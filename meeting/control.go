package meeting

import (
	"context"
	"errors"
	"strings"
)

// OperatedUser identifies one meeting instance of one participant.
type OperatedUser struct {
	TmpOpenID  string `json:"tmp_openid"`
	InstanceID int    `json:"instance_id,omitempty"`
}

// MemberSetting is the in-meeting configuration of one meeting. Pointer booleans
// keep "leave unchanged" distinct from "set to false".
type MemberSetting struct {
	MeetingID         string `json:"meetingid"`
	MuteAll           *bool  `json:"mute_all,omitempty"`
	AllowUnmuteSelf   *bool  `json:"allow_unmute_self,omitempty"`
	EnableEnterMute   int    `json:"enable_enter_mute,omitempty"`
	MeetingLocked     *bool  `json:"meeting_locked,omitempty"`
	HideMeetingCode   *bool  `json:"hide_meeting_code_password,omitempty"`
	AllowChat         int    `json:"allow_chat,omitempty"`
	AllowShareScreen  *bool  `json:"allow_share_screen,omitempty"`
	AllowExternalUser *bool  `json:"allow_external_user,omitempty"`
	PlayIVROnJoin     *bool  `json:"play_ivr_on_join,omitempty"`
	EnableWaitingRoom *bool  `json:"enable_waiting_room,omitempty"`
}

// SetMemberSetting updates the in-meeting configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98175
func (c *Meeting) SetMemberSetting(ctx context.Context, setting MemberSetting) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(setting.MeetingID) == "" {
		return errors.New("wecom: member setting: meetingid is required")
	}
	return c.api.Call(ctx, "meeting member setting", "/cgi-bin/meeting/realcontrol/set", setting, nil)
}

// SetCoHost makes one participant a co-host or revokes the role.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98180
func (c *Meeting) SetCoHost(ctx context.Context, meetingID string, enable bool, user OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(user.TmpOpenID) == "" {
		return errors.New("wecom: co-host: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID    string       `json:"meetingid"`
		Action       bool         `json:"action"`
		OperatedUser OperatedUser `json:"operated_user"`
	}{MeetingID: meetingID, Action: enable, OperatedUser: user}
	return c.api.Call(ctx, "meeting co-host", "/cgi-bin/meeting/realcontrol/set_cohost", request, nil)
}

// MuteUser mutes or unmutes one participant.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98184
func (c *Meeting) MuteUser(ctx context.Context, meetingID string, mute bool, user OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(user.TmpOpenID) == "" {
		return errors.New("wecom: mute user: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID    string       `json:"meetingid"`
		Option       bool         `json:"option"`
		OperatedUser OperatedUser `json:"operated_user"`
	}{MeetingID: meetingID, Option: mute, OperatedUser: user}
	return c.api.Call(ctx, "meeting mute user", "/cgi-bin/meeting/realcontrol/mute_user", request, nil)
}

// SwitchUserVideo turns the camera of one participant on or off.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98189
func (c *Meeting) SwitchUserVideo(ctx context.Context, meetingID string, on bool, user OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(user.TmpOpenID) == "" {
		return errors.New("wecom: user video: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID    string       `json:"meetingid"`
		Video        bool         `json:"video"`
		OperatedUser OperatedUser `json:"operated_user"`
	}{MeetingID: meetingID, Video: on, OperatedUser: user}
	return c.api.Call(ctx, "meeting user video", "/cgi-bin/meeting/realcontrol/switch_user_video", request, nil)
}

// CloseScreenShare stops the screen sharing of one participant.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98185
func (c *Meeting) CloseScreenShare(ctx context.Context, meetingID string, user OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(user.TmpOpenID) == "" {
		return errors.New("wecom: close screen share: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID    string       `json:"meetingid"`
		OperatedUser OperatedUser `json:"operated_user"`
	}{MeetingID: meetingID, OperatedUser: user}
	return c.api.Call(ctx, "meeting close screen share", "/cgi-bin/meeting/realcontrol/close_screen_share", request, nil)
}

// NicknameChange is one participant's new display name.
type NicknameChange struct {
	TmpOpenID  string `json:"tmp_openid"`
	InstanceID int    `json:"instance_id,omitempty"`
	Nickname   string `json:"nickname"`
}

// SetNicknames changes the in-meeting display names.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98188
func (c *Meeting) SetNicknames(ctx context.Context, meetingID string, changes []NicknameChange) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(changes) == 0 {
		return errors.New("wecom: nicknames: meetingid and operated_users are required")
	}
	request := struct {
		MeetingID string           `json:"meetingid"`
		Users     []NicknameChange `json:"operated_users"`
	}{MeetingID: meetingID, Users: changes}
	return c.api.Call(ctx, "meeting nicknames", "/cgi-bin/meeting/realcontrol/set_nicknames", request, nil)
}

// Waiting room operation types.
const (
	// WaitingAdmit lets waiting members into the meeting.
	WaitingAdmit = 1
	// WaitingMove moves members into the waiting room.
	WaitingMove = 2
)

// ManageWaitingRoom admits or moves participants.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98186
func (c *Meeting) ManageWaitingRoom(ctx context.Context, meetingID string, operation int, allowRejoin bool, users []OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(users) == 0 {
		return errors.New("wecom: waiting room manage: meetingid and operated_users are required")
	}
	request := struct {
		MeetingID     string         `json:"meetingid"`
		OperateType   int            `json:"operate_type"`
		AllowRejoin   bool           `json:"allow_rejoin,omitempty"`
		OperatedUsers []OperatedUser `json:"operated_users"`
	}{MeetingID: meetingID, OperateType: operation, AllowRejoin: allowRejoin, OperatedUsers: users}
	return c.api.Call(ctx, "meeting waiting room manage",
		"/cgi-bin/meeting/realcontrol/manage_waiting_room_users", request, nil)
}

// KickoutUsers removes participants from the meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98181
func (c *Meeting) KickoutUsers(ctx context.Context, meetingID string, allowRejoin bool, users []OperatedUser) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(users) == 0 {
		return errors.New("wecom: kickout: meetingid and operated_users are required")
	}
	request := struct {
		MeetingID     string         `json:"meetingid"`
		AllowRejoin   bool           `json:"allow_rejoin,omitempty"`
		OperatedUsers []OperatedUser `json:"operated_users"`
	}{MeetingID: meetingID, AllowRejoin: allowRejoin, OperatedUsers: users}
	return c.api.Call(ctx, "meeting kickout", "/cgi-bin/meeting/realcontrol/kickout_users", request, nil)
}

// DismissOptions controls how a meeting ends.
type DismissOptions struct {
	ForceDismiss bool `json:"force_dismiss,omitempty"`
	RetrieveCode bool `json:"retrieve_code,omitempty"`
}

// Dismiss ends one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98187
func (c *Meeting) Dismiss(ctx context.Context, meetingID string, options DismissOptions) error {
	if err := check(c); err != nil {
		return err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return errors.New("wecom: dismiss: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		DismissOptions
	}{MeetingID: meetingID, DismissOptions: options}
	return c.api.Call(ctx, "meeting dismiss", "/cgi-bin/meeting/realcontrol/dismiss", request, nil)
}
