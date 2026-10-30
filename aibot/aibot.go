// Package aibot implements the WeCom API-mode intelligent robot: parsing the
// callback payloads and replying through the response_url handed to the callback.
// It does not route HTTP requests; the application owns the callback URL.
package aibot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultHTTPTimeout bounds one active reply.
const DefaultHTTPTimeout = 10 * time.Second

// maxReplyBytes bounds a reply body, well above the official 20 KiB markdown cap.
const maxReplyBytes = 1 << 20

// Client posts replies to a response_url. It needs no corp token: the one-time
// response_url from the callback is the credential.
type Client struct {
	http *http.Client
}

// Option customizes the reply transport.
type Option func(*Client)

// WithHTTPClient injects the HTTP client used for replies.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		if c != nil {
			client.http = c
		}
	}
}

// New creates a reply client.
func New(options ...Option) *Client {
	c := &Client{http: &http.Client{Timeout: DefaultHTTPTimeout}}
	for _, option := range options {
		option(c)
	}
	return c
}

// Error is a non-zero errcode returned by the reply endpoint.
type Error struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("wecom: robot response failed with errcode %d: %s", e.Errcode, e.Errmsg)
}

// Response answers one callback through its response_url. The URL is valid for
// one hour and can be used once, so failures must be handled by the caller.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101138
func (c *Client) Response(ctx context.Context, responseURL string, reply Reply) error {
	if c == nil || c.http == nil {
		return errors.New("wecom: robot client is unavailable")
	}
	responseURL = strings.TrimSpace(responseURL)
	if !strings.HasPrefix(responseURL, "https://") && !strings.HasPrefix(responseURL, "http://") {
		return errors.New("wecom: robot response: response_url must be an absolute URL")
	}
	if err := reply.validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(reply)
	if err != nil {
		return fmt.Errorf("wecom: robot response: %w", err)
	}
	if len(payload) > maxReplyBytes {
		return errors.New("wecom: robot response: reply exceeds 1 MiB")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, responseURL, bytes.NewReader(payload))
	if err != nil {
		// The parse error echoes the URL, which is the credential, so it is
		// replaced with a stable message.
		return errors.New("wecom: robot response: invalid response_url")
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.http.Do(request)
	if err != nil {
		return replyTransportError(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, maxReplyBytes))
	if err != nil {
		return fmt.Errorf("wecom: robot response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom: robot response: unexpected status %s", response.Status)
	}
	var result Error
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("wecom: robot response: decode reply: %w", err)
	}
	if result.Errcode != 0 {
		return &result
	}
	return nil
}

// replyTransportError reports a failed reply request without echoing the
// response_url. That URL is the credential: it carries a one-hour response_code,
// so wrapping the underlying *url.Error would put a usable authorisation token
// into ordinary error logs. The message is therefore stable, matching the shared
// client, and cancellation stays recognisable through errors.Is.
func replyTransportError(err error) error {
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("wecom: robot response: %w", context.Canceled)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("wecom: robot response: %w", context.DeadlineExceeded)
	}
	return errors.New("wecom: robot response: request failed")
}
