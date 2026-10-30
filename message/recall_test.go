package message

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

func newFakeMessage(t *testing.T, reply func(w http.ResponseWriter, path string, body map[string]any)) (*Message, *[]string) {
	t.Helper()
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		paths = append(paths, r.URL.Path)
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		reply(w, r.URL.Path, body)
	}))
	t.Cleanup(server.Close)
	return New(client.New("corp", "secret", client.WithBaseURL(server.URL)), 1000002), &paths
}

func TestRecallAndUpdateTemplateCard(t *testing.T) {
	c, paths := newFakeMessage(t, func(w http.ResponseWriter, path string, body map[string]any) {
		switch path {
		case recallPath:
			if body["msgid"] != "m1" {
				t.Errorf("unexpected recall body: %+v", body)
			}
			_, _ = w.Write([]byte(`{"errcode":0}`))
		case updateTemplateCardPath:
			if body["response_code"] != "rc" {
				t.Errorf("unexpected card body: %+v", body)
			}
			if body["agentid"] != float64(1000002) {
				t.Errorf("agent id not defaulted: %+v", body)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"invaliduser":["u2"]}`))
		}
	})
	ctx := context.Background()
	if err := c.Recall(ctx, " m1 "); err != nil {
		t.Fatal(err)
	}
	if err := c.Recall(ctx, "  "); err == nil {
		t.Fatal("blank msgid must fail")
	}
	invalid, err := c.UpdateTemplateCard(ctx, TemplateCardUpdate{
		UserIDs:      []string{"u1", "u2"},
		ResponseCode: "rc",
		Button:       Button{ReplaceName: "已处理"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(invalid) != 1 || invalid[0] != "u2" {
		t.Fatalf("unexpected invalid users: %v", invalid)
	}
	if _, err := c.UpdateTemplateCard(ctx, TemplateCardUpdate{ResponseCode: "rc"}); err == nil {
		t.Fatal("missing recipients must fail")
	}
	if _, err := c.UpdateTemplateCard(ctx, TemplateCardUpdate{UserIDs: []string{"u"}}); err == nil {
		t.Fatal("missing response code must fail")
	}
	if len(*paths) != 2 {
		t.Fatalf("unexpected requests: %v", *paths)
	}
}
