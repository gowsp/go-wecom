// Package advanced implements the WeCom advanced feature approval API
// (/cgi-bin/advanced_feature): reporting approval progress and reading the
// pending applications, plus the related callbacks.
package advanced

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Advanced manages advanced feature applications with one shared transport.
type Advanced struct{ api *client.Client }

// New creates an advanced feature client.
func New(api *client.Client) *Advanced { return &Advanced{api: api} }

// Node is one approval node of the enterprise process.
type Node struct {
	CurrentApprovers   []string `json:"current_approvers,omitempty"`
	CompletedApprovers []string `json:"completed_approvers,omitempty"`
	Status             int      `json:"node_apv_status,omitempty"`
	Relation           int      `json:"node_apv_rel,omitempty"`
	UpdateTime         int64    `json:"apv_update_time,omitempty"`
}

// ProcessList is the approval flow reported back to WeCom.
type ProcessList struct {
	NodeList []Node `json:"node_list"`
}

// ApprovalDetail reports the approval state of one application.
type ApprovalDetail struct {
	ApprovalID     string      `json:"approval_id"`
	ApprovalStatus int         `json:"approval_status"`
	ApplyID        string      `json:"apply_id"`
	ApprovalURL    string      `json:"approval_url,omitempty"`
	ProcessList    ProcessList `json:"process_list"`
}

// ApplyQuery selects advanced feature applications.
type ApplyQuery struct {
	BusinessType int    `json:"business_type"`
	Limit        int    `json:"limit,omitempty"`
	Cursor       string `json:"cursor,omitempty"`
	ReqType      int    `json:"req_type,omitempty"`
}

// ApplyPage is one page of application ids.
type ApplyPage struct {
	NextCursor string   `json:"next_cursor"`
	ApplyIDs   []string `json:"apply_id_list"`
	HasMore    bool     `json:"has_more"`
}

// ApprovalInfo is the detail of one application.
type ApprovalInfo struct {
	Applicant      string `json:"applicant"`
	CreateTime     int64  `json:"create_time"`
	BusinessType   int    `json:"business_type"`
	ApprovalID     string `json:"approval_id"`
	ApplyID        string `json:"apply_id"`
	ApprovalURL    string `json:"approval_url,omitempty"`
	ApprovalStatus int    `json:"approval_status"`
	ApprovalType   int    `json:"approval_type,omitempty"`
	RequestReason  string `json:"request_reason,omitempty"`
}

// SetApprovalDetail reports the approval progress of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99880
func (c *Advanced) SetApprovalDetail(ctx context.Context, detail ApprovalDetail) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(detail.ApprovalID) == "" || strings.TrimSpace(detail.ApplyID) == "" {
		return errors.New("wecom: approval detail: approval_id and apply_id are required")
	}
	return c.api.Call(ctx, "advanced approval detail",
		"/cgi-bin/advanced_feature/set_approval_detail", detail, nil)
}

// ApplyIDs returns one page of application ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99883
func (c *Advanced) ApplyIDs(ctx context.Context, query ApplyQuery) (ApplyPage, error) {
	if err := check(c); err != nil {
		return ApplyPage{}, err
	}
	if query.BusinessType <= 0 {
		return ApplyPage{}, errors.New("wecom: apply ids: business_type is required")
	}
	var out ApplyPage
	err := c.api.Call(ctx, "advanced apply ids",
		"/cgi-bin/advanced_feature/get_apply_id_list", query, &out)
	return out, err
}

// Approval reads one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99885
func (c *Advanced) GetApproval(ctx context.Context, applyID string) (ApprovalInfo, error) {
	if err := check(c); err != nil {
		return ApprovalInfo{}, err
	}
	applyID = strings.TrimSpace(applyID)
	if applyID == "" {
		return ApprovalInfo{}, errors.New("wecom: approval: apply_id is required")
	}
	var out struct {
		Info ApprovalInfo `json:"approval_info"`
	}
	err := c.api.Call(ctx, "advanced approval", "/cgi-bin/advanced_feature/get_approval_info",
		map[string]string{"apply_id": applyID}, &out)
	return out.Info, err
}

// SubmitEvent is the callback that announces a new member application.
type SubmitEvent struct {
	ToUser       string `xml:"ToUserName"`
	FromUser     string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"`
	Name         string `xml:"Event"`
	ApplyReason  string `xml:"ApplyReason"`
	BusinessType int    `xml:"BusinessType"`
	ApplyID      string `xml:"ApplyId"`
	AgentID      int    `xml:"AgentID"`
}

// FinishEvent is the callback that announces a finished application.
type FinishEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUser"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	FinishType string `xml:"FinishType"`
	ApplyID    string `xml:"ApplyId"`
	AgentID    int    `xml:"AgentID"`
}

// ParseSubmitEvent decodes a decrypted submit_vip_account_approval callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99876
func ParseSubmitEvent(data []byte) (SubmitEvent, error) {
	var event SubmitEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return SubmitEvent{}, errors.New("wecom: advanced submit event: " + err.Error())
	}
	if event.Name != "submit_vip_account_approval" {
		return SubmitEvent{}, errors.New("wecom: advanced submit event: unexpected event " + event.Name)
	}
	if strings.TrimSpace(event.ApplyID) == "" {
		return SubmitEvent{}, errors.New("wecom: advanced submit event: ApplyId is required")
	}
	return event, nil
}

// ParseFinishEvent decodes a decrypted finish_vip_account_approval callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99877
func ParseFinishEvent(data []byte) (FinishEvent, error) {
	var event FinishEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return FinishEvent{}, errors.New("wecom: advanced finish event: " + err.Error())
	}
	if event.Name != "finish_vip_account_approval" {
		return FinishEvent{}, errors.New("wecom: advanced finish event: unexpected event " + event.Name)
	}
	if strings.TrimSpace(event.ApplyID) == "" {
		return FinishEvent{}, errors.New("wecom: advanced finish event: ApplyId is required")
	}
	return event, nil
}

func check(c *Advanced) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: advanced client is unavailable")
	}
	return nil
}
