// Package approval implements the WeCom approval API (/cgi-bin/oa/approval and
// /cgi-bin/oa/applyevent): templates, submissions, queries, vacation quota and
// the approval change callback.
package approval

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/url"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Approval manages approval templates and requests with one shared transport.
type Approval struct{ api *client.Client }

// New creates an approval client.
func New(api *client.Client) *Approval { return &Approval{api: api} }

// Text is one localized string of a template.
type Text struct {
	Text string `json:"text"`
	Lang string `json:"lang"`
}

// Template is one approval template. Controls stay raw because every control
// type carries its own property block.
type Template struct {
	Names   []Text          `json:"template_names"`
	Content json.RawMessage `json:"template_content"`
}

// Filter narrows an approval query by template, creator or department.
type Filter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Query selects approval numbers.
type Query struct {
	StartTime int64    `json:"starttime"`
	EndTime   int64    `json:"endtime"`
	Cursor    string   `json:"new_cursor"`
	Size      int      `json:"size,omitempty"`
	Filters   []Filter `json:"filters,omitempty"`
}

// SummaryItem is one row of the submitter-facing summary.
type SummaryItem struct {
	SummaryInfo []Text `json:"summary_info"`
}

// ApplyRequest submits one approval. Process and ApplyData follow the official
// control structures and are passed through.
type ApplyRequest struct {
	CreatorUserID       string          `json:"creator_userid"`
	TemplateID          string          `json:"template_id"`
	UseTemplateApprover int             `json:"use_template_approver,omitempty"`
	ChooseDepartment    int             `json:"choose_department,omitempty"`
	Process             json.RawMessage `json:"process,omitempty"`
	ApplyData           json.RawMessage `json:"apply_data"`
	SummaryList         []SummaryItem   `json:"summary_list,omitempty"`
}

// Applier identifies the member who submitted an approval.
type Applier struct {
	UserID  string `json:"userid,omitempty"`
	PartyID string `json:"partyid,omitempty"`
}

// Detail is one approval request. Records and detail blocks stay raw because the
// shape depends on the template controls.
type Detail struct {
	SPNo       string          `json:"sp_no"`
	SPName     string          `json:"sp_name"`
	SPStatus   int             `json:"sp_status"`
	TemplateID string          `json:"template_id"`
	ApplyTime  int64           `json:"apply_time"`
	Applier    Applier         `json:"applyer"`
	Records    json.RawMessage `json:"sp_record"`
}

// VacationQuota is one vacation balance of a member.
type VacationQuota struct {
	VacationID   int    `json:"id"`
	Name         string `json:"vacationname"`
	AssignedTime int    `json:"assignduration"`
	UsedTime     int    `json:"usedduration"`
	LeftTime     int    `json:"leftduration"`
	RealAssigned int    `json:"real_assignduration,omitempty"`
}

// SetQuota changes the balance of one vacation type for one member.
type SetQuota struct {
	UserID     string `json:"userid"`
	VacationID int    `json:"vacation_id"`
	LeftTime   int    `json:"leftduration"`
	TimeAttr   int    `json:"time_attr,omitempty"`
	Remarks    string `json:"remarks,omitempty"`
}

// GetTemplate reads one approval template.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91982
func (c *Approval) GetTemplate(ctx context.Context, templateID string) (Template, error) {
	if err := check(c); err != nil {
		return Template{}, err
	}
	templateID = strings.TrimSpace(templateID)
	if templateID == "" {
		return Template{}, errors.New("wecom: approval template: template_id is required")
	}
	var out Template
	err := c.api.Call(ctx, "approval template", "/cgi-bin/oa/gettemplatedetail",
		map[string]string{"template_id": templateID}, &out)
	return out, err
}

// Submit creates one approval request and returns its approval number.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91853
func (c *Approval) Submit(ctx context.Context, request ApplyRequest) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(request.CreatorUserID) == "" || strings.TrimSpace(request.TemplateID) == "" {
		return "", errors.New("wecom: approval submit: creator_userid and template_id are required")
	}
	if len(request.ApplyData) == 0 || !json.Valid(request.ApplyData) {
		return "", errors.New("wecom: approval submit: apply_data must be valid JSON")
	}
	var out struct {
		SPNo string `json:"sp_no"`
	}
	err := c.api.Call(ctx, "approval submit", "/cgi-bin/oa/applyevent", request, &out)
	return out.SPNo, err
}

