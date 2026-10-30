package longconn

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	cmdSubscribe      = "aibot_subscribe"
	cmdMsgCallback    = "aibot_msg_callback"
	cmdEventCallback  = "aibot_event_callback"
	cmdRespondMsg     = "aibot_respond_msg"
	cmdRespondWelcome = "aibot_respond_welcome_msg"
	cmdPing           = "ping"

	eventEnterChat    = "enter_chat"
	eventDisconnected = "disconnected_event"
	// maxFrameSize bounds a single inbound frame. Callbacks are small JSON
	// documents; anything larger is dropped without logging its content.
	maxFrameSize = 1 << 20
)

var (
	// errDisconnected reports that the server retired this connection because
	// another client subscribed for the same bot. Start reconnects.
	errDisconnected = errors.New("wecom: connection superseded by another client")
)

// WXWork message/event envelope. Errcode is a pointer because zero is a valid
// value and an absent field means "not a command acknowledgement".
type envelope struct {
	Cmd     string          `json:"cmd,omitempty"`
	Headers envelopeHeaders `json:"headers"`
	Body    json.RawMessage `json:"body,omitempty"`
	Errcode *int            `json:"errcode,omitempty"`
	Errmsg  string          `json:"errmsg,omitempty"`
}

type envelopeHeaders struct {
	ReqID string `json:"req_id"`
}

type outbound struct {
	Cmd     string          `json:"cmd"`
	Headers envelopeHeaders `json:"headers"`
	Body    any             `json:"body,omitempty"`
}

type subscribeBody struct {
	BotID  string `json:"bot_id"`
	Secret string `json:"secret"`
}

type eventBody struct {
	Event struct {
		EventType string `json:"eventtype"`
	} `json:"event"`
}

func (c *Bot) sendPing(ctx context.Context, conn *connection) error {
	return c.send(ctx, conn, cmdPing, newReqID(), nil)
}

// send writes one JSON text frame. The payload is built completely before the
// write lock is taken so a frame can never be split by another writer.
func (c *Bot) send(ctx context.Context, conn *connection, cmd, reqID string, body any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(outbound{
		Cmd:     cmd,
		Headers: envelopeHeaders{ReqID: reqID},
		Body:    body,
	})
	if err != nil {
		return errors.New("wecom: encode frame failed")
	}
	if err := conn.writeText(payload); err != nil {
		return &responseWriteError{err}
	}
	return nil
}

// newReqID returns a random request id used to correlate acknowledgements.
func newReqID() string {
	return randomHex(16)
}

// newStreamID returns a random stream id for one-shot replies.
func newStreamID() string {
	return randomHex(12)
}

func randomHex(size int) string {
	buffer := make([]byte, size)
	if _, err := cryptorand.Read(buffer); err != nil {
		// A failing system random source is unrecoverable for correlation
		// purposes; fall back to a time-based value rather than crashing.
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(buffer)
}
