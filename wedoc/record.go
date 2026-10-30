package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// CellValueKeyTypeFieldTitle addresses record cells by field title;
	// CellValueKeyTypeFieldID addresses them by field id.
	CellValueKeyTypeFieldTitle = "CELL_VALUE_KEY_TYPE_FIELD_TITLE"
	CellValueKeyTypeFieldID    = "CELL_VALUE_KEY_TYPE_FIELD_ID"
	addRecordsPath             = smartsheetPath + "/add_records"
	getRecordsPath             = smartsheetPath + "/get_records"
	updateRecordsPath          = smartsheetPath + "/update_records"
	deleteRecordsPath          = smartsheetPath + "/delete_records"
)

// TextValue is the wire form of a text cell. Type is "text" for plain content
// and "url" for a link; Link is only set for links.
type TextValue struct {
	Type string `json:"type"`
	Text string `json:"text"`
	Link string `json:"link,omitempty"`
}

// TextCell builds the value of a FIELD_TYPE_TEXT cell. Callers store the
// result in Record.Values under the column title instead of assembling the
// [{type,text}] array themselves.
func TextCell(content string) []TextValue {
	return []TextValue{{Type: "text", Text: content}}
}

// LinkCell builds a text cell that links to url.
func LinkCell(text, link string) []TextValue {
	return []TextValue{{Type: "url", Text: text, Link: link}}
}

// Record is one smart sheet row. AddRecords leaves RecordID empty; updates and
// responses carry it. Values is keyed by field title by default.
type Record struct {
	RecordID    string         `json:"record_id,omitempty"`
	CreateTime  string         `json:"create_time,omitempty"`
	UpdateTime  string         `json:"update_time,omitempty"`
	Values      map[string]any `json:"values,omitempty"`
	CreatorName string         `json:"creator_name,omitempty"`
	UpdaterName string         `json:"updater_name,omitempty"`
}

// Sort orders query results by a field title.
type Sort struct {
	FieldTitle string `json:"field_title"`
	Desc       bool   `json:"desc,omitempty"`
}

// TextRecord builds a record whose cells are all text, keyed by column title.
func TextRecord(cells map[string]string) Record {
	values := make(map[string]any, len(cells))
	for title, text := range cells {
		values[title] = TextCell(text)
	}
	return Record{Values: values}
}

// AddRecords appends one or more rows and returns the created rows.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99907
func (c *WeDoc) AddRecords(ctx context.Context, docID, sheetID string, records ...Record) ([]Record, error) {
	docID, sheetID, err := c.recordTarget("add records", docID, sheetID)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("wecom: add records: at least one record is required")
	}
	request := recordsRequest{
		DocID:   docID,
		SheetID: sheetID,
		KeyType: CellValueKeyTypeFieldTitle,
		Records: records,
	}
	var response struct {
		Records []Record `json:"records"`
	}
	if err := c.call(ctx, "add records", addRecordsPath, request, &response); err != nil {
		return nil, err
	}
	return response.Records, nil
}

// UpdateRecords updates existing rows and returns the updated rows. Every
// record must carry a non-empty RecordID.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99909
func (c *WeDoc) UpdateRecords(ctx context.Context, docID, sheetID string, records ...Record) ([]Record, error) {
	docID, sheetID, err := c.recordTarget("update records", docID, sheetID)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errors.New("wecom: update records: at least one record is required")
	}
	for index, record := range records {
		if strings.TrimSpace(record.RecordID) == "" {
			return nil, fmt.Errorf("wecom: update records: records[%d].recordId is required", index)
		}
	}
	request := recordsRequest{
		DocID:   docID,
		SheetID: sheetID,
		KeyType: CellValueKeyTypeFieldTitle,
		Records: records,
	}
	var response struct {
		Records []Record `json:"records"`
	}
	if err := c.call(ctx, "update records", updateRecordsPath, request, &response); err != nil {
		return nil, err
	}
	return response.Records, nil
}

// DeleteRecords removes the rows identified by recordIDs.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99908
func (c *WeDoc) DeleteRecords(ctx context.Context, docID, sheetID string, recordIDs ...string) error {
	docID, sheetID, err := c.recordTarget("delete records", docID, sheetID)
	if err != nil {
		return err
	}
	if len(recordIDs) == 0 {
		return errors.New("wecom: delete records: at least one recordId is required")
	}
	ids := make([]string, 0, len(recordIDs))
	for index, recordID := range recordIDs {
		trimmed := strings.TrimSpace(recordID)
		if trimmed == "" {
			return fmt.Errorf("wecom: delete records: recordIds[%d] is required", index)
		}
		ids = append(ids, trimmed)
	}
	request := struct {
		DocID     string   `json:"docid"`
		SheetID   string   `json:"sheet_id"`
		RecordIDs []string `json:"record_ids"`
	}{DocID: docID, SheetID: sheetID, RecordIDs: ids}
	if err := c.call(ctx, "delete records", deleteRecordsPath, request, nil); err != nil {
		return err
	}
	return nil
}

