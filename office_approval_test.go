package wecom_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gowsp/go-wecom/approval"
)

func TestApprovalTemplates(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"template_id":"tpl1",`+
		`"template_names":[{"text":"报销","lang":"zh_CN"}],`+
		`"template_content":{"controls":[{"property":{"control":"Text"}}]}}`)
	svc := approval.New(api)
	ctx := context.Background()

	tpl, err := svc.GetTemplate(ctx, "tpl1")
	if err != nil || rec.path != "/cgi-bin/oa/gettemplatedetail" || len(tpl.Names) != 1 {
		t.Fatalf("get template: %v %+v %s", err, tpl, rec.path)
	}
	if len(tpl.Content) == 0 {
		t.Fatal("template content not preserved")
	}

	created, err := svc.CreateTemplate(ctx, approval.Template{
		Names:   []approval.Text{{Text: "报销", Lang: "zh_CN"}},
		Content: json.RawMessage(`{"controls":[]}`),
	})
	if err != nil || created != "tpl1" || rec.path != "/cgi-bin/oa/approval/create_template" {
		t.Fatalf("create template: %q %v %s", created, err, rec.path)
	}

	if err := svc.UpdateTemplate(ctx, "tpl1", approval.Template{
		Names:   []approval.Text{{Text: "报销2", Lang: "zh_CN"}},
		Content: json.RawMessage(`{"controls":[]}`),
	}); err != nil || rec.path != "/cgi-bin/oa/approval/update_template" {
		t.Fatalf("update template: %v %s", err, rec.path)
	}

	if _, err := svc.GetTemplate(ctx, " "); err == nil {
		t.Fatal("empty template id accepted")
	}
	if _, err := svc.CreateTemplate(ctx, approval.Template{}); err == nil {
		t.Fatal("empty template accepted")
	}
	if err := svc.UpdateTemplate(ctx, "", approval.Template{}); err == nil {
		t.Fatal("update without id accepted")
	}
}

func TestApprovalSubmitAndQuery(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"sp_no":"SP1","sp_no_list":["SP1"],"new_next_cursor":"c2",`+
		`"info":{"sp_no":"SP1","sp_name":"报销","sp_status":1,"applyer":{"userid":"u1"},`+
		`"sp_record":[{"sp_status":1}]},"next_spnum":2,"data":[{"spname":"报销"}],`+
		`"lists":[{"id":1,"vacationname":"年假","leftduration":10}],"count":1}`)
	svc := approval.New(api)
	ctx := context.Background()

	spNo, err := svc.Submit(ctx, approval.ApplyRequest{
		CreatorUserID: "u1", TemplateID: "tpl1", UseTemplateApprover: 1,
		ApplyData: json.RawMessage(`{"contents":[{"control":"Text","value":{"text":"x"}}]}`),
	})
	if err != nil || spNo != "SP1" || rec.path != "/cgi-bin/oa/applyevent" {
		t.Fatalf("submit: %q %v %s", spNo, err, rec.path)
	}
	if rec.body["use_template_approver"] != float64(1) {
		t.Fatalf("approver flag not sent: %v", rec.body)
	}

	numbers, cursor, err := svc.List(ctx, approval.Query{StartTime: 1, EndTime: 2, Size: 100})
	if err != nil || len(numbers) != 1 || cursor != "c2" ||
		rec.path != "/cgi-bin/oa/getapprovalinfo" {
		t.Fatalf("list: %v %v %q %s", err, numbers, cursor, rec.path)
	}

	detail, err := svc.GetDetail(ctx, "SP1")
	if err != nil || detail.SPName != "报销" || detail.Applier.UserID != "u1" ||
		len(detail.Records) == 0 || rec.path != "/cgi-bin/oa/getapprovaldetail" {
		t.Fatalf("detail: %v %+v %s", err, detail, rec.path)
	}

	if _, _, err := svc.LegacyData(ctx, 1, 2, 0); err != nil ||
		rec.path != "/cgi-bin/corp/getapprovaldata" {
		t.Fatalf("legacy: %v %s", err, rec.path)
	}

	if _, err := svc.OpenApprovalData(ctx, "third1"); err != nil ||
		rec.path != "/cgi-bin/corp/getopenapprovaldata" {
		t.Fatalf("open data: %v %s", err, rec.path)
	}

	if _, err := svc.VacationQuota(ctx, "u1"); err != nil ||
		rec.path != "/cgi-bin/oa/vacation/getuservacationquota" {
		t.Fatalf("quota: %v %s", err, rec.path)
	}
	if _, err := svc.VacationConfig(ctx); err != nil ||
		rec.path != "/cgi-bin/oa/vacation/getcorpconf" {
		t.Fatalf("config: %v %s", err, rec.path)
	}
	if err := svc.SetVacationQuota(ctx, approval.SetQuota{UserID: "u1", VacationID: 1, LeftTime: 10}); err != nil ||
		rec.path != "/cgi-bin/oa/vacation/setoneuserquota" {
		t.Fatalf("set quota: %v %s", err, rec.path)
	}

	if _, err := svc.Submit(ctx, approval.ApplyRequest{CreatorUserID: "u1", TemplateID: "t"}); err == nil {
		t.Fatal("submit without apply_data accepted")
	}
	if _, _, err := svc.List(ctx, approval.Query{}); err == nil {
		t.Fatal("list without window accepted")
	}
	if _, err := svc.GetDetail(ctx, ""); err == nil {
		t.Fatal("detail without sp_no accepted")
	}
	if _, err := svc.VacationQuota(ctx, ""); err == nil {
		t.Fatal("quota without userid accepted")
	}
	if err := svc.SetVacationQuota(ctx, approval.SetQuota{}); err == nil {
		t.Fatal("set quota without ids accepted")
	}
}

func TestApprovalEvent(t *testing.T) {
	payload := `<xml><ToUserName>corp</ToUserName><FromUserName>sys</FromUserName>` +
		`<CreateTime>1571732272</CreateTime><MsgType>event</MsgType>` +
		`<Event><![CDATA[sys_approval_change]]></Event><AgentID>3010040</AgentID>` +
		`<ApprovalInfo><SpNoStr><![CDATA[202506110001]]></SpNoStr><SpNo>202506110001</SpNo>` +
		`<SpName><![CDATA[示例模板]]></SpName><SpStatus>1</SpStatus>` +
		`<TemplateId><![CDATA[tpl]]></TemplateId><ApplyTime>1571728713</ApplyTime>` +
		`<Applyer><UserId><![CDATA[u1]]></UserId><Party><![CDATA[1]]></Party></Applyer>` +
		`<SpRecord><SpStatus>1</SpStatus><ApproverAttr>2</ApproverAttr>` +
		`<Details><Approver><UserId><![CDATA[u2]]></UserId></Approver>` +
		`<Speech><![CDATA[]]></Speech><SpStatus>1</SpStatus><SpTime>0</SpTime></Details>` +
		`</SpRecord></ApprovalInfo></xml>`
	event, err := approval.ParseEvent([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if event.Info.SPNo != "202506110001" || event.Info.SPName != "示例模板" ||
		event.Info.Applier.UserID != "u1" || event.AgentID != 3010040 {
		t.Fatalf("unexpected event: %+v", event.Info)
	}
	if len(event.Info.Records) != 1 || event.Info.Records[0].Details[0].Approver.UserID != "u2" {
		t.Fatalf("records not parsed: %+v", event.Info.Records)
	}
	if _, err := approval.ParseEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
	if _, err := approval.ParseEvent([]byte(`<xml><Event>sys_approval_change</Event></xml>`)); err == nil {
		t.Fatal("event without SpNo accepted")
	}
}
