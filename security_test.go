package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/security"
)

func TestSecurityRecords(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"record_list":[{"time":1,"userid":"u1","operation":{"type":101},`+
		`"file_info":"a.jpg","file_size":10,"screen_shot_type":1,"system":"iPad"}],`+
		`"device_list":[{"device_code":"d1","system":"Windows"}],`+
		`"result":[{"device_index":1,"device_code":"d1","status":1}],`+
		`"success_list":["d1"],"fail_list":["d9"],`+
		`"domain_list":[{"domain":"i.work.weixin.qq.com","port":[443],"is_necessary":1}]}`)
	svc := security.New(api)
	ctx := context.Background()

	files, err := svc.FileRecords(ctx, 1, 2, []string{"u1"},
		&security.FileOperation{Type: 103, Source: 401}, "", 100)
	if err != nil || len(files.Records) != 1 || files.Records[0].FileInfo != "a.jpg" ||
		rec.path != "/cgi-bin/security/get_file_oper_record" {
		t.Fatalf("files: %v %+v %s", err, files, rec.path)
	}
	if rec.body["start_time"] != float64(1) {
		t.Fatalf("window not sent: %v", rec.body)
	}

	screens, err := svc.ScreenRecords(ctx, 1, 2, []string{"u1"}, []int{2}, 1, "", 100)
	if err != nil || len(screens.Records) != 1 ||
		rec.path != "/cgi-bin/security/get_screen_oper_record" {
		t.Fatalf("screens: %v %+v %s", err, screens, rec.path)
	}
	if _, ok := rec.body["department_id_list"]; !ok {
		t.Fatalf("departments not sent: %v", rec.body)
	}

	imports, err := svc.ImportTrustDevices(ctx, []security.Device{{System: "Windows"}}, security.DeviceCorporate)
	if err != nil || len(imports) != 1 || rec.path != "/cgi-bin/security/trustdevice/import" {
		t.Fatalf("import: %v %+v %s", err, imports, rec.path)
	}
	if _, err := svc.TrustDevices(ctx, security.DeviceCorporate, "", 100); err != nil ||
		rec.path != "/cgi-bin/security/trustdevice/list" {
		t.Fatalf("devices: %v %s", err, rec.path)
	}
	devices, err := svc.UserDevices(ctx, "u1", security.DevicePersonal)
	if err != nil || len(devices) != 1 || rec.path != "/cgi-bin/security/trustdevice/get_by_user" {
		t.Fatalf("user devices: %v %+v %s", err, devices, rec.path)
	}
	if err := svc.DeleteTrustDevices(ctx, security.DeviceCorporate, []string{"d1"}); err != nil ||
		rec.path != "/cgi-bin/security/trustdevice/delete" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}
	approved, err := svc.ApproveDevices(ctx, []string{"d1"})
	if err != nil || len(approved.Success) != 1 || rec.path != "/cgi-bin/security/trustdevice/approve" {
		t.Fatalf("approve: %v %+v %s", err, approved, rec.path)
	}
	if _, err := svc.RejectDevices(ctx, []string{"d1"}); err != nil ||
		rec.path != "/cgi-bin/security/trustdevice/reject" {
		t.Fatalf("reject: %v %s", err, rec.path)
	}

	for _, run := range []struct {
		name string
		fn   func() (security.OperLogPage, error)
		path string
	}{
		{"member", func() (security.OperLogPage, error) {
			return svc.MemberOperLogs(ctx, 1, 2, 1, "u1", "", 10)
		}, "/cgi-bin/security/member_oper_log/list"},
		{"admin", func() (security.OperLogPage, error) {
			return svc.AdminOperLogs(ctx, 1, 2, 1, "u1", "", 10)
		}, "/cgi-bin/security/admin_oper_log/list"},
	} {
		if _, err := run.fn(); err != nil || rec.path != run.path {
			t.Fatalf("%s logs: %v %s", run.name, err, rec.path)
		}
	}

	domains, err := svc.Domains(ctx)
	if err != nil || len(domains) != 1 || rec.path != "/cgi-bin/security/get_server_domain_ip" {
		t.Fatalf("domains: %v %+v %s", err, domains, rec.path)
	}

	if _, err := svc.FileRecords(ctx, 0, 0, nil, nil, "", 0); err == nil {
		t.Fatal("invalid window accepted")
	}
	if _, err := svc.ImportTrustDevices(ctx, nil, 0); err == nil {
		t.Fatal("empty device list accepted")
	}
	if _, err := svc.UserDevices(ctx, "", 0); err == nil {
		t.Fatal("user devices without id accepted")
	}
	if err := svc.DeleteTrustDevices(ctx, 0, nil); err == nil {
		t.Fatal("delete without codes accepted")
	}
	if _, err := svc.ApproveDevices(ctx, nil); err == nil {
		t.Fatal("approve without codes accepted")
	}
	if _, err := svc.RejectDevices(ctx, nil); err == nil {
		t.Fatal("reject without codes accepted")
	}
}

func TestSecurityEvent(t *testing.T) {
	payload := `<xml><MsgType>event</MsgType><Event><![CDATA[security]]></Event>` +
		`<ChangeType><![CDATA[change_domain_ip]]></ChangeType></xml>`
	event, err := security.ParseEvent([]byte(payload))
	if err != nil || event.ChangeType != "change_domain_ip" {
		t.Fatalf("event: %v %+v", err, event)
	}
	if _, err := security.ParseEvent([]byte(
		`<xml><Event>security</Event><ChangeType>other</ChangeType></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}
	if _, err := security.ParseEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
}

func TestSecurityVIP(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"jobid":"job1","invalid_userid_list":["u9"],`+
		`"status":1,"userid_list":["u1"],"has_more":true,"next_cursor":"c2"}`)
	svc := security.New(api)
	ctx := context.Background()

	job, err := svc.SubmitAddSecurityVIP(ctx, []string{"u1"})
	if err != nil || job.JobID != "job1" || rec.path != "/cgi-bin/security/vip/submit_batch_add_job" {
		t.Fatalf("add: %v %+v %s", err, job, rec.path)
	}
	if _, err := svc.AddSecurityVIPResult(ctx, "job1"); err != nil ||
		rec.path != "/cgi-bin/security/vip/batch_add_job_result" {
		t.Fatalf("add result: %v %s", err, rec.path)
	}
	if _, err := svc.SubmitRemoveSecurityVIP(ctx, []string{"u1"}); err != nil ||
		rec.path != "/cgi-bin/security/vip/submit_batch_del_job" {
		t.Fatalf("remove: %v %s", err, rec.path)
	}
	if _, err := svc.RemoveSecurityVIPResult(ctx, "job1"); err != nil ||
		rec.path != "/cgi-bin/security/vip/batch_del_job_result" {
		t.Fatalf("remove result: %v %s", err, rec.path)
	}
	page, err := svc.ListSecurityVIP(ctx, "", 2)
	if err != nil || len(page.UserIDs) != 1 || rec.path != "/cgi-bin/security/vip/list" {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}
	if _, err := svc.AddSecurityVIPResult(ctx, ""); err == nil {
		t.Fatal("result without job accepted")
	}
}
