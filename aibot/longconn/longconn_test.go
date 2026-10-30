package longconn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	"github.com/gowsp/go-wecom/aibot"
)

const (
	testBotID   = "bot-0001"
	testSecret  = "long-connection-secret"
	testHelp    = "已收到测试消息。"
	testWelcome = "欢迎使用测试机器人。"
)

// peerStream lets wsutil read buffered handshake data while control replies go
// straight to the socket.
type peerStream struct {
	io.Reader
	io.Writer
}

// peer is one accepted server connection in tests.
type peer struct {
	conn   net.Conn
	stream peerStream
}

func (p *peer) read(timeout time.Duration) (map[string]any, error) {
	if err := p.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}
	data, _, err := wsutil.ReadClientData(p.stream)
	if err != nil {
		return nil, err
	}
	var frame map[string]any
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, fmt.Errorf("invalid JSON frame: %w", err)
	}
	return frame, nil
}

func (p *peer) write(value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return wsutil.WriteServerText(p.conn, data)
}

func (p *peer) writeText(text string) error {
	return wsutil.WriteServerText(p.conn, []byte(text))
}

func (p *peer) writeBinary(data []byte) error {
	return wsutil.WriteServerBinary(p.conn, data)
}

func (p *peer) ack(reqID string, errcode int) error {
	return p.write(map[string]any{
		"headers": map[string]any{"req_id": reqID},
		"errcode": errcode,
		"errmsg":  "ok",
	})
}

// testServer accepts WebSocket connections and runs handle for each one.
type testServer struct {
	t          *testing.T
	url        string
	httpServer *httptest.Server
	conns      atomic.Int32
	handlerErr chan error
	mu         sync.Mutex
	open       []net.Conn
	wg         sync.WaitGroup
}

func startWXWorkServer(t *testing.T, handle func(index int, p *peer) error) *testServer {
	t.Helper()
	server := &testServer{t: t, handlerErr: make(chan error, 32)}
	server.httpServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, reader, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			server.record(fmt.Errorf("websocket upgrade failed: %w", err))
			return
		}
		server.mu.Lock()
		server.open = append(server.open, conn)
		server.mu.Unlock()

		index := int(server.conns.Add(1))
		server.wg.Add(1)
		go func() {
			defer server.wg.Done()
			defer conn.Close()
			if err := handle(index, &peer{conn: conn, stream: peerStream{Reader: reader, Writer: conn}}); err != nil {
				server.record(fmt.Errorf("connection %d: %w", index, err))
			}
		}()
	}))
	server.url = "ws" + strings.TrimPrefix(server.httpServer.URL, "http")
	t.Cleanup(func() {
		server.mu.Lock()
		for _, conn := range server.open {
			_ = conn.Close()
		}
		server.open = nil
		server.mu.Unlock()
		server.wg.Wait()
		server.httpServer.Close()
	})
	return server
}

// record keeps the first handler failure without blocking the handler.
func (s *testServer) record(err error) {
	select {
	case s.handlerErr <- err:
	default:
	}
}

// clientRun owns one Run goroutine so tests can stop it deterministically.
type clientRun struct {
	cancel context.CancelFunc
	done   chan error
	once   sync.Once
	err    error
}

func startClient(t *testing.T, client *Bot) *clientRun {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	run := &clientRun{cancel: cancel, done: make(chan error, 1)}
	go func() { run.done <- client.Start(ctx) }()
	t.Cleanup(func() {
		if err := run.stop(); err != nil {
			t.Errorf("client did not stop cleanly: %v", err)
		}
	})
	return run
}

func (r *clientRun) stop() error {
	r.cancel()
	r.once.Do(func() {
		select {
		case r.err = <-r.done:
		case <-time.After(3 * time.Second):
			r.err = errors.New("client did not stop within 3s")
		}
	})
	return r.err
}

// waitFor blocks until a signal arrives, the server handler fails, or the
// timeout elapses. It must only be called from the test goroutine.
func waitFor[T any](t *testing.T, ch <-chan T, server *testServer, timeout time.Duration, what string) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case err := <-server.handlerErr:
		t.Fatalf("%s: %v", what, err)
	case <-time.After(timeout):
		t.Fatalf("%s: timed out", what)
	}
	var zero T
	return zero
}

