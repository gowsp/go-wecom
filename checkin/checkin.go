// Package checkin implements the WeCom check-in API (/cgi-bin/checkin and the
// hardware check-in endpoint): punch records, day/month reports, schedules,
// corrections and check-in rule management.
package checkin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Checkin reads and maintains attendance data with one shared transport.
type Checkin struct{ api *client.Client }

// New creates a check-in client.
func New(api *client.Client) *Checkin { return &Checkin{api: api} }

// Record data types accepted by Records.
const (
	// DataTypeAll returns every punch in the window.
	DataTypeAll = 1
	// DataTypeAllWithException returns punches plus exception details.
	DataTypeAllWithException = 2
	// DataTypeExceptionOnly returns only exceptions.
	DataTypeExceptionOnly = 3
)

// Record is one punch record.
type Record struct {
	UserID         string   `json:"userid"`
	GroupName      string   `json:"groupname,omitempty"`
	CheckinType    string   `json:"checkin_type,omitempty"`
	ExceptionType  string   `json:"exception_type,omitempty"`
	CheckinTime    int64    `json:"checkin_time"`
	LocationTitle  string   `json:"location_title,omitempty"`
	LocationDetail string   `json:"location_detail,omitempty"`
	WifiName       string   `json:"wifiname,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	MediaIDs       []string `json:"mediaids,omitempty"`
	Latitude       int64    `json:"lat,omitempty"`
	Longitude      int64    `json:"lng,omitempty"`
	DeviceSN       string   `json:"device_sn,omitempty"`
	DeviceName     string   `json:"device_name,omitempty"`
}

// HardwareRecord is one punch taken on check-in hardware.
type HardwareRecord struct {
	UserID      string `json:"userid"`
	CheckinTime int64  `json:"checkin_time"`
	DeviceSN    string `json:"device_sn"`
	DeviceName  string `json:"device_name"`
}

// ScheduleItem assigns one shift to one member on one day.
type ScheduleItem struct {
	UserID     string `json:"userid"`
	Day        int    `json:"day"`
	ScheduleID int    `json:"schedule_id"`
}

// Correction fixes one missed punch.
type Correction struct {
	UserID              string `json:"userid"`
	ScheduleDateTime    int64  `json:"schedule_date_time"`
	ScheduleCheckinTime int    `json:"schedule_checkin_time"`
	CheckinTime         int64  `json:"checkin_time"`
	Remark              string `json:"remark,omitempty"`
}

// NewRecord is one punch imported from an external system.
type NewRecord struct {
	UserID         string   `json:"userid"`
	CheckinTime    int64    `json:"checkin_time"`
	LocationTitle  string   `json:"location_title,omitempty"`
	LocationDetail string   `json:"location_detail,omitempty"`
	MediaIDs       []string `json:"mediaids,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	DeviceType     int      `json:"device_type,omitempty"`
	Latitude       int64    `json:"lat,omitempty"`
	Longitude      int64    `json:"lng,omitempty"`
}

// Records returns the punch records of the given members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90262
func (c *Checkin) Records(ctx context.Context, dataType int, startTime, endTime int64, userIDs []string) ([]Record, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if err := checkWindowAndUsers(startTime, endTime, userIDs, "records"); err != nil {
		return nil, err
	}
	request := struct {
		DataType  int      `json:"opencheckindatatype"`
		StartTime int64    `json:"starttime"`
		EndTime   int64    `json:"endtime"`
		UserIDs   []string `json:"useridlist"`
	}{DataType: dataType, StartTime: startTime, EndTime: endTime, UserIDs: userIDs}
	var out struct {
		Data []Record `json:"checkindata"`
	}
	err := c.api.Call(ctx, "checkin records", "/cgi-bin/checkin/getcheckindata", request, &out)
	return out.Data, err
}

