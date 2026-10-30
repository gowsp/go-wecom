// Package payment implements the WeChat Pay merchant endpoints used by WeCom
// enterprise payments (/mmpaymkttransfers): red packets and payments to members.
// These endpoints use XML, merchant credentials and mutual TLS instead of the
// corp access token, so the package carries its own transport.
package payment

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// DefaultBaseURL is the WeChat Pay merchant endpoint root.
const DefaultBaseURL = "https://api.mch.weixin.qq.com"

// Option configures a payment client.
type Option func(*Client)

// WithBaseURL overrides the merchant endpoint root, mainly for tests.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.httpClient = client }
}

// WithCertificate configures mutual TLS with a merchant certificate.
func WithCertificate(certPEM, keyPEM []byte) Option {
	return func(c *Client) {
		c.certPEM = certPEM
		c.keyPEM = keyPEM
	}
}

// Client calls the merchant payment endpoints.
type Client struct {
	mchID      string
	appID      string
	secret     string
	apiKey     string
	baseURL    string
	httpClient *http.Client
	certPEM    []byte
	keyPEM     []byte

	// mtlsOnce builds the mutual TLS client lazily so the certificate is parsed
	// once per client instead of once per request.
	mtlsOnce   sync.Once
	mtlsClient *http.Client
	mtlsErr    error
}