func frameField(frame map[string]any, path ...string) (any, bool) {
	var current any = frame
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func stringAt(frame map[string]any, path ...string) string {
	value, _ := frameField(frame, path...)
	text, _ := value.(string)
	return text
}

func expectCommand(p *peer, want string, timeout time.Duration) (map[string]any, error) {
	frame, err := p.read(timeout)
	if err != nil {
		return nil, err
	}
	if got := stringAt(frame, "cmd"); got != want {
		return nil, fmt.Errorf("command = %q, want %q", got, want)
	}
	return frame, nil
}

func callback(cmd, reqID string, body map[string]any) map[string]any {
	return map[string]any{
		"cmd":     cmd,
		"headers": map[string]any{"req_id": reqID},
		"body":    body,
	}
}

func subscribeAck(p *peer) error {
	sub, err := expectCommand(p, cmdSubscribe, 2*time.Second)
	if err != nil {
		return err
	}
	return p.ack(stringAt(sub, "headers", "req_id"), 0)
}

// holdConnection blocks until the client closes the connection.
func holdConnection(p *peer) error {
	for {
		if _, err := p.read(10 * time.Second); err != nil {
			return nil
		}
	}
}

func TestSubscribeAndRespondHelp(t *testing.T) {
	replyCh := make(chan map[string]any, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		sub, err := expectCommand(p, cmdSubscribe, 2*time.Second)
		if err != nil {
			return err
		}
		if got := stringAt(sub, "headers", "req_id"); got == "" {
			return errors.New("subscribe request has no req_id")
		}
		if got := stringAt(sub, "body", "bot_id"); got != testBotID {
			return fmt.Errorf("subscribe bot_id = %q, want %q", got, testBotID)
		}
		if got := stringAt(sub, "body", "secret"); got != testSecret {
			return errors.New("subscribe request carries the wrong secret")
		}
		if err := p.ack(stringAt(sub, "headers", "req_id"), 0); err != nil {
			return err
		}
		if err := p.write(callback(cmdMsgCallback, "callback-1", map[string]any{
			"msgid":   "m-1",
			"msgtype": "text",
			"text":    map[string]any{"content": "hello robot"},
		})); err != nil {
			return err
		}
		reply, err := expectCommand(p, cmdRespondMsg, 2*time.Second)
		if err != nil {
			return err
		}
		replyCh <- reply
		return nil
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url), WithHandler(testReplyHandler)))

	reply := waitFor(t, replyCh, server, 3*time.Second, "help reply")
	if got := stringAt(reply, "headers", "req_id"); got != "callback-1" {
		t.Fatalf("reply req_id = %q, want the callback req_id", got)
	}
	if got := stringAt(reply, "body", "msgtype"); got != "stream" {
		t.Fatalf("reply msgtype = %q, want stream", got)
	}
	finish, ok := frameField(reply, "body", "stream", "finish")
	if !ok || finish != true {
		t.Fatalf("stream.finish = %v, want true", finish)
	}
	if got := stringAt(reply, "body", "stream", "content"); got != testHelp {
		t.Fatalf("stream.content = %q, want %q", got, testHelp)
	}
	if got := stringAt(reply, "body", "stream", "id"); got == "" {
		t.Fatal("stream.id must not be empty")
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestSubscribeBusyReconnects(t *testing.T) {
	resubscribed := make(chan struct{}, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		sub, err := expectCommand(p, cmdSubscribe, 2*time.Second)
		if err != nil {
			return err
		}
		reqID := stringAt(sub, "headers", "req_id")
		if index == 1 {
			return p.ack(reqID, -1)
		}
		if err := p.ack(reqID, 0); err != nil {
			return err
		}
		resubscribed <- struct{}{}
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url), WithReconnectBounds(5*time.Millisecond, 20*time.Millisecond)))

	waitFor(t, resubscribed, server, 3*time.Second, "reconnect after rejected subscription")
	if got := server.conns.Load(); got < 2 {
		t.Fatalf("connections = %d, want at least 2", got)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestEnterChatWelcomes(t *testing.T) {
	replyCh := make(chan map[string]any, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if err := p.write(callback(cmdEventCallback, "event-1", map[string]any{
			"msgid":   "e-1",
			"msgtype": "event",
			"event":   map[string]any{"eventtype": eventEnterChat},
		})); err != nil {
			return err
		}
		reply, err := expectCommand(p, cmdRespondWelcome, 2*time.Second)
		if err != nil {
			return err
		}
		replyCh <- reply
		return nil
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url), WithHandler(testReplyHandler)))

	reply := waitFor(t, replyCh, server, 3*time.Second, "welcome reply")
	if got := stringAt(reply, "headers", "req_id"); got != "event-1" {
		t.Fatalf("welcome req_id = %q, want the event req_id", got)
	}
	if got := stringAt(reply, "body", "msgtype"); got != "text" {
		t.Fatalf("welcome msgtype = %q, want text", got)
	}
	if got := stringAt(reply, "body", "text", "content"); got != testWelcome {
		t.Fatalf("welcome content = %q, want %q", got, testWelcome)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestHeartbeatPingPong(t *testing.T) {
	replied := make(chan map[string]any, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		sub, err := expectCommand(p, cmdSubscribe, 2*time.Second)
		if err != nil {
			return err
		}
		if err := p.ack(stringAt(sub, "headers", "req_id"), 0); err != nil {
			return err
		}
		for {
			frame, err := p.read(2 * time.Second)
			if err != nil {
				return err
			}
			if stringAt(frame, "cmd") != cmdPing {
				continue
			}
			reqID := stringAt(frame, "headers", "req_id")
			if reqID == "" {
				return errors.New("ping command has no req_id")
			}
			if err := p.ack(reqID, 0); err != nil {
				return err
			}
			if err := p.write(callback(cmdMsgCallback, "after-ping", map[string]any{"msgtype": "text"})); err != nil {
				return err
			}
			reply, err := expectCommand(p, cmdRespondMsg, 2*time.Second)
			if err != nil {
				return err
			}
			replied <- reply
			return nil
		}
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url), WithHandler(testReplyHandler), WithPingInterval(20*time.Millisecond)))

	reply := waitFor(t, replied, server, 3*time.Second, "reply after heartbeat")
	if got := stringAt(reply, "headers", "req_id"); got != "after-ping" {
		t.Fatalf("reply req_id = %q, want after-ping", got)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestDisconnectedEventReconnects(t *testing.T) {
	resumed := make(chan struct{}, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if index == 1 {
			return p.write(callback(cmdEventCallback, "event-disconnected", map[string]any{
				"msgtype": "event",
				"event":   map[string]any{"eventtype": eventDisconnected},
			}))
		}
		resumed <- struct{}{}
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url), WithReconnectBounds(5*time.Millisecond, 20*time.Millisecond)))

	waitFor(t, resumed, server, 3*time.Second, "reconnect after disconnected_event")
	if got := server.conns.Load(); got < 2 {
		t.Fatalf("connections = %d, want at least 2", got)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestContextCancellationStopsClient(t *testing.T) {
	subscribed := make(chan struct{}, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		subscribed <- struct{}{}
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret, WithEndpoint(server.url)))
	waitFor(t, subscribed, server, 3*time.Second, "subscription")
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestCancelledContextNeverDials(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- New(testBotID, testSecret).Start(ctx) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run ignored an already cancelled context")
	}
}

