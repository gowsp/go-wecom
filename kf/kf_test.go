package kf

import (
	"context"
	"encoding/json"
	"github.com/gowsp/go-wecom/client"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKfClientTransferQueuesForEmployee(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":3600}`))
			return
		}
		if r.URL.Path != "/cgi-bin/kf/service_state/trans" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	c := New(client.New("corp", "secret", client.WithBaseURL(srv.URL)))
	if err := c.Transfer(context.Background(), "account", "external", "employee"); err != nil {
		t.Fatal(err)
	}
	if got["open_kfid"] != "account" || got["external_userid"] != "external" || got["servicer_userid"] != "employee" || got["service_state"] != float64(3) {
		t.Fatalf("body=%v", got)
	}
}

func TestKfClientAccept(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":3600}`))
			return
		}
		if r.URL.Path != "/cgi-bin/kf/service_state/trans" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	}))
	defer srv.Close()
	c := New(client.New("corp", "secret", client.WithBaseURL(srv.URL)))
	if err := c.Accept(context.Background(), "account", "external"); err != nil {
		t.Fatal(err)
	}
	if got["service_state"] != float64(2) || got["open_kfid"] != "account" || got["external_userid"] != "external" {
		t.Fatalf("body=%v", got)
	}
	if _, ok := got["servicer_userid"]; ok {
		t.Fatalf("assistant state must not set a servicer: %v", got)
	}
}

func TestKfClientTransferRequiresServicer(t *testing.T) {
	c := New(client.New("corp", "secret"))
	if err := c.Transfer(context.Background(), "account", "external", "  "); err == nil {
		t.Fatal("a handoff without an employee must fail")
	}
}
