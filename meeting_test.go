package wecom_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gowsp/go-wecom/meeting"
)

func TestMeetingLifecycle(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"meetingid":"m1","excess_users":["u9"],`+
		`"meeting_info":{"meetingid":"m1","title":"周会","meeting_start":1,"meeting_duration":600,`+
		`"invitees":{"userid":["u1"]},"settings":{"password":"1234"}},`+
		`"meetingid_list":["m1"],"next_cursor":"c2","invitees":["u1"],`+
		`"meeting_list":[{"meetingid":"m1","title":"周会","meeting_start":1}],`+
		`"short_url":"https://short"}`)
	svc := meeting.New(api)
	ctx := context.Background()

	yes := true
	id, excess, err := svc.Create(ctx, meeting.CreateRequest{
		AdminUserID: "u1", Title: "周会", MeetingStart: 1, MeetingDuration: 600,
		Invitees: &meeting.UserList{UserID: []string{"u1"}},
		Settings: &meeting.Settings{Password: "1234", EnableWaitingRoom: &yes},
		Reminders: &meeting.Reminders{
			IsRepeat: 1, RepeatType: 0, RepeatUntil: 1606976813,
			RemindBefore: []int{0, 900},
		},
	})
	if err != nil || id != "m1" || len(excess) != 1 || rec.path != "/cgi-bin/meeting/create" {
		t.Fatalf("create: %q %v %v %s", id, excess, err, rec.path)
	}
	settings := rec.body["settings"].(map[string]any)
	if settings["enable_waiting_room"] != true {
		t.Fatalf("settings not sent: %v", settings)
	}
	// The repeat block is one nested object whose remind_before is a list of
	// offsets, not a list of per-member reminders.
	reminders := rec.body["reminders"].(map[string]any)
	if reminders["is_repeat"] != float64(1) || reminders["repeat_until"] != float64(1606976813) {
		t.Fatalf("reminders not sent: %v", reminders)
	}
	before, ok := reminders["remind_before"].([]any)
	if !ok || len(before) != 2 || before[1] != float64(900) {
		t.Fatalf("remind_before not sent: %v", reminders)
	}
	if _, ok := rec.body["is_repeat"]; ok {
		t.Fatalf("repeat must live inside reminders: %v", rec.body)
	}

	rec.body = nil // the recorder merges into a reused map; start clean
	excess, err = svc.Update(ctx, meeting.UpdateRequest{MeetingID: "m1", Title: ptrTo("改名")})
	if err != nil || rec.path != "/cgi-bin/meeting/update" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if len(excess) != 1 || excess[0] != "u9" {
		t.Fatalf("excess users = %v", excess)
	}
	// A partial update sends only what changed.
	if rec.body["title"] != "改名" || len(rec.body) != 2 {
		t.Fatalf("sparse update not sent: %v", rec.body)
	}

	// Clearing a description and emptying the invitees are expressible because
	// the update is a patch: nil leaves a value alone, an empty value clears it.
	if _, err := svc.Update(ctx, meeting.UpdateRequest{
		MeetingID: "m1", Description: ptrTo(""), Invitees: []string{},
		Reminders: &meeting.Reminders{}, Location: ptrTo(""),
	}); err != nil {
		t.Fatalf("clearing update: %v %s", err, rec.path)
	}
	if value, ok := rec.body["description"]; !ok || value != "" {
		t.Fatalf("empty description not sent: %v", rec.body)
	}
	if value, ok := rec.body["location"]; !ok || value != "" {
		t.Fatalf("empty location not sent: %v", rec.body)
	}
	inviteeBlock := rec.body["invitees"].(map[string]any)
	if list, ok := inviteeBlock["userid"].([]any); !ok || len(list) != 0 {
		t.Fatalf("empty invitees not sent: %v", inviteeBlock)
	}
	// reminders is an object with the repeat fields and a remind_before list,
	// not a list of per-member reminders.
	if block, ok := rec.body["reminders"].(map[string]any); !ok || len(block) != 0 {
		t.Fatalf("empty reminders not sent: %v", rec.body)
	}

	info, err := svc.GetInfo(ctx, "m1")
	if err != nil || info.Title != "周会" || rec.path != "/cgi-bin/meeting/get_info" {
		t.Fatalf("info: %v %+v %s", err, info, rec.path)
	}

	ids, cursor, err := svc.UserMeetingIDs(ctx, "u1", "", 20)
	if err != nil || len(ids) != 1 || cursor != "c2" ||
		rec.path != "/cgi-bin/meeting/get_user_meetingid" {
		t.Fatalf("user ids: %v %v %q %s", err, ids, cursor, rec.path)
	}

	invitees, nextCursor, err := svc.Invitees(ctx, "m1", "")
	if err != nil || len(invitees) != 1 || nextCursor != "c2" ||
		rec.path != "/cgi-bin/meeting/get_invitees" {
		t.Fatalf("invitees: %v %v %q %s", err, invitees, nextCursor, rec.path)
	}

	if err := svc.SetInvitees(ctx, "m1", []string{"u1", "u2"}); err != nil ||
		rec.path != "/cgi-bin/meeting/set_invitees" {
		t.Fatalf("set invitees: %v %s", err, rec.path)
	}

	records, _, err := svc.StartRecords(ctx, 1, 2, "", 10)
	if err != nil || len(records) != 1 ||
		rec.path != "/cgi-bin/meeting/statistics/get_start_list" {
		t.Fatalf("start records: %v %+v %s", err, records, rec.path)
	}

	url, err := svc.CreateCustomerShortURL(ctx, "m1", "data")
	if err != nil || url != "https://short" ||
		rec.path != "/cgi-bin/meeting/create_customer_short_url" {
		t.Fatalf("short url: %q %v %s", url, err, rec.path)
	}
	if _, err := svc.GetCustomerShortURL(ctx, "m1"); err != nil ||
		rec.path != "/cgi-bin/meeting/get_customer_short_url" {
		t.Fatalf("get short url: %v %s", err, rec.path)
	}

	if err := svc.Cancel(ctx, "m1"); err != nil || rec.path != "/cgi-bin/meeting/cancel" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}
}

func TestMeetingValidation(t *testing.T) {
	_, api := newOffice(t, `{"errcode":0}`)
	svc := meeting.New(api)
	ctx := context.Background()

	if _, _, err := svc.Create(ctx, meeting.CreateRequest{Title: "t", MeetingStart: 1, MeetingDuration: 1}); err == nil {
		t.Fatal("missing admin accepted")
	}
	if _, _, err := svc.Create(ctx, meeting.CreateRequest{AdminUserID: "u", Title: "t"}); err == nil {
		t.Fatal("missing schedule accepted")
	}
	if _, err := svc.Update(ctx, meeting.UpdateRequest{}); err == nil {
		t.Fatal("update without id accepted")
	}
	// meeting_start and meeting_duration must travel together.
	if _, err := svc.Update(ctx, meeting.UpdateRequest{MeetingID: "m1", MeetingStart: ptrTo[int64](1)}); err == nil {
		t.Fatal("update with meeting_start only accepted")
	}
	if _, err := svc.Update(ctx, meeting.UpdateRequest{MeetingID: "m1", MeetingDuration: ptrTo(600)}); err == nil {
		t.Fatal("update with meeting_duration only accepted")
	}
	if err := svc.Cancel(ctx, ""); err == nil {
		t.Fatal("cancel without id accepted")
	}
	if _, err := svc.GetInfo(ctx, " "); err == nil {
		t.Fatal("info without id accepted")
	}
	if _, _, err := svc.UserMeetingIDs(ctx, "", "", 0); err == nil {
		t.Fatal("ids without userid accepted")
	}
	if _, _, err := svc.Invitees(ctx, "", ""); err == nil {
		t.Fatal("invitees without id accepted")
	}
	if err := svc.SetInvitees(ctx, "m1", nil); err == nil {
		t.Fatal("set invitees without ids accepted")
	}
	if _, _, err := svc.StartRecords(ctx, 0, 0, "", 0); err == nil {
		t.Fatal("invalid window accepted")
	}
	if _, err := svc.CreateCustomerShortURL(ctx, "", ""); err == nil {
		t.Fatal("short url without input accepted")
	}
	if _, err := svc.GetCustomerShortURL(ctx, ""); err == nil {
		t.Fatal("get short url without id accepted")
	}
}

func TestMeetingAttendees(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2","status":"1",`+
		`"attendees":[{"userid":"u1","join_time":"1","audio_state":true}],`+
		`"user_list":[{"userid":"u2","instance_id":1}],`+
		`"result_list":[{"meetingid":"m1","instance_id":1}],`+
		`"meetingid":"m1","guests":[{"area":"86","phone_number":"138","guest_name":"g"}],`+
		`"quality":1,"audio_quality":2,"problems":["p"]}`)
	svc := meeting.New(api)
	ctx := context.Background()

	realtime, err := svc.RealtimeAttendees(ctx, "m1", "", "", 50)
	if err != nil || len(realtime.Attendees) != 1 || realtime.NextCursor != "c2" ||
		rec.path != "/cgi-bin/meeting/get_realtime_attendee_list" {
		t.Fatalf("realtime: %v %+v %s", err, realtime, rec.path)
	}
	if !realtime.Attendees[0].AudioState {
		t.Fatalf("attendee flags: %+v", realtime.Attendees[0])
	}

	history, err := svc.AttendeeList(ctx, "m1", "", 1, 2, "", 100)
	if err != nil || len(history.Attendees) != 1 || rec.path != "/cgi-bin/meeting/get_attendee_list" {
		t.Fatalf("attendees: %v %s", err, rec.path)
	}
	if rec.body["start_time"] != float64(1) || rec.body["end_time"] != float64(2) {
		t.Fatalf("window not sent: %v", rec.body)
	}

	waiting, err := svc.WaitingRoomMembers(ctx, "m1", "", 10)
	if err != nil || len(waiting.Members) != 1 ||
		rec.path != "/cgi-bin/meeting/waitingroom/get_current_user_list" {
		t.Fatalf("waiting: %v %+v %s", err, waiting, rec.path)
	}
	if _, err := svc.WaitingRoomRecords(ctx, "m1", "", 10); err != nil ||
		rec.path != "/cgi-bin/meeting/waitingroom/get_user_list" {
		t.Fatalf("waiting records: %v %s", err, rec.path)
	}

	devices, err := svc.DeviceInMeeting(ctx, "u1", []int{1}, []string{"m1"})
	if err != nil || len(devices) != 1 || rec.path != "/cgi-bin/meeting/check_device_in_meeting" {
		t.Fatalf("device: %v %+v %s", err, devices, rec.path)
	}

	guests, err := svc.Guests(ctx, "m1")
	if err != nil || len(guests.Guests) != 1 || rec.path != "/cgi-bin/meeting/get_guests" {
		t.Fatalf("guests: %v %+v %s", err, guests, rec.path)
	}
	if err := svc.SetGuests(ctx, "m1", []meeting.Guest{{Area: "86", PhoneNumber: "139"}}); err != nil ||
		rec.path != "/cgi-bin/meeting/set_guests" {
		t.Fatalf("set guests: %v %s", err, rec.path)
	}

	quality, err := svc.Quality(ctx, "m1", "", 1)
	if err != nil || quality.Quality != 1 || rec.path != "/cgi-bin/meeting/get_quality" {
		t.Fatalf("quality: %v %+v %s", err, quality, rec.path)
	}

	if _, err := svc.WaitingRoomMembers(ctx, "", "", 0); err == nil {
		t.Fatal("waiting without id accepted")
	}
	if _, err := svc.DeviceInMeeting(ctx, "", nil, nil); err == nil {
		t.Fatal("device without input accepted")
	}
	if err := svc.SetGuests(ctx, "", nil); err == nil {
		t.Fatal("set guests without id accepted")
	}
}

