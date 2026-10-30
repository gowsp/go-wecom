package external

import (
	"context"
	"errors"
	"strings"
)

// StrategyRange selects which members a rule group applies to. Type 1 is a
// department, 2 a member.
type StrategyRange struct {
	Type    int      `json:"type,omitempty"`
	UserID  []string `json:"userid,omitempty"`
	PartyID []int    `json:"partyid,omitempty"`
}

// StrategyRef is the minimal record returned by the rule group list.
type StrategyRef struct {
	StrategyID int `json:"strategy_id"`
}

// StrategyPage is one page of rule group ids.
type StrategyPage struct {
	NextCursor string        `json:"next_cursor"`
	Strategies []StrategyRef `json:"strategy"`
}

// RangePage is one page of rule group ranges.
type RangePage struct {
	NextCursor string          `json:"next_cursor"`
	Range      []StrategyRange `json:"range"`
}

// CustomerPrivilege lists the customer-contact permissions of a rule group.
// Every field is always sent so an update can revoke a permission explicitly.
type CustomerPrivilege struct {
	ViewCustomerList        bool `json:"view_customer_list"`
	ViewCustomerData        bool `json:"view_customer_data"`
	ViewRoomList            bool `json:"view_room_list"`
	ContactMe               bool `json:"contact_me"`
	JoinRoom                bool `json:"join_room"`
	ShareCustomer           bool `json:"share_customer"`
	OperResignCustomer      bool `json:"oper_resign_customer"`
	OperResignGroup         bool `json:"oper_resign_group"`
	SendCustomerMsg         bool `json:"send_customer_msg"`
	EditWelcomeMsg          bool `json:"edit_welcome_msg"`
	ViewBehaviorData        bool `json:"view_behavior_data"`
	SendGroupMsg            bool `json:"send_group_msg"`
	RoomDeduplication       bool `json:"room_deduplication"`
	RapidReply              bool `json:"rapid_reply"`
	OnJobCustomerTransfer   bool `json:"onjob_customer_transfer"`
	EditAntiSpamRule        bool `json:"edit_anti_spam_rule"`
	ExportCustomerList      bool `json:"export_customer_list"`
	ExportCustomerData      bool `json:"export_customer_data"`
	ExportCustomerGroupList bool `json:"export_customer_group_list"`
	ManageCustomerTag       bool `json:"manage_customer_tag"`
}

// CustomerStrategy is one customer-contact rule group.
type CustomerStrategy struct {
	StrategyID int                `json:"strategy_id,omitempty"`
	ParentID   int                `json:"parent_id,omitempty"`
	Name       string             `json:"strategy_name,omitempty"`
	CreateTime int64              `json:"create_time,omitempty"`
	AdminList  []string           `json:"admin_list,omitempty"`
	Privilege  *CustomerPrivilege `json:"privilege,omitempty"`
}

// ListCustomerStrategies returns one page of customer rule group ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) ListCustomerStrategies(ctx context.Context, cursor string, limit int) (StrategyPage, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out StrategyPage
	err := c.call(ctx, "external list customer strategies",
		"/cgi-bin/externalcontact/customer_strategy/list", request, &out)
	return out, err
}

// GetCustomerStrategy reads one customer rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) GetCustomerStrategy(ctx context.Context, strategyID int) (CustomerStrategy, error) {
	if strategyID <= 0 {
		return CustomerStrategy{}, errors.New("wecom: get customer strategy: strategy_id is required")
	}
	var out struct {
		Strategy CustomerStrategy `json:"strategy"`
	}
	err := c.call(ctx, "external get customer strategy",
		"/cgi-bin/externalcontact/customer_strategy/get",
		map[string]int{"strategy_id": strategyID}, &out)
	return out.Strategy, err
}

// CreateCustomerStrategy creates one customer rule group and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) CreateCustomerStrategy(ctx context.Context, strategy CustomerStrategy) (int, error) {
	if strings.TrimSpace(strategy.Name) == "" {
		return 0, errors.New("wecom: create customer strategy: strategy_name is required")
	}
	var out struct {
		StrategyID int `json:"strategy_id"`
	}
	err := c.call(ctx, "external create customer strategy",
		"/cgi-bin/externalcontact/customer_strategy/create", strategy, &out)
	return out.StrategyID, err
}

// UpdateCustomerStrategy changes one customer rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) UpdateCustomerStrategy(ctx context.Context, strategy CustomerStrategy) error {
	if strategy.StrategyID <= 0 {
		return errors.New("wecom: update customer strategy: strategy_id is required")
	}
	return c.call(ctx, "external update customer strategy",
		"/cgi-bin/externalcontact/customer_strategy/edit", strategy, nil)
}

// DeleteCustomerStrategy removes one customer rule group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) DeleteCustomerStrategy(ctx context.Context, strategyID int) error {
	if strategyID <= 0 {
		return errors.New("wecom: delete customer strategy: strategy_id is required")
	}
	return c.call(ctx, "external delete customer strategy",
		"/cgi-bin/externalcontact/customer_strategy/del",
		map[string]int{"strategy_id": strategyID}, nil)
}

// CustomerStrategyRange returns the members one rule group applies to.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94883
func (c *External) CustomerStrategyRange(ctx context.Context, strategyID int, cursor string, limit int) (RangePage, error) {
	if strategyID <= 0 {
		return RangePage{}, errors.New("wecom: customer strategy range: strategy_id is required")
	}
	request := struct {
		StrategyID int    `json:"strategy_id"`
		Cursor     string `json:"cursor,omitempty"`
		Limit      int    `json:"limit,omitempty"`
	}{StrategyID: strategyID, Cursor: cursor, Limit: limit}
	var out RangePage
	err := c.call(ctx, "external customer strategy range",
		"/cgi-bin/externalcontact/customer_strategy/get_range", request, &out)
	return out, err
}
