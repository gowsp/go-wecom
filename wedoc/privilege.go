package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	getSheetPrivPath    = smartsheetPath + "/content_priv/get_sheet_priv"
	updateSheetPrivPath = smartsheetPath + "/content_priv/update_sheet_priv"
	createRulePath      = smartsheetPath + "/content_priv/create_rule"
	modRuleMemberPath   = smartsheetPath + "/content_priv/mod_rule_member"
	deleteRulePath      = smartsheetPath + "/content_priv/delete_rule"
)

// Content permission rule types and sheet privilege levels.
const (
	ContentRuleAll   = 1
	ContentRuleExtra = 2

	SheetPrivFull = 1
	SheetPrivEdit = 2
	SheetPrivView = 3
	SheetPrivNone = 4

	FieldRangeAll     = 1
	FieldRangePartial = 2

	RecordRangeAll          = 1
	RecordRangeAnyCondition = 2
	RecordRangeAllCondition = 3
)

// ContentRule is one smart sheet content permission rule.
type ContentRule struct {
	RuleID   uint32           `json:"rule_id,omitempty"`
	Type     int              `json:"type"`
	Name     string           `json:"name,omitempty"`
	PrivList []SheetPrivilege `json:"priv_list,omitempty"`
}

// SheetPrivilege is the content permission of one sub-table.
type SheetPrivilege struct {
	SheetID                   string           `json:"sheet_id"`
	Priv                      int              `json:"priv"`
	CanInsertRecord           *bool            `json:"can_insert_record,omitempty"`
	CanDeleteRecord           *bool            `json:"can_delete_record,omitempty"`
	CanCreateModifyDeleteView *bool            `json:"can_create_modify_delete_view,omitempty"`
	FieldPriv                 *FieldPrivilege  `json:"field_priv,omitempty"`
	RecordPriv                *RecordPrivilege `json:"record_priv,omitempty"`
	Clear                     bool             `json:"clear,omitempty"`
}

// FieldPrivilege configures field-level permissions for a sub-table.
type FieldPrivilege struct {
	FieldRangeType   int                  `json:"field_range_type"`
	FieldRuleList    []FieldRule          `json:"field_rule_list,omitempty"`
	FieldDefaultRule *FieldRulePermission `json:"field_default_rule,omitempty"`
}

// FieldRule is the permission of one field.
type FieldRule struct {
	FieldID   string `json:"field_id"`
	FieldType string `json:"field_type,omitempty"`
	CanEdit   bool   `json:"can_edit"`
	CanInsert bool   `json:"can_insert"`
	CanView   bool   `json:"can_view"`
}

// FieldRulePermission is the default field permission.
type FieldRulePermission struct {
	CanEdit   bool `json:"can_edit"`
	CanInsert bool `json:"can_insert"`
	CanView   bool `json:"can_view"`
}

// RecordPrivilege configures record-level permissions for a sub-table.
type RecordPrivilege struct {
	RecordRangeType int          `json:"record_range_type"`
	RecordRuleList  []RecordRule `json:"record_rule_list,omitempty"`
	OtherPriv       int          `json:"other_priv,omitempty"`
}

// RecordRule is one record-permission condition.
type RecordRule struct {
	FieldID   string   `json:"field_id"`
	FieldType string   `json:"field_type,omitempty"`
	OperType  int      `json:"oper_type"`
	Value     []string `json:"value,omitempty"`
}

// GetContentPrivileges queries content permission rules. ruleType is
// ContentRuleAll or ContentRuleExtra; ruleIDs filters extra rules.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99935
func (c *WeDoc) GetContentPrivileges(ctx context.Context, docID string, ruleType int, ruleIDs []uint32) ([]ContentRule, error) {
	docID, err := required("get content privileges", "docId", docID)
	if err != nil {
		return nil, err
	}
	if ruleType != ContentRuleAll && ruleType != ContentRuleExtra {
		return nil, errors.New("wecom: get content privileges: ruleType must be 1 or 2")
	}
	request := struct {
		DocID      string   `json:"docid"`
		Type       int      `json:"type"`
		RuleIDList []uint32 `json:"rule_id_list,omitempty"`
	}{DocID: docID, Type: ruleType, RuleIDList: ruleIDs}
	var response struct {
		RuleList []ContentRule `json:"rule_list"`
	}
	if err := c.call(ctx, "get content privileges", getSheetPrivPath, request, &response); err != nil {
		return nil, err
	}
	return response.RuleList, nil
}

