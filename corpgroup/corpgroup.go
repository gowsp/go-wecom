// Package corpgroup implements the WeCom chain (上下游) and corp-group (企业互联)
// API: shared application discovery, downstream tokens, chain contacts, contact
// rules, pending ids and the related callbacks.
package corpgroup

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// CorpGroup manages chain relations with one shared transport.
type CorpGroup struct{ api *client.Client }

// New creates a chain client.
func New(api *client.Client) *CorpGroup { return &CorpGroup{api: api} }

// Business types.
const (
	// BusinessChain addresses the upstream/downstream (上下游) relation.
	BusinessChain = 1
	// BusinessCorpGroup addresses the corp group (企业互联) relation.
	BusinessCorpGroup = 2
)

// SharedCorp is one corp an application is shared to.
type SharedCorp struct {
	CorpID   string `json:"corpid"`
	CorpName string `json:"corp_name"`
	AgentID  int    `json:"agentid"`
}

// SharedCorpPage is one page of shared corps.
type SharedCorpPage struct {
	Ending int          `json:"ending"`
	Corps  []SharedCorp `json:"corp_list"`
	Cursor string       `json:"next_cursor,omitempty"`
}

// SharedCorps lists the corps an application is shared to.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93403
// 文档: https://developer.work.weixin.qq.com/document/path/95813
func (c *CorpGroup) SharedCorps(ctx context.Context, agentID, businessType int, corpID, cursor string, limit int) (SharedCorpPage, error) {
	if err := check(c); err != nil {
		return SharedCorpPage{}, err
	}
	if agentID <= 0 {
		return SharedCorpPage{}, errors.New("wecom: shared corps: agentid is required")
	}
	request := struct {
		AgentID      int    `json:"agentid"`
		BusinessType int    `json:"business_type,omitempty"`
		CorpID       string `json:"corpid,omitempty"`
		Limit        int    `json:"limit,omitempty"`
		Cursor       string `json:"cursor,omitempty"`
	}{AgentID: agentID, BusinessType: businessType, CorpID: corpID, Limit: limit, Cursor: cursor}
	var out SharedCorpPage
	err := c.api.Call(ctx, "corpgroup shared corps",
		"/cgi-bin/corpgroup/corp/list_app_share_info", request, &out)
	return out, err
}

// CorpToken is a downstream corp access token.
type CorpToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// DownstreamToken returns the access token of a downstream corp.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93359
// 文档: https://developer.work.weixin.qq.com/document/path/95816
func (c *CorpGroup) DownstreamToken(ctx context.Context, corpID string, businessType, agentID int) (CorpToken, error) {
	if err := check(c); err != nil {
		return CorpToken{}, err
	}
	corpID = strings.TrimSpace(corpID)
	if corpID == "" || agentID <= 0 {
		return CorpToken{}, errors.New("wecom: downstream token: corpid and agentid are required")
	}
	request := struct {
		CorpID       string `json:"corpid"`
		BusinessType int    `json:"business_type,omitempty"`
		AgentID      int    `json:"agentid"`
	}{CorpID: corpID, BusinessType: businessType, AgentID: agentID}
	var out CorpToken
	err := c.api.Call(ctx, "corpgroup downstream token",
		"/cgi-bin/corpgroup/corp/gettoken", request, &out)
	return out, err
}

// MiniProgramSession exchanges a login code for a mini program session.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93355
// 文档: https://developer.work.weixin.qq.com/document/path/95817
func (c *CorpGroup) MiniProgramSession(ctx context.Context, userID, sessionKey string) (string, string, error) {
	if err := check(c); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(sessionKey) == "" {
		return "", "", errors.New("wecom: miniprogram session: userid and session_key are required")
	}
	request := struct {
		UserID     string `json:"userid"`
		SessionKey string `json:"session_key"`
	}{UserID: userID, SessionKey: sessionKey}
	var out struct {
		UserID     string `json:"userid"`
		SessionKey string `json:"session_key"`
	}
	err := c.api.Call(ctx, "corpgroup miniprogram session",
		"/cgi-bin/miniprogram/transfer_session", request, &out)
	return out.UserID, out.SessionKey, err
}

