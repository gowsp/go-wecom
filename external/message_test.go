package external

import (
	"context"
	"strings"
	"testing"
)

func TestTransferCustomers(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"customer":[{"external_userid":"e1","errcode":0},{"external_userid":"e2","errcode":40096}]}`)
	failed, err := api.TransferCustomers(context.Background(), TransferRequest{
		HandoverUserID: "u1", TakeoverUserID: "u2", ExternalUserID: []string{"e1", "e2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/transfer_customer" {
		t.Fatalf("path = %s", rec.path)
	}
	if len(failed) != 2 || failed[1].Errcode != 40096 {
		t.Fatalf("unexpected result: %+v", failed)
	}
	if _, err := api.TransferCustomers(context.Background(), TransferRequest{
		HandoverUserID: "u1", TakeoverUserID: "u2",
	}); err == nil {
		t.Fatal("empty customer list accepted")
	}
}

func TestTransferResult(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"customer":[{"external_userid":"e1","status":2,"takeover_time":10}],"next_cursor":"n"}`)
	page, err := api.TransferResult(context.Background(), "u1", "u2", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/transfer_result" ||
		!strings.Contains(rec.query, "handover_userid=u1") ||
		!strings.Contains(rec.query, "takeover_userid=u2") ||
		!strings.Contains(rec.query, "cursor=c1") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if page.Customer[0].Status != 2 || page.NextCursor != "n" {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestTransferGroupChats(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"failed_chat_list":[{"chat_id":"c2","errcode":90500,"errmsg":"not resigned"}]}`)
	ctx := context.Background()

	failed, err := api.TransferGroupChats(ctx, []string{"c1", "c2"}, "u2")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/groupchat/transfer" ||
		len(rec.body["chat_id_list"].([]any)) != 2 {
		t.Fatalf("unexpected request %s %v", rec.path, rec.body)
	}
	if len(failed) != 1 || failed[0].ChatID != "c2" || failed[0].Errcode != 90500 {
		t.Fatalf("unexpected failed: %+v", failed)
	}
	if _, err := api.TransferGroupChats(ctx, nil, "u2"); err == nil {
		t.Fatal("empty chat list accepted")
	}

	if _, err := api.TransferOnJobGroupChats(ctx, []string{"c1"}, "u2"); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/groupchat/onjob_transfer" {
		t.Fatalf("path = %s", rec.path)
	}
}

func TestUnassignedAndResignedTransfer(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n","info":[`+
		`{"handover_userid":"u1","external_userid":"e1","dimission_time":10}],`+
		`"customer":[{"external_userid":"e1","status":1,"takeover_time":20}]}`)
	ctx := context.Background()

	page, err := api.ListUnassignedCustomers(ctx, "", 100)
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_unassigned_list" {
		t.Fatalf("unassigned: %v %s", err, rec.path)
	}
	if page.NextCursor != "n" || len(page.Info) != 1 || page.Info[0].HandoverUserID != "u1" {
		t.Fatalf("unexpected page: %+v", page)
	}
	if rec.body["page_size"] != float64(100) {
		t.Fatalf("page size not sent: %v", rec.body)
	}
	if _, err := api.ListUnassignedCustomers(ctx, "", 5000); err == nil {
		t.Fatal("oversized page accepted")
	}

	result, err := api.TransferResignedResult(ctx, "u1", "u2", "c1")
	if err != nil || rec.path != "/cgi-bin/externalcontact/resigned/transfer_result" {
		t.Fatalf("resigned result: %v %s", err, rec.path)
	}
	if len(result.Customer) != 1 || result.Customer[0].Status != 1 || rec.body["cursor"] != "c1" {
		t.Fatalf("unexpected result: %+v (body %v)", result, rec.body)
	}
	if _, err := api.TransferResignedResult(ctx, "", "u2", ""); err == nil {
		t.Fatal("missing handover accepted")
	}
}

