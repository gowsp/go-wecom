// Package wedrive implements the WeCom micro disk API (/cgi-bin/wedrive):
// spaces, files, permissions, sharing and change callbacks.
package wedrive

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// WeDrive manages micro disk spaces and files with one shared transport.
type WeDrive struct{ api *client.Client }

// New creates a micro disk client.
func New(api *client.Client) *WeDrive { return &WeDrive{api: api} }

// Member types used by access lists.
const (
	MemberUser       = 1
	MemberDepartment = 2
)

// Access rights.
const (
	// AuthRead can browse and download.
	AuthRead = 1
	// AuthUpload can add files.
	AuthUpload = 2
	// AuthEdit can modify and delete.
	AuthEdit = 3
	// AuthManage can change space settings.
	AuthManage = 7
)

// AuthInfo grants one member or department access to a space.
type AuthInfo struct {
	Type         int    `json:"type"`
	UserID       string `json:"userid,omitempty"`
	DepartmentID int    `json:"departmentid,omitempty"`
	Auth         int    `json:"auth,omitempty"`
}

// SpaceAuthList is the access list of a space.
type SpaceAuthList struct {
	AuthInfo []AuthInfo `json:"auth_info"`
	// QuitUserID lists members who dropped out of the space or lost their
	// permission while still belonging to an authorised department.
	QuitUserID []string `json:"quit_userid,omitempty"`
}

// SpaceInfo is one micro disk space. Quota fields are returned only by the
// detailed endpoint.
type SpaceInfo struct {
	SpaceID      string          `json:"spaceid"`
	SpaceName    string          `json:"space_name"`
	AuthList     *SpaceAuthList  `json:"auth_list,omitempty"`
	SpaceSubType int             `json:"space_sub_type,omitempty"`
	IsCorpSpace  int             `json:"is_corp_space,omitempty"`
	Quota        json.RawMessage `json:"quota,omitempty"`
	SourceFrom   json.RawMessage `json:"source_from,omitempty"`
	CreateTime   int64           `json:"create_time,omitempty"`
	Creator      string          `json:"creator,omitempty"`
}

// SpaceSetting is the security configuration of one space.
type SpaceSetting struct {
	SpaceID                      string `json:"spaceid"`
	EnableWatermark              *bool  `json:"enable_watermark,omitempty"`
	ShareURLNoApprove            *bool  `json:"share_url_no_approve,omitempty"`
	ShareURLNoApproveDefaultAuth int    `json:"share_url_no_approve_default_auth,omitempty"`
	EnableConfidentialMode       *bool  `json:"enable_confidential_mode,omitempty"`
	DefaultFileScope             int    `json:"default_file_scope,omitempty"`
	BanShareExternal             *bool  `json:"ban_share_external,omitempty"`
}

// CreateSpace creates one space and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93655
func (c *WeDrive) CreateSpace(ctx context.Context, name string, subType int, auth []AuthInfo) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("wecom: create space: space_name is required")
	}
	request := struct {
		SpaceName    string     `json:"space_name"`
		AuthInfo     []AuthInfo `json:"auth_info,omitempty"`
		SpaceSubType int        `json:"space_sub_type,omitempty"`
	}{SpaceName: name, AuthInfo: auth, SpaceSubType: subType}
	var out struct {
		SpaceID string `json:"spaceid"`
	}
	err := c.api.Call(ctx, "wedrive create space", "/cgi-bin/wedrive/space_create", request, &out)
	return out.SpaceID, err
}

// RenameSpace changes the name of one space.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97856
func (c *WeDrive) RenameSpace(ctx context.Context, spaceID, name string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(spaceID) == "" || strings.TrimSpace(name) == "" {
		return errors.New("wecom: rename space: spaceid and space_name are required")
	}
	request := struct {
		SpaceID   string `json:"spaceid"`
		SpaceName string `json:"space_name"`
	}{SpaceID: spaceID, SpaceName: name}
	return c.api.Call(ctx, "wedrive rename space", "/cgi-bin/wedrive/space_rename", request, nil)
}

