package meeting

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"
)

// WebinarGuest is one guest of a webinar. guest_type 1 is an internal member, 2 a
// phone guest and 3 an email guest.
type WebinarGuest struct {
	GuestType   int    `json:"guest_type,omitempty"`
	UserID      string `json:"userid,omitempty"`
	Area        any    `json:"area,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	GuestName   string `json:"guest_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

// WebinarHost is one host of a webinar.
type WebinarHost struct {
	UserID string `json:"userid"`
}

// WebinarRequest creates or updates one webinar.
type WebinarRequest struct {
	AdminUserID          string          `json:"admin_userid"`
	MeetingID            string          `json:"meetingid,omitempty"`
	Title                string          `json:"title,omitempty"`
	Sponsor              string          `json:"sponsor,omitempty"`
	StartTime            string          `json:"start_time,omitempty"`
	EndTime              string          `json:"end_time,omitempty"`
	AdmissionType        int             `json:"admission_type,omitempty"`
	Hosts                []WebinarHost   `json:"hosts,omitempty"`
	Password             string          `json:"password,omitempty"`
	CoverURL             string          `json:"cover_url,omitempty"`
	Description          string          `json:"description,omitempty"`
	EnableGuestInviteURL *bool           `json:"enable_guest_invite_link,omitempty"`
	MediaSetting         json.RawMessage `json:"media_setting,omitempty"`
}

// Webinar is one webinar with its join links.
type Webinar struct {
	MeetingID           string          `json:"meetingid"`
	MeetingCode         string          `json:"meeting_code,omitempty"`
	Title               string          `json:"title,omitempty"`
	Sponsor             string          `json:"sponsor,omitempty"`
	StartTime           string          `json:"start_time,omitempty"`
	EndTime             string          `json:"end_time,omitempty"`
	AdmissionType       int             `json:"admission_type,omitempty"`
	Hosts               []WebinarHost   `json:"hosts,omitempty"`
	Password            string          `json:"password,omitempty"`
	CoverURL            string          `json:"cover_url,omitempty"`
	Description         string          `json:"description,omitempty"`
	AudienceJoinLink    string          `json:"audience_join_link,omitempty"`
	GuestJoinLink       string          `json:"guest_join_link,omitempty"`
	ManualCheckLink     string          `json:"manual_check_link,omitempty"`
	ManualCheckPassword string          `json:"manual_check_password,omitempty"`
	MediaSetting        json.RawMessage `json:"media_setting,omitempty"`
}

// CreateWebinar creates one webinar and returns it with its join links.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98842
func (c *Meeting) CreateWebinar(ctx context.Context, request WebinarRequest) (Webinar, error) {
	if err := check(c); err != nil {
		return Webinar{}, err
	}
	if strings.TrimSpace(request.AdminUserID) == "" || strings.TrimSpace(request.Title) == "" {
		return Webinar{}, errors.New("wecom: create webinar: admin_userid and title are required")
	}
	if request.StartTime == "" || request.EndTime == "" {
		return Webinar{}, errors.New("wecom: create webinar: start_time and end_time are required")
	}
	var out Webinar
	err := c.api.Call(ctx, "webinar create", "/cgi-bin/meeting/webinar/create", request, &out)
	return out, err
}

// UpdateWebinar changes one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98843
func (c *Meeting) UpdateWebinar(ctx context.Context, request WebinarRequest) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(request.MeetingID) == "" {
		return errors.New("wecom: update webinar: meetingid is required")
	}
	return c.api.Call(ctx, "webinar update", "/cgi-bin/meeting/webinar/update", request, nil)
}

// CancelWebinar cancels one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98870
func (c *Meeting) CancelWebinar(ctx context.Context, meetingID string) error {
	if err := check(c); err != nil {
		return err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return errors.New("wecom: cancel webinar: meetingid is required")
	}
	return c.api.Call(ctx, "webinar cancel", "/cgi-bin/meeting/webinar/cancel",
		map[string]string{"meetingid": meetingID}, nil)
}

// Webinar returns one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98860
func (c *Meeting) Webinar(ctx context.Context, meetingID string) (Webinar, error) {
	if err := check(c); err != nil {
		return Webinar{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return Webinar{}, errors.New("wecom: webinar: meetingid is required")
	}
	var out Webinar
	err := c.api.Call(ctx, "webinar get", "/cgi-bin/meeting/webinar/get",
		map[string]string{"meetingid": meetingID}, &out)
	return out, err
}

// WebinarGuests returns the guest list of one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98871
func (c *Meeting) WebinarGuests(ctx context.Context, meetingID, meetingCode string) ([]WebinarGuest, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" && strings.TrimSpace(meetingCode) == "" {
		return nil, errors.New("wecom: webinar guests: meetingid or meeting_code is required")
	}
	request := struct {
		MeetingID   string `json:"meetingid,omitempty"`
		MeetingCode string `json:"meeting_code,omitempty"`
	}{MeetingID: meetingID, MeetingCode: meetingCode}
	var out struct {
		Guests []WebinarGuest `json:"guests"`
	}
	err := c.api.Call(ctx, "webinar guests", "/cgi-bin/meeting/webinar/list_guest", request, &out)
	return out.Guests, err
}

