package meeting

import (
	"context"
	"errors"
	"strings"
)

// PhoneNumber is one PSTN participant identified by area, number and optional
// extension.
type PhoneNumber struct {
	Area      int    `json:"area"`
	Phone     string `json:"phone"`
	Extension string `json:"extension_number,omitempty"`
	Status    string `json:"status,omitempty"`
	TmpOpenID string `json:"tmp_openid,omitempty"`
}

// CalloutResult is the outcome of a batch callout.
type CalloutResult struct {
	Numbers []PhoneNumber `json:"phone_numbers"`
	Invalid []PhoneNumber `json:"invalid_phone_numbers"`
}

// CalloutPage is one page of callout states.
type CalloutPage struct {
	Numbers    []PhoneNumber `json:"phone_numbers"`
	HasMore    bool          `json:"has_more"`
	NextCursor string        `json:"next_cursor"`
}

// Callout dials phone numbers into a meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98823
func (c *Meeting) Callout(ctx context.Context, meetingID string, numbers []PhoneNumber) (CalloutResult, error) {
	if err := check(c); err != nil {
		return CalloutResult{}, err
	}
	if strings.TrimSpace(meetingID) == "" || len(numbers) == 0 {
		return CalloutResult{}, errors.New("wecom: callout: meetingid and phone_numbers are required")
	}
	request := struct {
		MeetingID string        `json:"meetingid"`
		Numbers   []PhoneNumber `json:"phone_numbers"`
	}{MeetingID: meetingID, Numbers: numbers}
	var out CalloutResult
	err := c.api.Call(ctx, "meeting callout", "/cgi-bin/meeting/phone/callout", request, &out)
	return out, err
}

