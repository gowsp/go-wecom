// Package todo implements the WeCom to-do API (/cgi-bin/todo) on top of the
// shared authenticated client.
package todo

import (
	"context"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Todo reads and updates to-do items with one shared transport.
type Todo struct{ api *client.Client }

// New creates a to-do client.
func New(api *client.Client) *Todo { return &Todo{api: api} }

// Attendee is one participant of a to-do item.
type Attendee struct {
	UserID string `json:"userid"`
	Status int    `json:"status"`
}

// Reminder is one reminder of a to-do item.
type Reminder struct {
	RemindTime int64 `json:"remind_time"`
}

// Item is one to-do. Status 0 pending, 1 completed.
type Item struct {
	ID         string     `json:"todo_id"`
	Content    string     `json:"content"`
	Creator    string     `json:"creator"`
	Status     int        `json:"status"`
	CreateTime int64      `json:"create_time"`
	EndTime    int64      `json:"end_time,omitempty"`
	Attendees  []Attendee `json:"attendees,omitempty"`
	Reminders  []Reminder `json:"reminders,omitempty"`
}

// Get reads one to-do item.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101524
func (c *Todo) Get(ctx context.Context, todoID string) (Item, error) {
	if c == nil || c.api == nil {
		return Item{}, errors.New("wecom: todo client is unavailable")
	}
	todoID = strings.TrimSpace(todoID)
	if todoID == "" {
		return Item{}, errors.New("wecom: todo get: todo_id is required")
	}
	var out Item
	err := c.api.Call(ctx, "todo get", "/cgi-bin/todo/get", map[string]string{"todo_id": todoID}, &out)
	return out, err
}

// Update changes the status of one to-do and, when attendees are given, of the
// listed participants. Omitting attendees updates the whole item.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101534
func (c *Todo) Update(ctx context.Context, todoID string, status int, attendees []Attendee) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: todo client is unavailable")
	}
	todoID = strings.TrimSpace(todoID)
	if todoID == "" {
		return errors.New("wecom: todo update: todo_id is required")
	}
	if status != 0 && status != 1 {
		return errors.New("wecom: todo update: status must be 0 or 1")
	}
	request := struct {
		ID        string     `json:"todo_id"`
		Status    int        `json:"status"`
		Attendees []Attendee `json:"attendees,omitempty"`
	}{ID: todoID, Status: status, Attendees: attendees}
	return c.api.Call(ctx, "todo update", "/cgi-bin/todo/update", request, nil)
}
