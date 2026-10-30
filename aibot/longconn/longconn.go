// Package bot keeps one WeCom bot long connection alive: subscribe, heartbeat,
// reconnect and protocol frames. Callback handling and reply policy belong to
// the caller; by default no application responses are sent.
package longconn

import (
	"context"
	"errors"
	mathrand "math/rand"
	"strings"
	"time"

	"github.com/gobwas/ws"
)

const (
	// DefaultEndpoint is the public long-connection address. Only tests and
	// regional deployments override it.
	DefaultEndpoint      = "wss://openws.work.weixin.qq.com"
	defaultPingInterval  = 30 * time.Second
	defaultReconnectMin  = time.Second
	defaultReconnectMax  = 30 * time.Second
	defaultDialTimeout   = 10 * time.Second
	defaultSubscribeWait = 15 * time.Second
	defaultWriteTimeout  = 10 * time.Second
)

// Bot is one bot long connection. It is safe for a single Start call; all
// writes are serialized so replies, pings and protocol control frames never
// interleave.
type Bot struct {
	botID  string
	secret string

	endpoint string
	dialer   ws.Dialer

	pingInterval  time.Duration
	reconnectMin  time.Duration
	reconnectMax  time.Duration
	subscribeWait time.Duration
	writeTimeout  time.Duration
	handler       Handler
	observer      Observer
}

// Option customizes a Bot. Options exist for tests and regional endpoints.
type Option func(*Bot)

// WithEndpoint overrides the WebSocket address.
func WithEndpoint(endpoint string) Option {
	return func(c *Bot) {
		if trimmed := strings.TrimSpace(endpoint); trimmed != "" {
			c.endpoint = trimmed
		}
	}
}

// WithPingInterval overrides the heartbeat interval.
func WithPingInterval(interval time.Duration) Option {
	return func(c *Bot) {
		if interval > 0 {
			c.pingInterval = interval
		}
	}
}

// WithReconnectBounds overrides the exponential backoff window.
func WithReconnectBounds(min, max time.Duration) Option {
	return func(c *Bot) {
		if min > 0 {
			c.reconnectMin = min
		}
		if max > 0 && max >= c.reconnectMin {
			c.reconnectMax = max
		}
	}
}

// WithSubscribeTimeout bounds how long a subscription acknowledgement may take.
func WithSubscribeTimeout(timeout time.Duration) Option {
	return func(c *Bot) {
		if timeout > 0 {
			c.subscribeWait = timeout
		}
	}
}

// WithHandler installs the callback handler. Without one, callbacks receive no
// application response. A handler may also choose not to use its responder.
func WithHandler(handler Handler) Option {
	return func(c *Bot) { c.handler = handler }
}

// WithObserver installs an optional synchronous diagnostics hook. It must return
// promptly. No diagnostics are logged globally, and credentials, callback bodies
// and server-supplied error text are never included in diagnostics.
func WithObserver(observer Observer) Option {
	return func(c *Bot) { c.observer = observer }
}

// New creates a long-connection client. Start validates that credentials exist.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func New(botID, secret string, options ...Option) *Bot {
	client := &Bot{
		botID:         strings.TrimSpace(botID),
		secret:        strings.TrimSpace(secret),
		endpoint:      DefaultEndpoint,
		dialer:        ws.DefaultDialer,
		pingInterval:  defaultPingInterval,
		reconnectMin:  defaultReconnectMin,
		reconnectMax:  defaultReconnectMax,
		subscribeWait: defaultSubscribeWait,
		writeTimeout:  defaultWriteTimeout,
	}
	for _, option := range options {
		if option != nil {
			option(client)
		}
	}
	return client
}

// Start keeps the long connection alive until ctx is cancelled. It returns nil on
// cancellation. Transport failures and transient subscription failures retry with
// jittered exponential backoff. Invalid credentials, permanent subscription
// rejections and handler errors return to the caller.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (c *Bot) Start(ctx context.Context) error {
	if ctx.Err() != nil {
		return nil
	}
	if c.botID == "" || c.secret == "" {
		return ErrMissingCredentials
	}
	backoff := c.reconnectMin
	for {
		if ctx.Err() != nil {
			return nil
		}
		subscribed, err := c.runOnce(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if subscribed {
			backoff = c.reconnectMin
		}
		var rejection *SubscriptionError
		var handlerErr *HandlerError
		if (errors.As(err, &rejection) && !rejection.Temporary()) || errors.As(err, &handlerErr) {
			return err
		}
		wait := c.jitter(backoff)
		c.observe(Diagnostic{Kind: DiagnosticRetry, Err: err, RetryAfter: wait})
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil
		case <-timer.C:
		}
		if backoff < c.reconnectMax {
			backoff *= 2
			if backoff > c.reconnectMax {
				backoff = c.reconnectMax
			}
		}
	}
}

// jitter spreads reconnects within [delay, delay*1.25] so several instances do
// not retry in lockstep. The result never exceeds reconnectMax.
func (c *Bot) jitter(delay time.Duration) time.Duration {
	if delay <= 0 {
		return 0
	}
	if extra := int64(delay) / 4; extra > 0 {
		delay += time.Duration(mathrand.Int63n(extra + 1))
	}
	if delay > c.reconnectMax {
		delay = c.reconnectMax
	}
	return delay
}