func TestTransferResignedCustomers(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"customer":[{"external_userid":"e1","errcode":0}]}`)
	_, err := api.TransferResignedCustomers(context.Background(), TransferRequest{
		HandoverUserID: "u1", TakeoverUserID: "u2", ExternalUserID: []string{"e1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/resigned/transfer_customer" {
		t.Fatalf("path = %s", rec.path)
	}
	if _, err := api.TransferResignedCustomers(context.Background(), TransferRequest{
		HandoverUserID: "u1", TakeoverUserID: "u1", ExternalUserID: []string{"e1"},
	}); err == nil {
		t.Fatal("same handover and takeover accepted")
	}
}

func TestGroupSend(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"msgid":"m1","fail_list":["e2"]}`)
	ctx := context.Background()

	result, err := api.AddMsgTemplate(ctx, MsgTemplate{
		ChatType:       "single",
		ExternalUserID: []string{"e1", "e2"},
		Sender:         "u1",
		Text:           &TextContent{Content: "hi"},
		Attachments: []Attachment{{
			MsgType: "link", Link: &LinkAttachment{Title: "t", URL: "https://x"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/add_msg_template" ||
		result.MsgID != "m1" || len(result.FailList) != 1 {
		t.Fatalf("unexpected template result %s %+v", rec.path, result)
	}
	attachments := rec.body["attachments"].([]any)
	if attachments[0].(map[string]any)["msgtype"] != "link" {
		t.Fatalf("attachment not sent: %v", rec.body)
	}

	if _, err := api.AddMsgTemplate(ctx, MsgTemplate{Sender: "u1"}); err == nil {
		t.Fatal("empty content accepted")
	}
	if _, err := api.AddMsgTemplate(ctx, MsgTemplate{Text: &TextContent{Content: "x"}}); err == nil {
		t.Fatal("missing sender accepted")
	}

	if err := api.RemindGroupMsgSend(ctx, "m1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/remind_groupmsg_send" {
		t.Fatalf("remind: %v %s", err, rec.path)
	}
	if err := api.CancelGroupMsgSend(ctx, "m1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/cancel_groupmsg_send" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}
	if err := api.CancelGroupMsgSend(ctx, ""); err == nil {
		t.Fatal("empty msgid accepted")
	}
}

func TestGroupSendRecords(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n","group_msg_list":[{"msgid":"m1","creator":"u1"}]}`)
	ctx := context.Background()

	page, err := api.ListGroupMsgTasks(ctx, GroupMsgQuery{ChatType: "single", Limit: 50})
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_groupmsg_list_v2" || len(page.Tasks) != 1 {
		t.Fatalf("list tasks: %v %s %+v", err, rec.path, page)
	}

	if _, err := api.GetGroupMsgTask(ctx, "m1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/get_groupmsg_task" {
		t.Fatalf("get task: %v %s", err, rec.path)
	}

	if _, err := api.ListGroupMsgSendResults(ctx, "m1", "u1", 20, ""); err != nil ||
		rec.path != "/cgi-bin/externalcontact/get_groupmsg_send_result" {
		t.Fatalf("send results: %v %s", err, rec.path)
	}
	if _, err := api.ListGroupMsgSendResults(ctx, "", "", 0, ""); err == nil {
		t.Fatal("empty msgid accepted")
	}
}

func TestSendWelcomeMsg(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0}`)
	ctx := context.Background()

	if err := api.SendWelcomeMsg(ctx, WelcomeMessage{
		WelcomeCode: "code1",
		Text:        &TextContent{Content: "欢迎"},
	}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/send_welcome_msg" || rec.body["welcome_code"] != "code1" {
		t.Fatalf("unexpected request %s %v", rec.path, rec.body)
	}
	if err := api.SendWelcomeMsg(ctx, WelcomeMessage{}); err == nil {
		t.Fatal("missing welcome code accepted")
	}
	if err := api.SendWelcomeMsg(ctx, WelcomeMessage{WelcomeCode: "c"}); err == nil {
		t.Fatal("empty content accepted")
	}
}

func TestWelcomeTemplates(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"template_id":"tpl1","text":{"content":"hi"}}`)
	ctx := context.Background()

	id, err := api.AddWelcomeTemplate(ctx, WelcomeTemplate{
		Text: &TextContent{Content: "欢迎 %NICKNAME%"}, AgentID: 1000002,
	})
	if err != nil || id != "tpl1" {
		t.Fatalf("add template: %q %v", id, err)
	}
	if rec.path != "/cgi-bin/externalcontact/group_welcome_template/add" {
		t.Fatalf("path = %s", rec.path)
	}

	tpl, err := api.GetWelcomeTemplate(ctx, "tpl1")
	if err != nil || tpl.TemplateID != "tpl1" || tpl.Text == nil || tpl.Text.Content != "hi" {
		t.Fatalf("get template: %v %+v", err, tpl)
	}

	if err := api.UpdateWelcomeTemplate(ctx, WelcomeTemplate{
		TemplateID: "tpl1", Text: &TextContent{Content: "改"},
	}); err != nil || rec.path != "/cgi-bin/externalcontact/group_welcome_template/edit" {
		t.Fatalf("update template: %v %s", err, rec.path)
	}
	if err := api.DeleteWelcomeTemplate(ctx, "tpl1", 0); err != nil ||
		rec.path != "/cgi-bin/externalcontact/group_welcome_template/del" {
		t.Fatalf("delete template: %v %s", err, rec.path)
	}
	if _, err := api.AddWelcomeTemplate(ctx, WelcomeTemplate{}); err == nil {
		t.Fatal("empty template accepted")
	}
	if err := api.UpdateWelcomeTemplate(ctx, WelcomeTemplate{Text: &TextContent{Content: "x"}}); err == nil {
		t.Fatal("update without id accepted")
	}
}