// SetWebinarGuests replaces the guest list of one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98872
func (c *Meeting) SetWebinarGuests(ctx context.Context, meetingID string, guests []WebinarGuest) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || len(guests) == 0 {
		return errors.New("wecom: set webinar guests: meetingid and guests are required")
	}
	request := struct {
		MeetingID string         `json:"meetingid"`
		Guests    []WebinarGuest `json:"guests"`
	}{MeetingID: meetingID, Guests: guests}
	return c.api.Call(ctx, "webinar set guests", "/cgi-bin/meeting/webinar/update_guest_list", request, nil)
}

// WarmUp is the warm-up media configuration of a webinar.
type WarmUp struct {
	MeetingID            string `json:"meetingid"`
	Picture              string `json:"warm_up_picture,omitempty"`
	Video                string `json:"warm_up_video,omitempty"`
	AllowAttendeesInvite *bool  `json:"allow_attendees_invite_others,omitempty"`
}

// SetWebinarWarmUp updates the warm-up media of one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98882
func (c *Meeting) SetWebinarWarmUp(ctx context.Context, warmUp WarmUp) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(warmUp.MeetingID) == "" {
		return errors.New("wecom: webinar warm up: meetingid is required")
	}
	return c.api.Call(ctx, "webinar warm up", "/cgi-bin/meeting/webinar/update_warm_up", warmUp, nil)
}

// WebinarEnrollConfig is the registration configuration of one webinar.
type WebinarEnrollConfig struct {
	MeetingID              string          `json:"meetingid"`
	ApproveType            int             `json:"approve_type,omitempty"`
	IsCollectQuestion      int             `json:"is_collect_question,omitempty"`
	NoRegistrationForStaff *bool           `json:"no_registration_needed_for_staff,omitempty"`
	QuestionList           json.RawMessage `json:"question_list,omitempty"`
}

// SetWebinarEnrollConfig writes the registration configuration of one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98875
func (c *Meeting) SetWebinarEnrollConfig(ctx context.Context, config WebinarEnrollConfig) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(config.MeetingID) == "" {
		return 0, errors.New("wecom: webinar enroll config: meetingid is required")
	}
	var out EnrollQuestionCount
	err := c.api.Call(ctx, "webinar enroll config",
		"/cgi-bin/meeting/webinar/enroll/set_config", config, &out)
	return out.QuestionCount, err
}

