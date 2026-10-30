package kf

import (
	"context"
	"errors"
	"strings"
)

// CorpStatistic is the daily aggregate of one customer-service account.
type CorpStatistic struct {
	StatTime  int64              `json:"stat_time"`
	Statistic CorpStatisticValue `json:"statistic"`
}

// CorpStatisticValue holds the aggregate counters.
type CorpStatisticValue struct {
	SessionCount                int     `json:"session_cnt"`
	CustomerCount               int     `json:"customer_cnt"`
	CustomerMsgCount            int     `json:"customer_msg_cnt"`
	UpgradeServiceCustomerCount int     `json:"upgrade_service_customer_cnt"`
	AISessionReplyCount         int     `json:"ai_session_reply_cnt"`
	AITransferRate              float64 `json:"ai_transfer_rate"`
	AIKnowledgeHitRate          float64 `json:"ai_knowledge_hit_rate"`
	MsgRejectedCustomerCount    int     `json:"msg_rejected_customer_cnt"`
}

// ServicerStatistic is the daily aggregate of one reception employee.
type ServicerStatistic struct {
	StatTime  int64                  `json:"stat_time"`
	Statistic ServicerStatisticValue `json:"statistic"`
}

// ServicerStatisticValue holds the per-employee counters.
type ServicerStatisticValue struct {
	SessionCount                         int     `json:"session_cnt"`
	CustomerCount                        int     `json:"customer_cnt"`
	CustomerMsgCount                     int     `json:"customer_msg_cnt"`
	ReplyRate                            float64 `json:"reply_rate"`
	FirstReplyAverageSec                 int     `json:"first_reply_average_sec"`
	SatisfactionInvestgateCount          int     `json:"satisfaction_investgate_cnt"`
	SatisfactionParticipationRate        float64 `json:"satisfaction_participation_rate"`
	SatisfiedRate                        float64 `json:"satisfied_rate"`
	MiddlingRate                         float64 `json:"middling_rate"`
	DissatisfiedRate                     float64 `json:"dissatisfied_rate"`
	UpgradeServiceCustomerCount          int     `json:"upgrade_service_customer_cnt"`
	UpgradeServiceMemberInviteCount      int     `json:"upgrade_service_member_invite_cnt"`
	UpgradeServiceMemberCustomerCount    int     `json:"upgrade_service_member_customer_cnt"`
	UpgradeServiceGroupChatInviteCount   int     `json:"upgrade_service_groupchat_invite_cnt"`
	UpgradeServiceGroupChatCustomerCount int     `json:"upgrade_service_groupchat_customer_cnt"`
	MsgRejectedCustomerCount             int     `json:"msg_rejected_customer_cnt"`
}

// CorpStatistic returns daily aggregates for one customer-service account. The
// time range is inclusive and may not exceed 30 days.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95489
func (c *KF) CorpStatistic(ctx context.Context, openKfID string, startTime, endTime int64) ([]CorpStatistic, error) {
	if err := checkRange("customer service corp statistic", openKfID, startTime, endTime); err != nil {
		return nil, err
	}
	request := struct {
		OpenKFID  string `json:"open_kfid"`
		StartTime int64  `json:"start_time"`
		EndTime   int64  `json:"end_time"`
	}{OpenKFID: openKfID, StartTime: startTime, EndTime: endTime}
	var out struct {
		List []CorpStatistic `json:"statistic_list"`
	}
	err := c.api.Call(ctx, "customer service corp statistic", "/cgi-bin/kf/get_corp_statistic", request, &out)
	return out.List, err
}

// ServicerStatistic returns daily aggregates for one reception employee.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95490
func (c *KF) ServicerStatistic(ctx context.Context, openKfID, servicerUserID string, startTime, endTime int64) ([]ServicerStatistic, error) {
	if err := checkRange("customer service servicer statistic", openKfID, startTime, endTime); err != nil {
		return nil, err
	}
	if strings.TrimSpace(servicerUserID) == "" {
		return nil, errors.New("wecom: customer service servicer statistic: servicer_userid is required")
	}
	request := struct {
		OpenKFID       string `json:"open_kfid"`
		ServicerUserID string `json:"servicer_userid"`
		StartTime      int64  `json:"start_time"`
		EndTime        int64  `json:"end_time"`
	}{OpenKFID: openKfID, ServicerUserID: servicerUserID, StartTime: startTime, EndTime: endTime}
	var out struct {
		List []ServicerStatistic `json:"statistic_list"`
	}
	err := c.api.Call(ctx, "customer service servicer statistic", "/cgi-bin/kf/get_servicer_statistic", request, &out)
	return out.List, err
}

func checkRange(op, openKfID string, startTime, endTime int64) error {
	if strings.TrimSpace(openKfID) == "" {
		return errors.New("wecom: " + op + ": open_kfid is required")
	}
	if startTime <= 0 || endTime <= 0 || endTime < startTime {
		return errors.New("wecom: " + op + ": invalid time range")
	}
	return nil
}
