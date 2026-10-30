package wedoc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	addFieldsPath    = smartsheetPath + "/add_fields"
	updateFieldsPath = smartsheetPath + "/update_fields"
	deleteFieldsPath = smartsheetPath + "/delete_fields"
	getFieldsPath    = smartsheetPath + "/get_fields"
)

// Smart sheet field types, documented at
// https://developer.work.weixin.qq.com/document/path/99904.
const (
	FieldTypeText         = "FIELD_TYPE_TEXT"
	FieldTypeNumber       = "FIELD_TYPE_NUMBER"
	FieldTypeCheckbox     = "FIELD_TYPE_CHECKBOX"
	FieldTypeDateTime     = "FIELD_TYPE_DATE_TIME"
	FieldTypeImage        = "FIELD_TYPE_IMAGE"
	FieldTypeAttachment   = "FIELD_TYPE_ATTACHMENT"
	FieldTypeUser         = "FIELD_TYPE_USER"
	FieldTypeURL          = "FIELD_TYPE_URL"
	FieldTypeSelect       = "FIELD_TYPE_SELECT"
	FieldTypeCreatedUser  = "FIELD_TYPE_CREATED_USER"
	FieldTypeModifiedUser = "FIELD_TYPE_MODIFIED_USER"
	FieldTypeCreatedTime  = "FIELD_TYPE_CREATED_TIME"
	FieldTypeModifiedTime = "FIELD_TYPE_MODIFIED_TIME"
	FieldTypeProgress     = "FIELD_TYPE_PROGRESS"
	FieldTypePhoneNumber  = "FIELD_TYPE_PHONE_NUMBER"
	FieldTypeEmail        = "FIELD_TYPE_EMAIL"
	FieldTypeSingleSelect = "FIELD_TYPE_SINGLE_SELECT"
	FieldTypeReference    = "FIELD_TYPE_REFERENCE"
	FieldTypeLocation     = "FIELD_TYPE_LOCATION"
	FieldTypeCurrency     = "FIELD_TYPE_CURRENCY"
	FieldTypeWWGroup      = "FIELD_TYPE_WWGROUP"
	FieldTypePercentage   = "FIELD_TYPE_PERCENTAGE"
	FieldTypeBarcode      = "FIELD_TYPE_BARCODE"
	FieldTypeAutoNumber   = "FIELD_TYPE_AUTONUMBER"
	FieldTypeFormula      = "FIELD_TYPE_FORMULA"
)

// fieldPropertyKeys maps a field type to the property key WeCom expects when
// that field is created. A type without a property object needs no entry.
var fieldPropertyKeys = map[string]string{
	FieldTypeNumber:       "property_number",
	FieldTypeCheckbox:     "property_checkbox",
	FieldTypeDateTime:     "property_date_time",
	FieldTypeAttachment:   "property_attachment",
	FieldTypeUser:         "property_user",
	FieldTypeURL:          "property_url",
	FieldTypeSelect:       "property_select",
	FieldTypeCreatedTime:  "property_created_time",
	FieldTypeModifiedTime: "property_modified_time",
	FieldTypeProgress:     "property_progress",
	FieldTypeSingleSelect: "property_single_select",
	FieldTypeReference:    "property_reference",
	FieldTypeLocation:     "property_location",
	FieldTypeCurrency:     "property_currency",
	FieldTypeWWGroup:      "property_ww_group",
	FieldTypePercentage:   "property_percentage",
	FieldTypeBarcode:      "property_barcode",
	FieldTypeAutoNumber:   "property_auto_number"}

// Field describes one smart sheet column. Property is the type-specific
// property object documented for FieldTypeNumber, FieldTypeSelect, and so on;
// it is encoded under the matching property_* key.
type Field struct {
	FieldID    string `json:"field_id,omitempty"`
	FieldTitle string `json:"field_title,omitempty"`
	FieldType  string `json:"field_type,omitempty"`
	Property   any    `json:"-"`
}

// MarshalJSON encodes the field and, when set, its type-specific property under
// the key WeCom expects (property_number, property_select, ...).
func (f Field) MarshalJSON() ([]byte, error) {
	values := make(map[string]any, 4)
	if f.FieldTitle != "" {
		values["field_title"] = f.FieldTitle
	}
	if f.FieldType != "" {
		values["field_type"] = f.FieldType
	}
	if f.FieldID != "" {
		values["field_id"] = f.FieldID
	}
	if f.Property != nil {
		values[fieldPropertyKey(f.FieldType)] = f.Property
	}
	return json.Marshal(values)
}

// UnmarshalJSON captures the type-specific property object returned by
// get_fields, which lives under a dynamic property_* key.
func (f *Field) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.FieldID, f.FieldTitle, f.FieldType, f.Property = "", "", "", nil
	for key, value := range raw {
		switch key {
		case "field_id":
			if err := json.Unmarshal(value, &f.FieldID); err != nil {
				return err
			}
		case "field_title":
			if err := json.Unmarshal(value, &f.FieldTitle); err != nil {
				return err
			}
		case "field_type":
			if err := json.Unmarshal(value, &f.FieldType); err != nil {
				return err
			}
		default:
			if !strings.HasPrefix(key, "property_") {
				continue
			}
			var property map[string]any
			if err := json.Unmarshal(value, &property); err != nil {
				return err
			}
			f.Property = property
		}
	}
	return nil
}

