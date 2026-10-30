// Package callback implements the encrypted callback protocol of a WeCom
// self-built application: URL verification, envelope decryption and passive
// reply encryption.
//
// WeCom documents the protocol as 接收消息与事件
// (https://developer.work.weixin.qq.com/document/path/90238). WeCom calls the
// application URL twice: a GET request proves ownership of the URL, and POST
// requests deliver messages and events. Both directions use the same envelope
// encryption: AES-256-CBC with a key derived from EncodingAESKey, a 16-byte IV
// equal to the first key block, a random 16-byte prefix, a four-byte network
// order message length, the message, and the receive id (the corp id). Padding
// is PKCS#7 with a block size of 32.
//
// The plaintext is never logged and errors only carry a stable reason, so a bad
// key, a mismatched corp id, or attacker-supplied data cannot leak secrets.

package callback

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"sort"
	"strings"
)

const (
	callbackKeySize     = 32
	callbackBlockSize   = 32
	callbackRandomSize  = 16
	callbackLengthSize  = 4
	callbackPrefixSize  = callbackRandomSize + callbackLengthSize
	callbackMaxMessage  = 1 << 20
	callbackMaxKeyChars = 128
)

var (
	errCallbackConfig    = errors.New("wecom callback: invalid configuration")
	errCallbackSignature = errors.New("wecom callback: signature mismatch")
	errCallbackEnvelope  = errors.New("wecom callback: invalid encrypted envelope")
	errCallbackDecrypt   = errors.New("wecom callback: decryption failed")
	errCallbackReceiveID = errors.New("wecom callback: receive id mismatch")
)

// CallbackCrypto verifies and decrypts WeCom callback requests and encrypts
// passive replies. It is immutable after construction and safe for concurrent
// use.
type CallbackCrypto struct {
	token  string
	aesKey []byte
	corpID string
}

// callbackCDATA emits a named element whose text is wrapped in CDATA. The XML
// package splits a literal "]]>" into two CDATA sections, so no value can
// escape its element.
type callbackCDATA struct {
	Value string `xml:",cdata"`
}

// NewCallbackCrypto validates the callback credentials. The encoding key must
// be the 43-character EncodingAESKey from the WeCom administration console and
// must decode to exactly 32 bytes; the corp id is the receive id checked after
// every decryption. Errors never contain the token, the key, or the corp id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90238
func NewCallbackCrypto(token, encodingAESKey, corpID string) (*CallbackCrypto, error) {
	token = strings.TrimSpace(token)
	corpID = strings.TrimSpace(corpID)
	key, err := decodeCallbackAESKey(encodingAESKey)
	if err != nil {
		return nil, err
	}
	if token == "" || corpID == "" {
		return nil, errCallbackConfig
	}
	return &CallbackCrypto{token: token, aesKey: key, corpID: corpID}, nil
}

// CorpID is the receive id this instance accepts. The callback endpoint uses it
// as FromUserName in passive replies.
func (c *CallbackCrypto) CorpID() string {
	if c == nil {
		return ""
	}
	return c.corpID
}

// decodeCallbackAESKey applies the documented padding rule: append "=" until
// the base64 text is a multiple of four characters, decode it, and require
// exactly 32 bytes.
func decodeCallbackAESKey(encodingAESKey string) ([]byte, error) {
	key := strings.TrimSpace(encodingAESKey)
	if key == "" || len(key) > callbackMaxKeyChars {
		return nil, errCallbackConfig
	}
	if remainder := len(key) % 4; remainder != 0 {
		key += strings.Repeat("=", 4-remainder)
	}
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(decoded) != callbackKeySize {
		return nil, errCallbackConfig
	}
	return decoded, nil
}

