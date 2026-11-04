package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenWaiterCanCancelDuringRefresh(t *testing.T) {
	started, release := make(chan struct{}), make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
	}))
	defer server.Close()
	defer close(release)
	c := New("corp", "secret", WithBaseURL(server.URL))
	done := make(chan error, 1)
	go func() { _, err := c.Token(context.Background()); done <- err }()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	waiter := make(chan error, 1)
	go func() { _, err := c.Token(ctx); waiter <- err }()
	select {
	case err := <-waiter:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("canceled waiter blocked behind active refresh")
	}
	// The first refresh is still running: canceling a waiter must not cancel it.
	select {
	case err := <-done:
		t.Fatalf("active refresh unexpectedly ended: %v", err)
	default:
	}
}
