package wedoc

import "context"

const (
	docVIPAddPath  = "/cgi-bin/wedoc/vip/batch_add"
	docVIPDelPath  = "/cgi-bin/wedoc/vip/batch_del"
	docVIPListPath = "/cgi-bin/wedoc/vip/list"
)

// DocVIPResult reports which members received or lost the advanced document
// feature.
type DocVIPResult struct {
	Succeeded []string `json:"succ_userid_list"`
	Failed    []string `json:"fail_userid_list"`
}

// DocVIPList is one page of members with the advanced document feature.
type DocVIPList struct {
	HasMore    bool     `json:"has_more"`
	NextCursor string   `json:"next_cursor"`
	UserIDs    []string `json:"userid_list"`
}

// AddDocVIP grants the advanced document feature to up to 100 members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99516
func (c *WeDoc) AddDocVIP(ctx context.Context, userIDs []string) (DocVIPResult, error) {
	if len(userIDs) == 0 {
		return DocVIPResult{}, errEmpty("add doc vip", "userid_list")
	}
	var out DocVIPResult
	err := c.call(ctx, "add doc vip", docVIPAddPath, map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// RemoveDocVIP revokes the advanced document feature from up to 100 members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99517
func (c *WeDoc) RemoveDocVIP(ctx context.Context, userIDs []string) (DocVIPResult, error) {
	if len(userIDs) == 0 {
		return DocVIPResult{}, errEmpty("remove doc vip", "userid_list")
	}
	var out DocVIPResult
	err := c.call(ctx, "remove doc vip", docVIPDelPath, map[string][]string{"userid_list": userIDs}, &out)
	return out, err
}

// ListDocVIP returns one page of members with the advanced document feature.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99518
func (c *WeDoc) ListDocVIP(ctx context.Context, cursor string, limit int) (DocVIPList, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	body := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit"`
	}{Cursor: cursor, Limit: limit}
	var out DocVIPList
	err := c.call(ctx, "list doc vip", docVIPListPath, body, &out)
	return out, err
}
