package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom"
)

func TestIPRange(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"ip_list":["1.2.3.4","5.6.7.8"]}`)
	svc := wecom.NewIPRange(api)
	ctx := context.Background()

	segments, err := svc.APIIPs(ctx)
	if err != nil || len(segments) != 2 || rec.path != "/cgi-bin/get_api_domain_ip" {
		t.Fatalf("api ips: %v %v %s", err, segments, rec.path)
	}
	if _, err := svc.CallbackIPs(ctx); err != nil || rec.path != "/cgi-bin/getcallbackip" {
		t.Fatalf("callback ips: %v %s", err, rec.path)
	}
}
