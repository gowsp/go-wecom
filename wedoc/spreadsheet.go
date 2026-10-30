package wedoc

import (
	"context"
	"encoding/json"
	"strings"
)

const (
	spreadsheetBatchUpdatePath = "/cgi-bin/wedoc/spreadsheet/batch_update"
	spreadsheetPropertiesPath  = "/cgi-bin/wedoc/spreadsheet/get_sheet_properties"
	spreadsheetRangePath       = "/cgi-bin/wedoc/spreadsheet/get_sheet_range_data"
)

// BatchUpdateSpreadsheet applies a list of edit requests to a spreadsheet. Each
// entry keeps the WeCom request shape, for example {"add_sheet_request": {...}},
// and the matching responses are returned as raw JSON.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101168
func (c *WeDoc) BatchUpdateSpreadsheet(ctx context.Context, docID string, requests []map[string]any) (json.RawMessage, error) {
	docID, err := required("batch update spreadsheet", "docId", docID)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, errEmpty("batch update spreadsheet", "requests")
	}
	body := struct {
		DocID    string           `json:"docid"`
		Requests []map[string]any `json:"requests"`
	}{DocID: docID, Requests: requests}
	var out struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.call(ctx, "batch update spreadsheet", spreadsheetBatchUpdatePath, body, &out); err != nil {
		return nil, err
	}
	return out.Data, nil
}

// SpreadsheetProperties returns the sheet list and dimensions of a spreadsheet.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97711
func (c *WeDoc) SpreadsheetProperties(ctx context.Context, docID string) (json.RawMessage, error) {
	docID, err := required("get spreadsheet properties", "docId", docID)
	if err != nil {
		return nil, err
	}
	var out struct {
		Properties json.RawMessage `json:"properties"`
	}
	err = c.call(ctx, "get spreadsheet properties", spreadsheetPropertiesPath, map[string]string{"docid": docID}, &out)
	return out.Properties, err
}

// SpreadsheetRange returns the cell values of one A1:B2 style range.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97661
func (c *WeDoc) SpreadsheetRange(ctx context.Context, docID, sheetID, cellRange string) (json.RawMessage, error) {
	docID, err := required("get spreadsheet range", "docId", docID)
	if err != nil {
		return nil, err
	}
	if sheetID, err = required("get spreadsheet range", "sheetId", sheetID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(cellRange) == "" {
		return nil, errEmpty("get spreadsheet range", "range")
	}
	body := struct {
		DocID   string `json:"docid"`
		SheetID string `json:"sheet_id"`
		Range   string `json:"range"`
	}{DocID: docID, SheetID: sheetID, Range: cellRange}
	var out struct {
		Data json.RawMessage `json:"data"`
	}
	err = c.call(ctx, "get spreadsheet range", spreadsheetRangePath, body, &out)
	return out.Data, err
}
