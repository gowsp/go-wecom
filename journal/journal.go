// Package journal implements the WeCom report API (/cgi-bin/oa/journal).
package journal

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// Journal reads work reports with one shared transport.
type Journal struct{ api *client.Client }

// New creates a report client.
func New(api *client.Client) *Journal { return &Journal{api: api} }

// Filter narrows a report query. Key is one of creator, department or
// template_id; value is the matching id.
type Filter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RecordQuery is one page request for report record ids.
type RecordQuery struct {
	StartTime int64    `json:"starttime"`
	EndTime   int64    `json:"endtime"`
	Cursor    int      `json:"cursor,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Filters   []Filter `json:"filters,omitempty"`
}

// RecordPage is one page of report record ids.
type RecordPage struct {
	NextCursor int      `json:"next_cursor"`
	UUIDs      []string `json:"journaluuid_list"`
}

// Sender identifies a member in a report.
type Sender struct {
	UserID string `json:"userid"`
}

// Detail is one work report. Contents stay raw because report templates define
// their own question and answer structures.
type Detail struct {
	UUID          string          `json:"journal_uuid"`
	TemplateName  string          `json:"template_name"`
	ReportTime    int64           `json:"report_time"`
	Submitter     Sender          `json:"submitter"`
	Receivers     []Sender        `json:"receivers,omitempty"`
	ReadReceivers []Sender        `json:"readed_receivers,omitempty"`
	Content       json.RawMessage `json:"content,omitempty"`
}

// Records returns one page of report record ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93393
func (c *Journal) Records(ctx context.Context, query RecordQuery) (RecordPage, error) {
	if c == nil || c.api == nil {
		return RecordPage{}, errors.New("wecom: journal client is unavailable")
	}
	if query.StartTime <= 0 || query.EndTime <= query.StartTime {
		return RecordPage{}, errors.New("wecom: journal records: a valid time window is required")
	}
	var out RecordPage
	err := c.api.Call(ctx, "journal records", "/cgi-bin/oa/journal/get_record_list", query, &out)
	return out, err
}

// Detail reads one report by uuid.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93394
func (c *Journal) Detail(ctx context.Context, uuid string) (Detail, error) {
	if c == nil || c.api == nil {
		return Detail{}, errors.New("wecom: journal client is unavailable")
	}
	uuid = strings.TrimSpace(uuid)
	if uuid == "" {
		return Detail{}, errors.New("wecom: journal detail: journaluuid is required")
	}
	var out struct {
		Info Detail `json:"info"`
	}
	err := c.api.Call(ctx, "journal detail", "/cgi-bin/oa/journal/get_record_detail",
		map[string]string{"journaluuid": uuid}, &out)
	return out.Info, err
}

// Stat is the per-template statistics of one report window.
type Stat struct {
	TemplateID   string          `json:"template_id"`
	TemplateName string          `json:"template_name"`
	ReportRange  json.RawMessage `json:"report_range,omitempty"`
}

// Stats returns the report statistics of one template in a window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93395
func (c *Journal) Stats(ctx context.Context, templateID string, startTime, endTime int64) ([]Stat, error) {
	if c == nil || c.api == nil {
		return nil, errors.New("wecom: journal client is unavailable")
	}
	if strings.TrimSpace(templateID) == "" {
		return nil, errors.New("wecom: journal stats: template_id is required")
	}
	if startTime <= 0 || endTime <= startTime {
		return nil, errors.New("wecom: journal stats: a valid time window is required")
	}
	request := struct {
		TemplateID string `json:"template_id"`
		StartTime  int64  `json:"starttime"`
		EndTime    int64  `json:"endtime"`
	}{TemplateID: templateID, StartTime: startTime, EndTime: endTime}
	var out struct {
		Stats []Stat `json:"stat_list"`
	}
	err := c.api.Call(ctx, "journal stats", "/cgi-bin/oa/journal/get_stat_list", request, &out)
	return out.Stats, err
}

// Attachment is one downloadable report attachment in the micro disk.
type Attachment struct {
	URL         string `json:"download_url"`
	CookieName  string `json:"cookie_name"`
	CookieValue string `json:"cookie_value"`
}

// DownloadAttachment returns the download address and cookie of one report
// attachment stored in the micro disk.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98021
func (c *Journal) DownloadAttachment(ctx context.Context, journalUUID, fileID string) (Attachment, error) {
	if c == nil || c.api == nil {
		return Attachment{}, errors.New("wecom: journal client is unavailable")
	}
	if strings.TrimSpace(journalUUID) == "" || strings.TrimSpace(fileID) == "" {
		return Attachment{}, errors.New("wecom: journal attachment: journaluuid and fileid are required")
	}
	request := struct {
		JournalUUID string `json:"journaluuid"`
		FileID      string `json:"fileid"`
	}{JournalUUID: journalUUID, FileID: fileID}
	var out Attachment
	err := c.api.Call(ctx, "journal attachment",
		"/cgi-bin/oa/journal/download_wedrive_file", request, &out)
	return out, err
}
