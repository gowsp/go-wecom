package meeting

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"
)

// Recording states.
const (
	// RecordStateRecording means the meeting is still being recorded.
	RecordStateRecording = 1
	// RecordStateComplete means the recording finished and files exist.
	RecordStateComplete = 2
)

// RecordFile is one recording file of a meeting.
type RecordFile struct {
	FileID          string `json:"record_file_id"`
	StartTime       int64  `json:"record_start_time,omitempty"`
	EndTime         int64  `json:"record_end_time,omitempty"`
	State           int    `json:"record_state,omitempty"`
	ViewAddress     string `json:"view_address,omitempty"`
	DownloadAddress string `json:"download_address,omitempty"`
	FileType        string `json:"download_address_file_type,omitempty"`
	AudioAddress    string `json:"audio_address,omitempty"`
	AudioFileType   string `json:"audio_address_file_type,omitempty"`
}

// RecordMeeting is one recorded meeting.
type RecordMeeting struct {
	RecordID    string       `json:"meeting_record_id"`
	MeetingID   string       `json:"meetingid"`
	MeetingCode string       `json:"meeting_code,omitempty"`
	HostUserID  string       `json:"host_user_id,omitempty"`
	StartTime   int64        `json:"meeting_start_time,omitempty"`
	Title       string       `json:"title,omitempty"`
	State       int          `json:"state,omitempty"`
	Files       []RecordFile `json:"record_files"`
}

// RecordPage is one page of recorded meetings.
type RecordPage struct {
	HasMore    bool            `json:"has_more"`
	NextCursor string          `json:"next_cursor"`
	Meetings   []RecordMeeting `json:"record_meetings"`
}

// RecordQuery selects recordings.
type RecordQuery struct {
	MeetingID   string
	MeetingCode string
	UserID      string
	StartTime   int64
	EndTime     int64
	Cursor      string
	Limit       int
}

// SharingConfig is the sharing configuration of one recording.
type SharingConfig struct {
	EnableSharing       *bool  `json:"enable_sharing,omitempty"`
	SharingAuthType     int    `json:"sharing_auth_type,omitempty"`
	EnablePassword      *bool  `json:"enable_password,omitempty"`
	Password            string `json:"password,omitempty"`
	EnableSharingExpire *bool  `json:"enable_sharing_expire,omitempty"`
	SharingExpire       int64  `json:"sharing_expire,omitempty"`
	AllowDownload       *bool  `json:"allow_download,omitempty"`
}

// RecordStat is one day of recording access statistics.
type RecordStat struct {
	Date          string `json:"date"`
	ViewCount     int    `json:"view_count"`
	DownloadCount int    `json:"download_count"`
}

// TranscriptParagraph is one paragraph of a recording transcript.
type TranscriptParagraph struct {
	PID           string     `json:"pid"`
	StartTime     int64      `json:"start_time"`
	EndTime       int64      `json:"end_time"`
	SpeakerUserID string     `json:"-"`
	Sentences     []Sentence `json:"sentences,omitempty"`
}

