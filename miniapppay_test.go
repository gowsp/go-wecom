package wecom_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowsp/go-wecom/client"

	"github.com/gowsp/go-wecom/miniapppay"
)

func TestMiniAppPayOrders(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"prepay_id":"prepay1","pay_sign":"sig",`+
		`"mchid":"m1","out_trade_no":"o1","trade_state":"SUCCESS",`+
		`"out_refund_no":"r1","status":"SUCCESS",`+
		`"download_url":"https://bill","hash_type":"SHA1","auth":"auth-token"}`)
	svc := miniapppay.New(api)
	ctx := context.Background()

	prepayID, err := svc.CreateOrder(ctx, miniapppay.Order{
		AppID: "wx1", MchID: "m1", OutTradeNo: "o1",
		Amount: miniapppay.Amount{Total: 100, Currency: "CNY"},
	})
	if err != nil || prepayID != "prepay1" || rec.path != "/cgi-bin/miniapppay/create_order" {
		t.Fatalf("create: %q %v %s", prepayID, err, rec.path)
	}

	state, err := svc.Order(ctx, "m1", "o1")
	if err != nil || state.TradeState != "SUCCESS" || rec.path != "/cgi-bin/miniapppay/get_order" {
		t.Fatalf("order: %v %+v %s", err, state, rec.path)
	}

	if err := svc.CloseOrder(ctx, "m1", "o1"); err != nil || rec.path != "/cgi-bin/miniapppay/close_order" {
		t.Fatalf("close: %v %s", err, rec.path)
	}

	sign, err := svc.PaySign(ctx, "wx1", "prepay1", "RSA", "nonce", 1)
	if err != nil || sign != "sig" || rec.path != "/cgi-bin/miniapppay/get_sign" {
		t.Fatalf("sign: %q %v %s", sign, err, rec.path)
	}

	refundNo, err := svc.Refund(ctx, miniapppay.RefundRequest{
		MchID: "m1", AppID: "wx1", OutTradeNo: "o1", OutRefundNo: "r1",
		Amount: miniapppay.Amount{Refund: 50, Total: 100},
	})
	if err != nil || refundNo != "r1" || rec.path != "/cgi-bin/miniapppay/refund" {
		t.Fatalf("refund: %q %v %s", refundNo, err, rec.path)
	}

	refund, err := svc.RefundDetail(ctx, "m1", "r1")
	if err != nil || refund.Status != "SUCCESS" || rec.path != "/cgi-bin/miniapppay/get_refund_detail" {
		t.Fatalf("refund detail: %v %+v %s", err, refund, rec.path)
	}

	file, err := svc.Bill(ctx, miniapppay.BillRequest{BillDate: "2024-01-01", MchID: "m1"})
	if err != nil || file.DownloadURL != "https://bill" || rec.path != "/cgi-bin/miniapppay/get_bill" {
		t.Fatalf("bill: %v %+v %s", err, file, rec.path)
	}

	if _, err := svc.CreateOrder(ctx, miniapppay.Order{AppID: "wx1"}); err == nil {
		t.Fatal("order without ids accepted")
	}
	if _, err := svc.CreateOrder(ctx, miniapppay.Order{AppID: "wx1", MchID: "m1", OutTradeNo: "o1"}); err == nil {
		t.Fatal("order without amount accepted")
	}
	if _, err := svc.Order(ctx, "", ""); err == nil {
		t.Fatal("order without ids accepted")
	}
	if err := svc.CloseOrder(ctx, "m1", ""); err == nil {
		t.Fatal("close without order accepted")
	}
	if _, err := svc.PaySign(ctx, "", "", "", "", 0); err == nil {
		t.Fatal("sign without ids accepted")
	}
	if _, err := svc.Refund(ctx, miniapppay.RefundRequest{}); err == nil {
		t.Fatal("refund without ids accepted")
	}
	if _, err := svc.RefundDetail(ctx, "m1", ""); err == nil {
		t.Fatal("refund detail without id accepted")
	}
	if _, err := svc.Bill(ctx, miniapppay.BillRequest{}); err == nil {
		t.Fatal("bill without date accepted")
	}
	if _, err := svc.DownloadBill(ctx, miniapppay.BillFile{}); err == nil {
		t.Fatal("download without url accepted")
	}
}

func TestMiniAppPayNotification(t *testing.T) {
	paymentBody := `{"appid":"wx1","mchid":"m1","out_trade_no":"o1",` +
		`"transaction_id":"t1","trade_state":"SUCCESS","trade_type":"MICROPAY",` +
		`"success_time":"2018-06-08T10:34:56+08:00",` +
		`"amount":{"payer_total":100,"total":100,"currency":"CNY"},` +
		`"payer":{"openid":"open1"}}`
	result, err := miniapppay.ParsePaymentNotification([]byte(paymentBody))
	if err != nil {
		t.Fatal(err)
	}
	if result.OutTradeNo != "o1" || result.TradeState != "SUCCESS" ||
		result.Amount.Total != 100 || result.Payer.OpenID != "open1" {
		t.Fatalf("payment notification: %+v", result)
	}

	refundBody := `{"mchid":"m1","out_trade_no":"o1","refund_id":"r1",` +
		`"out_refund_no":"orn1","refund_status":"SUCCESS",` +
		`"amount":{"total":999,"refund":999}}`
	refund, err := miniapppay.ParseRefundNotification([]byte(refundBody))
	if err != nil {
		t.Fatal(err)
	}
	if refund.OutRefundNo != "orn1" || refund.RefundStatus != "SUCCESS" || refund.Amount.Refund != 999 {
		t.Fatalf("refund notification: %+v", refund)
	}

	if _, err := miniapppay.ParsePaymentNotification([]byte(`{"appid":"wx1"}`)); err == nil {
		t.Fatal("payment notification without identifiers accepted")
	}
	if _, err := miniapppay.ParseRefundNotification([]byte(`{"mchid":"m1"}`)); err == nil {
		t.Fatal("refund notification without identifiers accepted")
	}
	if _, err := miniapppay.ParsePaymentNotification([]byte(`{`)); err == nil {
		t.Fatal("malformed payment notification accepted")
	}
}

// TestMiniAppPayDownloadBill covers the descriptor download: the auth header is
// forwarded, the caller's transport is used, and an oversized file fails instead
// of being silently truncated.
func TestMiniAppPayDownloadBill(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.CopyN(w, io.LimitReader(zeroReader{}, 16), 16)
	}))
	defer server.Close()

	api := client.New("corp", "secret", client.WithBaseURL(server.URL))
	svc := miniapppay.New(api)
	data, err := svc.DownloadBill(context.Background(),
		miniapppay.BillFile{DownloadURL: server.URL + "/bill", Auth: "Authorization: auth-token"})
	if err != nil || len(data) != 16 {
		t.Fatalf("download: %d %v", len(data), err)
	}
	if gotAuth != "auth-token" {
		t.Fatalf("authorization not forwarded: %q", gotAuth)
	}

	// One byte over the configured limit must not look like a complete file.
	limited := client.New("corp", "secret",
		client.WithBaseURL(server.URL), client.WithDownloadLimit(8))
	if _, err := miniapppay.New(limited).DownloadBill(context.Background(),
		miniapppay.BillFile{DownloadURL: server.URL + "/bill"}); !errors.Is(err, client.ErrResponseTooLarge) {
		t.Fatalf("oversized download not rejected: %v", err)
	}

	// A bare value still goes out as the Authorization header.
	if _, err := svc.DownloadBill(context.Background(),
		miniapppay.BillFile{DownloadURL: server.URL + "/bill", Auth: "auth-token"}); err != nil {
		t.Fatalf("bare auth value: %v", err)
	}
	if gotAuth != "auth-token" {
		t.Fatalf("bare authorization not forwarded: %q", gotAuth)
	}

	if _, err := svc.DownloadBill(context.Background(),
		miniapppay.BillFile{DownloadURL: "not-a-url"}); err == nil {
		t.Fatal("relative download url accepted")
	}
	if _, err := svc.DownloadBill(context.Background(),
		miniapppay.BillFile{DownloadURL: server.URL, Auth: " : "}); err == nil {
		t.Fatal("malformed auth header accepted")
	}
}

// zeroReader yields an endless run of zero bytes.
type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
