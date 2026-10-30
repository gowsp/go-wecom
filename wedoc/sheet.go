package wedoc

import (
	"context"
	"errors"
	"strings"
)

const (
	smartsheetPath  = "/cgi-bin/wedoc/smartsheet"
	addSheetPath    = smartsheetPath + "/add_sheet"
	updateSheetPath = smartsheetPath + "/update_sheet"
	deleteSheetPath = smartsheetPath + "/delete_sheet"
	getSheetPath    = smartsheetPath + "/get_sheet"
)

// Sheet describes one smart sheet sub-table returned by get_sheet.
type Sheet struct {
	SheetID   string `json:"sheet_id,omitempty"`
	Title     string `json:"title,omitempty"`
	Index     int    `json:"index,omitempty"`
	IsVisible bool   `json:"is_visible,omitempty"`
	Type      string `json:"type,omitempty"`
}

// sheetProperties is the wire object used by add_sheet and update_sheet.
type sheetProperties struct {
	SheetID string `json:"sheet_id,omitempty"`
	Title   string `json:"title,omitempty"`
	Index   *int   `json:"index,omitempty"`
}

// AddSheet adds one sub-table to a smart sheet document. An optional index
// places it at that position; negative indexes are rejected.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99896
func (c *WeDoc) AddSheet(ctx context.Context, docID, title string, index ...int) (Sheet, error) {
	docID, err := required("add sheet", "docId", docID)
	if err != nil {
		return Sheet{}, err
	}
	title, err = required("add sheet", "title", title)
	if err != nil {
		return Sheet{}, err
	}
	properties := sheetProperties{Title: title}
	if len(index) > 0 {
		if index[0] < 0 {
			return Sheet{}, errors.New("wecom: add sheet: index must not be negative")
		}
		position := index[0]
		properties.Index = &position
	}
	request := struct {
		DocID      string          `json:"docid"`
		Properties sheetProperties `json:"properties"`
	}{DocID: docID, Properties: properties}
	var response struct {
		Properties Sheet `json:"properties"`
	}
	if err := c.call(ctx, "add sheet", addSheetPath, request, &response); err != nil {
		return Sheet{}, err
	}
	if strings.TrimSpace(response.Properties.SheetID) == "" {
		return Sheet{}, errors.New("wecom: add sheet response is missing sheet_id")
	}
	return response.Properties, nil
}

// UpdateSheet changes the title of one sub-table.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99898
func (c *WeDoc) UpdateSheet(ctx context.Context, docID, sheetID, title string) error {
	docID, err := required("update sheet", "docId", docID)
	if err != nil {
		return err
	}
	sheetID, err = required("update sheet", "sheetId", sheetID)
	if err != nil {
		return err
	}
	title, err = required("update sheet", "title", title)
	if err != nil {
		return err
	}
	request := struct {
		DocID      string          `json:"docid"`
		Properties sheetProperties `json:"properties"`
	}{DocID: docID, Properties: sheetProperties{SheetID: sheetID, Title: title}}
	return c.call(ctx, "update sheet", updateSheetPath, request, nil)
}

// DeleteSheet deletes one sub-table and its views, fields, and records.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99899
func (c *WeDoc) DeleteSheet(ctx context.Context, docID, sheetID string) error {
	docID, err := required("delete sheet", "docId", docID)
	if err != nil {
		return err
	}
	sheetID, err = required("delete sheet", "sheetId", sheetID)
	if err != nil {
		return err
	}
	request := struct {
		DocID   string `json:"docid"`
		SheetID string `json:"sheet_id"`
	}{DocID: docID, SheetID: sheetID}
	return c.call(ctx, "delete sheet", deleteSheetPath, request, nil)
}

// GetSheets returns every sub-table of a smart sheet document. WeCom's
// get_sheet response carries a sheet_list, so one request lists them all.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101154
func (c *WeDoc) GetSheets(ctx context.Context, docID string) ([]Sheet, error) {
	docID, err := required("get sheets", "docId", docID)
	if err != nil {
		return nil, err
	}
	request := struct {
		DocID string `json:"docid"`
	}{DocID: docID}
	var response struct {
		SheetList []Sheet `json:"sheet_list"`
	}
	if err := c.call(ctx, "get sheets", getSheetPath, request, &response); err != nil {
		return nil, err
	}
	return response.SheetList, nil
}
