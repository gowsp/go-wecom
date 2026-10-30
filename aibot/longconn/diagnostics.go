package longconn

import (
	"errors"
	"fmt"
	"time"
)

// ErrMissingCredentials reports empty bot ID or secret, before any dial attempt.
var ErrMissingCredentials = errors.New("wecom: bot ID and secret are required")

// SubscriptionError is a subscription rejection. Server-provided error text is
// intentionally omitted because it can contain sensitive data.
type SubscriptionError struct {
	Code int
}

func (e *SubscriptionError) Error() string {
	return fmt.Sprintf("wecom: subscribe rejected with errcode %d", e.Code)
}

// Temporary reports the protocol's system-busy response (-1). Other subscription
// rejections require caller intervention rather than endlessly retrying the same
// credentials. Unknown codes are returned so callers can apply their own policy.
func (e *SubscriptionError) Temporary() bool { return e.Code == -1 }

// HandlerError reports an application callback failure. It stops Start rather
// than reconnecting and potentially replaying failed business logic.
type HandlerError struct {
	Err error
}

// Error omits the application's error text to keep diagnostics safe.
func (e *HandlerError) Error() string { return "wecom: callback handler failed" }
func (e *HandlerError) Unwrap() error { return e.Err }

// DiagnosticKind identifies a connection lifecycle or acknowledgement notice.
type DiagnosticKind string

const (
	DiagnosticConnected DiagnosticKind = "connected"
	DiagnosticRetry     DiagnosticKind = "retry"
	DiagnosticRejected  DiagnosticKind = "rejected"
)

// Diagnostic contains SDK diagnostics, never credentials or callback payloads.
// RetryAfter is set for retry notices; Code is set for rejected acknowledgements.
// Terminal failures are returned by Start, not sent as retry notices.
type Diagnostic struct {
	Kind       DiagnosticKind
	Err        error
	Code       int
	RetryAfter time.Duration
}

// Observer receives optional synchronous SDK diagnostics. It must not block.
type Observer func(Diagnostic)

func (c *Bot) observe(d Diagnostic) {
	if c.observer != nil {
		c.observer(d)
	}
}

// responseWriteError distinguishes a recoverable response transport failure
// from a handler's own error or invalid response encoding.
type responseWriteError struct{ error }

func (e *responseWriteError) Unwrap() error { return e.error }
