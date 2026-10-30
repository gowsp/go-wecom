package wedoc

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/gowsp/go-wecom/internal/blockhash"
)

// FileBlockSize is the fixed chunk size of a chunked document upload: 2 MiB.
const FileBlockSize = 2 << 20

const (
	imageUploadPath      = "/cgi-bin/wedoc/image_upload"
	fileUploadPath       = "/cgi-bin/wedoc/file_upload"
	fileUploadInitPath   = "/cgi-bin/wedoc/file_upload_init"
	fileUploadPartPath   = "/cgi-bin/wedoc/file_upload_part"
	fileUploadFinishPath = "/cgi-bin/wedoc/file_upload_finish"
)

// ImageResult describes one uploaded document image.
type ImageResult struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
	Size   int    `json:"size"`
}

// ChunkUpload is the state of a chunked file upload. HitExist reports that the
// server already holds the file and returned FileURL directly.
type ChunkUpload struct {
	HitExist  bool   `json:"hit_exist"`
	UploadKey string `json:"upload_key"`
	FileURL   string `json:"file_url"`
}

// UploadImage stores one image inside a document and returns its URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99933
func (c *WeDoc) UploadImage(ctx context.Context, docID string, content []byte) (ImageResult, error) {
	docID, err := required("upload doc image", "docId", docID)
	if err != nil {
		return ImageResult{}, err
	}
	if len(content) == 0 {
		return ImageResult{}, errEmpty("upload doc image", "content")
	}
	body := map[string]string{"docid": docID, "base64_content": base64.StdEncoding.EncodeToString(content)}
	var out ImageResult
	err = c.call(ctx, "upload doc image", imageUploadPath, body, &out)
	return out, err
}

// UploadFile stores one file inside a document in a single request.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101870
func (c *WeDoc) UploadFile(ctx context.Context, docID, name string, content []byte) (string, error) {
	docID, err := required("upload doc file", "docId", docID)
	if err != nil {
		return "", err
	}
	if name, err = required("upload doc file", "file_name", name); err != nil {
		return "", err
	}
	if len(content) == 0 {
		return "", errEmpty("upload doc file", "content")
	}
	body := map[string]string{
		"docid":               docID,
		"file_name":           name,
		"file_base64_content": base64.StdEncoding.EncodeToString(content),
	}
	var out struct {
		URL string `json:"file_url"`
	}
	err = c.call(ctx, "upload doc file", fileUploadPath, body, &out)
	return out.URL, err
}

// InitChunkUpload starts a chunked upload. blockSHA holds the cumulative SHA of
// the 2 MB blocks and lets the server skip already uploaded content.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101871
func (c *WeDoc) InitChunkUpload(ctx context.Context, docID, name string, size int64, blockSHA []string) (ChunkUpload, error) {
	docID, err := required("init chunk upload", "docId", docID)
	if err != nil {
		return ChunkUpload{}, err
	}
	if name, err = required("init chunk upload", "file_name", name); err != nil {
		return ChunkUpload{}, err
	}
	if size <= 0 {
		return ChunkUpload{}, errors.New("wecom: init chunk upload: size is required")
	}
	if len(blockSHA) == 0 {
		return ChunkUpload{}, errors.New("wecom: init chunk upload: block_sha is required")
	}
	body := struct {
		DocID    string   `json:"docid"`
		FileName string   `json:"file_name"`
		Size     int64    `json:"size"`
		BlockSHA []string `json:"block_sha"`
	}{DocID: docID, FileName: name, Size: size, BlockSHA: blockSHA}
	var out ChunkUpload
	if err := c.call(ctx, "init chunk upload", fileUploadInitPath, body, &out); err != nil {
		return ChunkUpload{}, err
	}
	if !out.HitExist && out.UploadKey == "" {
		return ChunkUpload{}, errors.New("wecom: init chunk upload: response carried neither upload_key nor hit_exist")
	}
	return out, nil
}

// BlockSHA computes the cumulative SHA of every 2 MiB block for InitChunkUpload.
// The values cannot be produced with hash/sha1 alone, so use this helper rather
// than hashing each block.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101871
func BlockSHA(content []byte) []string {
	return blockhash.SHA1(content, FileBlockSize)
}

// UploadChunked uploads content with the three-step block protocol and returns the
// document file URL. It is the convenience wrapper for files above the 10 MiB
// inline limit; use the three steps directly to parallelise or resume.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101871
func (c *WeDoc) UploadChunked(ctx context.Context, docID, name string, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errEmpty("upload chunked", "content")
	}
	upload, err := c.InitChunkUpload(ctx, docID, name, int64(len(content)), BlockSHA(content))
	if err != nil {
		return "", err
	}
	if upload.HitExist {
		return upload.FileURL, nil
	}
	for index, offset := 0, 0; offset < len(content); index, offset = index+1, offset+FileBlockSize {
		end := offset + FileBlockSize
		if end > len(content) {
			end = len(content)
		}
		if err := c.UploadChunk(ctx, upload.UploadKey, index+1, content[offset:end]); err != nil {
			return "", err
		}
	}
	return c.FinishChunkUpload(ctx, upload.UploadKey)
}

// UploadChunk sends one 2 MB block. index starts at 1.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101871
func (c *WeDoc) UploadChunk(ctx context.Context, uploadKey string, index int, block []byte) error {
	uploadKey, err := required("upload chunk", "upload_key", uploadKey)
	if err != nil {
		return err
	}
	if index <= 0 {
		return errors.New("wecom: upload chunk: index must be positive")
	}
	body := map[string]any{
		"upload_key":          uploadKey,
		"index":               index,
		"file_base64_content": base64.StdEncoding.EncodeToString(block),
	}
	return c.call(ctx, "upload chunk", fileUploadPartPath, body, nil)
}

// FinishChunkUpload completes a chunked upload and returns the document file URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101871
func (c *WeDoc) FinishChunkUpload(ctx context.Context, uploadKey string) (string, error) {
	uploadKey, err := required("finish chunk upload", "upload_key", uploadKey)
	if err != nil {
		return "", err
	}
	var out struct {
		FileURL string `json:"file_url"`
	}
	if err := c.call(ctx, "finish chunk upload", fileUploadFinishPath,
		map[string]string{"upload_key": uploadKey}, &out); err != nil {
		return "", err
	}
	return out.FileURL, nil
}
