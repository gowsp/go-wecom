// Package schedule implements the WeCom calendar and schedule API
// (/cgi-bin/oa/calendar, /cgi-bin/oa/schedule).
package schedule

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Schedule manages calendars and schedule entries with one shared transport.
type Schedule struct{ api *client.Client }

// New creates a calendar client.
func New(api *client.Client) *Schedule { return &Schedule{api: api} }

// Share grants one member access to a calendar. Permission 1 reads, 3 manages.
type Share struct {
	UserID     string `json:"userid"`
	Permission int    `json:"permission"`
}

// ShareFailure reports one share that could not be granted.
type ShareFailure struct {
	UserID  string `json:"userid"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// PublicRange limits who can see a public calendar.
type PublicRange struct {
	UserIDs  []string `json:"userids,omitempty"`
	PartyIDs []int    `json:"partyids,omitempty"`
}

// Calendar is one calendar.
type Calendar struct {
	CalID          string       `json:"cal_id,omitempty"`
	Admins         []string     `json:"admins,omitempty"`
	SetAsDefault   int          `json:"set_as_default,omitempty"`
	Summary        string       `json:"summary,omitempty"`
	Color          string       `json:"color,omitempty"`
	Description    string       `json:"description,omitempty"`
	Shares         []Share      `json:"shares,omitempty"`
	IsPublic       int          `json:"is_public,omitempty"`
	PublicRange    *PublicRange `json:"public_range,omitempty"`
	IsCorpCalendar int          `json:"is_corp_calendar,omitempty"`
}

// CalendarResult is the result of creating a calendar.
type CalendarResult struct {
	CalID      string `json:"cal_id"`
	FailResult struct {
		Shares []ShareFailure `json:"shares"`
	} `json:"fail_result"`
}

// AddCalendar creates one calendar and returns its id plus failed shares.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93647
func (c *Schedule) AddCalendar(ctx context.Context, calendar Calendar, agentID int) (CalendarResult, error) {
	if err := check(c); err != nil {
		return CalendarResult{}, err
	}
	if strings.TrimSpace(calendar.Summary) == "" {
		return CalendarResult{}, errors.New("wecom: add calendar: summary is required")
	}
	request := struct {
		Calendar Calendar `json:"calendar"`
		AgentID  int      `json:"agentid,omitempty"`
	}{Calendar: calendar, AgentID: agentID}
	var out CalendarResult
	err := c.api.Call(ctx, "schedule add calendar", "/cgi-bin/oa/calendar/add", request, &out)
	return out, err
}

// CalendarUpdate is the overwrite body of one calendar update. The endpoint is a
// replacement, not a patch: a field left at its zero value is cleared on the
// server, so set everything the calendar should keep. Summary and Color are
// required; IsPublic and IsCorpCalendar cannot be changed after creation and are
// therefore absent here.
type CalendarUpdate struct {
	CalID       string
	Admins      []string
	Summary     string
	Color       string
	Description string
	Shares      []Share
	PublicRange *PublicRange
	// SkipPublicRange leaves the subscribable range untouched instead of
	// replacing it with PublicRange.
	SkipPublicRange bool
}

func (u CalendarUpdate) request() map[string]any {
	calendar := map[string]any{
		"cal_id":  u.CalID,
		"summary": u.Summary,
		"color":   u.Color,
	}
	if u.Admins != nil {
		calendar["admins"] = u.Admins
	}
	if u.Description != "" {
		calendar["description"] = u.Description
	}
	if u.Shares != nil {
		calendar["shares"] = u.Shares
	}
	if u.PublicRange != nil {
		calendar["public_range"] = map[string]any{
			"userids":  u.PublicRange.UserIDs,
			"partyids": u.PublicRange.PartyIDs,
		}
	}
	skip := 0
	if u.SkipPublicRange {
		skip = 1
	}
	return map[string]any{"skip_public_range": skip, "calendar": calendar}
}

// UpdateCalendar overwrites one calendar. Fields the update leaves empty are
// cleared, so read the calendar first if only part of it should change.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97716
func (c *Schedule) UpdateCalendar(ctx context.Context, update CalendarUpdate) (CalendarResult, error) {
	if err := check(c); err != nil {
		return CalendarResult{}, err
	}
	update.CalID = strings.TrimSpace(update.CalID)
	if update.CalID == "" {
		return CalendarResult{}, errors.New("wecom: update calendar: cal_id is required")
	}
	if strings.TrimSpace(update.Summary) == "" || strings.TrimSpace(update.Color) == "" {
		return CalendarResult{}, errors.New("wecom: update calendar: summary and color are required")
	}
	var out CalendarResult
	err := c.api.Call(ctx, "schedule update calendar", "/cgi-bin/oa/calendar/update", update.request(), &out)
	return out, err
}

// GetCalendars reads one or more calendars by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97717
func (c *Schedule) GetCalendars(ctx context.Context, calIDs []string) ([]Calendar, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(calIDs) == 0 {
		return nil, errors.New("wecom: get calendars: cal_id is empty")
	}
	var out struct {
		List []Calendar `json:"calendar_list"`
	}
	err := c.api.Call(ctx, "schedule get calendars", "/cgi-bin/oa/calendar/get",
		map[string][]string{"cal_id_list": calIDs}, &out)
	return out.List, err
}

// DeleteCalendar removes one calendar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97718
func (c *Schedule) DeleteCalendar(ctx context.Context, calID string) error {
	if err := check(c); err != nil {
		return err
	}
	calID = strings.TrimSpace(calID)
	if calID == "" {
		return errors.New("wecom: delete calendar: cal_id is required")
	}
	return c.api.Call(ctx, "schedule delete calendar", "/cgi-bin/oa/calendar/del",
		map[string]string{"cal_id": calID}, nil)
}

// Attendee is one schedule participant. ResponseStatus is 0 pending, 1 accepted,
// 2 declined, 3 tentative.
type Attendee struct {
	UserID         string `json:"userid"`
	ResponseStatus int    `json:"response_status,omitempty"`
}

// Reminders is the reminder and repeat configuration of a schedule.
type Reminders struct {
	IsRemind             int   `json:"is_remind,omitempty"`
	RemindBeforeEventSec int   `json:"remind_before_event_secs,omitempty"`
	RemindTimeDiffs      []int `json:"remind_time_diffs,omitempty"`
	IsRepeat             int   `json:"is_repeat,omitempty"`
	RepeatType           int   `json:"repeat_type,omitempty"`
	RepeatUntil          int64 `json:"repeat_until,omitempty"`
	IsCustomRepeat       int   `json:"is_custom_repeat,omitempty"`
	RepeatInterval       int   `json:"repeat_interval,omitempty"`
	RepeatDayOfWeek      []int `json:"repeat_day_of_week,omitempty"`
	RepeatDayOfMonth     []int `json:"repeat_day_of_month,omitempty"`
	Timezone             int   `json:"timezone,omitempty"`
}

// Entry is one calendar schedule.
type Entry struct {
	ScheduleID  string     `json:"schedule_id,omitempty"`
	Sequence    int        `json:"sequence,omitempty"`
	Admins      []string   `json:"admins,omitempty"`
	Attendees   []Attendee `json:"attendees,omitempty"`
	Summary     string     `json:"summary,omitempty"`
	Description string     `json:"description,omitempty"`
	StartTime   int64      `json:"start_time,omitempty"`
	EndTime     int64      `json:"end_time,omitempty"`
	IsWholeDay  int        `json:"is_whole_day,omitempty"`
	Reminders   *Reminders `json:"reminders,omitempty"`
	Location    string     `json:"location,omitempty"`
	CalID       string     `json:"cal_id,omitempty"`
}

// AddSchedule creates one schedule on a calendar and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93648
func (c *Schedule) AddSchedule(ctx context.Context, entry Entry, agentID int) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if entry.StartTime <= 0 || entry.EndTime <= entry.StartTime {
		return "", errors.New("wecom: add schedule: a valid time window is required")
	}
	if strings.TrimSpace(entry.Summary) == "" {
		return "", errors.New("wecom: add schedule: summary is required")
	}
	request := struct {
		Schedule Entry `json:"schedule"`
		AgentID  int   `json:"agentid,omitempty"`
	}{Schedule: entry, AgentID: agentID}
	var out struct {
		ScheduleID string `json:"schedule_id"`
	}
	err := c.api.Call(ctx, "schedule add", "/cgi-bin/oa/schedule/add", request, &out)
	return out.ScheduleID, err
}

// Update modes for repeating schedules.
const (
	// OpModeAll changes every occurrence.
	OpModeAll = 0
	// OpModeSingle changes one occurrence.
	OpModeSingle = 1
	// OpModeThisAndLater changes the occurrence and all later ones.
	OpModeThisAndLater = 2
)

// EntryUpdate is the overwrite body of one schedule update. The endpoint is a
// replacement, not a patch: fields left at their zero value are cleared on the
// server. StartTime and EndTime are required. CalID cannot be changed and is
// therefore absent here; use AddAttendees and DeleteAttendees to change the
// participants of one schedule incrementally.
type EntryUpdate struct {
	ScheduleID  string
	Admins      []string
	Attendees   []Attendee
	Summary     string
	Description string
	StartTime   int64
	EndTime     int64
	IsWholeDay  int
	Reminders   *Reminders
	Location    string

	// SkipAttendees keeps the current participants, which is the only way to
	// leave them alone because the endpoint replaces the list otherwise.
	SkipAttendees bool
	// OpMode and OpStartTime choose which occurrences of a repeating schedule
	// change: OpModeAll, OpModeSingle or OpModeThisAndLater.
	OpMode      int
	OpStartTime int64
}

func (u EntryUpdate) request() map[string]any {
	schedule := map[string]any{
		"schedule_id": u.ScheduleID,
		"start_time":  u.StartTime,
		"end_time":    u.EndTime,
	}
	if u.Admins != nil {
		schedule["admins"] = u.Admins
	}
	if u.Attendees != nil {
		schedule["attendees"] = u.Attendees
	}
	if u.Summary != "" {
		schedule["summary"] = u.Summary
	}
	if u.Description != "" {
		schedule["description"] = u.Description
	}
	if u.IsWholeDay != 0 {
		schedule["is_whole_day"] = u.IsWholeDay
	}
	if u.Reminders != nil {
		schedule["reminders"] = u.Reminders
	}
	if u.Location != "" {
		schedule["location"] = u.Location
	}
	body := map[string]any{"schedule": schedule}
	if u.SkipAttendees {
		body["skip_attendees"] = 1
	}
	if u.OpMode != 0 {
		body["op_mode"] = u.OpMode
	}
	if u.OpStartTime != 0 {
		body["op_start_time"] = u.OpStartTime
	}
	return body
}

// UpdateSchedule overwrites one schedule and returns the resulting schedule id.
// The id differs from the input when a repeating schedule is split by OpMode.
// Fields the update leaves empty are cleared, so read the schedule first if only
// part of it should change.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97720
// 文档: https://developer.work.weixin.qq.com/document/path/96204
func (c *Schedule) UpdateSchedule(ctx context.Context, update EntryUpdate) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	update.ScheduleID = strings.TrimSpace(update.ScheduleID)
	if update.ScheduleID == "" {
		return "", errors.New("wecom: update schedule: schedule_id is required")
	}
	if update.StartTime <= 0 || update.EndTime <= update.StartTime {
		return "", errors.New("wecom: update schedule: a valid time window is required")
	}
	if update.OpMode != OpModeAll && update.OpStartTime <= 0 {
		return "", errors.New("wecom: update schedule: op_start_time is required for a partial repeat update")
	}
	var out struct {
		ScheduleID string `json:"schedule_id"`
	}
	err := c.api.Call(ctx, "schedule update", "/cgi-bin/oa/schedule/update", update.request(), &out)
	if out.ScheduleID == "" {
		return update.ScheduleID, err
	}
	return out.ScheduleID, err
}

// AddAttendees adds participants to one schedule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97721
func (c *Schedule) AddAttendees(ctx context.Context, scheduleID string, attendees []Attendee) error {
	return c.changeAttendees(ctx, "schedule add attendees", "/cgi-bin/oa/schedule/add_attendees", scheduleID, attendees)
}

// DeleteAttendees removes participants from one schedule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97722
func (c *Schedule) DeleteAttendees(ctx context.Context, scheduleID string, attendees []Attendee) error {
	return c.changeAttendees(ctx, "schedule delete attendees", "/cgi-bin/oa/schedule/del_attendees", scheduleID, attendees)
}

func (c *Schedule) changeAttendees(ctx context.Context, op, path, scheduleID string, attendees []Attendee) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(scheduleID) == "" || len(attendees) == 0 {
		return errors.New("wecom: " + op + ": schedule_id and attendees are required")
	}
	request := struct {
		ScheduleID string     `json:"schedule_id"`
		Attendees  []Attendee `json:"attendees"`
	}{ScheduleID: scheduleID, Attendees: attendees}
	return c.api.Call(ctx, op, path, request, nil)
}

// CalendarSchedules returns one page of schedules on a calendar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97723
func (c *Schedule) CalendarSchedules(ctx context.Context, calID string, offset, limit int) ([]Entry, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	calID = strings.TrimSpace(calID)
	if calID == "" {
		return nil, errors.New("wecom: calendar schedules: cal_id is required")
	}
	request := struct {
		CalID  string `json:"cal_id"`
		Offset int    `json:"offset,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{CalID: calID, Offset: offset, Limit: limit}
	var out struct {
		List []Entry `json:"schedule_list"`
	}
	err := c.api.Call(ctx, "schedule list", "/cgi-bin/oa/schedule/get_by_calendar", request, &out)
	return out.List, err
}