// UpdateContentPrivilege sets the sub-table permissions of an existing rule.
// For ContentRuleExtra the rule id is required.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99935
func (c *WeDoc) UpdateContentPrivilege(ctx context.Context, docID string, rule ContentRule) error {
	docID, err := required("update content privilege", "docId", docID)
	if err != nil {
		return err
	}
	if rule.Type != ContentRuleAll && rule.Type != ContentRuleExtra {
		return errors.New("wecom: update content privilege: rule type must be 1 or 2")
	}
	if rule.Type == ContentRuleExtra && rule.RuleID == 0 {
		return errors.New("wecom: update content privilege: ruleId is required for extra rules")
	}
	for index, privilege := range rule.PrivList {
		if strings.TrimSpace(privilege.SheetID) == "" {
			return fmt.Errorf("wecom: update content privilege: privList[%d].sheetId is required", index)
		}
		if privilege.Priv < SheetPrivFull || privilege.Priv > SheetPrivNone {
			return fmt.Errorf("wecom: update content privilege: privList[%d].priv must be 1, 2, 3, or 4", index)
		}
	}
	request := struct {
		DocID    string           `json:"docid"`
		Type     int              `json:"type"`
		RuleID   uint32           `json:"rule_id,omitempty"`
		Name     string           `json:"name,omitempty"`
		PrivList []SheetPrivilege `json:"priv_list,omitempty"`
	}{
		DocID:    docID,
		Type:     rule.Type,
		RuleID:   rule.RuleID,
		Name:     strings.TrimSpace(rule.Name),
		PrivList: rule.PrivList,
	}
	return c.call(ctx, "update content privilege", updateSheetPrivPath, request, nil)
}

// CreateContentPrivilegeRule creates an extra member permission rule and
// returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99935
func (c *WeDoc) CreateContentPrivilegeRule(ctx context.Context, docID, name string) (uint32, error) {
	docID, err := required("create content privilege rule", "docId", docID)
	if err != nil {
		return 0, err
	}
	name, err = required("create content privilege rule", "name", name)
	if err != nil {
		return 0, err
	}
	request := struct {
		DocID string `json:"docid"`
		Name  string `json:"name"`
	}{DocID: docID, Name: name}
	var response struct {
		RuleID uint32 `json:"rule_id"`
	}
	if err := c.call(ctx, "create content privilege rule", createRulePath, request, &response); err != nil {
		return 0, err
	}
	if response.RuleID == 0 {
		return 0, errors.New("wecom: create content privilege rule response is missing rule_id")
	}
	return response.RuleID, nil
}

// ModifyContentPrivilegeMembers adds and removes members of an extra rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99935
func (c *WeDoc) ModifyContentPrivilegeMembers(ctx context.Context, docID string, ruleID uint32, addUserIDs, removeUserIDs []string) error {
	docID, err := required("modify content privilege members", "docId", docID)
	if err != nil {
		return err
	}
	if ruleID == 0 {
		return errors.New("wecom: modify content privilege members: ruleId is required")
	}
	add := normalizeUserIDs(addUserIDs)
	remove := normalizeUserIDs(removeUserIDs)
	if len(add) == 0 && len(remove) == 0 {
		return errors.New("wecom: modify content privilege members: at least one member change is required")
	}
	type memberRange struct {
		UserIDs []string `json:"userid_list,omitempty"`
	}
	request := struct {
		DocID          string       `json:"docid"`
		RuleID         uint32       `json:"rule_id"`
		AddMemberRange *memberRange `json:"add_member_range,omitempty"`
		DelMemberRange *memberRange `json:"del_member_range,omitempty"`
	}{DocID: docID, RuleID: ruleID}
	if len(add) > 0 {
		request.AddMemberRange = &memberRange{UserIDs: add}
	}
	if len(remove) > 0 {
		request.DelMemberRange = &memberRange{UserIDs: remove}
	}
	return c.call(ctx, "modify content privilege members", modRuleMemberPath, request, nil)
}

// DeleteContentPrivilegeRules deletes extra member permission rules.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99935
func (c *WeDoc) DeleteContentPrivilegeRules(ctx context.Context, docID string, ruleIDs ...uint32) error {
	docID, err := required("delete content privilege rules", "docId", docID)
	if err != nil {
		return err
	}
	if len(ruleIDs) == 0 {
		return errors.New("wecom: delete content privilege rules: at least one ruleId is required")
	}
	for index, ruleID := range ruleIDs {
		if ruleID == 0 {
			return fmt.Errorf("wecom: delete content privilege rules: ruleIds[%d] is required", index)
		}
	}
	request := struct {
		DocID      string   `json:"docid"`
		RuleIDList []uint32 `json:"rule_id_list"`
	}{DocID: docID, RuleIDList: ruleIDs}
	return c.call(ctx, "delete content privilege rules", deleteRulePath, request, nil)
}
