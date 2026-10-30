// Package mail implements the WeCom mailbox API (/cgi-bin/exmail/*) on top of
// the shared authenticated client.
package mail

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Mail provides mailbox operations using one shared authenticated client.
type Mail struct{ api *client.Client }

// New creates a mailbox client.
func New(api *client.Client) *Mail { return &Mail{api: api} }

func (c *Mail) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: mail client is unavailable")
	}
	// Mailbox POST endpoints that take no body still expect an empty object.
	if request == nil {
		request = struct{}{}
	}
	return c.api.Call(ctx, op, path, request, response)
}

func (c *Mail) get(ctx context.Context, op, path string, query map[string]string, response any) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: mail client is unavailable")
	}
	values := make(map[string][]string, len(query))
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			values[key] = []string{value}
		}
	}
	return c.api.Get(ctx, op, path, values, response)
}

func required(op, field, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("wecom: %s: %s is required", op, field)
	}
	return value, nil
}

func requireIDs[T any](op, field string, values []T) ([]T, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("wecom: %s: %s is required", op, field)
	}
	return values, nil
}

// List carries the {"list": [...]} envelope used by group and mailbox payloads.
type List[T any] struct {
	List []T `json:"list"`
}

func newList[T any](values []T) List[T] { return List[T]{List: values} }
