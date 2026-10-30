package meeting

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
)

// MRA addressing protocols.
const (
	// ProtocolSIP dials an H.323/SIP address.
	ProtocolSIP = 1
	// ProtocolH323 dials a raw H.323 address.
	ProtocolH323 = 2
)

// MRAAddress is one dial-in address of a meeting room or connector.
type MRAAddress struct {
	Protocol   int    `json:"protocol"`
	DialString string `json:"dial_string"`
}

// Room is one Rooms meeting room.
type Room struct {
	RoomID            string   `json:"meeting_room_id"`
	RoomName          string   `json:"meeting_room_name"`
	RoomLocation      string   `json:"meeting_room_location,omitempty"`
	AccountType       int      `json:"account_type,omitempty"`
	ActiveCode        string   `json:"active_code,omitempty"`
	ParticipantNumber int      `json:"participant_number,omitempty"`
	Status            int      `json:"meeting_room_status,omitempty"`
	ScheduledStatus   int      `json:"scheduled_status,omitempty"`
	RoomsIDs          []string `json:"rooms_id_list,omitempty"`
	City              string   `json:"city,omitempty"`
	Building          string   `json:"building,omitempty"`
	Floor             string   `json:"floor,omitempty"`
	Device            string   `json:"device,omitempty"`
	Desc              string   `json:"desc,omitempty"`
}

// RoomPage is one page of meeting rooms.
type RoomPage struct {
	HasMore    bool   `json:"has_more"`
	NextCursor string `json:"next_cursor"`
	Rooms      []Room `json:"meeting_room_list"`
}

// RoomMeeting is one meeting scheduled in a meeting room.
type RoomMeeting struct {
	MeetingID   string `json:"meetingid"`
	MeetingCode string `json:"meeting_code,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Status      string `json:"status,omitempty"`
	MeetingType int    `json:"meeting_type,omitempty"`
	StartTime   int64  `json:"start_time,omitempty"`
	EndTime     int64  `json:"end_time,omitempty"`
}

// RoomMeetingPage is one page of meetings of a meeting room.
type RoomMeetingPage struct {
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor"`
	Meetings   []RoomMeeting `json:"meeting_info_list"`
}

// RoomConfig is the configuration of one meeting room.
type RoomConfig struct {
	MeetingSettings struct {
		WaterMark        int  `json:"water_mark"`
		AutoResponse     int  `json:"auto_response"`
		Caption          bool `json:"caption"`
		RoomPMI          bool `json:"room_pmi"`
		RoomNotification bool `json:"room_notification"`
	} `json:"meeting_settings"`
	RecordSettings struct {
		ShareRecord    int  `json:"share_record"`
		DownloadRecord bool `json:"download_record"`
	} `json:"record_settings"`
}

// RoomInventory is the meeting room resource usage of the enterprise.
type RoomInventory struct {
	NormalCount         int `json:"normal_count"`
	SpecialCount        int `json:"special_count"`
	NormalUsedCount     int `json:"normal_used_count"`
	SpecialUsedCount    int `json:"special_used_count"`
	NormalExpiredCount  int `json:"normal_expired_count"`
	SpecialExpiredCount int `json:"special_expired_count"`
}

// BookRooms reserves meeting rooms for one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98791
func (c *Meeting) BookRooms(ctx context.Context, meetingID string, roomIDs []string, subjectVisible bool) ([]Room, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(roomIDs) == 0 {
		return nil, errors.New("wecom: book rooms: meetingid and meeting_room_id_list are required")
	}
	request := struct {
		MeetingID      string   `json:"meetingid"`
		RoomIDs        []string `json:"meeting_room_id_list"`
		SubjectVisible bool     `json:"subject_visible,omitempty"`
	}{MeetingID: meetingID, RoomIDs: roomIDs, SubjectVisible: subjectVisible}
	var out struct {
		Rooms []Room `json:"meeting_room_list"`
	}
	err := c.api.Call(ctx, "meeting book rooms", "/cgi-bin/meeting/rooms/book", request, &out)
	return out.Rooms, err
}

// ReleaseRooms frees meeting rooms from one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98792
func (c *Meeting) ReleaseRooms(ctx context.Context, meetingID string, roomIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(roomIDs) == 0 {
		return errors.New("wecom: release rooms: meetingid and meeting_room_id_list are required")
	}
	request := struct {
		MeetingID string   `json:"meetingid"`
		RoomIDs   []string `json:"meeting_room_id_list"`
	}{MeetingID: meetingID, RoomIDs: roomIDs}
	return c.api.Call(ctx, "meeting release rooms", "/cgi-bin/meeting/rooms/release", request, nil)
}

