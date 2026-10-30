// Package meeting implements the WeCom meeting API (/cgi-bin/meeting): scheduled
// meetings, invitees, customer short links and meeting statistics.
package meeting

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Meeting manages scheduled meetings with one shared transport.
type Meeting struct{ api *client.Client }

// New creates a meeting client.
func New(api *client.Client) *Meeting { return &Meeting{api: api} }

// UserList is a set of member ids, used by invitees, hosts and ring users.
type UserList struct {
	// UserID has no omitempty on purpose: a nil list means "unchanged" at the
	// enclosing pointer, while an empty list sent here clears the members.
	UserID []string `json:"userid"`
}

// Guest is an external participant invited by phone.
type Guest struct {
	Area        string `json:"area"`
	PhoneNumber string `json:"phone_number"`
	GuestName   string `json:"guest_name,omitempty"`
}

// Reminders is the repeat and reminder configuration of one meeting. It is a
// nested object in both the create and the update body, not a per-member list.
type Reminders struct {
	IsRepeat         int   `json:"is_repeat,omitempty"`
	IsCustomRepeat   int   `json:"is_custom_repeat,omitempty"`
	RepeatType       int   `json:"repeat_type,omitempty"`
	RepeatUntilType  int   `json:"repeat_until_type,omitempty"`
	RepeatUntilCount int   `json:"repeat_until_count,omitempty"`
	RepeatUntil      int64 `json:"repeat_until,omitempty"`
	RepeatInterval   int   `json:"repeat_interval,omitempty"`
	RepeatDayOfWeek  []int `json:"repeat_day_of_week,omitempty"`
	RepeatDayOfMonth []int `json:"repeat_day_of_month,omitempty"`
	// RemindBefore lists the offsets in seconds before the meeting starts, for
	// example 0, 300, 900, 3600 or 86400.
	RemindBefore []int `json:"remind_before,omitempty"`
}

// Settings is the meeting configuration. Tri-state flags use pointers so an
// update can turn a feature off explicitly.
type Settings struct {
	Password                  string    `json:"password,omitempty"`
	EnableWaitingRoom         *bool     `json:"enable_waiting_room,omitempty"`
	AllowEnterBeforeHost      *bool     `json:"allow_enter_before_host,omitempty"`
	RemindScope               *int      `json:"remind_scope,omitempty"`
	EnableEnterMute           *int      `json:"enable_enter_mute,omitempty"`
	AllowUnmuteSelf           *bool     `json:"allow_unmute_self,omitempty"`
	MuteAll                   *bool     `json:"mute_all,omitempty"`
	AllowExternalUser         *bool     `json:"allow_external_user,omitempty"`
	EnableScreenWatermark     *bool     `json:"enable_screen_watermark,omitempty"`
	WatermarkType             *int      `json:"watermark_type,omitempty"`
	AutoRecordType            *string   `json:"auto_record_type,omitempty"`
	AttendeeJoinAutoRecord    *bool     `json:"attendee_join_auto_record,omitempty"`
	EnableHostPauseAutoRecord *bool     `json:"enable_host_pause_auto_record,omitempty"`
	EnableInterpreter         *bool     `json:"enable_interpreter,omitempty"`
	EnableEnroll              *bool     `json:"enable_enroll,omitempty"`
	EnableHostKey             *bool     `json:"enable_host_key,omitempty"`
	HostKey                   *string   `json:"host_key,omitempty"`
	Hosts                     *UserList `json:"hosts,omitempty"`
	RingUsers                 *UserList `json:"ring_users,omitempty"`
}

// CreateRequest schedules one meeting.
type CreateRequest struct {
	AdminUserID     string    `json:"admin_userid"`
	Title           string    `json:"title"`
	MeetingStart    int64     `json:"meeting_start"`
	MeetingDuration int       `json:"meeting_duration"`
	Description     string    `json:"description,omitempty"`
	Location        string    `json:"location,omitempty"`
	AgentID         int       `json:"agentid,omitempty"`
	CalID           string    `json:"cal_id,omitempty"`
	Invitees        *UserList `json:"invitees,omitempty"`
	Guests          []Guest   `json:"guests,omitempty"`
	Settings        *Settings `json:"settings,omitempty"`
	// Reminders carries the repeat configuration; the repeat fields live inside
	// this block, not at the top level of the request.
	Reminders *Reminders `json:"reminders,omitempty"`
}

