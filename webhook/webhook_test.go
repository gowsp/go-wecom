package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendText(t *testing.T) {
	var gotKey, gotPath string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey, gotPath = r.URL.Query().Get("key"), r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		_, _ = w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
	}))
	defer server.Close()

	w := New("KEY1", WithBaseURL(server.URL))
	err := w.Send(context.Background(), Message{
		MsgType: "text",
		Text:    &Text{Content: "hello", MentionedList: []string{"@all"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotKey != "KEY1" || gotPath != "/cgi-bin/webhook/send" {
		t.Fatalf("unexpected request: %s %s", gotPath, gotKey)
	}
	if body["msgtype"] != "text" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestSendReportsErrcode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"errcode":93000,"errmsg":"invalid webhook url"}`))
	}))
	defer server.Close()
	w := New("KEY1", WithBaseURL(server.URL))
	err := w.Send(context.Background(), Message{MsgType: "text", Text: &Text{Content: "x"}})
	if err == nil || !strings.Contains(err.Error(), "93000") {
		t.Fatalf("expected errcode error: %v", err)
	}
}

func TestUploadMediaMultipart(t *testing.T) {
	var contentType, mediaType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		mediaType = r.URL.Query().Get("type")
		raw, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(raw), "a.png") {
			t.Errorf("missing file name in body")
		}
		_, _ = w.Write([]byte(`{"errcode":0,"media_id":"M1"}`))
	}))
	defer server.Close()
	w := New("KEY1", WithBaseURL(server.URL))
	id, err := w.UploadMedia(context.Background(), "image", "a.png", []byte("png"))
	if err != nil || id != "M1" {
		t.Fatalf("upload media: %v %q", err, id)
	}
	if !strings.HasPrefix(contentType, "multipart/form-data") || mediaType != "image" {
		t.Fatalf("unexpected request: %s %s", contentType, mediaType)
	}
}

func TestEmptyKeyRejected(t *testing.T) {
	w := New("  ")
	if err := w.Send(context.Background(), Message{MsgType: "text"}); err == nil {
		t.Fatal("empty key must fail")
	}
	if _, err := w.UploadMedia(context.Background(), "image", "a", []byte("x")); err == nil {
		t.Fatal("empty key must fail for upload")
	}
}
