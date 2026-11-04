package kf

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// ErrTextTooLong means the strict text API would exceed WeCom's 2048-byte limit.
var ErrTextTooLong = errors.New("customerservice: text exceeds 2048 bytes")

// SendTextRequest sends text without silently discarding any content. MsgID is
// optional and must be unique within the account; supplying it does not make
// automatic retries safe (WeCom can reject duplicate IDs).
type SendTextRequest struct {
	OpenKfID       string
	ExternalUserID string
	Content        string
	MsgID          string
}

// SendResult means WeCom accepted a send request, not confirmed delivery.
// Applications must also handle the asynchronous message-send-failure event.
type SendResult struct {
	MsgID string `json:"msgid"`
}

// SendTextWithResult validates the complete text and returns its message ID.
// Unlike SendText, it never truncates. See
// https://developer.work.weixin.qq.com/document/path/94677.
func (c *KF) SendTextWithResult(ctx context.Context, request SendTextRequest) (SendResult, error) {
	var out SendResult
	if len(request.Content) > maxTextBytes {
		return out, ErrTextTooLong
	}
	if strings.TrimSpace(request.OpenKfID) == "" || strings.TrimSpace(request.ExternalUserID) == "" || strings.TrimSpace(request.Content) == "" || !utf8.ValidString(request.Content) {
		return out, errors.New("customerservice: valid account, customer and text are required")
	}
	if len(request.MsgID) > 32 {
		return out, errors.New("customerservice: invalid message id")
	}
	for _, r := range request.MsgID {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
			return out, errors.New("customerservice: invalid message id")
		}
	}
	body := map[string]any{"open_kfid": request.OpenKfID, "touser": request.ExternalUserID, "msgtype": "text", "text": map[string]string{"content": request.Content}}
	if request.MsgID != "" {
		body["msgid"] = request.MsgID
	}
	err := c.api.Call(ctx, "customer service send_msg", "/cgi-bin/kf/send_msg", body, &out)
	return out, err
}
