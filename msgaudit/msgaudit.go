// Package msgaudit implements the WeCom message audit API (/cgi-bin/msgaudit):
// archive permissions, consent checks, internal group info, robot info and the
// audit callbacks.
//
// Pulling the archived messages themselves is done by the official archive SDK
// (NewSdk/GetChatData) inside the application process, not by an HTTP endpoint.
// The decryption helpers here operate on the SDK's output.
package msgaudit

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"net/url"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// MsgAudit reads archived messages with one shared transport.
type MsgAudit struct{ api *client.Client }

// New creates a message audit client.
func New(api *client.Client) *MsgAudit { return &MsgAudit{api: api} }

// ChatItem is one encrypted archived message as returned by the archive SDK.
type ChatItem struct {
	Seq              int64  `json:"seq"`
	MsgID            string `json:"msgid"`
	PublicKeyVersion int    `json:"publickey_ver"`
	EncryptRandomKey string `json:"encrypt_random_key"`
	EncryptChatMsg   string `json:"encrypt_chat_msg"`
}

// Robot is one archive robot bound to the application.
type Robot struct {
	RobotID       string `json:"robot_id"`
	Name          string `json:"name"`
	CreatorUserID string `json:"creator_userid"`
}

// RobotInfo returns the name and creator of one archive robot.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91774
func (c *MsgAudit) RobotInfo(ctx context.Context, robotID string) (Robot, error) {
	if err := check(c); err != nil {
		return Robot{}, err
	}
	robotID = strings.TrimSpace(robotID)
	if robotID == "" {
		return Robot{}, errors.New("wecom: robot info: robot_id is required")
	}
	var out struct {
		Data Robot `json:"data"`
	}
	err := c.api.Get(ctx, "msgaudit robot info", "/cgi-bin/msgaudit/get_robot_info",
		url.Values{"robot_id": {robotID}}, &out)
	return out.Data, err
}

// PrivateKeyFunc returns the private key registered for a public key version.
type PrivateKeyFunc func(version int) (*rsa.PrivateKey, error)

// Decrypt decrypts one archived message with the key of its public key version.
// The random key is RSA/PKCS#1 decrypted, then the message body is AES-256-CBC
// decrypted with the random key as both key and IV source.
func (c *MsgAudit) Decrypt(item ChatItem, keys PrivateKeyFunc) ([]byte, error) {
	if strings.TrimSpace(item.EncryptRandomKey) == "" || strings.TrimSpace(item.EncryptChatMsg) == "" {
		return nil, errors.New("wecom: decrypt: encrypt_random_key and encrypt_chat_msg are required")
	}
	if keys == nil {
		return nil, errors.New("wecom: decrypt: a private key resolver is required")
	}
	key, err := keys(item.PublicKeyVersion)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.New("wecom: decrypt: no private key for public key version")
	}
	randomKeyCipher, err := base64.StdEncoding.DecodeString(item.EncryptRandomKey)
	if err != nil {
		return nil, errors.New("wecom: decrypt: bad random key: " + err.Error())
	}
	randomKey, err := rsa.DecryptPKCS1v15(rand.Reader, key, randomKeyCipher)
	if err != nil {
		return nil, errors.New("wecom: decrypt: rsa decrypt failed: " + err.Error())
	}
	return DecryptData(randomKey, item.EncryptChatMsg)
}

// DecryptData decrypts one message body with an already decrypted random key.
func DecryptData(randomKey []byte, encryptChatMsg string) ([]byte, error) {
	if len(randomKey) < 32 {
		return nil, errors.New("wecom: decrypt: random key must be 32 bytes")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encryptChatMsg)
	if err != nil {
		return nil, errors.New("wecom: decrypt: bad chat msg: " + err.Error())
	}
	block, err := aes.NewCipher(randomKey[:32])
	if err != nil {
		return nil, err
	}
	if len(ciphertext)%aes.BlockSize != 0 || len(ciphertext) == 0 {
		return nil, errors.New("wecom: decrypt: chat msg is not block aligned")
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, randomKey[:16]).CryptBlocks(plain, ciphertext)
	return unpad(plain)
}

func unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("wecom: decrypt: empty plaintext")
	}
	pad := int(data[len(data)-1])
	if pad == 0 || pad > aes.BlockSize || pad > len(data) {
		return nil, errors.New("wecom: decrypt: invalid padding")
	}
	for _, b := range data[len(data)-pad:] {
		if int(b) != pad {
			return nil, errors.New("wecom: decrypt: invalid padding")
		}
	}
	return bytes.TrimSuffix(data, bytes.Repeat([]byte{byte(pad)}, pad)), nil
}

// PermitUsers returns the members whose conversations are archived. kind 1 is
// member, 2 is party, 3 is tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91614
func (c *MsgAudit) PermitUsers(ctx context.Context, kind int) ([]string, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	var out struct {
		IDs []string `json:"ids"`
	}
	err := c.api.Call(ctx, "msgaudit permit users", "/cgi-bin/msgaudit/get_permit_user_list",
		map[string]int{"type": kind}, &out)
	return out.IDs, err
}

