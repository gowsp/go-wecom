package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

type recorder struct {
	path  string
	query string
	body  map[string]any
}

func newAuth(t *testing.T, response string) (*recorder, *Auth) {
	t.Helper()
	rec := &recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		rec.path, rec.query = r.URL.Path, r.URL.RawQuery
		_ = json.NewDecoder(r.Body).Decode(&rec.body)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return rec, New(client.New("corp", "secret", client.WithBaseURL(server.URL)))
}

func TestOAuthURL(t *testing.T) {
	link, err := OAuthURL(OAuthOptions{
		CorpID:      "corp1",
		RedirectURL: "https://app.example.com/cb?x=1",
		State:       "s1",
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "open.weixin.qq.com" || parsed.Fragment != "wechat_redirect" {
		t.Fatalf("unexpected link: %s", link)
	}
	query := parsed.Query()
	if query.Get("appid") != "corp1" || query.Get("scope") != ScopeBase ||
		query.Get("state") != "s1" || query.Get("response_type") != "code" {
		t.Fatalf("unexpected query: %v", query)
	}
	if query.Get("redirect_uri") != "https://app.example.com/cb?x=1" {
		t.Fatalf("redirect uri not preserved: %s", query.Get("redirect_uri"))
	}
}

func TestOAuthURLValidation(t *testing.T) {
	cases := map[string]OAuthOptions{
		"no corpid":   {RedirectURL: "https://a/b"},
		"no redirect": {CorpID: "c"},
		"relative":    {CorpID: "c", RedirectURL: "/cb"},
		"bad scope":   {CorpID: "c", RedirectURL: "https://a/b", Scope: "snsapi_other"},
		"privateinfo": {CorpID: "c", RedirectURL: "https://a/b", Scope: ScopePrivateInfo},
	}
	for name, options := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := OAuthURL(options); err == nil {
				t.Fatal("invalid options accepted")
			}
		})
	}
	link, err := OAuthURL(OAuthOptions{CorpID: "c", RedirectURL: "https://a/b",
		Scope: ScopePrivateInfo, AgentID: 1000002})
	if err != nil || !strings.Contains(link, "agentid=1000002") {
		t.Fatalf("privateinfo link: %s %v", link, err)
	}
}

func TestGetUserInfoAndDetail(t *testing.T) {
	rec, api := newAuth(t, `{"errcode":0,"userid":"u1","user_ticket":"ticket"}`)
	info, err := api.GetUserInfo(context.Background(), "code1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/auth/getuserinfo" || !strings.Contains(rec.query, "code=code1") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if info.UserID != "u1" || info.UserTicket != "ticket" {
		t.Fatalf("unexpected info: %+v", info)
	}

	rec, api = newAuth(t, `{"errcode":0,"userid":"u1","mobile":"138","email":"a@b.c"}`)
	detail, err := api.GetUserDetail(context.Background(), "ticket")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/auth/getuserdetail" || rec.body["user_ticket"] != "ticket" {
		t.Fatalf("unexpected detail request %s %v", rec.path, rec.body)
	}
	if detail.Mobile != "138" || detail.Email != "a@b.c" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestLoginIdentity(t *testing.T) {
	rec, api := newAuth(t, `{"errcode":0,"userid":"u1","user_doc_ticket":"doc","openid":"o1"}`)
	identity, err := api.GetLoginIdentity(context.Background(), "code2")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/auth/getuserinfo" || identity.UserDocTicket != "doc" ||
		identity.OpenID != "o1" {
		t.Fatalf("unexpected identity: %s %+v", rec.path, identity)
	}
}

func TestTFAFlow(t *testing.T) {
	rec, api := newAuth(t, `{"errcode":0,"userid":"u1","tfa_code":"tfa"}`)
	ctx := context.Background()

	info, err := api.GetTFAInfo(ctx, "code3")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/auth/get_tfa_info" || rec.body["code"] != "code3" {
		t.Fatalf("unexpected tfa info request %s %v", rec.path, rec.body)
	}
	if info.UserID != "u1" || info.TFACode != "tfa" {
		t.Fatalf("unexpected tfa info: %+v", info)
	}

	if err := api.UseTFACode(ctx, "u1", "tfa"); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/user/tfa_succ" || rec.body["tfa_code"] != "tfa" {
		t.Fatalf("unexpected tfa succ %s %v", rec.path, rec.body)
	}

	if err := api.ConfirmAuth(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/user/authsucc" || !strings.Contains(rec.query, "userid=u1") {
		t.Fatalf("unexpected confirm %s?%s", rec.path, rec.query)
	}
}

func TestAuthValidation(t *testing.T) {
	_, api := newAuth(t, `{"errcode":0}`)
	ctx := context.Background()
	if _, err := api.GetUserInfo(ctx, ""); err == nil {
		t.Fatal("empty code accepted")
	}
	if _, err := api.GetUserDetail(ctx, ""); err == nil {
		t.Fatal("empty ticket accepted")
	}
	if _, err := api.GetLoginIdentity(ctx, ""); err == nil {
		t.Fatal("empty login code accepted")
	}
	if _, err := api.GetTFAInfo(ctx, ""); err == nil {
		t.Fatal("empty tfa code accepted")
	}
	if err := api.UseTFACode(ctx, "u1", ""); err == nil {
		t.Fatal("empty tfa code accepted")
	}
	if err := api.ConfirmAuth(ctx, ""); err == nil {
		t.Fatal("empty userid accepted")
	}
}
