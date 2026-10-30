package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/invoice"
)

func TestInvoice(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"card_id":"c1","type":"增值税电子普通发票",`+
		`"user_info":{"fee":4000,"title":"XX有限公司","reimburse_status":"INVOICE_REIMBURSE_INIT"},`+
		`"item_list":[{"card_id":"c1","user_info":{"fee":4000}}]}`)
	svc := invoice.New(api)
	ctx := context.Background()

	detail, err := svc.Get(ctx, invoice.Code{CardID: "c1", EncryptCode: "e1"})
	if err != nil || detail.UserInfo.Fee != 4000 ||
		rec.path != "/cgi-bin/card/invoice/reimburse/getinvoiceinfo" {
		t.Fatalf("get: %v %+v %s", err, detail, rec.path)
	}

	list, err := svc.GetBatch(ctx, []invoice.Code{{CardID: "c1", EncryptCode: "e1"}})
	if err != nil || len(list) != 1 ||
		rec.path != "/cgi-bin/card/invoice/reimburse/getinvoiceinfobatch" {
		t.Fatalf("batch: %v %+v %s", err, list, rec.path)
	}

	if err := svc.SetStatus(ctx, invoice.Code{CardID: "c1", EncryptCode: "e1"},
		invoice.StatusLocked); err != nil ||
		rec.path != "/cgi-bin/card/invoice/reimburse/updateinvoicestatus" {
		t.Fatalf("status: %v %s", err, rec.path)
	}
	if rec.body["reimburse_status"] != invoice.StatusLocked {
		t.Fatalf("status not sent: %v", rec.body)
	}
	if _, ok := rec.body["card_id"]; !ok {
		t.Fatalf("card_id not sent: %v", rec.body)
	}

	if err := svc.SetStatusBatch(ctx, "open1", invoice.StatusFinished,
		[]invoice.Code{{CardID: "c1", EncryptCode: "e1"}}); err != nil ||
		rec.path != "/cgi-bin/card/invoice/reimburse/updatestatusbatch" {
		t.Fatalf("batch status: %v %s", err, rec.path)
	}

	if _, err := svc.Get(ctx, invoice.Code{}); err == nil {
		t.Fatal("get without ids accepted")
	}
	if _, err := svc.GetBatch(ctx, nil); err == nil {
		t.Fatal("batch without codes accepted")
	}
	if err := svc.SetStatus(ctx, invoice.Code{}, invoice.StatusInit); err == nil {
		t.Fatal("status without ids accepted")
	}
	if err := svc.SetStatusBatch(ctx, "o", invoice.StatusInit, nil); err == nil {
		t.Fatal("batch status without codes accepted")
	}
}
