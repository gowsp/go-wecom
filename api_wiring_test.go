package wecom_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gowsp/go-wecom"
	"github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/invoice"
)

// TestAPINewWiresEveryField guards against a half-wired aggregate: a nil field
// panics on first use, so every declared pointer field must be non-nil.
func TestAPINewWiresEveryField(t *testing.T) {
	api := wecom.New("corp", "secret", 1000002)
	value := reflect.ValueOf(api).Elem()
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := value.Field(i)
		if field.Kind() != reflect.Ptr {
			continue
		}
		if field.IsNil() {
			t.Errorf("wecom.New left API.%s nil", typ.Field(i).Name)
		}
	}
}

// TestAPINewIsUsable calls one method per package through the aggregate against a
// local stub, so a nil dependency or a wrong constructor argument fails loudly.
// The default transport would reach the real endpoint, so the base URL is
// overridden here.
func TestAPINewIsUsable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0,"chains":[],"ip_list":[]}`))
	}))
	t.Cleanup(server.Close)

	api := wecom.New("corp", "secret", 1000002, client.WithBaseURL(server.URL))
	ctx := context.Background()

	checks := []struct {
		name string
		call func() error
	}{
		{"schedule", func() error { _, err := api.Schedule.GetCalendars(ctx, []string{"c"}); return err }},
		{"live", func() error { _, err := api.Live.Info(ctx, "l1"); return err }},
		{"security", func() error { _, err := api.Security.Domains(ctx); return err }},
		{"wedrive", func() error { _, err := api.WeDrive.ListFiles(ctx, "s1", "s1", 1, 0, 1); return err }},
		{"checkin", func() error { _, err := api.Checkin.Options(ctx); return err }},
		{"approval", func() error { _, err := api.Approval.VacationConfig(ctx); return err }},
		{"advanced", func() error { _, err := api.Advanced.GetApproval(ctx, "a1"); return err }},
		{"chatdata", func() error { _, err := api.ChatData.ArchiveUsers(ctx, "", 1); return err }},
		{"corpgroup", func() error { _, err := api.CorpGroup.Chains(ctx); return err }},
		{"externalpay", func() error { _, err := api.ExternalPay.Merchant(ctx, "m1"); return err }},
		{"idconvert", func() error { _, err := api.IDConvert.CorpID(ctx, "open", 1); return err }},
		{"invoice", func() error {
			_, err := api.Invoice.Get(ctx, invoice.Code{CardID: "c", EncryptCode: "e"})
			return err
		}},
		{"iprange", func() error { _, err := api.IPRange.APIIPs(ctx); return err }},
		{"miniapppay", func() error { _, err := api.MiniAppPay.Order(ctx, "m1", "o1"); return err }},
		{"msgaudit", func() error { _, err := api.MsgAudit.PermitUsers(ctx, 1); return err }},
	}
	for _, check := range checks {
		if err := check.call(); err != nil {
			t.Errorf("%s: %v", check.name, err)
		}
	}
}
