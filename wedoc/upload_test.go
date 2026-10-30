package wedoc

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"testing"
)

func TestUploadChunkedDocument(t *testing.T) {
	var bodies []map[string]any
	doc, paths := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		bodies = append(bodies, body)
		switch path {
		case fileUploadInitPath:
			_, _ = w.Write([]byte(`{"errcode":0,"hit_exist":false,"upload_key":"UK"}`))
		case fileUploadFinishPath:
			_, _ = w.Write([]byte(`{"errcode":0,"file_url":"https://file"}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	})

	content := make([]byte, FileBlockSize+5)
	for i := range content {
		content[i] = byte(i)
	}
	url, err := doc.UploadChunked(context.Background(), "d1", "big.bin", content)
	if err != nil || url != "https://file" {
		t.Fatalf("upload chunked: %q %v", url, err)
	}
	wantPaths := []string{
		fileUploadInitPath, fileUploadPartPath, fileUploadPartPath, fileUploadFinishPath,
	}
	if len(*paths) != len(wantPaths) {
		t.Fatalf("paths = %v", *paths)
	}
	for i, path := range wantPaths {
		if (*paths)[i] != path {
			t.Fatalf("paths = %v", *paths)
		}
	}

	init := bodies[0]
	if init["size"] != float64(len(content)) || init["docid"] != "d1" {
		t.Fatalf("init body: %v", init)
	}
	wantSHA := BlockSHA(content)
	if len(wantSHA) != 2 {
		t.Fatalf("blocks = %d, want 2", len(wantSHA))
	}
	sum := sha1.Sum(content)
	if last := hex.EncodeToString(sum[:]); wantSHA[1] != last {
		t.Fatalf("last block = %s, want %s", wantSHA[1], last)
	}
	gotSHA, ok := init["block_sha"].([]any)
	if !ok || len(gotSHA) != 2 || gotSHA[0] != wantSHA[0] || gotSHA[1] != wantSHA[1] {
		t.Fatalf("block_sha = %v, want %v", init["block_sha"], wantSHA)
	}
	if bodies[1]["index"] != float64(1) || bodies[2]["index"] != float64(2) {
		t.Fatalf("block indexes: %v %v", bodies[1], bodies[2])
	}
	if bodies[1]["upload_key"] != "UK" || bodies[3]["upload_key"] != "UK" {
		t.Fatalf("upload key not reused: %v", bodies)
	}
}

func TestUploadChunkedHitExist(t *testing.T) {
	var requests int
	doc, _ := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		requests++
		_, _ = w.Write([]byte(`{"errcode":0,"hit_exist":true,"file_url":"https://hit"}`))
	})
	url, err := doc.UploadChunked(context.Background(), "d1", "big.bin", []byte("tiny"))
	if err != nil || url != "https://hit" {
		t.Fatalf("hit exist: %q %v", url, err)
	}
	if requests != 1 {
		t.Fatalf("requests after a hit = %d, want 1", requests)
	}
}

func TestInitChunkUploadValidation(t *testing.T) {
	doc, _ := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		_, _ = w.Write([]byte(`{"errcode":0}`))
	})
	ctx := context.Background()
	if _, err := doc.InitChunkUpload(ctx, "d1", "a", 1, nil); err == nil {
		t.Fatal("empty block_sha accepted")
	}
	if _, err := doc.InitChunkUpload(ctx, "d1", "a", 1, []string{"s"}); err == nil {
		t.Fatal("response without upload_key accepted")
	}
}
