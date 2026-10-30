package external

import (
	"context"
	"errors"
	"strings"
)

// TextContent is a plain text block.
type TextContent struct {
	Content string `json:"content"`
}

// ImageAttachment is an image attachment.
type ImageAttachment struct {
	MediaID string `json:"media_id,omitempty"`
	PicURL  string `json:"pic_url,omitempty"`
}

// LinkAttachment is a link attachment.
type LinkAttachment struct {
	Title  string `json:"title"`
	PicURL string `json:"picurl,omitempty"`
	Desc   string `json:"desc,omitempty"`
	URL    string `json:"url"`
}

// MiniProgramAttachment is a mini program attachment.
type MiniProgramAttachment struct {
	Title      string `json:"title"`
	PicMediaID string `json:"pic_media_id,omitempty"`
	AppID      string `json:"appid"`
	Page       string `json:"page"`
}

// MediaAttachment is a file or video attachment.
type MediaAttachment struct {
	MediaID string `json:"media_id"`
}

// Attachment is one entry of a message template. Set the field matching
// MsgType.
type Attachment struct {
	MsgType     string                 `json:"msgtype"`
	Image       *ImageAttachment       `json:"image,omitempty"`
	Link        *LinkAttachment        `json:"link,omitempty"`
	MiniProgram *MiniProgramAttachment `json:"miniprogram,omitempty"`
	Video       *MediaAttachment       `json:"video,omitempty"`
	File        *MediaAttachment       `json:"file,omitempty"`
}

// TagFilterRequest selects recipients by tag groups for a group send.
type TagFilterRequest struct {
	GroupList []TagGroupFilter `json:"group_list"`
}

// TagGroupFilter is one tag group selection; tags inside a group are OR-ed.
type TagGroupFilter struct {
	TagList []string `json:"tag_list"`
}

// MsgTemplate is one enterprise group send.
type MsgTemplate struct {
	ChatType       string            `json:"chat_type,omitempty"`
	ExternalUserID []string          `json:"external_userid,omitempty"`
	ChatIDList     []string          `json:"chat_id_list,omitempty"`
	TagFilter      *TagFilterRequest `json:"tag_filter,omitempty"`
	Sender         string            `json:"sender,omitempty"`
	AllowSelect    bool              `json:"allow_select,omitempty"`
	Text           *TextContent      `json:"text,omitempty"`
	Attachments    []Attachment      `json:"attachments,omitempty"`
}

// MsgTemplateResult reports the send task and the recipients that failed.
type MsgTemplateResult struct {
	MsgID    string   `json:"msgid"`
	FailList []string `json:"fail_list"`
}

// AddMsgTemplate creates one enterprise group send and returns its task id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92135
func (c *External) AddMsgTemplate(ctx context.Context, template MsgTemplate) (MsgTemplateResult, error) {
	if template.Text == nil && len(template.Attachments) == 0 {
		return MsgTemplateResult{}, errors.New("wecom: add msg template: text or attachments are required")
	}
	if strings.TrimSpace(template.Sender) == "" {
		return MsgTemplateResult{}, errors.New("wecom: add msg template: sender is required")
	}
	var out MsgTemplateResult
	err := c.call(ctx, "external add msg template", "/cgi-bin/externalcontact/add_msg_template", template, &out)
	return out, err
}

// RemindGroupMsgSend reminds members who have not sent the task yet.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97610
func (c *External) RemindGroupMsgSend(ctx context.Context, msgID string) error {
	if strings.TrimSpace(msgID) == "" {
		return errors.New("wecom: remind group msg: msgid is required")
	}
	return c.call(ctx, "external remind group msg", "/cgi-bin/externalcontact/remind_groupmsg_send",
		map[string]string{"msgid": msgID}, nil)
}

// CancelGroupMsgSend stops a running group send.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97611
func (c *External) CancelGroupMsgSend(ctx context.Context, msgID string) error {
	if strings.TrimSpace(msgID) == "" {
		return errors.New("wecom: cancel group msg: msgid is required")
	}
	return c.call(ctx, "external cancel group msg", "/cgi-bin/externalcontact/cancel_groupmsg_send",
		map[string]string{"msgid": msgID}, nil)
}

