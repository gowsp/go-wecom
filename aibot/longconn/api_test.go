package longconn

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gowsp/go-wecom/aibot"
)

func TestPermanentSubscriptionRejectionReturns(t *testing.T) {
	for _, code := range []int{40001, 40013, 6000, 999999} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := startWXWorkServer(t, func(_ int, p *peer) error {
				sub, err := expectCommand(p, cmdSubscribe, time.Second)
				if err != nil {
					return err
				}
				// The remote text must not appear in the returned error.
				return p.write(map[string]any{
					"headers": map[string]any{"req_id": stringAt(sub, "headers", "req_id")},
					"errcode": code, "errmsg": testSecret,
				})
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			client := New(testBotID, testSecret, WithEndpoint(server.url),
				WithReconnectBounds(time.Millisecond, time.Millisecond))
			err := client.Start(ctx)
			var rejection *SubscriptionError
			if !errors.As(err, &rejection) || rejection.Code != code || rejection.Temporary() {
				t.Fatalf("Start error = %v, want permanent SubscriptionError(%d)", err, code)
			}
			if err.Error() != fmt.Sprintf("wecom: subscribe rejected with errcode %d", code) {
				t.Fatalf("unsafe error: %v", err)
			}
			if got := server.conns.Load(); got != 1 {
				t.Fatalf("connections = %d, want 1", got)
			}
		})
	}
}

