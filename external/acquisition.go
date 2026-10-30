package external

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// AcquisitionRange selects the members an acquisition link serves.
type AcquisitionRange struct {
	UserList       []string `json:"user_list,omitempty"`
	DepartmentList []int    `json:"department_list,omitempty"`
}

// PriorityOption controls how an acquisition link distributes customers.
type PriorityOption struct {
	PriorityType       int      `json:"priority_type,omitempty"`
	PriorityUserIDList []string `json:"priority_userid_list,omitempty"`
}

// AcquisitionLink is one customer acquisition link. The same shape covers the
// create and update requests and the detail response.
type AcquisitionLink struct {
	LinkID         string            `json:"link_id,omitempty"`
	LinkName       string            `json:"link_name,omitempty"`
	URL            string            `json:"url,omitempty"`
	CreateTime     int64             `json:"create_time,omitempty"`
	SkipVerify     bool              `json:"skip_verify,omitempty"`
	MarkSource     bool              `json:"mark_source,omitempty"`
	Range          *AcquisitionRange `json:"range,omitempty"`
	PriorityOption *PriorityOption   `json:"priority_option,omitempty"`
}

// AcquisitionLinkPage is one page of acquisition link ids.
type AcquisitionLinkPage struct {
	NextCursor string   `json:"next_cursor"`
	LinkIDs    []string `json:"link_id_list"`
}

// AcquisitionLinkDetail is the detail of one acquisition link.
type AcquisitionLinkDetail struct {
	Link           AcquisitionLink   `json:"link"`
	Range          *AcquisitionRange `json:"range,omitempty"`
	PriorityOption *PriorityOption   `json:"priority_option,omitempty"`
}

// ListAcquisitionLinks returns one page of valid acquisition link ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97297
func (c *External) ListAcquisitionLinks(ctx context.Context, cursor string, limit int) (AcquisitionLinkPage, error) {
	request := struct {
		Limit  int    `json:"limit,omitempty"`
		Cursor string `json:"cursor,omitempty"`
	}{Limit: limit, Cursor: cursor}
	var out AcquisitionLinkPage
	err := c.call(ctx, "external list acquisition links",
		"/cgi-bin/externalcontact/customer_acquisition/list_link", request, &out)
	return out, err
}

// GetAcquisitionLink reads one acquisition link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97297
func (c *External) GetAcquisitionLink(ctx context.Context, linkID string) (AcquisitionLinkDetail, error) {
	if strings.TrimSpace(linkID) == "" {
		return AcquisitionLinkDetail{}, errors.New("wecom: get acquisition link: link_id is required")
	}
	var out AcquisitionLinkDetail
	err := c.call(ctx, "external get acquisition link",
		"/cgi-bin/externalcontact/customer_acquisition/get",
		map[string]string{"link_id": linkID}, &out)
	return out, err
}

// CreateAcquisitionLink creates one acquisition link and returns the stored
// record, including its URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97297
func (c *External) CreateAcquisitionLink(ctx context.Context, link AcquisitionLink) (AcquisitionLink, error) {
	if strings.TrimSpace(link.LinkName) == "" {
		return AcquisitionLink{}, errors.New("wecom: create acquisition link: link_name is required")
	}
	if link.Range == nil || (len(link.Range.UserList) == 0 && len(link.Range.DepartmentList) == 0) {
		return AcquisitionLink{}, errors.New("wecom: create acquisition link: range is required")
	}
	var out struct {
		Link AcquisitionLink `json:"link"`
	}
	err := c.call(ctx, "external create acquisition link",
		"/cgi-bin/externalcontact/customer_acquisition/create_link", link, &out)
	return out.Link, err
}

// UpdateAcquisitionLink changes one acquisition link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97297
func (c *External) UpdateAcquisitionLink(ctx context.Context, link AcquisitionLink) error {
	if strings.TrimSpace(link.LinkID) == "" {
		return errors.New("wecom: update acquisition link: link_id is required")
	}
	return c.call(ctx, "external update acquisition link",
		"/cgi-bin/externalcontact/customer_acquisition/update_link", link, nil)
}

// DeleteAcquisitionLink removes one acquisition link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97297
func (c *External) DeleteAcquisitionLink(ctx context.Context, linkID string) error {
	if strings.TrimSpace(linkID) == "" {
		return errors.New("wecom: delete acquisition link: link_id is required")
	}
	return c.call(ctx, "external delete acquisition link",
		"/cgi-bin/externalcontact/customer_acquisition/delete_link",
		map[string]string{"link_id": linkID}, nil)
}