// ListRooms returns one page of meeting rooms, optionally filtered by name.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98795
func (c *Meeting) ListRooms(ctx context.Context, name, cursor string, limit int) (RoomPage, error) {
	if err := check(c); err != nil {
		return RoomPage{}, err
	}
	request := struct {
		Name   string `json:"meeting_room_name,omitempty"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Name: name, Cursor: cursor, Limit: limit}
	var out RoomPage
	err := c.api.Call(ctx, "meeting rooms", "/cgi-bin/meeting/rooms/list", request, &out)
	return out, err
}

// Room reads one meeting room. The response nests basic_info with fields that
// depend on the room device type, so it stays raw.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98793
func (c *Meeting) Room(ctx context.Context, roomID string) (json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return nil, errors.New("wecom: room info: meeting_room_id is required")
	}
	var out json.RawMessage
	err := c.api.Call(ctx, "meeting room info", "/cgi-bin/meeting/rooms/get_info",
		map[string]string{"meeting_room_id": roomID}, &out)
	return out, err
}

// RoomConfig reads the configuration of one meeting room.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98802
func (c *Meeting) RoomConfig(ctx context.Context, roomID string) (RoomConfig, error) {
	if err := check(c); err != nil {
		return RoomConfig{}, err
	}
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return RoomConfig{}, errors.New("wecom: room config: meeting_room_id is required")
	}
	var out RoomConfig
	err := c.api.Call(ctx, "meeting room config", "/cgi-bin/meeting/rooms/get_config",
		map[string]string{"meeting_room_id": roomID}, &out)
	return out, err
}

// RoomMeetings returns one page of meetings held in one meeting room.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98796
func (c *Meeting) RoomMeetings(ctx context.Context, roomID, roomsID string, startTime, endTime int64, cursor string, limit int) (RoomMeetingPage, error) {
	if err := check(c); err != nil {
		return RoomMeetingPage{}, err
	}
	if strings.TrimSpace(roomID) == "" && strings.TrimSpace(roomsID) == "" {
		return RoomMeetingPage{}, errors.New("wecom: room meetings: meeting_room_id or rooms_id is required")
	}
	if startTime <= 0 || endTime <= startTime {
		return RoomMeetingPage{}, errors.New("wecom: room meetings: a valid time window is required")
	}
	request := struct {
		RoomID    string `json:"meeting_room_id,omitempty"`
		RoomsID   string `json:"rooms_id,omitempty"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{RoomID: roomID, RoomsID: roomsID, StartTime: startTime, EndTime: endTime, Cursor: cursor, Limit: limit}
	var out RoomMeetingPage
	err := c.api.Call(ctx, "meeting room meetings", "/cgi-bin/meeting/rooms/list_meetings", request, &out)
	return out, err
}

// RoomInventory reads the meeting room resource usage.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98809
func (c *Meeting) RoomInventory(ctx context.Context) (RoomInventory, error) {
	if err := check(c); err != nil {
		return RoomInventory{}, err
	}
	var out RoomInventory
	err := c.api.Call(ctx, "meeting room inventory", "/cgi-bin/meeting/rooms/get_inventory", nil, &out)
	return out, err
}

// CallRoom dials a meeting room or MRA address and returns the invite id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98804
func (c *Meeting) CallRoom(ctx context.Context, meetingID, roomID string, address MRAAddress) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(roomID) == "" {
		return "", errors.New("wecom: call room: meetingid and meeting_room_id are required")
	}
	request := struct {
		MeetingID string     `json:"meetingid"`
		RoomID    string     `json:"meeting_room_id"`
		Address   MRAAddress `json:"mra_address"`
	}{MeetingID: meetingID, RoomID: roomID, Address: address}
	var out struct {
		InviteID string `json:"invite_id"`
	}
	err := c.api.Call(ctx, "meeting call room", "/cgi-bin/meeting/rooms/call", request, &out)
	return out.InviteID, err
}

// CancelRoomCall cancels an outgoing call to a meeting room.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98805
func (c *Meeting) CancelRoomCall(ctx context.Context, meetingID, inviteID, roomID string, address MRAAddress) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(roomID) == "" {
		return errors.New("wecom: cancel room call: meetingid and meeting_room_id are required")
	}
	request := struct {
		MeetingID string     `json:"meetingid"`
		InviteID  string     `json:"invite_id,omitempty"`
		RoomID    string     `json:"meeting_room_id"`
		Address   MRAAddress `json:"mra_address"`
	}{MeetingID: meetingID, InviteID: inviteID, RoomID: roomID, Address: address}
	return c.api.Call(ctx, "meeting cancel room call", "/cgi-bin/meeting/rooms/cancel_call", request, nil)
}

// RoomResponse is the answer state of an outgoing room call.
type RoomResponse struct {
	Status       int    `json:"status"`
	ResponseTime string `json:"response_time,omitempty"`
}

