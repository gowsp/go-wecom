package kf

import (
	"context"
	"errors"
)

// Customer is the basic information of one WeChat customer.
type Customer struct {
	ExternalUserID      string               `json:"external_userid"`
	Nickname            string               `json:"nickname"`
	Avatar              string               `json:"avatar"`
	Gender              int                  `json:"gender"`
	UnionID             string               `json:"unionid"`
	EnterSessionContext *EnterSessionContext `json:"enter_session_context"`
}

// EnterSessionContext describes how the customer entered the conversation.
type EnterSessionContext struct {
	Scene      string           `json:"scene"`
	SceneParam string           `json:"scene_param"`
	Channels   *ChannelsContext `json:"wechat_channels"`
}

// ChannelsContext is the video-channel entry of a session.
type ChannelsContext struct {
	Nickname string `json:"nickname"`
	Scene    int    `json:"scene"`
}

// BatchCustomers returns basic customer information. Unknown ids are returned
// separately instead of failing the whole call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95159
func (c *KF) BatchCustomers(ctx context.Context, externalUserIDs []string, needSessionContext bool) ([]Customer, []string, error) {
	if len(externalUserIDs) == 0 {
		return nil, nil, errors.New("wecom: customer service batchget: external_userid_list is required")
	}
	need := 0
	if needSessionContext {
		need = 1
	}
	request := struct {
		ExternalUserIDs []string `json:"external_userid_list"`
		NeedContext     int      `json:"need_enter_session_context"`
	}{ExternalUserIDs: externalUserIDs, NeedContext: need}
	var out struct {
		Customers []Customer `json:"customer_list"`
		Invalid   []string   `json:"invalid_external_userid"`
	}
	err := c.api.Call(ctx, "customer service customer batchget", "/cgi-bin/kf/customer/batchget", request, &out)
	return out.Customers, out.Invalid, err
}