// HardwareRecords returns punches taken on check-in hardware. filterType 1 means
// by member, 2 by department, 3 by tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94126
func (c *Checkin) HardwareRecords(ctx context.Context, filterType int, startTime, endTime int64, userIDs []string) ([]HardwareRecord, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if err := checkWindowAndUsers(startTime, endTime, userIDs, "hardware records"); err != nil {
		return nil, err
	}
	request := struct {
		FilterType int      `json:"filter_type"`
		StartTime  int64    `json:"starttime"`
		EndTime    int64    `json:"endtime"`
		UserIDs    []string `json:"useridlist"`
	}{FilterType: filterType, StartTime: startTime, EndTime: endTime, UserIDs: userIDs}
	var out struct {
		Data []HardwareRecord `json:"checkindata"`
	}
	err := c.api.Call(ctx, "checkin hardware records", "/cgi-bin/hardware/get_hardware_checkin_data", request, &out)
	return out.Data, err
}

// DayData returns daily attendance summaries. Records keep the official nested
// shape because the per-day fields depend on the rule configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93374
func (c *Checkin) DayData(ctx context.Context, startTime, endTime int64, userIDs []string) ([]json.RawMessage, error) {
	return c.report(ctx, "checkin day data", "/cgi-bin/checkin/getcheckin_daydata", startTime, endTime, userIDs)
}

// MonthData returns monthly attendance summaries with the official nested shape.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93387
func (c *Checkin) MonthData(ctx context.Context, startTime, endTime int64, userIDs []string) ([]json.RawMessage, error) {
	return c.report(ctx, "checkin month data", "/cgi-bin/checkin/getcheckin_monthdata", startTime, endTime, userIDs)
}

func (c *Checkin) report(ctx context.Context, op, path string, startTime, endTime int64, userIDs []string) ([]json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if err := checkWindowAndUsers(startTime, endTime, userIDs, op); err != nil {
		return nil, err
	}
	request := struct {
		StartTime int64    `json:"starttime"`
		EndTime   int64    `json:"endtime"`
		UserIDs   []string `json:"useridlist"`
	}{StartTime: startTime, EndTime: endTime, UserIDs: userIDs}
	var out struct {
		Data []json.RawMessage `json:"datas"`
	}
	err := c.api.Call(ctx, op, path, request, &out)
	return out.Data, err
}

// Schedules returns shift schedules in the window. Shift definitions vary by
// rule, so entries stay raw.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93380
func (c *Checkin) Schedules(ctx context.Context, startTime, endTime int64, userIDs []string) ([]json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if err := checkWindowAndUsers(startTime, endTime, userIDs, "schedules"); err != nil {
		return nil, err
	}
	request := struct {
		StartTime int64    `json:"starttime"`
		EndTime   int64    `json:"endtime"`
		UserIDs   []string `json:"useridlist"`
	}{StartTime: startTime, EndTime: endTime, UserIDs: userIDs}
	var out struct {
		List []json.RawMessage `json:"schedule_list"`
	}
	err := c.api.Call(ctx, "checkin schedules", "/cgi-bin/checkin/getcheckinschedulist", request, &out)
	return out.List, err
}

// SetSchedules assigns shifts to members of one group for one month.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93385
func (c *Checkin) SetSchedules(ctx context.Context, groupID, yearMonth int, items []ScheduleItem) error {
	if err := check(c); err != nil {
		return err
	}
	if groupID <= 0 || yearMonth <= 0 || len(items) == 0 {
		return errors.New("wecom: set schedules: groupid, yearmonth and items are required")
	}
	request := struct {
		GroupID   int            `json:"groupid"`
		Items     []ScheduleItem `json:"items"`
		YearMonth int            `json:"yearmonth"`
	}{GroupID: groupID, Items: items, YearMonth: yearMonth}
	return c.api.Call(ctx, "checkin set schedules", "/cgi-bin/checkin/setcheckinschedulist", request, nil)
}

// CorrectPunch files a correction for one missed punch.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95803
func (c *Checkin) CorrectPunch(ctx context.Context, correction Correction) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(correction.UserID) == "" || correction.ScheduleDateTime <= 0 ||
		correction.CheckinTime <= 0 {
		return errors.New("wecom: correct punch: userid, schedule_date_time and checkin_time are required")
	}
	return c.api.Call(ctx, "checkin correct punch", "/cgi-bin/checkin/punch_correction", correction, nil)
}

