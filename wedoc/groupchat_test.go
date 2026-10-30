package wedoc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

func TestGroupChatAutomation(t *testing.T) {
	var paths []string
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		paths = append(paths, r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		bodies = append(bodies, body)
		switch r.URL.Path {
		case groupChatListPath:
			_, _ = w.Write([]byte(`{"errcode":0,"has_more":true,"next_cursor":"N","chat_id_list":["c1","c2"]}`))
		case groupChatGetPath:
			_, _ = w.Write([]byte(`{"errcode":0,"name":"群","owner":"o","user_list":["u1"]}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	}))
	defer server.Close()

	c := New(client.New("corp", "secret", client.WithBaseURL(server.URL)))
	ctx := context.Background()

	list, err := c.ListGroupChats(ctx, "doc1", "", 0)
	if err != nil || len(list.ChatIDs) != 2 || !list.HasMore {
		t.Fatalf("list: %+v %v", list, err)
	}
	if bodies[0]["docid"] != "doc1" || bodies[0]["limit"] != float64(1000) {
		t.Fatalf("unexpected list body: %+v", bodies[0])
	}

	chat, err := c.GetGroupChat(ctx, "doc1", "c1")
	if err != nil || chat.Name != "群" || chat.Owner != "o" {
		t.Fatalf("get: %+v %v", chat, err)
	}

	if err := c.UpdateGroupChat(ctx, GroupChatUpdate{DocID: "doc1", ChatID: "c1", Owner: "u2"}); err != nil {
		t.Fatal(err)
	}
	if bodies[2]["owner"] != "u2" {
		t.Fatalf("unexpected update body: %+v", bodies[2])
	}
	if err := c.UpdateGroupChat(ctx, GroupChatUpdate{DocID: "doc1", ChatID: "c1"}); err == nil {
		t.Fatal("empty update must fail")
	}
}
