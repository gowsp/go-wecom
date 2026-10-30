// Package contact implements the WeCom contact API (/cgi-bin/user,
// /cgi-bin/batch, /cgi-bin/corp) on top of the shared authenticated client.
// Members, departments and tags live here; business rules stay in the caller.
package contact

import (
	"context"
	"errors"
	"strconv"

	"net/url"

	"github.com/gowsp/go-wecom/client"
)

// Contact manages members, departments and tags with one shared transport.
type Contact struct{ api *client.Client }

// New creates a contact client.
func New(api *client.Client) *Contact { return &Contact{api: api} }

var errClient = errors.New("wecom: contact client is unavailable")

func (c *Contact) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Call(ctx, op, path, request, response)
}

func (c *Contact) get(ctx context.Context, op, path string, query url.Values, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Get(ctx, op, path, query, response)
}

// itoa renders an int for a query parameter.
func itoa(v int) string { return strconv.Itoa(v) }

// boolString renders 1/0 because the contact API does not accept Go booleans.
func boolString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
