package meeting

import (
	"context"
	"errors"
	"strings"
)

// Attendee is one participant of a meeting.
type Attendee struct {
	UserID       string `json:"userid,omitempty"`
	TmpOpenID    string `json:"tmp_openid,omitempty"`
	JoinTime     string `json:"join_time,omitempty"`
	QuitTime     string `json:"quit_time,omitempty"`
	InstanceID   int    `json:"instance_id,omitempty"`
	Role         int    `json:"role,omitempty"`
	WebinarRole  int    `json:"webinar_role,omitempty"`
	JoinType     int    `json:"join_type,omitempty"`
	Net          string `json:"net,omitempty"`
	AudioState   bool   `json:"audio_state,omitempty"`
	VideoState   bool   `json:"video_state,omitempty"`
	ScreenShared bool   `json:"screen_shared,omitempty"`
}

// AttendeePage is one page of participants.
type AttendeePage struct {
	Status     string     `json:"status,omitempty"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor"`
	Attendees  []Attendee `json:"attendees"`
}

// RealtimeAttendees returns the members currently in one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98157
func (c *Meeting) RealtimeAttendees(ctx context.Context, meetingID, subMeetingID, cursor string, limit int) (AttendeePage, error) {
	if err := check(c); err != nil {
		return AttendeePage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return AttendeePage{}, errors.New("wecom: realtime attendees: meetingid is required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		SubMeetingID string `json:"sub_meetingid,omitempty"`
		Cursor       string `json:"cursor,omitempty"`
		Limit        int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, SubMeetingID: subMeetingID, Cursor: cursor, Limit: limit}
	var out AttendeePage
	err := c.api.Call(ctx, "meeting realtime attendees",
		"/cgi-bin/meeting/get_realtime_attendee_list", request, &out)
	return out, err
}

// AttendeeList returns the members who joined within one window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98156
func (c *Meeting) AttendeeList(ctx context.Context, meetingID, subMeetingID string, startTime, endTime int64, cursor string, limit int) (AttendeePage, error) {
	if err := check(c); err != nil {
		return AttendeePage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return AttendeePage{}, errors.New("wecom: attendees: meetingid is required")
	}
	if startTime <= 0 || endTime <= startTime {
		return AttendeePage{}, errors.New("wecom: attendees: a valid time window is required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		SubMeetingID string `json:"sub_meetingid,omitempty"`
		StartTime    int64  `json:"start_time"`
		EndTime      int64  `json:"end_time"`
		Cursor       string `json:"cursor,omitempty"`
		Limit        int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, SubMeetingID: subMeetingID, StartTime: startTime, EndTime: endTime, Cursor: cursor, Limit: limit}
	var out AttendeePage
	err := c.api.Call(ctx, "meeting attendees", "/cgi-bin/meeting/get_attendee_list", request, &out)
	return out, err
}

// WaitingMember is one member in the waiting room.
type WaitingMember struct {
	UserID       string `json:"userid,omitempty"`
	InstanceID   int    `json:"instance_id,omitempty"`
	CustomerData string `json:"customer_data,omitempty"`
	TmpOpenID    string `json:"tmp_openid,omitempty"`
	JoinTime     int64  `json:"join_time,omitempty"`
	QuitTime     int64  `json:"quit_time,omitempty"`
}

// WaitingPage is one page of waiting room members.
type WaitingPage struct {
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor"`
	Members    []WaitingMember `json:"user_list"`
}

// WaitingRoomMembers returns the members currently waiting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98163
func (c *Meeting) WaitingRoomMembers(ctx context.Context, meetingID string, cursor string, limit int) (WaitingPage, error) {
	if err := check(c); err != nil {
		return WaitingPage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return WaitingPage{}, errors.New("wecom: waiting room: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, Cursor: cursor, Limit: limit}
	var out WaitingPage
	err := c.api.Call(ctx, "meeting waiting room",
		"/cgi-bin/meeting/waitingroom/get_current_user_list", request, &out)
	return out, err
}

// WaitingRoomRecords returns the waiting room history of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98164
func (c *Meeting) WaitingRoomRecords(ctx context.Context, meetingID string, cursor string, limit int) (WaitingPage, error) {
	if err := check(c); err != nil {
		return WaitingPage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return WaitingPage{}, errors.New("wecom: waiting room records: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, Cursor: cursor, Limit: limit}
	var out WaitingPage
	err := c.api.Call(ctx, "meeting waiting room records",
		"/cgi-bin/meeting/waitingroom/get_user_list", request, &out)
	return out, err
}

// DeviceMeeting is one meeting instance a member's device joined.
type DeviceMeeting struct {
	MeetingID  string `json:"meetingid"`
	InstanceID int    `json:"instance_id"`
}

// DeviceInMeeting reports which of the given meetings a member's device is in.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98165
func (c *Meeting) DeviceInMeeting(ctx context.Context, userID string, instanceIDs []int, meetingIDs []string) ([]DeviceMeeting, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(userID) == "" || len(meetingIDs) == 0 {
		return nil, errors.New("wecom: device in meeting: userid and meetingid_list are required")
	}
	request := struct {
		UserID      string   `json:"userid"`
		InstanceIDs []int    `json:"instance_id_list,omitempty"`
		MeetingIDs  []string `json:"meetingid_list"`
	}{UserID: userID, InstanceIDs: instanceIDs, MeetingIDs: meetingIDs}
	var out struct {
		List []DeviceMeeting `json:"result_list"`
	}
	err := c.api.Call(ctx, "meeting device in meeting",
		"/cgi-bin/meeting/check_device_in_meeting", request, &out)
	return out.List, err
}

// GuestList is the guest list of one meeting.
type GuestList struct {
	MeetingID   string  `json:"meetingid"`
	MeetingCode string  `json:"meeting_code,omitempty"`
	Title       string  `json:"title,omitempty"`
	Guests      []Guest `json:"guests"`
}

// Guests returns the external guests of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99039
func (c *Meeting) Guests(ctx context.Context, meetingID string) (GuestList, error) {
	if err := check(c); err != nil {
		return GuestList{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return GuestList{}, errors.New("wecom: guests: meetingid is required")
	}
	var out GuestList
	err := c.api.Call(ctx, "meeting guests", "/cgi-bin/meeting/get_guests",
		map[string]string{"meetingid": meetingID}, &out)
	return out, err
}

// SetGuests replaces the external guests of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99040
func (c *Meeting) SetGuests(ctx context.Context, meetingID string, guests []Guest) error {
	if err := check(c); err != nil {
		return err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return errors.New("wecom: set guests: meetingid is required")
	}
	request := struct {
		MeetingID string  `json:"meetingid"`
		Guests    []Guest `json:"guests"`
	}{MeetingID: meetingID, Guests: guests}
	return c.api.Call(ctx, "meeting set guests", "/cgi-bin/meeting/set_guests", request, nil)
}

// Quality is the network quality summary of one meeting.
type Quality struct {
	Quality            int        `json:"quality"`
	AudioQuality       int        `json:"audio_quality"`
	VideoQuality       int        `json:"video_quality"`
	ScreenShareQuality int        `json:"screen_share_quality"`
	NetworkQuality     int        `json:"network_quality"`
	Problems           []string   `json:"problems,omitempty"`
	Attendees          []Attendee `json:"attendees,omitempty"`
}

// Quality returns the meeting health metrics of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98821
func (c *Meeting) Quality(ctx context.Context, meetingID, subMeetingID string, startTime int64) (Quality, error) {
	if err := check(c); err != nil {
		return Quality{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return Quality{}, errors.New("wecom: quality: meetingid is required")
	}
	request := struct {
		MeetingID    string `json:"meetingid"`
		SubMeetingID string `json:"sub_meetingid,omitempty"`
		StartTime    int64  `json:"start_time,omitempty"`
	}{MeetingID: meetingID, SubMeetingID: subMeetingID, StartTime: startTime}
	var out Quality
	err := c.api.Call(ctx, "meeting quality", "/cgi-bin/meeting/get_quality", request, &out)
	return out, err
}
