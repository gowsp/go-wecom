package external

import (
	"context"
	"errors"
	"strings"
)

// MomentPrivilege lists the moments permissions of a rule group.
type MomentPrivilege struct {
	ViewMomentList           bool `json:"view_moment_list"`
	SendMoment               bool `json:"send_moment"`
	ManageMomentCoverAndSign bool `json:"manage_moment_cover_and_sign"`
}

// MomentStrategy is one moments rule group.
type MomentStrategy struct {
	StrategyID int              `json:"strategy_id,omitempty"`
	ParentID   int              `json:"parent_id,omitempty"`
	Name       string           `json:"strategy_name,omitempty"`
	CreateTime int64            `json:"create_time,omitempty"`
	AdminList  []string         `json:"admin_list,omitempty"`
	Privilege  *MomentPrivilege `json:"privilege,omitempty"`
}

// ListMomentStrategies returns one page of moments rule group ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) ListMomentStrategies(ctx context.Context, cursor string, limit int) (StrategyPage, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out StrategyPage
	err := c.call(ctx, "external list moment strategies",
		"/cgi-bin/externalcontact/moment_strategy/list", request, &out)
	return out, err
}

// GetMomentStrategy reads one moments rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) GetMomentStrategy(ctx context.Context, strategyID int) (MomentStrategy, error) {
	if strategyID <= 0 {
		return MomentStrategy{}, errors.New("wecom: get moment strategy: strategy_id is required")
	}
	var out struct {
		Strategy MomentStrategy `json:"strategy"`
	}
	err := c.call(ctx, "external get moment strategy",
		"/cgi-bin/externalcontact/moment_strategy/get",
		map[string]int{"strategy_id": strategyID}, &out)
	return out.Strategy, err
}

// CreateMomentStrategy creates one moments rule group and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) CreateMomentStrategy(ctx context.Context, strategy MomentStrategy) (int, error) {
	if strings.TrimSpace(strategy.Name) == "" {
		return 0, errors.New("wecom: create moment strategy: strategy_name is required")
	}
	var out struct {
		StrategyID int `json:"strategy_id"`
	}
	err := c.call(ctx, "external create moment strategy",
		"/cgi-bin/externalcontact/moment_strategy/create", strategy, &out)
	return out.StrategyID, err
}

// UpdateMomentStrategy changes one moments rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) UpdateMomentStrategy(ctx context.Context, strategy MomentStrategy) error {
	if strategy.StrategyID <= 0 {
		return errors.New("wecom: update moment strategy: strategy_id is required")
	}
	return c.call(ctx, "external update moment strategy",
		"/cgi-bin/externalcontact/moment_strategy/edit", strategy, nil)
}

// DeleteMomentStrategy removes one moments rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) DeleteMomentStrategy(ctx context.Context, strategyID int) error {
	if strategyID <= 0 {
		return errors.New("wecom: delete moment strategy: strategy_id is required")
	}
	return c.call(ctx, "external delete moment strategy",
		"/cgi-bin/externalcontact/moment_strategy/del",
		map[string]int{"strategy_id": strategyID}, nil)
}

// MomentStrategyRange returns the members one moments rule group applies to.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94890
func (c *External) MomentStrategyRange(ctx context.Context, strategyID int, cursor string, limit int) (RangePage, error) {
	if strategyID <= 0 {
		return RangePage{}, errors.New("wecom: moment strategy range: strategy_id is required")
	}
	request := struct {
		StrategyID int    `json:"strategy_id"`
		Cursor     string `json:"cursor,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{StrategyID: strategyID, Cursor: cursor, Limit: limit}
	var out RangePage
	err := c.call(ctx, "external moment strategy range",
		"/cgi-bin/externalcontact/moment_strategy/get_range", request, &out)
	return out, err
}

// ListStrategyTags returns the tags belonging to one rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94882
func (c *External) ListStrategyTags(ctx context.Context, strategyID int, groupID, tagID, cursor string, limit int) ([]CorpTagGroup, error) {
	if strategyID <= 0 {
		return nil, errors.New("wecom: list strategy tags: strategy_id is required")
	}
	request := struct {
		StrategyID int      `json:"strategy_id"`
		GroupIDs   []string `json:"group_id,omitempty"`
		TagIDs     []string `json:"tag_id,omitempty"`
		Cursor     string   `json:"cursor,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{StrategyID: strategyID, Cursor: cursor, Limit: limit}
	if groupID != "" {
		request.GroupIDs = []string{groupID}
	}
	if tagID != "" {
		request.TagIDs = []string{tagID}
	}
	var out struct {
		TagGroups []CorpTagGroup `json:"tag_group"`
	}
	err := c.call(ctx, "external list strategy tags",
		"/cgi-bin/externalcontact/get_strategy_tag_list", request, &out)
	return out.TagGroups, err
}

// AddStrategyTags creates tags under one rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94882
func (c *External) AddStrategyTags(ctx context.Context, strategyID int, groupID, groupName string, tags []CorpTag) ([]CorpTagGroup, error) {
	if strategyID <= 0 {
		return nil, errors.New("wecom: add strategy tags: strategy_id is required")
	}
	if len(tags) == 0 && strings.TrimSpace(groupName) == "" {
		return nil, errors.New("wecom: add strategy tags: tags or group name is required")
	}
	request := struct {
		StrategyID int       `json:"strategy_id"`
		GroupID    string    `json:"group_id,omitempty"`
		GroupName  string    `json:"group_name,omitempty"`
		Tags       []CorpTag `json:"tag,omitempty"`
	}{StrategyID: strategyID, GroupID: groupID, GroupName: groupName, Tags: tags}
	var out struct {
		TagGroups []CorpTagGroup `json:"tag_group"`
	}
	err := c.call(ctx, "external add strategy tags",
		"/cgi-bin/externalcontact/add_strategy_tag", request, &out)
	return out.TagGroups, err
}

// EditStrategyTag renames one rule group tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94882
func (c *External) EditStrategyTag(ctx context.Context, id, name string, order int) error {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(name) == "" {
		return errors.New("wecom: edit strategy tag: id and name are required")
	}
	request := struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Order int    `json:"order,omitempty"`
	}{ID: id, Name: name, Order: order}
	return c.call(ctx, "external edit strategy tag",
		"/cgi-bin/externalcontact/edit_strategy_tag", request, nil)
}

// DeleteStrategyTags removes tags from one rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94882
func (c *External) DeleteStrategyTags(ctx context.Context, strategyID int, ids []string) error {
	if strategyID <= 0 || len(ids) == 0 {
		return errors.New("wecom: delete strategy tags: strategy_id and id are required")
	}
	request := struct {
		StrategyID int      `json:"strategy_id"`
		IDs        []string `json:"id"`
	}{StrategyID: strategyID, IDs: ids}
	return c.call(ctx, "external delete strategy tags",
		"/cgi-bin/externalcontact/del_strategy_tag", request, nil)
}
