package external

import (
	"context"
	"errors"
	"strings"
)

// MomentLocation is the location attached to a moment.
type MomentLocation struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

// MomentRecord is one published enterprise moment.
type MomentRecord struct {
	MomentID    string            `json:"moment_id"`
	Creator     string            `json:"creator"`
	CreateTime  int64             `json:"create_time"`
	CreateType  int               `json:"create_type"`
	VisibleType int               `json:"visible_type"`
	Text        *TextContent      `json:"text,omitempty"`
	Image       []MediaAttachment `json:"image,omitempty"`
	Video       *MomentVideo      `json:"video,omitempty"`
	Link        *MomentLink       `json:"link,omitempty"`
	Location    *MomentLocation   `json:"location,omitempty"`
}

// MomentVideo is the video of one moment.
type MomentVideo struct {
	MediaID      string `json:"media_id"`
	ThumbMediaID string `json:"thumb_media_id,omitempty"`
}

// MomentQuery filters the moment records.
type MomentQuery struct {
	StartTime  int64  `json:"start_time,omitempty"`
	EndTime    int64  `json:"end_time,omitempty"`
	Creator    string `json:"creator,omitempty"`
	FilterType int    `json:"filter_type,omitempty"`
	Cursor     string `json:"cursor,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// MomentRecordPage is one page of moment records.
type MomentRecordPage struct {
	NextCursor string         `json:"next_cursor"`
	Moments    []MomentRecord `json:"moment_list"`
}

// ListMomentRecords returns the enterprise moment publishing records.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93333
func (c *External) ListMomentRecords(ctx context.Context, query MomentQuery) (MomentRecordPage, error) {
	var out MomentRecordPage
	err := c.call(ctx, "external list moment records",
		"/cgi-bin/externalcontact/get_moment_list", query, &out)
	return out, err
}

// MomentPublish is the publishing state of one member for one moment.
type MomentPublish struct {
	UserID        string `json:"userid"`
	PublishStatus int    `json:"publish_status"`
}

// MomentTaskPage is one page of per-member publishing states.
type MomentTaskPage struct {
	NextCursor string          `json:"next_cursor"`
	Tasks      []MomentPublish `json:"task_list"`
}

// ListMomentTasks returns the publishing state of each member for one moment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93333
func (c *External) ListMomentTasks(ctx context.Context, momentID, cursor string, limit int) (MomentTaskPage, error) {
	if strings.TrimSpace(momentID) == "" {
		return MomentTaskPage{}, errors.New("wecom: moment tasks: moment_id is required")
	}
	request := struct {
		MomentID string `json:"moment_id"`
		Cursor   string `json:"cursor,omitempty"`
		Limit    int    `json:"limit,omitempty"`
	}{MomentID: momentID, Cursor: cursor, Limit: limit}
	var out MomentTaskPage
	err := c.call(ctx, "external moment tasks",
		"/cgi-bin/externalcontact/get_moment_task", request, &out)
	return out, err
}

// MomentReceiver is one member/customer pair that received the moment.
type MomentReceiver struct {
	UserID         string `json:"userid,omitempty"`
	ExternalUserID string `json:"external_userid,omitempty"`
}

// MomentReceiverPage is one page of receivers.
type MomentReceiverPage struct {
	NextCursor   string           `json:"next_cursor"`
	CustomerList []MomentReceiver `json:"customer_list"`
}

// ListMomentReceivers returns the customers of one member that received a moment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93333
func (c *External) ListMomentReceivers(ctx context.Context, momentID, userID, cursor string, limit int) (MomentReceiverPage, error) {
	if strings.TrimSpace(momentID) == "" || strings.TrimSpace(userID) == "" {
		return MomentReceiverPage{}, errors.New("wecom: moment receivers: moment_id and userid are required")
	}
	request := struct {
		MomentID string `json:"moment_id"`
		UserID   string `json:"userid"`
		Cursor   string `json:"cursor,omitempty"`
		Limit    int    `json:"limit,omitempty"`
	}{MomentID: momentID, UserID: userID, Cursor: cursor, Limit: limit}
	var out MomentReceiverPage
	err := c.call(ctx, "external moment receivers",
		"/cgi-bin/externalcontact/get_moment_customer_list", request, &out)
	return out, err
}

// MomentSendResult is one delivered moment for a customer.
type MomentSendResult struct {
	ExternalUserID string `json:"external_userid"`
}

// MomentSendPage is one page of delivered moments.
type MomentSendPage struct {
	NextCursor   string             `json:"next_cursor"`
	CustomerList []MomentSendResult `json:"customer_list"`
}

// ListMomentSendResults returns the moments one customer actually received.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93333
func (c *External) ListMomentSendResults(ctx context.Context, momentID, userID, cursor string, limit int) (MomentSendPage, error) {
	if strings.TrimSpace(momentID) == "" || strings.TrimSpace(userID) == "" {
		return MomentSendPage{}, errors.New("wecom: moment send results: moment_id and userid are required")
	}
	request := struct {
		MomentID string `json:"moment_id"`
		UserID   string `json:"userid"`
		Cursor   string `json:"cursor,omitempty"`
		Limit    int    `json:"limit,omitempty"`
	}{MomentID: momentID, UserID: userID, Cursor: cursor, Limit: limit}
	var out MomentSendPage
	err := c.call(ctx, "external moment send results",
		"/cgi-bin/externalcontact/get_moment_send_result", request, &out)
	return out, err
}

// MomentComment is one comment or like of a moment.
type MomentComment struct {
	UserID         string `json:"userid,omitempty"`
	ExternalUserID string `json:"external_userid,omitempty"`
	CreateTime     int64  `json:"create_time"`
}

// MomentComments is the comment and like list of one moment.
type MomentComments struct {
	Comments []MomentComment `json:"comment_list"`
	Likes    []MomentComment `json:"like_list"`
}

// GetMomentComments returns the comments and likes of one moment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93333
func (c *External) GetMomentComments(ctx context.Context, momentID, userID string) (MomentComments, error) {
	if strings.TrimSpace(momentID) == "" || strings.TrimSpace(userID) == "" {
		return MomentComments{}, errors.New("wecom: moment comments: moment_id and userid are required")
	}
	request := struct {
		MomentID string `json:"moment_id"`
		UserID   string `json:"userid"`
	}{MomentID: momentID, UserID: userID}
	var out MomentComments
	err := c.call(ctx, "external moment comments",
		"/cgi-bin/externalcontact/get_moment_comments", request, &out)
	return out, err
}
