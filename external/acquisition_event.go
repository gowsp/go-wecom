package external

import (
	"encoding/xml"
	"errors"
	"strings"
)

// Customer acquisition callback change types.
const (
	AcqBalanceLow          = "balance_low"
	AcqBalanceExhausted    = "balance_exhausted"
	AcqBalanceIncreased    = "balance_increased"
	AcqLinkUnavailable     = "link_unavailable"
	AcqDeleteLink          = "delete_link"
	AcqQuotaExpireSoon     = "quota_expire_soon"
	AcqOpenProfile         = "open_profile"
	AcqFriendRequest       = "friend_request"
	AcqCustomerStartChat   = "customer_start_chat"
	AcqMessageFromCustomer = "message_from_customer"
	AcqDelFollowUser       = "del_follow_user"
)

// AcquisitionEvent is one decrypted customer acquisition callback. Only the
// fields of the reported ChangeType are populated.
type AcquisitionEvent struct {
	ToUser         string `xml:"ToUserName"`
	FromUser       string `xml:"FromUserName"`
	CreateTime     int64  `xml:"CreateTime"`
	MsgType        string `xml:"MsgType"`
	Event          string `xml:"Event"`
	ChangeType     string `xml:"ChangeType"`
	LinkID         string `xml:"LinkId"`
	State          string `xml:"State"`
	UserID         string `xml:"UserID"`
	ExternalUserID string `xml:"ExternalUserID"`
	ExpireTime     int64  `xml:"ExpireTime"`
	ExpireQuotaNum int    `xml:"ExpireQuotaNum"`
}

// ParseAcquisitionEvent decodes a decrypted customer_acquisition callback.
// Decrypt the body with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97299
func ParseAcquisitionEvent(data []byte) (AcquisitionEvent, error) {
	var event AcquisitionEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return AcquisitionEvent{}, errors.New("wecom: acquisition event: " + err.Error())
	}
	if event.Event != "customer_acquisition" || strings.TrimSpace(event.ChangeType) == "" {
		return AcquisitionEvent{}, errors.New("wecom: acquisition event: expected customer_acquisition with a change type")
	}
	switch event.ChangeType {
	case AcqBalanceLow, AcqBalanceExhausted, AcqBalanceIncreased, AcqLinkUnavailable,
		AcqDeleteLink, AcqQuotaExpireSoon, AcqOpenProfile, AcqFriendRequest,
		AcqCustomerStartChat, AcqMessageFromCustomer, AcqDelFollowUser:
	default:
		return AcquisitionEvent{}, errors.New("wecom: acquisition event: unknown change type " + event.ChangeType)
	}
	return event, nil
}