func TestMeetingEnrollment(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"question_count":12,"approve_type":1,`+
		`"question_list":[{"question_title":"q"}],`+
		`"enroll_id_list":[{"tmp_openid":"o1","enroll_id":"e1"}],`+
		`"has_more":false,"next_cursor":"c2","enroll_list":[{"enroll_id":"e1","status":1}],`+
		`"handled_count":3,"total_count":1}`)
	svc := meeting.New(api)
	ctx := context.Background()

	count, err := svc.SetEnrollConfig(ctx, meeting.EnrollConfig{
		MeetingID: "m1", ApproveType: 1,
		QuestionList: json.RawMessage(`[{"question_title":"q"}]`),
	})
	if err != nil || count != 12 || rec.path != "/cgi-bin/meeting/enroll/set_config" {
		t.Fatalf("set config: %d %v %s", count, err, rec.path)
	}

	config, err := svc.GetEnrollConfig(ctx, "m1")
	if err != nil || config.ApproveType != 1 || len(config.QuestionList) == 0 ||
		rec.path != "/cgi-bin/meeting/enroll/get_config" {
		t.Fatalf("get config: %v %+v %s", err, config, rec.path)
	}

	ids, err := svc.EnrollIDs(ctx, "m1", 1, []string{"o1"})
	if err != nil || len(ids) != 1 || ids[0].EnrollID != "e1" ||
		rec.path != "/cgi-bin/meeting/enroll/query_by_tmp_openid" {
		t.Fatalf("enroll ids: %v %+v %s", err, ids, rec.path)
	}

	page, err := svc.Enrollments(ctx, "m1", 1, "", 10)
	if err != nil || len(page.List) != 1 || rec.path != "/cgi-bin/meeting/enroll/list" {
		t.Fatalf("enrollments: %v %+v %s", err, page, rec.path)
	}

	handled, err := svc.ApproveEnrollments(ctx, "m1", 1, []string{"e1"})
	if err != nil || handled != 3 || rec.path != "/cgi-bin/meeting/enroll/approve" {
		t.Fatalf("approve: %d %v %s", handled, err, rec.path)
	}

	total, imported, err := svc.ImportEnrollments(ctx, "m1", []meeting.ImportEnrollment{{NickName: "n"}})
	if err != nil || total != 1 || len(imported) != 1 ||
		rec.path != "/cgi-bin/meeting/enroll/import" {
		t.Fatalf("import: %d %v %+v %s", total, err, imported, rec.path)
	}

	deleted, err := svc.DeleteEnrollments(ctx, "m1", []string{"e1"})
	if err != nil || deleted != 1 || rec.path != "/cgi-bin/meeting/enroll/delete" {
		t.Fatalf("delete: %d %v %s", deleted, err, rec.path)
	}

	if _, err := svc.SetEnrollConfig(ctx, meeting.EnrollConfig{}); err == nil {
		t.Fatal("config without id accepted")
	}
	if _, err := svc.EnrollIDs(ctx, "m1", 0, nil); err == nil {
		t.Fatal("ids without list accepted")
	}
	if _, err := svc.ApproveEnrollments(ctx, "m1", 1, nil); err == nil {
		t.Fatal("approve without ids accepted")
	}
	if _, _, err := svc.ImportEnrollments(ctx, "m1", nil); err == nil {
		t.Fatal("import without list accepted")
	}
	if _, err := svc.DeleteEnrollments(ctx, "m1", nil); err == nil {
		t.Fatal("delete without ids accepted")
	}
}

func TestMeetingControl(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0}`)
	svc := meeting.New(api)
	ctx := context.Background()
	user := meeting.OperatedUser{TmpOpenID: "o1", InstanceID: 1}

	locked := true
	if err := svc.SetMemberSetting(ctx, meeting.MemberSetting{MeetingID: "m1", MeetingLocked: &locked}); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/set" {
		t.Fatalf("setting: %v %s", err, rec.path)
	}
	if rec.body["meeting_locked"] != true {
		t.Fatalf("setting not sent: %v", rec.body)
	}
	if _, ok := rec.body["mute_all"]; ok {
		t.Fatalf("unset flag must be omitted: %v", rec.body)
	}

	if err := svc.SetCoHost(ctx, "m1", true, user); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/set_cohost" || rec.body["action"] != true {
		t.Fatalf("cohost: %v %s %v", err, rec.path, rec.body)
	}
	if err := svc.MuteUser(ctx, "m1", true, user); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/mute_user" || rec.body["option"] != true {
		t.Fatalf("mute: %v %s", err, rec.path)
	}
	if err := svc.SwitchUserVideo(ctx, "m1", false, user); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/switch_user_video" {
		t.Fatalf("video: %v %s", err, rec.path)
	}
	if err := svc.CloseScreenShare(ctx, "m1", user); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/close_screen_share" {
		t.Fatalf("share: %v %s", err, rec.path)
	}
	if err := svc.SetNicknames(ctx, "m1", []meeting.NicknameChange{{TmpOpenID: "o1", Nickname: "昵称"}}); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/set_nicknames" {
		t.Fatalf("nicknames: %v %s", err, rec.path)
	}
	if err := svc.ManageWaitingRoom(ctx, "m1", meeting.WaitingAdmit, true, []meeting.OperatedUser{user}); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/manage_waiting_room_users" ||
		rec.body["operate_type"] != float64(1) {
		t.Fatalf("waiting: %v %s %v", err, rec.path, rec.body)
	}
	if err := svc.KickoutUsers(ctx, "m1", false, []meeting.OperatedUser{user}); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/kickout_users" {
		t.Fatalf("kickout: %v %s", err, rec.path)
	}
	if err := svc.Dismiss(ctx, "m1", meeting.DismissOptions{ForceDismiss: true, RetrieveCode: true}); err != nil ||
		rec.path != "/cgi-bin/meeting/realcontrol/dismiss" || rec.body["force_dismiss"] != true {
		t.Fatalf("dismiss: %v %s %v", err, rec.path, rec.body)
	}

	if err := svc.SetMemberSetting(ctx, meeting.MemberSetting{}); err == nil {
		t.Fatal("setting without id accepted")
	}
	if err := svc.SetCoHost(ctx, "m1", true, meeting.OperatedUser{}); err == nil {
		t.Fatal("cohost without user accepted")
	}
	if err := svc.MuteUser(ctx, "", true, user); err == nil {
		t.Fatal("mute without meeting accepted")
	}
	if err := svc.SetNicknames(ctx, "m1", nil); err == nil {
		t.Fatal("nicknames without users accepted")
	}
	if err := svc.ManageWaitingRoom(ctx, "m1", 1, false, nil); err == nil {
		t.Fatal("waiting without users accepted")
	}
	if err := svc.KickoutUsers(ctx, "m1", false, nil); err == nil {
		t.Fatal("kickout without users accepted")
	}
	if err := svc.Dismiss(ctx, "", meeting.DismissOptions{}); err == nil {
		t.Fatal("dismiss without id accepted")
	}
}