func fieldPropertyKey(fieldType string) string {
	if key, ok := fieldPropertyKeys[fieldType]; ok {
		return key
	}
	return "property"
}

// AddField adds exactly one column and returns the created column. It is a
// convenience wrapper around AddFields.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99904
func (c *WeDoc) AddField(ctx context.Context, docID, sheetID string, field Field) (Field, error) {
	fields, err := c.AddFields(ctx, docID, sheetID, field)
	if err != nil {
		return Field{}, err
	}
	if len(fields) == 0 {
		return Field{}, errors.New("wecom: add field response is empty")
	}
	return fields[0], nil
}

// AddFields adds one or more columns to a sub-table.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99904
func (c *WeDoc) AddFields(ctx context.Context, docID, sheetID string, fields ...Field) ([]Field, error) {
	docID, err := required("add fields", "docId", docID)
	if err != nil {
		return nil, err
	}
	sheetID, err = required("add fields", "sheetId", sheetID)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, errors.New("wecom: add fields: at least one field is required")
	}
	for index, field := range fields {
		if strings.TrimSpace(field.FieldTitle) == "" {
			return nil, fmt.Errorf("wecom: add fields: fields[%d].fieldTitle is required", index)
		}
		if strings.TrimSpace(field.FieldType) == "" {
			return nil, fmt.Errorf("wecom: add fields: fields[%d].fieldType is required", index)
		}
	}
	request := struct {
		DocID   string  `json:"docid"`
		SheetID string  `json:"sheet_id"`
		Fields  []Field `json:"fields"`
	}{DocID: docID, SheetID: sheetID, Fields: fields}
	var response struct {
		Fields []Field `json:"fields"`
	}
	if err := c.call(ctx, "add fields", addFieldsPath, request, &response); err != nil {
		return nil, err
	}
	return response.Fields, nil
}

// UpdateFields changes the titles or properties of existing columns. The field
// type must stay the same as the original column.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99906
func (c *WeDoc) UpdateFields(ctx context.Context, docID, sheetID string, fields ...Field) ([]Field, error) {
	docID, err := required("update fields", "docId", docID)
	if err != nil {
		return nil, err
	}
	sheetID, err = required("update fields", "sheetId", sheetID)
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, errors.New("wecom: update fields: at least one field is required")
	}
	for index, field := range fields {
		if strings.TrimSpace(field.FieldID) == "" {
			return nil, fmt.Errorf("wecom: update fields: fields[%d].fieldId is required", index)
		}
		if strings.TrimSpace(field.FieldType) == "" {
			return nil, fmt.Errorf("wecom: update fields: fields[%d].fieldType is required", index)
		}
	}
	request := struct {
		DocID   string  `json:"docid"`
		SheetID string  `json:"sheet_id"`
		Fields  []Field `json:"fields"`
	}{DocID: docID, SheetID: sheetID, Fields: fields}
	var response struct {
		Fields []Field `json:"fields"`
	}
	if err := c.call(ctx, "update fields", updateFieldsPath, request, &response); err != nil {
		return nil, err
	}
	return response.Fields, nil
}

// DeleteFields deletes one or more columns from a sub-table.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99905
func (c *WeDoc) DeleteFields(ctx context.Context, docID, sheetID string, fieldIDs ...string) error {
	docID, err := required("delete fields", "docId", docID)
	if err != nil {
		return err
	}
	sheetID, err = required("delete fields", "sheetId", sheetID)
	if err != nil {
		return err
	}
	if len(fieldIDs) == 0 {
		return errors.New("wecom: delete fields: at least one fieldId is required")
	}
	ids := make([]string, 0, len(fieldIDs))
	for index, fieldID := range fieldIDs {
		trimmed := strings.TrimSpace(fieldID)
		if trimmed == "" {
			return fmt.Errorf("wecom: delete fields: fieldIds[%d] is required", index)
		}
		ids = append(ids, trimmed)
	}
	request := struct {
		DocID    string   `json:"docid"`
		SheetID  string   `json:"sheet_id"`
		FieldIDs []string `json:"field_ids"`
	}{DocID: docID, SheetID: sheetID, FieldIDs: ids}
	return c.call(ctx, "delete fields", deleteFieldsPath, request, nil)
}

// GetFields returns every column of a sub-table, walking offset/limit pages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101157
func (c *WeDoc) GetFields(ctx context.Context, docID, sheetID string) ([]Field, error) {
	docID, err := required("get fields", "docId", docID)
	if err != nil {
		return nil, err
	}
	sheetID, err = required("get fields", "sheetId", sheetID)
	if err != nil {
		return nil, err
	}
	fields := make([]Field, 0)
	offset := 0
	for {
		request := struct {
			DocID   string `json:"docid"`
			SheetID string `json:"sheet_id"`
			Offset  int    `json:"offset"`
			Limit   int    `json:"limit"`
		}{DocID: docID, SheetID: sheetID, Offset: offset, Limit: maxListPageSize}
		var page struct {
			Total  uint32  `json:"total"`
			Fields []Field `json:"fields"`
		}
		if err := c.call(ctx, "get fields", getFieldsPath, request, &page); err != nil {
			return nil, err
		}
		fields = append(fields, page.Fields...)
		if len(page.Fields) == 0 {
			return fields, nil
		}
		offset += len(page.Fields)
		if page.Total > 0 {
			if offset >= int(page.Total) {
				return fields, nil
			}
		} else if len(page.Fields) < maxListPageSize {
			return fields, nil
		}
	}
}
