package appchat

import (
	"context"
	wxclient "github.com/gowsp/go-wecom/client"
	"strings"
	"testing"
)

func TestAppChatRequests(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		if call.Path == "/cgi-bin/appchat/create" {
			return map[string]any{"errcode": 0, "errmsg": "ok", "chatid": "chat-1"}
		}
		return map[string]any{"errcode": 0, "errmsg": "ok"}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)
	ctx := context.Background()

	chatID, err := client.CreateAppChat(ctx, " 官网商机 ", "owner", []string{" owner ", "member"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if chatID != "chat-1" {
		t.Fatalf("chat id = %q", chatID)
	}
	if err := client.SendAppChatText(ctx, " chat-1 ", "新商机：20261028-001"); err != nil {
		t.Fatal(err)
	}

	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	create := calls[0]
	if create.Path != "/cgi-bin/appchat/create" {
		t.Fatalf("create path = %q", create.Path)
	}
	if create.Body["name"] != "官网商机" || create.Body["owner"] != "owner" {
		t.Fatalf("unexpected create body: %+v", create.Body)
	}
	users := arrayField(t, create.Body["userlist"], "userlist")
	if len(users) != 2 || users[0] != "owner" || users[1] != "member" {
		t.Fatalf("userlist = %v", users)
	}
	if _, ok := create.Body["chatid"]; ok {
		t.Fatalf("auto-created chat must not send chatid: %+v", create.Body)
	}

	send := calls[1]
	if send.Path != "/cgi-bin/appchat/send" {
		t.Fatalf("send path = %q", send.Path)
	}
	if send.Body["chatid"] != "chat-1" || send.Body["msgtype"] != "text" {
		t.Fatalf("unexpected send body: %+v", send.Body)
	}
	text := objectField(t, send.Body["text"], "text")
	if text["content"] != "新商机：20261028-001" {
		t.Fatalf("unexpected content: %+v", text)
	}
}

func TestAppChatCreateValidation(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"empty name", func() error { _, err := client.CreateAppChat(ctx, " ", "owner", []string{"owner"}, ""); return err }, "name"},
		{"empty owner", func() error { _, err := client.CreateAppChat(ctx, "群", " ", []string{"owner"}, ""); return err }, "owner"},
		{"empty users", func() error { _, err := client.CreateAppChat(ctx, "群", "owner", nil, ""); return err }, "at least one user"},
		{"long chat id", func() error {
			_, err := client.CreateAppChat(ctx, "群", "owner", []string{"owner"}, strings.Repeat("x", 33))
			return err
		}, "32"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
	if tokenCalls, endpointCalls := fake.counts(); tokenCalls != 0 || endpointCalls != 0 {
		t.Fatalf("invalid arguments reached the API: %d token, %d endpoint calls", tokenCalls, endpointCalls)
	}
}

func TestAppChatSendValidation(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)
	ctx := context.Background()

	for _, test := range []struct{ name, chatID, content, want string }{
		{"empty chat", " ", "hello", "chatId"},
		{"empty content", "chat-1", " ", "content"},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := client.SendAppChatText(ctx, test.chatID, test.content)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
	if tokenCalls, endpointCalls := fake.counts(); tokenCalls != 0 || endpointCalls != 0 {
		t.Fatalf("invalid arguments reached the API: %d token, %d endpoint calls", tokenCalls, endpointCalls)
	}
}
