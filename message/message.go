// Package message implements the WeCom application message API
// (/cgi-bin/message/*) on top of the shared authenticated client.
package message

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Message sends application messages using one shared authenticated client.
// It does not own credentials or maintain a separate token cache.
type Message struct {
	api     *client.Client
	agentID int
}

// New creates an application messaging client. A non-positive agentID
// disables direct application messages, but group chats do not require agentID.
func New(api *client.Client, agentID int) *Message {
	return &Message{api: api, agentID: agentID}
}

func (c *Message) CanSendMessages() bool { return c != nil && c.agentID > 0 }

func (c *Message) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: application client is unavailable")
	}
	return c.api.Call(ctx, op, path, request, response)
}

func required(op, field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("wecom: %s: %s is required", op, field)
	}
	return value, nil
}

func normalizeUserIDs(userIDs []string) []string {
	result := make([]string, 0, len(userIDs))
	seen := make(map[string]bool, len(userIDs))
	for _, userID := range userIDs {
		id := strings.TrimSpace(userID)
		if id != "" && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}