// RecordQuery narrows a record query. The zero value asks for every row,
// walking offset/limit pages until the server reports the end.
type RecordQuery struct {
	// ViewID restricts the query to one view.
	ViewID string
	// RecordIDs restricts the query to specific rows.
	RecordIDs []string
	// KeyType selects field-title or field-id keys. Empty means field title.
	KeyType string
	// FieldTitles and FieldIDs select the columns to return; they must match
	// KeyType.
	FieldTitles []string
	FieldIDs    []string
	// Sort orders the returned rows. It cannot be combined with FilterSpec.
	Sort []Sort
	// Offset and Limit control the first page. Zero Limit uses the maximum.
	Offset int
	Limit  int
	// Ver pins a table version.
	Ver int
	// FilterSpec is the documented condition object.
	FilterSpec map[string]any
	// NeedFormulaResult asks for formula and reference results.
	NeedFormulaResult bool
}

// GetRecords returns every row of a sub-table, walking offset/limit pages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101158
func (c *WeDoc) GetRecords(ctx context.Context, docID, sheetID string) ([]Record, error) {
	return c.QueryRecords(ctx, docID, sheetID, RecordQuery{})
}

// QueryRecords returns every matching row, walking offset/limit pages until
// the server reports the end.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101158
func (c *WeDoc) QueryRecords(ctx context.Context, docID, sheetID string, query RecordQuery) ([]Record, error) {
	docID, sheetID, err := c.recordTarget("get records", docID, sheetID)
	if err != nil {
		return nil, err
	}
	if query.Offset < 0 {
		return nil, errors.New("wecom: get records: offset must not be negative")
	}
	if query.Limit < 0 {
		return nil, errors.New("wecom: get records: limit must not be negative")
	}
	limit := query.Limit
	if limit == 0 || limit > maxListPageSize {
		limit = maxListPageSize
	}
	keyType := strings.TrimSpace(query.KeyType)
	if keyType == "" {
		keyType = CellValueKeyTypeFieldTitle
	}
	offset := query.Offset
	records := make([]Record, 0)
	for {
		page, err := c.getRecordsPage(ctx, docID, sheetID, query, keyType, offset, limit)
		if err != nil {
			return nil, err
		}
		records = append(records, page.Records...)
		if len(page.Records) == 0 {
			return records, nil
		}
		next := offset + len(page.Records)
		if page.HasMore || page.Next != 0 {
			if int(page.Next) > offset {
				next = int(page.Next)
			}
		} else if page.Total == 0 || next >= int(page.Total) {
			return records, nil
		}
		if next <= offset {
			return nil, errors.New("wecom: get records pagination did not advance")
		}
		offset = next
	}
}

type recordsRequest struct {
	DocID   string   `json:"docid"`
	SheetID string   `json:"sheet_id"`
	KeyType string   `json:"key_type,omitempty"`
	Records []Record `json:"records"`
}

type recordPage struct {
	Records []Record `json:"records"`
	Total   uint32   `json:"total"`
	HasMore bool     `json:"has_more"`
	Next    uint32   `json:"next"`
	Ver     uint32   `json:"ver"`
}

func (c *WeDoc) getRecordsPage(ctx context.Context, docID, sheetID string, query RecordQuery, keyType string, offset, limit int) (recordPage, error) {
	request := struct {
		DocID             string         `json:"docid"`
		SheetID           string         `json:"sheet_id"`
		ViewID            string         `json:"view_id,omitempty"`
		RecordIDs         []string       `json:"record_ids,omitempty"`
		KeyType           string         `json:"key_type,omitempty"`
		FieldTitles       []string       `json:"field_titles,omitempty"`
		FieldIDs          []string       `json:"field_ids,omitempty"`
		Sort              []Sort         `json:"sort,omitempty"`
		Offset            int            `json:"offset"`
		Limit             int            `json:"limit"`
		Ver               int            `json:"ver,omitempty"`
		FilterSpec        map[string]any `json:"filter_spec,omitempty"`
		NeedFormulaResult bool           `json:"need_formula_result,omitempty"`
	}{
		DocID:             docID,
		SheetID:           sheetID,
		ViewID:            strings.TrimSpace(query.ViewID),
		RecordIDs:         query.RecordIDs,
		KeyType:           keyType,
		FieldTitles:       query.FieldTitles,
		FieldIDs:          query.FieldIDs,
		Sort:              query.Sort,
		Offset:            offset,
		Limit:             limit,
		Ver:               query.Ver,
		FilterSpec:        query.FilterSpec,
		NeedFormulaResult: query.NeedFormulaResult,
	}
	var response recordPage
	if err := c.call(ctx, "get records", getRecordsPath, request, &response); err != nil {
		return recordPage{}, err
	}
	return response, nil
}

// recordTarget validates the document and sub-table ids shared by the record
// methods. Error messages name the field, never its value.
func (c *WeDoc) recordTarget(op, docID, sheetID string) (string, string, error) {
	docID, err := required(op, "docId", docID)
	if err != nil {
		return "", "", err
	}
	sheetID, err = required(op, "sheetId", sheetID)
	if err != nil {
		return "", "", err
	}
	return docID, sheetID, nil
}
