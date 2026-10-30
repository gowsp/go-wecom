package agent

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// AgentInfo is one application. Fields mirror the official payload; writable and
// read-only fields share the struct, and SetAgent only sends what it needs.
type AgentInfo struct {
	AgentID           int             `json:"agentid,omitempty"`
	Name              string          `json:"name,omitempty"`
	SquareLogoURL     string          `json:"square_logo_url,omitempty"`
	Description       string          `json:"description,omitempty"`
	Close             int             `json:"close,omitempty"`
	RedirectDomain    string          `json:"redirect_domain,omitempty"`
	ReportLocation    int             `json:"report_location_flag,omitempty"`
	ReportEnter       int             `json:"isreportenter,omitempty"`
	HomeURL           string          `json:"home_url,omitempty"`
	AllowUserInfos    json.RawMessage `json:"allow_userinfos,omitempty"`
	AllowParties      json.RawMessage `json:"allow_partys,omitempty"`
	AllowTags         json.RawMessage `json:"allow_tags,omitempty"`
	DefaultCorpSecret string          `json:"-"`
}

// AgentBrief is one entry of the application list.
type AgentBrief struct {
	AgentID       int    `json:"agentid"`
	Name          string `json:"name"`
	SquareLogoURL string `json:"square_logo_url"`
}

// GetAgent reads one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90227
func (c *Agent) GetAgent(ctx context.Context, agentID int) (AgentInfo, error) {
	if agentID <= 0 {
		return AgentInfo{}, errors.New("wecom: get agent: agentid is required")
	}
	var out AgentInfo
	err := c.get(ctx, "agent get", "/cgi-bin/agent/get",
		url.Values{"agentid": {itoa(agentID)}}, &out)
	out.AgentID = agentID
	return out, err
}

// ListAgents returns the applications visible to the current secret.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90227
func (c *Agent) ListAgents(ctx context.Context) ([]AgentBrief, error) {
	var out struct {
		AgentList []AgentBrief `json:"agentlist"`
	}
	err := c.get(ctx, "agent list", "/cgi-bin/agent/list", nil, &out)
	return out.AgentList, err
}

// SetAgent updates one application. Only the fields set on the struct are sent.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90228
func (c *Agent) SetAgent(ctx context.Context, info AgentInfo) error {
	if info.AgentID <= 0 {
		return errors.New("wecom: set agent: agentid is required")
	}
	return c.call(ctx, "agent set", "/cgi-bin/agent/set", info, nil)
}

// Menu is the application menu. WeCom accepts at most three top-level buttons.
type Menu struct {
	Button []MenuButton `json:"button"`
}

// MenuButton is one menu entry, either a click/view action or a sub-menu.
type MenuButton struct {
	Type      string       `json:"type,omitempty"`
	Name      string       `json:"name"`
	Key       string       `json:"key,omitempty"`
	URL       string       `json:"url,omitempty"`
	AppID     string       `json:"appid,omitempty"`
	PagePath  string       `json:"pagepath,omitempty"`
	MediaID   string       `json:"media_id,omitempty"`
	ArticleID string       `json:"article_id,omitempty"`
	SubButton []MenuButton `json:"sub_button,omitempty"`
}

// CreateMenu replaces the menu of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90231
func (c *Agent) CreateMenu(ctx context.Context, agentID int, menu Menu) error {
	if agentID <= 0 {
		return errors.New("wecom: create menu: agentid is required")
	}
	if len(menu.Button) == 0 {
		return errors.New("wecom: create menu: at least one button is required")
	}
	if len(menu.Button) > 3 {
		return errors.New("wecom: create menu: at most three top-level buttons")
	}
	if err := validateButtons(menu.Button); err != nil {
		return err
	}
	return c.call(ctx, "menu create", "/cgi-bin/menu/create?agentid="+itoa(agentID), menu, nil)
}

// GetMenu returns the menu of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90232
func (c *Agent) GetMenu(ctx context.Context, agentID int) (Menu, error) {
	if agentID <= 0 {
		return Menu{}, errors.New("wecom: get menu: agentid is required")
	}
	var out Menu
	err := c.get(ctx, "menu get", "/cgi-bin/menu/get",
		url.Values{"agentid": {itoa(agentID)}}, &out)
	return out, err
}

// DeleteMenu removes the menu of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90233
func (c *Agent) DeleteMenu(ctx context.Context, agentID int) error {
	if agentID <= 0 {
		return errors.New("wecom: delete menu: agentid is required")
	}
	return c.get(ctx, "menu delete", "/cgi-bin/menu/delete",
		url.Values{"agentid": {itoa(agentID)}}, nil)
}

// SetWorkbenchData replaces the custom workbench items of one application. The
// payload is the documented template array and is passed through unchanged so
// new template types keep working.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92535
func (c *Agent) SetWorkbenchData(ctx context.Context, agentID int, data json.RawMessage) error {
	if agentID <= 0 {
		return errors.New("wecom: set workbench data: agentid is required")
	}
	if len(data) == 0 || !json.Valid(data) {
		return errors.New("wecom: set workbench data: data must be valid JSON")
	}
	request := struct {
		AgentID int             `json:"agentid"`
		Data    json.RawMessage `json:"data"`
	}{AgentID: agentID, Data: data}
	return c.call(ctx, "workbench set data", "/cgi-bin/agent/batch_set_workbench_data", request, nil)
}

// GetWorkbenchData returns the custom workbench items of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92535
func (c *Agent) GetWorkbenchData(ctx context.Context, agentID int) (json.RawMessage, error) {
	if agentID <= 0 {
		return nil, errors.New("wecom: get workbench data: agentid is required")
	}
	var out struct {
		Data json.RawMessage `json:"data"`
	}
	err := c.get(ctx, "workbench get data", "/cgi-bin/agent/get_workbench_data",
		url.Values{"agentid": {itoa(agentID)}}, &out)
	if err != nil {
		return nil, err
	}
	return out.Data, nil
}

// validateButtons checks the documented button shape: a button is either a
// container of sub-buttons or one typed action.
func validateButtons(buttons []MenuButton) error {
	for _, button := range buttons {
		if strings.TrimSpace(button.Name) == "" {
			return errors.New("wecom: create menu: every button needs a name")
		}
		if len(button.SubButton) > 0 {
			if len(button.SubButton) > 5 {
				return errors.New("wecom: create menu: at most five sub-buttons")
			}
			if err := validateButtons(button.SubButton); err != nil {
				return err
			}
			continue
		}
		if !menuType(button.Type) {
			return errors.New("wecom: create menu: unsupported button type " + button.Type)
		}
		switch button.Type {
		case "view":
			if strings.TrimSpace(button.URL) == "" {
				return errors.New("wecom: create menu: view button needs a url")
			}
		case "miniprogram":
			if button.AppID == "" || button.PagePath == "" {
				return errors.New("wecom: create menu: miniprogram button needs appid and pagepath")
			}
		}
	}
	return nil
}

// menuType validates the documented button types.
func menuType(name string) bool {
	switch strings.TrimSpace(name) {
	case "click", "view", "miniprogram", "scancode_push", "scancode_waitmsg",
		"pic_sysphoto", "pic_photo_or_album", "pic_weixin", "location_select":
		return true
	}
	return false
}
