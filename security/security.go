// Package security implements the WeCom security management API
// (/cgi-bin/security): file and screen operation records, trusted devices,
// operation logs, server domains and the domain change callback.
package security

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Security reads security audit data with one shared transport.
type Security struct{ api *client.Client }

// New creates a security client.
func New(api *client.Client) *Security { return &Security{api: api} }

// Device types.
const (
	// DeviceCorporate is a corporate device.
	DeviceCorporate = 1
	// DevicePersonal is a personal device.
	DevicePersonal = 2
	// DeviceUnknown is a device whose owner is not confirmed.
	DeviceUnknown = 3
)

// FileOperation selects the operation a file record describes.
type FileOperation struct {
	Type   int `json:"type"`
	Source int `json:"source,omitempty"`
}

// FileRecord is one file operation entry.
type FileRecord struct {
	Time      int64         `json:"time"`
	UserID    string        `json:"userid"`
	Operation FileOperation `json:"operation"`
	FileInfo  string        `json:"file_info,omitempty"`
	FileSize  int64         `json:"file_size,omitempty"`
	FileMD5   string        `json:"file_md5,omitempty"`
}

// FileRecordPage is one page of file operation records.
type FileRecordPage struct {
	HasMore    bool         `json:"has_more"`
	NextCursor string       `json:"next_cursor"`
	Records    []FileRecord `json:"record_list"`
}

// FileRecords returns file operation records for data leak protection.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98079
func (c *Security) FileRecords(ctx context.Context, startTime, endTime int64, userIDs []string, op *FileOperation, cursor string, limit int) (FileRecordPage, error) {
	if err := check(c); err != nil {
		return FileRecordPage{}, err
	}
	if err := checkWindow(startTime, endTime, "file records"); err != nil {
		return FileRecordPage{}, err
	}
	request := struct {
		StartTime int64          `json:"start_time"`
		EndTime   int64          `json:"end_time"`
		UserIDs   []string       `json:"userid_list,omitempty"`
		Operation *FileOperation `json:"operation,omitempty"`
		Cursor    string         `json:"cursor,omitempty"`
		Limit     int            `json:"limit,omitempty"`
	}{startTime, endTime, userIDs, op, cursor, limit}
	var out FileRecordPage
	err := c.api.Call(ctx, "security file records",
		"/cgi-bin/security/get_file_oper_record", request, &out)
	return out, err
}

// ScreenRecord is one screenshot or screen recording entry.
type ScreenRecord struct {
	Time         int64  `json:"time"`
	UserID       string `json:"userid"`
	DepartmentID int    `json:"department_id,omitempty"`
	Type         int    `json:"screen_shot_type,omitempty"`
	Content      string `json:"screen_shot_content,omitempty"`
	System       string `json:"system,omitempty"`
}

// ScreenRecordPage is one page of screen records.
type ScreenRecordPage struct {
	HasMore    bool           `json:"has_more"`
	NextCursor string         `json:"next_cursor"`
	Records    []ScreenRecord `json:"record_list"`
}

// ScreenRecords returns screenshot and screen recording records.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100128
func (c *Security) ScreenRecords(ctx context.Context, startTime, endTime int64, userIDs []string, departmentIDs []int, screenShotType int, cursor string, limit int) (ScreenRecordPage, error) {
	if err := check(c); err != nil {
		return ScreenRecordPage{}, err
	}
	if err := checkWindow(startTime, endTime, "screen records"); err != nil {
		return ScreenRecordPage{}, err
	}
	request := struct {
		StartTime     int64    `json:"start_time"`
		EndTime       int64    `json:"end_time"`
		UserIDs       []string `json:"userid_list,omitempty"`
		DepartmentIDs []int    `json:"department_id_list,omitempty"`
		ScreenType    int      `json:"screen_shot_type,omitempty"`
		Cursor        string   `json:"cursor,omitempty"`
		Limit         int      `json:"limit,omitempty"`
	}{startTime, endTime, userIDs, departmentIDs, screenShotType, cursor, limit}
	var out ScreenRecordPage
	err := c.api.Call(ctx, "security screen records",
		"/cgi-bin/security/get_screen_oper_record", request, &out)
	return out, err
}

// Device identifies one device by its hardware identifiers.
type Device struct {
	System          string   `json:"system"`
	MACAddresses    []string `json:"mac_addr,omitempty"`
	MotherboardUUID string   `json:"motherboard_uuid,omitempty"`
	HarddiskUUID    []string `json:"harddisk_uuid,omitempty"`
	Domain          string   `json:"domain,omitempty"`
	PCName          string   `json:"pc_name,omitempty"`
	DeviceCode      string   `json:"device_code,omitempty"`
	LastLoginUserID string   `json:"last_login_userid,omitempty"`
}

// DeviceImport is the outcome of importing one device.
type DeviceImport struct {
	Index      int    `json:"device_index"`
	DeviceCode string `json:"device_code,omitempty"`
	Duplicated string `json:"duplicated_device_code,omitempty"`
	Status     int    `json:"status"`
}

