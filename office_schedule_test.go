package wecom_test

import (
	"context"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/schedule"
)

func TestCalendars(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"cal_id":"cal1",`+
		`"fail_result":{"shares":[{"userid":"u3","errcode":40001,"errmsg":"not found"}]},`+
		`"calendar_list":[{"cal_id":"cal1","summary":"默认日历"}]}`)
	svc := schedule.New(api)
	ctx := context.Background()

	result, err := svc.AddCalendar(ctx, schedule.Calendar{
		Summary: "默认日历", Admins: []string{"u1"}, Shares: []schedule.Share{{UserID: "u2", Permission: 1}},
	}, 1000002)
	if err != nil || result.CalID != "cal1" || rec.path != "/cgi-bin/oa/calendar/add" {
		t.Fatalf("add: %v %+v %s", err, result, rec.path)
	}
	if len(result.FailResult.Shares) != 1 || result.FailResult.Shares[0].Errcode != 40001 {
		t.Fatalf("failed shares: %+v", result.FailResult.Shares)
	}
	if rec.body["agentid"] != float64(1000002) {
		t.Fatalf("agentid not sent: %v", rec.body)
	}

	rec.body = nil // the recorder merges into a reused map; start clean
	// The endpoint is an overwrite: summary and color are required, and the
	// field set is the update body, not a patch of the created calendar.
	if _, err := svc.UpdateCalendar(ctx, schedule.CalendarUpdate{
		CalID: "cal1", Summary: "改名", Color: "#FF3030",
		Shares: []schedule.Share{{UserID: "u2", Permission: 1}},
	}); err != nil || rec.path != "/cgi-bin/oa/calendar/update" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	inner := rec.body["calendar"].(map[string]any)
	if inner["summary"] != "改名" || inner["color"] != "#FF3030" || len(inner) != 4 {
		t.Fatalf("overwrite body not sent: %v", inner)
	}
	if rec.body["skip_public_range"] != float64(0) {
		t.Fatalf("skip_public_range not sent: %v", rec.body)
	}

	// Skipping the subscribable range is the only way to leave it alone.
	if _, err := svc.UpdateCalendar(ctx, schedule.CalendarUpdate{
		CalID: "cal1", Summary: "改名", Color: "#FF3030", SkipPublicRange: true,
	}); err != nil {
		t.Fatalf("skip update: %v", err)
	}
	if rec.body["skip_public_range"] != float64(1) {
		t.Fatalf("skip_public_range not raised: %v", rec.body)
	}
	inner = rec.body["calendar"].(map[string]any)
	if _, ok := inner["is_public"]; ok {
		t.Fatalf("is_public is not updatable: %v", inner)
	}

	// summary and color are mandatory on the overwrite endpoint.
	if _, err := svc.UpdateCalendar(ctx, schedule.CalendarUpdate{CalID: "cal1"}); err == nil {
		t.Fatal("update without summary and color accepted")
	}

	list, err := svc.GetCalendars(ctx, []string{"cal1"})
	if err != nil || len(list) != 1 || rec.path != "/cgi-bin/oa/calendar/get" {
		t.Fatalf("get: %v %+v %s", err, list, rec.path)
	}
	ids, ok := rec.body["cal_id_list"].([]any)
	if !ok || len(ids) != 1 {
		t.Fatalf("cal_id_list not sent: %v", rec.body)
	}

	if err := svc.DeleteCalendar(ctx, "cal1"); err != nil || rec.path != "/cgi-bin/oa/calendar/del" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := svc.AddCalendar(ctx, schedule.Calendar{}, 0); err == nil {
		t.Fatal("calendar without summary accepted")
	}
	if _, err := svc.UpdateCalendar(ctx, schedule.CalendarUpdate{
		Summary: "s", Color: "#000000",
	}); err == nil {
		t.Fatal("update without cal_id accepted")
	}
	if _, err := svc.GetCalendars(ctx, nil); err == nil {
		t.Fatal("empty cal_id list accepted")
	}
	if err := svc.DeleteCalendar(ctx, " "); err == nil {
		t.Fatal("delete without cal_id accepted")
	}
}

func TestScheduleEntries(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"schedule_id":"s1",`+
		`"schedule_list":[{"schedule_id":"s1","summary":"评审","start_time":1,"end_time":2}]}`)
	svc := schedule.New(api)
	ctx := context.Background()

	id, err := svc.AddSchedule(ctx, schedule.Entry{
		Summary: "评审", StartTime: 1, EndTime: 2, CalID: "cal1",
		Attendees: []schedule.Attendee{{UserID: "u1"}},
		Reminders: &schedule.Reminders{IsRemind: 1, RemindBeforeEventSec: 3600},
	}, 1000002)
	if err != nil || id != "s1" || rec.path != "/cgi-bin/oa/schedule/add" {
		t.Fatalf("add: %q %v %s", id, err, rec.path)
	}
	inner := rec.body["schedule"].(map[string]any)
	if inner["summary"] != "评审" || inner["reminders"] == nil {
		t.Fatalf("schedule not sent: %v", inner)
	}
	if rec.body["agentid"] != float64(1000002) {
		t.Fatalf("agentid not sent: %v", rec.body)
	}

	rec.body = nil // the recorder merges into a reused map; start clean
	// The overwrite endpoint requires a time window on every update.
	splitID, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{
		ScheduleID: "s1", Summary: "改名", StartTime: 10, EndTime: 20,
		OpMode: schedule.OpModeSingle, OpStartTime: 100,
	})
	if err != nil || rec.path != "/cgi-bin/oa/schedule/update" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if splitID != "s1" {
		t.Fatalf("kept schedule id = %q", splitID)
	}
	if rec.body["op_mode"] != float64(1) || rec.body["op_start_time"] != float64(100) {
		t.Fatalf("op mode not sent: %v", rec.body)
	}
	inner = rec.body["schedule"].(map[string]any)
	if inner["summary"] != "改名" || inner["start_time"] != float64(10) || inner["end_time"] != float64(20) {
		t.Fatalf("overwrite body not sent: %v", inner)
	}
	if _, ok := inner["cal_id"]; ok {
		t.Fatalf("cal_id is not updatable: %v", inner)
	}

	// skip_attendees is the only way to keep the participants.
	if _, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 20, SkipAttendees: true,
	}); err != nil {
		t.Fatalf("skip attendees: %v", err)
	}
	if rec.body["skip_attendees"] != float64(1) {
		t.Fatalf("skip_attendees not sent: %v", rec.body)
	}

	// A split repeating schedule reports the new id, which callers must keep.
	splitID, err = svc.UpdateSchedule(ctx, schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 20,
		OpMode: schedule.OpModeThisAndLater, OpStartTime: 100,
	})
	if err != nil || splitID != "s1" {
		t.Fatalf("split update: %q %v", splitID, err)
	}

	if err := svc.AddAttendees(ctx, "s1", []schedule.Attendee{{UserID: "u2"}}); err != nil ||
		rec.path != "/cgi-bin/oa/schedule/add_attendees" {
		t.Fatalf("add attendees: %v %s", err, rec.path)
	}
	if err := svc.DeleteAttendees(ctx, "s1", []schedule.Attendee{{UserID: "u2"}}); err != nil ||
		rec.path != "/cgi-bin/oa/schedule/del_attendees" {
		t.Fatalf("delete attendees: %v %s", err, rec.path)
	}

	entries, err := svc.CalendarSchedules(ctx, "cal1", 0, 100)
	if err != nil || len(entries) != 1 || rec.path != "/cgi-bin/oa/schedule/get_by_calendar" {
		t.Fatalf("by calendar: %v %+v %s", err, entries, rec.path)
	}

	got, err := svc.GetSchedules(ctx, []string{"s1"})
	if err != nil || len(got) != 1 || rec.path != "/cgi-bin/oa/schedule/get" {
		t.Fatalf("get: %v %+v %s", err, got, rec.path)
	}

	if err := svc.DeleteSchedule(ctx, "s1"); err != nil || rec.path != "/cgi-bin/oa/schedule/del" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := svc.AddSchedule(ctx, schedule.Entry{Summary: "x"}, 0); err == nil {
		t.Fatal("schedule without window accepted")
	}
	if _, err := svc.AddSchedule(ctx, schedule.Entry{StartTime: 1, EndTime: 2}, 0); err == nil {
		t.Fatal("schedule without summary accepted")
	}
	if _, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{StartTime: 1, EndTime: 2}); err == nil {
		t.Fatal("update without id accepted")
	}
	if _, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{ScheduleID: "s1"}); err == nil {
		t.Fatal("update without a time window accepted")
	}
	if _, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 10,
	}); err == nil {
		t.Fatal("inverted time window accepted")
	}
	if _, err := svc.UpdateSchedule(ctx, schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 20, OpMode: schedule.OpModeSingle,
	}); err == nil {
		t.Fatal("partial repeat update without op_start_time accepted")
	}
	if err := svc.AddAttendees(ctx, "s1", nil); err == nil {
		t.Fatal("empty attendees accepted")
	}
	if _, err := svc.CalendarSchedules(ctx, "", 0, 0); err == nil {
		t.Fatal("missing cal_id accepted")
	}
	if _, err := svc.GetSchedules(ctx, nil); err == nil {
		t.Fatal("empty schedule ids accepted")
	}
}