// Create schedules one meeting and returns its id plus the invitees that could
// not be added.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99104
// 文档: https://developer.work.weixin.qq.com/document/path/98148
func (c *Meeting) Create(ctx context.Context, request CreateRequest) (string, []string, error) {
	if err := check(c); err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(request.AdminUserID) == "" || strings.TrimSpace(request.Title) == "" {
		return "", nil, errors.New("wecom: create meeting: admin_userid and title are required")
	}
	if request.MeetingStart <= 0 || request.MeetingDuration <= 0 {
		return "", nil, errors.New("wecom: create meeting: meeting_start and meeting_duration are required")
	}
	var out struct {
		MeetingID   string   `json:"meetingid"`
		ExcessUsers []string `json:"excess_users"`
	}
	err := c.api.Call(ctx, "meeting create", "/cgi-bin/meeting/create", request, &out)
	return out.MeetingID, out.ExcessUsers, err
}

// UpdateRequest changes one scheduled meeting. Only the fields set here are
// sent, so a nil pointer or nil slice leaves the current value alone; an empty
// slice clears the collection. MeetingStart and MeetingDuration must be set
// together.
type UpdateRequest struct {
	MeetingID       string
	Title           *string
	MeetingStart    *int64
	MeetingDuration *int
	Description     *string
	Location        *string
	Invitees        []string
	Settings        *Settings
	Reminders       *Reminders
	AgentID         *int
	CalID           *string
	// RemindTime is the number of seconds before the meeting to remind members.
	RemindTime *int
}

// MarshalJSON builds the sparse update body. encoding/json drops struct zero
// values together with omitempty, which made "clear this field" and "leave it
// alone" indistinguishable, so the body is assembled field by field.
func (r UpdateRequest) MarshalJSON() ([]byte, error) {
	body := map[string]any{"meetingid": r.MeetingID}
	if r.Title != nil {
		body["title"] = *r.Title
	}
	if r.MeetingStart != nil {
		body["meeting_start"] = *r.MeetingStart
	}
	if r.MeetingDuration != nil {
		body["meeting_duration"] = *r.MeetingDuration
	}
	if r.Description != nil {
		body["description"] = *r.Description
	}
	if r.Location != nil {
		body["location"] = *r.Location
	}
	if r.Invitees != nil {
		body["invitees"] = map[string]any{"userid": r.Invitees}
	}
	if r.Settings != nil {
		body["settings"] = r.Settings
	}
	if r.Reminders != nil {
		body["reminders"] = r.Reminders
	}
	if r.AgentID != nil {
		body["agentid"] = *r.AgentID
	}
	if r.CalID != nil {
		body["cal_id"] = *r.CalID
	}
	if r.RemindTime != nil {
		body["remind_time"] = *r.RemindTime
	}
	return json.Marshal(body)
}

// Update changes one scheduled meeting and returns the invitees that have no
// valid meeting account. That list is only reported for some paid editions, so
// an empty result is normal.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99047
// 文档: https://developer.work.weixin.qq.com/document/path/98154
func (c *Meeting) Update(ctx context.Context, request UpdateRequest) ([]string, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(request.MeetingID) == "" {
		return nil, errors.New("wecom: update meeting: meetingid is required")
	}
	if (request.MeetingStart == nil) != (request.MeetingDuration == nil) {
		return nil, errors.New("wecom: update meeting: meeting_start and meeting_duration must be set together")
	}
	var out struct {
		ExcessUsers []string `json:"excess_users"`
	}
	if err := c.api.Call(ctx, "meeting update", "/cgi-bin/meeting/update", request, &out); err != nil {
		return nil, err
	}
	return out.ExcessUsers, nil
}

// Cancel cancels one scheduled meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99048
// 文档: https://developer.work.weixin.qq.com/document/path/98153
func (c *Meeting) Cancel(ctx context.Context, meetingID string) error {
	if err := check(c); err != nil {
		return err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return errors.New("wecom: cancel meeting: meetingid is required")
	}
	return c.api.Call(ctx, "meeting cancel", "/cgi-bin/meeting/cancel",
		map[string]string{"meetingid": meetingID}, nil)
}

// Info is one meeting. Nested blocks that vary by meeting type stay raw.
type Info struct {
	MeetingID       string          `json:"meetingid"`
	Title           string          `json:"title"`
	MeetingStart    int64           `json:"meeting_start"`
	MeetingDuration int             `json:"meeting_duration"`
	Status          int             `json:"status,omitempty"`
	Description     string          `json:"description,omitempty"`
	Location        string          `json:"location,omitempty"`
	AdminUserID     string          `json:"admin_userid,omitempty"`
	Invitees        *UserList       `json:"invitees,omitempty"`
	Settings        json.RawMessage `json:"settings,omitempty"`
}

