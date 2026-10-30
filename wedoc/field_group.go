package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	addGroupPath     = smartsheetPath + "/add_field_group"
	updateGroupPath  = smartsheetPath + "/update_field_group"
	deleteGroupsPath = smartsheetPath + "/delete_field_groups"
	getGroupsPath    = smartsheetPath + "/get_field_groups"
)

// FieldGroupChild is one field reference inside a field group.
type FieldGroupChild struct {
	FieldID string `json:"field_id,omitempty"`
}

// FieldGroup is a named group of columns.
type FieldGroup struct {
	FieldGroupID string            `json:"field_group_id,omitempty"`
	Name         string            `json:"name,omitempty"`
	Children     []FieldGroupChild `json:"children,omitempty"`
}

// AddFieldGroup creates a column group. A nil fieldIDs leaves children out;
// an empty non-nil slice sends an empty group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101100
func (c *WeDoc) AddFieldGroup(ctx context.Context, docID, sheetID, name string, fieldIDs []string) (FieldGroup, error) {
	docID, err := required("add field group", "docId", docID)
	if err != nil {
		return FieldGroup{}, err
	}
	sheetID, err = required("add field group", "sheetId", sheetID)
	if err != nil {
		return FieldGroup{}, err
	}
	name, err = required("add field group", "name", name)
	if err != nil {
		return FieldGroup{}, err
	}
	children, err := fieldGroupChildren("add field group", fieldIDs)
	if err != nil {
		return FieldGroup{}, err
	}
	request := struct {
		DocID    string            `json:"docid"`
		SheetID  string            `json:"sheet_id"`
		Name     string            `json:"name"`
		Children []FieldGroupChild `json:"children,omitempty"`
	}{DocID: docID, SheetID: sheetID, Name: name, Children: children}
	var response struct {
		Group FieldGroup `json:"field_group"`
	}
	if err := c.call(ctx, "add field group", addGroupPath, request, &response); err != nil {
		return FieldGroup{}, err
	}
	return response.Group, nil
}

// UpdateFieldGroup renames a group or replaces its columns. A nil fieldIDs
// leaves the column list unchanged; an empty non-nil slice clears it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101101
func (c *WeDoc) UpdateFieldGroup(ctx context.Context, docID, sheetID, groupID, name string, fieldIDs []string) (FieldGroup, error) {
	docID, err := required("update field group", "docId", docID)
	if err != nil {
		return FieldGroup{}, err
	}
	sheetID, err = required("update field group", "sheetId", sheetID)
	if err != nil {
		return FieldGroup{}, err
	}
	groupID, err = required("update field group", "fieldGroupId", groupID)
	if err != nil {
		return FieldGroup{}, err
	}
	children, err := fieldGroupChildren("update field group", fieldIDs)
	if err != nil {
		return FieldGroup{}, err
	}
	var childrenPointer *[]FieldGroupChild
	if fieldIDs != nil {
		childrenPointer = &children
	}
	request := struct {
		DocID    string             `json:"docid"`
		SheetID  string             `json:"sheet_id"`
		GroupID  string             `json:"field_group_id"`
		Name     string             `json:"name,omitempty"`
		Children *[]FieldGroupChild `json:"children,omitempty"`
	}{
		DocID:    docID,
		SheetID:  sheetID,
		GroupID:  groupID,
		Name:     strings.TrimSpace(name),
		Children: childrenPointer,
	}
	var response struct {
		Group FieldGroup `json:"field_group"`
	}
	if err := c.call(ctx, "update field group", updateGroupPath, request, &response); err != nil {
		return FieldGroup{}, err
	}
	return response.Group, nil
}

// DeleteFieldGroups deletes one or more column groups.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101102
func (c *WeDoc) DeleteFieldGroups(ctx context.Context, docID, sheetID string, groupIDs ...string) error {
	docID, err := required("delete field groups", "docId", docID)
	if err != nil {
		return err
	}
	sheetID, err = required("delete field groups", "sheetId", sheetID)
	if err != nil {
		return err
	}
	if len(groupIDs) == 0 {
		return errors.New("wecom: delete field groups: at least one fieldGroupId is required")
	}
	ids := make([]string, 0, len(groupIDs))
	for index, groupID := range groupIDs {
		trimmed := strings.TrimSpace(groupID)
		if trimmed == "" {
			return fmt.Errorf("wecom: delete field groups: fieldGroupIds[%d] is required", index)
		}
		ids = append(ids, trimmed)
	}
	request := struct {
		DocID    string   `json:"docid"`
		SheetID  string   `json:"sheet_id"`
		GroupIDs []string `json:"field_group_ids"`
	}{DocID: docID, SheetID: sheetID, GroupIDs: ids}
	return c.call(ctx, "delete field groups", deleteGroupsPath, request, nil)
}

// GetFieldGroups returns every column group of a sub-table, walking
// offset/limit pages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101103
func (c *WeDoc) GetFieldGroups(ctx context.Context, docID, sheetID string) ([]FieldGroup, error) {
	docID, err := required("get field groups", "docId", docID)
	if err != nil {
		return nil, err
	}
	sheetID, err = required("get field groups", "sheetId", sheetID)
	if err != nil {
		return nil, err
	}
	groups := make([]FieldGroup, 0)
	offset := 0
	for {
		request := struct {
			DocID   string `json:"docid"`
			SheetID string `json:"sheet_id"`
			Offset  int    `json:"offset"`
			Limit   int    `json:"limit"`
		}{DocID: docID, SheetID: sheetID, Offset: offset, Limit: maxListPageSize}
		var page struct {
			Total       uint32       `json:"total"`
			HasMore     bool         `json:"has_more"`
			Next        uint32       `json:"next"`
			FieldGroups []FieldGroup `json:"field_groups"`
		}
		if err := c.call(ctx, "get field groups", getGroupsPath, request, &page); err != nil {
			return nil, err
		}
		groups = append(groups, page.FieldGroups...)
		if len(page.FieldGroups) == 0 {
			return groups, nil
		}
		next := offset + len(page.FieldGroups)
		if page.HasMore || page.Next != 0 {
			if int(page.Next) > offset {
				next = int(page.Next)
			}
		} else if page.Total == 0 || next >= int(page.Total) {
			return groups, nil
		}
		if next <= offset {
			return nil, errors.New("wecom: get field groups pagination did not advance")
		}
		offset = next
	}
}

func fieldGroupChildren(op string, fieldIDs []string) ([]FieldGroupChild, error) {
	if fieldIDs == nil {
		return nil, nil
	}
	children := make([]FieldGroupChild, 0, len(fieldIDs))
	for index, fieldID := range fieldIDs {
		trimmed := strings.TrimSpace(fieldID)
		if trimmed == "" {
			return nil, fmt.Errorf("wecom: %s: fieldIds[%d] is required", op, index)
		}
		children = append(children, FieldGroupChild{FieldID: trimmed})
	}
	return children, nil
}
