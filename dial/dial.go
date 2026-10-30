// Package dial implements the WeCom paid-call record API (/cgi-bin/dial).
package dial

import (
	"context"
	"errors"

	"github.com/gowsp/go-wecom/client"
)

// Dial reads paid-call records with one shared transport.
type Dial struct{ api *client.Client }

// New creates a dial client.
func New(api *client.Client) *Dial { return &Dial{api: api} }

// Caller is the member who placed the call.
type Caller struct {
	UserID   string `json:"userid"`
	Duration int    `json:"duration"`
}

// Callee is one dialled number and its talk time.
type Callee struct {
	Phone    string `json:"phone"`
	Duration int    `json:"duration"`
}

// Record is one paid-call record.
type Record struct {
	CallTime      int64    `json:"call_time"`
	TotalDuration int      `json:"total_duration"`
	CallType      int      `json:"call_type"`
	Caller        Caller   `json:"caller"`
	Callee        []Callee `json:"callee"`
}

// Records returns the paid-call records of one time window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93662
func (c *Dial) Records(ctx context.Context, startTime, endTime int64, offset, limit int) ([]Record, error) {
	if c == nil || c.api == nil {
		return nil, errors.New("wecom: dial client is unavailable")
	}
	if startTime <= 0 || endTime <= startTime {
		return nil, errors.New("wecom: dial records: a valid time window is required")
	}
	request := struct {
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
		Offset    int   `json:"offset,omitempty"`
		Limit     int   `json:"limit,omitempty"`
	}{StartTime: startTime, EndTime: endTime, Offset: offset, Limit: limit}
	var out struct {
		Record []Record `json:"record"`
	}
	err := c.api.Call(ctx, "dial records", "/cgi-bin/dial/get_dial_record", request, &out)
	return out.Record, err
}
