// Package chatdata implements the WeCom chat data zone API (/cgi-bin/chatdata
// and /cgi-bin/docdata): public key registration, archive authorisation,
// callback routing, hidden sensitive info, program invocation and debug mode.
//
// The zone program itself calls a separate SDK for the session data; those
// interfaces are not HTTP endpoints and are out of scope for this client.
package chatdata

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// ChatData manages the chat data zone with one shared transport.
type ChatData struct{ api *client.Client }

// New creates a chat data zone client.
func New(api *client.Client) *ChatData { return &ChatData{api: api} }

// HiddenConfig hides sensitive fields in the session component.
type HiddenConfig struct {
	HideMobile bool `json:"hide_mobile"`
	HideIDCard bool `json:"hide_idcard"`
	HideBankNo bool `json:"hide_bankno"`
}

// AuthUser is one member authorised to read archived sessions.
type AuthUser struct {
	UserID      string `json:"userid"`
	EditionList []int  `json:"edition_list,omitempty"`
}

// AuthPage is one page of authorised members.
type AuthPage struct {
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor"`
	Users      []AuthUser `json:"auth_user_list"`
}

// SetPublicKey registers the RSA public key used to encrypt archived messages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99961
func (c *ChatData) SetPublicKey(ctx context.Context, publicKey string) error {
	if err := check(c); err != nil {
		return err
	}
	publicKey = strings.TrimSpace(publicKey)
	if publicKey == "" {
		return errors.New("wecom: chatdata public key: public_key is required")
	}
	return c.api.Call(ctx, "chatdata public key", "/cgi-bin/chatdata/set_public_key",
		map[string]string{"public_key": publicKey}, nil)
}

// ArchiveUsers returns one page of members authorised to read archived sessions.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99962
func (c *ChatData) ArchiveUsers(ctx context.Context, cursor string, limit int) (AuthPage, error) {
	if err := check(c); err != nil {
		return AuthPage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out AuthPage
	err := c.api.Call(ctx, "chatdata archive users", "/cgi-bin/chatdata/get_auth_user_list", request, &out)
	return out, err
}

// FileArchiveUsers returns one page of members authorised to read archived file
// contents.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101873
func (c *ChatData) FileArchiveUsers(ctx context.Context, cursor string, limit int) (AuthPage, error) {
	if err := check(c); err != nil {
		return AuthPage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out AuthPage
	err := c.api.Call(ctx, "chatdata file archive users", "/cgi-bin/docdata/get_auth_user_list", request, &out)
	return out, err
}

// SetReceiveCallback routes zone events to one program.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99963
func (c *ChatData) SetReceiveCallback(ctx context.Context, programID string) error {
	if err := check(c); err != nil {
		return err
	}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		return errors.New("wecom: chatdata callback: program_id is required")
	}
	return c.api.Call(ctx, "chatdata receive callback", "/cgi-bin/chatdata/set_receive_callback",
		map[string]string{"program_id": programID}, nil)
}

// Hidden returns the hidden-field configuration of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100139
func (c *ChatData) Hidden(ctx context.Context, userID string) (HiddenConfig, error) {
	if err := check(c); err != nil {
		return HiddenConfig{}, err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return HiddenConfig{}, errors.New("wecom: chatdata hidden: userid is required")
	}
	var out struct {
		Config HiddenConfig `json:"config"`
	}
	err := c.api.Call(ctx, "chatdata hidden",
		"/cgi-bin/chatdata/get_hide_sensitiveinfo_config",
		map[string]string{"userid": userID}, &out)
	return out.Config, err
}

// SetHidden updates the hidden-field configuration of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100139
func (c *ChatData) SetHidden(ctx context.Context, userID string, config HiddenConfig) error {
	if err := check(c); err != nil {
		return err
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("wecom: chatdata set hidden: userid is required")
	}
	request := struct {
		UserID string       `json:"userid"`
		Config HiddenConfig `json:"config"`
	}{UserID: userID, Config: config}
	return c.api.Call(ctx, "chatdata set hidden",
		"/cgi-bin/chatdata/set_hide_sensitiveinfo_config", request, nil)
}

// LogLevel returns the log level of one zone program.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100108
func (c *ChatData) LogLevel(ctx context.Context, programID string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		return 0, errors.New("wecom: chatdata log level: program_id is required")
	}
	var out struct {
		Level int `json:"log_level"`
	}
	err := c.api.Call(ctx, "chatdata log level", "/cgi-bin/chatdata/get_log_level",
		map[string]string{"program_id": programID}, &out)
	return out.Level, err
}

// SetLogLevel updates the log level of one zone program.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100108
func (c *ChatData) SetLogLevel(ctx context.Context, programID string, level int) error {
	if err := check(c); err != nil {
		return err
	}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		return errors.New("wecom: chatdata set log level: program_id is required")
	}
	request := struct {
		ProgramID string `json:"program_id"`
		Level     int    `json:"log_level"`
	}{ProgramID: programID, Level: level}
	return c.api.Call(ctx, "chatdata set log level", "/cgi-bin/chatdata/set_log_level", request, nil)
}

// UploadMedia uploads one temporary file into the zone and returns its media id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100174
func (c *ChatData) UploadMedia(ctx context.Context, filename string, data []byte) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(filename) == "" || len(data) == 0 {
		return "", errors.New("wecom: chatdata upload: filename and data are required")
	}
	var out struct {
		MediaID string `json:"media_id"`
	}
	err := c.api.Upload(ctx, "chatdata upload", "/cgi-bin/chatdata/upload_media",
		nil, "media", filename, data, &out)
	return out.MediaID, err
}

