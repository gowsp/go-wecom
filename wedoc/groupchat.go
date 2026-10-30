package wedoc

import (
	"context"
	"errors"
	"strings"
)

const (
	groupChatListPath   = "/cgi-bin/wedoc/smartsheet/groupchat/list"
	groupChatGetPath    = "/cgi-bin/wedoc/smartsheet/groupchat/get"
	groupChatUpdatePath = "/cgi-bin/wedoc/smartsheet/groupchat/update"
)

// GroupChat is one chat created by a smart-sheet automation.
type GroupChat struct {
	ChatID   string   `json:"chat_id"`
	Name     string   `json:"name"`
	Owner    string   `json:"owner"`
	UserList []string `json:"user_list"`
}

// GroupChatList is one page of automation chats.
type GroupChatList struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	ChatIDs    []string `json:"chat_id_list"`
}

// GroupChatUpdate changes the owner or members of an automation chat.
type GroupChatUpdate struct {
	DocID       string   `json:"docid"`
	ChatID      string   `json:"chat_id"`
	Owner       string   `json:"owner,omitempty"`
	AddUserList []string `json:"add_user_list,omitempty"`
	DelUserList []string `json:"del_user_list,omitempty"`
}

// ListGroupChats returns the chats created by a smart-sheet document's
// automation, page by page.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100989
func (c *WeDoc) ListGroupChats(ctx context.Context, docID, cursor string, limit int) (GroupChatList, error) {
	docID, err := required("list group chats", "docId", docID)
	if err != nil {
		return GroupChatList{}, err
	}
	if limit <= 0 || limit > maxListPageSize {
		limit = maxListPageSize
	}
	request := struct {
		DocID  string `json:"docid"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit"`
	}{DocID: docID, Cursor: cursor, Limit: limit}
	var out GroupChatList
	err = c.call(ctx, "list group chats", groupChatListPath, request, &out)
	return out, err
}

// GetGroupChat returns one automation chat.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101028
func (c *WeDoc) GetGroupChat(ctx context.Context, docID, chatID string) (GroupChat, error) {
	docID, err := required("get group chat", "docId", docID)
	if err != nil {
		return GroupChat{}, err
	}
	chatID, err = required("get group chat", "chat_id", chatID)
	if err != nil {
		return GroupChat{}, err
	}
	request := struct {
		DocID  string `json:"docid"`
		ChatID string `json:"chat_id"`
	}{DocID: docID, ChatID: chatID}
	var out GroupChat
	err = c.call(ctx, "get group chat", groupChatGetPath, request, &out)
	return out, err
}

// UpdateGroupChat changes the owner or members of one automation chat.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101029
func (c *WeDoc) UpdateGroupChat(ctx context.Context, update GroupChatUpdate) error {
	update.DocID = strings.TrimSpace(update.DocID)
	update.ChatID = strings.TrimSpace(update.ChatID)
	if update.DocID == "" || update.ChatID == "" {
		return errors.New("wecom: update group chat: docId and chat_id are required")
	}
	if update.Owner == "" && len(update.AddUserList) == 0 && len(update.DelUserList) == 0 {
		return errors.New("wecom: update group chat: nothing to update")
	}
	return c.call(ctx, "update group chat", groupChatUpdatePath, update, nil)
}