// GetSchedules reads schedules by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97724
func (c *Schedule) GetSchedules(ctx context.Context, scheduleIDs []string) ([]Entry, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(scheduleIDs) == 0 {
		return nil, errors.New("wecom: get schedules: schedule_id_list is empty")
	}
	var out struct {
		List []Entry `json:"schedule_list"`
	}
	err := c.api.Call(ctx, "schedule get", "/cgi-bin/oa/schedule/get",
		map[string][]string{"schedule_id_list": scheduleIDs}, &out)
	return out.List, err
}

// DeleteSchedule removes one schedule from a calendar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97725
func (c *Schedule) DeleteSchedule(ctx context.Context, scheduleID string) error {
	if err := check(c); err != nil {
		return err
	}
	scheduleID = strings.TrimSpace(scheduleID)
	if scheduleID == "" {
		return errors.New("wecom: delete schedule: schedule_id is required")
	}
	return c.api.Call(ctx, "schedule delete", "/cgi-bin/oa/schedule/del",
		map[string]string{"schedule_id": scheduleID}, nil)
}

// Calendar event names.
const (
	EventDeleteCalendar  = "delete_calendar"
	EventModifyCalendar  = "modify_calendar"
	EventModifySchedule  = "modify_schedule"
	EventDeleteSchedule  = "delete_schedule"
	EventRespondSchedule = "respond_schedule"
)

// Event is one decrypted calendar callback.
type Event struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	CalID      string `xml:"CalId"`
	ScheduleID string `xml:"ScheduleId"`
}

// ParseEvent decodes a decrypted calendar callback. Decrypt the body with
// callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97728
// 文档: https://developer.work.weixin.qq.com/document/path/97730
// 文档: https://developer.work.weixin.qq.com/document/path/97731
// 文档: https://developer.work.weixin.qq.com/document/path/97732
// 文档: https://developer.work.weixin.qq.com/document/path/98111
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: calendar event: " + err.Error())
	}
	switch event.Name {
	case EventDeleteCalendar, EventModifyCalendar, EventModifySchedule,
		EventDeleteSchedule, EventRespondSchedule:
	default:
		return Event{}, errors.New("wecom: calendar event: unexpected event " + event.Name)
	}
	if strings.TrimSpace(event.CalID) == "" {
		return Event{}, errors.New("wecom: calendar event: CalId is required")
	}
	return event, nil
}

func check(c *Schedule) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: schedule client is unavailable")
	}
	return nil
}
