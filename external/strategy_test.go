package external

import (
	"context"
	"strings"
	"testing"
)

func TestCustomerStrategyLifecycle(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"strategy_id":7,`+
		`"strategy":{"strategy_id":7,"strategy_name":"默认","privilege":{"view_customer_list":true}}}`)
	ctx := context.Background()

	id, err := api.CreateCustomerStrategy(ctx, CustomerStrategy{
		Name: "默认", AdminList: []string{"u1"},
		Privilege: &CustomerPrivilege{ViewCustomerList: true},
	})
	if err != nil || id != 7 {
		t.Fatalf("create: %d %v", id, err)
	}
	if rec.path != "/cgi-bin/externalcontact/customer_strategy/create" ||
		rec.body["strategy_name"] != "默认" {
		t.Fatalf("unexpected create %s %v", rec.path, rec.body)
	}
	privilege := rec.body["privilege"].(map[string]any)
	if privilege["view_customer_list"] != true {
		t.Fatalf("privilege not sent: %v", privilege)
	}
	if _, ok := privilege["send_moment"]; ok {
		t.Fatal("moment-only field must not appear in a customer privilege")
	}
	if privilege["share_customer"] != false {
		t.Fatalf("false privilege must be explicit: %v", privilege)
	}

	strategy, err := api.GetCustomerStrategy(ctx, 7)
	if err != nil || rec.path != "/cgi-bin/externalcontact/customer_strategy/get" {
		t.Fatalf("get: %v %s", err, rec.path)
	}
	if strategy.Name != "默认" || !strategy.Privilege.ViewCustomerList {
		t.Fatalf("unexpected strategy: %+v", strategy)
	}

	if err := api.UpdateCustomerStrategy(ctx, CustomerStrategy{StrategyID: 7, Name: "改名"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/customer_strategy/edit" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := api.DeleteCustomerStrategy(ctx, 7); err != nil ||
		rec.path != "/cgi-bin/externalcontact/customer_strategy/del" ||
		rec.body["strategy_id"] != float64(7) {
		t.Fatalf("delete: %v %s %v", err, rec.path, rec.body)
	}
}

func TestListCustomerStrategies(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"strategy":[{"strategy_id":7}],"next_cursor":"n"}`)
	page, err := api.ListCustomerStrategies(context.Background(), "", 100)
	if err != nil || rec.path != "/cgi-bin/externalcontact/customer_strategy/list" ||
		len(page.Strategies) != 1 || page.NextCursor != "n" {
		t.Fatalf("list: %v %s %+v", err, rec.path, page)
	}
	if rec.body["limit"] != float64(100) {
		t.Fatalf("limit not sent: %v", rec.body)
	}
}

