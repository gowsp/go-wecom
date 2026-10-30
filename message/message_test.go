package message

import (
	"context"

	"encoding/json"
	"errors"
	"fmt"
	wxclient "github.com/gowsp/go-wecom/client"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

// messagePayload mirrors the JSON body the client sends. Tests decode into it
// instead of a generic map so the agent id stays an integer.
type messagePayload struct {
	ToUser  string `json:"touser"`
	MsgType string `json:"msgtype"`
	AgentID int    `json:"agentid"`
	Text    *struct {
		Content string `json:"content"`
	} `json:"text"`
	Markdown *struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

// fakeWeCom is a deterministic httptest replacement for qyapi.weixin.qq.com. It
// records every call and answers gettoken/message-send from scripted responses;
// it never touches the network.
type fakeWeCom struct {
	server *httptest.Server

	mu           sync.Mutex
	tokenCalls   int
	messageCalls int
	tokenQueries []url.Values
	sendTokens   []string
	sendBodies   []messagePayload

	tokens          []string
	expiresIn       int
	tokenErrcode    int
	messageErrcodes []int
	// tokenGate, when set, blocks the first gettoken response until closed.
	tokenGate <-chan struct{}
}

func newFakeWeCom(t *testing.T) *fakeWeCom {
	t.Helper()
	fake := &fakeWeCom{tokens: []string{"access-token-1"}, expiresIn: 7200}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeWeCom) handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/cgi-bin/gettoken":
		f.mu.Lock()
		f.tokenCalls++
		call := f.tokenCalls
		f.tokenQueries = append(f.tokenQueries, r.URL.Query())
		token := f.tokens[len(f.tokens)-1]
		if call-1 < len(f.tokens) {
			token = f.tokens[call-1]
		}
		errcode, expiresIn, gate := f.tokenErrcode, f.expiresIn, f.tokenGate
		f.mu.Unlock()

		if gate != nil {
			<-gate
		}
		if errcode != 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{"errcode": errcode, "errmsg": "failed"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0, "errmsg": "ok", "access_token": token, "expires_in": expiresIn,
		})
	case "/cgi-bin/message/send":
		var payload messagePayload
		_ = json.NewDecoder(r.Body).Decode(&payload)

		f.mu.Lock()
		f.messageCalls++
		call := f.messageCalls
		f.sendTokens = append(f.sendTokens, r.URL.Query().Get("access_token"))
		f.sendBodies = append(f.sendBodies, payload)
		errcode := 0
		if call-1 < len(f.messageErrcodes) {
			errcode = f.messageErrcodes[call-1]
		}
		f.mu.Unlock()

		_ = json.NewEncoder(w).Encode(map[string]any{"errcode": errcode, "errmsg": "ok"})
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeWeCom) counts() (tokenCalls, messageCalls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokenCalls, f.messageCalls
}

func (f *fakeWeCom) queries() []url.Values {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]url.Values(nil), f.tokenQueries...)
}

func (f *fakeWeCom) sentTokens() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.sendTokens...)
}

func (f *fakeWeCom) sentBodies() []messagePayload {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]messagePayload(nil), f.sendBodies...)
}

func TestSendTextCachesToken(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.tokens = []string{"access-token-1"}
	// A secret with reserved characters must round-trip through the query.
	client := New(wxclient.New("corp-id", "s3cr3t+/=&", wxclient.WithBaseURL(fake.server.URL)), 1000002)

	for i := 0; i < 3; i++ {
		if err := client.SendText(context.Background(), "zhangsan", "hello"); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	tokenCalls, messageCalls := fake.counts()
	if tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
	if messageCalls != 3 {
		t.Fatalf("message endpoint called %d times, want 3", messageCalls)
	}
	queries := fake.queries()
	if len(queries) != 1 {
		t.Fatalf("recorded %d token queries, want 1", len(queries))
	}
	if got := queries[0].Get("corpid"); got != "corp-id" {
		t.Fatalf("corpid=%q", got)
	}
	if got := queries[0].Get("corpsecret"); got != "s3cr3t+/=&" {
		t.Fatalf("corpsecret did not round-trip: %q", got)
	}

	tokens := fake.sentTokens()
	bodies := fake.sentBodies()
	if len(bodies) != 3 || len(tokens) != 3 {
		t.Fatalf("recorded %d bodies and %d tokens, want 3 each", len(bodies), len(tokens))
	}
	for i, body := range bodies {
		if tokens[i] != "access-token-1" {
			t.Fatalf("send %d used token %q", i, tokens[i])
		}
		if body.ToUser != "zhangsan" || body.MsgType != "text" || body.AgentID != 1000002 {
			t.Fatalf("send %d unexpected envelope: %+v", i, body)
		}
		if body.Text == nil || body.Text.Content != "hello" {
			t.Fatalf("send %d unexpected text content: %+v", i, body.Text)
		}
		if body.Markdown != nil {
			t.Fatalf("send %d unexpectedly carried markdown: %+v", i, body.Markdown)
		}
	}
}

func TestSendMarkdownBody(t *testing.T) {
	fake := newFakeWeCom(t)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)

	if err := client.SendMarkdown(context.Background(), "lisi", "**bold**"); err != nil {
		t.Fatal(err)
	}

	bodies := fake.sentBodies()
	if len(bodies) != 1 {
		t.Fatalf("recorded %d bodies, want 1", len(bodies))
	}
	body := bodies[0]
	if body.ToUser != "lisi" || body.MsgType != "markdown" || body.AgentID != 1000002 {
		t.Fatalf("unexpected envelope: %+v", body)
	}
	if body.Markdown == nil || body.Markdown.Content != "**bold**" {
		t.Fatalf("unexpected markdown content: %+v", body.Markdown)
	}
	if body.Text != nil {
		t.Fatalf("markdown send unexpectedly carried text: %+v", body.Text)
	}
	if got := fake.sentTokens()[0]; got != "access-token-1" {
		t.Fatalf("unexpected access_token query: %q", got)
	}
}

