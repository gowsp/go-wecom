package longconn

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// runOnce dials, subscribes, and serves one connection until it fails or the
// context is cancelled. The boolean reports whether the subscription succeeded,
// allowing Start to distinguish connect failures from dropped connections.
func (c *Bot) runOnce(ctx context.Context) (bool, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()

	if err := c.subscribe(ctx, conn); err != nil {
		return false, err
	}
	c.observe(Diagnostic{Kind: DiagnosticConnected})
	if ctx.Err() != nil {
		return true, nil
	}
	return true, c.serve(ctx, conn)
}

// dial bounds the TCP connect and WebSocket handshake so a blackholed
// endpoint cannot stall the reconnect loop forever.
func (c *Bot) dial(ctx context.Context) (*connection, error) {
	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()
	raw, reader, _, err := c.dialer.Dial(dialCtx, c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("wecom: websocket dial failed: %w", err)
	}
	if reader == nil {
		reader = bufio.NewReader(raw)
	}
	return &connection{Conn: raw, reader: reader, writeTimeout: c.writeTimeout}, nil
}

// subscribe sends aibot_subscribe and waits for its acknowledgement. Frames
// that are not the matching acknowledgement are ignored; a non-zero errcode
// aborts with the code only, never the response text or credentials.
func (c *Bot) subscribe(ctx context.Context, conn *connection) error {
	reqID := newReqID()
	body := subscribeBody{BotID: c.botID, Secret: c.secret}
	if err := c.send(ctx, conn, cmdSubscribe, reqID, body); err != nil {
		return err
	}
	_ = conn.SetReadDeadline(time.Now().Add(c.subscribeWait))
	defer func() { _ = conn.SetReadDeadline(time.Time{}) }()
	for {
		data, op, err := wsutil.ReadServerData(conn)
		if err != nil {
			return fmt.Errorf("wecom: subscribe failed: %w", err)
		}
		if op != ws.OpText && op != ws.OpBinary {
			continue
		}
		var msg envelope
		if json.Unmarshal(data, &msg) != nil {
			continue
		}
		if msg.Headers.ReqID != reqID || msg.Errcode == nil {
			continue
		}
		if *msg.Errcode != 0 {
			return &SubscriptionError{Code: *msg.Errcode}
		}
		return nil
	}
}

// serve reads frames, dispatches callbacks, and sends a heartbeat on every
// interval. One goroutine reads; that goroutine may send protocol-level replies
// (pongs, close acknowledgements) while this loop sends business replies, so
// all writes go through connection's mutex. The loop returns nil when ctx is
// cancelled and the read error otherwise.
func (c *Bot) serve(ctx context.Context, conn *connection) error {
	frames := make(chan inbound, 64)
	readErr := make(chan error, 1)
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			data, op, err := wsutil.ReadServerData(conn)
			if err != nil {
				readErr <- err
				return
			}
			select {
			case frames <- inbound{data: data, op: op}:
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()
	// A pong (or any other frame) proves the peer is alive. Three unanswered
	// heartbeats mean a half-open connection that only a reconnect can repair.
	missedPongs := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-readErr:
			if ctx.Err() != nil {
				return nil
			}
			return err
		case <-ticker.C:
			if missedPongs >= 3 {
				return errors.New("wecom: heartbeat timed out")
			}
			if err := c.sendPing(ctx, conn); err != nil {
				return err
			}
			missedPongs++
		case frame := <-frames:
			missedPongs = 0
			if err := c.handleFrame(ctx, conn, frame); err != nil {
				return err
			}
		}
	}
}

type inbound struct {
	data []byte
	op   ws.OpCode
}