// GetRoomResponseStatus reads the answer state of an outgoing room call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98806
func (c *Meeting) GetRoomResponseStatus(ctx context.Context, meetingID, roomID string, address MRAAddress) (RoomResponse, error) {
	if err := check(c); err != nil {
		return RoomResponse{}, err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(roomID) == "" {
		return RoomResponse{}, errors.New("wecom: room response: meetingid and meeting_room_id are required")
	}
	request := struct {
		MeetingID string     `json:"meetingid"`
		RoomID    string     `json:"meeting_room_id"`
		Address   MRAAddress `json:"mra_address"`
	}{MeetingID: meetingID, RoomID: roomID, Address: address}
	var out RoomResponse
	err := c.api.Call(ctx, "meeting room response", "/cgi-bin/meeting/rooms/get_response_status", request, &out)
	return out, err
}

// MRAStatus is the state of one meeting room connector participant.
type MRAStatus struct {
	TmpOpenID         string `json:"tmp_openid"`
	InstanceID        int    `json:"instance_id"`
	UserRole          int    `json:"user_role"`
	WebinarMemberRole int    `json:"webinar_member_role,omitempty"`
	IP                string `json:"ip,omitempty"`
	Name              string `json:"name,omitempty"`
	AudioState        bool   `json:"audio_state"`
	VideoState        bool   `json:"video_state"`
	ScreenShared      bool   `json:"screen_shared_state"`
	DefaultLayout     int    `json:"default_layout,omitempty"`
	RaiseHandsState   bool   `json:"raise_hands_state"`
}

// MRAStatus reads the state of one connector participant.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98786
func (c *Meeting) MRAStatus(ctx context.Context, meetingID, tmpOpenID string) (MRAStatus, error) {
	if err := check(c); err != nil {
		return MRAStatus{}, err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(tmpOpenID) == "" {
		return MRAStatus{}, errors.New("wecom: mra status: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		TmpOpenID string `json:"tmp_openid"`
	}{MeetingID: meetingID, TmpOpenID: tmpOpenID}
	var out MRAStatus
	err := c.api.Call(ctx, "meeting mra status", "/cgi-bin/meeting/mra/query_status", request, &out)
	return out, err
}

// SetMRALayout switches the default layout of one connector.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98787
func (c *Meeting) SetMRALayout(ctx context.Context, meetingID string, layout int, noVideoUser int, tmpOpenID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(tmpOpenID) == "" {
		return errors.New("wecom: mra layout: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID     string `json:"meetingid"`
		DefaultLayout int    `json:"default_layout"`
		NoVideoUser   int    `json:"default_novideo_user,omitempty"`
		MRA           struct {
			TmpOpenID string `json:"tmp_openid"`
		} `json:"mra"`
	}{MeetingID: meetingID, DefaultLayout: layout, NoVideoUser: noVideoUser}
	request.MRA.TmpOpenID = tmpOpenID
	return c.api.Call(ctx, "meeting mra layout", "/cgi-bin/meeting/mra/set_default_layout", request, nil)
}

// SetMRAHand raises or lowers the hand of one connector.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98788
func (c *Meeting) SetMRAHand(ctx context.Context, meetingID string, raise bool, tmpOpenID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(tmpOpenID) == "" {
		return errors.New("wecom: mra hand: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		RaiseHand bool   `json:"raise_hand"`
		MRA       struct {
			TmpOpenID string `json:"tmp_openid"`
		} `json:"mra"`
	}{MeetingID: meetingID, RaiseHand: raise}
	request.MRA.TmpOpenID = tmpOpenID
	return c.api.Call(ctx, "meeting mra hand", "/cgi-bin/meeting/mra/set_raise_hand", request, nil)
}

// HangupMRA ends the call of one connector.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98789
func (c *Meeting) HangupMRA(ctx context.Context, meetingID, tmpOpenID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(tmpOpenID) == "" {
		return errors.New("wecom: mra hangup: meetingid and tmp_openid are required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		MRA       struct {
			TmpOpenID string `json:"tmp_openid"`
		} `json:"mra"`
	}{MeetingID: meetingID}
	request.MRA.TmpOpenID = tmpOpenID
	return c.api.Call(ctx, "meeting mra hangup", "/cgi-bin/meeting/mra/hangup", request, nil)
}

// RoomResponseEvent is one decrypted meeting room response callback.
type RoomResponseEvent struct {
	ToUser     string     `xml:"ToUserName"`
	FromUser   string     `xml:"FromUserName"`
	CreateTime int64      `xml:"CreateTime"`
	MsgType    string     `xml:"MsgType"`
	Name       string     `xml:"Event"`
	ChangeType string     `xml:"ChangeType"`
	MeetingID  string     `xml:"MeetingId"`
	RoomID     string     `xml:"MeetingRoomId"`
	Address    MRAAddress `xml:"MraAddress"`
	Status     int        `xml:"RoomResponseStatus"`
}

// ParseRoomResponseEvent decodes a decrypted meeting_room_response callback.
// Decrypt the body with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98783
func ParseRoomResponseEvent(data []byte) (RoomResponseEvent, error) {
	var event RoomResponseEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return RoomResponseEvent{}, errors.New("wecom: room response event: " + err.Error())
	}
	if event.Name != "meeting_change" || event.ChangeType != "meeting_room_response" {
		return RoomResponseEvent{}, errors.New("wecom: room response event: expected meeting_room_response")
	}
	if strings.TrimSpace(event.MeetingID) == "" || strings.TrimSpace(event.RoomID) == "" {
		return RoomResponseEvent{}, errors.New("wecom: room response event: MeetingId and MeetingRoomId are required")
	}
	return event, nil
}
