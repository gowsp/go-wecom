package mail

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

type recorded struct {
	Method string
	Path   string
	Query  string
	Body   map[string]any
}

func newFake(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, call recorded)) (*Mail, *[]recorded) {
	t.Helper()
	var calls []recorded
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"token","expires_in":7200}`))
			return
		}
		call := recorded{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery}
		if body, err := io.ReadAll(r.Body); err == nil && len(body) > 0 {
			_ = json.Unmarshal(body, &call.Body)
		}
		calls = append(calls, call)
		handler(w, r, call)
	}))
	t.Cleanup(server.Close)
	return New(client.New("corp", "secret", client.WithBaseURL(server.URL))), &calls
}

func TestSendMailBody(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
		_, _ = w.Write([]byte(`{"errcode":0}`))
	})
	ctx := context.Background()
	err := c.Send(ctx, SendRequest{
		To:             Address{Emails: []string{"a@example.com"}, UserIDs: []string{"zhangsan"}},
		Cc:             &Address{UserIDs: []string{"lisi"}},
		Subject:        "标题",
		Content:        "正文",
		AttachmentList: []Attachment{{FileName: "a.txt", Content: "BASE64"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	call := (*calls)[0]
	if call.Method != http.MethodPost || call.Path != composeSendPath {
		t.Fatalf("unexpected call: %+v", call)
	}
	if call.Body["subject"] != "标题" || call.Body["content"] != "正文" {
		t.Fatalf("unexpected body: %+v", call.Body)
	}
	to := call.Body["to"].(map[string]any)
	if to["emails"].([]any)[0] != "a@example.com" {
		t.Fatalf("unexpected recipients: %+v", to)
	}
	if _, ok := call.Body["enable_id_trans"]; ok {
		t.Fatalf("zero-valued optional field must be omitted: %+v", call.Body)
	}
}

func TestSendMailValidation(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
		_, _ = w.Write([]byte(`{"errcode":0}`))
	})
	ctx := context.Background()
	if err := c.Send(ctx, SendRequest{Subject: "s"}); err == nil || !strings.Contains(err.Error(), "to") {
		t.Fatalf("empty recipients must fail: %v", err)
	}
	if err := c.Send(ctx, SendRequest{To: Address{Emails: []string{"a@b.c"}}}); err == nil || !strings.Contains(err.Error(), "subject") {
		t.Fatalf("empty subject must fail: %v", err)
	}
}

func TestListMailReturnsCursor(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
		_, _ = w.Write([]byte(`{"errcode":0,"next_cursor":"NEXT","has_more":1,"mail_list":[{"mail_id":"m1"}]}`))
	})
	list, err := c.List(context.Background(), 100, 200, "CUR", 50)
	if err != nil {
		t.Fatal(err)
	}
	if list.NextCursor != "NEXT" || list.HasMore != 1 || len(list.Mails) != 1 || list.Mails[0].MailID != "m1" {
		t.Fatalf("unexpected list: %+v", list)
	}
	body := (*calls)[0].Body
	if body["begin_time"] != float64(100) || body["end_time"] != float64(200) || body["cursor"] != "CUR" {
		t.Fatalf("unexpected request: %+v", body)
	}
}

func TestGetGroupUsesQueryString(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
		_, _ = w.Write([]byte(`{"errcode":0,"groupid":"g@example.com","groupname":"g","email_list":{"list":["a@example.com"]}}`))
	})
	group, err := c.GetGroup(context.Background(), "g@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if group.GroupName != "g" || len(group.EmailList.List) != 1 {
		t.Fatalf("unexpected group: %+v", group)
	}
	call := (*calls)[0]
	if call.Method != http.MethodGet || call.Path != groupGetPath || !strings.Contains(call.Query, "groupid=g%40example.com") {
		t.Fatalf("unexpected call: %+v", call)
	}
}

func TestSearchPublicMailFuzzyFlag(t *testing.T) {
	for _, test := range []struct {
		fuzzy bool
		want  string
	}{
		{true, "fuzzy=1"},
		{false, "fuzzy=0"},
	} {
		c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
			_, _ = w.Write([]byte(`{"errcode":0,"list":[{"id":1,"email":"box@example.com"}]}`))
		})
		list, err := c.SearchPublicMail(context.Background(), test.fuzzy, "")
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 1 || list[0].Email != "box@example.com" {
			t.Fatalf("unexpected list: %+v", list)
		}
		if !strings.Contains((*calls)[0].Query, test.want) {
			t.Fatalf("query %q missing %q", (*calls)[0].Query, test.want)
		}
	}
}

func TestVIPResults(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request, call recorded) {
		if call.Path == vipListPath {
			_, _ = w.Write([]byte(`{"errcode":0,"has_more":false,"next_cursor":"N","userid_list":["a","b"]}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0,"succ_userid_list":["a"],"fail_userid_list":["b"]}`))
	})
	ctx := context.Background()
	result, err := c.AddVIP(ctx, []string{"a", "b"})
	if err != nil || len(result.Succeeded) != 1 || len(result.Failed) != 1 {
		t.Fatalf("unexpected vip result: %+v %v", result, err)
	}
	list, err := c.ListVIP(ctx, "", 0)
	if err != nil || len(list.UserIDs) != 2 || list.NextCursor != "N" {
		t.Fatalf("unexpected vip list: %+v %v", list, err)
	}
}

func TestAuthCodesAndValidation(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ recorded) {
		_, _ = w.Write([]byte(`{"errcode":0,"auth_code_list":[{"auth_code_id":7,"remark":"备注"}]}`))
	})
	codes, err := c.AuthCodes(context.Background(), 1)
	if err != nil || len(codes) != 1 || codes[0].ID != 7 || codes[0].Remark != "备注" {
		t.Fatalf("unexpected codes: %+v %v", codes, err)
	}
	if _, err := c.AuthCodes(context.Background(), 0); err == nil {
		t.Fatal("zero mailbox id must fail")
	}
	if err := c.DeleteAuthCode(context.Background(), 1, 0); err == nil {
		t.Fatal("zero auth code id must fail")
	}
}