// CalloutStatus returns one page of PSTN callout states.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98824
func (c *Meeting) CalloutStatus(ctx context.Context, meetingID, cursor string, limit int) (CalloutPage, error) {
	if err := check(c); err != nil {
		return CalloutPage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return CalloutPage{}, errors.New("wecom: callout status: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, Cursor: cursor, Limit: limit}
	var out CalloutPage
	err := c.api.Call(ctx, "meeting callout status",
		"/cgi-bin/meeting/phone/get_callout_status", request, &out)
	return out, err
}

// PhoneMemberIDs resolves the temporary open ids of PSTN participants.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98825
func (c *Meeting) PhoneMemberIDs(ctx context.Context, meetingID string, numbers []PhoneNumber) ([]PhoneNumber, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(numbers) == 0 {
		return nil, errors.New("wecom: phone member ids: meetingid and phone_numbers are required")
	}
	request := struct {
		MeetingID string        `json:"meetingid"`
		Numbers   []PhoneNumber `json:"phone_numbers"`
	}{MeetingID: meetingID, Numbers: numbers}
	var out struct {
		List []PhoneNumber `json:"tmp_openid_list"`
	}
	err := c.api.Call(ctx, "meeting phone member ids",
		"/cgi-bin/meeting/phone/get_tmp_openid", request, &out)
	return out.List, err
}

// Device is one hardware device registered to a meeting room.
type Device struct {
	RoomID       string `json:"meeting_room_id"`
	RoomsID      string `json:"rooms_id,omitempty"`
	RoomName     string `json:"meeting_room_name,omitempty"`
	RoomLocation string `json:"meeting_room_location,omitempty"`
	Model        string `json:"device_model,omitempty"`
	AppVersion   string `json:"app_version,omitempty"`
	Status       int    `json:"meeting_room_status,omitempty"`
	DeviceStatus int    `json:"device_status,omitempty"`
}

// DevicePage is one page of meeting room devices.
type DevicePage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	Devices    []Device `json:"device_info_list"`
}

// Devices returns one page of meeting room devices.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98798
func (c *Meeting) Devices(ctx context.Context, name, cursor string, limit int) (DevicePage, error) {
	if err := check(c); err != nil {
		return DevicePage{}, err
	}
	request := struct {
		Name   string `json:"meeting_room_name,omitempty"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Name: name, Cursor: cursor, Limit: limit}
	var out DevicePage
	err := c.api.Call(ctx, "meeting devices", "/cgi-bin/meeting/rooms/list_devices", request, &out)
	return out, err
}

// Controller is one controller paired with a meeting room.
type Controller struct {
	RoomsID      string `json:"rooms_id"`
	RoomName     string `json:"meeting_room_name,omitempty"`
	RoomLocation string `json:"meeting_room_location,omitempty"`
	Name         string `json:"controller_name"`
	Manufacture  string `json:"manufacture_name,omitempty"`
	Model        string `json:"controller_model,omitempty"`
	AppVersion   string `json:"app_version,omitempty"`
}

// ControllerPage is one page of meeting room controllers.
type ControllerPage struct {
	HasMore    bool         `json:"has_more"`
	NextCursor string       `json:"next_cursor"`
	Devices    []Controller `json:"controller_info_list"`
}

// Controllers returns one page of meeting room controllers.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98799
func (c *Meeting) Controllers(ctx context.Context, name, cursor string, limit int) (ControllerPage, error) {
	if err := check(c); err != nil {
		return ControllerPage{}, err
	}
	request := struct {
		Name   string `json:"controller_name,omitempty"`
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Name: name, Cursor: cursor, Limit: limit}
	var out ControllerPage
	err := c.api.Call(ctx, "meeting controllers", "/cgi-bin/meeting/rooms/list_controllers", request, &out)
	return out, err
}

// VIPJob is one asynchronous professional account job.
type VIPJob struct {
	JobID   string   `json:"jobid"`
	Invalid []string `json:"invalid_userid_list"`
}

// VIPPage is one page of professional accounts.
type VIPPage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	UserIDs    []string `json:"userid_list"`
}

// VIPJobResult is the outcome of one asynchronous professional account job.
type VIPJobResult struct {
	Status  int      `json:"status"`
	UserIDs []string `json:"userid_list"`
}

// SubmitAddVIP starts granting professional accounts and returns the job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99508
func (c *Meeting) SubmitAddVIP(ctx context.Context, userIDs []string) (VIPJob, error) {
	if err := check(c); err != nil {
		return VIPJob{}, err
	}
	if len(userIDs) == 0 {
		return VIPJob{}, errors.New("wecom: add vip: userid_list is empty")
	}
	var out VIPJob
	err := c.api.Call(ctx, "meeting add vip", "/cgi-bin/meeting/vip/submit_batch_add_job",
		map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// AddVIPResult reads the outcome of a grant job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99508
func (c *Meeting) AddVIPResult(ctx context.Context, jobID string) (VIPJobResult, error) {
	return c.vipResult(ctx, "meeting add vip result", "/cgi-bin/meeting/vip/batch_add_job_result", jobID)
}

// SubmitRemoveVIP starts revoking professional accounts and returns the job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99509
func (c *Meeting) SubmitRemoveVIP(ctx context.Context, userIDs []string) (VIPJob, error) {
	if err := check(c); err != nil {
		return VIPJob{}, err
	}
	if len(userIDs) == 0 {
		return VIPJob{}, errors.New("wecom: remove vip: userid_list is empty")
	}
	var out VIPJob
	err := c.api.Call(ctx, "meeting remove vip", "/cgi-bin/meeting/vip/submit_batch_del_job",
		map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// RemoveVIPResult reads the outcome of a revoke job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99509
func (c *Meeting) RemoveVIPResult(ctx context.Context, jobID string) (VIPJobResult, error) {
	return c.vipResult(ctx, "meeting remove vip result", "/cgi-bin/meeting/vip/batch_del_job_result", jobID)
}

func (c *Meeting) vipResult(ctx context.Context, op, path, jobID string) (VIPJobResult, error) {
	if err := check(c); err != nil {
		return VIPJobResult{}, err
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return VIPJobResult{}, errors.New("wecom: " + op + ": jobid is required")
	}
	var out VIPJobResult
	err := c.api.Call(ctx, op, path, map[string]string{"jobid": jobID}, &out)
	return out, err
}

// ListVIP returns one page of professional accounts.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99510
func (c *Meeting) ListVIP(ctx context.Context, cursor string, limit int) (VIPPage, error) {
	if err := check(c); err != nil {
		return VIPPage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out VIPPage
	err := c.api.Call(ctx, "meeting list vip", "/cgi-bin/meeting/vip/list", request, &out)
	return out, err
}
