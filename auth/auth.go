// Package auth implements the WeCom identity API: web authorisation links, code
// exchange, web login and second-factor verification.
package auth

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"github.com/gowsp/go-wecom/client"
)

// Auth resolves identities with one shared transport.
type Auth struct{ api *client.Client }

// New creates an identity client.
func New(api *client.Client) *Auth { return &Auth{api: api} }

var errClient = errors.New("wecom: auth client is unavailable")

func (c *Auth) get(ctx context.Context, op, path string, query url.Values, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Get(ctx, op, path, query, response)
}

func (c *Auth) call(ctx context.Context, op, path string, request, response any) error {
	if c == nil || c.api == nil {
		return errClient
	}
	return c.api.Call(ctx, op, path, request, response)
}

// OAuth scopes.
const (
	// ScopeBase authorises silently and yields only the member id.
	ScopeBase = "snsapi_base"
	// ScopeUserInfo asks the member to confirm and yields the profile ticket.
	ScopeUserInfo = "snsapi_userinfo"
	// ScopePrivateInfo exposes sensitive fields such as mobile and email.
	ScopePrivateInfo = "snsapi_privateinfo"
)

// OAuthOptions describes one authorisation link.
type OAuthOptions struct {
	CorpID      string
	RedirectURL string
	State       string
	Scope       string
	// AgentID is required when Scope is ScopePrivateInfo.
	AgentID int
}

// OAuthURL builds the authorisation URL the browser is sent to. The caller keeps
// the state and verifies it on return; this package does not store it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91022
func OAuthURL(options OAuthOptions) (string, error) {
	if options.CorpID == "" {
		return "", errors.New("wecom: oauth url: corpid is required")
	}
	if options.RedirectURL == "" {
		return "", errors.New("wecom: oauth url: redirect url is required")
	}
	parsed, err := url.Parse(options.RedirectURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("wecom: oauth url: redirect url must be absolute")
	}
	scope := options.Scope
	if scope == "" {
		scope = ScopeBase
	}
	switch scope {
	case ScopeBase, ScopeUserInfo, ScopePrivateInfo:
	default:
		return "", errors.New("wecom: oauth url: unsupported scope " + scope)
	}
	if scope == ScopePrivateInfo && options.AgentID <= 0 {
		return "", errors.New("wecom: oauth url: agentid is required for snsapi_privateinfo")
	}

	query := url.Values{
		"appid":         {options.CorpID},
		"redirect_uri":  {options.RedirectURL},
		"response_type": {"code"},
		"scope":         {scope},
	}
	if options.State != "" {
		query.Set("state", options.State)
	}
	if scope == ScopePrivateInfo {
		query.Set("agentid", strconv.Itoa(options.AgentID))
	}
	return "https://open.weixin.qq.com/connect/oauth2/authorize?" + query.Encode() + "#wechat_redirect", nil
}

// UserInfo is the identity behind one authorisation code.
type UserInfo struct {
	UserID         string `json:"userid"`
	OpenID         string `json:"openid"`
	ExternalUserID string `json:"external_userid"`
	UserTicket     string `json:"user_ticket"`
}

// GetUserInfo exchanges an authorisation code for the member identity.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91023
func (c *Auth) GetUserInfo(ctx context.Context, code string) (UserInfo, error) {
	if code == "" {
		return UserInfo{}, errors.New("wecom: get user info: code is required")
	}
	var out UserInfo
	err := c.get(ctx, "auth get user info", "/cgi-bin/auth/getuserinfo",
		url.Values{"code": {code}}, &out)
	return out, err
}

// UserDetail is the sensitive profile available with a user ticket.
type UserDetail struct {
	UserID  string `json:"userid"`
	Gender  string `json:"gender"`
	Avatar  string `json:"avatar"`
	QRCode  string `json:"qr_code"`
	Mobile  string `json:"mobile"`
	Email   string `json:"email"`
	BizMail string `json:"biz_mail"`
	Address string `json:"address"`
}

// GetUserDetail reads the sensitive profile behind a user ticket. The ticket
// comes from GetUserInfo when the link used ScopeUserInfo or ScopePrivateInfo.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95833
func (c *Auth) GetUserDetail(ctx context.Context, userTicket string) (UserDetail, error) {
	if userTicket == "" {
		return UserDetail{}, errors.New("wecom: get user detail: user_ticket is required")
	}
	var out UserDetail
	err := c.call(ctx, "auth get user detail", "/cgi-bin/auth/getuserdetail",
		map[string]string{"user_ticket": userTicket}, &out)
	return out, err
}

// LoginIdentity is the identity returned by the web login flow.
type LoginIdentity struct {
	UserID         string `json:"userid"`
	UserDocTicket  string `json:"user_doc_ticket"`
	OpenID         string `json:"openid"`
	ExternalUserID string `json:"external_userid"`
}

// GetLoginIdentity exchanges a web login code for the member identity. It uses
// the same endpoint as GetUserInfo but returns the web login fields.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98176
func (c *Auth) GetLoginIdentity(ctx context.Context, code string) (LoginIdentity, error) {
	if code == "" {
		return LoginIdentity{}, errors.New("wecom: get login identity: code is required")
	}
	var out LoginIdentity
	err := c.get(ctx, "auth get login identity", "/cgi-bin/auth/getuserinfo",
		url.Values{"code": {code}}, &out)
	return out, err
}

// TFAInfo is the second-factor challenge bound to one login code.
type TFAInfo struct {
	UserID  string `json:"userid"`
	TFACode string `json:"tfa_code"`
}

// GetTFAInfo resolves the member behind a second-factor login code.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99499
func (c *Auth) GetTFAInfo(ctx context.Context, code string) (TFAInfo, error) {
	if code == "" {
		return TFAInfo{}, errors.New("wecom: get tfa info: code is required")
	}
	var out TFAInfo
	err := c.call(ctx, "auth get tfa info", "/cgi-bin/auth/get_tfa_info",
		map[string]string{"code": code}, &out)
	return out, err
}

// UseTFACode reports that the member passed second-factor verification with the
// code returned by GetTFAInfo.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99500
func (c *Auth) UseTFACode(ctx context.Context, userID, tfaCode string) error {
	if userID == "" || tfaCode == "" {
		return errors.New("wecom: use tfa code: userid and tfa_code are required")
	}
	request := struct {
		UserID  string `json:"userid"`
		TFACode string `json:"tfa_code"`
	}{UserID: userID, TFACode: tfaCode}
	return c.call(ctx, "auth use tfa code", "/cgi-bin/user/tfa_succ", request, nil)
}

// ConfirmAuth reports that the member passed login verification.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99521
func (c *Auth) ConfirmAuth(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.New("wecom: confirm auth: userid is required")
	}
	return c.get(ctx, "auth confirm", "/cgi-bin/user/authsucc",
		url.Values{"userid": {userID}}, nil)
}
