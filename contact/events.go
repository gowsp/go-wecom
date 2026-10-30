package contact

import (
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
)

// Contact change types carried by ChangeType.
const (
	ChangeCreateUser  = "create_user"
	ChangeUpdateUser  = "update_user"
	ChangeDeleteUser  = "delete_user"
	ChangeCreateParty = "create_party"
	ChangeUpdateParty = "update_party"
	ChangeDeleteParty = "delete_party"
	ChangeUpdateTag   = "update_tag"
)

// StringList decodes a comma-separated list such as "lisi,wangwu".
type StringList []string

// UnmarshalXML splits the element text on commas.
func (s *StringList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}
	var out []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	*s = out
	return nil
}

// IntList decodes a comma-separated integer list such as "1,2,3".
type IntList []int

// UnmarshalXML splits the element text on commas.
func (s *IntList) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := d.DecodeElement(&raw, &start); err != nil {
		return err
	}
	var out []int
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		value, err := strconv.Atoi(item)
		if err != nil {
			return err
		}
		out = append(out, value)
	}
	*s = out
	return nil
}

// ChangeEvent is one decrypted change_contact callback. Member, department and
// tag changes share the envelope; only the fields of the reported ChangeType are
// populated.
type ChangeEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	Event      string `xml:"Event"`
	ChangeType string `xml:"ChangeType"`

	// Member changes.
	UserID         string     `xml:"UserID"`
	NewUserID      string     `xml:"NewUserID"`
	Name           string     `xml:"Name"`
	Department     IntList    `xml:"Department"`
	MainDepartment int        `xml:"MainDepartment"`
	IsLeaderInDept IntList    `xml:"IsLeaderInDept"`
	DirectLeader   StringList `xml:"DirectLeader"`
	Position       string     `xml:"Position"`
	Mobile         string     `xml:"Mobile"`
	Gender         string     `xml:"Gender"`
	Email          string     `xml:"Email"`
	BizMail        string     `xml:"BizMail"`
	Status         int        `xml:"Status"`
	Avatar         string     `xml:"Avatar"`
	Alias          string     `xml:"Alias"`
	Telephone      string     `xml:"Telephone"`
	Address        string     `xml:"Address"`
	EnglishName    string     `xml:"EnglishName"`

	// Department changes.
	ID       int `xml:"Id"`
	ParentID int `xml:"ParentId"`
	Order    int `xml:"Order"`

	// Tag changes.
	TagID      int        `xml:"TagId"`
	AddUsers   StringList `xml:"AddUserItems"`
	DelUsers   StringList `xml:"DelUserItems"`
	AddParties IntList    `xml:"AddPartyItems"`
	DelParties IntList    `xml:"DelPartyItems"`
}

// ParseChangeEvent decodes a decrypted change_contact callback for members,
// departments or tags. Decrypt the body with callback.CallbackCrypto first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90970
// 文档: https://developer.work.weixin.qq.com/document/path/90971
// 文档: https://developer.work.weixin.qq.com/document/path/90972
func ParseChangeEvent(data []byte) (ChangeEvent, error) {
	var event ChangeEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return ChangeEvent{}, errors.New("wecom: contact event: " + err.Error())
	}
	if event.Event != "change_contact" || strings.TrimSpace(event.ChangeType) == "" {
		return ChangeEvent{}, errors.New("wecom: contact event: expected change_contact with a change type")
	}
	return event, nil
}

// BatchJob is the task summary of an asynchronous import notification.
type BatchJob struct {
	JobID   string `xml:"JobId"`
	JobType string `xml:"JobType"`
	ErrCode int    `xml:"ErrCode"`
	ErrMsg  string `xml:"ErrMsg"`
}

// BatchJobEvent is one decrypted batch_job_result callback.
type BatchJobEvent struct {
	ToUser     string   `xml:"ToUserName"`
	FromUser   string   `xml:"FromUserName"`
	CreateTime int64    `xml:"CreateTime"`
	MsgType    string   `xml:"MsgType"`
	Event      string   `xml:"Event"`
	Job        BatchJob `xml:"BatchJob"`
}

// ParseBatchJobEvent decodes a decrypted batch_job_result callback, used both by
// asynchronous imports and by contact exports.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90973
// 文档: https://developer.work.weixin.qq.com/document/path/94946
func ParseBatchJobEvent(data []byte) (BatchJobEvent, error) {
	var event BatchJobEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return BatchJobEvent{}, errors.New("wecom: contact batch event: " + err.Error())
	}
	if event.Event != "batch_job_result" || strings.TrimSpace(event.Job.JobID) == "" {
		return BatchJobEvent{}, errors.New("wecom: contact batch event: expected batch_job_result with a job id")
	}
	return event, nil
}
