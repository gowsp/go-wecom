package kf

import (
	"encoding/xml"
	"errors"
	"strings"
)

// EventKFAuthChange is the customer-service account authorisation callback.
const EventKFAuthChange = "kf_account_auth_change"

// AuthChangeEvent is one decrypted customer-service authorisation callback.
// AuthAddOpenKFIDs lists newly authorised accounts, AuthDelOpenKFIDs the ones
// that lost authorisation.
type AuthChangeEvent struct {
	ToUser          string   `xml:"ToUserName"`
	FromUser        string   `xml:"FromUserName"`
	CreateTime      int64    `xml:"CreateTime"`
	MsgType         string   `xml:"MsgType"`
	Event           string   `xml:"Event"`
	AuthAddOpenKFID []string `xml:"AuthAddOpenKfId"`
	AuthDelOpenKFID []string `xml:"AuthDelOpenKfId"`
}

// ParseAuthChangeEvent decodes a decrypted kf_account_auth_change callback.
// Decrypt the body with callback.CallbackCrypto first; the SDK does not route
// HTTP requests.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97712
func ParseAuthChangeEvent(data []byte) (AuthChangeEvent, error) {
	var event AuthChangeEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return AuthChangeEvent{}, errors.New("wecom: customer service event: " + err.Error())
	}
	if strings.TrimSpace(event.Event) == "" {
		return AuthChangeEvent{}, errors.New("wecom: customer service event: event is required")
	}
	if event.Event != EventKFAuthChange {
		return AuthChangeEvent{}, errors.New("wecom: customer service event: unexpected event " + event.Event)
	}
	return event, nil
}
