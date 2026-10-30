package external

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// Group chat member types and join scenes are plain integers in the API; the
// values below name the ones the documentation lists explicitly.
const (
	GroupMemberUser  = 1
	GroupMemberCorp  = 2
	JoinSceneCreated = 1
	JoinSceneQRCode  = 2
)

// GroupChatMember is one member of a customer group chat.
type GroupChatMember struct {
	UserID        string           `json:"userid,omitempty"`
	Type          int              `json:"type"`
	JoinTime      int64            `json:"join_time,omitempty"`
	JoinScene     int              `json:"join_scene,omitempty"`
	Invitor       *GroupChatMember `json:"invitor,omitempty"`
	GroupNickname string           `json:"group_nickname,omitempty"`
	Name          string           `json:"name,omitempty"`
	UnionID       string           `json:"unionid,omitempty"`
}

// GroupChat is one customer group chat.
type GroupChat struct {
	ChatID     string            `json:"chat_id"`
	Status     int               `json:"status,omitempty"`
	Owner      string            `json:"owner,omitempty"`
	CreateTime int64             `json:"create_time,omitempty"`
	Notice     string            `json:"notice,omitempty"`
	MemberList []GroupChatMember `json:"member_list,omitempty"`
	AdminList  []GroupChatMember `json:"admin_list,omitempty"`
	RoomName   string            `json:"room_name,omitempty"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

// GroupChatPage is one page of the customer group chat list.
type GroupChatPage struct {
	Chats      []GroupChat `json:"group_chat_list"`
	NextCursor string      `json:"next_cursor"`
}

// GroupChatFilter narrows the group chat list by status or owner.
type GroupChatFilter struct {
	StatusFilter int          `json:"status_filter,omitempty"`
	OwnerFilter  *OwnerFilter `json:"owner_filter,omitempty"`
}

// OwnerFilter selects chats by owner userid.
type OwnerFilter struct {
	UserIDList []string `json:"userid_list"`
}

// ListGroupChats returns one page of customer group chats.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92120
func (c *External) ListGroupChats(ctx context.Context, filter GroupChatFilter, cursor string, limit int) (GroupChatPage, error) {
	request := struct {
		StatusFilter int          `json:"status_filter,omitempty"`
		OwnerFilter  *OwnerFilter `json:"owner_filter,omitempty"`
		Cursor       string       `json:"cursor,omitempty"`
		Limit        int          `json:"limit,omitempty"`
	}{StatusFilter: filter.StatusFilter, OwnerFilter: filter.OwnerFilter, Cursor: cursor, Limit: limit}
	var out GroupChatPage
	err := c.call(ctx, "external list group chats", "/cgi-bin/externalcontact/groupchat/list", request, &out)
	return out, err
}

// GetGroupChat reads one customer group chat. needName asks for member display
// names, which requires the contact permission.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92122
func (c *External) GetGroupChat(ctx context.Context, chatID string, needName bool) (GroupChat, error) {
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return GroupChat{}, errors.New("wecom: get group chat: chat_id is required")
	}
	request := struct {
		ChatID   string `json:"chat_id"`
		NeedName bool   `json:"need_name,omitempty"`
	}{ChatID: chatID, NeedName: needName}
	var out struct {
		GroupChat GroupChat `json:"group_chat"`
	}
	err := c.call(ctx, "external get group chat", "/cgi-bin/externalcontact/groupchat/get", request, &out)
	return out.GroupChat, err
}

// ConvertOpenGID maps an open GID from a mini program to a chat id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94822
func (c *External) ConvertOpenGID(ctx context.Context, openGID string) (string, error) {
	openGID = strings.TrimSpace(openGID)
	if openGID == "" {
		return "", errors.New("wecom: convert opengid: opengid is required")
	}
	var out struct {
		ChatID string `json:"chat_id"`
	}
	err := c.call(ctx, "external convert opengid", "/cgi-bin/externalcontact/opengid_to_chatid",
		map[string]string{"opengid": openGID}, &out)
	return out.ChatID, err
}

// GroupChatJoinWay is one "join chat" configuration.
type GroupChatJoinWay struct {
	ConfigID       string   `json:"config_id,omitempty"`
	Scene          int      `json:"scene,omitempty"`
	Remark         string   `json:"remark,omitempty"`
	AutoCreateRoom int      `json:"auto_create_room,omitempty"`
	RoomBaseName   string   `json:"room_base_name,omitempty"`
	RoomBaseID     int      `json:"room_base_id,omitempty"`
	ChatIDList     []string `json:"chat_id_list,omitempty"`
	QRCode         string   `json:"qr_code,omitempty"`
}

// AddJoinWay creates one join-chat configuration and returns its id, or the
// existing id when the same configuration already exists.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92229
func (c *External) AddJoinWay(ctx context.Context, way GroupChatJoinWay) (string, error) {
	if len(way.ChatIDList) == 0 {
		return "", errors.New("wecom: add join way: chat_id_list is empty")
	}
	var out struct {
		ConfigID string `json:"config_id"`
	}
	err := c.call(ctx, "external add join way", "/cgi-bin/externalcontact/groupchat/add_join_way", way, &out)
	return out.ConfigID, err
}

// GetJoinWay reads one join-chat configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92229
func (c *External) GetJoinWay(ctx context.Context, configID string) (GroupChatJoinWay, error) {
	if strings.TrimSpace(configID) == "" {
		return GroupChatJoinWay{}, errors.New("wecom: get join way: config_id is required")
	}
	var out struct {
		JoinWay GroupChatJoinWay `json:"join_way"`
	}
	err := c.call(ctx, "external get join way", "/cgi-bin/externalcontact/groupchat/get_join_way",
		map[string]string{"config_id": configID}, &out)
	return out.JoinWay, err
}

// UpdateJoinWay replaces one join-chat configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92229
func (c *External) UpdateJoinWay(ctx context.Context, way GroupChatJoinWay) error {
	if strings.TrimSpace(way.ConfigID) == "" {
		return errors.New("wecom: update join way: config_id is required")
	}
	return c.call(ctx, "external update join way", "/cgi-bin/externalcontact/groupchat/update_join_way", way, nil)
}

// DeleteJoinWay removes one join-chat configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92229
func (c *External) DeleteJoinWay(ctx context.Context, configID string) error {
	if strings.TrimSpace(configID) == "" {
		return errors.New("wecom: delete join way: config_id is required")
	}
	return c.call(ctx, "external delete join way", "/cgi-bin/externalcontact/groupchat/del_join_way",
		map[string]string{"config_id": configID}, nil)
}

// ContactWay is one "contact me" configuration. The request and response share
// the struct; Conclusions stays raw because its shape depends on the scene.
type ContactWay struct {
	ConfigID      string          `json:"config_id,omitempty"`
	Type          int             `json:"type,omitempty"`
	Scene         int             `json:"scene,omitempty"`
	Style         int             `json:"style,omitempty"`
	Remark        string          `json:"remark,omitempty"`
	SkipVerify    bool            `json:"skip_verify,omitempty"`
	State         string          `json:"state,omitempty"`
	User          []string        `json:"user,omitempty"`
	Party         []int           `json:"party,omitempty"`
	IsTemp        bool            `json:"is_temp,omitempty"`
	ExpiresIn     int             `json:"expires_in,omitempty"`
	ChatExpiresIn int             `json:"chat_expires_in,omitempty"`
	UnionID       string          `json:"unionid,omitempty"`
	Conclusions   json.RawMessage `json:"conclusions,omitempty"`
	QRCode        string          `json:"qr_code,omitempty"`
}

// AddContactWay creates one contact-me configuration and returns its id and QR
// code URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92228
func (c *External) AddContactWay(ctx context.Context, way ContactWay) (ContactWay, error) {
	if way.Type == 0 || way.Scene == 0 {
		return ContactWay{}, errors.New("wecom: add contact way: type and scene are required")
	}
	var out ContactWay
	err := c.call(ctx, "external add contact way", "/cgi-bin/externalcontact/add_contact_way", way, &out)
	return out, err
}

// GetContactWay reads one contact-me configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92228
func (c *External) GetContactWay(ctx context.Context, configID string) (ContactWay, error) {
	if strings.TrimSpace(configID) == "" {
		return ContactWay{}, errors.New("wecom: get contact way: config_id is required")
	}
	var out struct {
		ContactWay ContactWay `json:"contact_way"`
	}
	err := c.get(ctx, "external get contact way", "/cgi-bin/externalcontact/get_contact_way",
		map[string][]string{"config_id": {configID}}, &out)
	return out.ContactWay, err
}

// UpdateContactWay changes one contact-me configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92228
func (c *External) UpdateContactWay(ctx context.Context, way ContactWay) error {
	if strings.TrimSpace(way.ConfigID) == "" {
		return errors.New("wecom: update contact way: config_id is required")
	}
	return c.call(ctx, "external update contact way", "/cgi-bin/externalcontact/update_contact_way", way, nil)
}

// DeleteContactWay removes one contact-me configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92228
func (c *External) DeleteContactWay(ctx context.Context, configID string) error {
	if strings.TrimSpace(configID) == "" {
		return errors.New("wecom: delete contact way: config_id is required")
	}
	return c.call(ctx, "external delete contact way", "/cgi-bin/externalcontact/del_contact_way",
		map[string]string{"config_id": configID}, nil)
}