// List returns one page of approval numbers and the next cursor.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91816
func (c *Approval) List(ctx context.Context, query Query) ([]string, string, error) {
	if err := check(c); err != nil {
		return nil, "", err
	}
	if query.StartTime <= 0 || query.EndTime <= query.StartTime {
		return nil, "", errors.New("wecom: approval list: a valid time window is required")
	}
	var out struct {
		Numbers []string `json:"sp_no_list"`
		Cursor  string   `json:"new_next_cursor"`
	}
	err := c.api.Call(ctx, "approval list", "/cgi-bin/oa/getapprovalinfo", query, &out)
	return out.Numbers, out.Cursor, err
}

// GetDetail reads one approval request.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91983
func (c *Approval) GetDetail(ctx context.Context, spNo string) (Detail, error) {
	if err := check(c); err != nil {
		return Detail{}, err
	}
	spNo = strings.TrimSpace(spNo)
	if spNo == "" {
		return Detail{}, errors.New("wecom: approval detail: sp_no is required")
	}
	var out struct {
		Info Detail `json:"info"`
	}
	err := c.api.Call(ctx, "approval detail", "/cgi-bin/oa/getapprovaldetail",
		map[string]string{"sp_no": spNo}, &out)
	return out.Info, err
}

// LegacyData returns the deprecated approval export. Entries stay raw because the
// endpoint predates the current template model.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91530
func (c *Approval) LegacyData(ctx context.Context, startTime, endTime int64, nextSPNum int64) ([]json.RawMessage, int64, error) {
	if err := check(c); err != nil {
		return nil, 0, err
	}
	if startTime <= 0 || endTime <= startTime {
		return nil, 0, errors.New("wecom: approval legacy data: a valid time window is required")
	}
	query := url.Values{
		"starttime": {itoa64(startTime)},
		"endtime":   {itoa64(endTime)},
	}
	if nextSPNum > 0 {
		query.Set("next_spnum", itoa64(nextSPNum))
	}
	var out struct {
		Data      []json.RawMessage `json:"data"`
		NextSPNum int64             `json:"next_spnum"`
	}
	err := c.api.Get(ctx, "approval legacy data", "/cgi-bin/corp/getapprovaldata", query, &out)
	return out.Data, out.NextSPNum, err
}

// OpenApprovalData returns the open-platform approval payload of one request.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90269
func (c *Approval) OpenApprovalData(ctx context.Context, thirdNo string) (json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	thirdNo = strings.TrimSpace(thirdNo)
	if thirdNo == "" {
		return nil, errors.New("wecom: open approval data: thirdNo is required")
	}
	var out struct {
		Data json.RawMessage `json:"data"`
	}
	err := c.api.Get(ctx, "open approval data", "/cgi-bin/corp/getopenapprovaldata",
		url.Values{"thirdNo": {thirdNo}}, &out)
	return out.Data, err
}

// VacationConfig returns the enterprise vacation configuration. Entries stay raw
// because quota rules vary per vacation type.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93375
func (c *Approval) VacationConfig(ctx context.Context) ([]json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		Lists []json.RawMessage `json:"lists"`
	}
	err := c.api.Call(ctx, "approval vacation config", "/cgi-bin/oa/vacation/getcorpconf", nil, &out)
	return out.Lists, err
}

// VacationQuota returns the vacation balances of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93376
func (c *Approval) VacationQuota(ctx context.Context, userID string) ([]VacationQuota, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("wecom: vacation quota: userid is required")
	}
	var out struct {
		Lists []VacationQuota `json:"lists"`
	}
	err := c.api.Call(ctx, "approval vacation quota", "/cgi-bin/oa/vacation/getuservacationquota",
		map[string]string{"userid": userID}, &out)
	return out.Lists, err
}