func TestMeetingPolls(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"poll_theme_id":"theme1","poll_id":"poll1",`+
		`"polls_theme_info":[{"poll_theme_id":"theme1","polls_info":[{"poll_id":"poll1","status":1}]}],`+
		`"poll_topic":"主题","poll_question_data":[{"question_desc":"q","option_info":[{"option_desc":"a"}]}],`+
		`"vote_total_num":3}`)
	svc := meeting.New(api)
	ctx := context.Background()

	themeID, err := svc.CreatePollTheme(ctx, meeting.PollTheme{
		OperatorUserID: "u1", MeetingID: "m1", Topic: "主题",
		Questions: []meeting.PollQuestion{{QuestionType: 1, QuestionDesc: "q", Options: []string{"a"}}},
	})
	if err != nil || themeID != "theme1" || rec.path != "/cgi-bin/meeting/poll/create_theme" {
		t.Fatalf("create theme: %q %v %s", themeID, err, rec.path)
	}

	if err := svc.UpdatePollTheme(ctx, meeting.PollTheme{ThemeID: "theme1", MeetingID: "m1", Topic: "改"}); err != nil ||
		rec.path != "/cgi-bin/meeting/poll/update_theme" {
		t.Fatalf("update theme: %v %s", err, rec.path)
	}

	themes, err := svc.PollList(ctx, "u1", 1, "m1")
	if err != nil || len(themes) != 1 || len(themes[0].Polls) != 1 ||
		rec.path != "/cgi-bin/meeting/poll/get_poll_list" {
		t.Fatalf("poll list: %v %+v %s", err, themes, rec.path)
	}

	theme, err := svc.PollTheme(ctx, "u1", 1, "m1", "theme1")
	if err != nil || theme.Topic != "主题" || len(theme.Questions) != 1 ||
		rec.path != "/cgi-bin/meeting/poll/get_theme_info" {
		t.Fatalf("theme info: %v %+v %s", err, theme, rec.path)
	}

	detail, err := svc.Poll(ctx, "u1", 1, "m1", "poll1")
	if err != nil || detail.VoteTotal != 3 || len(detail.Questions) != 1 ||
		rec.path != "/cgi-bin/meeting/poll/get_poll_detail" {
		t.Fatalf("poll detail: %v %+v %s", err, detail, rec.path)
	}

	pollID, err := svc.StartPoll(ctx, "u1", 1, "m1", "theme1")
	if err != nil || pollID != "poll1" || rec.path != "/cgi-bin/meeting/poll/start" {
		t.Fatalf("start poll: %q %v %s", pollID, err, rec.path)
	}

	if err := svc.FinishPoll(ctx, "u1", 1, "m1", "theme1", "poll1"); err != nil ||
		rec.path != "/cgi-bin/meeting/poll/finish" {
		t.Fatalf("finish poll: %v %s", err, rec.path)
	}

	if err := svc.DeletePoll(ctx, "u1", 1, "m1", "theme1", "poll1"); err != nil ||
		rec.path != "/cgi-bin/meeting/poll/delete" {
		t.Fatalf("delete poll: %v %s", err, rec.path)
	}

	if _, err := svc.CreatePollTheme(ctx, meeting.PollTheme{MeetingID: "m1"}); err == nil {
		t.Fatal("theme without topic accepted")
	}
	if err := svc.UpdatePollTheme(ctx, meeting.PollTheme{}); err == nil {
		t.Fatal("update without theme id accepted")
	}
	if _, err := svc.PollList(ctx, "u1", 1, ""); err == nil {
		t.Fatal("list without meeting accepted")
	}
	if _, err := svc.PollTheme(ctx, "u1", 1, "m1", ""); err == nil {
		t.Fatal("theme info without id accepted")
	}
	if _, err := svc.Poll(ctx, "u1", 1, "m1", ""); err == nil {
		t.Fatal("poll detail without id accepted")
	}
	if _, err := svc.StartPoll(ctx, "u1", 1, "m1", ""); err == nil {
		t.Fatal("start poll without theme accepted")
	}
	if err := svc.FinishPoll(ctx, "u1", 1, "m1", "theme1", ""); err == nil {
		t.Fatal("finish poll without id accepted")
	}
	if err := svc.DeletePoll(ctx, "u1", 1, "m1", "", ""); err == nil {
		t.Fatal("delete poll without theme accepted")
	}
}

func TestWebinarLifecycle(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"title":"研讨会","meetingid":"w1","meeting_code":"123",`+
		`"audience_join_link":"https://aud","guest_join_link":"https://guest",`+
		`"guests":[{"guest_type":1,"userid":"u1"}],`+
		`"enroll_id_list":[{"tmp_openid":"o1","enroll_id":"e1"}],`+
		`"has_more":false,"next_cursor":"c2","enroll_list":[{"enroll_id":"e1"}],`+
		`"handled_count":2,"total_count":1,"question_count":3}`)
	svc := meeting.New(api)
	ctx := context.Background()

	webinar, err := svc.CreateWebinar(ctx, meeting.WebinarRequest{
		AdminUserID: "u1", Title: "研讨会", StartTime: "1", EndTime: "2",
	})
	if err != nil || webinar.MeetingID != "w1" || rec.path != "/cgi-bin/meeting/webinar/create" {
		t.Fatalf("create: %v %+v %s", err, webinar, rec.path)
	}

	if err := svc.UpdateWebinar(ctx, meeting.WebinarRequest{MeetingID: "w1", Title: "改"}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/update" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := svc.CancelWebinar(ctx, "w1"); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/cancel" {
		t.Fatalf("cancel: %v %s", err, rec.path)
	}
	got, err := svc.Webinar(ctx, "w1")
	if err != nil || got.AudienceJoinLink != "https://aud" ||
		rec.path != "/cgi-bin/meeting/webinar/get" {
		t.Fatalf("get: %v %+v %s", err, got, rec.path)
	}

	guests, err := svc.WebinarGuests(ctx, "w1", "")
	if err != nil || len(guests) != 1 || rec.path != "/cgi-bin/meeting/webinar/list_guest" {
		t.Fatalf("guests: %v %+v %s", err, guests, rec.path)
	}
	if err := svc.SetWebinarGuests(ctx, "w1", []meeting.WebinarGuest{{GuestType: 1, UserID: "u2"}}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/update_guest_list" {
		t.Fatalf("set guests: %v %s", err, rec.path)
	}

	if err := svc.SetWebinarWarmUp(ctx, meeting.WarmUp{MeetingID: "w1", Picture: "https://p"}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/update_warm_up" {
		t.Fatalf("warm up: %v %s", err, rec.path)
	}

	count, err := svc.SetWebinarEnrollConfig(ctx, meeting.WebinarEnrollConfig{MeetingID: "w1", ApproveType: 1})
	if err != nil || count != 3 || rec.path != "/cgi-bin/meeting/webinar/enroll/set_config" {
		t.Fatalf("set config: %d %v %s", count, err, rec.path)
	}
	if _, err := svc.GetWebinarEnrollConfig(ctx, "w1"); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/get_config" {
		t.Fatalf("get config: %v %s", err, rec.path)
	}
	if _, err := svc.WebinarEnrollIDs(ctx, "w1", 1, []string{"o1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/query_by_tmp_openid" {
		t.Fatalf("enroll ids: %v %s", err, rec.path)
	}
	if _, err := svc.WebinarEnrollments(ctx, "w1", 1, "", 10); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/list" {
		t.Fatalf("enrollments: %v %s", err, rec.path)
	}
	if handled, err := svc.ApproveWebinarEnrollments(ctx, "w1", 3, []string{"e1"}); err != nil || handled != 2 ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/approve" {
		t.Fatalf("approve: %d %v %s", handled, err, rec.path)
	}
	if _, _, err := svc.ImportWebinarEnrollments(ctx, "w1", []meeting.ImportEnrollment{{NickName: "n"}}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/import" {
		t.Fatalf("import: %v %s", err, rec.path)
	}
	if _, err := svc.DeleteWebinarEnrollments(ctx, "w1", []string{"e1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/webinar/enroll/delete" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := svc.CreateWebinar(ctx, meeting.WebinarRequest{Title: "t", StartTime: "1", EndTime: "2"}); err == nil {
		t.Fatal("webinar without admin accepted")
	}
	if _, err := svc.CreateWebinar(ctx, meeting.WebinarRequest{AdminUserID: "u", Title: "t"}); err == nil {
		t.Fatal("webinar without window accepted")
	}
	if err := svc.UpdateWebinar(ctx, meeting.WebinarRequest{}); err == nil {
		t.Fatal("update without id accepted")
	}
	if err := svc.CancelWebinar(ctx, ""); err == nil {
		t.Fatal("cancel without id accepted")
	}
	if _, err := svc.WebinarGuests(ctx, "", ""); err == nil {
		t.Fatal("guests without ids accepted")
	}
	if err := svc.SetWebinarWarmUp(ctx, meeting.WarmUp{}); err == nil {
		t.Fatal("warm up without id accepted")
	}
}

func TestWebinarEvent(t *testing.T) {
	rolePayload := `<xml><CreateTime>1</CreateTime><MsgType>event</MsgType>` +
		`<Event><![CDATA[meeting_change]]></Event>` +
		`<ChangeType><![CDATA[webinar_role_change]]></ChangeType>` +
		`<MeetingId><![CDATA[w1]]></MeetingId>` +
		`<OperatedUser><UserId><![CDATA[u1]]></UserId>` +
		`<TmpOpenId><![CDATA[o1]]></TmpOpenId><UserRole>2</UserRole></OperatedUser></xml>`
	event, err := meeting.ParseWebinarEvent([]byte(rolePayload))
	if err != nil || event.ChangeType != meeting.EventWebinarRoleChange ||
		event.OperatedUser == nil || event.OperatedUser.UserRole != 2 {
		t.Fatalf("role event: %v %+v", err, event)
	}

	warmPayload := `<xml><MsgType>event</MsgType><Event><![CDATA[meeting_change]]></Event>` +
		`<ChangeType><![CDATA[webinar_warm_up_upload]]></ChangeType>` +
		`<MeetingId><![CDATA[w1]]></MeetingId>` +
		`<WarmUpInfo><WarmUpPicture><![CDATA[https://p]]></WarmUpPicture>` +
		`<UploadStatus>1</UploadStatus><ErrorMsg><![CDATA[low]]></ErrorMsg></WarmUpInfo></xml>`
	event, err = meeting.ParseWebinarEvent([]byte(warmPayload))
	if err != nil || event.WarmUpInfo == nil || event.WarmUpInfo.UploadStatus != 1 {
		t.Fatalf("warm up event: %v %+v", err, event)
	}

	if _, err := meeting.ParseWebinarEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
	if _, err := meeting.ParseWebinarEvent([]byte(`<xml><Event>meeting_change</Event>` +
		`<ChangeType>other</ChangeType><MeetingId>w1</MeetingId></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}
}

func TestRoomsBooking(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"meeting_room_list":[{"meeting_room_id":"r1","meeting_room_name":"会议室1"}],`+
		`"meeting_settings":{"water_mark":2},"record_settings":{"share_record":0},`+
		`"meeting_info_list":[{"meetingid":"m1","subject":"周会"}],`+
		`"normal_count":10,"special_count":2,"normal_used_count":3,`+
		`"invite_id":"inv1","status":2,"response_time":"2022/11/22 14:35:26",`+
		`"basic_info":{"meeting_room_name":"会议室1"}}`)
	svc := meeting.New(api)
	ctx := context.Background()
	address := meeting.MRAAddress{Protocol: meeting.ProtocolSIP, DialString: "sip:1"}

	rooms, err := svc.BookRooms(ctx, "m1", []string{"r1"}, true)
	if err != nil || len(rooms) != 1 || rec.path != "/cgi-bin/meeting/rooms/book" {
		t.Fatalf("book: %v %+v %s", err, rooms, rec.path)
	}
	if err := svc.ReleaseRooms(ctx, "m1", []string{"r1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/rooms/release" {
		t.Fatalf("release: %v %s", err, rec.path)
	}

	page, err := svc.ListRooms(ctx, "", "", 50)
	if err != nil || len(page.Rooms) != 1 || page.NextCursor != "c2" ||
		rec.path != "/cgi-bin/meeting/rooms/list" {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}

	raw, err := svc.Room(ctx, "r1")
	if err != nil || len(raw) == 0 || rec.path != "/cgi-bin/meeting/rooms/get_info" {
		t.Fatalf("room: %v %s %s", err, raw, rec.path)
	}

	config, err := svc.RoomConfig(ctx, "r1")
	if err != nil || config.MeetingSettings.WaterMark != 2 ||
		rec.path != "/cgi-bin/meeting/rooms/get_config" {
		t.Fatalf("config: %v %+v %s", err, config, rec.path)
	}

	meetings, err := svc.RoomMeetings(ctx, "r1", "", 1, 2, "", 20)
	if err != nil || len(meetings.Meetings) != 1 ||
		rec.path != "/cgi-bin/meeting/rooms/list_meetings" {
		t.Fatalf("meetings: %v %+v %s", err, meetings, rec.path)
	}

	inventory, err := svc.RoomInventory(ctx)
	if err != nil || inventory.NormalCount != 10 ||
		rec.path != "/cgi-bin/meeting/rooms/get_inventory" {
		t.Fatalf("inventory: %v %+v %s", err, inventory, rec.path)
	}

	inviteID, err := svc.CallRoom(ctx, "m1", "r1", address)
	if err != nil || inviteID != "inv1" || rec.path != "/cgi-bin/meeting/rooms/call" {
		t.Fatalf("call: %q %v %s", inviteID, err, rec.path)
	}
	if err := svc.CancelRoomCall(ctx, "m1", inviteID, "r1", address); err != nil ||
		rec.path != "/cgi-bin/meeting/rooms/cancel_call" {
		t.Fatalf("cancel call: %v %s", err, rec.path)
	}
	response, err := svc.GetRoomResponseStatus(ctx, "m1", "r1", address)
	if err != nil || response.Status != 2 ||
		rec.path != "/cgi-bin/meeting/rooms/get_response_status" {
		t.Fatalf("response: %v %+v %s", err, response, rec.path)
	}

	if _, err := svc.BookRooms(ctx, "m1", nil, false); err == nil {
		t.Fatal("book without rooms accepted")
	}
	if err := svc.ReleaseRooms(ctx, "", nil); err == nil {
		t.Fatal("release without ids accepted")
	}
	if _, err := svc.Room(ctx, ""); err == nil {
		t.Fatal("room info without id accepted")
	}
	if _, err := svc.RoomMeetings(ctx, "", "", 1, 2, "", 0); err == nil {
		t.Fatal("room meetings without room accepted")
	}
	if _, err := svc.CallRoom(ctx, "m1", "", address); err == nil {
		t.Fatal("call without room accepted")
	}
}

func TestMeetingMRA(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"tmp_openid":"o1","instance_id":9,`+
		`"audio_state":true,"video_state":false,"default_layout":1}`)
	svc := meeting.New(api)
	ctx := context.Background()

	status, err := svc.MRAStatus(ctx, "m1", "o1")
	if err != nil || status.InstanceID != 9 || !status.AudioState ||
		rec.path != "/cgi-bin/meeting/mra/query_status" {
		t.Fatalf("status: %v %+v %s", err, status, rec.path)
	}
	if err := svc.SetMRALayout(ctx, "m1", 2, 1, "o1"); err != nil ||
		rec.path != "/cgi-bin/meeting/mra/set_default_layout" {
		t.Fatalf("layout: %v %s", err, rec.path)
	}
	mra, ok := rec.body["mra"].(map[string]any)
	if !ok || mra["tmp_openid"] != "o1" {
		t.Fatalf("mra not nested: %v", rec.body)
	}
	if err := svc.SetMRAHand(ctx, "m1", true, "o1"); err != nil ||
		rec.path != "/cgi-bin/meeting/mra/set_raise_hand" {
		t.Fatalf("hand: %v %s", err, rec.path)
	}
	if err := svc.HangupMRA(ctx, "m1", "o1"); err != nil ||
		rec.path != "/cgi-bin/meeting/mra/hangup" {
		t.Fatalf("hangup: %v %s", err, rec.path)
	}

	if _, err := svc.MRAStatus(ctx, "", ""); err == nil {
		t.Fatal("status without ids accepted")
	}
	if err := svc.SetMRALayout(ctx, "m1", 1, 0, ""); err == nil {
		t.Fatal("layout without openid accepted")
	}
	if err := svc.SetMRAHand(ctx, "", true, "o1"); err == nil {
		t.Fatal("hand without meeting accepted")
	}
	if err := svc.HangupMRA(ctx, "m1", ""); err == nil {
		t.Fatal("hangup without openid accepted")
	}
}

func TestRoomResponseEvent(t *testing.T) {
	payload := `<xml><MsgType>event</MsgType><Event><![CDATA[meeting_change]]></Event>` +
		`<ChangeType><![CDATA[meeting_room_response]]></ChangeType>` +
		`<MeetingId><![CDATA[m1]]></MeetingId><MeetingRoomId><![CDATA[r1]]></MeetingRoomId>` +
		`<MraAddress><Protocol>1</Protocol><DialString><![CDATA[sip:1]]></DialString></MraAddress>` +
		`<RoomResponseStatus>2</RoomResponseStatus></xml>`
	event, err := meeting.ParseRoomResponseEvent([]byte(payload))
	if err != nil || event.RoomID != "r1" || event.Status != 2 ||
		event.Address.Protocol != meeting.ProtocolSIP {
		t.Fatalf("event: %v %+v", err, event)
	}
	if _, err := meeting.ParseRoomResponseEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>other</ChangeType></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}
	if _, err := meeting.ParseRoomResponseEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>meeting_room_response</ChangeType>` +
			`<MeetingId>m1</MeetingId></xml>`)); err == nil {
		t.Fatal("event without room id accepted")
	}
}

func TestMeetingLayouts(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"selected_layout_id":"lay1","layout_name":"高级",`+
		`"layout_template_list":[{"layout_template_id":"1","thumbnail_url":"https://t"}],`+
		`"layout_list":[{"layout_id":"lay1","layout_name":"高级","page_list":[{"layout_template_id":"3"}]}]}`)
	svc := meeting.New(api)
	ctx := context.Background()
	pages := []meeting.LayoutPage{{TemplateID: "1", Seats: []meeting.Seat{{GridID: "10", TmpOpenID: "o1"}}}}

	templates, err := svc.LayoutTemplates(ctx)
	if err != nil || len(templates) != 1 ||
		rec.path != "/cgi-bin/meeting/layout/list_template" {
		t.Fatalf("templates: %v %+v %s", err, templates, rec.path)
	}

	selected, layouts, err := svc.AddLayout(ctx, "m1", pages, 1)
	if err != nil || selected != "lay1" || len(layouts) != 1 ||
		rec.path != "/cgi-bin/meeting/layout/add" {
		t.Fatalf("add layout: %q %v %+v %s", selected, err, layouts, rec.path)
	}

	if err := svc.UpdateLayout(ctx, "m1", "lay1", pages, true); err != nil ||
		rec.path != "/cgi-bin/meeting/layout/update" {
		t.Fatalf("update layout: %v %s", err, rec.path)
	}
	if rec.body["enable_set_default"] != true {
		t.Fatalf("set default flag not sent: %v", rec.body)
	}

	advanced, err := svc.AddAdvancedLayout(ctx, "m1", []meeting.Layout{{LayoutName: "高级", Pages: pages}})
	if err != nil || len(advanced) != 1 || rec.path != "/cgi-bin/meeting/advanced_layout/add" {
		t.Fatalf("add advanced: %v %+v %s", err, advanced, rec.path)
	}
	if err := svc.UpdateAdvancedLayout(ctx, "m1", meeting.Layout{LayoutID: "lay1", Pages: pages}); err != nil ||
		rec.path != "/cgi-bin/meeting/advanced_layout/update" {
		t.Fatalf("update advanced: %v %s", err, rec.path)
	}
	if err := svc.SetDefaultLayout(ctx, "m1", "lay1"); err != nil ||
		rec.path != "/cgi-bin/meeting/layout/set_default" {
		t.Fatalf("default layout: %v %s", err, rec.path)
	}
	if err := svc.ApplyAdvancedLayout(ctx, "m1", "lay1", []string{"o1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/advanced_layout/apply" {
		t.Fatalf("apply layout: %v %s", err, rec.path)
	}

	current, err := svc.ListLayouts(ctx, "m1")
	if err != nil || current.SelectedLayoutID != "lay1" || len(current.Layouts) != 1 ||
		rec.path != "/cgi-bin/meeting/advanced_layout/list" {
		t.Fatalf("list layouts: %v %+v %s", err, current, rec.path)
	}

	userLayout, err := svc.UserLayout(ctx, "m1", "o1", 1)
	if err != nil || userLayout.LayoutID != "lay1" ||
		rec.path != "/cgi-bin/meeting/advanced_layout/get_user_layout" {
		t.Fatalf("user layout: %v %+v %s", err, userLayout, rec.path)
	}

	if err := svc.BatchDeleteLayouts(ctx, "m1", []string{"lay1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/advanced_layout/batch_delete" {
		t.Fatalf("delete layouts: %v %s", err, rec.path)
	}

	if _, _, err := svc.AddLayout(ctx, "m1", nil, 0); err == nil {
		t.Fatal("layout without pages accepted")
	}
	if err := svc.UpdateLayout(ctx, "m1", "", nil, false); err == nil {
		t.Fatal("update layout without id accepted")
	}
	if err := svc.ApplyAdvancedLayout(ctx, "m1", "lay1", nil); err == nil {
		t.Fatal("apply without users accepted")
	}
	if _, err := svc.ListLayouts(ctx, ""); err == nil {
		t.Fatal("list without meeting accepted")
	}
	if _, err := svc.UserLayout(ctx, "m1", "", 0); err == nil {
		t.Fatal("user layout without openid accepted")
	}
	if err := svc.BatchDeleteLayouts(ctx, "m1", nil); err == nil {
		t.Fatal("delete layouts without ids accepted")
	}
}

func TestMeetingBackgrounds(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"selected_background_id":"bg1",`+
		`"background_list":[{"background_id":"bg1","image_md5":"m1"}]}`)
	svc := meeting.New(api)
	ctx := context.Background()

	result, err := svc.AddBackground(ctx, "m1",
		[]meeting.BackgroundImage{{ImageMD5: "m1", ImageURL: "https://i"}}, 1)
	if err != nil || result.SelectedID != "bg1" || len(result.List) != 1 ||
		rec.path != "/cgi-bin/meeting/layout/add_background" {
		t.Fatalf("add: %v %+v %s", err, result, rec.path)
	}
	if err := svc.SetDefaultBackground(ctx, "m1", "bg1"); err != nil ||
		rec.path != "/cgi-bin/meeting/layout/set_default_background" {
		t.Fatalf("default: %v %s", err, rec.path)
	}
	list, err := svc.Backgrounds(ctx, "m1")
	if err != nil || list.SelectedID != "bg1" || rec.path != "/cgi-bin/meeting/layout/list_background" {
		t.Fatalf("list: %v %+v %s", err, list, rec.path)
	}
	if err := svc.DeleteBackground(ctx, "m1", "bg1"); err != nil ||
		rec.path != "/cgi-bin/meeting/layout/delete_background" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}
	if err := svc.BatchDeleteBackgrounds(ctx, "m1", []string{"bg1", "bg2"}); err != nil ||
		rec.path != "/cgi-bin/meeting/layout/batch_delete_background" {
		t.Fatalf("batch delete: %v %s", err, rec.path)
	}

	if _, err := svc.AddBackground(ctx, "m1", nil, 0); err == nil {
		t.Fatal("add without images accepted")
	}
	if err := svc.SetDefaultBackground(ctx, "m1", ""); err == nil {
		t.Fatal("default without id accepted")
	}
	if _, err := svc.Backgrounds(ctx, ""); err == nil {
		t.Fatal("list without meeting accepted")
	}
	if err := svc.DeleteBackground(ctx, "", "bg1"); err == nil {
		t.Fatal("delete without meeting accepted")
	}
	if err := svc.BatchDeleteBackgrounds(ctx, "m1", nil); err == nil {
		t.Fatal("batch delete without ids accepted")
	}
}