// GroupMsgQuery filters the group send records.
type GroupMsgQuery struct {
	ChatType   string `json:"chat_type,omitempty"`
	StartTime  int64  `json:"start_time,omitempty"`
	EndTime    int64  `json:"end_time,omitempty"`
	Creator    string `json:"creator,omitempty"`
	FilterType int    `json:"filter_type,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
}

// GroupMsgTask is one group send record.
type GroupMsgTask struct {
	MsgID       string       `json:"msgid"`
	Creator     string       `json:"creator,omitempty"`
	CreateTime  string       `json:"create_time,omitempty"`
	CreateType  int          `json:"create_type,omitempty"`
	Text        *TextContent `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// GroupMsgTaskPage is one page of group send records.
type GroupMsgTaskPage struct {
	NextCursor string         `json:"next_cursor"`
	Tasks      []GroupMsgTask `json:"group_msg_list"`
}

// ListGroupMsgTasks returns the group send records of the enterprise.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93338
func (c *External) ListGroupMsgTasks(ctx context.Context, query GroupMsgQuery) (GroupMsgTaskPage, error) {
	var out GroupMsgTaskPage
	err := c.call(ctx, "external list group msg tasks",
		"/cgi-bin/externalcontact/get_groupmsg_list_v2", query, &out)
	return out, err
}

// GetGroupMsgTask reads one group send record.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93338
func (c *External) GetGroupMsgTask(ctx context.Context, msgID string) (GroupMsgTask, error) {
	if strings.TrimSpace(msgID) == "" {
		return GroupMsgTask{}, errors.New("wecom: get group msg task: msgid is required")
	}
	var out struct {
		Task GroupMsgTask `json:"task"`
	}
	err := c.call(ctx, "external get group msg task",
		"/cgi-bin/externalcontact/get_groupmsg_task", map[string]string{"msgid": msgID}, &out)
	return out.Task, err
}

// GroupMsgSendResult is one recipient's delivery state.
type GroupMsgSendResult struct {
	ExternalUserID string `json:"external_userid,omitempty"`
	UserID         string `json:"userid,omitempty"`
	ChatID         string `json:"chat_id,omitempty"`
	Status         int    `json:"status,omitempty"`
	SendTime       int64  `json:"send_time,omitempty"`
}

// GroupMsgSendPage is one page of delivery results.
type GroupMsgSendPage struct {
	NextCursor string               `json:"next_cursor"`
	SendList   []GroupMsgSendResult `json:"send_list"`
}

// ListGroupMsgSendResults returns the delivery state of one group send.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93338
func (c *External) ListGroupMsgSendResults(ctx context.Context, msgID, userID string, limit int, cursor string) (GroupMsgSendPage, error) {
	if strings.TrimSpace(msgID) == "" {
		return GroupMsgSendPage{}, errors.New("wecom: group msg send results: msgid is required")
	}
	request := struct {
		MsgID  string `json:"msgid"`
		UserID string `json:"userid,omitempty"`
		Limit  int    `json:"limit,omitempty"`
		Cursor string `json:"cursor,omitempty"`
	}{MsgID: msgID, UserID: userID, Limit: limit, Cursor: cursor}
	var out GroupMsgSendPage
	err := c.call(ctx, "external group msg send results",
		"/cgi-bin/externalcontact/get_groupmsg_send_result", request, &out)
	return out, err
}

// WelcomeMessage is the greeting sent to one new customer.
type WelcomeMessage struct {
	WelcomeCode string       `json:"welcome_code"`
	Text        *TextContent `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// SendWelcomeMsg answers a new-customer callback. The welcome code is valid for
// 20 seconds and can be used once, so failures cannot be retried later.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92137
func (c *External) SendWelcomeMsg(ctx context.Context, message WelcomeMessage) error {
	if strings.TrimSpace(message.WelcomeCode) == "" {
		return errors.New("wecom: send welcome msg: welcome_code is required")
	}
	if message.Text == nil && len(message.Attachments) == 0 {
		return errors.New("wecom: send welcome msg: text or attachments are required")
	}
	return c.call(ctx, "external send welcome msg", "/cgi-bin/externalcontact/send_welcome_msg", message, nil)
}
