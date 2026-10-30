package wedoc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// webhookClient bounds webhook calls, which do not share the token transport.
var webhookClient = &http.Client{Timeout: 15 * time.Second}

const smartsheetWebhookPath = "/cgi-bin/wedoc/smartsheet/webhook"

// WebhookRecord is one record written through a smart-sheet webhook. Values are
// keyed by field id and keep the WeCom cell shape.
type WebhookRecord struct {
	RecordID string                     `json:"record_id,omitempty"`
	Values   map[string]json.RawMessage `json:"values"`
}

// AddWebhookRecords appends records through the API-key webhook of a smart
// sheet. The webhook key is not a corp access token.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101240
func (c *WeDoc) AddWebhookRecords(ctx context.Context, key string, records []WebhookRecord) error {
	return c.webhook(ctx, "add webhook records", key, map[string]any{"add_records": records})
}

// UpdateWebhookRecords updates existing records through a smart-sheet webhook.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101241
func (c *WeDoc) UpdateWebhookRecords(ctx context.Context, key string, records []WebhookRecord) error {
	for _, record := range records {
		if strings.TrimSpace(record.RecordID) == "" {
			return errors.New("wecom: update webhook records: record_id is required")
		}
	}
	return c.webhook(ctx, "update webhook records", key, map[string]any{"update_records": records})
}

func (c *WeDoc) webhook(ctx context.Context, op, key string, payload map[string]any) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: document client is unavailable")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("wecom: " + op + ": key is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return errors.New("wecom: " + op + ": encode failed")
	}
	endpoint := c.api.BaseURL() + smartsheetWebhookPath + "?key=" + url.QueryEscape(key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("wecom: " + op + ": request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := webhookClient.Do(req)
	if err != nil {
		return fmt.Errorf("wecom: %s failed", op)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("wecom: %s failed", op)
	}
	var envelope struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Errcode != 0 {
		return fmt.Errorf("wecom: %s rejected with errcode %d", op, envelope.Errcode)
	}
	return nil
}
