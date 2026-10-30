package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// tokenServer scripts token responses without importing any business package.
type tokenServer struct {
	server       *httptest.Server
	tokens       []string
	expiresIn    int
	tokenErrcode int
	calls        int
	mu           sync.Mutex
}

func newFakeWeCom(t *testing.T) *tokenServer {
	t.Helper()
	f := &tokenServer{tokens: []string{"access-token-1"}, expiresIn: 7200}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cgi-bin/gettoken" {
			t.Errorf("unexpected token client endpoint: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		f.mu.Lock()
		defer f.mu.Unlock()
		f.calls++
		index := f.calls - 1
		if index >= len(f.tokens) {
			index = len(f.tokens) - 1
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": f.tokenErrcode, "access_token": f.tokens[index], "expires_in": f.expiresIn,
		})
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *tokenServer) counts() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls, 0
}

func TestTokenRefreshesFiveMinutesBeforeExpiry(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.tokens = []string{"token-1", "token-2"}
	fake.expiresIn = 600
	client := New("corp-id", "corp-secret", WithBaseURL(fake.server.URL))
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	client.now = func() time.Time { return now }

	token, err := client.Token(context.Background())
	if err != nil || token != "token-1" {
		t.Fatalf("first token=%q err=%v", token, err)
	}
	// 10 minute lifetime with a 5 minute early-refresh margin: cached for the
	// first five minutes, refreshed afterwards.
	now = now.Add(4*time.Minute + 59*time.Second)
	if token, err = client.Token(context.Background()); err != nil || token != "token-1" {
		t.Fatalf("cached token=%q err=%v", token, err)
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("token refreshed too early: %d calls", tokenCalls)
	}
	now = now.Add(2 * time.Second)
	if token, err = client.Token(context.Background()); err != nil || token != "token-2" {
		t.Fatalf("refreshed token=%q err=%v", token, err)
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 2 {
		t.Fatalf("token not refreshed: %d calls", tokenCalls)
	}
}

func TestShortLivedTokenCachesHalfLifetime(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.tokens = []string{"short-1", "short-2"}
	fake.expiresIn = 60
	client := New("corp-id", "corp-secret", WithBaseURL(fake.server.URL))
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	client.now = func() time.Time { return now }

	if token, err := client.Token(context.Background()); err != nil || token != "short-1" {
		t.Fatalf("first token=%q err=%v", token, err)
	}
	now = now.Add(29 * time.Second)
	if token, err := client.Token(context.Background()); err != nil || token != "short-1" {
		t.Fatalf("cached token=%q err=%v", token, err)
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("short-lived token refreshed too early: %d calls", tokenCalls)
	}
	now = now.Add(2 * time.Second)
	if token, err := client.Token(context.Background()); err != nil || token != "short-2" {
		t.Fatalf("refreshed token=%q err=%v", token, err)
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 2 {
		t.Fatalf("short-lived token not refreshed: %d calls", tokenCalls)
	}
}

func TestTokenErrorIsStableAndHidesSecret(t *testing.T) {
	fake := newFakeWeCom(t)
	fake.tokenErrcode = 40001
	client := New("corp-id", "corp-secret", WithBaseURL(fake.server.URL))

	_, err := client.Token(context.Background())
	if err == nil {
		t.Fatal("expected a token error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Errcode != 40001 {
		t.Fatalf("expected APIError 40001, got %v", err)
	}
	for _, forbidden := range []string{"corp-secret", "corp-id"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaks %q: %v", forbidden, err)
		}
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
}

func TestTransportErrorHidesCredentials(t *testing.T) {
	fake := newFakeWeCom(t)
	client := New("corp-id", "corp-secret", WithBaseURL(fake.server.URL))
	fake.server.Close()

	_, err := client.Token(context.Background())
	if err == nil {
		t.Fatal("expected a transport error")
	}
	for _, forbidden := range []string{"corp-secret", "corp-id"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaks %q: %v", forbidden, err)
		}
	}
}

func TestTokenHonorsContextCancellation(t *testing.T) {
	fake := newFakeWeCom(t)
	client := New("corp-id", "corp-secret", WithBaseURL(fake.server.URL))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.Token(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 0 {
		t.Fatalf("cancelled call reached gettoken %d times", tokenCalls)
	}
}

func TestInvalidTokenResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()
	client := New("corp-id", "corp-secret", WithBaseURL(server.URL))

	_, err := client.Token(context.Background())
	if err == nil || !strings.Contains(err.Error(), "invalid token response") {
		t.Fatalf("unexpected error: %v", err)
	}
}
