package callback

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"sort"
	"strings"
	"testing"
)

const (
	testCallbackToken  = "callback-token"
	testCallbackCorpID = "ww-callback-corp"
	testCallbackTime   = "1700000000"
	testCallbackNonce  = "callback-nonce"
)

// testCallbackKey returns a deterministic 32-byte key and the 43-character
// EncodingAESKey form WeCom's console displays.
func testCallbackKey(t *testing.T) (string, []byte) {
	t.Helper()
	key := make([]byte, 32)
	for index := range key {
		key[index] = byte((index*17 + 11) % 256)
	}
	encodingKey := base64.RawStdEncoding.EncodeToString(key)
	if len(encodingKey) != 43 {
		t.Fatalf("test encoding key has %d characters, want 43", len(encodingKey))
	}
	return encodingKey, key
}

func testCallbackCrypto(t *testing.T) (*CallbackCrypto, []byte) {
	t.Helper()
	encodingKey, key := testCallbackKey(t)
	crypto, err := NewCallbackCrypto(testCallbackToken, encodingKey, testCallbackCorpID)
	if err != nil {
		t.Fatal(err)
	}
	return crypto, key
}

// testCallbackEncrypt is an independent implementation of the WeCom encryption
// documented for callbacks. Tests use it to build requests and to verify
// replies, so production code is never the only source of truth.
func testCallbackEncrypt(t *testing.T, key []byte, corpID string, plain []byte) string {
	t.Helper()
	random := make([]byte, 16)
	for index := range random {
		random[index] = byte(index + 1)
	}
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(plain)))
	buffer := make([]byte, 0, 16+4+len(plain)+len(corpID)+32)
	buffer = append(buffer, random...)
	buffer = append(buffer, length[:]...)
	buffer = append(buffer, plain...)
	buffer = append(buffer, corpID...)
	padding := 32 - len(buffer)%32
	buffer = append(buffer, bytes.Repeat([]byte{byte(padding)}, padding)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := make([]byte, len(buffer))
	cipher.NewCBCEncrypter(block, key[:aes.BlockSize]).CryptBlocks(ciphertext, buffer)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func testCallbackDecrypt(t *testing.T, key []byte, corpID, encrypt string) []byte {
	t.Helper()
	ciphertext, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		t.Fatal(err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	plain := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, key[:aes.BlockSize]).CryptBlocks(plain, ciphertext)
	padding := int(plain[len(plain)-1])
	if padding < 1 || padding > 32 || padding > len(plain) {
		t.Fatalf("invalid test padding %d", padding)
	}
	plain = plain[:len(plain)-padding]
	messageLength := binary.BigEndian.Uint32(plain[16:20])
	messageEnd := 20 + int(messageLength)
	if messageEnd > len(plain) {
		t.Fatal("test message length exceeds plaintext")
	}
	if got := string(plain[messageEnd:]); got != corpID {
		t.Fatalf("test receive id = %q, want %q", got, corpID)
	}
	return plain[20:messageEnd]
}

func testCallbackSignature(token, timestamp, nonce, encrypt string) string {
	parts := []string{token, timestamp, nonce, encrypt}
	sort.Strings(parts)
	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	return hex.EncodeToString(sum[:])
}

func TestNewCallbackCryptoValidation(t *testing.T) {
	encodingKey, _ := testCallbackKey(t)
	if _, err := NewCallbackCrypto(testCallbackToken, encodingKey, testCallbackCorpID); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, token, key, corpID string
	}{
		{"empty token", "", encodingKey, testCallbackCorpID},
		{"empty corp id", testCallbackToken, encodingKey, ""},
		{"empty key", testCallbackToken, "", testCallbackCorpID},
		{"not base64", testCallbackToken, "!!!not-base64!!!", testCallbackCorpID},
		{"too short", testCallbackToken, base64.RawStdEncoding.EncodeToString(make([]byte, 31)), testCallbackCorpID},
		{"too long", testCallbackToken, base64.RawStdEncoding.EncodeToString(make([]byte, 33)), testCallbackCorpID},
		{"padded but short", testCallbackToken, base64.StdEncoding.EncodeToString(make([]byte, 16)), testCallbackCorpID},
	} {
		t.Run(test.name, func(t *testing.T) {
			crypto, err := NewCallbackCrypto(test.token, test.key, test.corpID)
			if err == nil || crypto != nil {
				t.Fatal("invalid callback configuration must not construct a crypto")
			}
			for _, secret := range []string{test.token, test.key, test.corpID, encodingKey} {
				if secret != "" && strings.Contains(err.Error(), secret) {
					t.Fatalf("error leaks input: %v", err)
				}
			}
		})
	}
}

func TestCallbackVerifyURL(t *testing.T) {
	crypto, key := testCallbackCrypto(t)
	echo := "echo-plaintext-123"
	encrypt := testCallbackEncrypt(t, key, testCallbackCorpID, []byte(echo))
	signature := testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, encrypt)
	got, err := crypto.VerifyURL(signature, testCallbackTime, testCallbackNonce, encrypt)
	if err != nil || got != echo {
		t.Fatalf("VerifyURL = %q, %v", got, err)
	}
	if _, err := crypto.VerifyURL("deadbeef", testCallbackTime, testCallbackNonce, encrypt); !errors.Is(err, errCallbackSignature) {
		t.Fatalf("bad signature error = %v", err)
	}
	if _, err := crypto.VerifyURL(signature, "", testCallbackNonce, encrypt); !errors.Is(err, errCallbackEnvelope) {
		t.Fatalf("missing timestamp error = %v", err)
	}
	otherCorp := testCallbackEncrypt(t, key, "ww-other-corp", []byte(echo))
	if _, err := crypto.VerifyURL(
		testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, otherCorp),
		testCallbackTime, testCallbackNonce, otherCorp,
	); !errors.Is(err, errCallbackReceiveID) {
		t.Fatalf("receive id mismatch error = %v", err)
	}
}

