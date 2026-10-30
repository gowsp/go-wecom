package wecom_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/dial"
	"github.com/gowsp/go-wecom/hr"
	"github.com/gowsp/go-wecom/pstn"
	"github.com/gowsp/go-wecom/todo"
)

// officeRecorder answers gettoken and records the last domain request.
type officeRecorder struct {
	method string
	path   string
	body   map[string]any
}

func newOffice(t *testing.T, response string) (*officeRecorder, *client.Client) {
	t.Helper()
	rec := &officeRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		rec.method, rec.path = r.Method, r.URL.Path
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&rec.body)
		}
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return rec, client.New("corp", "secret", client.WithBaseURL(server.URL))
}

func TestTodo(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"todo_id":"t1","content":"述职","status":1,`+
		`"attendees":[{"userid":"u1","status":1}]}`)
	svc := todo.New(api)
	ctx := context.Background()

	item, err := svc.Get(ctx, "t1")
	if err != nil || rec.path != "/cgi-bin/todo/get" || rec.body["todo_id"] != "t1" {
		t.Fatalf("get: %v %s %v", err, rec.path, rec.body)
	}
	if item.Content != "述职" || item.Status != 1 || item.Attendees[0].UserID != "u1" {
		t.Fatalf("unexpected item: %+v", item)
	}

	if err := svc.Update(ctx, "t1", 1, nil); err != nil || rec.path != "/cgi-bin/todo/update" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if _, ok := rec.body["attendees"]; ok {
		t.Fatalf("empty attendees must be omitted: %v", rec.body)
	}
	if err := svc.Update(ctx, "t1", 0, []todo.Attendee{{UserID: "u1", Status: 2}}); err != nil {
		t.Fatal(err)
	}
	if len(rec.body["attendees"].([]any)) != 1 {
		t.Fatalf("attendees not sent: %v", rec.body)
	}

	if _, err := svc.Get(ctx, " "); err == nil {
		t.Fatal("empty todo id accepted")
	}
	if err := svc.Update(ctx, "t1", 5, nil); err == nil {
		t.Fatal("invalid status accepted")
	}
}

func TestDial(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"record":[{"call_time":1,"total_duration":10,`+
		`"call_type":1,"caller":{"userid":"u1","duration":10},"callee":[{"phone":"138","duration":5}]}]}`)
	records, err := dial.New(api).Records(context.Background(), 100, 200, 0, 100)
	if err != nil || rec.path != "/cgi-bin/dial/get_dial_record" || len(records) != 1 {
		t.Fatalf("records: %v %s %+v", err, rec.path, records)
	}
	if records[0].Caller.UserID != "u1" || records[0].Callee[0].Phone != "138" {
		t.Fatalf("unexpected record: %+v", records[0])
	}
	if _, err := dial.New(api).Records(context.Background(), 0, 0, 0, 0); err == nil {
		t.Fatal("invalid window accepted")
	}
}

func TestPSTN(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"states":[{"code":0,"callid":"c1","userid":"u1"}],`+
		`"istalked":1,"calltime":10,"talktime":2}`)
	svc := pstn.New(api)
	ctx := context.Background()

	states, err := svc.Call(ctx, []string{"u1"})
	if err != nil || rec.path != "/cgi-bin/pstncc/call" || len(states) != 1 {
		t.Fatalf("call: %v %s %+v", err, rec.path, states)
	}
	if states[0].CallID != "c1" {
		t.Fatalf("unexpected state: %+v", states[0])
	}
	state, err := svc.AnswerState(ctx, "u1", "c1")
	if err != nil || rec.path != "/cgi-bin/pstncc/getstates" || state.IsTalked != 1 {
		t.Fatalf("state: %v %s %+v", err, rec.path, state)
	}
	if _, err := svc.Call(ctx, nil); err == nil {
		t.Fatal("empty callee accepted")
	}
	if _, err := svc.AnswerState(ctx, "u1", ""); err == nil {
		t.Fatal("missing callid accepted")
	}
}

func TestHR(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,`+
		`"group_list":[{"group_id":1,"group_name":"个人信息","field_list":[{"fieldid":11001,"field_name":"姓名"}]}],`+
		`"field_info":[{"fieldid":11001,"value_type":1,"value_string":"张三"}],`+
		`"update_results":[{"fieldid":11001,"result":1}]}`)
	svc := hr.New(api)
	ctx := context.Background()

	groups, err := svc.Fields(ctx)
	if err != nil || rec.path != "/cgi-bin/hr/get_fields" || len(groups) != 1 {
		t.Fatalf("fields: %v %s %+v", err, rec.path, groups)
	}
	if groups[0].Fields[0].FieldName != "姓名" {
		t.Fatalf("unexpected field: %+v", groups[0].Fields[0])
	}

	info, err := svc.StaffInfo(ctx, "u1", true, nil)
	if err != nil || rec.path != "/cgi-bin/hr/get_staff_info" || len(info) != 1 {
		t.Fatalf("staff info: %v %s %v", err, rec.path, info)
	}
	var decoded struct {
		ValueString string `json:"value_string"`
	}
	if err := json.Unmarshal(info[0], &decoded); err != nil || decoded.ValueString != "张三" {
		t.Fatalf("raw value: %v %+v", err, decoded)
	}

	result, err := svc.UpdateStaff(ctx, hr.StaffUpdate{
		UserID:      "u1",
		UpdateItems: []json.RawMessage{json.RawMessage(`{"fieldid":11001,"value_string":"李四"}`)},
	})
	if err != nil || rec.path != "/cgi-bin/hr/update_staff_info" || len(result.UpdateResults) != 1 {
		t.Fatalf("update: %v %s %+v", err, rec.path, result)
	}

	if _, err := svc.StaffInfo(ctx, "u1", false, nil); err == nil {
		t.Fatal("staff info without fields accepted")
	}
	if _, err := svc.UpdateStaff(ctx, hr.StaffUpdate{UserID: "u1"}); err == nil {
		t.Fatal("empty update accepted")
	}
	if _, err := svc.StaffInfo(ctx, " ", true, nil); err == nil {
		t.Fatal("empty userid accepted")
	}
}
