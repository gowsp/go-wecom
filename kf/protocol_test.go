package kf

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

// Values are literal protocol fixtures, not derived from implementation constants.
func TestDocumentedSessionStates(t *testing.T) {
	if StatePending != 0 || StateAssistant != 1 || StateQueue != 2 || StateHuman != 3 || StateEnded != 4 {
		t.Fatal("session states differ from official 94669 table")
	}
}

func TestStrictTextSendResultAndValidation(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			w.Write([]byte(`{"access_token":"t","expires_in":7200}`))
			return
		}
		calls++
		if r.URL.Path != "/cgi-bin/kf/send_msg" {
			t.Error("wrong send path")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["msgid"] != "request-1" || body["msgtype"] != "text" || body["text"].(map[string]any)["content"] != "你好" {
			t.Errorf("body=%v", body)
		}
		w.Write([]byte(`{"errcode":0,"msgid":"request-1"}`))
	}))
	defer srv.Close()
	c := New(client.New("corp", "secret", client.WithBaseURL(srv.URL)))
	request := SendTextRequest{OpenKfID: "account", ExternalUserID: "customer", Content: "你好", MsgID: "request-1"}
	out, err := c.SendTextWithResult(context.Background(), request)
	if err != nil || out.MsgID != "request-1" {
		t.Fatalf("out=%v err=%v", out, err)
	}
	request.Content = strings.Repeat("你", 683)
	if _, err = c.SendTextWithResult(context.Background(), request); !errors.Is(err, ErrTextTooLong) {
		t.Fatalf("err=%v", err)
	}
	request.Content = "ok"
	request.MsgID = "invalid/id"
	if _, err = c.SendTextWithResult(context.Background(), request); err == nil {
		t.Fatal("expected invalid ID")
	}
	if calls != 1 {
		t.Fatalf("invalid input reached server: %d", calls)
	}
}

func TestChangeStateQueueAndResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			w.Write([]byte(`{"access_token":"t","expires_in":7200}`))
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["service_state"] != float64(2) {
			t.Errorf("body=%v", body)
		}
		if _, ok := body["servicer_userid"]; ok {
			t.Error("queue must not include employee")
		}
		w.Write([]byte(`{"errcode":0,"msg_code":"event-code"}`))
	}))
	defer srv.Close()
	c := New(client.New("corp", "secret", client.WithBaseURL(srv.URL)))
	out, err := c.ChangeState(context.Background(), ChangeStateRequest{OpenKfID: "account", ExternalUserID: "customer", ServiceState: StateQueue})
	if err != nil || out.MsgCode != "event-code" {
		t.Fatalf("out=%v err=%v", out, err)
	}
}
