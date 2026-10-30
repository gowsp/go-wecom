// Package miniapppay implements WeCom mini program payments
// (/cgi-bin/miniapppay): orders, signing, refunds, bills and the payment and
// refund notifications.
package miniapppay

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// MiniAppPay manages mini program payments with one shared transport.
type MiniAppPay struct{ api *client.Client }

// New creates a mini program payment client.
func New(api *client.Client) *MiniAppPay { return &MiniAppPay{api: api} }

// Amount is a payment amount in cents.
type Amount struct {
	Total         int64  `json:"total"`
	Currency      string `json:"currency,omitempty"`
	PayerTotal    int64  `json:"payer_total,omitempty"`
	PayerCurrency string `json:"payer_currency,omitempty"`
	Refund        int64  `json:"refund,omitempty"`
}

// Payer is the paying customer.
type Payer struct {
	OpenID string `json:"openid"`
}

// Order creates one mini program payment order.
type Order struct {
	AppID       string          `json:"appid"`
	MchID       string          `json:"mchid"`
	OutTradeNo  string          `json:"out_trade_no"`
	Description string          `json:"description,omitempty"`
	SceneKey    string          `json:"scenekey,omitempty"`
	Amount      Amount          `json:"amount"`
	Payer       Payer           `json:"payer,omitempty"`
	TimeExpire  string          `json:"time_expire,omitempty"`
	Attach      string          `json:"attach,omitempty"`
	GoodsTag    string          `json:"goods_tag,omitempty"`
	SceneInfo   json.RawMessage `json:"scene_info,omitempty"`
}

// OrderState is the trade state of one order.
type OrderState struct {
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc,omitempty"`
	TransactionID  string `json:"transaction_id,omitempty"`
	Payer          Payer  `json:"payer"`
}

// RefundRequest asks for a refund of one order.
type RefundRequest struct {
	MchID        string `json:"mchid"`
	AppID        string `json:"appid"`
	OutTradeNo   string `json:"out_trade_no"`
	OutRefundNo  string `json:"out_refund_no"`
	Reason       string `json:"reason,omitempty"`
	FundsAccount string `json:"funds_account,omitempty"`
	Amount       Amount `json:"amount"`
}

// Refund is the state of one refund.
type Refund struct {
	OutRefundNo     string `json:"out_refund_no"`
	OutTradeNo      string `json:"out_trade_no,omitempty"`
	Channel         string `json:"channel,omitempty"`
	Status          string `json:"status"`
	ReceivedAccount string `json:"user_received_account,omitempty"`
	SuccessTime     string `json:"success_time,omitempty"`
	CreateTime      string `json:"create_time,omitempty"`
	Amount          Amount `json:"amount"`
}

// BillRequest selects one transaction bill.
type BillRequest struct {
	BillDate string `json:"bill_date"`
	MchID    string `json:"mchid"`
	BillType string `json:"bill_type,omitempty"`
	TarType  string `json:"tar_type,omitempty"`
}

// authHeader turns the descriptor's auth value into a request header. The
// official documents describe it as a whole header line, for example
// "Authorization: xxx", so sending it as a value alone would double the name.
// A value without a colon is treated as the Authorization value.
func authHeader(auth string) (http.Header, error) {
	header := http.Header{}
	auth = strings.TrimSpace(auth)
	if auth == "" {
		return header, nil
	}
	name, value, found := strings.Cut(auth, ":")
	if !found {
		header.Set("Authorization", auth)
		return header, nil
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" || value == "" {
		return nil, errors.New("wecom: download bill: auth header is malformed")
	}
	header.Set(name, value)
	return header, nil
}

// BillFile is the download descriptor of one transaction bill.
type BillFile struct {
	DownloadURL string `json:"download_url"`
	HashType    string `json:"hash_type,omitempty"`
	HashValue   string `json:"hash_value,omitempty"`
	Auth        string `json:"auth,omitempty"`
}

