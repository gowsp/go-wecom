package appchat

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

func TestUpdateAndGetChat(t *testing.T) {
	var calls []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		bodies = append(bodies, body)
		if strings.HasSuffix(r.URL.Path, "/get") {
			_, _ = w.Write([]byte(`{"errcode":0,"chat_info":{"chatid":"c1","name":"群","owner":"o","userlist":["u1"],"chat_type":0}}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer server.Close()

	c := New(client.New("corp", "secret", client.WithBaseURL(server.URL)), 1)
	ctx := context.Background()
	if err := c.Update(ctx, UpdateRequest{ChatID: "c1", AddUserList: []string{"u2"}}); err != nil {
		t.Fatal(err)
	}
	if bodies[0]["chatid"] != "c1" || len(bodies[0]["add_user_list"].([]any)) != 1 {
		t.Fatalf("unexpected update body: %+v", bodies[0])
	}
	chat, err := c.Get(ctx, "c1")
	if err != nil || chat.Name != "群" || len(chat.UserList) != 1 {
		t.Fatalf("get chat: %+v %v", chat, err)
	}
	if !strings.Contains(calls[1], "chatid=c1") {
		t.Fatalf("get must use query: %s", calls[1])
	}
	if err := c.Update(ctx, UpdateRequest{ChatID: "c1"}); err == nil {
		t.Fatal("empty update must fail")
	}
	if err := c.Update(ctx, UpdateRequest{}); err == nil {
		t.Fatal("missing chat id must fail")
	}
}
