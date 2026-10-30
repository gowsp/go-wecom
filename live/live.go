// Package live implements the WeCom live streaming API (/cgi-bin/living):
// creating, editing and cancelling reservations, reading details, watch
// statistics and share codes.
package live

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Live manages live streams with one shared transport.
type Live struct{ api *client.Client }

// New creates a live streaming client.
func New(api *client.Client) *Live { return &Live{api: api} }

// Live types accepted by Type.
const (
	TypeGeneral  = 0
	TypeDisplay  = 1
	TypeActivity = 2
	TypeShow     = 3
	TypeTraining = 4
)

// ActivityDetail carries the extra fields of an activity-type live stream.
type ActivityDetail struct {
	Description string   `json:"description,omitempty"`
	ImageList   []string `json:"image_list,omitempty"`
}

// CreateRequest reserves one live stream.
type CreateRequest struct {
	AnchorUserID         string          `json:"anchor_userid"`
	Theme                string          `json:"theme"`
	LivingStart          int64           `json:"living_start"`
	LivingDuration       int             `json:"living_duration"`
	Description          string          `json:"description,omitempty"`
	Type                 int             `json:"type,omitempty"`
	AgentID              int             `json:"agentid,omitempty"`
	RemindTime           int             `json:"remind_time,omitempty"`
	ActivityCoverMediaID string          `json:"activity_cover_mediaid,omitempty"`
	ActivityShareMediaID string          `json:"activity_share_mediaid,omitempty"`
	ActivityDetail       *ActivityDetail `json:"activity_detail,omitempty"`
}

// ModifyRequest changes one reserved live stream.
type ModifyRequest struct {
	LivingID       string `json:"livingid"`
	Theme          string `json:"theme,omitempty"`
	LivingStart    int64  `json:"living_start,omitempty"`
	LivingDuration int    `json:"living_duration,omitempty"`
	Description    string `json:"description,omitempty"`
	Type           int    `json:"type,omitempty"`
	RemindTime     int    `json:"remind_time,omitempty"`
}

// Info is the detail of one live stream.
type Info struct {
	Theme           string `json:"theme"`
	LivingStart     int64  `json:"living_start"`
	LivingDuration  int    `json:"living_duration"`
	Status          int    `json:"status"`
	ReserveStart    int64  `json:"reserve_start,omitempty"`
	ReserveDuration int    `json:"reserve_living_duration,omitempty"`
	Description     string `json:"description,omitempty"`
	AnchorUserID    string `json:"anchor_userid"`
	MainDepartment  int    `json:"main_department,omitempty"`
	ViewerNum       int    `json:"viewer_num,omitempty"`
	CommentNum      int    `json:"comment_num,omitempty"`
	MicNum          int    `json:"mic_num,omitempty"`
	OpenReplay      int    `json:"open_replay,omitempty"`
	ReplayStatus    int    `json:"replay_status,omitempty"`
	Type            int    `json:"type,omitempty"`
}

// WatchStat is one viewer's watch record.
type WatchUser struct {
	UserID    string `json:"userid"`
	WatchTime int    `json:"watch_time"`
	IsComment int    `json:"is_comment,omitempty"`
	IsMic     int    `json:"is_mic,omitempty"`
}

// WatchExternalUser is one external viewer's watch record.
type WatchExternalUser struct {
	ExternalUserID string `json:"external_userid"`
	Type           int    `json:"type,omitempty"`
	Name           string `json:"name,omitempty"`
	WatchTime      int    `json:"watch_time"`
	IsComment      int    `json:"is_comment,omitempty"`
	IsMic          int    `json:"is_mic,omitempty"`
}

// WatchStat is one page of watch statistics.
type WatchStat struct {
	Ending        int                 `json:"ending"`
	NextKey       string              `json:"next_key"`
	Users         []WatchUser         `json:"users"`
	ExternalUsers []WatchExternalUser `json:"external_users"`
}

// Create reserves one live stream and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93637
func (c *Live) Create(ctx context.Context, request CreateRequest) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(request.AnchorUserID) == "" || strings.TrimSpace(request.Theme) == "" {
		return "", errors.New("wecom: live create: anchor_userid and theme are required")
	}
	if request.LivingStart <= 0 || request.LivingDuration <= 0 {
		return "", errors.New("wecom: live create: living_start and living_duration are required")
	}
	var out struct {
		LivingID string `json:"livingid"`
	}
	err := c.api.Call(ctx, "live create", "/cgi-bin/living/create", request, &out)
	return out.LivingID, err
}

// Modify changes one reserved live stream.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93640
func (c *Live) Modify(ctx context.Context, request ModifyRequest) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(request.LivingID) == "" {
		return errors.New("wecom: live modify: livingid is required")
	}
	return c.api.Call(ctx, "live modify", "/cgi-bin/living/modify", request, nil)
}

// Cancel cancels one reserved live stream.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93638
func (c *Live) Cancel(ctx context.Context, livingID string) error {
	if err := check(c); err != nil {
		return err
	}
	livingID = strings.TrimSpace(livingID)
	if livingID == "" {
		return errors.New("wecom: live cancel: livingid is required")
	}
	return c.api.Call(ctx, "live cancel", "/cgi-bin/living/cancel",
		map[string]string{"livingid": livingID}, nil)
}

