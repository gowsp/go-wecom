package aibot

import (
	"encoding/json"
	"errors"
	"strings"
)

// Chat types reported in a callback.
const (
	ChatSingle = "single"
	ChatGroup  = "group"
)

// Message types received from users.
const (
	TypeText  = "text"
	TypeImage = "image"
	TypeMixed = "mixed"
	TypeVoice = "voice"
	TypeFile  = "file"
	TypeVideo = "video"
	TypeEvent = "event"
)

// Event types delivered as msgtype=event.
const (
	EventEnterChat         = "enter_chat"
	EventTemplateCardEvent = "template_card_event"
)

// From identifies the sender of a callback.
type From struct {
	CorpID string `json:"corpid,omitempty"`
	UserID string `json:"userid,omitempty"`
}

// Text is a text body.
type Text struct {
	Content string `json:"content"`
}

// Image is an image body; the callback supplies a temporary download URL.
type Image struct {
	URL string `json:"url"`
}

// Media is a voice, file or video body.
type Media struct {
	URL      string `json:"url,omitempty"`
	MediaID  string `json:"media_id,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
}

// MixedItem is one part of a mixed message.
type MixedItem struct {
	MsgType string `json:"msgtype"`
	Text    *Text  `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

// Mixed is a mixed body.
type Mixed struct {
	MsgItem []MixedItem `json:"msg_item"`
}

// Quote is the message the user replied to.
type Quote struct {
	MsgType string `json:"msgtype"`
	Text    *Text  `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

// Event is the payload of msgtype=event. Digit is kept raw: the nested object is
// named after the event type.
type Event struct {
	EventType string          `json:"eventtype"`
	Raw       json.RawMessage `json:"-"`
}

// Message is one decrypted bot callback. Raw keeps the original JSON so callers
// can read fields this package does not model yet.
type Message struct {
	MsgID       string          `json:"msgid"`
	AibotID     string          `json:"aibotid"`
	ChatID      string          `json:"chatid"`
	ChatType    string          `json:"chattype"`
	From        From            `json:"from"`
	ResponseURL string          `json:"response_url"`
	CreateTime  int64           `json:"create_time"`
	MsgType     string          `json:"msgtype"`
	Text        *Text           `json:"text,omitempty"`
	Image       *Image          `json:"image,omitempty"`
	Mixed       *Mixed          `json:"mixed,omitempty"`
	Voice       *Media          `json:"voice,omitempty"`
	File        *Media          `json:"file,omitempty"`
	Video       *Media          `json:"video,omitempty"`
	Quote       *Quote          `json:"quote,omitempty"`
	Event       *Event          `json:"event,omitempty"`
	Raw         json.RawMessage `json:"-"`
}

// ParseMessage decodes one decrypted message callback. Decrypt the body with the
// scheme of the callback configuration first.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100719
func ParseMessage(data []byte) (Message, error) {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return Message{}, errors.New("wecom: robot message: " + err.Error())
	}
	if strings.TrimSpace(message.MsgType) == "" || strings.TrimSpace(message.MsgID) == "" {
		return Message{}, errors.New("wecom: robot message: msgid and msgtype are required")
	}
	if message.ChatType != "" && message.ChatType != ChatSingle && message.ChatType != ChatGroup {
		return Message{}, errors.New("wecom: robot message: unknown chattype " + message.ChatType)
	}
	if message.MsgType == TypeEvent {
		return Message{}, errors.New("wecom: robot message: use ParseEvent for msgtype=event")
	}
	message.Raw = append(json.RawMessage(nil), data...)
	return message, nil
}

// ParseEvent decodes one decrypted event callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101027
func ParseEvent(data []byte) (Message, error) {
	var message Message
	if err := json.Unmarshal(data, &message); err != nil {
		return Message{}, errors.New("wecom: robot event: " + err.Error())
	}
	if message.MsgType != TypeEvent || message.Event == nil {
		return Message{}, errors.New("wecom: robot event: msgtype must be event")
	}
	if strings.TrimSpace(message.Event.EventType) == "" {
		return Message{}, errors.New("wecom: robot event: eventtype is required")
	}
	var envelope struct {
		Event map[string]json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil {
		if raw, ok := envelope.Event[message.Event.EventType]; ok {
			message.Event.Raw = append(json.RawMessage(nil), raw...)
		}
	}
	message.Raw = append(json.RawMessage(nil), data...)
	return message, nil
}