func TestMeetingRecords(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"record_meetings":[{"meeting_record_id":"rec1","meetingid":"m1",`+
		`"record_files":[{"record_file_id":"f1","view_address":"https://v"}]}],`+
		`"summaries":[{"date":"2033-01-01","view_count":1}],`+
		`"audio_detect":1,"paragraphs":[{"pid":"p1","start_time":1,"end_time":2}],`+
		`"transcripts":{"paragraphs":[{"pid":"p1","sentences":[{"sid":"s1","text":"hi"}]}]},`+
		`"hits":[{"pid":"p1","sid":"s1","offset":1,"length":2}],`+
		`"timelines":[{"pid":"p1","sid":"s1","start_time":1}]}`)
	svc := meeting.New(api)
	ctx := context.Background()

	page, err := svc.RecordList(ctx, meeting.RecordQuery{MeetingID: "m1", Cursor: "", Limit: 10})
	if err != nil || len(page.Meetings) != 1 || page.NextCursor != "c2" ||
		rec.path != "/cgi-bin/meeting/record/list" {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}

	stats, err := svc.RecordStats(ctx, "rec1", "m1", 1, 2)
	if err != nil || len(stats) != 1 || rec.path != "/cgi-bin/meeting/record/get_statistics" {
		t.Fatalf("stats: %v %+v %s", err, stats, rec.path)
	}

	sharing := true
	if err := svc.SetRecordSharing(ctx, "rec1", "m1", meeting.SharingConfig{
		EnableSharing: &sharing, Password: "1234",
	}); err != nil || rec.path != "/cgi-bin/meeting/record/update_sharing_config" {
		t.Fatalf("sharing: %v %s", err, rec.path)
	}
	cfg, ok := rec.body["sharing_config"].(map[string]any)
	if !ok || cfg["enable_sharing"] != true || cfg["password"] != "1234" {
		t.Fatalf("sharing not nested: %v", rec.body)
	}
	if _, has := cfg["allow_download"]; has {
		t.Fatalf("unset sharing flag must be omitted: %v", cfg)
	}

	if err := svc.DeleteRecord(ctx, "rec1", "m1"); err != nil ||
		rec.path != "/cgi-bin/meeting/record/delete" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}
	if err := svc.DeleteRecordFile(ctx, "f1", "m1"); err != nil ||
		rec.path != "/cgi-bin/meeting/record/delete_file" {
		t.Fatalf("delete file: %v %s", err, rec.path)
	}

	file, err := svc.RecordFile(ctx, "f1", "m1")
	if err != nil || rec.path != "/cgi-bin/meeting/record/get_file" {
		t.Fatalf("file: %v %+v %s", err, file, rec.path)
	}

	if _, err := svc.RecordFiles(ctx, "rec1", ""); err != nil ||
		rec.path != "/cgi-bin/meeting/record/get_file_list" {
		t.Fatalf("files: %v %s", err, rec.path)
	}

	audio, paragraphs, err := svc.TranscriptParagraphs(ctx, "f1", "m1")
	if err != nil || !audio || len(paragraphs) != 1 ||
		rec.path != "/cgi-bin/meeting/record/transcript/get_paragraph_list" {
		t.Fatalf("paragraphs: %v %v %+v %s", err, audio, paragraphs, rec.path)
	}

	transcript, hasMore, err := svc.Transcript(ctx, "f1", "m1", "p1", 10)
	if err != nil || len(transcript) != 1 || !hasMore ||
		rec.path != "/cgi-bin/meeting/record/transcript/get_detail" {
		t.Fatalf("transcript: %v %v %+v %s", err, hasMore, transcript, rec.path)
	}

	hits, timelines, err := svc.SearchTranscript(ctx, "f1", "m1", "test")
	if err != nil || len(hits) != 1 || len(timelines) != 1 ||
		rec.path != "/cgi-bin/meeting/record/transcript/search" {
		t.Fatalf("search: %v %+v %+v %s", err, hits, timelines, rec.path)
	}

	if _, err := svc.RecordList(ctx, meeting.RecordQuery{}); err == nil {
		t.Fatal("list without selector accepted")
	}
	if _, err := svc.RecordStats(ctx, "", "", 0, 0); err == nil {
		t.Fatal("stats without ids accepted")
	}
	if err := svc.SetRecordSharing(ctx, "", "", meeting.SharingConfig{}); err == nil {
		t.Fatal("sharing without ids accepted")
	}
	if err := svc.DeleteRecordFile(ctx, "", ""); err == nil {
		t.Fatal("delete file without id accepted")
	}
	if _, err := svc.RecordFile(ctx, "", ""); err == nil {
		t.Fatal("file without id accepted")
	}
	if _, _, err := svc.SearchTranscript(ctx, "f1", "", ""); err == nil {
		t.Fatal("search without text accepted")
	}
}

func TestRecordingEvent(t *testing.T) {
	changes := []string{
		meeting.EventStartRecording, meeting.EventPauseRecording,
		meeting.EventResumeRecording, meeting.EventStopRecording,
		meeting.EventRecordingComplete, meeting.EventDeleteRecording,
	}
	for _, change := range changes {
		payload := `<xml><MsgType>event</MsgType><Event>meeting_change</Event>` +
			`<ChangeType>` + change + `</ChangeType>` +
			`<MeetingId>m1</MeetingId></xml>`
		event, err := meeting.ParseRecordingEvent([]byte(payload))
		if err != nil || event.ChangeType != change || event.MeetingID != "m1" {
			t.Fatalf("%s: %v %+v", change, err, event)
		}
	}
	if _, err := meeting.ParseRecordingEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>other</ChangeType>` +
			`<MeetingId>m1</MeetingId></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}
	if _, err := meeting.ParseRecordingEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>start_recording</ChangeType></xml>`)); err == nil {
		t.Fatal("event without meeting id accepted")
	}
}

