// Package idconvert implements the WeCom id conversion helpers
// (/cgi-bin/batch/openuserid_to_userid, /cgi-bin/corp/opencorpid_to_corpid,
// /cgi-bin/externalcontact/from_service_external_userid and
// /cgi-bin/idconvert/convert_tmp_external_userid).
package idconvert

import (
	"context"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// IDConvert converts ids across application boundaries with one shared
// transport.
type IDConvert struct{ api *client.Client }

// New creates an id conversion client.
func New(api *client.Client) *IDConvert { return &IDConvert{api: api} }

// UserID is one open_userid resolved to a corp userid.
type UserID struct {
	OpenUserID string `json:"open_userid"`
	UserID     string `json:"userid"`
}

// UserIDs converts open_userid values of one application into corp userids.
// sourceAgentID is the agent that originally issued the open ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95884
// 文档: https://developer.work.weixin.qq.com/document/path/101521
func (c *IDConvert) UserIDs(ctx context.Context, openUserIDs []string, sourceAgentID int) ([]UserID, []string, error) {
	if err := check(c); err != nil {
		return nil, nil, err
	}
	if len(openUserIDs) == 0 {
		return nil, nil, errors.New("wecom: id convert: open_userid_list is empty")
	}
	request := struct {
		List          []string `json:"open_userid_list"`
		SourceAgentID int      `json:"source_agentid,omitempty"`
	}{List: openUserIDs, SourceAgentID: sourceAgentID}
	var out struct {
		List    []UserID `json:"userid_list"`
		Invalid []string `json:"invalid_open_userid_list"`
	}
	err := c.api.Call(ctx, "id convert userids",
		"/cgi-bin/batch/openuserid_to_userid", request, &out)
	return out.List, out.Invalid, err
}

// CorpID converts an open_corpid into a corp corpid.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95884
func (c *IDConvert) CorpID(ctx context.Context, openCorpID string, sourceAgentID int) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	openCorpID = strings.TrimSpace(openCorpID)
	if openCorpID == "" {
		return "", errors.New("wecom: id convert: open_corpid is required")
	}
	request := struct {
		OpenCorpID    string `json:"open_corpid"`
		SourceAgentID int    `json:"source_agentid,omitempty"`
	}{OpenCorpID: openCorpID, SourceAgentID: sourceAgentID}
	var out struct {
		CorpID string `json:"corpid"`
	}
	err := c.api.Call(ctx, "id convert corpid", "/cgi-bin/corp/opencorpid_to_corpid", request, &out)
	return out.CorpID, err
}

// ExternalUserID converts a service-provider external_userid into a corpid-scoped
// one.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95884
func (c *IDConvert) ExternalUserID(ctx context.Context, externalUserID string, sourceAgentID int) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	externalUserID = strings.TrimSpace(externalUserID)
	if externalUserID == "" {
		return "", errors.New("wecom: id convert: external_userid is required")
	}
	request := struct {
		ExternalUserID string `json:"external_userid"`
		SourceAgentID  int    `json:"source_agentid,omitempty"`
	}{ExternalUserID: externalUserID, SourceAgentID: sourceAgentID}
	var out struct {
		ExternalUserID string `json:"external_userid"`
	}
	err := c.api.Call(ctx, "id convert external userid",
		"/cgi-bin/externalcontact/from_service_external_userid", request, &out)
	return out.ExternalUserID, err
}

// TmpResult is one converted temporary external_userid.
type TmpResult struct {
	TmpExternalUserID string `json:"tmp_external_userid"`
	ExternalUserID    string `json:"external_userid,omitempty"`
	CorpID            string `json:"corpid,omitempty"`
	UserID            string `json:"userid,omitempty"`
}

// TmpExternalUserIDs converts temporary customer ids used in upstream/downstream
// and corp group scenarios. businessType 1 is 上下游 and 2 is 企业互联;
// userType 1 means the id belongs to a member and 2 to a customer.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98729
func (c *IDConvert) TmpExternalUserIDs(ctx context.Context, businessType, userType int, tmpIDs []string) ([]TmpResult, []string, error) {
	if err := check(c); err != nil {
		return nil, nil, err
	}
	if len(tmpIDs) == 0 {
		return nil, nil, errors.New("wecom: id convert: tmp_external_userid_list is empty")
	}
	request := struct {
		BusinessType int      `json:"business_type"`
		UserType     int      `json:"user_type"`
		List         []string `json:"tmp_external_userid_list"`
	}{BusinessType: businessType, UserType: userType, List: tmpIDs}
	var out struct {
		Results []TmpResult `json:"results"`
		Invalid []string    `json:"invalid_tmp_external_userid_list"`
	}
	err := c.api.Call(ctx, "id convert tmp external userids",
		"/cgi-bin/idconvert/convert_tmp_external_userid", request, &out)
	return out.Results, out.Invalid, err
}

func check(c *IDConvert) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: idconvert client is unavailable")
	}
	return nil
}