func TestInvalidFramesDoNotStopConnection(t *testing.T) {
	replyCh := make(chan map[string]any, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if err := p.writeText("this is not json"); err != nil {
			return err
		}
		if err := p.writeText("[]"); err != nil {
			return err
		}
		if err := p.writeText(""); err != nil {
			return err
		}
		if err := p.writeBinary([]byte{0xff, 0x00, 0xfe}); err != nil {
			return err
		}
		if err := p.write(map[string]any{"cmd": "aibot_unknown_command"}); err != nil {
			return err
		}
		if err := p.write(map[string]any{"headers": map[string]any{"req_id": "no-command"}}); err != nil {
			return err
		}
		if err := p.write(callback(cmdMsgCallback, "after-invalid", map[string]any{"msgtype": "image"})); err != nil {
			return err
		}
		reply, err := expectCommand(p, cmdRespondMsg, 2*time.Second)
		if err != nil {
			return err
		}
		replyCh <- reply
		return nil
	})
	run := startClient(t, New(testBotID, testSecret, WithEndpoint(server.url), WithHandler(testReplyHandler)))

	reply := waitFor(t, replyCh, server, 3*time.Second, "reply after invalid frames")
	if got := stringAt(reply, "headers", "req_id"); got != "after-invalid" {
		t.Fatalf("reply req_id = %q, want after-invalid", got)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

func TestHeartbeatTimeoutReconnects(t *testing.T) {
	reconnected := make(chan struct{}, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if index >= 2 {
			reconnected <- struct{}{}
			return holdConnection(p)
		}
		// Accept pings without answering: the client must give up and
		// reconnect after three unanswered heartbeats.
		for {
			if _, err := p.read(2 * time.Second); err != nil {
				return nil
			}
		}
	})
	run := startClient(t, New(testBotID, testSecret,
		WithEndpoint(server.url),
		WithPingInterval(10*time.Millisecond),
		WithReconnectBounds(5*time.Millisecond, 20*time.Millisecond)))

	waitFor(t, reconnected, server, 3*time.Second, "reconnect after unanswered heartbeats")
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}