func TestMeetingPhoneAndDevices(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"phone_numbers":[{"area":86,"phone":"133","status":"CALLING","tmp_openid":"o1"}],`+
		`"invalid_phone_numbers":[{"area":86,"phone":"000"}],`+
		`"tmp_openid_list":[{"area":86,"phone":"133","tmp_openid":"o1"}],`+
		`"device_info_list":[{"meeting_room_id":"r1","device_model":"TT"}],`+
		`"controller_info_list":[{"rooms_id":"r1","controller_name":"ctl"}]}`)
	svc := meeting.New(api)
	ctx := context.Background()

	out, err := svc.Callout(ctx, "m1", []meeting.PhoneNumber{{Area: 86, Phone: "133"}})
	if err != nil || len(out.Numbers) != 1 || len(out.Invalid) != 1 ||
		rec.path != "/cgi-bin/meeting/phone/callout" {
		t.Fatalf("callout: %v %+v %s", err, out, rec.path)
	}

	status, err := svc.CalloutStatus(ctx, "m1", "", 10)
	if err != nil || len(status.Numbers) != 1 ||
		rec.path != "/cgi-bin/meeting/phone/get_callout_status" {
		t.Fatalf("status: %v %+v %s", err, status, rec.path)
	}

	ids, err := svc.PhoneMemberIDs(ctx, "m1", []meeting.PhoneNumber{{Area: 86, Phone: "133"}})
	if err != nil || len(ids) != 1 || rec.path != "/cgi-bin/meeting/phone/get_tmp_openid" {
		t.Fatalf("ids: %v %+v %s", err, ids, rec.path)
	}

	devices, err := svc.Devices(ctx, "", "", 50)
	if err != nil || len(devices.Devices) != 1 ||
		rec.path != "/cgi-bin/meeting/rooms/list_devices" {
		t.Fatalf("devices: %v %+v %s", err, devices, rec.path)
	}
	controllers, err := svc.Controllers(ctx, "", "", 20)
	if err != nil || len(controllers.Devices) != 1 ||
		rec.path != "/cgi-bin/meeting/rooms/list_controllers" {
		t.Fatalf("controllers: %v %+v %s", err, controllers, rec.path)
	}

	if _, err := svc.Callout(ctx, "", nil); err == nil {
		t.Fatal("callout without input accepted")
	}
	if _, err := svc.CalloutStatus(ctx, "", "", 0); err == nil {
		t.Fatal("status without meeting accepted")
	}
	if _, err := svc.PhoneMemberIDs(ctx, "m1", nil); err == nil {
		t.Fatal("ids without numbers accepted")
	}
}

func TestMeetingVIP(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"jobid":"job1","invalid_userid_list":["u9"],`+
		`"status":1,"userid_list":["u1"],"has_more":true,"next_cursor":"c2"}`)
	svc := meeting.New(api)
	ctx := context.Background()

	job, err := svc.SubmitAddVIP(ctx, []string{"u1"})
	if err != nil || job.JobID != "job1" || len(job.Invalid) != 1 ||
		rec.path != "/cgi-bin/meeting/vip/submit_batch_add_job" {
		t.Fatalf("add vip: %v %+v %s", err, job, rec.path)
	}
	result, err := svc.AddVIPResult(ctx, "job1")
	if err != nil || result.Status != 1 || rec.path != "/cgi-bin/meeting/vip/batch_add_job_result" {
		t.Fatalf("add result: %v %+v %s", err, result, rec.path)
	}

	if _, err := svc.SubmitRemoveVIP(ctx, []string{"u1"}); err != nil ||
		rec.path != "/cgi-bin/meeting/vip/submit_batch_del_job" {
		t.Fatalf("remove vip: %v %s", err, rec.path)
	}
	if _, err := svc.RemoveVIPResult(ctx, "job1"); err != nil ||
		rec.path != "/cgi-bin/meeting/vip/batch_del_job_result" {
		t.Fatalf("remove result: %v %s", err, rec.path)
	}

	page, err := svc.ListVIP(ctx, "", 2)
	if err != nil || len(page.UserIDs) != 1 || rec.path != "/cgi-bin/meeting/vip/list" {
		t.Fatalf("list vip: %v %+v %s", err, page, rec.path)
	}

	if _, err := svc.SubmitAddVIP(ctx, nil); err == nil {
		t.Fatal("empty add accepted")
	}
	if _, err := svc.SubmitRemoveVIP(ctx, nil); err == nil {
		t.Fatal("empty remove accepted")
	}
	if _, err := svc.AddVIPResult(ctx, ""); err == nil {
		t.Fatal("result without job accepted")
	}
}

