// Package client provides the shared monolithic WeCom transport: one
// concurrency-safe access-token cache plus JSON, query-string, multipart upload
// and binary download helpers used by every subpackage.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	// DefaultBaseURL is the production WeCom API origin. Tests and regional
	// deployments override it.
	DefaultBaseURL = "https://qyapi.weixin.qq.com"

	defaultHTTPTimeout = 5 * time.Second
	// tokenRefreshMargin refreshes well before WeCom expires the token so a
	// request never races the expiry.
	tokenRefreshMargin         = 5 * time.Minute
	maxResponseBytes           = 1 << 20
	defaultDownloadLimit int64 = 32 << 20

	// maxErrmsgRunes bounds the server message echoed in an APIError.
	maxErrmsgRunes = 160

	// errcodeInvalidToken and errcodeExpiredToken are the two WeCom error
	// codes that mean "fetch a new access_token and try again".
	errcodeInvalidToken = 40014
	errcodeExpiredToken = 42001
)

// ErrResponseTooLarge means a response exceeded its size limit. No partial
// payload is returned. JSON responses are limited to 1 MiB; downloads default
// to 32 MiB and can be configured with WithDownloadLimit.
var ErrResponseTooLarge = errors.New("wecom: response exceeds size limit")

// APIError reports a non-zero WeCom errcode. Errmsg is the server message,
// collapsed and truncated; it is a human hint only. The client never adds the
// access_token, the corp secret, message content, or cell values.
type APIError struct {
	// Op names the failed operation, for example "message send" or
	// "add records".
	Op string
	// Errcode is the WeCom error code from the response.
	Errcode int
	// Errmsg is the bounded server message, with the corp secret redacted.
	Errmsg string
}

func (e *APIError) Error() string {
	if e.Errmsg == "" {
		return fmt.Sprintf("wecom: %s failed with errcode %d", e.Op, e.Errcode)
	}
	return fmt.Sprintf("wecom: %s failed with errcode %d: %s", e.Op, e.Errcode, e.Errmsg)
}

// Client is the shared WeCom transport with a concurrency-safe token cache.
type Client struct {
	corpID     string
	corpSecret string

	baseURL       string
	httpClient    *http.Client
	downloadLimit int64
	responseLimit int64
	now           func() time.Time

	mu        sync.Mutex // guards token and expiresAt
	token     string
	expiresAt time.Time

	refreshMu sync.Mutex // serializes token refreshes
}

// Option customizes the shared client. Options exist for tests and for
// deployments with a proxy or a regional WeCom origin.
type Option func(*Client)

// WithBaseURL overrides the API origin (for example an httptest server).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/"); trimmed != "" {
			c.baseURL = trimmed
		}
	}
}

// WithHTTPClient injects the HTTP client, which is what tests use to inspect
// requests or force failures. A nil client leaves the default in place.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}

// WithResponseLimit raises the JSON response limit for bulk endpoints that
// legitimately return more than the 1 MiB default, for example a batch customer
// query or a large record page. Positive values override the default;
// non-positive values leave it unchanged.
func WithResponseLimit(bytes int64) Option {
	return func(c *Client) {
		if bytes > 0 && bytes < 1<<63-1 {
			c.responseLimit = bytes
		}
	}
}

// WithDownloadLimit sets the maximum buffered download size in bytes. Positive
// values override the 32 MiB default; non-positive values leave it unchanged.
// It does not change the JSON response limit (see WithResponseLimit).
func WithDownloadLimit(bytes int64) Option {
	return func(c *Client) {
		if bytes > 0 && bytes < 1<<63-1 {
			c.downloadLimit = bytes
		}
	}
}

// New creates a shared authenticated HTTP client. Subpackage clients receive
// the same instance so they reuse one concurrency-safe token cache.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91039
func New(corpID, corpSecret string, options ...Option) *Client {
	client := &Client{
		corpID:        strings.TrimSpace(corpID),
		corpSecret:    strings.TrimSpace(corpSecret),
		baseURL:       DefaultBaseURL,
		httpClient:    &http.Client{Timeout: defaultHTTPTimeout},
		downloadLimit: defaultDownloadLimit,
		responseLimit: maxResponseBytes,
		now:           time.Now,
	}
	for _, option := range options {
		if option != nil {
			option(client)
		}
	}
	return client
}

// BaseURL exposes the configured origin (without a trailing slash) so callers
// can log the target without leaking credentials.
func (c *Client) BaseURL() string { return c.baseURL }

// Token returns a cached access_token, fetching and refreshing it at most once
// concurrently. A refresh happens five minutes before expiry; a token that
// lives less than five minutes is cached for half its lifetime instead of being
// fetched on every call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91039
func (c *Client) Token(ctx context.Context) (string, error) {
	if token := c.cachedToken(); token != "" {
		return token, nil
	}
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()
	// Another goroutine may have refreshed while this one waited.
	if token := c.cachedToken(); token != "" {
		return token, nil
	}
	token, expiresAt, err := c.fetchToken(ctx)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.token = token
	c.expiresAt = expiresAt
	c.mu.Unlock()
	return token, nil
}

