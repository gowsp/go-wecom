package contact

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// encodingAESKeyLength is the fixed length of a 43-character AES key.
const encodingAESKeyLength = 43

// ExportRequest configures one asynchronous contact export. The exported CSV is
// encrypted with EncodingAESKey, whose matching key the caller keeps.
type ExportRequest struct {
	EncodingAESKey string `json:"encoding_aeskey"`
	BlockSize      int    `json:"block_size,omitempty"`
}

// ExportFile is one encrypted CSV chunk returned by an export task.
type ExportFile struct {
	URL  string `json:"url"`
	Size int64  `json:"size"`
	MD5  string `json:"md5"`
}

// ExportResult is the state of one export task. Status 2 means the files are
// ready in DataList; status 1 means the task is still running.
type ExportResult struct {
	Status   int          `json:"status"`
	DataList []ExportFile `json:"data_list"`
}

// ExportMembers starts exporting member ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94849
func (c *Contact) ExportMembers(ctx context.Context, request ExportRequest) (string, error) {
	return c.startExport(ctx, "contact export members", "/cgi-bin/export/simple_user", request, nil)
}

// ExportMemberDetails starts exporting full member details.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94851
func (c *Contact) ExportMemberDetails(ctx context.Context, request ExportRequest) (string, error) {
	return c.startExport(ctx, "contact export member details", "/cgi-bin/export/user", request, nil)
}

// ExportDepartments starts exporting departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94852
func (c *Contact) ExportDepartments(ctx context.Context, request ExportRequest) (string, error) {
	return c.startExport(ctx, "contact export departments", "/cgi-bin/export/department", request, nil)
}

// ExportTagMembers starts exporting the members of one tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94853
func (c *Contact) ExportTagMembers(ctx context.Context, tagID int, request ExportRequest) (string, error) {
	if tagID <= 0 {
		return "", errors.New("wecom: contact export tag members: tagid is required")
	}
	return c.startExport(ctx, "contact export tag members", "/cgi-bin/export/taguser", request, &tagID)
}

func (c *Contact) startExport(ctx context.Context, op, path string, request ExportRequest, tagID *int) (string, error) {
	key := strings.TrimSpace(request.EncodingAESKey)
	if len(key) != encodingAESKeyLength {
		return "", errors.New("wecom: " + op + ": encoding_aeskey must be 43 characters")
	}
	body := struct {
		EncodingAESKey string `json:"encoding_aeskey"`
		BlockSize      int    `json:"block_size,omitempty"`
		TagID          int    `json:"tagid,omitempty"`
	}{EncodingAESKey: key, BlockSize: request.BlockSize}
	if tagID != nil {
		body.TagID = *tagID
	}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.call(ctx, op, path, body, &out)
	return out.JobID, err
}

// ExportResult returns the state of one export task. The returned URLs are
// temporary, so download the chunks when the caller is ready.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94854
func (c *Contact) ExportResult(ctx context.Context, jobID string) (ExportResult, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return ExportResult{}, errors.New("wecom: export result: jobid is required")
	}
	var out ExportResult
	err := c.get(ctx, "contact export result", "/cgi-bin/export/get_result",
		url.Values{"jobid": {jobID}}, &out)
	return out, err
}