// CallProgram invokes one zone program synchronously and returns its response
// payload.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99965
func (c *ChatData) CallProgram(ctx context.Context, programID, abilityID, notifyID string, request []byte) ([]byte, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(programID) == "" || strings.TrimSpace(abilityID) == "" {
		return nil, errors.New("wecom: chatdata call: program_id and ability_id are required")
	}
	if len(request) > 0 && !json.Valid(request) {
		return nil, errors.New("wecom: chatdata call: request_data must be valid JSON")
	}
	payload := struct {
		ProgramID   string `json:"program_id"`
		AbilityID   string `json:"ability_id"`
		NotifyID    string `json:"notify_id,omitempty"`
		RequestData string `json:"request_data,omitempty"`
	}{ProgramID: programID, AbilityID: abilityID, NotifyID: notifyID, RequestData: string(request)}
	var out struct {
		ResponseData string `json:"response_data"`
	}
	if err := c.api.Call(ctx, "chatdata call", "/cgi-bin/chatdata/sync_call_program", payload, &out); err != nil {
		return nil, err
	}
	return []byte(out.ResponseData), nil
}

// SubmitProgramTask starts one asynchronous zone program call and returns the
// job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99966
func (c *ChatData) SubmitProgramTask(ctx context.Context, programID, abilityID string, request []byte) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(programID) == "" || strings.TrimSpace(abilityID) == "" {
		return "", errors.New("wecom: chatdata task: program_id and ability_id are required")
	}
	if len(request) > 0 && !json.Valid(request) {
		return "", errors.New("wecom: chatdata task: request_data must be valid JSON")
	}
	payload := struct {
		ProgramID   string `json:"program_id"`
		AbilityID   string `json:"ability_id"`
		RequestData string `json:"request_data,omitempty"`
	}{ProgramID: programID, AbilityID: abilityID, RequestData: string(request)}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.api.Call(ctx, "chatdata task", "/cgi-bin/chatdata/async_program_task", payload, &out)
	return out.JobID, err
}

// ProgramResult reads the result of one asynchronous zone program call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99966
func (c *ChatData) ProgramResult(ctx context.Context, jobID string) (json.RawMessage, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return nil, errors.New("wecom: chatdata result: jobid is required")
	}
	var out struct {
		Status       int             `json:"status"`
		ResponseData json.RawMessage `json:"response_data"`
	}
	err := c.api.Call(ctx, "chatdata result", "/cgi-bin/chatdata/async_program_result",
		map[string]string{"jobid": jobID}, &out)
	return out.ResponseData, err
}

// DebugMode reports whether the zone program runs in debug mode.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100113
func (c *ChatData) DebugMode(ctx context.Context, programID string) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		return 0, errors.New("wecom: chatdata debug mode: program_id is required")
	}
	var out struct {
		Status int `json:"debug_mode_status"`
	}
	err := c.api.Call(ctx, "chatdata debug mode", "/cgi-bin/chatdata/check_debug_mode",
		map[string]string{"program_id": programID}, &out)
	return out.Status, err
}

// OpenDebugMode enables debug mode for one zone program.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100087
func (c *ChatData) OpenDebugMode(ctx context.Context, programID, debugToken string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(programID) == "" || strings.TrimSpace(debugToken) == "" {
		return errors.New("wecom: chatdata open debug: program_id and debug_token are required")
	}
	request := struct {
		ProgramID  string `json:"program_id"`
		DebugToken string `json:"debug_token"`
	}{ProgramID: programID, DebugToken: debugToken}
	return c.api.Call(ctx, "chatdata open debug", "/cgi-bin/chatdata/open_debug_mode", request, nil)
}

// CloseDebugMode disables debug mode for one zone program.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100088
func (c *ChatData) CloseDebugMode(ctx context.Context, programID string) error {
	if err := check(c); err != nil {
		return err
	}
	programID = strings.TrimSpace(programID)
	if programID == "" {
		return errors.New("wecom: chatdata close debug: program_id is required")
	}
	return c.api.Call(ctx, "chatdata close debug", "/cgi-bin/chatdata/close_debug_mode",
		map[string]string{"program_id": programID}, nil)
}

func check(c *ChatData) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: chatdata client is unavailable")
	}
	return nil
}

// Notify is the app-side callback triggered when a zone program notifies the
// application. Decrypt the body with callback.CallbackCrypto first.
type Notify struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUser"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	NotifyID   string `xml:"NotifyId"`
	Scene      int    `xml:"NotifyScene"`
}

// ParseNotify decodes a decrypted program_notify callback. The notify id can be
// passed to CallProgram as notify_id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99998
func ParseNotify(data []byte) (Notify, error) {
	var notify Notify
	if err := xml.Unmarshal(data, &notify); err != nil {
		return Notify{}, errors.New("wecom: chatdata notify: " + err.Error())
	}
	if notify.Name != "program_notify" {
		return Notify{}, errors.New("wecom: chatdata notify: unexpected event " + notify.Name)
	}
	if strings.TrimSpace(notify.NotifyID) == "" {
		return Notify{}, errors.New("wecom: chatdata notify: NotifyId is required")
	}
	return notify, nil
}
