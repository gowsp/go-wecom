package wecom

import (
	"context"
	"net/url"

	"github.com/gowsp/go-wecom/client"
)

// IPRange reports the WeCom server IP segments that applications and callbacks
// use, so firewalls can be configured.
type IPRange struct{ api *client.Client }

// NewIPRange creates an IP range reader on top of the shared transport.
func NewIPRange(api *client.Client) *IPRange { return &IPRange{api: api} }

// APIIPs returns the IP segments of qyapi.weixin.qq.com. WeCom recommends
// refreshing them daily.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92520
func (r *IPRange) APIIPs(ctx context.Context) ([]string, error) {
	return r.list(ctx, "ip api segments", "/cgi-bin/get_api_domain_ip")
}

// CallbackIPs returns the IP segments WeCom uses when calling a corp callback
// URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92521
func (r *IPRange) CallbackIPs(ctx context.Context) ([]string, error) {
	return r.list(ctx, "ip callback segments", "/cgi-bin/getcallbackip")
}

func (r *IPRange) list(ctx context.Context, op, path string) ([]string, error) {
	var out struct {
		List []string `json:"ip_list"`
	}
	if err := r.api.Get(ctx, op, path, url.Values{}, &out); err != nil {
		return nil, err
	}
	return out.List, nil
}