// SetVacationQuota changes one vacation balance of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93377
func (c *Approval) SetVacationQuota(ctx context.Context, quota SetQuota) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(quota.UserID) == "" || quota.VacationID <= 0 {
		return errors.New("wecom: set vacation quota: userid and vacation_id are required")
	}
	return c.api.Call(ctx, "approval set vacation quota",
		"/cgi-bin/oa/vacation/setoneuserquota", quota, nil)
}

// CreateTemplate creates one approval template and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97437
func (c *Approval) CreateTemplate(ctx context.Context, template Template) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if len(template.Names) == 0 || len(template.Content) == 0 {
		return "", errors.New("wecom: create template: template_name and template_content are required")
	}
	var out struct {
		TemplateID string `json:"template_id"`
	}
	err := c.api.Call(ctx, "approval create template", "/cgi-bin/oa/approval/create_template",
		template, &out)
	return out.TemplateID, err
}

// UpdateTemplate replaces one approval template.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97438
func (c *Approval) UpdateTemplate(ctx context.Context, templateID string, template Template) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(templateID) == "" || len(template.Content) == 0 {
		return errors.New("wecom: update template: template_id and template_content are required")
	}
	request := struct {
		TemplateID string          `json:"template_id"`
		Names      []Text          `json:"template_name,omitempty"`
		Content    json.RawMessage `json:"template_content"`
		Template
	}{TemplateID: templateID, Names: template.Names, Content: template.Content}
	return c.api.Call(ctx, "approval update template", "/cgi-bin/oa/approval/update_template", request, nil)
}

// EventApprovalChange is the approval change callback name.
const EventApprovalChange = "sys_approval_change"

// Approver identifies one approver of a callback.
type Approver struct {
	UserID string `xml:"UserId"`
}

// ApprovalStep is one approver result inside a callback.
type ApprovalStep struct {
	Approver *Approver `xml:"Approver"`
	Speech   string    `xml:"Speech"`
	Status   int       `xml:"SpStatus"`
	Time     int64     `xml:"SpTime"`
}

// ApprovalRecord is one approval node of a callback.
type ApprovalRecord struct {
	Status       int            `xml:"SpStatus"`
	ApproverAttr int            `xml:"ApproverAttr"`
	Details      []ApprovalStep `xml:"Details"`
}

// ApprovalInfo is the approval block of a callback.
type ApprovalInfo struct {
	SPNoStr    string `xml:"SpNoStr"`
	SPNo       string `xml:"SpNo"`
	SPName     string `xml:"SpName"`
	Status     int    `xml:"SpStatus"`
	TemplateID string `xml:"TemplateId"`
	ApplyTime  int64  `xml:"ApplyTime"`
	Applier    struct {
		UserID string `xml:"UserId"`
		Party  string `xml:"Party"`
	} `xml:"Applyer"`
	Records []ApprovalRecord `xml:"SpRecord"`
}

// Event is one decrypted sys_approval_change callback.
type Event struct {
	ToUser     string       `xml:"ToUserName"`
	FromUser   string       `xml:"FromUserName"`
	CreateTime int64        `xml:"CreateTime"`
	MsgType    string       `xml:"MsgType"`
	Name       string       `xml:"Event"`
	AgentID    int          `xml:"AgentID"`
	Info       ApprovalInfo `xml:"ApprovalInfo"`
}

// ParseEvent decodes a decrypted sys_approval_change callback. Decrypt the body
// with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91815
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: approval event: " + err.Error())
	}
	if event.Name != EventApprovalChange {
		return Event{}, errors.New("wecom: approval event: unexpected event " + event.Name)
	}
	if strings.TrimSpace(event.Info.SPNo) == "" {
		return Event{}, errors.New("wecom: approval event: SpNo is required")
	}
	return event, nil
}

func check(c *Approval) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: approval client is unavailable")
	}
	return nil
}

func itoa(v int) string { return itoa64(int64(v)) }

func itoa64(v int64) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var digits []byte
	for v > 0 {
		digits = append([]byte{byte('0' + v%10)}, digits...)
		v /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
