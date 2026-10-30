// Package kf implements the WeCom customer-service API (/cgi-bin/kf/*) on top
// of the shared authenticated client.
package kf

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/gowsp/go-wecom/client"
)

// KF provides the WeCom customer-service API.
type KF struct{ api *client.Client }

func New(api *client.Client) *KF { return &KF{api: api} }

// Session states accepted by the WeCom customer-service state API.
const (
	StateAssistant = 2 // this service handles the conversation
	// StateQueue parks the conversation in the reception queue. WeCom only
	// accepts a direct switch to the human state from specific states, so the
	// handoff goes through the queue with the target employee attached.
	StateQueue = 3
)

const (
	// stateTransPath is the documented "分配客服会话" endpoint (变更会话状态).
	stateTransPath = "/cgi-bin/kf/service_state/trans"
	// stateGetPath reads the current conversation state.
	stateGetPath = "/cgi-bin/kf/service_state/get"
)

// State returns the current customer-service state for one conversation:
// 0 未处理, 1 由接待人员接待, 2 由智能助手接待, 3 待接入池排队.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94669
func (c *KF) State(ctx context.Context, openKfID, externalUserID string) (int, error) {
	var out struct {
		ServiceState int `json:"service_state"`
	}
	err := c.api.Call(ctx, "customer service state get", stateGetPath, map[string]any{
		"open_kfid": openKfID, "external_userid": externalUserID,
	}, &out)
	return out.ServiceState, err
}

// Accept moves the conversation to the assistant state. WeCom only accepts
// send_msg while the session is in an assistant or human state; a queued
// session is rejected with errcode 95018.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94669
func (c *KF) Accept(ctx context.Context, openKfID, externalUserID string) error {
	return c.api.Call(ctx, "customer service state", stateTransPath, map[string]any{
		"open_kfid": openKfID, "external_userid": externalUserID,
		"service_state": StateAssistant,
	}, nil)
}

// Transfer hands a customer-service conversation to one reception employee by
// parking it in the reception queue with that employee as the servicer. WeCom
// rejects a direct jump to the human state (errcode 95016), and a queue entry
// without a servicer is rejected too (95014).
//
// 文档: https://developer.work.weixin.qq.com/document/path/94669
func (c *KF) Transfer(ctx context.Context, openKfID, externalUserID, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("customerservice: transfer requires a servicer user id")
	}
	return c.api.Call(ctx, "customer service transfer", stateTransPath, map[string]any{
		"open_kfid": openKfID, "external_userid": externalUserID,
		"service_state": StateQueue, "servicer_userid": userID,
	}, nil)
}

// Message origins reported by WeCom.
const (
	OriginCustomer = 3 // the customer sent it
	OriginSystem   = 4 // a system event
	OriginServicer = 5 // a reception employee sent it
)

// Voice formats accepted by SyncMsg.
const (
	VoiceFormatAmr  = 0
	VoiceFormatSilk = 1
)

// TextMessage is a text body. MenuID is set when the customer tapped a menu.
type TextMessage struct {
	Content string `json:"content"`
	MenuID  string `json:"menu_id,omitempty"`
}

// MediaMessage is an image, voice, video or file body: the media id is fetched
// with the media endpoint.
type MediaMessage struct {
	MediaID string `json:"media_id"`
}

