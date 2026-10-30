// Package wedoc implements WeCom documents, smart sheets, and their permissions.
package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// maxListPageSize is the largest offset/limit page WeCom accepts.
const maxListPageSize = 1000

// WeDoc uses an injected shared transport for document operations.
type WeDoc struct {
	api *client.Client
}

// New creates a document client sharing the supplied token cache.
func New(api *client.Client) *WeDoc {
	return &WeDoc{api: api}
}

func (c *WeDoc) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: document client is unavailable")
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
