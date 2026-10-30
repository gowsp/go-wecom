package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/live"
)

func TestLiveLifecycle(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"livingid":"L1","living_code":"code",`+
		`"next_cursor":"n","livingid_list":["L1"],`+
		`"living_info":{"theme":"课程","anchor_userid":"u1","viewer_num":10},`+
		`"ending":1,"next_key":"k2","stat_info":{"users":[{"userid":"u1","watch_time":30}],`+
		`"external_users":[{"external_userid":"e1","watch_time":20}]},`+
		`"viewer_userid":"u1","invitor_userid":"u2"}`)
	svc := live.New(api)
	ctx := context.Background()

	id, err := svc.Create(ctx, live.CreateRequest{
		AnchorUserID: "u1", Theme: "课程", LivingStart: 100, LivingDuration: 600,
	})
	if err != nil || id != "L1" || rec.path != "/cgi-bin/living/create" {
		t.Fatalf("create: %q %v %s", id, err, rec.path)
	}
	if err := svc.Modify(ctx, live.ModifyRequest{LivingID: "L1", Theme: "改名"}); err != nil ||
		rec.path != "/cgi-bin/living/modify" {
		t.Fatalf("modify: %v %s", err, rec.path)
	}
	if err := svc.Cancel(ctx, "L1"); err != nil || rec.path != "/cgi-bin/living/cancel" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}
	if err := svc.DeleteReplay(ctx, "L1"); err != nil ||
		rec.path != "/cgi-bin/living/delete_replay_data" {
		t.Fatalf("delete replay: %v %s", err, rec.path)
	}

	code, err := svc.ShareCode(ctx, "L1", "open1")
	if err != nil || code != "code" || rec.path != "/cgi-bin/living/get_living_code" {
		t.Fatalf("share code: %q %v %s", code, err, rec.path)
	}

	page, err := svc.ListLivingIDs(ctx, "u1", "", 20)
	if err != nil || len(page.LivingIDs) != 1 || page.NextCursor != "n" ||
		rec.path != "/cgi-bin/living/get_user_all_livingid" {
		t.Fatalf("ids: %v %+v %s", err, page, rec.path)
	}

	info, err := svc.Info(ctx, "L1")
	if err != nil || info.Theme != "课程" || info.ViewerNum != 10 ||
		rec.path != "/cgi-bin/living/get_living_info" {
		t.Fatalf("info: %v %+v %s", err, info, rec.path)
	}

	stat, err := svc.WatchStats(ctx, "L1", "")
	if err != nil || stat.Ending != 1 || stat.NextKey != "k2" ||
		len(stat.Users) != 1 || len(stat.ExternalUsers) != 1 ||
		rec.path != "/cgi-bin/living/get_watch_stat" {
		t.Fatalf("stats: %v %+v %s", err, stat, rec.path)
	}

	share, err := svc.ShareInfo(ctx, "CODE")
	if err != nil || share.ViewerUserID != "u1" || share.InvitorUserID != "u2" ||
		rec.path != "/cgi-bin/living/get_living_share_info" {
		t.Fatalf("share info: %v %+v %s", err, share, rec.path)
	}
}

func TestLiveValidation(t *testing.T) {
	_, api := newOffice(t, `{"errcode":0}`)
	svc := live.New(api)
	ctx := context.Background()

	if _, err := svc.Create(ctx, live.CreateRequest{Theme: "t", LivingStart: 1, LivingDuration: 2}); err == nil {
		t.Fatal("missing anchor accepted")
	}
	if _, err := svc.Create(ctx, live.CreateRequest{AnchorUserID: "u", Theme: "t"}); err == nil {
		t.Fatal("missing schedule accepted")
	}
	if err := svc.Modify(ctx, live.ModifyRequest{}); err == nil {
		t.Fatal("modify without id accepted")
	}
	if err := svc.Cancel(ctx, " "); err == nil {
		t.Fatal("cancel without id accepted")
	}
	if _, err := svc.ShareCode(ctx, "L1", ""); err == nil {
		t.Fatal("share code without openid accepted")
	}
	if _, err := svc.ListLivingIDs(ctx, "", "", 0); err == nil {
		t.Fatal("ids without userid accepted")
	}
	if _, err := svc.WatchStats(ctx, "", ""); err == nil {
		t.Fatal("stats without id accepted")
	}
	if _, err := svc.ShareInfo(ctx, " "); err == nil {
		t.Fatal("share info without code accepted")
	}
}

func TestLiveEvent(t *testing.T) {
	payload := `<xml><ToUserName>toUser</ToUserName><CreateTime>1</CreateTime>` +
		`<MsgType>event</MsgType><Event>living_status_change</Event>` +
		`<LivingId>L1</LivingId><Status>1</Status><AgentID>7</AgentID></xml>`
	event, err := live.ParseEvent([]byte(payload))
	if err != nil || event.LivingID != "L1" || event.Status != 1 || event.AgentID != 7 {
		t.Fatalf("event: %v %+v", err, event)
	}
	if _, err := live.ParseEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
}
