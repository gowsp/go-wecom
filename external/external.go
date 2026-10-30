// Package external implements the WeCom customer contact API
// (/cgi-bin/externalcontact): customers, tags, customer group chats and
// "contact me" ways. Business rules stay in the application.
package external

import (
	"context"
	"errors"
	"net/url"

	"github.com/gowsp/go-wecom/client"
)

// External manages customer contacts with one shared transport.
type External struct{ api *client.Client }

// New creates a customer contact client.
func New(api *client.Client) *External { return &External{api: api} }

var errClient = errors.New("wecom: external contact client is unavailable")

func (c *External) get(ctx context.Context, op, path string, query url.Values, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Get(ctx, op, path, query, response)
}

func (c *External) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Call(ctx, op, path, request, response)
}