// Chain is one upstream/downstream relation.
type Chain struct {
	ChainID   string `json:"chain_id"`
	ChainName string `json:"chain_name"`
}

// ChainGroups returns the groups of one chain.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95820
func (c *CorpGroup) ChainGroups(ctx context.Context, chainID string, groupID int) ([]Chain, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(chainID) == "" {
		return nil, errors.New("wecom: chain groups: chain_id is required")
	}
	request := struct {
		ChainID string `json:"chain_id"`
		GroupID int    `json:"groupid,omitempty"`
	}{ChainID: chainID, GroupID: groupID}
	var out struct {
		Groups []Chain `json:"groups"`
	}
	err := c.api.Call(ctx, "corpgroup chain groups",
		"/cgi-bin/corpgroup/corp/get_chain_group", request, &out)
	return out.Groups, err
}

// Chains returns the chains of the enterprise.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95820
func (c *CorpGroup) Chains(ctx context.Context) ([]Chain, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		Chains []Chain `json:"chains"`
	}
	err := c.api.Call(ctx, "corpgroup chains",
		"/cgi-bin/corpgroup/corp/get_chain_corpinfo_list", nil, &out)
	return out.Chains, err
}

// ChainCorp is one corp inside a chain group.
type ChainCorp struct {
	GroupID      int    `json:"groupid"`
	CorpID       string `json:"corpid"`
	CorpName     string `json:"corp_name"`
	CustomID     string `json:"custom_id,omitempty"`
	InviteUserID string `json:"invite_userid,omitempty"`
}

// ChainCorpPage is one page of chain corps.
type ChainCorpPage struct {
	HasMore    bool        `json:"has_more"`
	NextCursor string      `json:"next_cursor"`
	Corps      []ChainCorp `json:"group_corps"`
}

// ChainCorps returns one page of corps inside one chain group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95820
func (c *CorpGroup) ChainCorps(ctx context.Context, chainID string, groupID int, needPending bool, cursor string, limit int) (ChainCorpPage, error) {
	if err := check(c); err != nil {
		return ChainCorpPage{}, err
	}
	if strings.TrimSpace(chainID) == "" {
		return ChainCorpPage{}, errors.New("wecom: chain corps: chain_id is required")
	}
	request := struct {
		ChainID     string `json:"chain_id"`
		GroupID     int    `json:"groupid"`
		NeedPending bool   `json:"need_pending,omitempty"`
		Cursor      string `json:"cursor,omitempty"`
		Limit       int    `json:"limit,omitempty"`
	}{ChainID: chainID, GroupID: groupID, NeedPending: needPending, Cursor: cursor, Limit: limit}
	var out ChainCorpPage
	err := c.api.Call(ctx, "corpgroup chain corps",
		"/cgi-bin/corpgroup/corp/get_chain_corpinfo", request, &out)
	return out, err
}

// UnionIDToExternalUserID converts WeChat unionids into customer external ids.
// Returns the resolved pairs and the unionids that could not be converted.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95818
func (c *CorpGroup) UnionIDToExternalUserID(ctx context.Context, unionID, openID, corpID, massCallTicket string) ([]ExternalUserIDInfo, []string, error) {
	if err := check(c); err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(unionID) == "" {
		return nil, nil, errors.New("wecom: unionid to external userid: unionid is required")
	}
	request := struct {
		UnionID        string `json:"unionid"`
		OpenID         string `json:"openid,omitempty"`
		CorpID         string `json:"corpid,omitempty"`
		MassCallTicket string `json:"mass_call_ticket,omitempty"`
	}{UnionID: unionID, OpenID: openID, CorpID: corpID, MassCallTicket: massCallTicket}
	var out struct {
		Info    []ExternalUserIDInfo `json:"external_userid_info"`
		Invalid []string             `json:"invalid_unionid_list"`
	}
	err := c.api.Call(ctx, "corpgroup unionid to external userid",
		"/cgi-bin/corpgroup/unionid_to_external_userid", request, &out)
	return out.Info, out.Invalid, err
}

// ExternalUserIDInfo is one resolved customer id.
type ExternalUserIDInfo struct {
	CorpID         string `json:"corpid"`
	ExternalUserID string `json:"external_userid"`
}