// VerifyURL validates the GET request that proves URL ownership and returns the
// decrypted echo string. The caller has already URL-decoded echostr.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90238
func (c *CallbackCrypto) VerifyURL(msgSignature, timestamp, nonce, echostr string) (string, error) {
	if c == nil || strings.TrimSpace(timestamp) == "" || strings.TrimSpace(nonce) == "" ||
		strings.TrimSpace(echostr) == "" {
		return "", errCallbackEnvelope
	}
	if err := c.verifySignature(msgSignature, timestamp, nonce, echostr); err != nil {
		return "", err
	}
	plain, err := c.decrypt(echostr)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// DecryptXML parses an XML encrypted envelope, verifies msg_signature over its
// Encrypt value, decrypts it, and returns the plaintext message.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90238
func (c *CallbackCrypto) DecryptXML(body []byte, msgSignature, timestamp, nonce string) ([]byte, error) {
	if c == nil {
		return nil, errCallbackEnvelope
	}
	var envelope struct {
		XMLName xml.Name `xml:"xml"`
		Encrypt string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(body, &envelope); err != nil {
		return nil, errCallbackEnvelope
	}
	return c.decryptEnvelope(envelope.Encrypt, msgSignature, timestamp, nonce)
}

// DecryptJSON parses a JSON encrypted envelope, verifies msg_signature over its
// Encrypt value, decrypts it, and returns the plaintext message. AgentID is
// accepted as either a string or a number because WeCom has used both shapes.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90238
func (c *CallbackCrypto) DecryptJSON(body []byte, msgSignature, timestamp, nonce string) ([]byte, error) {
	if c == nil {
		return nil, errCallbackEnvelope
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil, errCallbackEnvelope
	}
	var envelope struct {
		ToUserName string          `json:"ToUserName"`
		AgentID    json.RawMessage `json:"AgentID"`
		Encrypt    string          `json:"Encrypt"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, errCallbackEnvelope
	}
	return c.decryptEnvelope(envelope.Encrypt, msgSignature, timestamp, nonce)
}

func (c *CallbackCrypto) decryptEnvelope(encrypt, msgSignature, timestamp, nonce string) ([]byte, error) {
	encrypt = strings.TrimSpace(encrypt)
	if encrypt == "" || strings.TrimSpace(timestamp) == "" || strings.TrimSpace(nonce) == "" {
		return nil, errCallbackEnvelope
	}
	if err := c.verifySignature(msgSignature, timestamp, nonce, encrypt); err != nil {
		return nil, err
	}
	return c.decrypt(encrypt)
}

// verifySignature compares the hexadecimal SHA-1 signature in constant time.
// Both values are decoded first so an invalid signature cannot be distinguished
// from a mismatching one.
func (c *CallbackCrypto) verifySignature(msgSignature, timestamp, nonce, encrypt string) error {
	want, err := hex.DecodeString(callbackSignature(c.token, timestamp, nonce, encrypt))
	if err != nil {
		return errCallbackSignature
	}
	got, err := hex.DecodeString(strings.TrimSpace(msgSignature))
	if err != nil || len(got) != len(want) {
		return errCallbackSignature
	}
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return errCallbackSignature
	}
	return nil
}

// callbackSignature is SHA-1 over the four values sorted lexicographically and
// concatenated without separators, encoded as lower-case hexadecimal.
func callbackSignature(token, timestamp, nonce, encrypt string) string {
	parts := []string{token, timestamp, nonce, encrypt}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(sum[:])
}

// EncryptReply encrypts one plaintext reply and wraps it in the passive reply
// envelope. timestamp and nonce must be the values used to sign the reply; the
// callback endpoint reuses the request values.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90238
func (c *CallbackCrypto) EncryptReply(plain []byte, timestamp, nonce string) ([]byte, error) {
	if c == nil || strings.TrimSpace(timestamp) == "" || strings.TrimSpace(nonce) == "" {
		return nil, errCallbackEnvelope
	}
	encrypt, err := c.encrypt(plain)
	if err != nil {
		return nil, err
	}
	reply := struct {
		XMLName      xml.Name      `xml:"xml"`
		Encrypt      callbackCDATA `xml:"Encrypt"`
		MsgSignature callbackCDATA `xml:"MsgSignature"`
		TimeStamp    string        `xml:"TimeStamp"`
		Nonce        callbackCDATA `xml:"Nonce"`
	}{
		Encrypt:      callbackCDATA{Value: encrypt},
		MsgSignature: callbackCDATA{Value: callbackSignature(c.token, timestamp, nonce, encrypt)},
		TimeStamp:    timestamp,
		Nonce:        callbackCDATA{Value: nonce},
	}
	body, err := xml.Marshal(reply)
	if err != nil {
		return nil, errCallbackEnvelope
	}
	return body, nil
}

func (c *CallbackCrypto) encrypt(plain []byte) (string, error) {
	if len(plain) > callbackMaxMessage {
		return "", errCallbackEnvelope
	}
	random := make([]byte, callbackRandomSize)
	if _, err := rand.Read(random); err != nil {
		return "", errCallbackEnvelope
	}
	buffer := make([]byte, 0, callbackPrefixSize+len(plain)+len(c.corpID)+callbackBlockSize)
	buffer = append(buffer, random...)
	var length [callbackLengthSize]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(plain)))
	buffer = append(buffer, length[:]...)
	buffer = append(buffer, plain...)
	buffer = append(buffer, c.corpID...)
	padding := callbackBlockSize - len(buffer)%callbackBlockSize
	buffer = append(buffer, bytes.Repeat([]byte{byte(padding)}, padding)...)

	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return "", errCallbackConfig
	}
	ciphertext := make([]byte, len(buffer))
	cipher.NewCBCEncrypter(block, c.aesKey[:aes.BlockSize]).CryptBlocks(ciphertext, buffer)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *CallbackCrypto) decrypt(encrypt string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil || len(ciphertext) < callbackBlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return nil, errCallbackDecrypt
	}
	block, err := aes.NewCipher(c.aesKey)
	if err != nil {
		return nil, errCallbackConfig
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, c.aesKey[:aes.BlockSize]).CryptBlocks(plain, ciphertext)

	padding := int(plain[len(plain)-1])
	if padding < 1 || padding > callbackBlockSize || padding > len(plain) {
		return nil, errCallbackDecrypt
	}
	for _, value := range plain[len(plain)-padding:] {
		if int(value) != padding {
			return nil, errCallbackDecrypt
		}
	}
	plain = plain[:len(plain)-padding]
	if len(plain) < callbackPrefixSize {
		return nil, errCallbackDecrypt
	}
	messageLength := binary.BigEndian.Uint32(plain[callbackRandomSize:callbackPrefixSize])
	if uint64(messageLength) > uint64(len(plain)-callbackPrefixSize) {
		return nil, errCallbackDecrypt
	}
	messageEnd := callbackPrefixSize + int(messageLength)
	receiveID := string(plain[messageEnd:])
	if subtle.ConstantTimeCompare([]byte(receiveID), []byte(c.corpID)) != 1 {
		return nil, errCallbackReceiveID
	}
	return plain[callbackPrefixSize:messageEnd], nil
}