func TestMissingCredentialsReturnBeforeDial(t *testing.T) {
	for _, pair := range [][2]string{{"", testSecret}, {testBotID, " "}} {
		err := New(pair[0], pair[1]).Start(context.Background())
		if !errors.Is(err, ErrMissingCredentials) {
			t.Fatalf("Start = %v, want ErrMissingCredentials", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := New("", "").Start(ctx); err != nil {
		t.Fatalf("cancelled Start = %v, want nil", err)
	}
}

func TestCallbacksDefaultToNoResponse(t *testing.T) {
	for _, installHandler := range []bool{false, true} {
		t.Run(fmt.Sprint(installHandler), func(t *testing.T) {
			quiet := make(chan struct{}, 1)
			events := make(chan Event, 3)
			server := startWXWorkServer(t, func(_ int, p *peer) error {
				if err := subscribeAck(p); err != nil {
					return err
				}
				for _, cb := range []map[string]any{
					callback(cmdMsgCallback, "message", map[string]any{"msgtype": "text", "text": map[string]any{"content": "hello"}}),
					callback(cmdEventCallback, "enter", map[string]any{"event": map[string]any{"eventtype": EnterChat}}),
					callback(cmdEventCallback, "future", map[string]any{"event": map[string]any{"eventtype": "future_event"}, "custom": 42}),
				} {
					if err := p.write(cb); err != nil {
						return err
					}
				}
				_, err := p.read(75 * time.Millisecond)
				var timeout net.Error
				if !errors.As(err, &timeout) || !timeout.Timeout() {
					return fmt.Errorf("expected no response, read error = %v", err)
				}
				quiet <- struct{}{}
				return holdConnection(p)
			})
			options := []Option{WithEndpoint(server.url)}
			if installHandler {
				options = append(options, WithHandler(func(_ context.Context, e Event, _ *Responder) error {
					events <- e
					return nil
				}))
			}
			run := startClient(t, New(testBotID, testSecret, options...))
			waitFor(t, quiet, server, time.Second, "no implicit response")
			if installHandler {
				for i, want := range []string{"message", "enter", "future"} {
					event := waitFor(t, events, server, time.Second, "callback dispatch")
					if event.RequestID != want {
						t.Fatalf("event %d = %+v, want request %q", i, event, want)
					}
					if i == 2 {
						var body map[string]any
						if err := json.Unmarshal(event.Body, &body); err != nil || body["custom"] != float64(42) || event.EventType != "future_event" {
							t.Fatalf("raw event not preserved: %+v", event)
						}
					}
				}
			}
			if err := run.stop(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestHandlerErrorReturnsWithoutReconnect(t *testing.T) {
	failure := errors.New("application failure")
	server := startWXWorkServer(t, func(_ int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if err := p.write(callback(cmdMsgCallback, "fail", map[string]any{"msgtype": "text"})); err != nil {
			return err
		}
		return holdConnection(p)
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := New(testBotID, testSecret, WithEndpoint(server.url), WithHandler(func(context.Context, Event, *Responder) error {
		return failure
	})).Start(ctx)
	var handlerErr *HandlerError
	if !errors.As(err, &handlerErr) || !errors.Is(err, failure) {
		t.Fatalf("Start = %v, want wrapped handler error", err)
	}
	if got := server.conns.Load(); got != 1 {
		t.Fatalf("connections = %d, want 1", got)
	}
}

func TestResponderValidatesCallback(t *testing.T) {
	ctx := context.Background()
	// A nil connection would panic if validation attempted a write.
	r := &Responder{event: Event{Command: MessageCallback}}
	if err := r.ReplyRaw(ctx, map[string]any{}); !errors.Is(err, ErrNoRequestID) {
		t.Fatalf("Reply = %v, want ErrNoRequestID", err)
	}
	if err := r.WelcomeText(ctx, "hello"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("WelcomeText = %v, want ErrInvalidResponse", err)
	}
	r.event = Event{Command: EventCallback, RequestID: "x", EventType: "feedback_event"}
	if err := r.WelcomeText(ctx, "hello"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("WelcomeText = %v, want ErrInvalidResponse", err)
	}
	if err := r.ReplyStream(ctx, "hello"); !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("ReplyStream = %v, want ErrInvalidResponse", err)
	}
}

func TestExplicitEventResponse(t *testing.T) {
	replies := make(chan map[string]any, 1)
	server := startWXWorkServer(t, func(_ int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if err := p.write(callback(cmdEventCallback, "card-event", map[string]any{"event": map[string]any{"eventtype": "template_card_event"}})); err != nil {
			return err
		}
		reply, err := expectCommand(p, cmdRespondMsg, time.Second)
		if err != nil {
			return err
		}
		replies <- reply
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret, WithEndpoint(server.url), WithHandler(func(ctx context.Context, e Event, r *Responder) error {
		return r.Reply(ctx, aibot.TemplateCardReply(json.RawMessage(`{"card_type":"text_notice"}`)))
	})))
	reply := waitFor(t, replies, server, time.Second, "explicit event response")
	if stringAt(reply, "headers", "req_id") != "card-event" || stringAt(reply, "body", "msgtype") != "template_card" {
		t.Fatalf("response = %v", reply)
	}
	if err := run.stop(); err != nil {
		t.Fatal(err)
	}
}

func TestTransportFailureRetriesAndObserverReceivesNotices(t *testing.T) {
	notices := make(chan Diagnostic, 10)
	ready := make(chan struct{}, 1)
	server := startWXWorkServer(t, func(index int, p *peer) error {
		if err := subscribeAck(p); err != nil {
			return err
		}
		if index == 1 {
			return nil // transport closes after successful subscription
		}
		if err := p.write(map[string]any{"errcode": 123, "errmsg": testSecret}); err != nil {
			return err
		}
		ready <- struct{}{}
		return holdConnection(p)
	})
	run := startClient(t, New(testBotID, testSecret, WithEndpoint(server.url),
		WithReconnectBounds(time.Millisecond, time.Millisecond), WithObserver(func(d Diagnostic) { notices <- d })))
	waitFor(t, ready, server, time.Second, "transport reconnect")
	want := []DiagnosticKind{DiagnosticConnected, DiagnosticRetry, DiagnosticConnected, DiagnosticRejected}
	for _, kind := range want {
		notice := waitFor(t, notices, server, time.Second, "diagnostic")
		if notice.Kind != kind {
			t.Fatalf("notice = %+v, want kind %s", notice, kind)
		}
		if kind == DiagnosticRetry && (notice.Err == nil || notice.RetryAfter != time.Millisecond) {
			t.Fatalf("retry notice = %+v", notice)
		}
		if kind == DiagnosticRejected && notice.Code != 123 {
			t.Fatalf("rejected notice = %+v", notice)
		}
	}
	if err := run.stop(); err != nil {
		t.Fatal(err)
	}
}

func TestCancellationDuringSubscriptionAndBackoff(t *testing.T) {
	for _, reject := range []bool{false, true} {
		t.Run(fmt.Sprint(reject), func(t *testing.T) {
			ready := make(chan struct{}, 1)
			server := startWXWorkServer(t, func(_ int, p *peer) error {
				sub, err := expectCommand(p, cmdSubscribe, time.Second)
				if err != nil {
					return err
				}
				if reject {
					if err := p.ack(stringAt(sub, "headers", "req_id"), -1); err != nil {
						return err
					}
				} else {
					ready <- struct{}{}
				}
				return holdConnection(p)
			})
			options := []Option{WithEndpoint(server.url), WithReconnectBounds(time.Hour, time.Hour)}
			if reject {
				options = append(options, WithObserver(func(d Diagnostic) {
					if d.Kind == DiagnosticRetry {
						ready <- struct{}{}
					}
				}))
			}
			run := startClient(t, New(testBotID, testSecret, options...))
			waitFor(t, ready, server, time.Second, "subscription or backoff")
			if err := run.stop(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
