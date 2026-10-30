package meeting

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

// EnrollConfig is the registration configuration of one meeting. Questions stay
// raw because their option structure depends on question_type.
type EnrollConfig struct {
	MeetingID              string          `json:"meetingid"`
	ApproveType            int             `json:"approve_type,omitempty"`
	IsCollectQuestion      int             `json:"is_collect_question,omitempty"`
	NoRegistrationForStaff *bool           `json:"no_registration_needed_for_staff,omitempty"`
	QuestionList           json.RawMessage `json:"question_list,omitempty"`
}

// EnrollQuestionCount is the result of setting a registration configuration.
type EnrollQuestionCount struct {
	QuestionCount int `json:"question_count"`
}

// SetEnrollConfig writes the registration configuration of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98797
func (c *Meeting) SetEnrollConfig(ctx context.Context, config EnrollConfig) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(config.MeetingID) == "" {
		return 0, errors.New("wecom: enroll config: meetingid is required")
	}
	var out EnrollQuestionCount
	err := c.api.Call(ctx, "meeting enroll config", "/cgi-bin/meeting/enroll/set_config", config, &out)
	return out.QuestionCount, err
}

// GetEnrollConfig reads the registration configuration of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98800
func (c *Meeting) GetEnrollConfig(ctx context.Context, meetingID string) (EnrollConfig, error) {
	if err := check(c); err != nil {
		return EnrollConfig{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return EnrollConfig{}, errors.New("wecom: get enroll config: meetingid is required")
	}
	var out EnrollConfig
	err := c.api.Call(ctx, "meeting get enroll config", "/cgi-bin/meeting/enroll/get_config",
		map[string]string{"meetingid": meetingID}, &out)
	out.MeetingID = meetingID
	return out, err
}

// EnrollID pairs a temporary open id with its registration id.
type EnrollID struct {
	TmpOpenID string `json:"tmp_openid"`
	EnrollID  string `json:"enroll_id"`
}

// EnrollIDs resolves registration ids for temporary open ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98794
func (c *Meeting) EnrollIDs(ctx context.Context, meetingID string, sortingRules int, tmpOpenIDs []string) ([]EnrollID, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(tmpOpenIDs) == 0 {
		return nil, errors.New("wecom: enroll ids: meetingid and tmp_openid_list are required")
	}
	request := struct {
		MeetingID     string   `json:"meetingid"`
		SortingRules  int      `json:"sorting_rules,omitempty"`
		TmpOpenIDList []string `json:"tmp_openid_list"`
	}{MeetingID: meetingID, SortingRules: sortingRules, TmpOpenIDList: tmpOpenIDs}
	var out struct {
		List []EnrollID `json:"enroll_id_list"`
	}
	err := c.api.Call(ctx, "meeting enroll ids", "/cgi-bin/meeting/enroll/query_by_tmp_openid", request, &out)
	return out.List, err
}

// Enrollment is one registration record.
type Enrollment struct {
	EnrollID     string `json:"enroll_id"`
	EnrollTime   string `json:"enroll_time,omitempty"`
	EnrollSource int    `json:"enroll_source_type,omitempty"`
	NickName     string `json:"nick_name,omitempty"`
	Status       int    `json:"status,omitempty"`
	UserID       string `json:"userid,omitempty"`
	TmpOpenID    string `json:"tmp_openid,omitempty"`
	EnrollCode   string `json:"enroll_code,omitempty"`
	Area         string `json:"area,omitempty"`
	PhoneNumber  string `json:"phone_number,omitempty"`
}

// EnrollPage is one page of registrations.
type EnrollPage struct {
	HasMore    bool         `json:"has_more"`
	NextCursor string       `json:"next_cursor"`
	List       []Enrollment `json:"enroll_list"`
}

// Enrollments returns one page of registrations filtered by status.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98810
func (c *Meeting) Enrollments(ctx context.Context, meetingID string, status int, cursor string, limit int) (EnrollPage, error) {
	if err := check(c); err != nil {
		return EnrollPage{}, err
	}
	meetingID = strings.TrimSpace(meetingID)
	if meetingID == "" {
		return EnrollPage{}, errors.New("wecom: enrollments: meetingid is required")
	}
	request := struct {
		MeetingID string `json:"meetingid"`
		Status    int    `json:"status,omitempty"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{MeetingID: meetingID, Status: status, Cursor: cursor, Limit: limit}
	var out EnrollPage
	err := c.api.Call(ctx, "meeting enrollments", "/cgi-bin/meeting/enroll/list", request, &out)
	return out, err
}

// ApproveEnrollments approves or rejects registrations. action 1 approves and 2
// rejects.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98807
func (c *Meeting) ApproveEnrollments(ctx context.Context, meetingID string, action int, enrollIDs []string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(meetingID) == "" || len(enrollIDs) == 0 {
		return 0, errors.New("wecom: approve enrollments: meetingid and enroll_id_list are required")
	}
	request := struct {
		MeetingID string   `json:"meetingid"`
		Action    int      `json:"action"`
		EnrollIDs []string `json:"enroll_id_list"`
	}{MeetingID: meetingID, Action: action, EnrollIDs: enrollIDs}
	var out struct {
		Handled int `json:"handled_count"`
	}
	err := c.api.Call(ctx, "meeting approve enrollments", "/cgi-bin/meeting/enroll/approve", request, &out)
	return out.Handled, err
}

// ImportEnrollment is one registration imported from an external system.
type ImportEnrollment struct {
	UserID      string `json:"userid,omitempty"`
	Area        string `json:"area,omitempty"`
	PhoneNumber string `json:"phone_number,omitempty"`
	NickName    string `json:"nick_name,omitempty"`
}

// ImportEnrollments adds registrations from an external system.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98816
func (c *Meeting) ImportEnrollments(ctx context.Context, meetingID string, list []ImportEnrollment) (int, []Enrollment, error) {
	if err := check(c); err != nil {
		return 0, nil, err
	}
	if strings.TrimSpace(meetingID) == "" || len(list) == 0 {
		return 0, nil, errors.New("wecom: import enrollments: meetingid and enroll_list are required")
	}
	request := struct {
		MeetingID string             `json:"meetingid"`
		List      []ImportEnrollment `json:"enroll_list"`
	}{MeetingID: meetingID, List: list}
	var out struct {
		Total int          `json:"total_count"`
		List  []Enrollment `json:"enroll_list"`
	}
	err := c.api.Call(ctx, "meeting import enrollments", "/cgi-bin/meeting/enroll/import", request, &out)
	return out.Total, out.List, err
}

// DeleteEnrollments removes registrations by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98817
func (c *Meeting) DeleteEnrollments(ctx context.Context, meetingID string, enrollIDs []string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	if strings.TrimSpace(meetingID) == "" || len(enrollIDs) == 0 {
		return 0, errors.New("wecom: delete enrollments: meetingid and enroll_id_list are required")
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
	err := c.api.Call(ctx, "meeting delete enrollments", "/cgi-bin/meeting/enroll/delete", request, &out)
	return out.Total, err
}