// DismissSpace deletes one space and everything in it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97857
func (c *WeDrive) DismissSpace(ctx context.Context, spaceID string) error {
	if err := check(c); err != nil {
		return err
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return errors.New("wecom: dismiss space: spaceid is required")
	}
	return c.api.Call(ctx, "wedrive dismiss space", "/cgi-bin/wedrive/space_dismiss",
		map[string]string{"spaceid": spaceID}, nil)
}

// GetSpace reads one space, including its access list and quota.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97858
func (c *WeDrive) GetSpace(ctx context.Context, spaceID string) (SpaceInfo, error) {
	if err := check(c); err != nil {
		return SpaceInfo{}, err
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return SpaceInfo{}, errors.New("wecom: get space: spaceid is required")
	}
	var out struct {
		Info SpaceInfo `json:"space_info"`
	}
	err := c.api.Call(ctx, "wedrive get space", "/cgi-bin/wedrive/space_info",
		map[string]string{"spaceid": spaceID}, &out)
	return out.Info, err
}

// GetNewSpaceInfo reads the detailed space information, including the newer
// quota and source fields.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97878
func (c *WeDrive) GetNewSpaceInfo(ctx context.Context, spaceID string) (SpaceInfo, error) {
	if err := check(c); err != nil {
		return SpaceInfo{}, err
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return SpaceInfo{}, errors.New("wecom: new space info: spaceid is required")
	}
	var out struct {
		Info SpaceInfo `json:"space_info"`
	}
	err := c.api.Call(ctx, "wedrive new space info", "/cgi-bin/wedrive/new_space_info",
		map[string]string{"spaceid": spaceID}, &out)
	return out.Info, err
}

// AddSpaceMembers grants access to members and departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93656
func (c *WeDrive) AddSpaceMembers(ctx context.Context, spaceID string, auth []AuthInfo) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(spaceID) == "" || len(auth) == 0 {
		return errors.New("wecom: add space members: spaceid and auth_info are required")
	}
	// auth is mandatory here and carries the granted permission: 1 downloads
	// only, 7 makes the member a space administrator.
	if err := checkAuthInfo(auth, true); err != nil {
		return errors.New("wecom: add space members: " + err.Error())
	}
	request := struct {
		SpaceID  string     `json:"spaceid"`
		AuthInfo []AuthInfo `json:"auth_info"`
	}{SpaceID: spaceID, AuthInfo: auth}
	return c.api.Call(ctx, "wedrive add space members", "/cgi-bin/wedrive/space_acl_add", request, nil)
}

// RemoveSpaceMembers revokes access of members and departments.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97875
func (c *WeDrive) RemoveSpaceMembers(ctx context.Context, spaceID string, auth []AuthInfo) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(spaceID) == "" || len(auth) == 0 {
		return errors.New("wecom: remove space members: spaceid and auth_info are required")
	}
	if err := checkAuthInfo(auth, false); err != nil {
		return errors.New("wecom: remove space members: " + err.Error())
	}
	request := struct {
		SpaceID  string     `json:"spaceid"`
		AuthInfo []AuthInfo `json:"auth_info"`
	}{SpaceID: spaceID, AuthInfo: auth}
	return c.api.Call(ctx, "wedrive remove space members", "/cgi-bin/wedrive/space_acl_del", request, nil)
}

// SetSpaceSetting updates the security configuration of one space.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97876
func (c *WeDrive) SetSpaceSetting(ctx context.Context, setting SpaceSetting) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(setting.SpaceID) == "" {
		return errors.New("wecom: space setting: spaceid is required")
	}
	return c.api.Call(ctx, "wedrive space setting", "/cgi-bin/wedrive/space_setting", setting, nil)
}

// SpaceShareLink returns the invite link of one space.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97877
func (c *WeDrive) SpaceShareLink(ctx context.Context, spaceID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return "", errors.New("wecom: space share link: spaceid is required")
	}
	var out struct {
		URL string `json:"space_share_url"`
	}
	err := c.api.Call(ctx, "wedrive space share link", "/cgi-bin/wedrive/space_share",
		map[string]string{"spaceid": spaceID}, &out)
	return out.URL, err
}

func check(c *WeDrive) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: wedrive client is unavailable")
	}
	return nil
}