// PendingIDs converts unionids of not-yet-added customers into pending ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97357
func (c *CorpGroup) PendingIDs(ctx context.Context, unionIDs []string, openIDs []string) ([]PendingID, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(unionIDs) != len(openIDs) || len(unionIDs) == 0 {
		return nil, errors.New("wecom: pending ids: unionid and openid lists must be non-empty and equal length")
	}
	type pair struct {
		UnionID string `json:"unionid"`
		OpenID  string `json:"openid"`
	}
	pairs := make([]pair, 0, len(unionIDs))
	for i := range unionIDs {
		pairs = append(pairs, pair{UnionID: unionIDs[i], OpenID: openIDs[i]})
	}
	request := struct {
		List []pair `json:"userid_list"`
	}{List: pairs}
	var out struct {
		List []PendingID `json:"pending_id_list"`
	}
	err := c.api.Call(ctx, "corpgroup pending ids",
		"/cgi-bin/corpgroup/batch/external_userid_to_pending_id", request, &out)
	return out.List, err
}

// PendingID is one pending customer id.
type PendingID struct {
	PendingID string `json:"pending_id"`
	UnionID   string `json:"unionid,omitempty"`
	OpenID    string `json:"openid,omitempty"`
}

// PendingID converts one unionid into a pending customer id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97357
func (c *CorpGroup) PendingID(ctx context.Context, unionID, openID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(unionID) == "" {
		return "", errors.New("wecom: pending id: unionid is required")
	}
	request := struct {
		UnionID string `json:"unionid"`
		OpenID  string `json:"openid,omitempty"`
	}{UnionID: unionID, OpenID: openID}
	var out struct {
		PendingID string `json:"pending_id"`
	}
	err := c.api.Call(ctx, "corpgroup pending id",
		"/cgi-bin/corpgroup/unionid_to_pending_id", request, &out)
	return out.PendingID, err
}

// Contact is one contact imported into a chain.
type Contact struct {
	Name         string `json:"name"`
	IdentityType int    `json:"identity_type"`
	Mobile       string `json:"mobile"`
	UserCustomID string `json:"user_custom_id,omitempty"`
}

// ContactCorp is one corp with its contacts to import.
type ContactCorp struct {
	CorpName  string    `json:"corp_name"`
	GroupPath string    `json:"group_path,omitempty"`
	CustomID  string    `json:"custom_id,omitempty"`
	Contacts  []Contact `json:"contact_info_list"`
}

// ImportChainContacts imports contacts into a chain and returns the job id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95821
func (c *CorpGroup) ImportChainContacts(ctx context.Context, chainID string, corps []ContactCorp) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(chainID) == "" || len(corps) == 0 {
		return "", errors.New("wecom: import chain contacts: chain_id and contact_list are required")
	}
	request := struct {
		ChainID string        `json:"chain_id"`
		List    []ContactCorp `json:"contact_list"`
	}{ChainID: chainID, List: corps}
	var out struct {
		JobID string `json:"jobid"`
	}
	err := c.api.Call(ctx, "corpgroup import contacts",
		"/cgi-bin/corpgroup/import_chain_contact", request, &out)
	return out.JobID, err
}

// JobResult is the outcome of one asynchronous chain import.
type JobResult struct {
	Status int           `json:"status"`
	Result *ImportResult `json:"result"`
}

// ImportResult is the import outcome of one chain.
type ImportResult struct {
	ChainID      string       `json:"chain_id"`
	ImportStatus int          `json:"import_status"`
	FailList     []ImportFail `json:"fail_list"`
}

// ImportFail is one corp or contact that failed to import.
type ImportFail struct {
	CorpName string    `json:"corp_name"`
	CustomID string    `json:"custom_id"`
	Errcode  int       `json:"errcode"`
	Errmsg   string    `json:"errmsg"`
	Contacts []Contact `json:"contact_info_list,omitempty"`
}

// ImportResult reads the outcome of one chain import job.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95823
func (c *CorpGroup) ImportResult(ctx context.Context, jobID string) (JobResult, error) {
	if err := check(c); err != nil {
		return JobResult{}, err
	}
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return JobResult{}, errors.New("wecom: import result: jobid is required")
	}
	var out JobResult
	err := c.api.Call(ctx, "corpgroup import result", "/cgi-bin/corpgroup/getresult",
		map[string]string{"jobid": jobID}, &out)
	return out, err
}