func TestScheduleEvent(t *testing.T) {
	cases := map[string]string{
		schedule.EventDeleteCalendar:  `<xml><Event>delete_calendar</Event><CalId>cal1</CalId></xml>`,
		schedule.EventModifyCalendar:  `<xml><Event>modify_calendar</Event><CalId>cal1</CalId></xml>`,
		schedule.EventModifySchedule:  `<xml><Event>modify_schedule</Event><CalId>cal1</CalId><ScheduleId>s1</ScheduleId></xml>`,
		schedule.EventDeleteSchedule:  `<xml><Event>delete_schedule</Event><CalId>cal1</CalId><ScheduleId>s1</ScheduleId></xml>`,
		schedule.EventRespondSchedule: `<xml><Event>respond_schedule</Event><CalId>cal1</CalId><ScheduleId>s1</ScheduleId></xml>`,
	}
	for want, payload := range cases {
		event, err := schedule.ParseEvent([]byte(payload))
		if err != nil {
			t.Fatalf("%s: %v", want, err)
		}
		if event.Name != want || event.CalID != "cal1" {
			t.Fatalf("unexpected event: %+v", event)
		}
	}
	if _, err := schedule.ParseEvent([]byte(`<xml><Event>other</Event><CalId>c</CalId></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
	if _, err := schedule.ParseEvent([]byte(`<xml><Event>delete_calendar</Event></xml>`)); err == nil {
		t.Fatal("event without CalId accepted")
	}
	if !strings.Contains(cases[schedule.EventRespondSchedule], "respond_schedule") {
		t.Fatal("fixture mismatch")
	}
}

// ptrTo is the tri-state marker used by the update models: nil means "leave it
// alone", a pointer to the zero value means "set it to zero".
func ptrTo[T any](value T) *T { return &value }

// TestUpdateScheduleReturnsSplitID covers the response id: splitting a repeating
// schedule creates a new one, and callers must keep that id.
func TestUpdateScheduleReturnsSplitID(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"schedule_id":"s2"}`)
	svc := schedule.New(api)
	id, err := svc.UpdateSchedule(context.Background(), schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 20,
		OpMode: schedule.OpModeSingle, OpStartTime: 10,
	})
	if err != nil || id != "s2" || rec.path != "/cgi-bin/oa/schedule/update" {
		t.Fatalf("split id: %q %v %s", id, err, rec.path)
	}

	// A response without an id keeps the caller's id instead of emptying it.
	_, plain := newOffice(t, `{"errcode":0}`)
	id, err = schedule.New(plain).UpdateSchedule(context.Background(), schedule.EntryUpdate{
		ScheduleID: "s1", StartTime: 10, EndTime: 20,
	})
	if err != nil || id != "s1" {
		t.Fatalf("kept id: %q %v", id, err)
	}
}
