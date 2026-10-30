package aibot

import (
	"encoding/json"
	"errors"
	"strings"
)

// Reply types accepted by the response_url endpoint and by passive replies.
const (
	ReplyText         = "text"
	ReplyMarkdown     = "markdown"
	ReplyTemplateCard = "template_card"
	ReplyStream       = "stream"
)

// maxMarkdownBytes is the official content limit for a markdown reply.
const maxMarkdownBytes = 20480

// Feedback makes a reply report user feedback to the callback URL.
type Feedback struct {
	ID string `json:"id,omitempty"`
}

// Reply is one reply payload. Set the field matching MsgType; TemplateCard is
// kept raw because card types differ per scenario.
type Reply struct {
	MsgType      string          `json:"msgtype"`
	Text         *Text           `json:"text,omitempty"`
	Markdown     *Markdown       `json:"markdown,omitempty"`
	TemplateCard json.RawMessage `json:"template_card,omitempty"`
	Stream       *Stream         `json:"stream,omitempty"`
}

// Markdown is a markdown reply body.
type Markdown struct {
	Content  string    `json:"content"`
	Feedback *Feedback `json:"feedback,omitempty"`
}

// Stream is a streaming reply body.
type Stream struct {
	ID       string       `json:"id,omitempty"`
	Finish   bool         `json:"finish"`
	Content  string       `json:"content,omitempty"`
	MsgItem  []StreamItem `json:"msg_item,omitempty"`
	Feedback *Feedback    `json:"feedback,omitempty"`
}

// StreamItem is one attachment of a stream reply.
type StreamItem struct {
	MsgType string       `json:"msgtype"`
	Image   *StreamImage `json:"image,omitempty"`
}

// StreamImage is an inline base64 image of a stream reply.
type StreamImage struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

// TextReply builds a text reply, used for the welcome message.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101031
func TextReply(content string) Reply {
	return Reply{MsgType: ReplyText, Text: &Text{Content: content}}
}

// MarkdownReply builds a markdown reply for an active response.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101138
func MarkdownReply(content string) Reply {
	return Reply{MsgType: ReplyMarkdown, Markdown: &Markdown{Content: content}}
}

// TemplateCardReply passes a template card through verbatim.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101031
// 文档: https://developer.work.weixin.qq.com/document/path/101138
func TemplateCardReply(card json.RawMessage) Reply {
	return Reply{MsgType: ReplyTemplateCard, TemplateCard: card}
}

// StreamReply builds a streaming reply; set Finish on the last chunk.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101031
func StreamReply(content string, finish bool) Reply {
	return Reply{MsgType: ReplyStream, Stream: &Stream{Content: content, Finish: finish}}
}

// Validate reports whether the reply matches its declared type. Both delivery
// modes call it before sending.
func (r Reply) Validate() error {
	return r.validate()
}

// validate reports whether the reply matches its declared type.
func (r Reply) validate() error {
	switch r.MsgType {
	case ReplyText:
		if r.Text == nil || strings.TrimSpace(r.Text.Content) == "" {
			return errors.New("wecom: robot reply: text.content is required")
		}
	case ReplyMarkdown:
		if r.Markdown == nil {
			return errors.New("wecom: robot reply: markdown body is required")
		}
		if strings.TrimSpace(r.Markdown.Content) == "" {
			return errors.New("wecom: robot reply: markdown.content is required")
		}
		if len(r.Markdown.Content) > maxMarkdownBytes {
			return errors.New("wecom: robot reply: markdown.content exceeds 20480 bytes")
		}
	case ReplyTemplateCard:
		if len(r.TemplateCard) == 0 || !json.Valid(r.TemplateCard) {
			return errors.New("wecom: robot reply: template_card must be valid JSON")
		}
	case ReplyStream:
		if r.Stream == nil {
			return errors.New("wecom: robot reply: stream body is required")
		}
		if !r.Stream.Finish && strings.TrimSpace(r.Stream.Content) == "" && len(r.Stream.MsgItem) == 0 {
			return errors.New("wecom: robot reply: unfinished stream needs content")
		}
	default:
		return errors.New("wecom: robot reply: unsupported msgtype " + r.MsgType)
	}
	return nil
}