func TestMeetingChangeEvent(t *testing.T) {
	changes := []string{
		meeting.ChangeModifyMeeting, meeting.ChangeCancelMeeting,
		meeting.ChangeMeetingStart, meeting.ChangeMeetingEnd,
		meeting.ChangeMeetingMuteAll, meeting.ChangeMeetingUnmuteAll,
		meeting.ChangeJoinMeeting, meeting.ChangeQuitMeeting,
		meeting.ChangeJoinBeforeHost, meeting.ChangeJoinWaitingRoom,
		meeting.ChangeQuitWaitingRoom, meeting.ChangeJoinFromMeetingRoom,
		meeting.ChangeMoveToWaitingRoom, meeting.ChangeOpenScreenShare,
		meeting.ChangeCloseScreenShare, meeting.ChangeRoleChange,
		meeting.ChangePSTNStatusUpdate, meeting.ChangeMediumUpload,
		meeting.ChangeEnroll, meeting.ChangeCancelEnroll,
	}
	for _, change := range changes {
		payload := `<xml><MsgType>event</MsgType><Event>meeting_change</Event>` +
			`<ChangeType>` + change + `</ChangeType><MeetingId>m1</MeetingId></xml>`
		event, err := meeting.ParseMeetingChangeEvent([]byte(payload))
		if err != nil || event.ChangeType != change || event.MeetingID != "m1" {
			t.Fatalf("%s: %v %+v", change, err, event)
		}
	}

	rolePayload := `<xml><Event>meeting_change</Event><ChangeType>role_change</ChangeType>` +
		`<MeetingId>m1</MeetingId><OperatedUser><UserId>u1</UserId>` +
		`<TmpOpenId>o1</TmpOpenId><UserRole>2</UserRole></OperatedUser></xml>`
	role, err := meeting.ParseMeetingChangeEvent([]byte(rolePayload))
	if err != nil || role.OperatedUser == nil || role.OperatedUser.UserRole != 2 {
		t.Fatalf("role change: %v %+v", err, role.OperatedUser)
	}

	pstnPayload := `<xml><Event>meeting_change</Event><ChangeType>pstn_status_update</ChangeType>` +
		`<MeetingId>m1</MeetingId><PstnStatus>START_ACCEPT</PstnStatus></xml>`
	pstn, err := meeting.ParseMeetingChangeEvent([]byte(pstnPayload))
	if err != nil || pstn.PSTNStatus != "START_ACCEPT" {
		t.Fatalf("pstn: %v %+v", err, pstn)
	}

	uploadPayload := `<xml><Event>meeting_change</Event><ChangeType>medium_upload</ChangeType>` +
		`<MeetingId>m1</MeetingId><AllUploadStatus>false</AllUploadStatus>` +
		`<UploadInfo><MediumUrl>https://i/1.png</MediumUrl><MediumType>2</MediumType>` +
		`<UploadStatus>1</UploadStatus></UploadInfo></xml>`
	upload, err := meeting.ParseMeetingChangeEvent([]byte(uploadPayload))
	if err != nil || len(upload.Uploads) != 1 || upload.Uploads[0].MediumType != 2 {
		t.Fatalf("upload: %v %+v", err, upload.Uploads)
	}

	if _, err := meeting.ParseMeetingChangeEvent([]byte(
		`<xml><Event>other</Event><ChangeType>join_meeting</ChangeType><MeetingId>m1</MeetingId></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
	if _, err := meeting.ParseMeetingChangeEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>nope</ChangeType>` +
			`<MeetingId>m1</MeetingId></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}
	if _, err := meeting.ParseMeetingChangeEvent([]byte(
		`<xml><Event>meeting_change</Event><ChangeType>join_meeting</ChangeType></xml>`)); err == nil {
		t.Fatal("missing meeting id accepted")
	}

	stats, err := meeting.ParseMeetingStatistics([]byte(
		`<xml><Event>meeting_statistics</Event><ChangeType>start_meeting</ChangeType>` +
			`<MeetingId>m1</MeetingId><Status>1</Status></xml>`))
	if err != nil || stats.Status != 1 || stats.MeetingID != "m1" {
		t.Fatalf("statistics: %v %+v", err, stats)
	}
	if _, err := meeting.ParseMeetingStatistics([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected statistics event accepted")
	}
}
