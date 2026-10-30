package wecom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gowsp/go-wecom"
	"github.com/gowsp/go-wecom/client"
)

// Business clients must reuse the injected transport rather than allocating
// independent token caches for messages and documents.
func TestBusinessClientsShareTokenCache(t *testing.T) {
	var tokens, messages, documents atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/cgi-bin/gettoken" {
			tokens.Add(1)
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"shared-token","expires_in":7200}`))
			return
		}
		if r.URL.Query().Get("access_token") != "shared-token" {
			t.Error("business request did not use the shared token")
		}
		switch r.URL.Path {
		case "/cgi-bin/message/send":
			messages.Add(1)
			_, _ = w.Write([]byte(`{"errcode":0}`))
		case "/cgi-bin/wedoc/get_doc_base_info":
			documents.Add(1)
			_, _ = w.Write([]byte(`{"errcode":0,"doc_base_info":{"docid":"doc-1","doc_name":"商机"}}`))
		default:
			t.Errorf("unexpected endpoint: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	api := wecom.New("corp", "secret", 1000002, client.WithBaseURL(server.URL))
	messagesAPI := api.Message
	documentsAPI := api.WeDoc
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := messagesAPI.SendText(ctx, "user", "hello"); err != nil {
			t.Fatal(err)
		}
		info, err := documentsAPI.GetDocBaseInfo(ctx, "doc-1")
		if err != nil {
			t.Fatal(err)
		}
		if info.DocID != "doc-1" || info.DocName != "商机" {
			t.Fatalf("unexpected document: %+v", info)
		}
	}
	if tokens.Load() != 1 || messages.Load() != 3 || documents.Load() != 3 {
		t.Fatalf("requests: token=%d message=%d document=%d; want 1/3/3", tokens.Load(), messages.Load(), documents.Load())
	}
}
