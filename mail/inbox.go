package mail

import "context"

const (
	mailListPath     = "/cgi-bin/exmail/app/get_mail_list"
	readMailPath     = "/cgi-bin/exmail/app/read_mail"
	newMailCountPath = "/cgi-bin/exmail/mail/get_newcount"
)

// MailRef is one entry of the inbox list.
type MailRef struct {
	MailID string `json:"mail_id"`
}

// MailList is one page of inbox mail ids.
type MailList struct {
	NextCursor string    `json:"next_cursor"`
	HasMore    int       `json:"has_more"`
	Mails      []MailRef `json:"mail_list"`
}

// List returns inbox mail ids in a time range. The first call may omit cursor;
// limit defaults to 100 on the server and is capped at 1000.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97369
func (c *Mail) List(ctx context.Context, beginTime, endTime int64, cursor string, limit int) (MailList, error) {
	request := struct {
		BeginTime int64  `json:"begin_time"`
		EndTime   int64  `json:"end_time"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{BeginTime: beginTime, EndTime: endTime, Cursor: cursor, Limit: limit}
	var out MailList
	err := c.call(ctx, "list mail", mailListPath, request, &out)
	return out, err
}

// Read returns the raw eml content of one mail.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97979
func (c *Mail) Read(ctx context.Context, mailID string) (string, error) {
	mailID, err := required("read mail", "mail_id", mailID)
	if err != nil {
		return "", err
	}
	var out struct {
		Data string `json:"mail_data"`
	}
	err = c.call(ctx, "read mail", readMailPath, map[string]string{"mail_id": mailID}, &out)
	return out.Data, err
}

// UnreadCount returns how many unread mails one member has.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95514
func (c *Mail) UnreadCount(ctx context.Context, userID string) (int, error) {
	userID, err := required("mail unread count", "userid", userID)
	if err != nil {
		return 0, err
	}
	var out struct {
		Count int `json:"count"`
	}
	err = c.call(ctx, "mail unread count", newMailCountPath, map[string]string{"userid": userID}, &out)
	return out.Count, err
}
