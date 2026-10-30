package aibot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const groupTextJSON = `{"msgid":"m1","aibotid":"AIBOTID","chatid":"CHATID","chattype":"group",` +
	`"from":{"userid":"USERID"},"response_url":"https://qyapi.weixin.qq.com/cgi-bin/aibot/response?response_code=X",` +
	`"msgtype":"text","text":{"content":"@RobotA hello"},"quote":{"msgtype":"text","text":{"content":"prev"}}}`

const enterChatJSON = `{"msgid":"m2","create_time":1700000000,"aibotid":"AIBOTID",` +
	`"from":{"corpid":"wpxxxx","userid":"USERID"},"msgtype":"event",` +
	`"event":{"eventtype":"enter_chat"}}`

func TestParseMessage(t *testing.T) {
	message, err := ParseMessage([]byte(groupTextJSON))
	if err != nil {
		t.Fatal(err)
	}
	if message.MsgType != TypeText || message.ChatType != ChatGroup {
		t.Fatalf("unexpected envelope: %+v", message)
	}
	if message.Text == nil || message.Text.Content != "@RobotA hello" {
		t.Fatalf("text not parsed: %+v", message.Text)
	}
	if message.Quote == nil || message.Quote.Text.Content != "prev" {
		t.Fatalf("quote not parsed: %+v", message.Quote)
	}
	if len(message.Raw) == 0 {
		t.Fatal("raw payload not preserved")
	}
}

func TestParseMessageRejects(t *testing.T) {
	cases := map[string]string{
		"malformed":    `{`,
		"no msgtype":   `{"msgid":"m"}`,
		"no msgid":     `{"msgtype":"text"}`,
		"bad chattype": `{"msgid":"m","msgtype":"text","chattype":"channel"}`,
		"event":        enterChatJSON,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseMessage([]byte(payload)); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}

func TestParseEvent(t *testing.T) {
	message, err := ParseEvent([]byte(enterChatJSON))
	if err != nil {
		t.Fatal(err)
	}
	if message.Event == nil || message.Event.EventType != EventEnterChat {
		t.Fatalf("event not parsed: %+v", message.Event)
	}
	if message.From.UserID != "USERID" || message.From.CorpID != "wpxxxx" {
		t.Fatalf("sender not parsed: %+v", message.From)
	}
}

func TestParseTemplateCardEventKeepsNestedPayload(t *testing.T) {
	payload := `{"msgid":"m3","aibotid":"A","chattype":"group","msgtype":"event",` +
		`"event":{"eventtype":"template_card_event","template_card_event":{"event_key":"k","task_id":"t"}}}`
	message, err := ParseEvent([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	var nested struct {
		EventKey string `json:"event_key"`
	}
	if err := json.Unmarshal(message.Event.Raw, &nested); err != nil {
		t.Fatalf("nested payload: %v", err)
	}
	if nested.EventKey != "k" {
		t.Fatalf("event_key = %q", nested.EventKey)
	}
	if _, err := ParseEvent([]byte(`{"msgtype":"event","event":{}}`)); err == nil {
		t.Fatal("event without eventtype accepted")
	}
}

func TestResponsePost(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode: %v", err)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	if err := New().Response(context.Background(), server.URL, MarkdownReply("hi")); err != nil {
		t.Fatal(err)
	}
	if got["msgtype"] != ReplyMarkdown {
		t.Fatalf("unexpected body: %v", got)
	}
}

func TestResponseErrors(t *testing.T) {
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer failing.Close()
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":40001,"errmsg":"bad"}`))
	}))
	defer errServer.Close()

	ctx := context.Background()
	cases := map[string]struct {
		url   string
		reply Reply
	}{
		"relative url": {"not-a-url", MarkdownReply("x")},
		"empty reply":  {errServer.URL, Reply{}},
		"no content":   {errServer.URL, Reply{MsgType: ReplyMarkdown, Markdown: &Markdown{}}},
		"empty card":   {errServer.URL, Reply{MsgType: ReplyTemplateCard}},
		"bad stream":   {errServer.URL, Reply{MsgType: ReplyStream, Stream: &Stream{}}},
		"server error": {failing.URL, MarkdownReply("x")},
		"errcode":      {errServer.URL, MarkdownReply("x")},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := New().Response(ctx, tc.url, tc.reply); err == nil {
				t.Fatal("error not reported")
			}
		})
	}
}

func TestReplyBuilders(t *testing.T) {
	if r := TextReply("hi"); r.MsgType != ReplyText || r.Text.Content != "hi" {
		t.Fatalf("text reply: %+v", r)
	}
	if r := StreamReply("chunk", false); !r.Stream.Finish == false {
		t.Fatalf("stream reply: %+v", r)
	}
	if r := TemplateCardReply(json.RawMessage(`{"card_type":"text_notice"}`)); r.MsgType != ReplyTemplateCard {
		t.Fatalf("card reply: %+v", r)
	}
	if err := MarkdownReply("x").validate(); err != nil {
		t.Fatal(err)
	}
}

// roundTripperFunc stands in for a failing transport.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

// TestResponseDoesNotLeakCredential guards the response_url: it carries a
// one-hour response_code, so a transport error must not echo it.
func TestResponseDoesNotLeakCredential(t *testing.T) {
	const credential = "https://qyapi.weixin.qq.com/cgi-bin/aibot/response?response_code=TOPSECRET"
	cases := map[string]error{
		"dial":     &url.Error{Op: "Post", URL: credential, Err: errors.New("dial tcp: connection refused")},
		"canceled": &url.Error{Op: "Post", URL: credential, Err: context.Canceled},
		"timeout":  &url.Error{Op: "Post", URL: credential, Err: context.DeadlineExceeded},
		"plain":    errors.New("transport exploded"),
	}
	for name, failure := range cases {
		t.Run(name, func(t *testing.T) {
			client := New(WithHTTPClient(&http.Client{Transport: roundTripperFunc(
				func(*http.Request) (*http.Response, error) { return nil, failure })}))
			err := client.Response(context.Background(), credential, MarkdownReply("x"))
			if err == nil {
				t.Fatal("transport error not reported")
			}
			if strings.Contains(err.Error(), "TOPSECRET") || strings.Contains(err.Error(), "response_code") {
				t.Fatalf("credential leaked: %v", err)
			}
			if !strings.Contains(err.Error(), "wecom: robot response") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}

	// Cancellation stays recognisable even though the URL is stripped.
	client := New(WithHTTPClient(&http.Client{Transport: roundTripperFunc(
		func(*http.Request) (*http.Response, error) {
			return nil, &url.Error{Op: "Post", URL: credential, Err: context.Canceled}
		})}))
	if err := client.Response(context.Background(), credential, MarkdownReply("x")); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

// TestResponseReportsStableTransportError keeps the failure diagnosable without
// the credential: the message is stable and cancellation is still detectable.
func TestResponseReportsStableTransportError(t *testing.T) {
	client := New(WithHTTPClient(&http.Client{Transport: roundTripperFunc(
		func(*http.Request) (*http.Response, error) {
			return nil, &url.Error{Op: "Post", URL: "https://example.com/?response_code=X",
				Err: errors.New("dial tcp 10.0.0.1:443: connect: connection refused")}
		})}))
	err := client.Response(context.Background(), "https://example.com/?response_code=X", MarkdownReply("x"))
	if err == nil || !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(err.Error(), "response_code") || strings.Contains(err.Error(), "10.0.0.1") {
		t.Fatalf("request detail leaked: %v", err)
	}
}