// DeleteReplay removes the replay of one live stream.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93874
func (c *Live) DeleteReplay(ctx context.Context, livingID string) error {
	if err := check(c); err != nil {
		return err
	}
	livingID = strings.TrimSpace(livingID)
	if livingID == "" {
		return errors.New("wecom: live delete replay: livingid is required")
	}
	return c.api.Call(ctx, "live delete replay", "/cgi-bin/living/delete_replay_data",
		map[string]string{"livingid": livingID}, nil)
}

// ShareCode returns the code used to open one live stream in WeChat.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93641
func (c *Live) ShareCode(ctx context.Context, livingID, openID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(livingID) == "" || strings.TrimSpace(openID) == "" {
		return "", errors.New("wecom: live share code: livingid and openid are required")
	}
	request := struct {
		LivingID string `json:"livingid"`
		OpenID   string `json:"openid"`
	}{LivingID: livingID, OpenID: openID}
	var out struct {
		Code string `json:"living_code"`
	}
	err := c.api.Call(ctx, "live share code", "/cgi-bin/living/get_living_code", request, &out)
	return out.Code, err
}

// LivingIDPage is one page of live ids.
type LivingIDPage struct {
	NextCursor string   `json:"next_cursor"`
	LivingIDs  []string `json:"livingid_list"`
}

// ListLivingIDs returns the live streams of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93634
func (c *Live) ListLivingIDs(ctx context.Context, userID, cursor string, limit int) (LivingIDPage, error) {
	if err := check(c); err != nil {
		return LivingIDPage{}, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return LivingIDPage{}, errors.New("wecom: live ids: userid is required")
	}
	request := struct {
		UserID string `json:"userid"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{UserID: userID, Cursor: cursor, Limit: limit}
	var out LivingIDPage
	err := c.api.Call(ctx, "live ids", "/cgi-bin/living/get_user_all_livingid", request, &out)
	return out, err
}

// Info returns the detail of one live stream.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93635
func (c *Live) Info(ctx context.Context, livingID string) (Info, error) {
	if err := check(c); err != nil {
		return Info{}, err
	}
	livingID = strings.TrimSpace(livingID)
	if livingID == "" {
		return Info{}, errors.New("wecom: live info: livingid is required")
	}
	var out struct {
		Info Info `json:"living_info"`
	}
	err := c.api.Call(ctx, "live info", "/cgi-bin/living/get_living_info",
		map[string]string{"livingid": livingID}, &out)
	return out.Info, err
}

// WatchStats returns one page of watch records for a live stream.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93636
func (c *Live) WatchStats(ctx context.Context, livingID, nextKey string) (WatchStat, error) {
	if err := check(c); err != nil {
		return WatchStat{}, err
	}
	livingID = strings.TrimSpace(livingID)
	if livingID == "" {
		return WatchStat{}, errors.New("wecom: watch stats: livingid is required")
	}
	request := struct {
		LivingID string `json:"livingid"`
		NextKey  string `json:"next_key,omitempty"`
	}{LivingID: livingID, NextKey: nextKey}
	var out struct {
		Ending  int    `json:"ending"`
		NextKey string `json:"next_key"`
		Stat    struct {
			Users         []WatchUser         `json:"users"`
			ExternalUsers []WatchExternalUser `json:"external_users"`
		} `json:"stat_info"`
	}
	if err := c.api.Call(ctx, "live watch stats", "/cgi-bin/living/get_watch_stat", request, &out); err != nil {
		return WatchStat{}, err
	}
	return WatchStat{
		Ending: out.Ending, NextKey: out.NextKey,
		Users: out.Stat.Users, ExternalUsers: out.Stat.ExternalUsers,
	}, nil
}

// ShareInfo is the viewer resolved from a mall share code.
type ShareInfo struct {
	LivingID              string `json:"livingid"`
	ViewerUserID          string `json:"viewer_userid"`
	ViewerExternalUserID  string `json:"viewer_external_userid"`
	InvitorUserID         string `json:"invitor_userid"`
	InvitorExternalUserID string `json:"invitor_external_userid"`
}

// ShareInfo resolves the viewer behind a mini program mall share code.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94442
func (c *Live) ShareInfo(ctx context.Context, shareCode string) (ShareInfo, error) {
	if err := check(c); err != nil {
		return ShareInfo{}, err
	}
	shareCode = strings.TrimSpace(shareCode)
	if shareCode == "" {
		return ShareInfo{}, errors.New("wecom: live share info: ww_share_code is required")
	}
	var out ShareInfo
	err := c.api.Call(ctx, "live share info", "/cgi-bin/living/get_living_share_info",
		map[string]string{"ww_share_code": shareCode}, &out)
	return out, err
}

// Event is one decrypted live status callback.
type Event struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	LivingID   string `xml:"LivingId"`
	Status     int    `xml:"Status"`
	AgentID    int    `xml:"AgentID"`
}

// ParseEvent decodes a decrypted living_status_change callback. Decrypt the body
// with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94145
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: live event: " + err.Error())
	}
	if event.Name != "living_status_change" || strings.TrimSpace(event.LivingID) == "" {
		return Event{}, errors.New("wecom: live event: expected living_status_change with a living id")
	}
	return event, nil
}

func check(c *Live) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: live client is unavailable")
	}
	return nil
}
