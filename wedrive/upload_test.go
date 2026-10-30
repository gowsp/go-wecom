package wedrive

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

// sha1Hex is the reference digest the last block must match.
func sha1Hex(content []byte) string {
	sum := sha1.Sum(content)
	return hex.EncodeToString(sum[:])
}

// uploadRecorder captures the chunked upload conversation.
type uploadRecorder struct {
	mu     sync.Mutex
	paths  []string
	bodies []map[string]any
}

func (r *uploadRecorder) record(path string, body map[string]any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.paths = append(r.paths, path)
	r.bodies = append(r.bodies, body)
}

func newUploadServer(t *testing.T, initResponse string) (*uploadRecorder, *client.Client) {
	t.Helper()
	recorder := &uploadRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		recorder.record(r.URL.Path, body)
		switch r.URL.Path {
		case "/cgi-bin/wedrive/file_upload_init":
			_, _ = w.Write([]byte(initResponse))
		case "/cgi-bin/wedrive/file_upload_finish":
			_, _ = w.Write([]byte(`{"errcode":0,"fileid":"f1"}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	}))
	t.Cleanup(server.Close)
	return recorder, client.New("corp", "secret", client.WithBaseURL(server.URL))
}

func TestUploadChunked(t *testing.T) {
	recorder, api := newUploadServer(t, `{"errcode":0,"hit_exist":false,"upload_key":"uk"}`)
	svc := New(api)

	content := make([]byte, BlockSize+10)
	for i := range content {
		content[i] = byte(i)
	}
	id, err := svc.UploadChunked(context.Background(), UploadInitRequest{SpaceID: "sp1", FileName: "big.bin"}, content)
	if err != nil || id != "f1" {
		t.Fatalf("upload: %q %v", id, err)
	}

	wantPaths := []string{
		"/cgi-bin/wedrive/file_upload_init",
		"/cgi-bin/wedrive/file_upload_part",
		"/cgi-bin/wedrive/file_upload_part",
		"/cgi-bin/wedrive/file_upload_finish",
	}
	if strings.Join(recorder.paths, ",") != strings.Join(wantPaths, ",") {
		t.Fatalf("paths = %v", recorder.paths)
	}
	init := recorder.bodies[0]
	if init["size"] != float64(len(content)) || init["spaceid"] != "sp1" {
		t.Fatalf("init body: %v", init)
	}
	sha, ok := init["block_sha"].([]any)
	if !ok || len(sha) != 2 || sha[1] != sha1Hex(content) {
		t.Fatalf("block_sha: %v", init["block_sha"])
	}
	for index, body := range recorder.bodies[1:3] {
		if body["index"] != float64(index+1) || body["upload_key"] != "uk" {
			t.Fatalf("block %d body: %v", index+1, body)
		}
	}
	if recorder.bodies[3]["upload_key"] != "uk" {
		t.Fatalf("finish body: %v", recorder.bodies[3])
	}
}

func TestUploadChunkedHitExistSkipsBlocks(t *testing.T) {
	recorder, api := newUploadServer(t, `{"errcode":0,"hit_exist":true,"fileid":"f9"}`)
	id, err := New(api).UploadChunked(context.Background(), UploadInitRequest{Ticket: "tk", FileName: "big.bin"},
		make([]byte, 32))
	if err != nil || id != "f9" {
		t.Fatalf("hit exist: %q %v", id, err)
	}
	if len(recorder.paths) != 1 {
		t.Fatalf("blocks sent after a hit: %v", recorder.paths)
	}
	if ticket := recorder.bodies[0]["selected_ticket"]; ticket != "tk" {
		t.Fatalf("ticket not sent: %v", recorder.bodies[0])
	}
}

func TestUploadLocationRules(t *testing.T) {
	_, api := newUploadServer(t, `{"errcode":0,"upload_key":"uk"}`)
	svc := New(api)
	ctx := context.Background()

	// A space location and a picker ticket are alternatives, never both.
	if _, err := svc.UploadInit(ctx, UploadInitRequest{
		SpaceID: "sp1", Ticket: "tk", FileName: "a", Size: 1, BlockSHA: []string{"x"},
	}); err == nil {
		t.Fatal("space and ticket accepted together")
	}
	if _, err := svc.UploadInit(ctx, UploadInitRequest{
		FileName: "a", Size: 1, BlockSHA: []string{"x"},
	}); err == nil {
		t.Fatal("upload init without a location accepted")
	}
	if _, err := svc.UploadFile(ctx, "", "", "tk", "a", "YQ=="); err != nil {
		t.Fatalf("ticket-only upload: %v", err)
	}
	if _, err := svc.DownloadFile(ctx, "", "tk"); err != nil {
		t.Fatalf("ticket-only download: %v", err)
	}
	if _, err := svc.DownloadFile(ctx, "", ""); err == nil {
		t.Fatal("download without a location accepted")
	}
	if _, err := InlineBase64(make([]byte, 10<<20+1)); err == nil {
		t.Fatal("inline upload above 10 MiB accepted")
	}
	if encoded, err := InlineBase64([]byte("x")); err != nil || encoded != "eA==" {
		t.Fatalf("inline encoding: %q %v", encoded, err)
	}
}
