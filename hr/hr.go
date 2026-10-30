// Package hr implements the WeCom HR assistant API (/cgi-bin/hr): employee field
// configuration and roster records.
package hr

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// HR reads and updates employee roster data with one shared transport.
type HR struct{ api *client.Client }

// New creates an HR client.
func New(api *client.Client) *HR { return &HR{api: api} }

// Field describes one configurable roster field.
type Field struct {
	FieldID   int    `json:"fieldid"`
	FieldName string `json:"field_name"`
	FieldType int    `json:"field_type"`
	ValueType int    `json:"value_type"`
	IsMust    bool   `json:"is_must"`
}

// FieldGroup groups fields by section, for example 个人信息.
type FieldGroup struct {
	GroupID   int     `json:"group_id"`
	GroupName string  `json:"group_name"`
	Fields    []Field `json:"field_list"`
}

// Fields returns the field configuration of the enterprise roster.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99131
func (c *HR) Fields(ctx context.Context) ([]FieldGroup, error) {
	if c == nil || c.api == nil {
		return nil, errors.New("wecom: hr client is unavailable")
	}
	var out struct {
		Groups []FieldGroup `json:"group_list"`
	}
	err := c.api.Call(ctx, "hr fields", "/cgi-bin/hr/get_fields", nil, &out)
	return out.Groups, err
}

// FieldRef selects one field value; SubIdx addresses list-valued fields.
type FieldRef struct {
	FieldID int `json:"fieldid"`
	SubIdx  int `json:"sub_idx,omitempty"`
}

// StaffInfo reads the roster of one employee. Values are returned raw because
// their shape depends on the field type (string, number, mobile, file, ...);
// callers decode the value_* key matching ValueType.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99132
func (c *HR) StaffInfo(ctx context.Context, userID string, getAll bool, fields []FieldRef) ([]json.RawMessage, error) {
	if c == nil || c.api == nil {
		return nil, errors.New("wecom: hr client is unavailable")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("wecom: hr staff info: userid is required")
	}
	if !getAll && len(fields) == 0 {
		return nil, errors.New("wecom: hr staff info: get_all or fieldids is required")
	}
	request := struct {
		UserID   string     `json:"userid"`
		GetAll   bool       `json:"get_all,omitempty"`
		FieldIDs []FieldRef `json:"fieldids,omitempty"`
	}{UserID: userID, GetAll: getAll, FieldIDs: fields}
	var out struct {
		Info []json.RawMessage `json:"field_info"`
	}
	err := c.api.Call(ctx, "hr staff info", "/cgi-bin/hr/get_staff_info", request, &out)
	return out.Info, err
}

// StaffUpdate carries the polymorphic update payload. Items stay raw because
// each value_* key depends on the field type.
type StaffUpdate struct {
	UserID      string            `json:"userid"`
	UpdateItems []json.RawMessage `json:"update_items,omitempty"`
	RemoveItems []json.RawMessage `json:"remove_items,omitempty"`
	InsertItems []json.RawMessage `json:"insert_items,omitempty"`
}

// StaffUpdateResult is the per-item outcome of one roster update.
type StaffUpdateResult struct {
	UpdateResults []json.RawMessage `json:"update_results"`
	RemoveResults []json.RawMessage `json:"remove_results"`
	InsertResults []json.RawMessage `json:"insert_result"`
}

// UpdateStaff writes roster values of one employee.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99133
func (c *HR) UpdateStaff(ctx context.Context, update StaffUpdate) (StaffUpdateResult, error) {
	if c == nil || c.api == nil {
		return StaffUpdateResult{}, errors.New("wecom: hr client is unavailable")
	}
	if strings.TrimSpace(update.UserID) == "" {
		return StaffUpdateResult{}, errors.New("wecom: hr update staff: userid is required")
	}
	if len(update.UpdateItems) == 0 && len(update.RemoveItems) == 0 && len(update.InsertItems) == 0 {
		return StaffUpdateResult{}, errors.New("wecom: hr update staff: no item to change")
	}
	var out StaffUpdateResult
	err := c.api.Call(ctx, "hr update staff", "/cgi-bin/hr/update_staff_info", update, &out)
	return out, err
}
