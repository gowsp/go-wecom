package wecom_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gowsp/go-wecom/externalpay"
)

func TestExternalPayMerchant(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"mch_id":"m1","bind_status":1,"merchant_name":"商户",`+
		`"allow_use_scope":{"user":["u1"],"partyid":[1],"tagid":[2]}}`)
	svc := externalpay.New(api)
	ctx := context.Background()

	merchant, err := svc.Merchant(ctx, "m1")
	if err != nil || merchant.MerchantName != "商户" || merchant.Scope == nil ||
		rec.path != "/cgi-bin/externalpay/getmerchant" {
		t.Fatalf("merchant: %v %+v %s", err, merchant, rec.path)
	}
	if err := svc.SetMerchantScope(ctx, "m1", externalpay.Scope{Users: []string{"u1"}}); err != nil ||
		rec.path != "/cgi-bin/externalpay/set_mch_use_scope" {
		t.Fatalf("scope: %v %s", err, rec.path)
	}
	if _, err := svc.Merchant(ctx, ""); err == nil {
		t.Fatal("merchant without id accepted")
	}
	if err := svc.SetMerchantScope(ctx, "", externalpay.Scope{}); err == nil {
		t.Fatal("scope without id accepted")
	}
}

func TestExternalPayRecords(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"next_cursor":"c2",`+
		`"bill_list":[{"transaction_id":"tx1","total_fee":100,"out_trade_no":"o1"}],`+
		`"fund_flow_list":[{"request_no":"r1","transaction_amount":20}]}`)
	svc := externalpay.New(api)
	ctx := context.Background()

	bills, err := svc.Bills(ctx, 1, 2, "u1", "", 10)
	if err != nil || len(bills.Bills) != 1 || bills.Bills[0].TotalFee != 100 ||
		rec.path != "/cgi-bin/externalpay/get_bill_list" {
		t.Fatalf("bills: %v %+v %s", err, bills, rec.path)
	}

	orders, err := svc.PaymentOrders(ctx, "pay1")
	if err != nil || len(orders) != 1 || orders[0] != "o1" ||
		rec.path != "/cgi-bin/externalpay/get_payment_info" {
		t.Fatalf("orders: %v %v %s", err, orders, rec.path)
	}

	flows, err := svc.FundFlows(ctx, 1, 2, "m1", "", 10)
	if err != nil || len(flows.Flows) != 1 || rec.path != "/cgi-bin/externalpay/get_fund_flow" {
		t.Fatalf("flows: %v %+v %s", err, flows, rec.path)
	}

	if _, err := svc.Bills(ctx, 0, 0, "", "", 0); err == nil {
		t.Fatal("invalid bill window accepted")
	}
	if _, err := svc.PaymentOrders(ctx, ""); err == nil {
		t.Fatal("orders without id accepted")
	}
	if _, err := svc.FundFlows(ctx, 2, 1, "", "", 0); err == nil {
		t.Fatal("invalid flow window accepted")
	}
}

func TestExternalPayApplyment(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"out_request_no":"req1",`+
		`"status":{"applyment_state":"AUDITING","sign_url":"https://sign",`+
		`"audit_detail":[{"param_name":"merchant_name","reject_reason":"bad"}]},`+
		`"open_wx_pay_media_id":"media1"}`)
	svc := externalpay.New(api)
	ctx := context.Background()

	payload := json.RawMessage(`{"out_request_no":"req1","organization_type":0}`)
	requestNo, err := svc.SubmitApplyment(ctx, payload)
	if err != nil || requestNo != "req1" || rec.path != "/cgi-bin/miniapppay/apply_mch" {
		t.Fatalf("submit: %q %v %s", requestNo, err, rec.path)
	}

	status, err := svc.ApplymentStatus(ctx, "req1")
	if err != nil || status.ApplymentState != "AUDITING" || len(status.AuditDetail) != 1 ||
		rec.path != "/cgi-bin/miniapppay/get_applyment_status" {
		t.Fatalf("status: %v %+v %s", err, status, rec.path)
	}

	mediaID, err := svc.UploadImage(ctx, "a.png", []byte("png"))
	if err != nil || mediaID != "media1" || rec.path != "/cgi-bin/miniapppay/upload_image" {
		t.Fatalf("upload: %q %v %s", mediaID, err, rec.path)
	}

	if _, err := svc.SubmitApplyment(ctx, json.RawMessage(`{`)); err == nil {
		t.Fatal("invalid payload accepted")
	}
	if _, err := svc.ApplymentStatus(ctx, ""); err == nil {
		t.Fatal("status without id accepted")
	}
	if _, err := svc.UploadImage(ctx, "", nil); err == nil {
		t.Fatal("upload without data accepted")
	}
}
