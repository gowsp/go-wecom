package media

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

func newFake(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, raw []byte)) (*Media, *[]string) {
	t.Helper()
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		calls = append(calls, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
		handler(w, r, raw)
	}))
	t.Cleanup(server.Close)
	return New(client.New("corp", "secret", client.WithBaseURL(server.URL))), &calls
}

func TestUploadSendsMultipart(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, r *http.Request, raw []byte) {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("unexpected content type: %s", contentType)
		}
		if !bytes.Contains(raw, []byte("a.txt")) || !bytes.Contains(raw, []byte("hello")) {
			t.Errorf("multipart body missing file: %d bytes", len(raw))
		}
		_, _ = w.Write([]byte(`{"errcode":0,"type":"file","media_id":"M1","created_at":"1"}`))
	})
	item, err := c.Upload(context.Background(), TypeFile, "a.txt", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if item.MediaID != "M1" || item.Type != "file" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if !strings.Contains((*calls)[0], "type=file") {
		t.Fatalf("missing type query: %s", (*calls)[0])
	}
}

func TestUploadValidation(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ []byte) {})
	ctx := context.Background()
	if _, err := c.Upload(ctx, "unknown", "a", []byte("x")); err == nil {
		t.Fatal("unsupported type must fail")
	}
	if _, err := c.Upload(ctx, TypeFile, "a", nil); err == nil {
		t.Fatal("empty content must fail")
	}
	if _, err := c.Upload(ctx, TypeFile, " ", []byte("x")); err == nil {
		t.Fatal("blank name must fail")
	}
}

func TestUploadImageReturnsURL(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		_, _ = w.Write([]byte(`{"errcode":0,"url":"https://img"}`))
	})
	url, err := c.UploadImage(context.Background(), "a.png", []byte("png"))
	if err != nil || url != "https://img" {
		t.Fatalf("upload image: %v %q", err, url)
	}
	if strings.Contains((*calls)[0], "type=") {
		t.Fatalf("uploadimg carries no type: %s", (*calls)[0])
	}
}

func TestDownloadReturnsBinary(t *testing.T) {
	c, calls := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write([]byte{0x00, 0x01, 0x02})
	})
	data, err := c.Download(context.Background(), "M1")
	if err != nil || !bytes.Equal(data, []byte{0, 1, 2}) {
		t.Fatalf("download: %v %v", data, err)
	}
	if !strings.Contains((*calls)[0], "media_id=M1") {
		t.Fatalf("missing media id: %s", (*calls)[0])
	}
}

func TestDownloadReportsJSONError(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		_, _ = w.Write([]byte(`{"errcode":40007,"errmsg":"invalid media_id"}`))
	})
	if _, err := c.Download(context.Background(), "M1"); err == nil || !strings.Contains(err.Error(), "40007") {
		t.Fatalf("expected api error: %v", err)
	}
}

func TestAsyncUpload(t *testing.T) {
	c, _ := newFake(t, func(w http.ResponseWriter, r *http.Request, raw []byte) {
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		if strings.HasSuffix(r.URL.Path, "upload_by_url") {
			if body["url"] != "https://src" {
				t.Errorf("unexpected body: %+v", body)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"jobid":"J1"}`))
			return
		}
		if body["jobid"] != "J1" {
			t.Errorf("unexpected job body: %+v", body)
		}
		_, _ = w.Write([]byte(`{"errcode":0,"status":2,"detail":{"media_id":"M2"}}`))
	})
	ctx := context.Background()
	jobID, err := c.UploadByURL(ctx, UploadRequest{Type: TypeVideo, URL: "https://src", FileName: "v.mp4"})
	if err != nil || jobID != "J1" {
		t.Fatalf("upload by url: %v %q", err, jobID)
	}
	job, err := c.UploadResult(ctx, "J1")
	if err != nil || job.Status != 2 || job.Detail == nil || job.Detail.MediaID != "M2" {
		t.Fatalf("upload result: %+v %v", job, err)
	}
	if _, err := c.UploadByURL(ctx, UploadRequest{Type: TypeVideo}); err == nil {
		t.Fatal("missing url must fail")
	}
}