// RemoveCorp removes a corp from a chain.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95822
func (c *CorpGroup) RemoveCorp(ctx context.Context, chainID, corpID, pendingCorpID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(chainID) == "" || (corpID == "" && pendingCorpID == "") {
		return errors.New("wecom: remove corp: chain_id and corpid or pending_corpid are required")
	}
	request := struct {
		ChainID       string `json:"chain_id"`
		CorpID        string `json:"corpid,omitempty"`
		PendingCorpID string `json:"pending_corpid,omitempty"`
	}{ChainID: chainID, CorpID: corpID, PendingCorpID: pendingCorpID}
	return c.api.Call(ctx, "corpgroup remove corp",
		"/cgi-bin/corpgroup/corp/remove_corp", request, nil)
}

// ChainUserCustomID returns the custom id of one member of a chain corp.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97441
func (c *CorpGroup) ChainUserCustomID(ctx context.Context, chainID, corpID, userID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(chainID) == "" || strings.TrimSpace(corpID) == "" || strings.TrimSpace(userID) == "" {
		return "", errors.New("wecom: chain user custom id: chain_id, corpid and userid are required")
	}
	request := struct {
		ChainID string `json:"chain_id"`
		CorpID  string `json:"corpid"`
		UserID  string `json:"userid"`
	}{ChainID: chainID, CorpID: corpID, UserID: userID}
	var out struct {
		UserCustomID string `json:"user_custom_id"`
	}
	err := c.api.Call(ctx, "corpgroup chain user custom id",
		"/cgi-bin/corpgroup/corp/get_chain_user_custom_id", request, &out)
	return out.UserCustomID, err
}

// SharedChains returns the chains a downstream corp joined.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97442
func (c *CorpGroup) SharedChains(ctx context.Context, corpID string) ([]Chain, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	corpID = strings.TrimSpace(corpID)
	if corpID == "" {
		return nil, errors.New("wecom: shared chains: corpid is required")
	}
	var out struct {
		Chains []Chain `json:"chains"`
	}
	err := c.api.Call(ctx, "corpgroup shared chains",
		"/cgi-bin/corpgroup/get_corp_shared_chain_list", map[string]string{"corpid": corpID}, &out)
	return out.Chains, err
}

// RuleRange is the owner or member range of one contact rule.
type RuleRange struct {
	DepartmentIDs []string `json:"departmentids,omitempty"`
	UserIDs       []string `json:"userids,omitempty"`
	GroupIDs      []string `json:"groupids,omitempty"`
	CorpIDs       []string `json:"corpids,omitempty"`
}

// RuleInfo is the scope of one contact rule.
type RuleInfo struct {
	OwnerRange  RuleRange `json:"owner_corp_range"`
	MemberRange RuleRange `json:"member_corp_range"`
}

// RuleIDs returns the contact rule ids of one chain.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95631
func (c *CorpGroup) RuleIDs(ctx context.Context, chainID string) ([]int, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	chainID = strings.TrimSpace(chainID)
	if chainID == "" {
		return nil, errors.New("wecom: rule ids: chain_id is required")
	}
	var out struct {
		IDs []int `json:"rule_ids"`
	}
	err := c.api.Call(ctx, "corpgroup rule ids", "/cgi-bin/corpgroup/rule/list_ids",
		map[string]string{"chain_id": chainID}, &out)
	return out.IDs, err
}

// Rule reads one contact rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95633
func (c *CorpGroup) Rule(ctx context.Context, chainID string, ruleID int) (RuleInfo, error) {
	if err := check(c); err != nil {
		return RuleInfo{}, err
	}
	if strings.TrimSpace(chainID) == "" || ruleID <= 0 {
		return RuleInfo{}, errors.New("wecom: rule: chain_id and rule_id are required")
	}
	request := struct {
		ChainID string `json:"chain_id"`
		RuleID  int    `json:"rule_id"`
	}{ChainID: chainID, RuleID: ruleID}
	var out struct {
		Info RuleInfo `json:"rule_info"`
	}
	err := c.api.Call(ctx, "corpgroup rule", "/cgi-bin/corpgroup/rule/get_rule_info", request, &out)
	return out.Info, err
}