// GetInfo reads one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99049
// 文档: https://developer.work.weixin.qq.com/document/path/98149
func (c *Meeting) GetInfo(ctx context.Context, meetingID string) (Info, error) {
	if err := check(c); err != nil {
		return Info{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return Info{}, errors.New("wecom: meeting info: meetingid is required")
	}
	var out struct {
		Info Info `json:"meeting_info"`
	}
	err := c.api.Call(ctx, "meeting info", "/cgi-bin/meeting/get_info",
		map[string]string{"meetingid": meetingID}, &out)
	return out.Info, err
}

// UserMeetingIDs returns one page of meeting ids created by one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99050
// 文档: https://developer.work.weixin.qq.com/document/path/98714
func (c *Meeting) UserMeetingIDs(ctx context.Context, userID, cursor string, limit int) ([]string, string, error) {
	if err := check(c); err != nil {
		return nil, "", err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, "", errors.New("wecom: user meeting ids: userid is required")
	}
	request := struct {
		UserID string `json:"userid"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{UserID: userID, Cursor: cursor, Limit: limit}
	var out struct {
		List   []string `json:"meetingid_list"`
		Cursor string   `json:"next_cursor"`
	}
	err := c.api.Call(ctx, "meeting user ids", "/cgi-bin/meeting/get_user_meetingid", request, &out)
	return out.List, out.Cursor, err
}

// Invitees returns one page of invited members of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98160
func (c *Meeting) Invitees(ctx context.Context, meetingID, cursor string) ([]string, string, error) {
	if err := check(c); err != nil {
		return nil, "", err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return nil, "", errors.New("wecom: invitees: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Cursor    string `json:"cursor,omitempty"`
	}{MeetingID: meetingID, Cursor: cursor}
	var out struct {
		List   []string `json:"invitees"`
		Cursor string   `json:"next_cursor"`
	}
	err := c.api.Call(ctx, "meeting invitees", "/cgi-bin/meeting/get_invitees", request, &out)
	return out.List, out.Cursor, err
}

// SetInvitees replaces the invited members of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98162
func (c *Meeting) SetInvitees(ctx context.Context, meetingID string, userIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(userIDs) == 0 {
		return errors.New("wecom: set invitees: meetingid and userid are required")
	}
	request := struct {
		MeetingID string   `json:"meetingid"`
		UserIDs   []string `json:"userid"`
	}{MeetingID: meetingID, UserIDs: userIDs}
	return c.api.Call(ctx, "meeting set invitees", "/cgi-bin/meeting/set_invitees", request, nil)
}

// StartRecord is one meeting that started in the queried window.
type StartRecord struct {
	MeetingID string `json:"meetingid"`
	Title     string `json:"title,omitempty"`
	StartTime int64  `json:"meeting_start,omitempty"`
	Status    int    `json:"status,omitempty"`
}

// StartRecords returns the meetings started in one window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99651
func (c *Meeting) StartRecords(ctx context.Context, beginTime, endTime int64, cursor string, limit int) ([]StartRecord, string, error) {
	if err := check(c); err != nil {
		return nil, "", err
	}
	if beginTime <= 0 || endTime <= beginTime {
		return nil, "", errors.New("wecom: start records: a valid time window is required")
	}
	request := struct {
		BeginTime int64  `json:"begin_time"`
		EndTime   int64  `json:"end_time"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{BeginTime: beginTime, EndTime: endTime, Cursor: cursor, Limit: limit}
	var out struct {
		List   []StartRecord `json:"meeting_list"`
		Cursor string        `json:"next_cursor"`
	}
	err := c.api.Call(ctx, "meeting start records", "/cgi-bin/meeting/statistics/get_start_list", request, &out)
	return out.List, out.Cursor, err
}

// CreateCustomerShortURL creates the per-customer short URL of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98818
func (c *Meeting) CreateCustomerShortURL(ctx context.Context, meetingID, customerData string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(customerData) == "" {
		return "", errors.New("wecom: customer short url: meetingid and customer_data are required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		CustomerData string `json:"customer_data"`
	}{MeetingID: meetingID, CustomerData: customerData}
	var out struct {
		URL string `json:"short_url"`
	}
	err := c.api.Call(ctx, "meeting customer short url", "/cgi-bin/meeting/create_customer_short_url", request, &out)
	return out.URL, err
}

// GetCustomerShortURL returns the per-customer short URL of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98819
func (c *Meeting) GetCustomerShortURL(ctx context.Context, meetingID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return "", errors.New("wecom: get customer short url: meetingid is required")
	}
	var out struct {
		URL string `json:"short_url"`
	}
	err := c.api.Call(ctx, "meeting get customer short url", "/cgi-bin/meeting/get_customer_short_url",
		map[string]string{"meetingid": meetingID}, &out)
	return out.URL, err
}

func check(c *Meeting) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: meeting client is unavailable")
	}
	return nil
}