// Reply policy belongs to the caller, not the connection library.
func testReplyHandler(ctx context.Context, event Event, r *Responder) error {
	if event.Command == MessageCallback {
		return r.ReplyStream(ctx, testHelp)
	}
	if event.EventType == EnterChat {
		return r.WelcomeText(ctx, testWelcome)
	}
	return nil
}

// TestCallbackCarriesTypedModel checks that both delivery modes expose the same
// model: a message callback and an event callback both populate Event.Message,
// and an event keeps its nested payload. Before the fix the event was parsed
// with the message parser, which rejects msgtype=event, so Message stayed nil.
func TestCallbackCarriesTypedModel(t *testing.T) {
	handled := make(chan Event, 2)
	server := startWXWorkServer(t, func(_ int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if err := p.write(callback(cmdMsgCallback, "msg-1", map[string]any{
			"msgid": "m-1", "msgtype": "text", "chattype": "single",
			"from": map[string]any{"userid": "u1"},
			"text": map[string]any{"content": "hello"},
		})); err != nil {
			return err
		}
		if err := p.write(callback(cmdEventCallback, "event-1", map[string]any{
			"msgid": "e-1", "msgtype": "event", "aibotid": "A",
			"from": map[string]any{"userid": "u2"},
			"event": map[string]any{
				"eventtype":           "template_card_event",
				"template_card_event": map[string]any{"event_key": "k1", "task_id": "t1"},
			},
		})); err != nil {
			return err
		}
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret, WithEndpoint(server.url),
		WithHandler(func(_ context.Context, event Event, _ *Responder) error {
			handled <- event
			return nil
		})))

	message := waitFor(t, handled, server, 3*time.Second, "message callback")
	if message.Message == nil {
		t.Fatal("message callback has no typed message")
	}
	if message.Message.MsgType != aibot.TypeText || message.Message.Text == nil ||
		message.Message.Text.Content != "hello" {
		t.Fatalf("unexpected message model: %+v", message.Message)
	}
	if len(message.Message.Raw) == 0 {
		t.Fatal("raw message not preserved")
	}

	event := waitFor(t, handled, server, 3*time.Second, "event callback")
	if event.EventType != "template_card_event" {
		t.Fatalf("event type = %q", event.EventType)
	}
	if event.Message == nil {
		t.Fatal("event callback has no typed message")
	}
	if event.Message.Event == nil || event.Message.Event.EventType != "template_card_event" {
		t.Fatalf("event payload not parsed: %+v", event.Message)
	}
	if event.Message.From.UserID != "u2" {
		t.Fatalf("sender not parsed: %+v", event.Message.From)
	}
	var nested struct {
		EventKey string `json:"event_key"`
	}
	if err := json.Unmarshal(event.Message.Event.Raw, &nested); err != nil || nested.EventKey != "k1" {
		t.Fatalf("nested event payload: %v %+v", err, nested)
	}
	if err := run.stop(); err != nil {
		t.Fatalf("client run: %v", err)
	}
}
