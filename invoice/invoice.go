// Package invoice implements the WeCom electronic invoice API
// (/cgi-bin/card/invoice/reimburse): reading invoice details and updating
// reimbursement state.
package invoice

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Invoice reads and updates electronic invoices with one shared transport.
type Invoice struct{ api *client.Client }

// New creates an invoice client.
func New(api *client.Client) *Invoice { return &Invoice{api: api} }

// Code identifies one invoice.
type Code struct {
	CardID      string `json:"card_id"`
	EncryptCode string `json:"encrypt_code"`
}

// Reimbursement states.
const (
	StatusInit     = "INVOICE_REIMBURSE_INIT"
	StatusLocked   = "INVOICE_REIMBURSE_LOCK"
	StatusFinished = "INVOICE_REIMBURSE_CLOSURE"
)

// UserInfo is the invoice header entered by the user.
type UserInfo struct {
	Fee         int64  `json:"fee,omitempty"`
	Title       string `json:"title,omitempty"`
	BillingTime int64  `json:"billing_time,omitempty"`
	BillingNo   string `json:"billing_no,omitempty"`
	BillingCode string `json:"billing_code,omitempty"`
	Info        []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"info,omitempty"`
	FeeWithoutTax   int64  `json:"fee_without_tax,omitempty"`
	Tax             int64  `json:"tax,omitempty"`
	Detail          string `json:"detail,omitempty"`
	PDFURL          string `json:"pdf_url,omitempty"`
	ReimburseStatus string `json:"reimburse_status,omitempty"`
	OrderID         string `json:"order_id,omitempty"`
	CheckCode       string `json:"check_code,omitempty"`
	BuyerNumber     string `json:"buyer_number,omitempty"`
}

// Detail is one electronic invoice.
type Detail struct {
	CardID    string          `json:"card_id"`
	BeginTime int64           `json:"begin_time,omitempty"`
	EndTime   int64           `json:"end_time,omitempty"`
	OpenID    string          `json:"openid,omitempty"`
	Type      string          `json:"type,omitempty"`
	Payee     string          `json:"payee,omitempty"`
	Detail    string          `json:"detail,omitempty"`
	UserInfo  UserInfo        `json:"user_info"`
	Raw       json.RawMessage `json:"-"`
}

// Get reads one invoice by card id and encrypted code.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90284
func (c *Invoice) Get(ctx context.Context, code Code) (Detail, error) {
	if err := check(c); err != nil {
		return Detail{}, err
	}
	if strings.TrimSpace(code.CardID) == "" || strings.TrimSpace(code.EncryptCode) == "" {
		return Detail{}, errors.New("wecom: invoice: card_id and encrypt_code are required")
	}
	var out Detail
	err := c.api.Call(ctx, "invoice get",
		"/cgi-bin/card/invoice/reimburse/getinvoiceinfo", code, &out)
	return out, err
}

// GetBatch reads several invoices at once.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90287
func (c *Invoice) GetBatch(ctx context.Context, codes []Code) ([]Detail, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(codes) == 0 {
		return nil, errors.New("wecom: invoice batch: item_list is empty")
	}
	type itemList struct {
		Items []Code `json:"item_list"`
	}
	var out struct {
		List []Detail `json:"item_list"`
	}
	err := c.api.Call(ctx, "invoice batch", "/cgi-bin/card/invoice/reimburse/getinvoiceinfobatch",
		itemList{Items: codes}, &out)
	return out.List, err
}

// SetStatus updates the reimbursement state of one invoice.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90285
func (c *Invoice) SetStatus(ctx context.Context, code Code, status string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(code.CardID) == "" || strings.TrimSpace(code.EncryptCode) == "" {
		return errors.New("wecom: invoice status: card_id and encrypt_code are required")
	}
	request := struct {
		Code        Code   `json:"-"`
		CardID      string `json:"card_id"`
		EncryptCode string `json:"encrypt_code"`
		Status      string `json:"reimburse_status"`
	}{CardID: code.CardID, EncryptCode: code.EncryptCode, Status: status}
	return c.api.Call(ctx, "invoice status",
		"/cgi-bin/card/invoice/reimburse/updateinvoicestatus", request, nil)
}

// SetStatusBatch updates the reimbursement state of several invoices.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90286
func (c *Invoice) SetStatusBatch(ctx context.Context, openID, status string, codes []Code) error {
	if err := check(c); err != nil {
		return err
	}
	if len(codes) == 0 {
		return errors.New("wecom: invoice batch status: invoice_list is empty")
	}
	request := struct {
		OpenID string `json:"openid,omitempty"`
		Status string `json:"reimburse_status"`
		List   []Code `json:"invoice_list"`
	}{OpenID: openID, Status: status, List: codes}
	return c.api.Call(ctx, "invoice batch status",
		"/cgi-bin/card/invoice/reimburse/updatestatusbatch", request, nil)
}

func check(c *Invoice) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: invoice client is unavailable")
	}
	return nil
}