func (c *Client) cachedToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && c.now().Before(c.expiresAt) {
		return c.token
	}
	return ""
}

// invalidate drops the cached token only if it is still the one that failed. A
// concurrent refresh may already have stored a newer token, which must survive.
func (c *Client) invalidate(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == token {
		c.token = ""
		c.expiresAt = time.Time{}
	}
}

func (c *Client) fetchToken(ctx context.Context) (string, time.Time, error) {
	endpoint := c.baseURL + "/cgi-bin/gettoken?corpid=" + url.QueryEscape(c.corpID) +
		"&corpsecret=" + url.QueryEscape(c.corpSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", time.Time{}, errors.New("wecom: create token request failed")
	}
	body, err := c.do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	var result struct {
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", time.Time{}, errors.New("wecom: invalid token response")
	}
	if result.Errcode != 0 {
		return "", time.Time{}, c.newAPIError("token request", result.Errcode, result.Errmsg)
	}
	if result.AccessToken == "" || result.ExpiresIn <= 0 {
		return "", time.Time{}, errors.New("wecom: token response is incomplete")
	}
	lifetime := time.Duration(result.ExpiresIn) * time.Second
	refreshAfter := lifetime - tokenRefreshMargin
	if refreshAfter <= 0 {
		// Short-lived tokens still cache for half their lifetime instead of
		// refreshing on every call.
		refreshAfter = lifetime / 2
	}
	return result.AccessToken, c.now().Add(refreshAfter), nil
}

// Call sends one JSON request and retries exactly once after a token-invalid
// or token-expired response. The body is encoded before the first attempt so
// both attempts send identical bytes.
func (c *Client) Call(ctx context.Context, op, path string, request, response any) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("wecom: encode %s request failed", op)
	}
	token, err := c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.postJSON(ctx, op, path, token, body, response)
	if !isTokenInvalid(err) {
		return err
	}
	c.invalidate(token)
	token, err = c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.postJSON(ctx, op, path, token, body, response)
	if isTokenInvalid(err) {
		c.invalidate(token)
	}
	return err
}

// Get sends one query-string request and retries exactly once after a
// token-invalid or token-expired response, mirroring Call.
func (c *Client) Get(ctx context.Context, op, path string, query url.Values, response any) error {
	token, err := c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.getJSON(ctx, op, path, query, token, response)
	if !isTokenInvalid(err) {
		return err
	}
	c.invalidate(token)
	token, err = c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.getJSON(ctx, op, path, query, token, response)
	if isTokenInvalid(err) {
		c.invalidate(token)
	}
	return err
}

func (c *Client) getJSON(ctx context.Context, op, path string, query url.Values, token string, response any) error {
	values := url.Values{}
	for key, entries := range query {
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}
	values.Set("access_token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+values.Encode(), nil)
	if err != nil {
		return fmt.Errorf("wecom: create %s request failed", op)
	}
	data, err := c.do(req)
	if err != nil {
		return err
	}
	return c.decode(op, data, response)
}

func (c *Client) postJSON(ctx context.Context, op, path, token string, body []byte, response any) error {
	endpoint := c.baseURL + path + "?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("wecom: create %s request failed", op)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	data, err := c.do(req)
	if err != nil {
		return err
	}
	return c.decode(op, data, response)
}

// decode applies the WeCom error envelope and unmarshals a successful body.
func (c *Client) decode(op string, data []byte, response any) error {
	var envelope struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("wecom: invalid %s response", op)
	}
	if envelope.Errcode != 0 {
		return c.newAPIError(op, envelope.Errcode, envelope.Errmsg)
	}
	if response == nil {
		return nil
	}
	if err := json.Unmarshal(data, response); err != nil {
		return fmt.Errorf("wecom: invalid %s response", op)
	}
	return nil
}

// newAPIError builds an error with a bounded, secret-free errmsg summary.
func (c *Client) newAPIError(op string, errcode int, errmsg string) *APIError {
	return &APIError{Op: op, Errcode: errcode, Errmsg: c.sanitizeErrmsg(errmsg)}
}

// sanitizeErrmsg collapses whitespace, redacts the corp secret, and bounds the
// length. Request content and access tokens are never passed here.
func (c *Client) sanitizeErrmsg(errmsg string) string {
	summary := strings.Join(strings.Fields(errmsg), " ")
	if secret := strings.TrimSpace(c.corpSecret); secret != "" {
		summary = strings.ReplaceAll(summary, secret, "[redacted]")
	}
	runes := []rune(summary)
	if len(runes) > maxErrmsgRunes {
		summary = string(runes[:maxErrmsgRunes]) + "..."
	}
	return summary
}

// isTokenInvalid reports whether err is a WeCom error that a fresh token can
// fix. Only the two documented codes trigger a retry.
func isTokenInvalid(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Errcode == errcodeInvalidToken || apiErr.Errcode == errcodeExpiredToken
}

