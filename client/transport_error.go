package client

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/url"
)

// TransportError describes a failed HTTP exchange without retaining the request
// URL, response body, or underlying error text (which can contain credentials).
// It deliberately has no Unwrap method. Use errors.As to inspect its fields.
type TransportError struct {
	Stage    string // "request" or "read"
	Kind     string // "dns", "tls", "timeout", "network", or "read"
	TimedOut bool
}

func (e *TransportError) Error() string { return "wecom: " + e.Stage + " failed (" + e.Kind + ")" }

// Timeout implements the timeout portion of net.Error.
func (e *TransportError) Timeout() bool { return e.TimedOut }

func transportError(stage string, err error) *TransportError {
	e := &TransportError{Stage: stage, Kind: "network"}
	if stage == "read" {
		e.Kind = "read"
	}
	// Inspect typed errors only; never classify by attacker-controlled text.
	var ue *url.Error
	if errors.As(err, &ue) {
		err = ue.Err
	}
	var ne net.Error
	if errors.As(err, &ne) {
		e.TimedOut = ne.Timeout()
	}
	var dns *net.DNSError
	var cert *tls.CertificateVerificationError
	var unknown x509.UnknownAuthorityError
	var hostname x509.HostnameError
	var invalid x509.CertificateInvalidError
	var record tls.RecordHeaderError
	switch {
	case errors.As(err, &dns):
		e.Kind = "dns"
	case errors.As(err, &cert), errors.As(err, &unknown), errors.As(err, &hostname), errors.As(err, &invalid), errors.As(err, &record):
		e.Kind = "tls"
	case e.TimedOut:
		e.Kind = "timeout"
	}
	return e
}
