package message

import (
	"context"
	"errors"
)

const (
	recallPath             = "/cgi-bin/message/recall"
	updateTemplateCardPath = "/cgi-bin/message/update_template_card"
)

// Recall withdraws one message previously sent by the application. WeCom only
// allows recalling messages sent within 24 hours.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94867
func (c *Message) Recall(ctx context.Context, msgID string) error {
	msgID, err := required("recall message", "msgid", msgID)
	if err != nil {
		return err
	}
	return c.call(ctx, "recall message", recallPath, map[string]string{"msgid": msgID}, nil)
}

// TemplateCardUpdate replaces the content of one sent template card. At least
// one recipient selector and a response code are required.
type TemplateCardUpdate struct {
	UserIDs      []string    `json:"userids,omitempty"`
	PartyIDs     []int64     `json:"partyids,omitempty"`
	TagIDs       []int64     `json:"tagids,omitempty"`
	AtAll        int         `json:"atall,omitempty"`
	AgentID      int         `json:"agentid"`
	ResponseCode string      `json:"response_code"`
	Button       Button      `json:"button,omitempty"`
	CardAction   *CardAction `json:"card_action,omitempty"`
}

// Button replaces the text of the card button.
type Button struct {
	ReplaceName string `json:"replace_name,omitempty"`
}

// CardAction replaces the click target of the card.
type CardAction struct {
	Type     int    `json:"type"`
	URL      string `json:"url,omitempty"`
	AppID    string `json:"appid,omitempty"`
	PagePath string `json:"pagepath,omitempty"`
}

// UpdateTemplateCard updates a sent template card and returns the recipients
// that could not be reached.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94888
func (c *Message) UpdateTemplateCard(ctx context.Context, update TemplateCardUpdate) ([]string, error) {
	if !c.CanSendMessages() {
		return nil, errors.New("wecom: agent id is required to update template cards")
	}
	if len(update.UserIDs) == 0 && len(update.PartyIDs) == 0 && len(update.TagIDs) == 0 {
		return nil, errors.New("wecom: update template card: at least one recipient is required")
	}
	if _, err := required("update template card", "response_code", update.ResponseCode); err != nil {
		return nil, err
	}
	if update.AgentID == 0 {
		update.AgentID = c.agentID
	}
	var out struct {
		Invalid []string `json:"invaliduser"`
	}
	err := c.call(ctx, "update template card", updateTemplateCardPath, update, &out)
	return out.Invalid, err
}