// do executes one request and returns the bounded response body. Request URLs
// carry the corp secret or the access token, so transport failures are reported
// with a stable message instead of the underlying url.Error.
func (c *Client) do(req *http.Request) ([]byte, error) {
	return c.doLimited(req, c.responseLimit)
}

func (c *Client) doLimited(req *http.Request, limit int64) ([]byte, error) {
	response, err := c.httpClient.Do(req)
	if err != nil {
		if ctxErr := req.Context().Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, errors.New("wecom: request failed")
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, errors.New("wecom: read response failed")
	}
	if int64(len(body)) > limit {
		return nil, ErrResponseTooLarge
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		// WeCom normally answers 200 even for errors, but a proxy or gateway
		// may return an error envelope with a non-2xx status. Decoding it keeps
		// the errcode contract without echoing the request.
		var envelope struct {
			Errcode int    `json:"errcode"`
			Errmsg  string `json:"errmsg"`
		}
		if json.Unmarshal(body, &envelope) == nil && envelope.Errcode != 0 {
			return nil, c.newAPIError("request", envelope.Errcode, envelope.Errmsg)
		}
		return nil, fmt.Errorf("wecom: request failed with status %d", response.StatusCode)
	}
	return body, nil
}

// Upload sends one multipart/form-data request carrying a single file part and
// decodes the JSON response, retrying once after a token-invalid response.
func (c *Client) Upload(ctx context.Context, op, path string, query url.Values, field, filename string, data []byte, response any) error {
	token, err := c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.upload(ctx, op, path, query, token, field, filename, data, response)
	if !isTokenInvalid(err) {
		return err
	}
	c.invalidate(token)
	token, err = c.Token(ctx)
	if err != nil {
		return err
	}
	err = c.upload(ctx, op, path, query, token, field, filename, data, response)
	if isTokenInvalid(err) {
		c.invalidate(token)
	}
	return err
}

func (c *Client) upload(ctx context.Context, op, path string, query url.Values, token, field, filename string, data []byte, response any) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		return fmt.Errorf("wecom: encode %s request failed", op)
	}
	if _, err := part.Write(data); err != nil {
		return fmt.Errorf("wecom: encode %s request failed", op)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("wecom: encode %s request failed", op)
	}
	values := url.Values{}
	for key, entries := range query {
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}
	values.Set("access_token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+"?"+values.Encode(), &body)
	if err != nil {
		return fmt.Errorf("wecom: create %s request failed", op)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	raw, err := c.do(req)
	if err != nil {
		return err
	}
	return c.decode(op, raw, response)
}

// Download returns the raw response body of one authenticated request. A JSON
// error envelope is reported as an APIError; any other body is returned as-is.
func (c *Client) Download(ctx context.Context, op, path string, query url.Values) ([]byte, error) {
	token, err := c.Token(ctx)
	if err != nil {
		return nil, err
	}
	data, err := c.download(ctx, op, path, query, token)
	if !isTokenInvalid(err) {
		return data, err
	}
	c.invalidate(token)
	token, err = c.Token(ctx)
	if err != nil {
		return nil, err
	}
	data, err = c.download(ctx, op, path, query, token)
	if isTokenInvalid(err) {
		c.invalidate(token)
	}
	return data, err
}

func (c *Client) download(ctx context.Context, op, path string, query url.Values, token string) ([]byte, error) {
	values := url.Values{}
	for key, entries := range query {
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}
	values.Set("access_token", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path+"?"+values.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("wecom: create %s request failed", op)
	}
	data, err := c.doLimited(req, c.downloadLimit)
	if err != nil {
		return nil, err
	}
	// WeCom answers errors as JSON even on binary endpoints.
	if len(data) > 0 && data[0] == '{' {
		if apiErr := c.decodeError(op, data); apiErr != nil {
			return nil, apiErr
		}
	}
	return data, nil
}

// FetchURL performs one GET against an absolute URL that WeCom handed back, using
// the shared HTTP client and download limit. Some endpoints return a download
// descriptor (a bill file, a media link) instead of a WeCom API path; the caller
// supplies the headers that descriptor requires. The body is bounded by the
// download limit, and a body at or over it fails with ErrResponseTooLarge rather
// than returning a silently truncated file.
func (c *Client) FetchURL(ctx context.Context, rawURL string, header http.Header) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, errors.New("wecom: create download request failed")
	}
	for key, values := range header {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	response, err := c.httpClient.Do(request)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, errors.New("wecom: download failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("wecom: download failed with status %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, c.downloadLimit+1))
	if err != nil {
		return nil, errors.New("wecom: read download failed")
	}
	if int64(len(data)) > c.downloadLimit {
		return nil, ErrResponseTooLarge
	}
	return data, nil
}

// decodeError reports the JSON error envelope, or nil when the body is a normal
// payload.
func (c *Client) decodeError(op string, data []byte) error {
	var envelope struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}
	if envelope.Errcode == 0 && envelope.Errmsg == "" {
		return nil
	}
	if envelope.Errcode == 0 {
		return nil
	}
	return c.newAPIError(op, envelope.Errcode, envelope.Errmsg)
}