// Message is one archived customer-service message. The body of the message
// depends on MsgType; typed bodies are decoded for the common kinds and Raw keeps
// the whole payload for the rest, including the event variants whose fields differ
// per event type.
type Message struct {
	MsgID          string `json:"msgid"`
	AccountID      string `json:"open_kfid,omitempty"`
	ExternalUserID string `json:"external_userid,omitempty"`
	SendTime       int64  `json:"send_time,omitempty"`
	Origin         int    `json:"origin,omitempty"`
	ServicerUserID string `json:"servicer_userid,omitempty"`
	MsgType        string `json:"msgtype"`
	Cursor         string `json:"cursor,omitempty"`

	Text  *TextMessage  `json:"text,omitempty"`
	Image *MediaMessage `json:"image,omitempty"`
	Voice *MediaMessage `json:"voice,omitempty"`
	Video *MediaMessage `json:"video,omitempty"`
	File  *MediaMessage `json:"file,omitempty"`
	// Event is the raw event object delivered with OriginSystem.
	Event json.RawMessage `json:"event,omitempty"`
	// Raw keeps the original JSON of the message.
	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON keeps the original payload alongside the typed fields.
func (m *Message) UnmarshalJSON(data []byte) error {
	type plain Message
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*m = Message(value)
	m.Raw = append(json.RawMessage(nil), data...)
	return nil
}

// SyncMsgRequest selects one page of archived messages.
type SyncMsgRequest struct {
	// OpenKfID is the customer-service account to read.
	OpenKfID string
	// Token is the token delivered by the callback; it is valid for ten
	// minutes and omitting it applies a strict rate limit.
	Token string
	// Cursor is the previous page's NextCursor; empty starts from the oldest
	// message kept, which is three days back.
	Cursor string
	// Limit is the requested page size. The documented default and maximum are
	// both 1000.
	Limit int
	// VoiceFormat selects the voice encoding: VoiceFormatAmr (default) or
	// VoiceFormatSilk.
	VoiceFormat int
}

// SyncPage is one page of archived messages. HasMore, not an empty message list,
// decides whether to keep going: a page may be empty and still have more.
type SyncPage struct {
	Messages   []Message
	NextCursor string
	HasMore    bool
}

// SyncMsg reads one page of archived messages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94670
func (c *KF) SyncMsg(ctx context.Context, request SyncMsgRequest) (SyncPage, error) {
	if strings.TrimSpace(request.OpenKfID) == "" {
		return SyncPage{}, errors.New("wecom: sync_msg: open_kfid is required")
	}
	limit := request.Limit
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}
	voiceFormat := request.VoiceFormat
	if voiceFormat != VoiceFormatAmr && voiceFormat != VoiceFormatSilk {
		return SyncPage{}, errors.New("wecom: sync_msg: voice_format must be 0 (amr) or 1 (silk)")
	}
	body := struct {
		OpenKfID    string `json:"open_kfid"`
		Token       string `json:"token,omitempty"`
		Cursor      string `json:"cursor,omitempty"`
		Limit       int    `json:"limit"`
		VoiceFormat int    `json:"voice_format"`
	}{OpenKfID: request.OpenKfID, Token: request.Token, Cursor: request.Cursor,
		Limit: limit, VoiceFormat: voiceFormat}
	var out struct {
		NextCursor string `json:"next_cursor"`
		// has_more is a uint32 flag, so it must not be decoded into a bool.
		HasMore int       `json:"has_more"`
		MsgList []Message `json:"msg_list"`
	}
	if err := c.api.Call(ctx, "customer service sync_msg", "/cgi-bin/kf/sync_msg", body, &out); err != nil {
		return SyncPage{}, err
	}
	return SyncPage{
		Messages:   out.MsgList,
		NextCursor: out.NextCursor,
		HasMore:    out.HasMore != 0,
	}, nil
}

// maxTextBytes is the documented limit for one customer-service text body.
const maxTextBytes = 2048

// 文档: https://developer.work.weixin.qq.com/document/path/94677
func (c *KF) SendText(ctx context.Context, openKfID, externalUserID, text string) error {
	return c.api.Call(ctx, "customer service send_msg", "/cgi-bin/kf/send_msg", map[string]any{
		"touser":    externalUserID,
		"open_kfid": openKfID,
		"msgtype":   "text",
		"text":      map[string]string{"content": truncateUTF8(text, maxTextBytes)},
	}, nil)
}

// truncateUTF8 cuts text to at most limit bytes without splitting a rune, which
// keeps the request valid when a model answer runs long.
func truncateUTF8(text string, limit int) string {
	if len(text) <= limit {
		return text
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(text[cut]) {
		cut--
	}
	return strings.TrimSpace(text[:cut])
}

// ServicerList is the reception roster of one customer-service account.
type ServicerList struct {
	UserIDs       []string
	DepartmentIDs []int64
}

// ListServicers returns the current reception users and departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94645
func (c *KF) ListServicers(ctx context.Context, openKfID string) (ServicerList, error) {
	var out struct {
		ServicerList []struct {
			UserID       string `json:"userid"`
			DepartmentID int64  `json:"department_id"`
		} `json:"servicer_list"`
	}
	if err := c.api.Get(ctx, "customer service servicer list", "/cgi-bin/kf/servicer/list", url.Values{
		"open_kfid": {openKfID},
	}, &out); err != nil {
		return ServicerList{}, err
	}
	var list ServicerList
	for _, entry := range out.ServicerList {
		if entry.UserID != "" {
			list.UserIDs = append(list.UserIDs, entry.UserID)
		}
		if entry.DepartmentID != 0 {
			list.DepartmentIDs = append(list.DepartmentIDs, entry.DepartmentID)
		}
	}
	return list, nil
}

// AddServicers grants reception duty to users and departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94646
func (c *KF) AddServicers(ctx context.Context, openKfID string, users []string, departments []int64) error {
	return c.api.Call(ctx, "customer service servicer add", "/cgi-bin/kf/servicer/add", servicerBody(openKfID, users, departments), nil)
}

// RemoveServicers revokes reception duty from users and departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94647
func (c *KF) RemoveServicers(ctx context.Context, openKfID string, users []string, departments []int64) error {
	return c.api.Call(ctx, "customer service servicer del", "/cgi-bin/kf/servicer/del", servicerBody(openKfID, users, departments), nil)
}

func servicerBody(openKfID string, users []string, departments []int64) map[string]any {
	if users == nil {
		users = []string{}
	}
	if departments == nil {
		departments = []int64{}
	}
	return map[string]any{
		"open_kfid": openKfID, "userid_list": users, "department_id_list": departments,
	}
}
