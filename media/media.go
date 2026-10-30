// Package media implements the WeCom material API (/cgi-bin/media/*) on top of
// the shared authenticated client.
package media

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/gowsp/go-wecom/client"
)

// Media types accepted by Upload and UploadByURL.
const (
	TypeImage = "image"
	TypeVoice = "voice"
	TypeVideo = "video"
	TypeFile  = "file"
)

var errClient = errors.New("wecom: media client is unavailable")

// Media provides material upload and download using one shared client.
type Media struct{ api *client.Client }

// New creates a material client.
func New(api *client.Client) *Media { return &Media{api: api} }

// Item is the result of uploading one temporary material.
type Item struct {
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

// AsyncJob is the state of one asynchronous upload task.
type AsyncJob struct {
	Status int   `json:"status"`
	Detail *Item `json:"detail"`
}

// UploadRequest starts an asynchronous upload from a public URL.
type UploadRequest struct {
	Scene    int    `json:"scene,omitempty"`
	Type     string `json:"type"`
	FileName string `json:"filename,omitempty"`
	URL      string `json:"url"`
	MD5      string `json:"md5,omitempty"`
}

// Upload stores one temporary material and returns its media id. name is the
// original file name and is required for voice and video materials.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90253
func (c *Media) Upload(ctx context.Context, mediaType, name string, data []byte) (Item, error) {
	if c == nil || c.api == nil {
		return Item{}, errClient
	}
	if err := validate(mediaType, name, data); err != nil {
		return Item{}, err
	}
	var out Item
	err := c.api.Upload(ctx, "media upload", "/cgi-bin/media/upload", map[string][]string{"type": {mediaType}},
		"media", name, data, &out)
	return out, err
}

// UploadImage stores one image and returns its URL. WeCom requires JPG or PNG
// no larger than 2 MB for this endpoint.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90256
func (c *Media) UploadImage(ctx context.Context, name string, data []byte) (string, error) {
	if c == nil || c.api == nil {
		return "", errClient
	}
	if err := validate(TypeImage, name, data); err != nil {
		return "", err
	}
	var out struct {
		URL string `json:"url"`
	}
	err := c.api.Upload(ctx, "media upload image", "/cgi-bin/media/uploadimg", nil, "media", name, data, &out)
	return out.URL, err
}

// Download returns the content of one temporary material.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90254
func (c *Media) Download(ctx context.Context, mediaID string) ([]byte, error) {
	return c.download(ctx, "media get", "/cgi-bin/media/get", mediaID)
}

// DownloadVoice returns the high-quality voice material used by JS-SDK.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90255
func (c *Media) DownloadVoice(ctx context.Context, mediaID string) ([]byte, error) {
	return c.download(ctx, "media get voice", "/cgi-bin/media/get/jssdk", mediaID)
}

func (c *Media) download(ctx context.Context, op, path, mediaID string) ([]byte, error) {
	if c == nil || c.api == nil {
		return nil, errClient
	}
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		return nil, errors.New("wecom: " + op + ": media_id is required")
	}
	return c.api.Download(ctx, op, path, map[string][]string{"media_id": {mediaID}})
}

// UploadByURL starts an asynchronous material upload and returns the job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/96219
func (c *Media) UploadByURL(ctx context.Context, request UploadRequest) (string, error) {
	if c == nil || c.api == nil {
		return "", errClient
	}
	if strings.TrimSpace(request.URL) == "" {
		return "", errors.New("wecom: media upload_by_url: url is required")
	}
	if strings.TrimSpace(request.Type) == "" {
		return "", errors.New("wecom: media upload_by_url: type is required")
	}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.api.Call(ctx, "media upload_by_url", "/cgi-bin/media/upload_by_url", request, &out)
	return out.JobID, err
}

// UploadResult returns the state of an asynchronous upload. Status 2 means the
// job finished and Detail carries the material.
//
// 文档: https://developer.work.weixin.qq.com/document/path/96219
func (c *Media) UploadResult(ctx context.Context, jobID string) (AsyncJob, error) {
	if c == nil || c.api == nil {
		return AsyncJob{}, errClient
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return AsyncJob{}, errors.New("wecom: media upload result: jobid is required")
	}
	var out AsyncJob
	err := c.api.Call(ctx, "media upload result", "/cgi-bin/media/get_upload_by_url_result",
		map[string]string{"jobid": jobID}, &out)
	return out, err
}

func validate(mediaType, name string, data []byte) error {
	switch mediaType {
	case TypeImage, TypeVoice, TypeVideo, TypeFile:
	default:
		return errors.New("wecom: media upload: unsupported type")
	}
	if len(data) == 0 {
		return errors.New("wecom: media upload: content is empty")
	}
	if strings.TrimSpace(name) == "" || utf8.RuneCountInString(name) > 120 {
		return errors.New("wecom: media upload: file name is required")
	}
	return nil
}