// DevicePage is one page of trusted devices.
type DevicePage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	Devices    []Device `json:"device_list"`
}

// DeviceCodes is the outcome of approving or rejecting devices.
type DeviceCodes struct {
	Success []string `json:"success_list"`
	Fail    []string `json:"fail_list"`
}

// ImportTrustDevices imports corporate devices as trusted.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) ImportTrustDevices(ctx context.Context, devices []Device, deviceType int) ([]DeviceImport, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, errors.New("wecom: import trust devices: device_list is empty")
	}
	request := struct {
		Devices []Device `json:"device_list"`
		Type    int      `json:"type,omitempty"`
	}{Devices: devices, Type: deviceType}
	var out struct {
		Result []DeviceImport `json:"result"`
	}
	err := c.api.Call(ctx, "security import devices",
		"/cgi-bin/security/trustdevice/import", request, &out)
	return out.Result, err
}

// TrustDevices returns one page of trusted devices.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) TrustDevices(ctx context.Context, deviceType int, cursor string, limit int) (DevicePage, error) {
	if err := check(c); err != nil {
		return DevicePage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
		Type   int    `json:"type,omitempty"`
	}{Cursor: cursor, Limit: limit, Type: deviceType}
	var out DevicePage
	err := c.api.Call(ctx, "security trust devices",
		"/cgi-bin/security/trustdevice/list", request, &out)
	return out, err
}

// UserDevices returns the devices a member used.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) UserDevices(ctx context.Context, userID string, deviceType int) ([]Device, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("wecom: user devices: last_login_userid is required")
	}
	request := struct {
		UserID string `json:"last_login_userid"`
		Type   int    `json:"type,omitempty"`
	}{UserID: userID, Type: deviceType}
	var out struct {
		Devices []Device `json:"device_list"`
	}
	err := c.api.Call(ctx, "security user devices",
		"/cgi-bin/security/trustdevice/get_by_user", request, &out)
	return out.Devices, err
}

// DeleteTrustDevices removes devices from the trusted list.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) DeleteTrustDevices(ctx context.Context, deviceType int, deviceCodes []string) error {
	if err := check(c); err != nil {
		return err
	}
	if len(deviceCodes) == 0 {
		return errors.New("wecom: delete trust devices: device_code_list is empty")
	}
	request := struct {
		Type  int      `json:"type,omitempty"`
		Codes []string `json:"device_code_list"`
	}{Type: deviceType, Codes: deviceCodes}
	return c.api.Call(ctx, "security delete devices",
		"/cgi-bin/security/trustdevice/delete", request, nil)
}

// ApproveDevices confirms unknown devices as trusted.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) ApproveDevices(ctx context.Context, deviceCodes []string) (DeviceCodes, error) {
	if err := check(c); err != nil {
		return DeviceCodes{}, err
	}
	if len(deviceCodes) == 0 {
		return DeviceCodes{}, errors.New("wecom: approve devices: device_code_list is empty")
	}
	var out DeviceCodes
	err := c.api.Call(ctx, "security approve devices",
		"/cgi-bin/security/trustdevice/approve",
		map[string][]string{"device_code_list": deviceCodes}, &out)
	return out, err
}

// RejectDevices rejects unknown devices.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98920
func (c *Security) RejectDevices(ctx context.Context, deviceCodes []string) (DeviceCodes, error) {
	if err := check(c); err != nil {
		return DeviceCodes{}, err
	}
	if len(deviceCodes) == 0 {
		return DeviceCodes{}, errors.New("wecom: reject devices: device_code_list is empty")
	}
	var out DeviceCodes
	err := c.api.Call(ctx, "security reject devices",
		"/cgi-bin/security/trustdevice/reject",
		map[string][]string{"device_code_list": deviceCodes}, &out)
	return out, err
}

// SecurityVIPJob is one asynchronous professional account job.
type SecurityVIPJob struct {
	JobID   string   `json:"jobid"`
	Invalid []string `json:"invalid_userid_list"`
}

// SecurityVIPPage is one page of professional accounts.
type SecurityVIPPage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	UserIDs    []string `json:"userid_list"`
}

// SecurityVIPResult is the outcome of one professional account job.
type SecurityVIPResult struct {
	Status  int      `json:"status"`
	UserIDs []string `json:"userid_list"`
}

