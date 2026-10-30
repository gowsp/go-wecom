package wedrive

import (
	"context"
	"errors"
)

// VIPPage is one page of micro disk professional accounts.
type VIPPage struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	UserIDs    []string `json:"userid_list"`
}

// AddVIPAccounts grants professional accounts to members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99512
func (c *WeDrive) AddVIPAccounts(ctx context.Context, userIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return errors.New("wecom: add vip accounts: userid_list is empty")
	}
	return c.api.Call(ctx, "wedrive add vip", "/cgi-bin/wedrive/vip/batch_add",
		map[string][]string{"userid_list": userIDs}, nil)
}

// RemoveVIPAccounts revokes professional accounts from members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99513
func (c *WeDrive) RemoveVIPAccounts(ctx context.Context, userIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return errors.New("wecom: remove vip accounts: userid_list is empty")
	}
	return c.api.Call(ctx, "wedrive remove vip", "/cgi-bin/wedrive/vip/batch_del",
		map[string][]string{"userid_list": userIDs}, nil)
}

// ListVIPAccounts returns one page of professional accounts.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99514
func (c *WeDrive) ListVIPAccounts(ctx context.Context, cursor string, limit int) (VIPPage, error) {
	if err := check(c); err != nil {
		return VIPPage{}, err
	}
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out VIPPage
	err := c.api.Call(ctx, "wedrive list vip", "/cgi-bin/wedrive/vip/list", request, &out)
	return out, err
}

// ProInfo is the professional edition state of the enterprise.
type ProInfo struct {
	IsPro          bool  `json:"is_pro"`
	TotalVIPNumber int   `json:"total_vip_acct_num"`
	UsedVIPNumber  int   `json:"use_vip_acct_num"`
	ExpireTime     int64 `json:"pro_expire_time"`
}

// Capacity is the total micro disk capacity.
type Capacity struct {
	TotalForAll int64 `json:"total_capacity_for_all"`
	TotalForVIP int64 `json:"total_capacity_for_vip"`
}

// ProInfo reads the professional edition state.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95856
func (c *WeDrive) ProInfo(ctx context.Context) (ProInfo, error) {
	if err := check(c); err != nil {
		return ProInfo{}, err
	}
	var out ProInfo
	err := c.api.Call(ctx, "wedrive pro info", "/cgi-bin/wedrive/mng_pro_info", nil, &out)
	return out, err
}

// Capacity reads the total capacity of the enterprise disk.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95856
func (c *WeDrive) Capacity(ctx context.Context) (Capacity, error) {
	if err := check(c); err != nil {
		return Capacity{}, err
	}
	var out Capacity
	err := c.api.Call(ctx, "wedrive capacity", "/cgi-bin/wedrive/mng_capacity", nil, &out)
	return out, err
}
