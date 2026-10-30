package external

import (
	"context"
	"errors"
)

// BehaviorData is one day of customer-contact behaviour for a member or
// department.
type BehaviorData struct {
	StatTime            int64   `json:"stat_time"`
	ChatCnt             int     `json:"chat_cnt"`
	MessageCnt          int     `json:"message_cnt"`
	ReplyPercentage     float64 `json:"reply_percentage"`
	AvgReplyTime        int     `json:"avg_reply_time"`
	NegativeFeedbackCnt int     `json:"negative_feedback_cnt"`
	NewApplyCnt         int     `json:"new_apply_cnt"`
	NewContactCnt       int     `json:"new_contact_cnt"`
}

// UserBehaviorData returns the daily customer-contact statistics of members or
// departments. The window may not exceed 30 days.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92132
func (c *External) UserBehaviorData(ctx context.Context, userIDs []string, partyIDs []int, startTime, endTime int64) ([]BehaviorData, error) {
	if len(userIDs) == 0 && len(partyIDs) == 0 {
		return nil, errors.New("wecom: user behavior data: userid or partyid is required")
	}
	if startTime <= 0 || endTime < startTime {
		return nil, errors.New("wecom: user behavior data: a valid time window is required")
	}
	request := struct {
		UserID    []string `json:"userid,omitempty"`
		PartyID   []int    `json:"partyid,omitempty"`
		StartTime int64    `json:"start_time"`
		EndTime   int64    `json:"end_time"`
	}{UserID: userIDs, PartyID: partyIDs, StartTime: startTime, EndTime: endTime}
	var out struct {
		Data []BehaviorData `json:"behavior_data"`
	}
	err := c.call(ctx, "external user behavior data",
		"/cgi-bin/externalcontact/get_user_behavior_data", request, &out)
	return out.Data, err
}

// GroupChatStat is the group chat total of one owner.
type GroupChatStat struct {
	Owner string `json:"owner"`
	Data  struct {
		NewChatCnt            int `json:"new_chat_cnt"`
		ChatTotal             int `json:"chat_total"`
		ChatHasMsg            int `json:"chat_has_msg"`
		NewMemberCnt          int `json:"new_member_cnt"`
		MemberTotal           int `json:"member_total"`
		MemberHasMsg          int `json:"member_has_msg"`
		MsgTotal              int `json:"msg_total"`
		MigrateTraineeChatCnt int `json:"migrate_trainee_chat_cnt"`
	} `json:"data"`
}

// GroupChatStatPage is one page of group chat statistics.
type GroupChatStatPage struct {
	Total      int             `json:"total"`
	NextOffset int             `json:"next_offset"`
	Items      []GroupChatStat `json:"items"`
}

// GroupChatStatOptions configures one group chat statistic query.
type GroupChatStatOptions struct {
	DayBeginTime int64    `json:"day_begin_time"`
	DayEndTime   int64    `json:"day_end_time,omitempty"`
	OwnerFilter  []string `json:"-"`
	OrderBy      int      `json:"order_by,omitempty"`
	OrderAsc     int      `json:"order_asc,omitempty"`
	Offset       int      `json:"offset,omitempty"`
	Limit        int      `json:"limit,omitempty"`
}

// GroupChatStatistic returns the group chat totals of the owners in a window.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92133
func (c *External) GroupChatStatistic(ctx context.Context, options GroupChatStatOptions) (GroupChatStatPage, error) {
	if options.DayBeginTime <= 0 {
		return GroupChatStatPage{}, errors.New("wecom: group chat statistic: day_begin_time is required")
	}
	var out GroupChatStatPage
	err := c.call(ctx, "external group chat statistic",
		"/cgi-bin/externalcontact/groupchat/statistic", options.body(), &out)
	return out, err
}

// GroupChatStatisticByDay returns per-day group chat totals.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92133
func (c *External) GroupChatStatisticByDay(ctx context.Context, options GroupChatStatOptions) (GroupChatStatPage, error) {
	if options.DayBeginTime <= 0 || options.DayEndTime <= 0 {
		return GroupChatStatPage{}, errors.New("wecom: group chat statistic by day: day_begin_time and day_end_time are required")
	}
	var out GroupChatStatPage
	err := c.call(ctx, "external group chat statistic by day",
		"/cgi-bin/externalcontact/groupchat/statistic_group_by_day", options.body(), &out)
	return out, err
}

// body renders the request, wrapping the owner filter only when it is set.
func (o GroupChatStatOptions) body() map[string]any {
	body := map[string]any{
		"day_begin_time": o.DayBeginTime,
		"order_by":       o.OrderBy,
		"order_asc":      o.OrderAsc,
		"offset":         o.Offset,
		"limit":          o.Limit,
	}
	if o.DayEndTime > 0 {
		body["day_end_time"] = o.DayEndTime
	}
	if len(o.OwnerFilter) > 0 {
		body["owner_filter"] = map[string]any{"userid_list": o.OwnerFilter}
	}
	return body
}