// WebinarEnrollConfig reads the registration configuration of one webinar.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98874
func (c *Meeting) GetWebinarEnrollConfig(ctx context.Context, meetingID string) (WebinarEnrollConfig, error) {
	if err := check(c); err != nil {
		return WebinarEnrollConfig{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return WebinarEnrollConfig{}, errors.New("wecom: get webinar enroll config: meetingid is required")
	}
	var out WebinarEnrollConfig
	err := c.api.Call(ctx, "webinar get enroll config",
		"/cgi-bin/meeting/webinar/enroll/get_config", map[string]string{"meetingid": meetingID}, &out)
	out.MeetingID = meetingID
	return out, err
}

// WebinarEnrollIDs resolves registration ids for temporary open ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98873
func (c *Meeting) WebinarEnrollIDs(ctx context.Context, meetingID string, sortingRules int, tmpOpenIDs []string) ([]EnrollID, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(tmpOpenIDs) == 0 {
		return nil, errors.New("wecom: webinar enroll ids: meetingid and tmp_openid_list are required")
	}
	request := struct {
		MeetingID     string   `json:"meetingid"`
		SortingRules  int      `json:"sorting_rules,omitempty"`
		TmpOpenIDList []string `json:"tmp_openid_list"`
	}{MeetingID: meetingID, SortingRules: sortingRules, TmpOpenIDList: tmpOpenIDs}
	var out struct {
		List []EnrollID `json:"enroll_id_list"`
	}
	err := c.api.Call(ctx, "webinar enroll ids",
		"/cgi-bin/meeting/webinar/enroll/query_by_tmp_openid", request, &out)
	return out.List, err
}

// WebinarEnrollments returns one page of webinar registrations.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98876
func (c *Meeting) WebinarEnrollments(ctx context.Context, meetingID string, status int, cursor string, limit int) (EnrollPage, error) {
	if err := check(c); err != nil {
		return EnrollPage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return EnrollPage{}, errors.New("wecom: webinar enrollments: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Status    int    `json:"status,omitempty"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, Status: status, Cursor: cursor, Limit: limit}
	var out EnrollPage
	err := c.api.Call(ctx, "webinar enrollments", "/cgi-bin/meeting/webinar/enroll/list", request, &out)
	return out, err
}

// ApproveWebinarEnrollments approves or rejects webinar registrations.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98877
func (c *Meeting) ApproveWebinarEnrollments(ctx context.Context, meetingID string, action int, enrollIDs []string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(meetingID) == "" || len(enrollIDs) == 0 {
		return 0, errors.New("wecom: approve webinar enrollments: meetingid and enroll_id_list are required")
	}
	request := struct {
		MeetingID string   `json:"meetingid"`
		EnrollIDs []string `json:"enroll_id_list"`
		Action    int      `json:"action"`
	}{MeetingID: meetingID, EnrollIDs: enrollIDs, Action: action}
	var out struct {
		Handled int `json:"handled_count"`
	}
	err := c.api.Call(ctx, "webinar approve enrollments",
		"/cgi-bin/meeting/webinar/enroll/approve", request, &out)
	return out.Handled, err
}

// ImportWebinarEnrollments adds webinar registrations from an external system.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98880
func (c *Meeting) ImportWebinarEnrollments(ctx context.Context, meetingID string, list []ImportEnrollment) (int, []Enrollment, error) {
	if err := check(c); err != nil {
		return 0, nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(list) == 0 {
		return 0, nil, errors.New("wecom: import webinar enrollments: meetingid and enroll_list are required")
	}
	request := struct {
		MeetingID string             `json:"meetingid"`
		List      []ImportEnrollment `json:"enroll_list"`
	}{MeetingID: meetingID, List: list}
	var out struct {
		Total int          `json:"total_count"`
		List  []Enrollment `json:"enroll_list"`
	}
	err := c.api.Call(ctx, "webinar import enrollments",
		"/cgi-bin/meeting/webinar/enroll/import", request, &out)
	return out.Total, out.List, err
}

// DeleteWebinarEnrollments removes webinar registrations by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98881
func (c *Meeting) DeleteWebinarEnrollments(ctx context.Context, meetingID string, enrollIDs []string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(meetingID) == "" || len(enrollIDs) == 0 {
		return 0, errors.New("wecom: delete webinar enrollments: meetingid and enroll_id_list are required")
	}
	items := make([]map[string]string, 0, len(enrollIDs))
	for _, id := range enrollIDs {
		items = append(items, map[string]string{"enroll_id": id})
	}
	request := struct {
		MeetingID string              `json:"meetingid"`
		Items     []map[string]string `json:"enroll_id_list"`
	}{MeetingID: meetingID, Items: items}
	var out struct {
		Total int `json:"total_count"`
	}
	err := c.api.Call(ctx, "webinar delete enrollments",
		"/cgi-bin/meeting/webinar/enroll/delete", request, &out)
	return out.Total, err
}

// Webinar event types.
const (
	EventWebinarRoleChange   = "webinar_role_change"
	EventWebinarWarmUpUpload = "webinar_warm_up_upload"
)

// WebinarOperatedUser is the participant whose role changed.
type WebinarOperatedUser struct {
	UserID    string `xml:"UserId"`
	TmpOpenID string `xml:"TmpOpenId"`
	UserRole  int    `xml:"UserRole"`
}

// WarmUpInfo is the upload result of warm-up media.
type WarmUpInfo struct {
	Picture      string `xml:"WarmUpPicture"`
	Video        string `xml:"WarmUpVideo"`
	UploadStatus int    `xml:"UploadStatus"`
	ErrorMsg     string `xml:"ErrorMsg"`
}

// WebinarEvent is one decrypted webinar callback.
type WebinarEvent struct {
	ToUser            string               `xml:"ToUserName"`
	FromUser          string               `xml:"FromUserName"`
	FromUserTmpOpenID string               `xml:"FromUserTmpOpenId"`
	CreateTime        int64                `xml:"CreateTime"`
	MsgType           string               `xml:"MsgType"`
	Name              string               `xml:"Event"`
	ChangeType        string               `xml:"ChangeType"`
	MeetingID         string               `xml:"MeetingId"`
	OperatedUser      *WebinarOperatedUser `xml:"OperatedUser"`
	WarmUpInfo        *WarmUpInfo          `xml:"WarmUpInfo"`
}

// ParseWebinarEvent decodes a decrypted webinar callback. Decrypt the body with
// callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98771
// 文档: https://developer.work.weixin.qq.com/document/path/98773
func ParseWebinarEvent(data []byte) (WebinarEvent, error) {
	var event WebinarEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return WebinarEvent{}, errors.New("wecom: webinar event: " + err.Error())
	}
	if event.Name != "meeting_change" {
		return WebinarEvent{}, errors.New("wecom: webinar event: unexpected event " + event.Name)
	}
	switch event.ChangeType {
	case EventWebinarRoleChange, EventWebinarWarmUpUpload:
	default:
		return WebinarEvent{}, errors.New("wecom: webinar event: unexpected change type " + event.ChangeType)
	}
	if strings.TrimSpace(event.MeetingID) == "" {
		return WebinarEvent{}, errors.New("wecom: webinar event: MeetingId is required")
	}
	return event, nil
}
