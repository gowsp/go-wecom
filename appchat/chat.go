package appchat

import (
	"context"
	"errors"
	"strings"
)

const (
	createAppChatPath = "/cgi-bin/appchat/create"
	sendAppChatPath   = "/cgi-bin/appchat/send"
)

// CreateAppChat creates a group chat for the self-built application and
// returns its chatid. Owner and every member must be in the application's
// visible range. A non-empty chatID is used verbatim and may be at most 32
// characters; empty lets WeCom generate one.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90245
func (c *AppChat) CreateAppChat(ctx context.Context, name, owner string, userIDs []string, chatID string) (string, error) {
	name, err := required("create app chat", "name", name)
	if err != nil {
		return "", err
	}
	owner, err = required("create app chat", "owner", owner)
	if err != nil {
		return "", err
	}
	users := normalizeUserIDs(userIDs)
	if len(users) == 0 {
		return "", errors.New("wecom: create app chat: at least one user is required")
	}
	chatID = strings.TrimSpace(chatID)
	if len([]rune(chatID)) > 32 {
		return "", errors.New("wecom: create app chat: chatId must be at most 32 characters")
	}
	request := struct {
		Name     string   `json:"name"`
		Owner    string   `json:"owner"`
		UserList []string `json:"userlist"`
		ChatID   string   `json:"chatid,omitempty"`
	}{Name: name, Owner: owner, UserList: users, ChatID: chatID}
	var response struct {
		ChatID string `json:"chatid"`
	}
	if err := c.call(ctx, "create app chat", createAppChatPath, request, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.ChatID) == "" {
		return "", errors.New("wecom: create app chat response is missing chatid")
	}
	return response.ChatID, nil
}

// SendAppChatText sends one plain-text message to an application group chat.
// Unlike message/send, this endpoint is addressed by chatid and carries no
// agent id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90248
func (c *AppChat) SendAppChatText(ctx context.Context, chatID, content string) error {
	chatID, err := required("send app chat message", "chatId", chatID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("wecom: send app chat message: content is required")
	}
	request := struct {
		ChatID  string         `json:"chatid"`
		MsgType string         `json:"msgtype"`
		Text    messageContent `json:"text"`
	}{ChatID: chatID, MsgType: msgTypeText, Text: messageContent{Content: content}}
	return c.call(ctx, "send app chat message", sendAppChatPath, request, nil)
}
