package external

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// TransferRequest hands customers from one member to another.
type TransferRequest struct {
	HandoverUserID     string   `json:"handover_userid"`
	TakeoverUserID     string   `json:"takeover_userid"`
	ExternalUserID     []string `json:"external_userid"`
	TransferSuccessMsg string   `json:"transfer_success_msg,omitempty"`
}

// TransferFailure is one customer that could not be handed over.
type TransferFailure struct {
	ExternalUserID string `json:"external_userid"`
	Errcode        int    `json:"errcode"`
}

// TransferStatus is the state of one handover: 1 waiting, 2 succeeded,
// 3 the customer refused, 4 the member left.
type TransferStatus struct {
	ExternalUserID string `json:"external_userid"`
	Status         int    `json:"status"`
	TakeoverTime   int64  `json:"takeover_time"`
}

// TransferPage is one page of handover results.
type TransferPage struct {
	NextCursor string           `json:"next_cursor"`
	Customer   []TransferStatus `json:"customer"`
}

// TransferCustomers hands over customers while both members are employed.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92125
func (c *External) TransferCustomers(ctx context.Context, request TransferRequest) ([]TransferFailure, error) {
	if strings.TrimSpace(request.HandoverUserID) == "" || strings.TrimSpace(request.TakeoverUserID) == "" {
		return nil, errors.New("wecom: transfer customers: handover and takeover userid are required")
	}
	if len(request.ExternalUserID) == 0 {
		return nil, errors.New("wecom: transfer customers: external_userid is empty")
	}
	var out struct {
		Customer []TransferFailure `json:"customer"`
	}
	err := c.call(ctx, "external transfer customers", "/cgi-bin/externalcontact/transfer_customer", request, &out)
	return out.Customer, err
}

// TransferResult lists the handover state of one member pair.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94088
func (c *External) TransferResult(ctx context.Context, handoverUserID, takeoverUserID, cursor string) (TransferPage, error) {
	if strings.TrimSpace(handoverUserID) == "" || strings.TrimSpace(takeoverUserID) == "" {
		return TransferPage{}, errors.New("wecom: transfer result: handover and takeover userid are required")
	}
	query := url.Values{
		"handover_userid": {handoverUserID},
		"takeover_userid": {takeoverUserID},
	}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	var out TransferPage
	err := c.get(ctx, "external transfer result", "/cgi-bin/externalcontact/transfer_result", query, &out)
	return out, err
}

// ChatFailure is one group chat that could not be transferred.
type ChatFailure struct {
	ChatID  string `json:"chat_id"`
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

// TransferGroupChats hands the group chats of a departed member to a new owner.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92127
func (c *External) TransferGroupChats(ctx context.Context, chatIDs []string, newOwner string) ([]ChatFailure, error) {
	return c.transferChats(ctx, "external transfer group chats",
		"/cgi-bin/externalcontact/groupchat/transfer", chatIDs, newOwner)
}

// TransferOnJobGroupChats hands group chats between two employed members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95703
func (c *External) TransferOnJobGroupChats(ctx context.Context, chatIDs []string, newOwner string) ([]ChatFailure, error) {
	return c.transferChats(ctx, "external transfer onjob group chats",
		"/cgi-bin/externalcontact/groupchat/onjob_transfer", chatIDs, newOwner)
}

func (c *External) transferChats(ctx context.Context, op, path string, chatIDs []string, newOwner string) ([]ChatFailure, error) {
	if len(chatIDs) == 0 {
		return nil, errors.New("wecom: " + op + ": chat_id_list is empty")
	}
	if strings.TrimSpace(newOwner) == "" {
		return nil, errors.New("wecom: " + op + ": new owner is required")
	}
	request := struct {
		ChatIDList []string `json:"chat_id_list"`
		NewOwner   string   `json:"new_owner"`
	}{ChatIDList: chatIDs, NewOwner: newOwner}
	var out struct {
		Failed []ChatFailure `json:"failed_chat_list"`
	}
	err := c.call(ctx, op, path, request, &out)
	return out.Failed, err
}

// UnassignedCustomer is one customer left behind by a departed member.
type UnassignedCustomer struct {
	HandoverUserID string `json:"handover_userid"`
	ExternalUserID string `json:"external_userid"`
	DimissionTime  int64  `json:"dimission_time"`
}

// UnassignedPage is one page of customers waiting for reassignment.
type UnassignedPage struct {
	NextCursor string               `json:"next_cursor"`
	Info       []UnassignedCustomer `json:"info"`
}

// ListUnassignedCustomers returns the customers of departed members that are
// still waiting for a new owner.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92124
func (c *External) ListUnassignedCustomers(ctx context.Context, cursor string, pageSize int) (UnassignedPage, error) {
	if pageSize < 0 || pageSize > 1000 {
		return UnassignedPage{}, errors.New("wecom: unassigned customers: page_size must be between 1 and 1000")
	}
	request := struct {
		Cursor   string `json:"cursor,omitempty"`
		PageSize int    `json:"page_size,omitempty"`
	}{Cursor: cursor, PageSize: pageSize}
	var out UnassignedPage
	err := c.call(ctx, "external list unassigned customers",
		"/cgi-bin/externalcontact/get_unassigned_list", request, &out)
	return out, err
}

// TransferResignedResult lists the handover state after reassigning the
// customers of a departed member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94082
func (c *External) TransferResignedResult(ctx context.Context, handoverUserID, takeoverUserID, cursor string) (TransferPage, error) {
	if strings.TrimSpace(handoverUserID) == "" || strings.TrimSpace(takeoverUserID) == "" {
		return TransferPage{}, errors.New("wecom: resigned transfer result: handover and takeover userid are required")
	}
	request := struct {
		HandoverUserID string `json:"handover_userid"`
		TakeoverUserID string `json:"takeover_userid"`
		Cursor         string `json:"cursor,omitempty"`
	}{HandoverUserID: handoverUserID, TakeoverUserID: takeoverUserID, Cursor: cursor}
	var out TransferPage
	err := c.call(ctx, "external resigned transfer result",
		"/cgi-bin/externalcontact/resigned/transfer_result", request, &out)
	return out, err
}

// TransferResignedCustomers hands over the customers of a departed member. The
// takeover member cannot be the same person as the departed one.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94081
func (c *External) TransferResignedCustomers(ctx context.Context, request TransferRequest) ([]TransferFailure, error) {
	if strings.TrimSpace(request.HandoverUserID) == "" || strings.TrimSpace(request.TakeoverUserID) == "" {
		return nil, errors.New("wecom: transfer resigned customers: handover and takeover userid are required")
	}
	if request.HandoverUserID == request.TakeoverUserID {
		return nil, errors.New("wecom: transfer resigned customers: handover and takeover must differ")
	}
	if len(request.ExternalUserID) == 0 {
		return nil, errors.New("wecom: transfer resigned customers: external_userid is empty")
	}
	var out struct {
		Customer []TransferFailure `json:"customer"`
	}
	err := c.call(ctx, "external transfer resigned customers",
		"/cgi-bin/externalcontact/resigned/transfer_customer", request, &out)
	return out.Customer, err
}