// CreateOrder creates one order and returns its prepay id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97322
func (c *MiniAppPay) CreateOrder(ctx context.Context, order Order) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(order.AppID) == "" || strings.TrimSpace(order.MchID) == "" ||
		strings.TrimSpace(order.OutTradeNo) == "" {
		return "", errors.New("wecom: create order: appid, mchid and out_trade_no are required")
	}
	if order.Amount.Total <= 0 {
		return "", errors.New("wecom: create order: amount.total is required")
	}
	var out struct {
		PrepayID string `json:"prepay_id"`
	}
	err := c.api.Call(ctx, "miniapppay create order",
		"/cgi-bin/miniapppay/create_order", order, &out)
	return out.PrepayID, err
}

// Order returns the current state of one order.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97323
func (c *MiniAppPay) Order(ctx context.Context, mchID, outTradeNo string) (OrderState, error) {
	if err := check(c); err != nil {
		return OrderState{}, err
	}
	if strings.TrimSpace(mchID) == "" || strings.TrimSpace(outTradeNo) == "" {
		return OrderState{}, errors.New("wecom: order: mchid and out_trade_no are required")
	}
	var out OrderState
	err := c.api.Call(ctx, "miniapppay order", "/cgi-bin/miniapppay/get_order",
		map[string]string{"mchid": mchID, "out_trade_no": outTradeNo}, &out)
	return out, err
}

// CloseOrder closes one unpaid order.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97324
func (c *MiniAppPay) CloseOrder(ctx context.Context, mchID, outTradeNo string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(mchID) == "" || strings.TrimSpace(outTradeNo) == "" {
		return errors.New("wecom: close order: mchid and out_trade_no are required")
	}
	return c.api.Call(ctx, "miniapppay close order", "/cgi-bin/miniapppay/close_order",
		map[string]string{"mchid": mchID, "out_trade_no": outTradeNo}, nil)
}

// PaySign returns the signature the mini program uses to invoke payment.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98130
func (c *MiniAppPay) PaySign(ctx context.Context, appID, prepayID, signType, nonce string, timestamp int64) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(appID) == "" || strings.TrimSpace(prepayID) == "" {
		return "", errors.New("wecom: pay sign: appid and prepay_id are required")
	}
	request := struct {
		AppID     string `json:"appid"`
		PrepayID  string `json:"prepay_id"`
		SignType  string `json:"sign_type,omitempty"`
		Nonce     string `json:"nonce,omitempty"`
		Timestamp int64  `json:"timestamp,omitempty"`
	}{AppID: appID, PrepayID: prepayID, SignType: signType, Nonce: nonce, Timestamp: timestamp}
	var out struct {
		Sign string `json:"pay_sign"`
	}
	err := c.api.Call(ctx, "miniapppay sign", "/cgi-bin/miniapppay/get_sign", request, &out)
	return out.Sign, err
}

// Refund applies for a refund and returns the refund number.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97333
func (c *MiniAppPay) Refund(ctx context.Context, request RefundRequest) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(request.MchID) == "" || strings.TrimSpace(request.OutTradeNo) == "" ||
		strings.TrimSpace(request.OutRefundNo) == "" {
		return "", errors.New("wecom: refund: mchid, out_trade_no and out_refund_no are required")
	}
	var out struct {
		OutRefundNo string `json:"out_refund_no"`
	}
	err := c.api.Call(ctx, "miniapppay refund", "/cgi-bin/miniapppay/refund", request, &out)
	return out.OutRefundNo, err
}

// RefundDetail returns the state of one refund.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97352
func (c *MiniAppPay) RefundDetail(ctx context.Context, mchID, outRefundNo string) (Refund, error) {
	if err := check(c); err != nil {
		return Refund{}, err
	}
	if strings.TrimSpace(mchID) == "" || strings.TrimSpace(outRefundNo) == "" {
		return Refund{}, errors.New("wecom: refund detail: mchid and out_refund_no are required")
	}
	var out Refund
	err := c.api.Call(ctx, "miniapppay refund detail",
		"/cgi-bin/miniapppay/get_refund_detail",
		map[string]string{"mchid": mchID, "out_refund_no": outRefundNo}, &out)
	return out, err
}

