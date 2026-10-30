package mail

import "context"

const (
	groupCreatePath = "/cgi-bin/exmail/group/create"
	groupUpdatePath = "/cgi-bin/exmail/group/update"
	groupDeletePath = "/cgi-bin/exmail/group/delete"
	groupGetPath    = "/cgi-bin/exmail/group/get"
	groupSearchPath = "/cgi-bin/exmail/group/search"
)

// Group is a mail group. Members come from member addresses, mail groups,
// departments and tags; on update an omitted list keeps the current members
// and an explicit empty list clears them.
type Group struct {
	GroupID             string       `json:"groupid"`
	GroupName           string       `json:"groupname,omitempty"`
	EmailList           List[string] `json:"email_list,omitempty"`
	TagList             List[int64]  `json:"tag_list,omitempty"`
	DepartmentList      List[int64]  `json:"department_list,omitempty"`
	GroupList           List[string] `json:"group_list,omitempty"`
	AllowType           int          `json:"allow_type,omitempty"`
	AllowEmailList      List[string] `json:"allow_emaillist,omitempty"`
	AllowDepartmentList List[int64]  `json:"allow_departmentlist,omitempty"`
	AllowTagList        List[int64]  `json:"allow_taglist,omitempty"`
}

// CreateGroup creates one mail group. GroupID must be a mail address.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95510
func (c *Mail) CreateGroup(ctx context.Context, group Group) error {
	if _, err := required("create mail group", "groupid", group.GroupID); err != nil {
		return err
	}
	return c.call(ctx, "create mail group", groupCreatePath, group, nil)
}

// UpdateGroup changes a mail group. Omitted member lists are left untouched.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97995
func (c *Mail) UpdateGroup(ctx context.Context, group Group) error {
	if _, err := required("update mail group", "groupid", group.GroupID); err != nil {
		return err
	}
	return c.call(ctx, "update mail group", groupUpdatePath, group, nil)
}

// DeleteGroup removes one mail group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97996
func (c *Mail) DeleteGroup(ctx context.Context, groupID string) error {
	groupID, err := required("delete mail group", "groupid", groupID)
	if err != nil {
		return err
	}
	return c.call(ctx, "delete mail group", groupDeletePath, map[string]string{"groupid": groupID}, nil)
}

// GetGroup returns one mail group created by the calling application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97997
func (c *Mail) GetGroup(ctx context.Context, groupID string) (Group, error) {
	groupID, err := required("get mail group", "groupid", groupID)
	if err != nil {
		return Group{}, err
	}
	var out Group
	err = c.get(ctx, "get mail group", groupGetPath, map[string]string{"groupid": groupID}, &out)
	return out, err
}

// SearchGroups lists mail groups. fuzzy enables fuzzy matching by name; groupID
// narrows the result to one group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97998
func (c *Mail) SearchGroups(ctx context.Context, fuzzy bool, groupID string) ([]Group, error) {
	query := map[string]string{"groupid": groupID}
	if fuzzy {
		query["fuzzy"] = "1"
	} else {
		query["fuzzy"] = "0"
	}
	var out struct {
		Groups []Group `json:"groups"`
	}
	err := c.get(ctx, "search mail groups", groupSearchPath, query, &out)
	return out.Groups, err
}
