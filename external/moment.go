package external

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// MomentLink is a link attachment of a moment; media_id is the cover image.
type MomentLink struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	MediaID string `json:"media_id,omitempty"`
}

// MomentAttachment is one attachment of a moment.
type MomentAttachment struct {
	MsgType string           `json:"msgtype"`
	Image   *MediaAttachment `json:"image,omitempty"`
	Video   *MediaAttachment `json:"video,omitempty"`
	Link    *MomentLink      `json:"link,omitempty"`
}

// MomentSenderList selects the members that may publish the moment.
type MomentSenderList struct {
	UserList       []string `json:"user_list,omitempty"`
	DepartmentList []int    `json:"department_list,omitempty"`
}

// MomentContactList selects the customers that can see the moment by tag.
type MomentContactList struct {
	TagList []string `json:"tag_list,omitempty"`
}

// MomentRange is the visible range of a moment.
type MomentRange struct {
	SenderList          *MomentSenderList  `json:"sender_list,omitempty"`
	ExternalContactList *MomentContactList `json:"external_contact_list,omitempty"`
}

// MomentTask is one enterprise moment to publish.
type MomentTask struct {
	Text         *TextContent       `json:"text,omitempty"`
	Attachments  []MomentAttachment `json:"attachments,omitempty"`
	VisibleRange *MomentRange       `json:"visible_range,omitempty"`
}

// MomentTaskResult is the outcome reported for one moment job.
type MomentTaskResult struct {
	Errcode            int                `json:"errcode"`
	Errmsg             string             `json:"errmsg"`
	MomentID           string             `json:"moment_id"`
	InvalidSenderList  *MomentSenderList  `json:"invalid_sender_list,omitempty"`
	InvalidContactList *MomentContactList `json:"invalid_external_contact_list,omitempty"`
}

// MomentJob is the state of one moment publishing job.
type MomentJob struct {
	Status int               `json:"status"`
	Type   string            `json:"type"`
	Result *MomentTaskResult `json:"result,omitempty"`
}

// AddMomentTask publishes one enterprise moment and returns the job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95094
func (c *External) AddMomentTask(ctx context.Context, task MomentTask) (string, error) {
	if task.Text == nil && len(task.Attachments) == 0 {
		return "", errors.New("wecom: add moment task: text or attachments are required")
	}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.call(ctx, "external add moment task",
		"/cgi-bin/externalcontact/add_moment_task", task, &out)
	return out.JobID, err
}

// GetMomentTaskResult returns the state of one publishing job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95094
func (c *External) GetMomentTaskResult(ctx context.Context, jobID string) (MomentJob, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return MomentJob{}, errors.New("wecom: moment task result: jobid is required")
	}
	var out MomentJob
	err := c.get(ctx, "external moment task result",
		"/cgi-bin/externalcontact/get_moment_task_result",
		url.Values{"jobid": {jobID}}, &out)
	return out, err
}

// CancelMomentTask stops publishing one moment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97612
func (c *External) CancelMomentTask(ctx context.Context, momentID string) error {
	momentID = strings.TrimSpace(momentID)
	if momentID == "" {
		return errors.New("wecom: cancel moment task: moment_id is required")
	}
	return c.call(ctx, "external cancel moment task",
		"/cgi-bin/externalcontact/cancel_moment_task",
		map[string]string{"moment_id": momentID}, nil)
}
