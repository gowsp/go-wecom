package mail

import "context"

const (
	vipAddPath  = "/cgi-bin/exmail/vip/batch_add"
	vipDelPath  = "/cgi-bin/exmail/vip/batch_del"
	vipListPath = "/cgi-bin/exmail/vip/list"
)

// VIPResult reports which members received or lost the advanced mailbox
// feature. A member that could not be changed is listed in Failed.
type VIPResult struct {
	Succeeded []string `json:"succ_userid_list"`
	Failed    []string `json:"fail_userid_list"`
}

// VIPList is one page of members with the advanced mailbox feature.
type VIPList struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	UserIDs    []string `json:"userid_list"`
}

// AddVIP grants the advanced mailbox feature to up to 100 members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99316
func (c *Mail) AddVIP(ctx context.Context, userIDs []string) (VIPResult, error) {
	if _, err := requireIDs("add vip accounts", "userid_list", userIDs); err != nil {
		return VIPResult{}, err
	}
	var out VIPResult
	err := c.call(ctx, "add vip accounts", vipAddPath, map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// RemoveVIP revokes the advanced mailbox feature from up to 100 members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99317
func (c *Mail) RemoveVIP(ctx context.Context, userIDs []string) (VIPResult, error) {
	if _, err := requireIDs("remove vip accounts", "userid_list", userIDs); err != nil {
		return VIPResult{}, err
	}
	var out VIPResult
	err := c.call(ctx, "remove vip accounts", vipDelPath, map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// ListVIP returns one page of members with the advanced mailbox feature.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99318
func (c *Mail) ListVIP(ctx context.Context, cursor string, limit int) (VIPList, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out VIPList
	err := c.call(ctx, "list vip accounts", vipListPath, request, &out)
	return out, err
}
