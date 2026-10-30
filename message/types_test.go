package message

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	wxclient "github.com/gowsp/go-wecom/client"
)

// sendRecorder answers gettoken and records the /cgi-bin/message/send body.
type sendRecorder struct {
	bodies []map[string]any
	status int
	body   string
}

func newSendRecorder(t *testing.T) (*sendRecorder, *Message) {
	t.Helper()
	rec := &sendRecorder{status: http.StatusOK, body: `{"errcode":0,"errmsg":"ok"}`}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cgi-bin/gettoken":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"errcode": 0, "access_token": "token", "expires_in": 7200,
			})
		case "/cgi-bin/message/send":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode body: %v", err)
			}
			rec.bodies = append(rec.bodies, body)
			w.WriteHeader(rec.status)
			_, _ = w.Write([]byte(rec.body))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	api := wxclient.New("corp", "secret", wxclient.WithBaseURL(server.URL))
	return rec, New(api, 1000002)
}

func TestSendMessageTypes(t *testing.T) {
	cases := []struct {
		name    string
		msg     AppMessage
		msgType string
		fields  []string
	}{
		{"image", AppMessage{Image: &Media{MediaID: "m1"}}, TypeImage, []string{"image", "media_id"}},
		{"voice", AppMessage{Voice: &Media{MediaID: "m2"}}, TypeVoice, []string{"voice", "media_id"}},
		{"file", AppMessage{File: &Media{MediaID: "m3"}}, TypeFile, []string{"file", "media_id"}},
		{"video", AppMessage{Video: &Video{MediaID: "m4", Title: "t"}}, TypeVideo, []string{"video", "title"}},
		{"textcard", AppMessage{TextCard: &TextCard{Title: "t", Description: "d", URL: "https://x", BtnTxt: "查看"}},
			TypeTextCard, []string{"textcard", "btntxt"}},
		{"news", AppMessage{News: &News{Articles: []NewsArticle{{Title: "t", URL: "https://x"}}}},
			TypeNews, []string{"news", "articles"}},
		{"mpnews", AppMessage{MPNews: &MPNews{Articles: []MPArticle{{Title: "t", ThumbMediaID: "m", Content: "c"}}}},
			TypeMPNews, []string{"mpnews", "thumb_media_id"}},
		{"miniprogram_notice", AppMessage{MiniProgramNotice: &MiniProgramNotice{
			AppID: "wx1", Title: "t", ContentItem: []MiniProgramContentItem{{Key: "k", Value: "v"}},
		}}, TypeMiniProgramNotice, []string{"miniprogram_notice", "content_item"}},
		{"template_card", AppMessage{TemplateCard: json.RawMessage(`{"card_type":"text_notice"}`)},
			TypeTemplateCard, []string{"template_card", "text_notice"}},
	}

	rec, api := newSendRecorder(t)
	ctx := context.Background()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec.bodies = nil
			if err := api.Send(ctx, "user1", tc.msg); err != nil {
				t.Fatalf("send: %v", err)
			}
			body := rec.bodies[0]
			if body["msgtype"] != tc.msgType {
				t.Errorf("msgtype = %v, want %s", body["msgtype"], tc.msgType)
			}
			if body["touser"] != "user1" || body["agentid"] != float64(1000002) {
				t.Errorf("bad envelope: %v", body)
			}
			raw, _ := json.Marshal(body)
			for _, field := range tc.fields {
				if !strings.Contains(string(raw), field) {
					t.Errorf("body lacks %q: %s", field, raw)
				}
			}
		})
	}
}

func TestSendMessageValidation(t *testing.T) {
	rec, api := newSendRecorder(t)
	ctx := context.Background()

	invalid := map[string]AppMessage{
		"empty body":        {},
		"empty content":     {Text: &Text{Content: "  "}},
		"missing media id":  {Image: &Media{}},
		"missing card url":  {TextCard: &TextCard{Title: "t"}},
		"too many articles": {News: &News{Articles: make([]NewsArticle, 9)}},
		"missing appid":     {MiniProgramNotice: &MiniProgramNotice{Title: "t"}},
		"type mismatch":     {Type: TypeImage, Text: &Text{Content: "x"}},
		"ambiguous":         {Text: &Text{Content: "x"}, Image: &Media{MediaID: "m"}},
	}
	for name, msg := range invalid {
		t.Run(name, func(t *testing.T) {
			if err := api.Send(ctx, "user1", msg); err == nil {
				t.Fatal("invalid message accepted")
			}
			if len(rec.bodies) != 0 {
				t.Fatalf("sent %d requests for an invalid message", len(rec.bodies))
			}
		})
	}

	if err := api.Send(ctx, "  ", AppMessage{Text: &Text{Content: "x"}}); err == nil {
		t.Fatal("empty touser accepted")
	}
	if err := New(nil, 0).Send(ctx, "user1", AppMessage{Text: &Text{Content: "x"}}); err == nil {
		t.Fatal("missing agent id accepted")
	}
}

func TestSendTypeInferenceAndFlags(t *testing.T) {
	rec, api := newSendRecorder(t)
	msg := AppMessage{
		Text:                   &Text{Content: "hello"},
		Safe:                   1,
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 1800,
	}
	if err := api.Send(context.Background(), "user1|user2", msg); err != nil {
		t.Fatal(err)
	}
	body := rec.bodies[0]
	if body["msgtype"] != TypeText || body["safe"] != float64(1) ||
		body["duplicate_check_interval"] != float64(1800) {
		t.Fatalf("unexpected body: %v", body)
	}
	if _, ok := body["enable_id_trans"]; ok {
		t.Fatal("unset flag must be omitted")
	}
	if err := api.Send(context.Background(), "user1", AppMessage{
		Type: TypeMarkdown, Markdown: &Text{Content: "# hi"},
	}); err != nil {
		t.Fatal(err)
	}
	if rec.bodies[1]["msgtype"] != TypeMarkdown {
		t.Fatalf("explicit type lost: %v", rec.bodies[1])
	}
}
