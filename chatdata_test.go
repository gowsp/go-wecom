package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/chatdata"
)

func TestChatDataConfig(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"auth_user_list":[{"userid":"u1","edition_list":[1]}],`+
		`"config":{"hide_mobile":true,"hide_idcard":false,"hide_bankno":true},`+
		`"log_level":2,"media_id":"media1","debug_mode_status":1}`)
	svc := chatdata.New(api)
	ctx := context.Background()

	if err := svc.SetPublicKey(ctx, "-----BEGIN PUBLIC KEY-----"); err != nil ||
		rec.path != "/cgi-bin/chatdata/set_public_key" {
		t.Fatalf("public key: %v %s", err, rec.path)
	}
	users, err := svc.ArchiveUsers(ctx, "", 100)
	if err != nil || len(users.Users) != 1 || rec.path != "/cgi-bin/chatdata/get_auth_user_list" {
		t.Fatalf("archive users: %v %+v %s", err, users, rec.path)
	}
	if _, err := svc.FileArchiveUsers(ctx, "", 100); err != nil ||
		rec.path != "/cgi-bin/docdata/get_auth_user_list" {
		t.Fatalf("file archive users: %v %s", err, rec.path)
	}
	if err := svc.SetReceiveCallback(ctx, "prog1"); err != nil ||
		rec.path != "/cgi-bin/chatdata/set_receive_callback" {
		t.Fatalf("callback: %v %s", err, rec.path)
	}

	hidden, err := svc.Hidden(ctx, "u1")
	if err != nil || !hidden.HideMobile ||
		rec.path != "/cgi-bin/chatdata/get_hide_sensitiveinfo_config" {
		t.Fatalf("hidden: %v %+v %s", err, hidden, rec.path)
	}
	if err := svc.SetHidden(ctx, "u1", chatdata.HiddenConfig{HideBankNo: true}); err != nil ||
		rec.path != "/cgi-bin/chatdata/set_hide_sensitiveinfo_config" {
		t.Fatalf("set hidden: %v %s", err, rec.path)
	}

	level, err := svc.LogLevel(ctx, "prog1")
	if err != nil || level != 2 || rec.path != "/cgi-bin/chatdata/get_log_level" {
		t.Fatalf("log level: %v %d %s", err, level, rec.path)
	}
	if err := svc.SetLogLevel(ctx, "prog1", 3); err != nil || rec.path != "/cgi-bin/chatdata/set_log_level" {
		t.Fatalf("set log level: %v %s", err, rec.path)
	}

	mediaID, err := svc.UploadMedia(ctx, "a.txt", []byte("data"))
	if err != nil || mediaID != "media1" || rec.path != "/cgi-bin/chatdata/upload_media" {
		t.Fatalf("upload: %q %v %s", mediaID, err, rec.path)
	}

	status, err := svc.DebugMode(ctx, "prog1")
	if err != nil || status != 1 || rec.path != "/cgi-bin/chatdata/check_debug_mode" {
		t.Fatalf("debug mode: %v %d %s", err, status, rec.path)
	}
	if err := svc.OpenDebugMode(ctx, "prog1", "token"); err != nil ||
		rec.path != "/cgi-bin/chatdata/open_debug_mode" {
		t.Fatalf("open debug: %v %s", err, rec.path)
	}
	if err := svc.CloseDebugMode(ctx, "prog1"); err != nil ||
		rec.path != "/cgi-bin/chatdata/close_debug_mode" {
		t.Fatalf("close debug: %v %s", err, rec.path)
	}

	if err := svc.SetPublicKey(ctx, " "); err == nil {
		t.Fatal("empty public key accepted")
	}
	if err := svc.SetReceiveCallback(ctx, ""); err == nil {
		t.Fatal("callback without program accepted")
	}
	if _, err := svc.Hidden(ctx, ""); err == nil {
		t.Fatal("hidden without user accepted")
	}
	if _, err := svc.DebugMode(ctx, ""); err == nil {
		t.Fatal("debug mode without program accepted")
	}
	if err := svc.OpenDebugMode(ctx, "p", ""); err == nil {
		t.Fatal("open debug without token accepted")
	}
}

func TestChatDataProgram(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"response_data":"{\"output\":\"ok\"}",`+
		`"jobid":"job1","status":1}`)
	svc := chatdata.New(api)
	ctx := context.Background()

	response, err := svc.CallProgram(ctx, "prog1", "ability1", "", []byte(`{"input":"x"}`))
	if err != nil || len(response) == 0 || rec.path != "/cgi-bin/chatdata/sync_call_program" {
		t.Fatalf("call: %v %s %s", err, response, rec.path)
	}
	if rec.body["request_data"] != `{"input":"x"}` {
		t.Fatalf("request_data not stringified: %v", rec.body["request_data"])
	}

	jobID, err := svc.SubmitProgramTask(ctx, "prog1", "ability1", []byte(`{"input":"y"}`))
	if err != nil || jobID != "job1" || rec.path != "/cgi-bin/chatdata/async_program_task" {
		t.Fatalf("task: %q %v %s", jobID, err, rec.path)
	}
	if _, err := svc.ProgramResult(ctx, "job1"); err != nil ||
		rec.path != "/cgi-bin/chatdata/async_program_result" {
		t.Fatalf("result: %v %s", err, rec.path)
	}

	if _, err := svc.CallProgram(ctx, "", "", "", nil); err == nil {
		t.Fatal("call without ids accepted")
	}
	if _, err := svc.CallProgram(ctx, "p", "a", "", []byte(`{`)); err == nil {
		t.Fatal("invalid request accepted")
	}
	if _, err := svc.SubmitProgramTask(ctx, "p", "", nil); err == nil {
		t.Fatal("task without ability accepted")
	}
	if _, err := svc.ProgramResult(ctx, ""); err == nil {
		t.Fatal("result without job accepted")
	}
}

func TestChatDataNotify(t *testing.T) {
	payload := `<xml><MsgType>event</MsgType><Event><![CDATA[program_notify]]></Event>` +
		`<NotifyId><![CDATA[n1]]></NotifyId><NotifyScene>1</NotifyScene></xml>`
	notify, err := chatdata.ParseNotify([]byte(payload))
	if err != nil || notify.NotifyID != "n1" || notify.Scene != 1 {
		t.Fatalf("notify: %v %+v", err, notify)
	}
	if _, err := chatdata.ParseNotify([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
	if _, err := chatdata.ParseNotify([]byte(`<xml><Event>program_notify</Event></xml>`)); err == nil {
		t.Fatal("notify without id accepted")
	}
}
