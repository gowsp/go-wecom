package wecom_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gowsp/go-wecom/checkin"
)

func TestCheckinQueries(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,`+
		`"checkindata":[{"userid":"u1","checkin_time":10,"groupname":"一组","device_sn":"sn1"}],`+
		`"datas":[{"base_info":{"name":"张三"}}],`+
		`"schedule_list":[{"userid":"u1","yearmonth":202011}],`+
		`"group":[{"groupid":1,"grouptype":1}],`+
		`"info":[{"userid":"u1","group":{"groupid":1}}]}`)
	svc := checkin.New(api)
	ctx := context.Background()

	records, err := svc.Records(ctx, checkin.DataTypeExceptionOnly, 1, 2, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/checkin/getcheckindata" || len(records) != 1 {
		t.Fatalf("records: %v %s %+v", err, rec.path, records)
	}
	if rec.body["opencheckindatatype"] != float64(3) {
		t.Fatalf("data type not sent: %v", rec.body)
	}
	if records[0].GroupName != "一组" {
		t.Fatalf("unexpected record: %+v", records[0])
	}

	hardware, err := svc.HardwareRecords(ctx, 1, 1, 2, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/hardware/get_hardware_checkin_data" || len(hardware) != 1 {
		t.Fatalf("hardware: %v %s %+v", err, rec.path, hardware)
	}
	if hardware[0].DeviceSN != "sn1" {
		t.Fatalf("unexpected hardware record: %+v", hardware[0])
	}

	day, err := svc.DayData(ctx, 1, 2, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/checkin/getcheckin_daydata" || len(day) != 1 {
		t.Fatalf("day data: %v %s %v", err, rec.path, day)
	}
	var decoded struct {
		BaseInfo struct {
			Name string `json:"name"`
		} `json:"base_info"`
	}
	if err := json.Unmarshal(day[0], &decoded); err != nil || decoded.BaseInfo.Name != "张三" {
		t.Fatalf("raw day data: %v %+v", err, decoded)
	}

	if _, err := svc.MonthData(ctx, 1, 2, []string{"u1"}); err != nil ||
		rec.path != "/cgi-bin/checkin/getcheckin_monthdata" {
		t.Fatalf("month data: %v %s", err, rec.path)
	}

	schedules, err := svc.Schedules(ctx, 1, 2, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/checkin/getcheckinschedulist" || len(schedules) != 1 {
		t.Fatalf("schedules: %v %s %v", err, rec.path, schedules)
	}

	groups, err := svc.Options(ctx)
	if err != nil || rec.path != "/cgi-bin/checkin/getcorpcheckinoption" || len(groups) != 1 {
		t.Fatalf("options: %v %s %v", err, rec.path, groups)
	}

	infos, err := svc.UserOptions(ctx, 100, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/checkin/getcheckinoption" || len(infos) != 1 {
		t.Fatalf("user options: %v %s %v", err, rec.path, infos)
	}
}

func TestCheckinMutations(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0}`)
	svc := checkin.New(api)
	ctx := context.Background()

	if err := svc.SetSchedules(ctx, 226, 202012, []checkin.ScheduleItem{
		{UserID: "u1", Day: 5, ScheduleID: 234},
	}); err != nil || rec.path != "/cgi-bin/checkin/setcheckinschedulist" {
		t.Fatalf("set schedules: %v %s", err, rec.path)
	}
	if rec.body["yearmonth"] != float64(202012) {
		t.Fatalf("yearmonth not sent: %v", rec.body)
	}

	if err := svc.CorrectPunch(ctx, checkin.Correction{
		UserID: "u1", ScheduleDateTime: 100, CheckinTime: 200, Remark: "补卡",
	}); err != nil || rec.path != "/cgi-bin/checkin/punch_correction" {
		t.Fatalf("correct: %v %s", err, rec.path)
	}

	if err := svc.AddRecords(ctx, []checkin.NewRecord{{UserID: "u1", CheckinTime: 300}}); err != nil ||
		rec.path != "/cgi-bin/checkin/add_checkin_record" {
		t.Fatalf("add records: %v %s", err, rec.path)
	}

	if err := svc.AddFace(ctx, "u1", "BASE64"); err != nil ||
		rec.path != "/cgi-bin/checkin/addcheckinuserface" {
		t.Fatalf("add face: %v %s", err, rec.path)
	}

	option := json.RawMessage(`{"effective_now":true,"group":{"grouptype":1}}`)
	if err := svc.AddOption(ctx, option); err != nil ||
		rec.path != "/cgi-bin/checkin/add_checkin_option" {
		t.Fatalf("add option: %v %s", err, rec.path)
	}
	if err := svc.ClearOptionField(ctx, 226, "special_days"); err != nil ||
		rec.path != "/cgi-bin/checkin/clear_checkin_option_array_field" {
		t.Fatalf("clear field: %v %s", err, rec.path)
	}
}

func TestCheckinValidation(t *testing.T) {
	_, api := newOffice(t, `{"errcode":0}`)
	svc := checkin.New(api)
	ctx := context.Background()

	if _, err := svc.Records(ctx, 1, 0, 0, []string{"u1"}); err == nil {
		t.Fatal("invalid window accepted")
	}
	if _, err := svc.Records(ctx, 1, 1, 2, nil); err == nil {
		t.Fatal("empty user list accepted")
	}
	if err := svc.SetSchedules(ctx, 0, 202012, []checkin.ScheduleItem{{UserID: "u"}}); err == nil {
		t.Fatal("missing group accepted")
	}
	if err := svc.CorrectPunch(ctx, checkin.Correction{}); err == nil {
		t.Fatal("empty correction accepted")
	}
	if err := svc.AddRecords(ctx, nil); err == nil {
		t.Fatal("empty records accepted")
	}
	if err := svc.AddRecords(ctx, []checkin.NewRecord{{UserID: "u"}}); err == nil {
		t.Fatal("record without time accepted")
	}
	if err := svc.AddFace(ctx, "", ""); err == nil {
		t.Fatal("empty face accepted")
	}
	if err := svc.AddOption(ctx, json.RawMessage(`{`)); err == nil {
		t.Fatal("invalid json accepted")
	}
	if err := svc.ClearOptionField(ctx, 0, ""); err == nil {
		t.Fatal("empty field accepted")
	}
	if _, err := svc.UserOptions(ctx, 0, nil); err == nil {
		t.Fatal("empty user options accepted")
	}
}