// AddRule creates one contact rule and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95634
func (c *CorpGroup) AddRule(ctx context.Context, chainID string, info RuleInfo) (int, error) {
	if err := check(c); err != nil {
		return 0, err
	}
	chainID = strings.TrimSpace(chainID)
	if chainID == "" {
		return 0, errors.New("wecom: add rule: chain_id is required")
	}
	request := struct {
		ChainID string   `json:"chain_id"`
		Info    RuleInfo `json:"rule_info"`
	}{ChainID: chainID, Info: info}
	var out struct {
		RuleID int `json:"rule_id"`
	}
	err := c.api.Call(ctx, "corpgroup add rule", "/cgi-bin/corpgroup/rule/add_rule", request, &out)
	return out.RuleID, err
}

// UpdateRule replaces one contact rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95635
func (c *CorpGroup) UpdateRule(ctx context.Context, chainID string, ruleID int, info RuleInfo) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(chainID) == "" || ruleID <= 0 {
		return errors.New("wecom: update rule: chain_id and rule_id are required")
	}
	request := struct {
		ChainID string   `json:"chain_id"`
		RuleID  int      `json:"rule_id"`
		Info    RuleInfo `json:"rule_info"`
	}{ChainID: chainID, RuleID: ruleID, Info: info}
	return c.api.Call(ctx, "corpgroup update rule", "/cgi-bin/corpgroup/rule/modify_rule", request, nil)
}

// DeleteRule removes one contact rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95632
func (c *CorpGroup) DeleteRule(ctx context.Context, chainID string, ruleID int) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(chainID) == "" || ruleID <= 0 {
		return errors.New("wecom: delete rule: chain_id and rule_id are required")
	}
	request := struct {
		ChainID string `json:"chain_id"`
		RuleID  int    `json:"rule_id"`
	}{ChainID: chainID, RuleID: ruleID}
	return c.api.Call(ctx, "corpgroup delete rule", "/cgi-bin/corpgroup/rule/delete_rule", request, nil)
}

// Chain change types carried by change_chain.
const (
	ChangeCreateChain = "create_chain"
	ChangeUpdateChain = "update_chain"
	ChangeDeleteChain = "delete_chain"
)

// ChainEvent is one decrypted chain change callback.
type ChainEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	ChangeType string `xml:"ChangeType"`
	ChainID    string `xml:"ChainId"`
}

// ParseChainEvent decodes a decrypted change_chain callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95796
func ParseChainEvent(data []byte) (ChainEvent, error) {
	var event ChainEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return ChainEvent{}, errors.New("wecom: chain event: " + err.Error())
	}
	if event.Name != "change_chain" {
		return ChainEvent{}, errors.New("wecom: chain event: unexpected event " + event.Name)
	}
	switch event.ChangeType {
	case ChangeCreateChain, ChangeUpdateChain, ChangeDeleteChain:
	default:
		return ChainEvent{}, errors.New("wecom: chain event: unexpected change type " + event.ChangeType)
	}
	if strings.TrimSpace(event.ChainID) == "" {
		return ChainEvent{}, errors.New("wecom: chain event: ChainId is required")
	}
	return event, nil
}

// BatchJobEvent is one decrypted asynchronous job notification.
type BatchJobEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Name       string `xml:"Event"`
	JobID      string `xml:"BatchJob>JobId"`
	JobType    string `xml:"BatchJob>JobType"`
	ErrCode    int    `xml:"BatchJob>ErrCode"`
	ErrMsg     string `xml:"BatchJob>ErrMsg"`
}

// ParseBatchJobEvent decodes a decrypted batch_job_result callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95797
func ParseBatchJobEvent(data []byte) (BatchJobEvent, error) {
	var event BatchJobEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return BatchJobEvent{}, errors.New("wecom: batch job event: " + err.Error())
	}
	if event.Name != "batch_job_result" {
		return BatchJobEvent{}, errors.New("wecom: batch job event: unexpected event " + event.Name)
	}
	if strings.TrimSpace(event.JobID) == "" {
		return BatchJobEvent{}, errors.New("wecom: batch job event: JobId is required")
	}
	return event, nil
}

func check(c *CorpGroup) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: corpgroup client is unavailable")
	}
	return nil
}
