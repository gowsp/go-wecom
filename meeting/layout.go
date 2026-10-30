package meeting

import (
	"context"
	"errors"
	"strings"
)

// LayoutTemplate is one official layout template.
type LayoutTemplate struct {
	TemplateID   string `json:"layout_template_id"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	PictureURL   string `json:"picture_url,omitempty"`
	RenderRule   string `json:"render_rule,omitempty"`
}

// PollingSetting is the auto-polling configuration of an advanced layout page.
type PollingSetting struct {
	IntervalUnit      int  `json:"polling_interval_unit,omitempty"`
	Interval          int  `json:"polling_interval,omitempty"`
	IgnoreUserNoVideo bool `json:"ignore_user_novideo,omitempty"`
	IgnoreUserAbsence bool `json:"ignore_user_absence,omitempty"`
}

// Seat assigns one participant to one grid cell of a layout page.
type Seat struct {
	GridID    string `json:"grid_id"`
	GridType  int    `json:"grid_type,omitempty"`
	UserID    string `json:"userid,omitempty"`
	TmpOpenID string `json:"tmp_openid,omitempty"`
	NickName  string `json:"nick_name,omitempty"`
	ToolSDKID string `json:"tool_sdkid,omitempty"`
}

// LayoutPage is one screen page of a layout.
type LayoutPage struct {
	TemplateID    string          `json:"layout_template_id"`
	EnablePolling *bool           `json:"enable_polling,omitempty"`
	Polling       *PollingSetting `json:"polling_setting,omitempty"`
	Seats         []Seat          `json:"user_seat_list,omitempty"`
}

// Layout is one basic or advanced layout.
type Layout struct {
	LayoutID   string       `json:"layout_id"`
	LayoutName string       `json:"layout_name,omitempty"`
	LayoutType int          `json:"layout_type,omitempty"`
	Pages      []LayoutPage `json:"page_list"`
}

// Layouts reads the current layouts of one meeting.
type Layouts struct {
	SelectedLayoutID string   `json:"selected_layout_id"`
	Layouts          []Layout `json:"layout_list"`
}

// LayoutTemplates returns the official layout templates.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98844
func (c *Meeting) LayoutTemplates(ctx context.Context) ([]LayoutTemplate, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		List []LayoutTemplate `json:"layout_template_list"`
	}
	err := c.api.Call(ctx, "meeting layout templates",
		"/cgi-bin/meeting/layout/list_template", nil, &out)
	return out.List, err
}

// AddLayout creates one basic layout and returns it with the selected id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98845
func (c *Meeting) AddLayout(ctx context.Context, meetingID string, pages []LayoutPage, defaultOrder int) (string, []Layout, error) {
	if err := check(c); err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(pages) == 0 {
		return "", nil, errors.New("wecom: add layout: meetingid and page_list are required")
	}
	request := struct {
		MeetingID    string   `json:"meetingid"`
		LayoutList   []Layout `json:"layout_list"`
		DefaultOrder int      `json:"default_layout_order,omitempty"`
	}{MeetingID: meetingID, LayoutList: []Layout{{Pages: pages}}, DefaultOrder: defaultOrder}
	var out Layouts
	if err := c.api.Call(ctx, "meeting add layout", "/cgi-bin/meeting/layout/add", request, &out); err != nil {
		return "", nil, err
	}
	return out.SelectedLayoutID, out.Layouts, nil
}

// AddAdvancedLayout creates one advanced layout.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98861
func (c *Meeting) AddAdvancedLayout(ctx context.Context, meetingID string, layouts []Layout) ([]Layout, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(layouts) == 0 {
		return nil, errors.New("wecom: add advanced layout: meetingid and layout_list are required")
	}
	request := struct {
		MeetingID  string   `json:"meetingid"`
		LayoutList []Layout `json:"layout_list"`
	}{MeetingID: meetingID, LayoutList: layouts}
	var out Layouts
	err := c.api.Call(ctx, "meeting add advanced layout",
		"/cgi-bin/meeting/advanced_layout/add", request, &out)
	return out.Layouts, err
}

// UpdateLayout replaces the pages of one basic layout.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98846
func (c *Meeting) UpdateLayout(ctx context.Context, meetingID, layoutID string, pages []LayoutPage, setAsDefault bool) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(layoutID) == "" || len(pages) == 0 {
		return errors.New("wecom: update layout: meetingid, layout_id and page_list are required")
	}
	request := struct {
		MeetingID        string       `json:"meetingid"`
		LayoutID         string       `json:"layout_id"`
		Pages            []LayoutPage `json:"page_list"`
		EnableSetDefault bool         `json:"enable_set_default,omitempty"`
	}{MeetingID: meetingID, LayoutID: layoutID, Pages: pages, EnableSetDefault: setAsDefault}
	return c.api.Call(ctx, "meeting update layout", "/cgi-bin/meeting/layout/update", request, nil)
}

// UpdateAdvancedLayout replaces one advanced layout.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98868
func (c *Meeting) UpdateAdvancedLayout(ctx context.Context, meetingID string, layout Layout) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(layout.LayoutID) == "" {
		return errors.New("wecom: update advanced layout: meetingid and layout_id are required")
	}
	request := struct {
		MeetingID  string       `json:"meetingid"`
		LayoutID   string       `json:"layout_id"`
		LayoutName string       `json:"layout_name,omitempty"`
		Pages      []LayoutPage `json:"page_list"`
	}{MeetingID: meetingID, LayoutID: layout.LayoutID, LayoutName: layout.LayoutName, Pages: layout.Pages}
	return c.api.Call(ctx, "meeting update advanced layout",
		"/cgi-bin/meeting/advanced_layout/update", request, nil)
}

// SetDefaultLayout selects the default layout of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98847
func (c *Meeting) SetDefaultLayout(ctx context.Context, meetingID, layoutID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(layoutID) == "" {
		return errors.New("wecom: default layout: meetingid and selected_layout_id are required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		LayoutID  string `json:"selected_layout_id"`
	}{MeetingID: meetingID, LayoutID: layoutID}
	return c.api.Call(ctx, "meeting default layout", "/cgi-bin/meeting/layout/set_default", request, nil)
}

// ApplyAdvancedLayout applies one advanced layout to specific participants.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98869
func (c *Meeting) ApplyAdvancedLayout(ctx context.Context, meetingID, layoutID string, tmpOpenIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(layoutID) == "" || len(tmpOpenIDs) == 0 {
		return errors.New("wecom: apply layout: meetingid, layout_id and user_list are required")
	}
	users := make([]map[string]string, 0, len(tmpOpenIDs))
	for _, id := range tmpOpenIDs {
		users = append(users, map[string]string{"tmp_openid": id})
	}
	request := struct {
		MeetingID string              `json:"meetingid"`
		LayoutID  string              `json:"layout_id"`
		Users     []map[string]string `json:"user_list"`
	}{MeetingID: meetingID, LayoutID: layoutID, Users: users}
	return c.api.Call(ctx, "meeting apply layout",
		"/cgi-bin/meeting/advanced_layout/apply", request, nil)
}

// ListLayouts returns the current layouts of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98862
func (c *Meeting) ListLayouts(ctx context.Context, meetingID string) (Layouts, error) {
	if err := check(c); err != nil {
		return Layouts{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return Layouts{}, errors.New("wecom: list layouts: meetingid is required")
	}
	var out Layouts
	err := c.api.Call(ctx, "meeting layouts", "/cgi-bin/meeting/advanced_layout/list",
		map[string]string{"meetingid": meetingID}, &out)
	return out, err
}

// UserLayout returns the layout a specific participant currently sees.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98865
func (c *Meeting) UserLayout(ctx context.Context, meetingID, tmpOpenID string, instanceID int) (Layout, error) {
	if err := check(c); err != nil {
		return Layout{}, err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(tmpOpenID) == "" {
		return Layout{}, errors.New("wecom: user layout: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID  string `json:"meetingid"`
		TmpOpenID  string `json:"tmp_openid"`
		InstanceID int    `json:"instance_id,omitempty"`
	}{MeetingID: meetingID, TmpOpenID: tmpOpenID, InstanceID: instanceID}
	var out struct {
		LayoutID   string       `json:"selected_layout_id"`
		LayoutName string       `json:"layout_name"`
		LayoutType int          `json:"layout_type"`
		Pages      []LayoutPage `json:"page_list"`
	}
	err := c.api.Call(ctx, "meeting user layout",
		"/cgi-bin/meeting/advanced_layout/get_user_layout", request, &out)
	return Layout{
		LayoutID: out.LayoutID, LayoutName: out.LayoutName,
		LayoutType: out.LayoutType, Pages: out.Pages,
	}, err
}

// BatchDeleteLayouts removes several layouts of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98866
func (c *Meeting) BatchDeleteLayouts(ctx context.Context, meetingID string, layoutIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(layoutIDs) == 0 {
		return errors.New("wecom: delete layouts: meetingid and layout_id_list are required")
	}
	request := struct {
		MeetingID string   `json:"meetingid"`
		LayoutIDs []string `json:"layout_id_list"`
	}{MeetingID: meetingID, LayoutIDs: layoutIDs}
	return c.api.Call(ctx, "meeting delete layouts",
		"/cgi-bin/meeting/advanced_layout/batch_delete", request, nil)
}

// Background is one meeting background image.
type Background struct {
	BackgroundID string `json:"background_id"`
	ImageMD5     string `json:"image_md5,omitempty"`
}

// BackgroundImage is one uploaded background candidate.
type BackgroundImage struct {
	ImageMD5 string `json:"image_md5"`
	ImageURL string `json:"image_url"`
}

// Backgrounds is the background list of one meeting.
type Backgrounds struct {
	SelectedID string       `json:"selected_background_id"`
	List       []Background `json:"background_list"`
}

// AddBackground adds background images to one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98851
func (c *Meeting) AddBackground(ctx context.Context, meetingID string, images []BackgroundImage, defaultOrder int) (Backgrounds, error) {
	if err := check(c); err != nil {
		return Backgrounds{}, err
	}
	if strings.TrimSpace(meetingID) == "" || len(images) == 0 {
		return Backgrounds{}, errors.New("wecom: add background: meetingid and image_list are required")
	}
	request := struct {
		MeetingID    string            `json:"meetingid"`
		Images       []BackgroundImage `json:"image_list"`
		DefaultOrder int               `json:"default_image_order,omitempty"`
	}{MeetingID: meetingID, Images: images, DefaultOrder: defaultOrder}
	var out Backgrounds
	err := c.api.Call(ctx, "meeting add background", "/cgi-bin/meeting/layout/add_background", request, &out)
	return out, err
}

// SetDefaultBackground selects the default background of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98852
func (c *Meeting) SetDefaultBackground(ctx context.Context, meetingID, backgroundID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(backgroundID) == "" {
		return errors.New("wecom: default background: meetingid and selected_background_id are required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		BackgroundID string `json:"selected_background_id"`
	}{MeetingID: meetingID, BackgroundID: backgroundID}
	return c.api.Call(ctx, "meeting default background",
		"/cgi-bin/meeting/layout/set_default_background", request, nil)
}

// Backgrounds returns the backgrounds of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98856
func (c *Meeting) Backgrounds(ctx context.Context, meetingID string) (Backgrounds, error) {
	if err := check(c); err != nil {
		return Backgrounds{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return Backgrounds{}, errors.New("wecom: backgrounds: meetingid is required")
	}
	var out Backgrounds
	err := c.api.Call(ctx, "meeting backgrounds", "/cgi-bin/meeting/layout/list_background",
		map[string]string{"meetingid": meetingID}, &out)
	return out, err
}

// DeleteBackground removes one background from a meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98853
func (c *Meeting) DeleteBackground(ctx context.Context, meetingID, backgroundID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(backgroundID) == "" {
		return errors.New("wecom: delete background: meetingid and background_id are required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		BackgroundID string `json:"background_id"`
	}{MeetingID: meetingID, BackgroundID: backgroundID}
	return c.api.Call(ctx, "meeting delete background",
		"/cgi-bin/meeting/layout/delete_background", request, nil)
}

// BatchDeleteBackgrounds removes several backgrounds from a meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98854
func (c *Meeting) BatchDeleteBackgrounds(ctx context.Context, meetingID string, backgroundIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(backgroundIDs) == 0 {
		return errors.New("wecom: delete backgrounds: meetingid and background_id_list are required")
	}
	request := struct {
		MeetingID     string   `json:"meetingid"`
		BackgroundIDs []string `json:"background_id_list"`
	}{MeetingID: meetingID, BackgroundIDs: backgroundIDs}
	return c.api.Call(ctx, "meeting delete backgrounds",
		"/cgi-bin/meeting/layout/batch_delete_background", request, nil)
}
