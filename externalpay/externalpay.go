// Package externalpay implements the WeCom external payment API
// (/cgi-bin/externalpay and /cgi-bin/miniapppay): merchant binding, payment and
// fund-flow records, and merchant application handling.
package externalpay

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// ExternalPay reads external payment data with one shared transport.
type ExternalPay struct{ api *client.Client }

// New creates an external payment client.
func New(api *client.Client) *ExternalPay { return &ExternalPay{api: api} }

// Merchant is one bound payment merchant.
type Merchant struct {
	MchID        string `json:"mch_id"`
	BindStatus   int    `json:"bind_status,omitempty"`
	MerchantName string `json:"merchant_name,omitempty"`
	Scope        *Scope `json:"allow_use_scope,omitempty"`
}

// Scope limits which members, departments and tags may use a merchant.
type Scope struct {
	Users    []string `json:"user,omitempty"`
	PartyIDs []int    `json:"partyid,omitempty"`
	TagIDs   []int    `json:"tagid,omitempty"`
}

// Merchant reads one bound merchant and its allowed scope.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93666
func (c *ExternalPay) Merchant(ctx context.Context, mchID string) (Merchant, error) {
	if err := check(c); err != nil {
		return Merchant{}, err
	}
	mchID = strings.TrimSpace(mchID)
	if mchID == "" {
		return Merchant{}, errors.New("wecom: merchant: mch_id is required")
	}
	var out Merchant
	err := c.api.Call(ctx, "externalpay merchant", "/cgi-bin/externalpay/getmerchant",
		map[string]string{"mch_id": mchID}, &out)
	return out, err
}

// SetMerchantScope replaces the members allowed to use one merchant.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93666
func (c *ExternalPay) SetMerchantScope(ctx context.Context, mchID string, scope Scope) error {
	if err := check(c); err != nil {
		return err
	}
	mchID = strings.TrimSpace(mchID)
	if mchID == "" {
		return errors.New("wecom: merchant scope: mch_id is required")
	}
	request := struct {
		MchID string `json:"mch_id"`
		Scope Scope  `json:"allow_use_scope"`
	}{MchID: mchID, Scope: scope}
	return c.api.Call(ctx, "externalpay merchant scope",
		"/cgi-bin/externalpay/set_mch_use_scope", request, nil)
}

// Bill is one external payment bill.
type Bill struct {
	TransactionID  string `json:"transaction_id"`
	BillType       int    `json:"bill_type,omitempty"`
	TradeState     int    `json:"trade_state,omitempty"`
	PayTime        int64  `json:"pay_time,omitempty"`
	OutTradeNo     string `json:"out_trade_no,omitempty"`
	OutRefundNo    string `json:"out_refund_no,omitempty"`
	ExternalUserID string `json:"external_userid,omitempty"`
	TotalFee       int64  `json:"total_fee,omitempty"`
	PayeeUserID    string `json:"payee_userid,omitempty"`
}

// BillPage is one page of payment bills.
type BillPage struct {
	NextCursor string `json:"next_cursor"`
	Bills      []Bill `json:"bill_list"`
}

// Bills returns one page of external payment records.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93667
func (c *ExternalPay) Bills(ctx context.Context, beginTime, endTime int64, payeeUserID, cursor string, limit int) (BillPage, error) {
	if err := check(c); err != nil {
		return BillPage{}, err
	}
	if beginTime <= 0 || endTime <= beginTime {
		return BillPage{}, errors.New("wecom: bills: a valid time window is required")
	}
	request := struct {
		BeginTime   int64  `json:"begin_time"`
		EndTime     int64  `json:"end_time"`
		PayeeUserID string `json:"payee_userid,omitempty"`
		Cursor      string `json:"cursor,omitempty"`
		Limit       int    `json:"limit,omitempty"`
	}{beginTime, endTime, payeeUserID, cursor, limit}
	var out BillPage
	err := c.api.Call(ctx, "externalpay bills", "/cgi-bin/externalpay/get_bill_list", request, &out)
	return out, err
}

// PaymentOrders returns the merchant order numbers of one payment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95944
func (c *ExternalPay) PaymentOrders(ctx context.Context, paymentID string) ([]string, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	paymentID = strings.TrimSpace(paymentID)
	if paymentID == "" {
		return nil, errors.New("wecom: payment orders: payment_id is required")
	}
	var out struct {
		Bills []struct {
			OutTradeNo string `json:"out_trade_no"`
		} `json:"bill_list"`
	}
	err := c.api.Call(ctx, "externalpay payment orders",
		"/cgi-bin/externalpay/get_payment_info", map[string]string{"payment_id": paymentID}, &out)
	if err != nil {
		return nil, err
	}
	numbers := make([]string, 0, len(out.Bills))
	for _, bill := range out.Bills {
		numbers = append(numbers, bill.OutTradeNo)
	}
	return numbers, nil
}