func TestCallbackDecryptXMLAndJSON(t *testing.T) {
	crypto, key := testCallbackCrypto(t)
	plaintext := []byte(`<xml><ToUserName><![CDATA[ww-callback-corp]]></ToUserName><FromUserName><![CDATA[user-1]]></FromUserName><CreateTime>1700000000</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[hello]]></Content><MsgId>123456789</MsgId></xml>`)
	encrypt := testCallbackEncrypt(t, key, testCallbackCorpID, plaintext)
	signature := testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, encrypt)

	xmlEnvelope := []byte(`<xml><ToUserName><![CDATA[ww-callback-corp]]></ToUserName><AgentID><![CDATA[1000002]]></AgentID><Encrypt><![CDATA[` + encrypt + `]]></Encrypt></xml>`)
	got, err := crypto.DecryptXML(xmlEnvelope, signature, testCallbackTime, testCallbackNonce)
	if err != nil || !bytes.Equal(got, plaintext) {
		t.Fatalf("DecryptXML = %q, %v", got, err)
	}

	jsonEnvelope, err := json.Marshal(map[string]any{
		"ToUserName": testCallbackCorpID,
		"AgentID":    1000002,
		"Encrypt":    encrypt,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err = crypto.DecryptJSON(jsonEnvelope, signature, testCallbackTime, testCallbackNonce)
	if err != nil || !bytes.Equal(got, plaintext) {
		t.Fatalf("DecryptJSON = %q, %v", got, err)
	}
}

func TestCallbackDecryptRejectsInvalidInput(t *testing.T) {
	crypto, key := testCallbackCrypto(t)
	plaintext := []byte("<xml><MsgType>text</MsgType></xml>")
	encrypt := testCallbackEncrypt(t, key, testCallbackCorpID, plaintext)
	signature := testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, encrypt)
	xmlEnvelope := []byte(`<xml><Encrypt><![CDATA[` + encrypt + `]]></Encrypt></xml>`)
	jsonEnvelope := []byte(`{"Encrypt":"` + encrypt + `"}`)

	for _, test := range []struct {
		name      string
		decrypt   func() ([]byte, error)
		wantError error
	}{
		{"xml bad signature", func() ([]byte, error) {
			return crypto.DecryptXML(xmlEnvelope, testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, "other"), testCallbackTime, testCallbackNonce)
		}, errCallbackSignature},
		{"json bad signature", func() ([]byte, error) {
			return crypto.DecryptJSON(jsonEnvelope, "not-hex", testCallbackTime, testCallbackNonce)
		}, errCallbackSignature},
		{"xml malformed envelope", func() ([]byte, error) {
			return crypto.DecryptXML([]byte(`<xml>`), signature, testCallbackTime, testCallbackNonce)
		}, errCallbackEnvelope},
		{"json malformed envelope", func() ([]byte, error) {
			return crypto.DecryptJSON([]byte(`{"Encrypt":`), signature, testCallbackTime, testCallbackNonce)
		}, errCallbackEnvelope},
		{"json empty body", func() ([]byte, error) {
			return crypto.DecryptJSON(nil, signature, testCallbackTime, testCallbackNonce)
		}, errCallbackEnvelope},
		{"wrong receive id", func() ([]byte, error) {
			other := testCallbackEncrypt(t, key, "ww-other-corp", plaintext)
			return crypto.DecryptXML([]byte(`<xml><Encrypt>`+other+`</Encrypt></xml>`),
				testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, other), testCallbackTime, testCallbackNonce)
		}, errCallbackReceiveID},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := test.decrypt(); !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestCallbackEncryptReply(t *testing.T) {
	crypto, key := testCallbackCrypto(t)
	plaintext := []byte(`<xml><ToUserName><![CDATA[user-1]]></ToUserName><FromUserName><![CDATA[ww-callback-corp]]></FromUserName><CreateTime>1700000001</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[reply]]></Content></xml>`)
	body, err := crypto.EncryptReply(plaintext, testCallbackTime, testCallbackNonce)
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		XMLName      xml.Name `xml:"xml"`
		Encrypt      string   `xml:"Encrypt"`
		MsgSignature string   `xml:"MsgSignature"`
		TimeStamp    string   `xml:"TimeStamp"`
		Nonce        string   `xml:"Nonce"`
	}
	if err := xml.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Encrypt == "" || envelope.MsgSignature == "" {
		t.Fatalf("incomplete reply envelope: %s", body)
	}
	if envelope.TimeStamp != testCallbackTime || envelope.Nonce != testCallbackNonce {
		t.Fatalf("reply timestamp/nonce = %q/%q", envelope.TimeStamp, envelope.Nonce)
	}
	if want := testCallbackSignature(testCallbackToken, testCallbackTime, testCallbackNonce, envelope.Encrypt); envelope.MsgSignature != want {
		t.Fatalf("reply signature = %q, want %q", envelope.MsgSignature, want)
	}
	if got := testCallbackDecrypt(t, key, testCallbackCorpID, envelope.Encrypt); !bytes.Equal(got, plaintext) {
		t.Fatalf("reply plaintext = %q", got)
	}
	if _, err := crypto.EncryptReply(plaintext, "", testCallbackNonce); !errors.Is(err, errCallbackEnvelope) {
		t.Fatalf("missing timestamp error = %v", err)
	}
}