// Sentence is one transcribed sentence.
type Sentence struct {
	SID       string `json:"sid"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	Text      string `json:"text,omitempty"`
}

// TranscriptHit is one search hit inside a transcript.
type TranscriptHit struct {
	PID    string `json:"pid"`
	SID    string `json:"sid"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// TranscriptTimeline is one timestamped location of a search hit.
type TranscriptTimeline struct {
	PID       string `json:"pid"`
	SID       string `json:"sid"`
	StartTime int64  `json:"start_time"`
}

// RecordList returns one page of recorded meetings.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98192
func (c *Meeting) RecordList(ctx context.Context, query RecordQuery) (RecordPage, error) {
	if err := check(c); err != nil {
		return RecordPage{}, err
	}
	if query.MeetingID == "" && query.MeetingCode == "" && query.UserID == "" {
		return RecordPage{}, errors.New("wecom: record list: meetingid, meeting_code or userid is required")
	}
	request := struct {
		MeetingID   string `json:"meetingid,omitempty"`
		MeetingCode string `json:"meeting_code,omitempty"`
		StartTime   int64  `json:"start_time,omitempty"`
		EndTime     int64  `json:"end_time,omitempty"`
		UserID      string `json:"userid,omitempty"`
		Cursor      string `json:"cursor,omitempty"`
		Limit       int    `json:"limit,omitempty"`
	}{query.MeetingID, query.MeetingCode, query.StartTime, query.EndTime, query.UserID, query.Cursor, query.Limit}
	var out RecordPage
	err := c.api.Call(ctx, "meeting records", "/cgi-bin/meeting/record/list", request, &out)
	return out, err
}

// RecordStats returns the daily access statistics of one recording.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98209
func (c *Meeting) RecordStats(ctx context.Context, recordID, meetingID string, startTime, endTime int64) ([]RecordStat, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if recordID == "" && meetingID == "" {
		return nil, errors.New("wecom: record stats: meeting_record_id or meetingid is required")
	}
	request := struct {
		RecordID  string `json:"meeting_record_id,omitempty"`
		MeetingID string `json:"meetingid,omitempty"`
		StartTime int64  `json:"start_time,omitempty"`
		EndTime   int64  `json:"end_time,omitempty"`
	}{RecordID: recordID, MeetingID: meetingID, StartTime: startTime, EndTime: endTime}
	var out struct {
		Summaries []RecordStat `json:"summaries"`
	}
	err := c.api.Call(ctx, "meeting record stats", "/cgi-bin/meeting/record/get_statistics", request, &out)
	return out.Summaries, err
}

// SetRecordSharing updates the sharing configuration of one recording.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98208
func (c *Meeting) SetRecordSharing(ctx context.Context, recordID, meetingID string, config SharingConfig) error {
	if err := check(c); err != nil {
		return err
	}
	if recordID == "" && meetingID == "" {
		return errors.New("wecom: record sharing: meeting_record_id or meetingid is required")
	}
	request := struct {
		RecordID  string        `json:"meeting_record_id,omitempty"`
		MeetingID string        `json:"meetingid,omitempty"`
		Config    SharingConfig `json:"sharing_config"`
	}{RecordID: recordID, MeetingID: meetingID, Config: config}
	return c.api.Call(ctx, "meeting record sharing",
		"/cgi-bin/meeting/record/update_sharing_config", request, nil)
}

// DeleteRecord removes every file of one recording.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98206
func (c *Meeting) DeleteRecord(ctx context.Context, recordID, meetingID string) error {
	if err := check(c); err != nil {
		return err
	}
	if recordID == "" && meetingID == "" {
		return errors.New("wecom: delete record: meeting_record_id or meetingid is required")
	}
	request := struct {
		RecordID  string `json:"meeting_record_id,omitempty"`
		MeetingID string `json:"meetingid,omitempty"`
	}{RecordID: recordID, MeetingID: meetingID}
	return c.api.Call(ctx, "meeting delete record", "/cgi-bin/meeting/record/delete", request, nil)
}

// DeleteRecordFile removes one recording file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98207
func (c *Meeting) DeleteRecordFile(ctx context.Context, fileID, meetingID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(fileID) == "" {
		return errors.New("wecom: delete record file: record_file_id is required")
	}
	request := struct {
		FileID    string `json:"record_file_id"`
		MeetingID string `json:"meetingid,omitempty"`
	}{FileID: fileID, MeetingID: meetingID}
	return c.api.Call(ctx, "meeting delete record file", "/cgi-bin/meeting/record/delete_file", request, nil)
}

// RecordFileDetail is the detail of one recording file.
type RecordFileDetail struct {
	FileID       string `json:"record_file_id"`
	MeetingID    string `json:"meetingid"`
	MeetingCode  string `json:"meeting_code,omitempty"`
	ViewAddress  string `json:"view_address,omitempty"`
	DownloadURL  string `json:"download_address,omitempty"`
	DownloadType string `json:"download_address_file_type,omitempty"`
	AudioURL     string `json:"audio_address,omitempty"`
	AudioType    string `json:"audio_address_file_type,omitempty"`
	Summary      struct {
		DownloadAddress string `json:"download_address"`
	} `json:"meeting_summary"`
}

// RecordFile returns the detail of one recording file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98205
func (c *Meeting) RecordFile(ctx context.Context, fileID, meetingID string) (RecordFileDetail, error) {
	if err := check(c); err != nil {
		return RecordFileDetail{}, err
	}
	if strings.TrimSpace(fileID) == "" {
		return RecordFileDetail{}, errors.New("wecom: record file: record_file_id is required")
	}
	request := struct {
		FileID    string `json:"record_file_id"`
		MeetingID string `json:"meetingid,omitempty"`
	}{FileID: fileID, MeetingID: meetingID}
	var out RecordFileDetail
	err := c.api.Call(ctx, "meeting record file", "/cgi-bin/meeting/record/get_file", request, &out)
	return out, err
}

// RecordFiles returns the file list of one recording.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98196
func (c *Meeting) RecordFiles(ctx context.Context, recordID, meetingID string) (RecordMeeting, error) {
	if err := check(c); err != nil {
		return RecordMeeting{}, err
	}
	if recordID == "" && meetingID == "" {
		return RecordMeeting{}, errors.New("wecom: record files: meeting_record_id or meetingid is required")
	}
	request := struct {
		RecordID  string `json:"meeting_record_id,omitempty"`
		MeetingID string `json:"meetingid,omitempty"`
	}{RecordID: recordID, MeetingID: meetingID}
	var out RecordMeeting
	err := c.api.Call(ctx, "meeting record files", "/cgi-bin/meeting/record/get_file_list", request, &out)
	return out, err
}

// TranscriptParagraphs returns the paragraph timeline of one recording.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98212
func (c *Meeting) TranscriptParagraphs(ctx context.Context, fileID, meetingID string) (bool, []TranscriptParagraph, error) {
	if err := check(c); err != nil {
		return false, nil, err
	}
	if strings.TrimSpace(fileID) == "" {
		return false, nil, errors.New("wecom: transcript paragraphs: record_file_id is required")
	}
	request := struct {
		FileID    string `json:"record_file_id"`
		MeetingID string `json:"meetingid,omitempty"`
	}{FileID: fileID, MeetingID: meetingID}
	var out struct {
		AudioDetect int                   `json:"audio_detect"`
		Paragraphs  []TranscriptParagraph `json:"paragraphs"`
	}
	err := c.api.Call(ctx, "meeting transcript paragraphs",
		"/cgi-bin/meeting/record/transcript/get_paragraph_list", request, &out)
	return out.AudioDetect == 1, out.Paragraphs, err
}

// Transcript returns one paragraph of the transcript with its sentences.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98211
func (c *Meeting) Transcript(ctx context.Context, fileID, meetingID, pid string, limit int) ([]TranscriptParagraph, bool, error) {
	if err := check(c); err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(fileID) == "" {
		return nil, false, errors.New("wecom: transcript: record_file_id is required")
	}
	request := struct {
		FileID    string `json:"record_file_id"`
		MeetingID string `json:"meetingid,omitempty"`
		PID       string `json:"pid,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{FileID: fileID, MeetingID: meetingID, PID: pid, Limit: limit}
	var out struct {
		HasMore     bool `json:"has_more"`
		Transcripts struct {
			Paragraphs []TranscriptParagraph `json:"paragraphs"`
		} `json:"transcripts"`
	}
	err := c.api.Call(ctx, "meeting transcript",
		"/cgi-bin/meeting/record/transcript/get_detail", request, &out)
	return out.Transcripts.Paragraphs, out.HasMore, err
}

// SearchTranscript finds text inside one recording transcript.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98213
func (c *Meeting) SearchTranscript(ctx context.Context, fileID, meetingID, text string) ([]TranscriptHit, []TranscriptTimeline, error) {
	if err := check(c); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(fileID) == "" || strings.TrimSpace(text) == "" {
		return nil, nil, errors.New("wecom: search transcript: record_file_id and text are required")
	}
	request := struct {
		FileID    string `json:"record_file_id"`
		MeetingID string `json:"meetingid,omitempty"`
		Text      string `json:"text"`
	}{FileID: fileID, MeetingID: meetingID, Text: text}
	var out struct {
		Hits      []TranscriptHit      `json:"hits"`
		Timelines []TranscriptTimeline `json:"timelines"`
	}
	err := c.api.Call(ctx, "meeting search transcript",
		"/cgi-bin/meeting/record/transcript/search", request, &out)
	return out.Hits, out.Timelines, err
}

// Recording event change types.
const (
	EventStartRecording    = "start_recording"
	EventPauseRecording    = "pause_recording"
	EventResumeRecording   = "resume_recording"
	EventStopRecording     = "stop_recording"
	EventRecordingComplete = "recording_complete"
	EventDeleteRecording   = "delete_recording"
)

// RecordingEvent is one decrypted cloud recording callback.
type RecordingEvent struct {
	ToUser            string `xml:"ToUserName"`
	FromUser          string `xml:"FromUserName"`
	FromUserTmpOpenID string `xml:"FromUserTmpOpenId"`
	CreateTime        int64  `xml:"CreateTime"`
	MsgType           string `xml:"MsgType"`
	Name              string `xml:"Event"`
	ChangeType        string `xml:"ChangeType"`
	MeetingID         string `xml:"MeetingId"`
}

// ParseRecordingEvent decodes a decrypted cloud recording callback. Decrypt the
// body with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98398
// 文档: https://developer.work.weixin.qq.com/document/path/98399
// 文档: https://developer.work.weixin.qq.com/document/path/98400
// 文档: https://developer.work.weixin.qq.com/document/path/98401
// 文档: https://developer.work.weixin.qq.com/document/path/98402
// 文档: https://developer.work.weixin.qq.com/document/path/98404
func ParseRecordingEvent(data []byte) (RecordingEvent, error) {
	var event RecordingEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return RecordingEvent{}, errors.New("wecom: recording event: " + err.Error())
	}
	if event.Name != "meeting_change" {
		return RecordingEvent{}, errors.New("wecom: recording event: unexpected event " + event.Name)
	}
	switch event.ChangeType {
	case EventStartRecording, EventPauseRecording, EventResumeRecording,
		EventStopRecording, EventRecordingComplete, EventDeleteRecording:
	default:
		return RecordingEvent{}, errors.New("wecom: recording event: unexpected change type " + event.ChangeType)
	}
	if strings.TrimSpace(event.MeetingID) == "" {
		return RecordingEvent{}, errors.New("wecom: recording event: MeetingId is required")
	}
	return event, nil
}
