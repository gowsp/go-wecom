package contact

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// BatchCallback configures the notification sent when one asynchronous contact
// task finishes. It must point at a URL that can decrypt with Token.
type BatchCallback struct {
	URL            string `json:"url"`
	Token          string `json:"token"`
	EncodingAESKey string `json:"encodingaeskey"`
}

// BatchRequest starts one asynchronous import from a CSV media id.
type BatchRequest struct {
	MediaID  string         `json:"media_id"`
	ToInvite bool           `json:"to_invite,omitempty"`
	Callback *BatchCallback `json:"callback,omitempty"`
}

// BatchResultItem is one row result of an asynchronous task.
type BatchResultItem struct {
	UserID  string `json:"userid,omitempty"`
	PartyID int    `json:"partyid,omitempty"`
	Action  string `json:"action,omitempty"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// BatchResult is the state of one asynchronous task.
type BatchResult struct {
	Status     int               `json:"status"`
	Type       string            `json:"type"`
	Total      int               `json:"total"`
	Percentage int               `json:"percentage"`
	Result     []BatchResultItem `json:"result"`
}

// SyncMembers incrementally updates members from a CSV media id. Upload the CSV
// with media.Upload using type "file" first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90980
func (c *Contact) SyncMembers(ctx context.Context, request BatchRequest) (string, error) {
	return c.startBatch(ctx, "contact sync members", "/cgi-bin/batch/syncuser", request)
}

// ReplaceMembers overwrites all members from a CSV media id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90981
func (c *Contact) ReplaceMembers(ctx context.Context, request BatchRequest) (string, error) {
	return c.startBatch(ctx, "contact replace members", "/cgi-bin/batch/replaceuser", request)
}

// ReplaceDepartments overwrites all departments from a CSV media id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90982
func (c *Contact) ReplaceDepartments(ctx context.Context, request BatchRequest) (string, error) {
	return c.startBatch(ctx, "contact replace departments", "/cgi-bin/batch/replaceparty", request)
}

func (c *Contact) startBatch(ctx context.Context, op, path string, request BatchRequest) (string, error) {
	if strings.TrimSpace(request.MediaID) == "" {
		return "", errors.New("wecom: " + op + ": media_id is required")
	}
	if request.Callback != nil {
		callback := request.Callback
		if strings.TrimSpace(callback.URL) == "" || strings.TrimSpace(callback.Token) == "" {
			return "", errors.New("wecom: " + op + ": callback url and token are required")
		}
	}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.call(ctx, op, path, request, &out)
	return out.JobID, err
}

// BatchResult returns the state of one asynchronous task. Status 1 means the
// task is still running; poll only when the caller decides to.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90983
func (c *Contact) BatchResult(ctx context.Context, jobID string) (BatchResult, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return BatchResult{}, errors.New("wecom: batch result: jobid is required")
	}
	var out BatchResult
	err := c.get(ctx, "contact batch result", "/cgi-bin/batch/getresult",
		url.Values{"jobid": {jobID}}, &out)
	return out, err
}