// FundFlow is one fund flow entry.
type FundFlow struct {
	Timestamp         int64  `json:"timestamp"`
	RequestNo         string `json:"request_no"`
	TransactionType   int    `json:"transaction_type,omitempty"`
	FundFlowType      int    `json:"fund_flow_type,omitempty"`
	TransactionAmount int64  `json:"transaction_amount,omitempty"`
	AccountBalance    int64  `json:"account_balance,omitempty"`
	OutTradeNo        string `json:"out_trade_no,omitempty"`
	MchID             string `json:"mch_id,omitempty"`
}

// FundFlowPage is one page of fund flows.
type FundFlowPage struct {
	NextCursor string     `json:"next_cursor"`
	Flows      []FundFlow `json:"fund_flow_list"`
}

// FundFlows returns one page of merchant fund flows.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98100
func (c *ExternalPay) FundFlows(ctx context.Context, beginTime, endTime int64, mchID, cursor string, limit int) (FundFlowPage, error) {
	if err := check(c); err != nil {
		return FundFlowPage{}, err
	}
	if beginTime <= 0 || endTime <= beginTime {
		return FundFlowPage{}, errors.New("wecom: fund flows: a valid time window is required")
	}
	request := struct {
		BeginTime int64  `json:"begin_time"`
		EndTime   int64  `json:"end_time"`
		MchID     string `json:"mch_id,omitempty"`
		Cursor    string `json:"cursor,omitempty"`
		Limit     int    `json:"limit,omitempty"`
	}{beginTime, endTime, mchID, cursor, limit}
	var out FundFlowPage
	err := c.api.Call(ctx, "externalpay fund flows", "/cgi-bin/externalpay/get_fund_flow", request, &out)
	return out, err
}

// Applyment is one merchant application. The body follows the official
// applyment schema and is passed through by the caller.
type Applyment struct {
	OutRequestNo     string `json:"out_request_no"`
	OrganizationType int    `json:"organization_type,omitempty"`
	Raw              []byte `json:"-"`
}

// SubmitApplyment creates a merchant application for external payment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98973
func (c *ExternalPay) SubmitApplyment(ctx context.Context, payload []byte) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if len(payload) == 0 || !json.Valid(payload) {
		return "", errors.New("wecom: submit applyment: payload must be valid JSON")
	}
	var out struct {
		OutRequestNo string `json:"out_request_no"`
	}
	err := c.api.Call(ctx, "externalpay submit applyment", "/cgi-bin/miniapppay/apply_mch",
		json.RawMessage(payload), &out)
	return out.OutRequestNo, err
}

// ApplymentStatus is the review state of one application.
type ApplymentStatus struct {
	ApplymentState     string        `json:"applyment_state"`
	ApplymentStateDesc string        `json:"applyment_state_desc"`
	SignState          string        `json:"sign_state,omitempty"`
	SignURL            string        `json:"sign_url,omitempty"`
	SubMchID           string        `json:"sub_mchid,omitempty"`
	AuditDetail        []AuditDetail `json:"audit_detail,omitempty"`
}

// AuditDetail is one rejection detail.
type AuditDetail struct {
	ParamName    string `json:"param_name"`
	RejectReason string `json:"reject_reason"`
}

// ApplymentStatus reads the review state of one application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98974
func (c *ExternalPay) ApplymentStatus(ctx context.Context, outRequestNo string) (ApplymentStatus, error) {
	if err := check(c); err != nil {
		return ApplymentStatus{}, err
	}
	outRequestNo = strings.TrimSpace(outRequestNo)
	if outRequestNo == "" {
		return ApplymentStatus{}, errors.New("wecom: applyment status: out_request_no is required")
	}
	var out struct {
		Status ApplymentStatus `json:"status"`
	}
	err := c.api.Call(ctx, "externalpay applyment status", "/cgi-bin/miniapppay/get_applyment_status",
		map[string]string{"out_request_no": outRequestNo}, &out)
	return out.Status, err
}

// UploadImage uploads one image for a merchant application and returns its media
// id. The id expires after 30 days.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98972
func (c *ExternalPay) UploadImage(ctx context.Context, filename string, data []byte) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(filename) == "" || len(data) == 0 {
		return "", errors.New("wecom: upload image: filename and data are required")
	}
	var out struct {
		MediaID string `json:"open_wx_pay_media_id"`
	}
	err := c.api.Upload(ctx, "externalpay upload image", "/cgi-bin/miniapppay/upload_image",
		nil, "media", filename, data, &out)
	return out.MediaID, err
}

func check(c *ExternalPay) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: externalpay client is unavailable")
	}
	return nil
}
