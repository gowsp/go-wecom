package contact

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Tag is one contact tag.
type Tag struct {
	ID   int    `json:"tagid"`
	Name string `json:"tagname"`
}

// TagMember is one member of a tag.
type TagMember struct {
	UserID string `json:"userid"`
	Name   string `json:"name"`
}

// TagDetail is the content of one tag.
type TagDetail struct {
	Name    string      `json:"tagname"`
	Users   []TagMember `json:"userlist"`
	Parties []int       `json:"partylist"`
}

// TagResult reports members that could not be added or removed.
type TagResult struct {
	InvalidUsers   []string `json:"invalidlist"`
	InvalidParties []int    `json:"invalidparty"`
}

// CreateTag adds one tag and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90210
func (c *Contact) CreateTag(ctx context.Context, name string) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, errors.New("wecom: create tag: tagname is required")
	}
	var out struct {
		ID int `json:"tagid"`
	}
	err := c.call(ctx, "contact create tag", "/cgi-bin/tag/create",
		map[string]string{"tagname": name}, &out)
	return out.ID, err
}

// UpdateTag renames one tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90211
func (c *Contact) UpdateTag(ctx context.Context, id int, name string) error {
	if id <= 0 {
		return errors.New("wecom: update tag: tagid is required")
	}
	if strings.TrimSpace(name) == "" {
		return errors.New("wecom: update tag: tagname is required")
	}
	request := struct {
		ID   int    `json:"tagid"`
		Name string `json:"tagname"`
	}{ID: id, Name: name}
	return c.call(ctx, "contact update tag", "/cgi-bin/tag/update", request, nil)
}

// DeleteTag removes one tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90212
func (c *Contact) DeleteTag(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("wecom: delete tag: tagid is required")
	}
	return c.get(ctx, "contact delete tag", "/cgi-bin/tag/delete",
		url.Values{"tagid": {itoa(id)}}, nil)
}

// GetTag reads the members and departments of one tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90213
func (c *Contact) GetTag(ctx context.Context, id int) (TagDetail, error) {
	if id <= 0 {
		return TagDetail{}, errors.New("wecom: get tag: tagid is required")
	}
	var out TagDetail
	err := c.get(ctx, "contact get tag", "/cgi-bin/tag/get",
		url.Values{"tagid": {itoa(id)}}, &out)
	return out, err
}

// AddTagMembers adds members and departments to a tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90214
func (c *Contact) AddTagMembers(ctx context.Context, id int, users []string, parties []int) (TagResult, error) {
	return c.changeTagMembers(ctx, "contact add tag members", "/cgi-bin/tag/addtagusers", id, users, parties)
}

// DeleteTagMembers removes members and departments from a tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90215
func (c *Contact) DeleteTagMembers(ctx context.Context, id int, users []string, parties []int) (TagResult, error) {
	return c.changeTagMembers(ctx, "contact delete tag members", "/cgi-bin/tag/deltagusers", id, users, parties)
}

func (c *Contact) changeTagMembers(ctx context.Context, op, path string, id int, users []string, parties []int) (TagResult, error) {
	if id <= 0 {
		return TagResult{}, errors.New("wecom: " + op + ": tagid is required")
	}
	if len(users) == 0 && len(parties) == 0 {
		return TagResult{}, errors.New("wecom: " + op + ": userlist and partylist are empty")
	}
	request := struct {
		ID        int      `json:"tagid"`
		UserList  []string `json:"userlist,omitempty"`
		PartyList []int    `json:"partylist,omitempty"`
	}{ID: id, UserList: users, PartyList: parties}
	var out TagResult
	err := c.call(ctx, op, path, request, &out)
	return out, err
}

// ListTags returns every tag of the enterprise.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90216
func (c *Contact) ListTags(ctx context.Context) ([]Tag, error) {
	var out struct {
		Tags []Tag `json:"taglist"`
	}
	err := c.get(ctx, "contact list tags", "/cgi-bin/tag/list", nil, &out)
	return out.Tags, err
}
