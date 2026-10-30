package meeting

import (
	"encoding/xml"
	"errors"
	"strings"
)

// Meeting change types carried by Event=meeting_change callbacks.
const (
	ChangeModifyMeeting       = "modify_meeting"
	ChangeCancelMeeting       = "cancel_meeting"
	ChangeMeetingStart        = "meeting_start"
	ChangeMeetingEnd          = "meeting_end"
	ChangeMeetingMuteAll      = "meeting_mute_all"
	ChangeMeetingUnmuteAll    = "meeting_unmute_all"
	ChangeJoinMeeting         = "join_meeting"
	ChangeQuitMeeting         = "quit_meeting"
	ChangeJoinBeforeHost      = "join_meeting_before_host"
	ChangeJoinWaitingRoom     = "join_waiting_room"
	ChangeQuitWaitingRoom     = "quit_waiting_room"
	ChangeJoinFromMeetingRoom = "join_from_meeting_room"
	ChangeMoveToWaitingRoom   = "move_to_waiting_room"
	ChangeOpenScreenShare     = "open_screen_share"
	ChangeCloseScreenShare    = "close_screen_share"
	ChangeRoleChange          = "role_change"
	ChangePSTNStatusUpdate    = "pstn_status_update"
	ChangeMediumUpload        = "medium_upload"
	ChangeEnroll              = "enroll"
	ChangeCancelEnroll        = "cancel_enroll"
)

// meetingChangeTypes is the set of meeting_change callbacks the parser accepts.
var meetingChangeTypes = map[string]bool{
	ChangeModifyMeeting: true, ChangeCancelMeeting: true,
	ChangeMeetingStart: true, ChangeMeetingEnd: true,
	ChangeMeetingMuteAll: true, ChangeMeetingUnmuteAll: true,
	ChangeJoinMeeting: true, ChangeQuitMeeting: true,
	ChangeJoinBeforeHost: true, ChangeJoinWaitingRoom: true,
	ChangeQuitWaitingRoom: true, ChangeJoinFromMeetingRoom: true,
	ChangeMoveToWaitingRoom: true, ChangeOpenScreenShare: true,
	ChangeCloseScreenShare: true, ChangeRoleChange: true,
	ChangePSTNStatusUpdate: true, ChangeMediumUpload: true,
	ChangeEnroll: true, ChangeCancelEnroll: true,
	EventWebinarRoleChange: true, EventWebinarWarmUpUpload: true,
}

// OperatedUserInfo is the participant a meeting_change callback refers to.
type OperatedUserInfo struct {
	UserID    string `xml:"UserId"`
	TmpOpenID string `xml:"TmpOpenId"`
	UserRole  int    `xml:"UserRole"`
}

// UploadInfo is one medium upload result.
type UploadInfo struct {
	MediumURL    string `xml:"MediumUrl"`
	MediumType   int    `xml:"MediumType"`
	UploadStatus int    `xml:"UploadStatus"`
	ErrorMsg     string `xml:"ErrorMsg"`
}

// MeetingChangeEvent is one decrypted meeting change callback. The payload is a
// superset: fields not relevant to a change type stay empty.
type MeetingChangeEvent struct {
	ToUser            string            `xml:"ToUserName"`
	FromUser          string            `xml:"FromUserName"`
	FromUserTmpOpenID string            `xml:"FromUserTmpOpenId"`
	CreateTime        int64             `xml:"CreateTime"`
	MsgType           string            `xml:"MsgType"`
	Name              string            `xml:"Event"`
	ChangeType        string            `xml:"ChangeType"`
	MeetingID         string            `xml:"MeetingId"`
	MeetingCode       string            `xml:"MeetingCode"`
	OperatedUser      *OperatedUserInfo `xml:"OperatedUser"`
	PSTNStatus        string            `xml:"PstnStatus"`
	AllUploadStatus   string            `xml:"AllUploadStatus"`
	Uploads           []UploadInfo      `xml:"UploadInfo"`
	WarmUpInfo        *WarmUpInfo       `xml:"WarmUpInfo"`
}

// ParseMeetingChangeEvent decodes a decrypted meeting_change callback. Decrypt
// the body with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99081
// 文档: https://developer.work.weixin.qq.com/document/path/99082
// 文档: https://developer.work.weixin.qq.com/document/path/98333
// 文档: https://developer.work.weixin.qq.com/document/path/98337
// 文档: https://developer.work.weixin.qq.com/document/path/98341
// 文档: https://developer.work.weixin.qq.com/document/path/98345
// 文档: https://developer.work.weixin.qq.com/document/path/98348
// 文档: https://developer.work.weixin.qq.com/document/path/98352
// 文档: https://developer.work.weixin.qq.com/document/path/98353
// 文档: https://developer.work.weixin.qq.com/document/path/98354
// 文档: https://developer.work.weixin.qq.com/document/path/98355
// 文档: https://developer.work.weixin.qq.com/document/path/98393
// 文档: https://developer.work.weixin.qq.com/document/path/98394
// 文档: https://developer.work.weixin.qq.com/document/path/98395
// 文档: https://developer.work.weixin.qq.com/document/path/98396
// 文档: https://developer.work.weixin.qq.com/document/path/98397
// 文档: https://developer.work.weixin.qq.com/document/path/98774
// 文档: https://developer.work.weixin.qq.com/document/path/98775
// 文档: https://developer.work.weixin.qq.com/document/path/98781
// 文档: https://developer.work.weixin.qq.com/document/path/98782
func ParseMeetingChangeEvent(data []byte) (MeetingChangeEvent, error) {
	var event MeetingChangeEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return MeetingChangeEvent{}, errors.New("wecom: meeting change event: " + err.Error())
	}
	if event.Name != "meeting_change" {
		return MeetingChangeEvent{}, errors.New("wecom: meeting change event: unexpected event " + event.Name)
	}
	if !meetingChangeTypes[event.ChangeType] {
		return MeetingChangeEvent{}, errors.New("wecom: meeting change event: unexpected change type " + event.ChangeType)
	}
	if strings.TrimSpace(event.MeetingID) == "" {
		return MeetingChangeEvent{}, errors.New("wecom: meeting change event: MeetingId is required")
	}
	return event, nil
}

// MeetingStatistics is the meeting_statistics callback of one started meeting.
type MeetingStatistics struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUser"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	ChangeType string `xml:"ChangeType"`
	MeetingID  string `xml:"MeetingId"`
	Status     int    `xml:"Status"`
}

// ParseMeetingStatistics decodes a decrypted meeting_statistics callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99648
func ParseMeetingStatistics(data []byte) (MeetingStatistics, error) {
	var event MeetingStatistics
	if err := xml.Unmarshal(data, &event); err != nil {
		return MeetingStatistics{}, errors.New("wecom: meeting statistics event: " + err.Error())
	}
	if event.Name != "meeting_statistics" {
		return MeetingStatistics{}, errors.New("wecom: meeting statistics event: unexpected event " + event.Name)
	}
	return event, nil
}
