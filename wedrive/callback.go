package wedrive

import (
	"encoding/xml"
	"errors"
	"strings"
)

// Space change types carried by wedrive_space_change.
const (
	ChangeDismissSpace         = "dismiss_space"
	ChangeSpaceMember          = "space_member_change"
	ChangeSpaceSecuritySetting = "space_security_settings_change"
)

// Event is one decrypted micro disk callback. Space-change events can list
// several spaces, so SpaceIDs is a slice.
type Event struct {
	ToUser     string   `xml:"ToUserName"`
	FromUser   string   `xml:"FromUserName"`
	CreateTime int64    `xml:"CreateTime"`
	MsgType    string   `xml:"MsgType"`
	Name       string   `xml:"Event"`
	ChangeType string   `xml:"ChangeType"`
	SpaceIDs   []string `xml:"SpaceId"`
	FileID     string   `xml:"FileId"`
}

// ParseEvent decodes a decrypted micro disk callback. Decrypt the body with
// callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97898
// 文档: https://developer.work.weixin.qq.com/document/path/97899
// 文档: https://developer.work.weixin.qq.com/document/path/97900
// 文档: https://developer.work.weixin.qq.com/document/path/97901
// 文档: https://developer.work.weixin.qq.com/document/path/97902
// 文档: https://developer.work.weixin.qq.com/document/path/97903
func ParseEvent(data []byte) (Event, error) {
	var event Event
	if err := xml.Unmarshal(data, &event); err != nil {
		return Event{}, errors.New("wecom: wedrive event: " + err.Error())
	}
	switch event.Name {
	case "wedrive_insufficient_capacity":
	case "wedrive_space_change":
		switch event.ChangeType {
		case ChangeDismissSpace, ChangeSpaceMember, ChangeSpaceSecuritySetting:
		default:
			return Event{}, errors.New("wecom: wedrive event: unknown space change type " + event.ChangeType)
		}
		if len(event.SpaceIDs) == 0 {
			return Event{}, errors.New("wecom: wedrive event: SpaceId is required")
		}
	case "wedrive_file_change":
		if strings.TrimSpace(event.FileID) == "" {
			return Event{}, errors.New("wecom: wedrive event: FileId is required")
		}
	default:
		return Event{}, errors.New("wecom: wedrive event: unexpected event " + event.Name)
	}
	return event, nil
}