// SubmitAddSecurityVIP starts granting security professional accounts.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99503
func (c *Security) SubmitAddSecurityVIP(ctx context.Context, userIDs []string) (SecurityVIPJob, error) {
	if err := check(c); err != nil {
		return SecurityVIPJob{}, err
	}
	if len(userIDs) == 0 {
		return SecurityVIPJob{}, errors.New("wecom: add security vip: userid_list is empty")
	}
	var out SecurityVIPJob
	err := c.api.Call(ctx, "security add vip", "/cgi-bin/security/vip/submit_batch_add_job",
		map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// AddSecurityVIPResult reads the outcome of a grant job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99503
func (c *Security) AddSecurityVIPResult(ctx context.Context, jobID string) (SecurityVIPResult, error) {
	return c.vipResult(ctx, "security add vip result", "/cgi-bin/security/vip/batch_add_job_result", jobID)
}

// SubmitRemoveSecurityVIP starts revoking security professional accounts.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99505
func (c *Security) SubmitRemoveSecurityVIP(ctx context.Context, userIDs []string) (SecurityVIPJob, error) {
	if err := check(c); err != nil {
		return SecurityVIPJob{}, err
	}
	if len(userIDs) == 0 {
		return SecurityVIPJob{}, errors.New("wecom: remove security vip: userid_list is empty")
	}
	var out SecurityVIPJob
	err := c.api.Call(ctx, "security remove vip", "/cgi-bin/security/vip/submit_batch_del_job",
		map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// RemoveSecurityVIPResult reads the outcome of a revoke job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99505
func (c *Security) RemoveSecurityVIPResult(ctx context.Context, jobID string) (SecurityVIPResult, error) {
	return c.vipResult(ctx, "security remove vip result", "/cgi-bin/security/vip/batch_del_job_result", jobID)
}

func (c *Security) vipResult(ctx context.Context, op, path, jobID string) (SecurityVIPResult, error) {
	if err := check(c); err != nil {
		return SecurityVIPResult{}, err
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return SecurityVIPResult{}, errors.New("wecom: " + op + ": jobid is required")
	}
	var out SecurityVIPResult
	err := c.api.Call(ctx, op, path, map[string]string{"jobid": jobID}, &out)
	return out, err
}

// ListSecurityVIP returns one page of security professional accounts.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99506
func (c *Security) ListSecurityVIP(ctx context.Context, cursor string, limit int) (SecurityVIPPage, error) {
	if err := check(c); err != nil {
		return SecurityVIPPage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out SecurityVIPPage
	err := c.api.Call(ctx, "security list vip", "/cgi-bin/security/vip/list", request, &out)
	return out, err
}

// OperLog is one member or admin operation log entry.
type OperLog struct {
	Time       int64  `json:"time"`
	UserID     string `json:"userid"`
	OperType   int    `json:"oper_type"`
	DetailType int    `json:"detail_type,omitempty"`
	DetailInfo string `json:"detail_info,omitempty"`
	IP         string `json:"ip,omitempty"`
}

// OperLogPage is one page of operation logs.
type OperLogPage struct {
	HasMore    bool      `json:"has_more"`
	NextCursor string    `json:"next_cursor"`
	Records    []OperLog `json:"record_list"`
}

// MemberOperLogs returns member operation logs.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100178
func (c *Security) MemberOperLogs(ctx context.Context, startTime, endTime int64, operType int, userID, cursor string, limit int) (OperLogPage, error) {
	return c.operLogs(ctx, "security member logs", "/cgi-bin/security/member_oper_log/list",
		startTime, endTime, operType, userID, cursor, limit)
}

// AdminOperLogs returns admin console operation logs.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100179
func (c *Security) AdminOperLogs(ctx context.Context, startTime, endTime int64, operType int, userID, cursor string, limit int) (OperLogPage, error) {
	return c.operLogs(ctx, "security admin logs", "/cgi-bin/security/admin_oper_log/list",
		startTime, endTime, operType, userID, cursor, limit)
}

func (c *Security) operLogs(ctx context.Context, op, path string, startTime, endTime int64, operType int, userID, cursor string, limit int) (OperLogPage, error) {
	if err := check(c); err != nil {
		return OperLogPage{}, err
	}
	if err := checkWindow(startTime, endTime, op); err != nil {
		return OperLogPage{}, err
	}
	request := struct {
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
		OperType  int    `json:"oper_type,omitempty"`
		UserID    string `json:"userid,omitempty"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{startTime, endTime, operType, userID, cursor, limit}
	var out OperLogPage
	err := c.api.Call(ctx, op, path, request, &out)
	return out, err
}

// Domain is one WeCom server domain and its ports.
type Domain struct {
	Domain      string `json:"domain"`
	Universal   string `json:"universal_domian,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	Ports       []int  `json:"port,omitempty"`
	Necessary   int    `json:"is_necessary,omitempty"`
	Description string `json:"description,omitempty"`
}

// Domains returns the WeCom server domains and ports.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100079
func (c *Security) Domains(ctx context.Context) ([]Domain, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		List []Domain `json:"domain_list"`
	}
	err := c.api.Call(ctx, "security domains", "/cgi-bin/security/get_server_domain_ip", nil, &out)
	return out.List, err
}

// Event is one decrypted security callback.
type Event struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	ChangeType string `xml:"ChangeType"`
}

// ParseEvent decodes a decrypted security callback. Decrypt the body with
// callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100080
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: security event: " + err.Error())
	}
	if event.Name != "security" || event.ChangeType != "change_domain_ip" {
		return Event{}, errors.New("wecom: security event: expected security/change_domain_ip")
	}
	return event, nil
}

func checkWindow(startTime, endTime int64, op string) error {
	if startTime <= 0 || endTime <= startTime {
		return errors.New("wecom: " + op + ": a valid time window is required")
	}
	return nil
}

func check(c *Security) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: security client is unavailable")
	}
	return nil
}