func TestCustomerStrategyRange(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n","range":[{"type":1,"userid":["u1"],"partyid":[2]}]}`)
	page, err := api.CustomerStrategyRange(context.Background(), 7, "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/customer_strategy/get_range" {
		t.Fatalf("path = %s", rec.path)
	}
	if len(page.Range) != 1 || page.Range[0].UserID[0] != "u1" || page.NextCursor != "n" {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestMomentStrategy(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"strategy_id":9,"strategy":{"strategy_id":9,"privilege":{"send_moment":true}}}`)
	ctx := context.Background()

	id, err := api.CreateMomentStrategy(ctx, MomentStrategy{
		Name: "朋友圈规则", Privilege: &MomentPrivilege{SendMoment: true},
	})
	if err != nil || id != 9 || rec.path != "/cgi-bin/externalcontact/moment_strategy/create" {
		t.Fatalf("create: %d %v %s", id, err, rec.path)
	}
	strategy, err := api.GetMomentStrategy(ctx, 9)
	if err != nil || !strategy.Privilege.SendMoment {
		t.Fatalf("get: %v %+v", err, strategy)
	}
	if err := api.UpdateMomentStrategy(ctx, MomentStrategy{StrategyID: 9, Name: "x"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/moment_strategy/edit" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := api.DeleteMomentStrategy(ctx, 9); err != nil ||
		rec.path != "/cgi-bin/externalcontact/moment_strategy/del" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}
	if _, err := api.MomentStrategyRange(ctx, 9, "", 100); err != nil ||
		rec.path != "/cgi-bin/externalcontact/moment_strategy/get_range" {
		t.Fatalf("range: %v %s", err, rec.path)
	}
}

func TestStrategyTags(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"tag_group":[{"group_id":"g1","group_name":"组","strategy_id":7,`+
		`"tag":[{"id":"t1","name":"标签"}]}]}`)
	ctx := context.Background()

	groups, err := api.ListStrategyTags(ctx, 7, "", "", "", 100)
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_strategy_tag_list" || len(groups) != 1 {
		t.Fatalf("list tags: %v %s %+v", err, rec.path, groups)
	}
	if groups[0].GroupID != "g1" || groups[0].Tags[0].Name != "标签" {
		t.Fatalf("unexpected groups: %+v", groups)
	}

	if _, err := api.AddStrategyTags(ctx, 7, "g1", "", []CorpTag{{Name: "新"}}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/add_strategy_tag" {
		t.Fatalf("add tags: %v %s", err, rec.path)
	}
	if err := api.EditStrategyTag(ctx, "t1", "改名", 1); err != nil ||
		rec.path != "/cgi-bin/externalcontact/edit_strategy_tag" {
		t.Fatalf("edit tag: %v %s", err, rec.path)
	}
	if err := api.DeleteStrategyTags(ctx, 7, []string{"t1"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/del_strategy_tag" {
		t.Fatalf("delete tags: %v %s", err, rec.path)
	}

	if _, err := api.ListStrategyTags(ctx, 0, "", "", "", 0); err == nil {
		t.Fatal("missing strategy id accepted")
	}
	if _, err := api.AddStrategyTags(ctx, 7, "", "", nil); err == nil {
		t.Fatal("empty tag body accepted")
	}
	if err := api.DeleteStrategyTags(ctx, 7, nil); err == nil {
		t.Fatal("empty tag ids accepted")
	}
	if err := api.EditStrategyTag(ctx, "", "x", 0); err == nil {
		t.Fatal("missing tag id accepted")
	}
}

func TestMomentTask(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"jobid":"job1","status":1,"type":"add_moment_task",`+
		`"result":{"errcode":0,"moment_id":"m1"}}`)
	ctx := context.Background()

	jobID, err := api.AddMomentTask(ctx, MomentTask{
		Text: &TextContent{Content: "内容"},
		Attachments: []MomentAttachment{{
			MsgType: "image", Image: &MediaAttachment{MediaID: "MEDIA"},
		}},
		VisibleRange: &MomentRange{
			SenderList:          &MomentSenderList{UserList: []string{"u1"}},
			ExternalContactList: &MomentContactList{TagList: []string{"et1"}},
		},
	})
	if err != nil || jobID != "job1" || rec.path != "/cgi-bin/externalcontact/add_moment_task" {
		t.Fatalf("add moment: %q %v %s", jobID, err, rec.path)
	}
	rangeBlock := rec.body["visible_range"].(map[string]any)
	if rangeBlock["sender_list"] == nil || rangeBlock["external_contact_list"] == nil {
		t.Fatalf("visible range not sent: %v", rangeBlock)
	}

	job, err := api.GetMomentTaskResult(ctx, "job1")
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_moment_task_result" ||
		!strings.Contains(rec.query, "jobid=job1") {
		t.Fatalf("result: %v %s?%s", err, rec.path, rec.query)
	}
	if job.Status != 1 || job.Result.MomentID != "m1" {
		t.Fatalf("unexpected job: %+v", job)
	}

	if err := api.CancelMomentTask(ctx, "m1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/cancel_moment_task" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}
	if _, err := api.AddMomentTask(ctx, MomentTask{}); err == nil {
		t.Fatal("empty moment accepted")
	}
	if _, err := api.GetMomentTaskResult(ctx, ""); err == nil {
		t.Fatal("missing jobid accepted")
	}
	if err := api.CancelMomentTask(ctx, " "); err == nil {
		t.Fatal("missing moment id accepted")
	}
}