// AgreeInfo is one consent record.
type AgreeInfo struct {
	ChangeTime     int64  `json:"status_change_time"`
	UserID         string `json:"userid"`
	ExternalOpenID string `json:"exteranalopenid"`
	Status         string `json:"agree_status"`
}

// SingleAgree returns the consent state of single chats between members and
// customers.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91782
func (c *MsgAudit) SingleAgree(ctx context.Context, pairs [][2]string) ([]AgreeInfo, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(pairs) == 0 {
		return nil, errors.New("wecom: single agree: info is empty")
	}
	type item struct {
		UserID         string `json:"userid"`
		ExternalOpenID string `json:"exteranalopenid"`
	}
	items := make([]item, 0, len(pairs))
	for _, pair := range pairs {
		items = append(items, item{UserID: pair[0], ExternalOpenID: pair[1]})
	}
	var out struct {
		Info []AgreeInfo `json:"agreeinfo"`
	}
	err := c.api.Call(ctx, "msgaudit single agree", "/cgi-bin/msgaudit/check_single_agree",
		map[string]any{"info": items}, &out)
	return out.Info, err
}

// RoomAgree returns the consent state of group chats owned by one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91782
func (c *MsgAudit) RoomAgree(ctx context.Context, roomIDs []string) ([]AgreeInfo, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if len(roomIDs) == 0 {
		return nil, errors.New("wecom: room agree: roomid_list is empty")
	}
	var out struct {
		Info []AgreeInfo `json:"agreeinfo"`
	}
	err := c.api.Call(ctx, "msgaudit room agree", "/cgi-bin/msgaudit/check_room_agree",
		map[string][]string{"roomid_list": roomIDs}, &out)
	return out.Info, err
}

// GroupMember is one member of an internal group.
type GroupMember struct {
	MemberID string `json:"memberid"`
	JoinTime int64  `json:"jointime,omitempty"`
	Type     int    `json:"type,omitempty"`
}

// GroupChat is one internal group chat.
type GroupChat struct {
	RoomName   string        `json:"roomname"`
	Creator    string        `json:"creator"`
	CreateTime int64         `json:"room_create_time"`
	Notice     string        `json:"notice,omitempty"`
	Members    []GroupMember `json:"members"`
}

// GroupChat returns one internal group chat.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92951
func (c *MsgAudit) GroupChat(ctx context.Context, roomID string) (GroupChat, error) {
	if err := check(c); err != nil {
		return GroupChat{}, err
	}
	roomID = strings.TrimSpace(roomID)
	if roomID == "" {
		return GroupChat{}, errors.New("wecom: group chat: roomid is required")
	}
	var out GroupChat
	err := c.api.Call(ctx, "msgaudit group chat", "/cgi-bin/msgaudit/groupchat/get",
		map[string]string{"roomid": roomID}, &out)
	return out, err
}

// ApprovalEvent is the consent change callback of one member.
type ApprovalEvent struct {
	ToUser         string `xml:"ToUserName"`
	FromUser       string `xml:"FromUserName"`
	CreateTime     int64  `xml:"CreateTime"`
	MsgType        string `xml:"MsgType"`
	Name           string `xml:"Event"`
	ChangeType     string `xml:"ChangeType"`
	UserID         string `xml:"UserID"`
	ExternalOpenID string `xml:"ExternalOpenID"`
	Status         string `xml:"Status"`
}

// NotifyEvent is the callback that tells the app new archived messages exist.
type NotifyEvent struct {
	ToUser     string `xml:"ToUserName"`
	FromUser   string `xml:"FromUserName"`
	CreateTime int64  `xml:"CreateTime"`
	MsgType    string `xml:"MsgType"`
	AgentID    int    `xml:"AgentID"`
	Name       string `xml:"Event"`
}

// ParseApprovalEvent decodes a decrypted msg_audit_approved callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92005
func ParseApprovalEvent(data []byte) (ApprovalEvent, error) {
	var event ApprovalEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return ApprovalEvent{}, errors.New("wecom: audit approval event: " + err.Error())
	}
	if event.Name != "change_external_contact" || event.ChangeType != "msg_audit_approved" {
		return ApprovalEvent{}, errors.New("wecom: audit approval event: expected change_external_contact/msg_audit_approved")
	}
	return event, nil
}

// ParseNotifyEvent decodes a decrypted msgaudit_notify callback.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95039
func ParseNotifyEvent(data []byte) (NotifyEvent, error) {
	var event NotifyEvent
	if err := xml.Unmarshal(data, &event); err != nil {
		return NotifyEvent{}, errors.New("wecom: audit notify event: " + err.Error())
	}
	if event.Name != "msgaudit_notify" {
		return NotifyEvent{}, errors.New("wecom: audit notify event: unexpected event " + event.Name)
	}
	return event, nil
}

func check(c *MsgAudit) error {
	if c == nil || c.api == nil {
		return errors.New("wecom: msgaudit client is unavailable")
	}
	return nil
}
