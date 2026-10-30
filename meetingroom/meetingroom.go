// Package meetingroom implements the WeCom meeting room API
// (/cgi-bin/oa/meetingroom): room inventory, bookings and booking callbacks.
package meetingroom

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// MeetingRoom manages rooms and bookings with one shared transport.
type MeetingRoom struct{ api *client.Client }

// New creates a meeting room client.
func New(api *client.Client) *MeetingRoom { return &MeetingRoom{api: api} }

// Coordinate is the geographic position of a room.
type Coordinate struct {
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

// Range selects the members allowed to book a room.
type Range struct {
	UserList       []string `json:"user_list,omitempty"`
	DepartmentList []int    `json:"department_list,omitempty"`
}

// Room is one meeting room.
type Room struct {
	ID         int         `json:"meetingroom_id,omitempty"`
	Name       string      `json:"name,omitempty"`
	Capacity   int         `json:"capacity,omitempty"`
	City       string      `json:"city,omitempty"`
	Building   string      `json:"building,omitempty"`
	Floor      string      `json:"floor,omitempty"`
	Equipment  []int       `json:"equipment,omitempty"`
	Coordinate *Coordinate `json:"coordinate,omitempty"`
	Range      *Range      `json:"range,omitempty"`
}

// RoomFilter narrows the room list.
type RoomFilter struct {
	City      string `json:"city,omitempty"`
	Building  string `json:"building,omitempty"`
	Floor     string `json:"floor,omitempty"`
	Equipment []int  `json:"equipment,omitempty"`
}

// Booking is one reservation returned for a room and window.
type Booking struct {
	BookingID  string `json:"booking_id"`
	ScheduleID string `json:"schedule_id,omitempty"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	Booker     string `json:"booker,omitempty"`
	Status     int    `json:"status,omitempty"`
}

// RoomBookings groups the reservations of one room.
type RoomBookings struct {
	RoomID   int       `json:"meetingroom_id"`
	Schedule []Booking `json:"schedule"`
}

// BookRequest books one room.
type BookRequest struct {
	RoomID    int      `json:"meetingroom_id"`
	Subject   string   `json:"subject,omitempty"`
	StartTime int64    `json:"start_time"`
	EndTime   int64    `json:"end_time"`
	Booker    string   `json:"booker"`
	Attendees []string `json:"attendees,omitempty"`
}

// AddRoom creates one room and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93619
func (c *MeetingRoom) AddRoom(ctx context.Context, room Room) (int, error) {
	if err := checkClient(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(room.Name) == "" {
		return 0, errors.New("wecom: add room: name is required")
	}
	var out struct {
		ID int `json:"meetingroom_id"`
	}
	err := c.api.Call(ctx, "meetingroom add", "/cgi-bin/oa/meetingroom/add", room, &out)
	return out.ID, err
}

// UpdateRoom changes one room.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93619
func (c *MeetingRoom) UpdateRoom(ctx context.Context, room Room) error {
	if err := checkClient(c); err != nil {
		return err
	}
	if room.ID <= 0 {
		return errors.New("wecom: update room: meetingroom_id is required")
	}
	return c.api.Call(ctx, "meetingroom edit", "/cgi-bin/oa/meetingroom/edit", room, nil)
}

// DeleteRoom removes one room.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93619
func (c *MeetingRoom) DeleteRoom(ctx context.Context, roomID int) error {
	if err := checkClient(c); err != nil {
		return err
	}
	if roomID <= 0 {
		return errors.New("wecom: delete room: meetingroom_id is required")
	}
	return c.api.Call(ctx, "meetingroom del", "/cgi-bin/oa/meetingroom/del",
		map[string]int{"meetingroom_id": roomID}, nil)
}

// ListRooms returns the rooms matching the filter.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93619
func (c *MeetingRoom) ListRooms(ctx context.Context, filter RoomFilter) ([]Room, error) {
	if err := checkClient(c); err != nil {
		return nil, err
	}
	var out struct {
		Rooms []Room `json:"meetingroom_list"`
	}
	err := c.api.Call(ctx, "meetingroom list", "/cgi-bin/oa/meetingroom/list", filter, &out)
	return out.Rooms, err
}

// Book reserves one room and returns the booking id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) Book(ctx context.Context, request BookRequest) (string, error) {
	if err := checkClient(c); err != nil {
		return "", err
	}
	if request.RoomID <= 0 || strings.TrimSpace(request.Booker) == "" {
		return "", errors.New("wecom: book room: meetingroom_id and booker are required")
	}
	if request.StartTime <= 0 || request.EndTime <= request.StartTime {
		return "", errors.New("wecom: book room: a valid time window is required")
	}
	var out struct {
		BookingID string `json:"booking_id"`
	}
	err := c.api.Call(ctx, "meetingroom book", "/cgi-bin/oa/meetingroom/book", request, &out)
	return out.BookingID, err
}

// BookByMeeting reserves a room for an existing meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) BookByMeeting(ctx context.Context, meetingID string) (string, error) {
	if err := checkClient(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(meetingID) == "" {
		return "", errors.New("wecom: book by meeting: meetingid is required")
	}
	var out struct {
		BookingID string `json:"booking_id"`
	}
	err := c.api.Call(ctx, "meetingroom book by meeting", "/cgi-bin/oa/meetingroom/book_by_meeting",
		map[string]string{"meetingid": meetingID}, &out)
	return out.BookingID, err
}

// BookBySchedule reserves a room for an existing schedule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) BookBySchedule(ctx context.Context, scheduleID string) (string, error) {
	if err := checkClient(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(scheduleID) == "" {
		return "", errors.New("wecom: book by schedule: schedule_id is required")
	}
	var out struct {
		BookingID string `json:"booking_id"`
	}
	err := c.api.Call(ctx, "meetingroom book by schedule", "/cgi-bin/oa/meetingroom/book_by_schedule",
		map[string]string{"schedule_id": scheduleID}, &out)
	return out.BookingID, err
}

// CancelBooking releases one reservation.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) CancelBooking(ctx context.Context, bookingID string) error {
	if err := checkClient(c); err != nil {
		return err
	}
	bookingID = strings.TrimSpace(bookingID)
	if bookingID == "" {
		return errors.New("wecom: cancel booking: booking_id is required")
	}
	return c.api.Call(ctx, "meetingroom cancel booking", "/cgi-bin/oa/meetingroom/cancel_book",
		map[string]string{"booking_id": bookingID}, nil)
}

// Bookings returns the reservations of one room in a window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) Bookings(ctx context.Context, roomID int, startTime, endTime int64) ([]Booking, error) {
	if err := checkClient(c); err != nil {
		return nil, err
	}
	if roomID <= 0 {
		return nil, errors.New("wecom: room bookings: meetingroom_id is required")
	}
	request := struct {
		RoomID    int   `json:"meetingroom_id"`
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
	}{RoomID: roomID, StartTime: startTime, EndTime: endTime}
	var out struct {
		List []RoomBookings `json:"booking_list"`
	}
	err := c.api.Call(ctx, "meetingroom bookings", "/cgi-bin/oa/meetingroom/get_booking_info", request, &out)
	if err != nil {
		return nil, err
	}
	for _, group := range out.List {
		if group.RoomID == roomID {
			return group.Schedule, nil
		}
	}
	return nil, nil
}

// BookingInfo returns one booking by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93620
func (c *MeetingRoom) BookingInfo(ctx context.Context, bookingID string) (Booking, error) {
	if err := checkClient(c); err != nil {
		return Booking{}, err
	}
	bookingID = strings.TrimSpace(bookingID)
	if bookingID == "" {
		return Booking{}, errors.New("wecom: booking info: booking_id is required")
	}
	var out struct {
		Info Booking `json:"booking_info"`
	}
	err := c.api.Call(ctx, "meetingroom booking info", "/cgi-bin/oa/meetingroom/bookinfo/get",
		map[string]string{"booking_id": bookingID}, &out)
	return out.Info, err
}

// Event is one decrypted meeting room callback.
type Event struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	RoomID     int    `xml:"MeetingRoomId"`
	BookingID  string `xml:"BookingId"`
}

// Booking events.
const (
	EventBook   = "book_meeting_room"
	EventCancel = "cancel_meeting_room"
)

// ParseEvent decodes a decrypted meeting room callback. Decrypt the body with
// callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95333
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: meeting room event: " + err.Error())
	}
	if event.Name != EventBook && event.Name != EventCancel {
		return Event{}, errors.New("wecom: meeting room event: unexpected event " + event.Name)
	}
	return event, nil
}

func checkClient(c *MeetingRoom) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: meeting room client is unavailable")
	}
	return nil
}