func TestConcurrentSendsFetchTokenOnce(t *testing.T) {
	fake := newFakeWeCom(t)
	gate := make(chan struct{})
	fake.tokenGate = gate
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	t.Cleanup(release)

	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)
	const workers = 16
	start := make(chan struct{})
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			errs <- client.SendText(context.Background(), fmt.Sprintf("user-%d", i), "hello")
		}(i)
	}
	close(start)

	// Wait until the single token request is in flight, then let it finish.
	deadline := time.Now().Add(5 * time.Second)
	for {
		tokenCalls, _ := fake.counts()
		if tokenCalls > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("token request did not start")
		}
		time.Sleep(time.Millisecond)
	}
	release()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent send failed: %v", err)
		}
	}

	tokenCalls, messageCalls := fake.counts()
	if tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
	if messageCalls != workers {
		t.Fatalf("message endpoint called %d times, want %d", messageCalls, workers)
	}
}

func TestTokenInvalidRetriesOnce(t *testing.T) {
	for _, code := range []int{40014, 42001} {
		t.Run(fmt.Sprintf("errcode %d", code), func(t *testing.T) {
			fake := newFakeWeCom(t)
			fake.tokens = []string{"token-1", "token-2"}
			fake.messageErrcodes = []int{code, 0}
			client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)

			if err := client.SendText(context.Background(), "zhangsan", "hello"); err != nil {
				t.Fatalf("send after token refresh: %v", err)
			}
			tokenCalls, messageCalls := fake.counts()
			if tokenCalls != 2 {
				t.Fatalf("gettoken called %d times, want 2", tokenCalls)
			}
			if messageCalls != 2 {
				t.Fatalf("message endpoint called %d times, want 2", messageCalls)
			}
			tokens := fake.sentTokens()
			if len(tokens) != 2 || tokens[0] != "token-1" || tokens[1] != "token-2" {
				t.Fatalf("retry did not use the refreshed token: %v", tokens)
			}
		})
	}
}

func TestTokenInvalidRetriesOnlyOnce(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.tokens = []string{"token-1", "token-2"}
	fake.messageErrcodes = []int{42001, 42001}
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)

	err := client.SendMarkdown(context.Background(), "zhangsan", "hello")
	if err == nil {
		t.Fatal("expected the retried send to fail")
	}
	if !strings.Contains(err.Error(), "42001") {
		t.Fatalf("error does not mention the errcode: %v", err)
	}
	tokenCalls, messageCalls := fake.counts()
	if tokenCalls != 2 {
		t.Fatalf("gettoken called %d times, want 2", tokenCalls)
	}
	if messageCalls != 2 {
		t.Fatalf("message endpoint called %d times, want 2", messageCalls)
	}
}

func TestMessageErrorIsStableAndHidesToken(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.messageErrcodes = []int{40013}
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)

	err := client.SendText(context.Background(), "zhangsan", "private-message-content")
	if err == nil {
		t.Fatal("expected a message error")
	}
	var apiErr *wxclient.APIError
	if !errors.As(err, &apiErr) || apiErr.Errcode != 40013 {
		t.Fatalf("expected APIError 40013, got %v", err)
	}
	if !strings.Contains(err.Error(), "40013") {
		t.Fatalf("error does not mention the errcode: %v", err)
	}
	for _, forbidden := range []string{"access-token-1", "private-message-content", "corp-secret"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaks %q: %v", forbidden, err)
		}
	}
}

func TestSendRejectsEmptyToUser(t *testing.T) {
	fake := newFakeWeCom(t)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002)

	if err := client.SendText(context.Background(), "   ", "hello"); err == nil {
		t.Fatal("expected an empty touser error")
	}
	tokenCalls, messageCalls := fake.counts()
	if tokenCalls != 0 || messageCalls != 0 {
		t.Fatalf("empty touser made %d token and %d message calls", tokenCalls, messageCalls)
	}
}
