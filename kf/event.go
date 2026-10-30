package kf

import (
	"context"
	"errors"
	"strings"
)

// EventMessage is the body of a welcome or other event response message. Only
// the field matching MsgType is sent.
type EventMessage struct {
	// Code is the event code from the callback, for example welcome_code. It can
	// be used only once.
	Code    string `json:"code"`
	MsgID   string `json:"msgid,omitempty"`
	MsgType string `json:"msgtype"`

	Text        *TextContent        `json:"text,omitempty"`
	Image       *MediaContent       `json:"image,omitempty"`
	MiniProgram *MiniProgramContent `json:"miniprogram,omitempty"`
	Menu        *MenuContent        `json:"msgmenu,omitempty"`
	Link        *LinkContent        `json:"link,omitempty"`
}

// TextContent is a plain text body.
type TextContent struct {
	Content string `json:"content"`
}

// MediaContent references an uploaded media id.
type MediaContent struct {
	MediaID string `json:"media_id"`
}

// MiniProgramContent opens a mini program.
type MiniProgramContent struct {
	AppID        string `json:"appid"`
	Title        string `json:"title,omitempty"`
	ThumbMediaID string `json:"thumb_media_id,omitempty"`
	PagePath     string `json:"pagepath,omitempty"`
}

// MenuContent is a clickable menu body.
type MenuContent struct {
	HeadContent string     `json:"head_content,omitempty"`
	List        []MenuItem `json:"list"`
	TailContent string     `json:"tail_content,omitempty"`
}

// MenuItem is one clickable menu entry.
type MenuItem struct {
	Type    string `json:"type"`
	ClickID string `json:"click_id,omitempty"`
	Content string `json:"content"`
}

// LinkContent is a card-style link body.
type LinkContent struct {
	Title        string `json:"title"`
	Desc         string `json:"desc,omitempty"`
	URL          string `json:"url"`
	ThumbMediaID string `json:"thumb_media_id,omitempty"`
}

// SendEventText answers an event, typically the enter-session welcome, with one
// text message and returns the message id assigned by the server.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95122
func (c *KF) SendEventText(ctx context.Context, code, content string) (string, error) {
	return c.SendEventMessage(ctx, EventMessage{
		Code:    code,
		MsgType: "text",
		Text:    &TextContent{Content: truncateUTF8(content, maxTextBytes)},
	})
}

// SendEventMessage answers an event with the message described by request. The
// event code can be used only once, so a failure has to be handled by the
// caller instead of being retried later.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95122
func (c *KF) SendEventMessage(ctx context.Context, request EventMessage) (string, error) {
	request.Code = strings.TrimSpace(request.Code)
	if request.Code == "" {
		return "", errors.New("wecom: customer service send_msg_on_event: code is required")
	}
	if strings.TrimSpace(request.MsgType) == "" {
		return "", errors.New("wecom: customer service send_msg_on_event: msgtype is required")
	}
	var out struct {
		MsgID string `json:"msgid"`
	}
	err := c.api.Call(ctx, "customer service send_msg_on_event", "/cgi-bin/kf/send_msg_on_event", request, &out)
	return out.MsgID, err
}