// AddRecords imports punches recorded by an external system.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99647
func (c *Checkin) AddRecords(ctx context.Context, records []NewRecord) error {
	if err := check(c); err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("wecom: add records: records are empty")
	}
	for i, record := range records {
		if strings.TrimSpace(record.UserID) == "" || record.CheckinTime <= 0 {
			return errors.New("wecom: add records: every record needs userid and checkin_time (" + itoa(i) + ")")
		}
	}
	request := struct {
		Records []NewRecord `json:"records"`
	}{Records: records}
	return c.api.Call(ctx, "checkin add records", "/cgi-bin/checkin/add_checkin_record", request, nil)
}

// AddFace registers the face image of one member. image is the base64 payload
// accepted by the official API.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93378
func (c *Checkin) AddFace(ctx context.Context, userID, image string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(image) == "" {
		return errors.New("wecom: add face: userid and userface are required")
	}
	request := struct {
		UserID   string `json:"userid"`
		UserFace string `json:"userface"`
	}{UserID: userID, UserFace: image}
	return c.api.Call(ctx, "checkin add face", "/cgi-bin/checkin/addcheckinuserface", request, nil)
}

// Options returns every check-in rule of the enterprise. The rule tree is large
// and depends on the configured rule type, so it stays raw.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93384
func (c *Checkin) Options(ctx context.Context) ([]json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		Groups []json.RawMessage `json:"group"`
	}
	err := c.api.Call(ctx, "checkin options", "/cgi-bin/checkin/getcorpcheckinoption", nil, &out)
	return out.Groups, err
}

// UserOptions returns the rules that apply to the given members on one day.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90263
func (c *Checkin) UserOptions(ctx context.Context, datetime int64, userIDs []string) ([]json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if datetime <= 0 || len(userIDs) == 0 {
		return nil, errors.New("wecom: user options: datetime and useridlist are required")
	}
	request := struct {
		DateTime int64    `json:"datetime"`
		UserIDs  []string `json:"useridlist"`
	}{DateTime: datetime, UserIDs: userIDs}
	var out struct {
		Info []json.RawMessage `json:"info"`
	}
	err := c.api.Call(ctx, "checkin user options", "/cgi-bin/checkin/getcheckinoption", request, &out)
	return out.Info, err
}

// AddOption creates or replaces one check-in rule. The rule body is passed
// through because its fields depend on grouptype.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98041
func (c *Checkin) AddOption(ctx context.Context, option json.RawMessage) error {
	if err := check(c); err != nil {
		return err
	}
	if len(option) == 0 || !json.Valid(option) {
		return errors.New("wecom: add option: option must be valid JSON")
	}
	return c.api.Call(ctx, "checkin add option", "/cgi-bin/checkin/add_checkin_option",
		json.RawMessage(option), nil)
}

// ClearOptionField resets one array field of a rule, for example 特殊日期.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98041
func (c *Checkin) ClearOptionField(ctx context.Context, groupID int, field string) error {
	if err := check(c); err != nil {
		return err
	}
	if groupID <= 0 || strings.TrimSpace(field) == "" {
		return errors.New("wecom: clear option field: groupid and field are required")
	}
	request := struct {
		GroupID int    `json:"groupid"`
		Field   string `json:"field"`
	}{GroupID: groupID, Field: field}
	return c.api.Call(ctx, "checkin clear option field",
		"/cgi-bin/checkin/clear_checkin_option_array_field", request, nil)
}

func checkWindowAndUsers(startTime, endTime int64, userIDs []string, op string) error {
	if startTime <= 0 || endTime <= startTime {
		return errors.New("wecom: checkin " + op + ": a valid time window is required")
	}
	if len(userIDs) == 0 {
		return errors.New("wecom: checkin " + op + ": useridlist is empty")
	}
	return nil
}

func check(c *Checkin) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: checkin client is unavailable")
	}
	return nil
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	return string(digits)
}
