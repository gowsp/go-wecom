package message

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

const messageSendPath = "/cgi-bin/message/send"

// Send delivers one application message to toUser, which may be a single user
// id, several ids joined by "|", or "@all". The message body is validated before
// the request is sent.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90236
func (c *Message) Send(ctx context.Context, toUser string, msg AppMessage) error {
	if !c.CanSendMessages() {
		return errors.New("wecom: agent id is required to send application messages")
	}
	toUser = strings.TrimSpace(toUser)
	if toUser == "" {
		return errors.New("wecom: touser is required")
	}
	msgType, err := msg.body()
	if err != nil {
		return err
	}
	request := sendRequest{
		ToUser:                 toUser,
		MsgType:                msgType,
		AgentID:                c.agentID,
		Safe:                   msg.Safe,
		EnableIDTrans:          msg.EnableIDTrans,
		EnableDuplicateCheck:   msg.EnableDuplicateCheck,
		DuplicateCheckInterval: msg.DuplicateCheckInterval,
		Text:                   msg.Text,
		Image:                  msg.Image,
		Voice:                  msg.Voice,
		Video:                  msg.Video,
		File:                   msg.File,
		TextCard:               msg.TextCard,
		News:                   msg.News,
		MPNews:                 msg.MPNews,
		Markdown:               msg.Markdown,
		MiniProgramNotice:      msg.MiniProgramNotice,
		TemplateCard:           msg.TemplateCard,
	}
	return c.call(ctx, "message send", messageSendPath, request, nil)
}

// sendRequest mirrors the wire format of /cgi-bin/message/send.
type sendRequest struct {
	ToUser                 string             `json:"touser"`
	MsgType                string             `json:"msgtype"`
	AgentID                int                `json:"agentid"`
	Safe                   int                `json:"safe,omitempty"`
	EnableIDTrans          int                `json:"enable_id_trans,omitempty"`
	EnableDuplicateCheck   int                `json:"enable_duplicate_check,omitempty"`
	DuplicateCheckInterval int                `json:"duplicate_check_interval,omitempty"`
	Text                   *Text              `json:"text,omitempty"`
	Image                  *Media             `json:"image,omitempty"`
	Voice                  *Media             `json:"voice,omitempty"`
	Video                  *Video             `json:"video,omitempty"`
	File                   *Media             `json:"file,omitempty"`
	TextCard               *TextCard          `json:"textcard,omitempty"`
	News                   *News              `json:"news,omitempty"`
	MPNews                 *MPNews            `json:"mpnews,omitempty"`
	Markdown               *Text              `json:"markdown,omitempty"`
	MiniProgramNotice      *MiniProgramNotice `json:"miniprogram_notice,omitempty"`
	TemplateCard           json.RawMessage    `json:"template_card,omitempty"`
}
