package external

import (
	"context"
	"errors"
	"strings"
)

// CorpTag is one enterprise customer tag.
type CorpTag struct {
	GroupName  string `json:"group_name,omitempty"`
	GroupID    string `json:"group_id,omitempty"`
	ID         string `json:"id,omitempty"`
	Name       string `json:"name,omitempty"`
	CreateTime int64  `json:"create_time,omitempty"`
	Order      int    `json:"order,omitempty"`
	Deleted    bool   `json:"deleted,omitempty"`
}

// CorpTagGroup is one tag group returned by the tag list.
type CorpTagGroup struct {
	GroupName  string    `json:"group_name"`
	GroupID    string    `json:"group_id"`
	CreateTime int64     `json:"create_time"`
	Order      int       `json:"order"`
	Deleted    bool      `json:"deleted"`
	Tags       []CorpTag `json:"tag"`
}

// TagFilter selects tag groups or tags when reading the tag list.
type TagFilter struct {
	GroupIDs []string `json:"group_id,omitempty"`
	TagIDs   []string `json:"tag_id,omitempty"`
	Type     int      `json:"type,omitempty"`
}

// ListCorpTags returns the enterprise tags, optionally filtered.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92117
func (c *External) ListCorpTags(ctx context.Context, filter TagFilter) ([]CorpTagGroup, error) {
	var out struct {
		TagGroups []CorpTagGroup `json:"tag_group"`
	}
	err := c.call(ctx, "external list corp tags", "/cgi-bin/externalcontact/get_corp_tag_list", filter, &out)
	return out.TagGroups, err
}

// AddCorpTags creates tags inside one group. An empty GroupID creates a new
// group named GroupName.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92117
func (c *External) AddCorpTags(ctx context.Context, groupID, groupName string, tags []CorpTag) ([]CorpTagGroup, error) {
	if len(tags) == 0 && strings.TrimSpace(groupName) == "" {
		return nil, errors.New("wecom: add corp tags: tags or group name is required")
	}
	request := struct {
		GroupID   string    `json:"group_id,omitempty"`
		GroupName string    `json:"group_name,omitempty"`
		Order     int       `json:"order,omitempty"`
		Tags      []CorpTag `json:"tag,omitempty"`
	}{GroupID: groupID, GroupName: groupName, Tags: tags}
	var out struct {
		TagGroups []CorpTagGroup `json:"tag_group"`
	}
	err := c.call(ctx, "external add corp tags", "/cgi-bin/externalcontact/add_corp_tag", request, &out)
	return out.TagGroups, err
}

// EditCorpTag renames one tag or group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92117
func (c *External) EditCorpTag(ctx context.Context, id, name string, order int) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return errors.New("wecom: edit corp tag: id and name are required")
	}
	request := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Order int    `json:"order,omitempty"`
	}{ID: id, Name: name, Order: order}
	return c.call(ctx, "external edit corp tag", "/cgi-bin/externalcontact/edit_corp_tag", request, nil)
}

// DeleteCorpTags removes tags or whole groups.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92117
func (c *External) DeleteCorpTags(ctx context.Context, tagIDs, groupIDs []string) error {
	if len(tagIDs) == 0 && len(groupIDs) == 0 {
		return errors.New("wecom: delete corp tags: tag_id or group_id is required")
	}
	request := struct {
		TagIDs   []string `json:"tag_id,omitempty"`
		GroupIDs []string `json:"group_id,omitempty"`
	}{TagIDs: tagIDs, GroupIDs: groupIDs}
	return c.call(ctx, "external delete corp tags", "/cgi-bin/externalcontact/del_corp_tag", request, nil)
}

// MarkCustomerTags adds and removes tags of one customer for one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92118
func (c *External) MarkCustomerTags(ctx context.Context, userID, externalUserID string, addTags, removeTags []string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(externalUserID) == "" {
		return errors.New("wecom: mark customer tags: userid and external_userid are required")
	}
	if len(addTags) == 0 && len(removeTags) == 0 {
		return errors.New("wecom: mark customer tags: add_tag or remove_tag is required")
	}
	request := struct {
		UserID         string   `json:"userid"`
		ExternalUserID string   `json:"external_userid"`
		AddTag         []string `json:"add_tag,omitempty"`
		RemoveTag      []string `json:"remove_tag,omitempty"`
	}{UserID: userID, ExternalUserID: externalUserID, AddTag: addTags, RemoveTag: removeTags}
	return c.call(ctx, "external mark customer tags", "/cgi-bin/externalcontact/mark_tag", request, nil)
}