// Bill requests one transaction bill and returns its download descriptor.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98115
func (c *MiniAppPay) Bill(ctx context.Context, request BillRequest) (BillFile, error) {
	if err := check(c); err != nil {
		return BillFile{}, err
	}
	if strings.TrimSpace(request.BillDate) == "" || strings.TrimSpace(request.MchID) == "" {
		return BillFile{}, errors.New("wecom: bill: bill_date and mchid are required")
	}
	var out BillFile
	err := c.api.Call(ctx, "miniapppay bill", "/cgi-bin/miniapppay/get_bill", request, &out)
	return out, err
}

// DownloadBill fetches the bill file using the descriptor returned by Bill. The
// auth value is sent as the Authorization header. The shared client performs the
// request, so a proxy or custom TLS setup is honoured, and a file larger than the
// download limit fails instead of being truncated.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98115
func (c *MiniAppPay) DownloadBill(ctx context.Context, file BillFile) ([]byte, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(file.DownloadURL) == "" {
		return nil, errors.New("wecom: download bill: download_url is required")
	}
	if !strings.HasPrefix(file.DownloadURL, "https://") && !strings.HasPrefix(file.DownloadURL, "http://") {
		return nil, errors.New("wecom: download bill: download_url must be an absolute URL")
	}
	header, err := authHeader(file.Auth)
	if err != nil {
		return nil, err
	}
	return c.api.FetchURL(ctx, file.DownloadURL, header)
}

// PaymentResult is the body of one payment notification. WeCom delivers it as
// plain JSON, not as an encrypted WeChat Pay v3 envelope.
type PaymentResult struct {
	AppID          string `json:"appid"`
	MchID          string `json:"mchid"`
	OutTradeNo     string `json:"out_trade_no"`
	TransactionID  string `json:"transaction_id,omitempty"`
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc,omitempty"`
	TradeType      string `json:"trade_type,omitempty"`
	Attach         string `json:"attach,omitempty"`
	SuccessTime    string `json:"success_time,omitempty"`
	Amount         Amount `json:"amount"`
	Payer          Payer  `json:"payer"`
}

// RefundResult is the body of one refund notification.
type RefundResult struct {
	MchID           string `json:"mchid"`
	OutTradeNo      string `json:"out_trade_no"`
	RefundID        string `json:"refund_id"`
	OutRefundNo     string `json:"out_refund_no"`
	RefundStatus    string `json:"refund_status"`
	SuccessTime     string `json:"success_time,omitempty"`
	ReceivedAccount string `json:"user_received_account,omitempty"`
	Amount          Amount `json:"amount"`
}

// ParsePaymentNotification decodes one payment notification body and checks that
// the fields needed to settle an order are present.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97335
func ParsePaymentNotification(data []byte) (PaymentResult, error) {
	var result PaymentResult
	if err := json.Unmarshal(data, &result); err != nil {
		return PaymentResult{}, errors.New("wecom: payment notification: " + err.Error())
	}
	if strings.TrimSpace(result.MchID) == "" || strings.TrimSpace(result.OutTradeNo) == "" ||
		strings.TrimSpace(result.TradeState) == "" {
		return PaymentResult{}, errors.New("wecom: payment notification: mchid, out_trade_no and trade_state are required")
	}
	return result, nil
}

// ParseRefundNotification decodes one refund notification body.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97338
func ParseRefundNotification(data []byte) (RefundResult, error) {
	var result RefundResult
	if err := json.Unmarshal(data, &result); err != nil {
		return RefundResult{}, errors.New("wecom: refund notification: " + err.Error())
	}
	if strings.TrimSpace(result.MchID) == "" || strings.TrimSpace(result.OutRefundNo) == "" ||
		strings.TrimSpace(result.RefundStatus) == "" {
		return RefundResult{}, errors.New("wecom: refund notification: mchid, out_refund_no and refund_status are required")
	}
	return result, nil
}

func check(c *MiniAppPay) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: miniapppay client is unavailable")
	}
	return nil
}
