package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestResponseLimits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		download bool
		size     int
		limit    int64
		tooLarge bool
	}{
		{"json exact", false, maxResponseBytes, 0, false},
		{"json overflow", false, maxResponseBytes + 1, 0, true},
		{"download above json limit", true, maxResponseBytes + 1, 0, false},
		{"download exact", true, 32, 32, false},
		{"download overflow", true, 33, 32, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := strings.Repeat("x", tc.size)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/cgi-bin/gettoken" {
					_, _ = w.Write([]byte(`{"access_token":"token","expires_in":7200}`))
					return
				}
				// Flush forces chunked responses, exercising limits without Content-Length.
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				_, _ = w.Write([]byte(payload))
			}))
			defer server.Close()
			c := New("corp", "secret", WithBaseURL(server.URL), WithDownloadLimit(tc.limit))
			var data []byte
			var err error
			if tc.download {
				data, err = c.Download(context.Background(), "download", "/file", nil)
			} else {
				req, e := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
				if e != nil {
					t.Fatal(e)
				}
				data, err = c.do(req)
			}
			if tc.tooLarge {
				if !errors.Is(err, ErrResponseTooLarge) || data != nil {
					t.Fatalf("data len=%d error=%v", len(data), err)
				}
			} else if err != nil || string(data) != payload {
				t.Fatalf("data len=%d want=%d error=%v", len(data), len(payload), err)
			}
		})
	}
}

func TestDownloadLimitIgnoresInvalidValues(t *testing.T) {
	c := New("corp", "secret", WithDownloadLimit(0), WithDownloadLimit(-1), WithDownloadLimit(1<<63-1))
	if c.downloadLimit != defaultDownloadLimit {
		t.Fatalf("limit=%d", c.downloadLimit)
	}
}
