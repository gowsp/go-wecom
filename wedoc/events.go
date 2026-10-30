package wedoc

import (
	"encoding/xml"
	"errors"
	"strings"
)

// Document change types carried by ChangeType.
const (
	ChangeDocMember    = "doc_member_change"
	ChangeDeleteDoc    = "delete_doc"
	ChangeFormComplete = "form_complete"
	ChangeDeleteForm   = "delete_form"
	ChangeFormSettings = "form_settings_change"
	ChangeAddRecord    = "add_record"
	ChangeUpdateRecord = "update_record"
	ChangeDeleteRecord = "delete_record"
	// The field change types really are spelled "filed" in the official
	// documentation; keep the wire values as published.
	ChangeAddField    = "add_filed"
	ChangeUpdateField = "update_filed"
	ChangeDeleteField = "delete_filed"
)

// ChangeEvent is one decrypted documentation callback. Official payloads repeat
// DocId, FormId, RecordId and FieldId, so they are slices even when a single
// change usually carries one value.
type ChangeEvent struct {
	ToUser     string   `xml:"ToUserName"`
	FromUser   string   `xml:"FromUserName"`
	CreateTime int64    `xml:"CreateTime"`
	MsgType    string   `xml:"MsgType"`
	Event      string   `xml:"Event"`
	ChangeType string   `xml:"ChangeType"`
	DocIDs     []string `xml:"DocId"`
	FormIDs    []string `xml:"FormId"`
	SheetID    string   `xml:"SheetId"`
	RecordIDs  []string `xml:"RecordId"`
	FieldIDs   []string `xml:"FieldId"`
}

// ParseDocumentEvent decodes a decrypted doc_change callback. Decrypt the body
// with callback.CallbackCrypto first; the SDK does not route HTTP requests.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97833
// 文档: https://developer.work.weixin.qq.com/document/path/97834
// 文档: https://developer.work.weixin.qq.com/document/path/97835
// 文档: https://developer.work.weixin.qq.com/document/path/98095
// 文档: https://developer.work.weixin.qq.com/document/path/98096
func ParseDocumentEvent(data []byte) (ChangeEvent, error) {
	return parseChangeEvent(data, "doc_change")
}

// ParseSmartSheetEvent decodes a decrypted smart_sheet_change callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100986
// 文档: https://developer.work.weixin.qq.com/document/path/100987
func ParseSmartSheetEvent(data []byte) (ChangeEvent, error) {
	return parseChangeEvent(data, "smart_sheet_change")
}

func parseChangeEvent(data []byte, want string) (ChangeEvent, error) {
	var event ChangeEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return ChangeEvent{}, errors.New("wecom: documentation event: " + err.Error())
	}
	if strings.TrimSpace(event.Event) == "" || strings.TrimSpace(event.ChangeType) == "" {
		return ChangeEvent{}, errors.New("wecom: documentation event: event and change type are required")
	}
	if event.Event != want {
		return ChangeEvent{}, errors.New("wecom: documentation event: unexpected event " + event.Event)
	}
	return event, nil
}
