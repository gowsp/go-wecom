package external

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Online customer constants.
const (
	CustomerTypeWeChat = 1
	CustomerTypeWeCom  = 2
)

// OnlineChannel is the WeChat Channels profile of a customer.
type OnlineChannel struct {
	Nickname string `json:"nickname,omitempty"`
}

// Customer is the external contact profile.
type Customer struct {
	ExternalUserID string         `json:"external_userid"`
	Name           string         `json:"name"`
	Position       string         `json:"position,omitempty"`
	Avatar         string         `json:"avatar,omitempty"`
	CorpName       string         `json:"corp_name,omitempty"`
	CorpFullName   string         `json:"corp_full_name,omitempty"`
	Type           int            `json:"type,omitempty"`
	Gender         int            `json:"gender,omitempty"`
	UnionID        string         `json:"unionid,omitempty"`
	IsExternal     bool           `json:"-"`
	Channels       *OnlineChannel `json:"wechat_channels,omitempty"`
}

// FollowUser is one member's relation with one customer.
type FollowUser struct {
	UserID         string   `json:"userid"`
	Remark         string   `json:"remark,omitempty"`
	Description    string   `json:"description,omitempty"`
	CreateTime     int64    `json:"createtime,omitempty"`
	TagIDs         []string `json:"tags,omitempty"`
	RemarkCorpName string   `json:"remark_corp_name,omitempty"`
	RemarkMobiles  []string `json:"remark_mobiles,omitempty"`
	OperUserID     string   `json:"oper_userid,omitempty"`
	AddWay         int      `json:"add_way,omitempty"`
	UnionID        string   `json:"unionid,omitempty"`
}

// CustomerDetail is one customer with the members following them.
type CustomerDetail struct {
	Customer   Customer     `json:"external_contact"`
	FollowUser []FollowUser `json:"follow_user"`
	NextCursor string       `json:"next_cursor"`
}

// RemarkRequest changes how one member sees one customer.
type RemarkRequest struct {
	UserID           string   `json:"userid"`
	ExternalUserID   string   `json:"external_userid"`
	Remark           string   `json:"remark,omitempty"`
	Description      string   `json:"description,omitempty"`
	RemarkCompany    string   `json:"remark_company,omitempty"`
	RemarkMobiles    []string `json:"remark_mobiles,omitempty"`
	RemarkPicMediaID string   `json:"remark_pic_mediaid,omitempty"`
}

// ListFollowUsers returns the members configured for customer contact.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92571
func (c *External) ListFollowUsers(ctx context.Context) ([]string, error) {
	var out struct {
		UserIDs []string `json:"follow_user"`
	}
	err := c.get(ctx, "external list follow users", "/cgi-bin/externalcontact/get_follow_user_list", nil, &out)
	return out.UserIDs, err
}

// ListCustomers returns the external user ids of one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92113
func (c *External) ListCustomers(ctx context.Context, userID string) ([]string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, errors.New("wecom: list customers: userid is required")
	}
	var out struct {
		ExternalUserIDs []string `json:"external_userid"`
	}
	err := c.get(ctx, "external list customers", "/cgi-bin/externalcontact/list",
		url.Values{"userid": {userID}}, &out)
	return out.ExternalUserIDs, err
}

// GetCustomer reads one customer with the members following them. The call
// returns at most one member's relation per request; pass NextCursor until it is
// empty to collect the rest.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92114
func (c *External) GetCustomer(ctx context.Context, externalUserID, cursor string) (CustomerDetail, error) {
	externalUserID = strings.TrimSpace(externalUserID)
	if externalUserID == "" {
		return CustomerDetail{}, errors.New("wecom: get customer: external_userid is required")
	}
	query := url.Values{"external_userid": {externalUserID}}
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	var out CustomerDetail
	err := c.get(ctx, "external get customer", "/cgi-bin/externalcontact/get", query, &out)
	return out, err
}

// BatchCustomerDetail is one page of the batch customer query.
type BatchCustomerDetail struct {
	Details    []CustomerDetail `json:"external_contact_list"`
	NextCursor string           `json:"next_cursor"`
}

// GetCustomersByUser batch-reads the customers of several members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92994
func (c *External) GetCustomersByUser(ctx context.Context, userIDs []string, cursor string, limit int) (BatchCustomerDetail, error) {
	if len(userIDs) == 0 {
		return BatchCustomerDetail{}, errors.New("wecom: batch get customers: userid_list is empty")
	}
	request := struct {
		UserIDList []string `json:"userid_list"`
		Cursor     string   `json:"cursor,omitempty"`
		Limit      int      `json:"limit,omitempty"`
	}{UserIDList: userIDs, Cursor: cursor, Limit: limit}
	var out BatchCustomerDetail
	err := c.call(ctx, "external batch get customers", "/cgi-bin/externalcontact/batch/get_by_user", request, &out)
	return out, err
}

// RemarkCustomer changes the remark, description or company of one customer.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92115
func (c *External) RemarkCustomer(ctx context.Context, request RemarkRequest) error {
	if strings.TrimSpace(request.UserID) == "" || strings.TrimSpace(request.ExternalUserID) == "" {
		return errors.New("wecom: remark customer: userid and external_userid are required")
	}
	return c.call(ctx, "external remark customer", "/cgi-bin/externalcontact/remark", request, nil)
}

// NewExternalProfile builds the official 成员对外信息 structure. The admin
// console defines the fields first; unknown names are ignored by WeCom.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92230
func NewExternalProfile(attrs []ExternalAttr) ExternalProfile {
	return ExternalProfile{Attrs: attrs}
}

// ExternalProfile is the 成员对外信息 block.
type ExternalProfile struct {
	Attrs []ExternalAttr `json:"attrs"`
}

// ExternalAttr is one对外信息 entry: type 0 text, 1 web, 2 miniprogram.
type ExternalAttr struct {
	Type        int               `json:"type"`
	Name        string            `json:"name"`
	Text        *ExternalAttrText `json:"text,omitempty"`
	Web         *ExternalAttrWeb  `json:"web,omitempty"`
	MiniProgram *ExternalAttrMini `json:"miniprogram,omitempty"`
}

// ExternalAttrText is a text attribute.
type ExternalAttrText struct {
	Value string `json:"value"`
}

// ExternalAttrWeb is a link attribute.
type ExternalAttrWeb struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// ExternalAttrMini is a mini program attribute.
type ExternalAttrMini struct {
	AppID    string `json:"appid"`
	PagePath string `json:"pagepath"`
	Title    string `json:"title"`
}