// New creates a payment client. secret is the self-built application secret used
// for the WeCom signature, apiKey is the merchant API key used for sign.
func New(mchID, appID, secret, apiKey string, options ...Option) *Client {
	c := &Client{
		mchID: mchID, appID: appID, secret: secret, apiKey: apiKey,
		baseURL: DefaultBaseURL, httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, option := range options {
		option(c)
	}
	return c
}

// Sign computes the WeChat Pay MD5 signature: sorted non-empty parameters joined
// as k=v&..., followed by &key=APIKEY, hashed with MD5 and upper-cased.
func Sign(params map[string]string, apiKey string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if value == "" || key == "sign" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	for _, key := range keys {
		builder.WriteString(key)
		builder.WriteByte('=')
		builder.WriteString(params[key])
		builder.WriteByte('&')
	}
	builder.WriteString("key=")
	builder.WriteString(apiKey)
	sum := md5.Sum([]byte(builder.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// WorkWXSign computes the WeCom signature over the fixed field set of one
// interface, followed by the application secret, hashed with MD5 and upper-cased.
// redPacket uses the red packet fields, payment the payment fields.
func WorkWXSign(params map[string]string, secret string, redPacket bool) string {
	fields := []string{"amount", "appid", "desc", "mch_id", "nonce_str", "openid", "partner_trade_no", "ww_msg_type"}
	if redPacket {
		fields = []string{"act_name", "mch_billno", "mch_id", "nonce_str", "re_openid", "total_amount", "wxappid"}
	}
	var builder strings.Builder
	for _, field := range fields {
		value := params[field]
		if value == "" {
			continue
		}
		builder.WriteString(field)
		builder.WriteByte('=')
		builder.WriteString(value)
		builder.WriteByte('&')
	}
	builder.WriteString("secret=")
	builder.WriteString(secret)
	sum := md5.Sum([]byte(builder.String()))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

// Nonce returns a 32 character random string suitable for nonce_str.
func Nonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// RedPacketRequest sends one enterprise red packet to a member.
type RedPacketRequest struct {
	MchBillno           string `xml:"mch_billno"`
	ReOpenID            string `xml:"re_openid"`
	TotalAmount         int64  `xml:"total_amount"`
	Wishing             string `xml:"wishing"`
	ActName             string `xml:"act_name"`
	Remark              string `xml:"remark"`
	SenderName          string `xml:"sender_name,omitempty"`
	SenderHeaderMediaID string `xml:"sender_header_media_id,omitempty"`
	SceneID             string `xml:"scene_id,omitempty"`
	AgentID             string `xml:"agentid,omitempty"`
	WXAppID             string `xml:"wxappid,omitempty"`
}

// RedPacketResult is the outcome of one red packet.
type RedPacketResult struct {
	ReturnCode  string `xml:"return_code"`
	ReturnMsg   string `xml:"return_msg"`
	ResultCode  string `xml:"result_code"`
	ErrCode     string `xml:"err_code"`
	ErrCodeDes  string `xml:"err_code_des"`
	MchBillno   string `xml:"mch_billno"`
	MchID       string `xml:"mch_id"`
	WXAppID     string `xml:"wxappid"`
	ReOpenID    string `xml:"re_openid"`
	TotalAmount int64  `xml:"total_amount"`
	SendListID  string `xml:"send_listid"`
	SendTime    string `xml:"send_time"`
	Status      string `xml:"status,omitempty"`
}

// err reports a business failure carried inside an HTTP 200 response. WeCom
// answers SUCCESS only when both return_code and result_code say so; treating any
// other value as success would make a failed payout look settled.
func (r RedPacketResult) err() error {
	return paymentErr("red packet", r.ReturnCode, r.ReturnMsg, r.ResultCode, r.ErrCode, r.ErrCodeDes)
}

func (r PaymentResult) err() error {
	return paymentErr("payment", r.ReturnCode, r.ReturnMsg, r.ResultCode, r.ErrCode, r.ErrCodeDes)
}

func paymentErr(op, returnCode, returnMsg, resultCode, errCode, errCodeDes string) error {
	if returnCode == "" && resultCode == "" {
		return errors.New("wecom: payment " + op + ": empty response")
	}
	if returnCode != "SUCCESS" {
		return fmt.Errorf("wecom: payment %s: %s (%s)", op, nonEmpty(returnMsg, "failed"), returnCode)
	}
	if resultCode != "SUCCESS" {
		return fmt.Errorf("wecom: payment %s: %s [%s] %s", op,
			nonEmpty(errCodeDes, "failed"), nonEmpty(errCode, "unknown"), errCode)
	}
	return nil
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// PaymentRequest pays one member through the enterprise payment wallet.
type PaymentRequest struct {
	PartnerTradeNo string `xml:"partner_trade_no"`
	OpenID         string `xml:"openid"`
	CheckName      string `xml:"check_name"`
	ReUserName     string `xml:"re_user_name,omitempty"`
	Amount         int64  `xml:"amount"`
	Desc           string `xml:"desc"`
	SpbillCreateIP string `xml:"spbill_create_ip"`
	DeviceInfo     string `xml:"device_info,omitempty"`
	WWMsgType      string `xml:"ww_msg_type,omitempty"`
	WorkWXSign     string `xml:"workwx_sign,omitempty"`
	AgentID        string `xml:"agentid,omitempty"`
	WXAppID        string `xml:"wxappid,omitempty"`
}

// PaymentResult is the outcome of one payment.
type PaymentResult struct {
	ReturnCode     string `xml:"return_code"`
	ReturnMsg      string `xml:"return_msg"`
	ResultCode     string `xml:"result_code"`
	ErrCode        string `xml:"err_code"`
	ErrCodeDes     string `xml:"err_code_des"`
	PartnerTradeNo string `xml:"partner_trade_no"`
	PaymentNo      string `xml:"payment_no"`
	PaymentTime    string `xml:"payment_time"`
	Status         string `xml:"status,omitempty"`
	OpenID         string `xml:"openid"`
	Amount         int64  `xml:"amount"`
	Desc           string `xml:"desc"`
	MchID          string `xml:"mch_id"`
}

// SendRedPacket sends one red packet. The merchant certificate is required.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90275
func (c *Client) SendRedPacket(ctx context.Context, request RedPacketRequest) (RedPacketResult, error) {
	if err := c.check(); err != nil {
		return RedPacketResult{}, err
	}
	if request.MchBillno == "" || request.ReOpenID == "" || request.TotalAmount <= 0 {
		return RedPacketResult{}, errors.New("wecom: payment: mch_billno, re_openid and total_amount are required")
	}
	nonce, err := Nonce()
	if err != nil {
		return RedPacketResult{}, err
	}
	params := map[string]string{
		"nonce_str": nonce, "mch_billno": request.MchBillno, "mch_id": c.mchID,
		"wxappid": c.appID, "re_openid": request.ReOpenID,
		"total_amount": intString(request.TotalAmount), "wishing": request.Wishing,
		"act_name": request.ActName, "remark": request.Remark,
		"sender_name": request.SenderName, "sender_header_media_id": request.SenderHeaderMediaID,
		"scene_id": request.SceneID, "agentid": request.AgentID,
	}
	params["workwx_sign"] = WorkWXSign(params, c.secret, true)
	params["sign"] = Sign(params, c.apiKey)
	body := encode(params)
	var out RedPacketResult
	if err := c.post(ctx, "/mmpaymkttransfers/sendworkwxredpack", body, &out); err != nil {
		return RedPacketResult{}, err
	}
	if err := out.err(); err != nil {
		return out, err
	}
	return out, nil
}

// QueryRedPacket queries one red packet sent in the last 30 days.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90276
func (c *Client) QueryRedPacket(ctx context.Context, mchBillno string) (RedPacketResult, error) {
	if err := c.check(); err != nil {
		return RedPacketResult{}, err
	}
	if strings.TrimSpace(mchBillno) == "" {
		return RedPacketResult{}, errors.New("wecom: payment: mch_billno is required")
	}
	nonce, err := Nonce()
	if err != nil {
		return RedPacketResult{}, err
	}
	params := map[string]string{
		"nonce_str": nonce, "mch_billno": mchBillno, "mch_id": c.mchID, "appid": c.appID,
	}
	params["sign"] = Sign(params, c.apiKey)
	var out RedPacketResult
	if err := c.post(ctx, "/mmpaymkttransfers/queryworkwxredpack", encode(params), &out); err != nil {
		return RedPacketResult{}, err
	}
	if err := out.err(); err != nil {
		return out, err
	}
	return out, nil
}

// Pay pays one member through the enterprise payment wallet.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90278
func (c *Client) Pay(ctx context.Context, request PaymentRequest) (PaymentResult, error) {
	if err := c.check(); err != nil {
		return PaymentResult{}, err
	}
	if request.PartnerTradeNo == "" || request.OpenID == "" || request.Amount <= 0 {
		return PaymentResult{}, errors.New("wecom: payment: partner_trade_no, openid and amount are required")
	}
	nonce, err := Nonce()
	if err != nil {
		return PaymentResult{}, err
	}
	params := map[string]string{
		"nonce_str": nonce, "mch_id": c.mchID, "appid": c.appID,
		"partner_trade_no": request.PartnerTradeNo, "openid": request.OpenID,
		"check_name": request.CheckName, "re_user_name": request.ReUserName,
		"amount": intString(request.Amount), "desc": request.Desc,
		"spbill_create_ip": request.SpbillCreateIP, "device_info": request.DeviceInfo,
		"ww_msg_type": request.WWMsgType, "agentid": request.AgentID,
	}
	params["workwx_sign"] = WorkWXSign(params, c.secret, false)
	params["sign"] = Sign(params, c.apiKey)
	var out PaymentResult
	if err := c.post(ctx, "/mmpaymkttransfers/promotion/paywwsptrans2pocket", encode(params), &out); err != nil {
		return PaymentResult{}, err
	}
	if err := out.err(); err != nil {
		return out, err
	}
	return out, nil
}

// QueryPayment queries one payment made in the last 30 days.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90279
func (c *Client) QueryPayment(ctx context.Context, partnerTradeNo string) (PaymentResult, error) {
	if err := c.check(); err != nil {
		return PaymentResult{}, err
	}
	if strings.TrimSpace(partnerTradeNo) == "" {
		return PaymentResult{}, errors.New("wecom: payment: partner_trade_no is required")
	}
	nonce, err := Nonce()
	if err != nil {
		return PaymentResult{}, err
	}
	params := map[string]string{
		"nonce_str": nonce, "mch_id": c.mchID, "appid": c.appID,
		"partner_trade_no": partnerTradeNo,
	}
	params["sign"] = Sign(params, c.apiKey)
	var out PaymentResult
	if err := c.post(ctx, "/mmpaymkttransfers/promotion/querywwsptrans2pocket", encode(params), &out); err != nil {
		return PaymentResult{}, err
	}
	if err := out.err(); err != nil {
		return out, err
	}
	return out, nil
}

func (c *Client) check() error {
	if c == nil || c.mchID == "" || c.appID == "" {
		return errors.New("wecom: payment: merchant id and app id are required")
	}
	if c.secret == "" {
		return errors.New("wecom: payment: application secret is required for workwx_sign")
	}
	if c.apiKey == "" {
		return errors.New("wecom: payment: merchant api key is required for sign")
	}
	return nil
}

func (c *Client) post(ctx context.Context, path string, body []byte, out any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "text/xml; charset=utf-8")
	client, err := c.transportClient()
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("wecom: payment: unexpected status " + response.Status)
	}
	if err := xml.Unmarshal(data, out); err != nil {
		return errors.New("wecom: payment: decode response: " + err.Error())
	}
	return nil
}

// transportClient returns the HTTP client for one request. Without a merchant
// certificate it is the caller's client unchanged. With one it is a shallow copy
// whose transport carries the certificate on top of the caller's TLS settings, so
// a custom CA, ServerName or verification callback keeps working.
func (c *Client) transportClient() (*http.Client, error) {
	if len(c.certPEM) == 0 {
		return c.httpClient, nil
	}
	c.mtlsOnce.Do(func() {
		cert, err := tls.X509KeyPair(c.certPEM, c.keyPEM)
		if err != nil {
			c.mtlsErr = errors.New("wecom: payment: invalid certificate: " + err.Error())
			return
		}
		transport, err := cloneTransport(c.httpClient.Transport)
		if err != nil {
			c.mtlsErr = err
			return
		}
		config := transport.TLSClientConfig
		if config == nil {
			config = &tls.Config{}
		} else {
			config = config.Clone()
		}
		// The merchant endpoints require TLS 1.2 or later. Raise the floor of an
		// inherited configuration instead of replacing it.
		if config.MinVersion < tls.VersionTLS12 {
			config.MinVersion = tls.VersionTLS12
		}
		config.Certificates = append(append([]tls.Certificate(nil), config.Certificates...), cert)
		transport.TLSClientConfig = config

		cloned := *c.httpClient
		cloned.Transport = transport
		c.mtlsClient = &cloned
	})
	return c.mtlsClient, c.mtlsErr
}

// cloneTransport copies the caller's transport settings so a proxy or custom TLS
// configuration survives when a certificate is added. A custom round tripper
// cannot accept a client certificate, and silently swapping it for a default
// transport would drop the caller's proxy and TLS setup, so that case fails.
func cloneTransport(base http.RoundTripper) (*http.Transport, error) {
	switch transport := base.(type) {
	case nil:
		fallback, ok := http.DefaultTransport.(*http.Transport)
		if !ok {
			return nil, errors.New("wecom: payment: default transport is not an HTTP transport")
		}
		return fallback.Clone(), nil
	case *http.Transport:
		return transport.Clone(), nil
	default:
		return nil, errors.New("wecom: payment: cannot attach a merchant certificate to a custom round tripper")
	}
}

// encode renders parameters as one flat XML document.
func encode(params map[string]string) []byte {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if value == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var builder strings.Builder
	builder.WriteString("<xml>")
	for _, key := range keys {
		builder.WriteString("<")
		builder.WriteString(key)
		builder.WriteString("><![CDATA[")
		builder.WriteString(escapeCDATA(params[key]))
		builder.WriteString("]]></")
		builder.WriteString(key)
		builder.WriteString(">")
	}
	builder.WriteString("</xml>")
	return []byte(builder.String())
}

// escapeCDATA splits the only sequence that can terminate a CDATA section.
func escapeCDATA(value string) string {
	return strings.ReplaceAll(value, "]]>", "]]]]><![CDATA[>")
}

func intString(v int64) string {
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
