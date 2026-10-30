package contact

import (
	"context"
	"errors"
)

// RuleTypeHide hides departments, members and tags from the visible range.
const RuleTypeHide = 1

// RuleRange selects members by id, department or tag.
type RuleRange struct {
	UserIDs  []string `json:"userid,omitempty"`
	PartyIDs []int    `json:"partyid,omitempty"`
	TagIDs   []int    `json:"tagid,omitempty"`
}

// Rule is one contact visibility rule. Whitelist is exempt from the rule and
// Exclude lists entries the rule does not hide.
type Rule struct {
	ID                    int        `json:"rule_id,omitempty"`
	Type                  int        `json:"rule_type,omitempty"`
	Range                 RuleRange  `json:"range"`
	Whitelist             *RuleRange `json:"whitelist,omitempty"`
	Exclude               *RuleRange `json:"exclude,omitempty"`
	IsAllowedSearch       *bool      `json:"is_allowed_search,omitempty"`
	IsAllowedConversation *bool      `json:"is_allowed_conversation,omitempty"`
}

// CreateRules adds visibility rules and returns their ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101536
func (c *Contact) CreateRules(ctx context.Context, rules []Rule) ([]int, error) {
	if len(rules) == 0 {
		return nil, errors.New("wecom: create rules: at least one rule is required")
	}
	for i, rule := range rules {
		if rule.Type == 0 {
			return nil, errors.New("wecom: create rules: rule_type is required on every rule")
		}
		if emptyRange(rule.Range) {
			return nil, errors.New("wecom: create rules: rule " + itoa(i) + " has an empty range")
		}
	}
	request := struct {
		Rules []Rule `json:"rules"`
	}{Rules: rules}
	var out struct {
		IDs []int `json:"rule_ids"`
	}
	err := c.call(ctx, "contact create rules", "/cgi-bin/contactrule/create", request, &out)
	return out.IDs, err
}

// ListRules returns every visibility rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101537
func (c *Contact) ListRules(ctx context.Context) ([]Rule, error) {
	var out struct {
		Rules []Rule `json:"rules"`
	}
	err := c.call(ctx, "contact list rules", "/cgi-bin/contactrule/list", nil, &out)
	return out.Rules, err
}

// UpdateRules changes existing rules; every entry needs a rule_id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101539
func (c *Contact) UpdateRules(ctx context.Context, rules []Rule) error {
	if len(rules) == 0 {
		return errors.New("wecom: update rules: at least one rule is required")
	}
	for _, rule := range rules {
		if rule.ID <= 0 {
			return errors.New("wecom: update rules: rule_id is required on every rule")
		}
	}
	request := struct {
		Rules []Rule `json:"rules"`
	}{Rules: rules}
	return c.call(ctx, "contact update rules", "/cgi-bin/contactrule/update", request, nil)
}

// DeleteRules removes rules by id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101540
func (c *Contact) DeleteRules(ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return errors.New("wecom: delete rules: rule_ids is empty")
	}
	request := struct {
		IDs []int `json:"rule_ids"`
	}{IDs: ids}
	return c.call(ctx, "contact delete rules", "/cgi-bin/contactrule/delete", request, nil)
}

func emptyRange(r RuleRange) bool {
	return len(r.UserIDs) == 0 && len(r.PartyIDs) == 0 && len(r.TagIDs) == 0
}
