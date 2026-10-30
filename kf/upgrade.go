package kf

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Upgrade service types accepted by UpgradeService.
const (
	UpgradeToMember    = 1 // recommend a dedicated employee
	UpgradeToGroupChat = 2 // recommend a customer group
)

// UpgradeConfig lists the employees and groups configured for upgrade service
// in the WeCom administration console.
type UpgradeConfig struct {
	MemberRange    UpgradeMembers `json:"member_range"`
	GroupChatRange UpgradeChats   `json:"groupchat_range"`
}

// UpgradeMembers is the employee range enabled for upgrade service.
type UpgradeMembers struct {
	UserIDs       []string `json:"userid_list"`
	DepartmentIDs []int64  `json:"department_id_list"`
}

// UpgradeChats is the customer-group range enabled for upgrade service.
type UpgradeChats struct {
	ChatIDs []string `json:"chat_id_list"`
}

// UpgradeRequest asks WeCom to recommend one employee or customer group to a
// customer. Member applies when Type is UpgradeToMember, GroupChat when Type is
// UpgradeToGroupChat.
type UpgradeRequest struct {
	OpenKFID       string            `json:"open_kfid"`
	ExternalUserID string            `json:"external_userid"`
	Type           int               `json:"type"`
	Member         *UpgradeMember    `json:"member,omitempty"`
	GroupChat      *UpgradeGroupChat `json:"groupchat,omitempty"`
}

// UpgradeMember is the recommended employee.
type UpgradeMember struct {
	UserID  string `json:"userid"`
	Wording string `json:"wording,omitempty"`
}

// UpgradeGroupChat is the recommended customer group.
type UpgradeGroupChat struct {
	ChatID  string `json:"chat_id"`
	Wording string `json:"wording,omitempty"`
}

// UpgradeConfig returns the employees and groups configured for upgrade service.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94674
func (c *KF) UpgradeConfig(ctx context.Context) (UpgradeConfig, error) {
	var out UpgradeConfig
	err := c.api.Get(ctx, "customer service upgrade config", "/cgi-bin/kf/customer/get_upgrade_service_config",
		url.Values{}, &out)
	return out, err
}

// UpgradeService recommends one employee or customer group to a customer. The
// recommendation must already be configured in the WeCom console, otherwise the
// server answers 95021.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94674
func (c *KF) UpgradeService(ctx context.Context, request UpgradeRequest) error {
	if strings.TrimSpace(request.OpenKFID) == "" || strings.TrimSpace(request.ExternalUserID) == "" {
		return errors.New("wecom: customer service upgrade_service: open_kfid and external_userid are required")
	}
	switch request.Type {
	case UpgradeToMember:
		if request.Member == nil || strings.TrimSpace(request.Member.UserID) == "" {
			return errors.New("wecom: customer service upgrade_service: member.userid is required")
		}
	case UpgradeToGroupChat:
		if request.GroupChat == nil || strings.TrimSpace(request.GroupChat.ChatID) == "" {
			return errors.New("wecom: customer service upgrade_service: groupchat.chat_id is required")
		}
	default:
		return errors.New("wecom: customer service upgrade_service: type must be 1 or 2")
	}
	return c.api.Call(ctx, "customer service upgrade_service", "/cgi-bin/kf/customer/upgrade_service", request, nil)
}

// CancelUpgradeService clears a previously recommended employee or group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94674
func (c *KF) CancelUpgradeService(ctx context.Context, openKfID, externalUserID string) error {
	openKfID = strings.TrimSpace(openKfID)
	externalUserID = strings.TrimSpace(externalUserID)
	if openKfID == "" || externalUserID == "" {
		return errors.New("wecom: customer service cancel_upgrade_service: open_kfid and external_userid are required")
	}
	return c.api.Call(ctx, "customer service cancel_upgrade_service", "/cgi-bin/kf/customer/cancel_upgrade_service",
		map[string]string{"open_kfid": openKfID, "external_userid": externalUserID}, nil)
}
