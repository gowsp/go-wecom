package mail

import (
	"encoding/xml"
	"errors"
	"strings"
)

// Mail event names carried by Event.
const (
	EventAppEmailChange    = "app_email_change"
	EventPublicEmailChange = "public_email_change"
)

// Mail change types carried by ChangeType.
const (
	ChangeReceiveEmail = "receive_email"
)

// ChangeEvent is one decrypted mailbox callback. Event is app_email_change for
// an application mailbox and public_email_change for a public mailbox; Id is
// the public mailbox id.
type ChangeEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Event      string `xml:"Event"`
	ChangeType string `xml:"ChangeType"`
	ID         string `xml:"Id"`
	Amount     int    `xml:"Amount"`
}

// ParseChangeEvent decodes a decrypted mailbox callback. Decrypt the body with
// callback.CallbackCrypto first; the SDK does not route HTTP requests.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97495
// 文档: https://developer.work.weixin.qq.com/document/path/100180
func ParseChangeEvent(data []byte) (ChangeEvent, error) {
	var event ChangeEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return ChangeEvent{}, errors.New("wecom: mail event: " + err.Error())
	}
	if strings.TrimSpace(event.Event) == "" || strings.TrimSpace(event.ChangeType) == "" {
		return ChangeEvent{}, errors.New("wecom: mail event: event and change type are required")
	}
	return event, nil
}
