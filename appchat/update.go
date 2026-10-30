package appchat

import (
	"context"
	"errors"
	"strings"
)

// Chat is one application group chat.
type Chat struct {
	ChatID   string   `json:"chatid"`
	Name     string   `json:"name"`
	Owner    string   `json:"owner"`
	UserList []string `json:"userlist"`
	ChatType int      `json:"chat_type"`
}

// UpdateRequest changes a group chat. Empty fields are left unchanged.
type UpdateRequest struct {
	ChatID      string   `json:"chatid"`
	Name        string   `json:"name,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	AddUserList []string `json:"add_user_list,omitempty"`
	DelUserList []string `json:"del_user_list,omitempty"`
}

// Update changes the name, owner or members of an application group chat.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98913
func (c *AppChat) Update(ctx context.Context, request UpdateRequest) error {
	request.ChatID = strings.TrimSpace(request.ChatID)
	if request.ChatID == "" {
		return errors.New("wecom: update app chat: chatId is required")
	}
	if request.Name == "" && request.Owner == "" && len(request.AddUserList) == 0 && len(request.DelUserList) == 0 {
		return errors.New("wecom: update app chat: nothing to update")
	}
	return c.call(ctx, "update app chat", "/cgi-bin/appchat/update", request, nil)
}

// Get returns one application group chat by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98914
func (c *AppChat) Get(ctx context.Context, chatID string) (Chat, error) {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return Chat{}, errors.New("wecom: get app chat: chatId is required")
	}
	var out struct {
		Info Chat `json:"chat_info"`
	}
	err := c.api.Get(ctx, "get app chat", "/cgi-bin/appchat/get",
		map[string][]string{"chatid": {chatID}}, &out)
	if err != nil {
		return Chat{}, err
	}
	return out.Info, nil
}
