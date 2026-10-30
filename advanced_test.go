package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/advanced"
)

func TestAdvancedFeatures(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"next_cursor":"c2","has_more":true,`+
		`"apply_id_list":["a1","a2"],`+
		`"approval_info":{"applicant":"u1","apply_id":"a1","approval_id":"ap1","approval_status":1}}`)
	svc := advanced.New(api)
	ctx := context.Background()

	page, err := svc.ApplyIDs(ctx, advanced.ApplyQuery{BusinessType: 1, Limit: 10})
	if err != nil || len(page.ApplyIDs) != 2 || !page.HasMore ||
		rec.path != "/cgi-bin/advanced_feature/get_apply_id_list" {
		t.Fatalf("ids: %v %+v %s", err, page, rec.path)
	}

	info, err := svc.GetApproval(ctx, "a1")
	if err != nil || info.Applicant != "u1" ||
		rec.path != "/cgi-bin/advanced_feature/get_approval_info" {
		t.Fatalf("approval: %v %+v %s", err, info, rec.path)
	}

	if err := svc.SetApprovalDetail(ctx, advanced.ApprovalDetail{
		ApprovalID: "ap1", ApplyID: "a1", ApprovalStatus: 1,
		ProcessList: advanced.ProcessList{NodeList: []advanced.Node{{Status: 1}}},
	}); err != nil || rec.path != "/cgi-bin/advanced_feature/set_approval_detail" {
		t.Fatalf("set detail: %v %s", err, rec.path)
	}

	if _, err := svc.ApplyIDs(ctx, advanced.ApplyQuery{}); err == nil {
		t.Fatal("ids without business type accepted")
	}
	if _, err := svc.GetApproval(ctx, ""); err == nil {
		t.Fatal("approval without id accepted")
	}
	if err := svc.SetApprovalDetail(ctx, advanced.ApprovalDetail{}); err == nil {
		t.Fatal("detail without ids accepted")
	}
}

func TestAdvancedEvents(t *testing.T) {
	submit := `<xml><MsgType>event</MsgType><Event><![CDATA[submit_vip_account_approval]]></Event>` +
		`<ApplyReason><![CDATA[需要空间]]></ApplyReason><BusinessType>1</BusinessType>` +
		`<ApplyId><![CDATA[a1]]></ApplyId><AgentID>1</AgentID></xml>`
	event, err := advanced.ParseSubmitEvent([]byte(submit))
	if err != nil || event.ApplyID != "a1" || event.BusinessType != 1 {
		t.Fatalf("submit: %v %+v", err, event)
	}
	if _, err := advanced.ParseSubmitEvent([]byte(
		`<xml><Event>submit_vip_account_approval</Event></xml>`)); err == nil {
		t.Fatal("submit without apply id accepted")
	}

	finish := `<xml><MsgType>event</MsgType><Event><![CDATA[finish_vip_account_approval]]></Event>` +
		`<FinishType>1</FinishType><ApplyId><![CDATA[a1]]></ApplyId></xml>`
	done, err := advanced.ParseFinishEvent([]byte(finish))
	if err != nil || done.ApplyID != "a1" || done.FinishType != "1" {
		t.Fatalf("finish: %v %+v", err, done)
	}
	if _, err := advanced.ParseFinishEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected finish event accepted")
	}
}
