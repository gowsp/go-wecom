package wedrive

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/internal/blockhash"
)

// BlockSize is the fixed chunk size of a chunked upload: 2 MiB.
const BlockSize = 2 << 20

// maxInlineBytes is the documented ceiling of UploadFile; larger files must be
// uploaded in blocks.
const maxInlineBytes = 10 << 20

// BlockSHA computes the cumulative SHA of every upload block, which is what
// UploadInit expects. See the package note on blockhash for why the values cannot
// be produced with hash/sha1 alone.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98004
func BlockSHA(content []byte) []string {
	return blockhash.SHA1(content, BlockSize)
}

// UploadChunked uploads content with the three-step block protocol and returns the
// new file id. It is the convenience wrapper for files above the 10 MiB inline
// limit; use UploadInit, UploadBlock and UploadFinish directly to parallelise the
// blocks, which the protocol allows, or to resume a broken upload. Size and
// BlockSHA in the request are computed from content and may be left empty.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98004
func (c *WeDrive) UploadChunked(ctx context.Context, location UploadInitRequest, content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("wecom: upload chunked: content is empty")
	}
	location.FileName = strings.TrimSpace(location.FileName)
	location.Size = int64(len(content))
	location.BlockSHA = BlockSHA(content)
	session, err := c.UploadInit(ctx, location)
	if err != nil {
		return "", err
	}
	if session.HitExist {
		// The server already stored an identical file, so no block is needed.
		return session.FileID, nil
	}
	for index, offset := 0, 0; offset < len(content); index, offset = index+1, offset+BlockSize {
		end := offset + BlockSize
		if end > len(content) {
			end = len(content)
		}
		block := base64.StdEncoding.EncodeToString(content[offset:end])
		if err := c.UploadBlock(ctx, session.UploadKey, index+1, block); err != nil {
			return "", err
		}
	}
	return c.UploadFinish(ctx, session.UploadKey)
}

// InlineBase64 encodes content for UploadFile and rejects files that must use the
// block protocol instead.
func InlineBase64(content []byte) (string, error) {
	if len(content) == 0 {
		return "", errors.New("wecom: upload file: content is empty")
	}
	if len(content) > maxInlineBytes {
		return "", errors.New("wecom: upload file: content exceeds the 10 MiB inline limit, use UploadChunked")
	}
	return base64.StdEncoding.EncodeToString(content), nil
}