// AcquisitionCustomer is one customer added through an acquisition link.
type AcquisitionCustomer struct {
	ExternalUserID string `json:"external_userid"`
	UserID         string `json:"userid"`
	ChatStatus     int    `json:"chat_status"`
	State          string `json:"state,omitempty"`
}

// AcquisitionCustomerPage is one page of customers added by a link.
type AcquisitionCustomerPage struct {
	NextCursor string                `json:"next_cursor"`
	Customers  []AcquisitionCustomer `json:"customer_list"`
}

// ListAcquisitionCustomers returns the customers added through one link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97298
func (c *External) ListAcquisitionCustomers(ctx context.Context, linkID, cursor string, limit int) (AcquisitionCustomerPage, error) {
	if strings.TrimSpace(linkID) == "" {
		return AcquisitionCustomerPage{}, errors.New("wecom: acquisition customers: link_id is required")
	}
	request := struct {
		LinkID string `json:"link_id"`
		Limit  int    `json:"limit,omitempty"`
		Cursor string `json:"cursor,omitempty"`
	}{LinkID: linkID, Limit: limit, Cursor: cursor}
	var out AcquisitionCustomerPage
	err := c.call(ctx, "external acquisition customers",
		"/cgi-bin/externalcontact/customer_acquisition/customer", request, &out)
	return out, err
}

// AcquisitionQuotaItem is one expiring quota grant.
type AcquisitionQuotaItem struct {
	ExpireDate int64 `json:"expire_date"`
	Balance    int   `json:"balance"`
}

// AcquisitionQuota is the remaining acquisition quota.
type AcquisitionQuota struct {
	Total     int                    `json:"total"`
	Balance   int                    `json:"balance"`
	QuotaList []AcquisitionQuotaItem `json:"quota_list"`
}

// AcquisitionQuota reads the remaining acquisition quota.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97375
func (c *External) AcquisitionQuota(ctx context.Context) (AcquisitionQuota, error) {
	var out AcquisitionQuota
	err := c.get(ctx, "external acquisition quota",
		"/cgi-bin/externalcontact/customer_acquisition_quota", url.Values{}, &out)
	return out, err
}

// AcquisitionUsage is the click and addition count of one link.
type AcquisitionUsage struct {
	ClickLinkCustomerCnt int `json:"click_link_customer_cnt"`
	NewCustomerCnt       int `json:"new_customer_cnt"`
}

// AcquisitionUsage reports how many customers one link reached and added.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97375
func (c *External) AcquisitionUsage(ctx context.Context, linkID string, startTime, endTime int64) (AcquisitionUsage, error) {
	if strings.TrimSpace(linkID) == "" {
		return AcquisitionUsage{}, errors.New("wecom: acquisition usage: link_id is required")
	}
	request := struct {
		LinkID    string `json:"link_id"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
	}{LinkID: linkID, StartTime: startTime, EndTime: endTime}
	var out AcquisitionUsage
	err := c.call(ctx, "external acquisition usage",
		"/cgi-bin/externalcontact/customer_acquisition/statistic", request, &out)
	return out, err
}

// AcquisitionChat is the acquisition context of one chat session.
type AcquisitionChat struct {
	RecvMsgCnt int    `json:"recv_msg_cnt"`
	LinkID     string `json:"link_id,omitempty"`
	State      string `json:"state,omitempty"`
}

// AcquisitionChatInfo is the result of one chat lookup.
type AcquisitionChatInfo struct {
	UserID         string          `json:"userid"`
	ExternalUserID string          `json:"external_userid"`
	ChatInfo       AcquisitionChat `json:"chat_info"`
}

// AcquisitionChatInfo reports how often one member received messages from an
// acquisition customer.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100130
func (c *External) AcquisitionChatInfo(ctx context.Context, chatKey string) (AcquisitionChatInfo, error) {
	chatKey = strings.TrimSpace(chatKey)
	if chatKey == "" {
		return AcquisitionChatInfo{}, errors.New("wecom: acquisition chat info: chat_key is required")
	}
	var out AcquisitionChatInfo
	err := c.call(ctx, "external acquisition chat info",
		"/cgi-bin/externalcontact/customer_acquisition/get_chat_info",
		map[string]string{"chat_key": chatKey}, &out)
	return out, err
}
