package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/journal"
	"github.com/gowsp/go-wecom/meetingroom"
)

func TestJournal(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"next_cursor":2,"journaluuid_list":["u1"],`+
		`"info":{"journal_uuid":"u1","template_name":"日报","content":[{"text":"内容"}]},`+
		`"stat_list":[{"template_id":"t1","template_name":"日报"}]}`)
	svc := journal.New(api)
	ctx := context.Background()

	page, err := svc.Records(ctx, journal.RecordQuery{StartTime: 1, EndTime: 2, Limit: 10})
	if err != nil || rec.path != "/cgi-bin/oa/journal/get_record_list" ||
		len(page.UUIDs) != 1 || page.NextCursor != 2 {
		t.Fatalf("records: %v %s %+v", err, rec.path, page)
	}

	detail, err := svc.Detail(ctx, "u1")
	if err != nil || rec.path != "/cgi-bin/oa/journal/get_record_detail" ||
		detail.TemplateName != "日报" || len(detail.Content) == 0 {
		t.Fatalf("detail: %v %s %+v", err, rec.path, detail)
	}

	stats, err := svc.Stats(ctx, "t1", 1, 2)
	if err != nil || rec.path != "/cgi-bin/oa/journal/get_stat_list" || len(stats) != 1 {
		t.Fatalf("stats: %v %s %+v", err, rec.path, stats)
	}

	if _, err := svc.Records(ctx, journal.RecordQuery{}); err == nil {
		t.Fatal("invalid window accepted")
	}
	if _, err := svc.Detail(ctx, " "); err == nil {
		t.Fatal("empty uuid accepted")
	}
	if _, err := svc.Stats(ctx, "", 1, 2); err == nil {
		t.Fatal("missing template accepted")
	}
}

func TestMeetingRooms(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"meetingroom_id":1,"booking_id":"b1",`+
		`"meetingroom_list":[{"meetingroom_id":1,"name":"18F"}],`+
		`"booking_list":[{"meetingroom_id":1,"schedule":[{"booking_id":"b1","start_time":1,"end_time":2}]}],`+
		`"booking_info":{"booking_id":"b1","booker":"u1"}}`)
	svc := meetingroom.New(api)
	ctx := context.Background()

	id, err := svc.AddRoom(ctx, meetingroom.Room{Name: "18F", Capacity: 10})
	if err != nil || id != 1 || rec.path != "/cgi-bin/oa/meetingroom/add" {
		t.Fatalf("add room: %d %v %s", id, err, rec.path)
	}
	rooms, err := svc.ListRooms(ctx, meetingroom.RoomFilter{City: "深圳"})
	if err != nil || rec.path != "/cgi-bin/oa/meetingroom/list" || len(rooms) != 1 {
		t.Fatalf("list rooms: %v %s %+v", err, rec.path, rooms)
	}
	if err := svc.UpdateRoom(ctx, meetingroom.Room{ID: 1, Name: "19F"}); err != nil ||
		rec.path != "/cgi-bin/oa/meetingroom/edit" {
		t.Fatalf("update room: %v %s", err, rec.path)
	}
	if err := svc.DeleteRoom(ctx, 1); err != nil || rec.path != "/cgi-bin/oa/meetingroom/del" {
		t.Fatalf("delete room: %v %s", err, rec.path)
	}

	bookingID, err := svc.Book(ctx, meetingroom.BookRequest{
		RoomID: 1, StartTime: 1, EndTime: 2, Booker: "u1", Attendees: []string{"u2"},
	})
	if err != nil || bookingID != "b1" || rec.path != "/cgi-bin/oa/meetingroom/book" {
		t.Fatalf("book: %q %v %s", bookingID, err, rec.path)
	}
	bookings, err := svc.Bookings(ctx, 1, 1, 2)
	if err != nil || rec.path != "/cgi-bin/oa/meetingroom/get_booking_info" ||
		len(bookings) != 1 || bookings[0].BookingID != "b1" {
		t.Fatalf("bookings: %v %s %+v", err, rec.path, bookings)
	}
	info, err := svc.BookingInfo(ctx, "b1")
	if err != nil || rec.path != "/cgi-bin/oa/meetingroom/bookinfo/get" || info.Booker != "u1" {
		t.Fatalf("booking info: %v %s %+v", err, rec.path, info)
	}
	if err := svc.CancelBooking(ctx, "b1"); err != nil ||
		rec.path != "/cgi-bin/oa/meetingroom/cancel_book" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}

	if _, err := svc.AddRoom(ctx, meetingroom.Room{}); err == nil {
		t.Fatal("room without name accepted")
	}
	if _, err := svc.Book(ctx, meetingroom.BookRequest{}); err == nil {
		t.Fatal("booking without room accepted")
	}
	if _, err := svc.Bookings(ctx, 0, 1, 2); err == nil {
		t.Fatal("bookings without room accepted")
	}
	if err := svc.DeleteRoom(ctx, 0); err == nil {
		t.Fatal("delete without id accepted")
	}
}

func TestMeetingRoomEvent(t *testing.T) {
	book := `<xml><ToUserName>toUser</ToUserName><CreateTime>1</CreateTime>` +
		`<MsgType>event</MsgType><Event>book_meeting_room</Event>` +
		`<MeetingRoomId>1</MeetingRoomId><BookingId>b1</BookingId></xml>`
	event, err := meetingroom.ParseEvent([]byte(book))
	if err != nil || event.RoomID != 1 || event.BookingID != "b1" || event.Name != meetingroom.EventBook {
		t.Fatalf("book event: %v %+v", err, event)
	}
	cancel := `<xml><Event>cancel_meeting_room</Event><BookingId>b1</BookingId></xml>`
	if _, err := meetingroom.ParseEvent([]byte(cancel)); err != nil {
		t.Fatal(err)
	}
	if _, err := meetingroom.ParseEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
}

func TestJournalAttachment(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"download_url":"https://dl",`+
		`"cookie_name":"ck","cookie_value":"cv"}`)
	svc := journal.New(api)
	ctx := context.Background()

	attachment, err := svc.DownloadAttachment(ctx, "j1", "f1")
	if err != nil || attachment.URL != "https://dl" || attachment.CookieValue != "cv" ||
		rec.path != "/cgi-bin/oa/journal/download_wedrive_file" {
		t.Fatalf("attachment: %v %+v %s", err, attachment, rec.path)
	}
	if _, err := svc.DownloadAttachment(ctx, "", ""); err == nil {
		t.Fatal("attachment without ids accepted")
	}
}
