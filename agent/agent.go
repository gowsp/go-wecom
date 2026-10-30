// Package agent implements the WeCom application, menu and workbench API
// (/cgi-bin/agent, /cgi-bin/menu) on top of the shared authenticated client.
package agent

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/gowsp/go-wecom/client"
)

// Agent manages applications and their menus with one shared transport.
type Agent struct{ api *client.Client }

// New creates an application client.
func New(api *client.Client) *Agent { return &Agent{api: api} }

var errClient = errors.New("wecom: agent client is unavailable")

// get performs one GET request with the shared token.
func (c *Agent) get(ctx context.Context, op, path string, query url.Values, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Get(ctx, op, path, query, response)
}

// call performs one POST request.
func (c *Agent) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Call(ctx, op, path, request, response)
}

func itoa(v int) string { return strconv.Itoa(v) }
