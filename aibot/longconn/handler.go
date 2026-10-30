package longconn

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gobwas/ws"
	"github.com/gowsp/go-wecom/aibot"
)

// Callback commands and the currently recognized lifecycle event types.
const (
	MessageCallback = cmdMsgCallback
	EventCallback   = cmdEventCallback
	EnterChat       = eventEnterChat
	Disconnected    = eventDisconnected
)

// Event is an inbound message or event callback. Message carries the same typed
// payload the HTTP mode parses, so both delivery modes expose one model: a
// message callback is decoded with aibot.ParseMessage and an event callback with
// aibot.ParseEvent, whose Event.Raw keeps the nested payload. Body is kept for
// fields this package does not model yet. RequestID may be empty, in which case
// the event cannot be answered.
type Event struct {
	Command   string
	RequestID string
	EventType string
	Message   *aibot.Message
	Body      json.RawMessage
}

// Handler handles callbacks serially, in arrival order. It must return promptly
// and respect ctx so it does not delay heartbeats or shutdown. Returning nil
// without using responder deliberately sends no response. A non-nil error stops
// Start with a HandlerError, except responder transport errors, which reconnect.
// The responder is valid only during this call.
type Handler func(ctx context.Context, event Event, responder *Responder) error

// Responder sends explicit responses correlated to one callback. Reply and
// Welcome accept JSON-marshalable protocol bodies, including future types.
// Sending successfully means the frame was written, not acknowledged by WeCom.
type Responder struct {
	bot   *Bot
	conn  *connection
	event Event
}

// ErrNoRequestID means the callback did not contain a response correlation ID.
var ErrNoRequestID = errors.New("wecom: callback has no request ID")

// ErrInvalidResponse means the response command is not valid for this callback.
var ErrInvalidResponse = errors.New("wecom: response is not valid for this callback")

// Reply sends one typed reply to a message or event callback. Use ReplyRaw for
// protocol bodies this package does not model yet. Welcome uses its own command.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (r *Responder) Reply(ctx context.Context, reply aibot.Reply) error {
	if err := reply.Validate(); err != nil {
		return err
	}
	return r.ReplyRaw(ctx, reply)
}

// ReplyRaw sends an arbitrary JSON-marshalable protocol body.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (r *Responder) ReplyRaw(ctx context.Context, body any) error {
	if r.event.Command != MessageCallback && r.event.Command != EventCallback {
		return ErrInvalidResponse
	}
	return r.send(ctx, cmdRespondMsg, body)
}

// Welcome sends a welcome response. It is valid only for enter_chat events.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (r *Responder) Welcome(ctx context.Context, body any) error {
	if r.event.Command != EventCallback || r.event.EventType != EnterChat {
		return ErrInvalidResponse
	}
	return r.send(ctx, cmdRespondWelcome, body)
}

// ReplyStream sends one finished stream response to a message callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (r *Responder) ReplyStream(ctx context.Context, content string) error {
	if r.event.Command != MessageCallback {
		return ErrInvalidResponse
	}
	reply := aibot.StreamReply(content, true)
	reply.Stream.ID = newStreamID()
	return r.Reply(ctx, reply)
}

// WelcomeText sends a text welcome to an enter_chat callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101463
func (r *Responder) WelcomeText(ctx context.Context, content string) error {
	return r.Welcome(ctx, map[string]any{
		"msgtype": "text", "text": map[string]any{"content": content},
	})
}

func (r *Responder) send(ctx context.Context, command string, body any) error {
	if r.event.RequestID == "" {
		return ErrNoRequestID
	}
	return r.bot.send(ctx, r.conn, command, r.event.RequestID, body)
}

// handleFrame ignores malformed frames and never exposes their contents through
// diagnostics. Lifecycle events still reach the handler before reconnection.
func (c *Bot) handleFrame(ctx context.Context, conn *connection, frame inbound) error {
	if frame.op != ws.OpText && frame.op != ws.OpBinary {
		return nil
	}
	if len(frame.data) == 0 || len(frame.data) > maxFrameSize {
		return nil
	}
	var msg envelope
	if json.Unmarshal(frame.data, &msg) != nil {
		return nil
	}
	if msg.Cmd != cmdMsgCallback && msg.Cmd != cmdEventCallback {
		if msg.Errcode != nil && *msg.Errcode != 0 {
			c.observe(Diagnostic{Kind: DiagnosticRejected, Code: *msg.Errcode})
		}
		return nil
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(msg.Body, &body) != nil || body == nil {
		return nil
	}
	event := Event{Command: msg.Cmd, RequestID: msg.Headers.ReqID, Body: msg.Body}
	// Each command has its own parser, so both delivery modes hand the caller
	// the same typed model. ParseEvent also carries the nested payload of the
	// event, which a generic message parse would reject.
	if msg.Cmd == cmdEventCallback {
		if parsed, err := aibot.ParseEvent(msg.Body); err == nil {
			event.Message = &parsed
			event.EventType = parsed.Event.EventType
		}
	}
	if event.EventType == "" {
		var value eventBody
		if json.Unmarshal(msg.Body, &value) == nil {
			event.EventType = value.Event.EventType
		}
	}
	if event.Message == nil {
		if message, err := aibot.ParseMessage(msg.Body); err == nil {
			event.Message = &message
		}
	}
	if c.handler != nil {
		if err := c.handler(ctx, event, &Responder{bot: c, conn: conn, event: event}); err != nil {
			var writeErr *responseWriteError
			if errors.As(err, &writeErr) {
				return err
			}
			return &HandlerError{Err: err}
		}
	}
	if event.Command == EventCallback && event.EventType == Disconnected {
		return errDisconnected
	}
	return nil
}
