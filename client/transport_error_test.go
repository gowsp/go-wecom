package client

import (
	"crypto/x509"
	"errors"
	"net"
	"net/url"
	"strings"
	"testing"
)

func TestTransportErrorClassificationAndRedaction(t *testing.T) {
	for _, tc := range []struct {
		name, kind string
		err        error
	}{
		{"dns", "dns", &net.DNSError{Err: "secret-token", Name: "secret-host"}},
		{"timeout", "timeout", &net.OpError{Op: "dial", Err: &net.DNSError{IsTimeout: true}}},
		{"certificate", "tls", x509.UnknownAuthorityError{}},
		{"opaque", "network", errors.New("secret-token")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := transportError("request", &url.Error{Op: "Get", URL: "https://example.com?corpsecret=secret-token", Err: tc.err})
			// A DNS timeout retains the more specific DNS classification.
			want := tc.kind
			if tc.name == "timeout" {
				want = "dns"
			}
			if e.Kind != want {
				t.Fatalf("kind=%s want %s", e.Kind, want)
			}
			if strings.Contains(e.Error(), "secret") || errors.Unwrap(e) != nil {
				t.Fatal("unsafe transport error")
			}
			if tc.name == "timeout" && !e.Timeout() {
				t.Fatal("timeout flag missing")
			}
		})
	}
}
