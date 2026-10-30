package wecom_test

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"testing"

	"github.com/gowsp/go-wecom/msgaudit"
)

func TestMsgAuditQueries(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,`+
		`"data":{"robot_id":"r1","name":"机器人A","creator_userid":"u1"},`+
		`"ids":["u1"],`+
		`"agreeinfo":[{"userid":"u1","agree_status":"Agree"}],`+
		`"roomname":"群","creator":"u1","members":[{"memberid":"u1"}]}`)
	svc := msgaudit.New(api)
	ctx := context.Background()

	robot, err := svc.RobotInfo(ctx, "r1")
	if err != nil || robot.Name != "机器人A" ||
		rec.path != "/cgi-bin/msgaudit/get_robot_info" {
		t.Fatalf("robot info: %v %+v %s", err, robot, rec.path)
	}

	users, err := svc.PermitUsers(ctx, 1)
	if err != nil || len(users) != 1 || rec.path != "/cgi-bin/msgaudit/get_permit_user_list" {
		t.Fatalf("permit: %v %v %s", err, users, rec.path)
	}

	single, err := svc.SingleAgree(ctx, [][2]string{{"u1", "o1"}})
	if err != nil || len(single) != 1 || rec.path != "/cgi-bin/msgaudit/check_single_agree" {
		t.Fatalf("single agree: %v %+v %s", err, single, rec.path)
	}
	if _, err := svc.RoomAgree(ctx, []string{"r1"}); err != nil ||
		rec.path != "/cgi-bin/msgaudit/check_room_agree" {
		t.Fatalf("room agree: %v %s", err, rec.path)
	}

	group, err := svc.GroupChat(ctx, "r1")
	if err != nil || group.RoomName != "群" || len(group.Members) != 1 ||
		rec.path != "/cgi-bin/msgaudit/groupchat/get" {
		t.Fatalf("group: %v %+v %s", err, group, rec.path)
	}

	if _, err := svc.RobotInfo(ctx, " "); err == nil {
		t.Fatal("robot info without id accepted")
	}
	if _, err := svc.SingleAgree(ctx, nil); err == nil {
		t.Fatal("single agree without pairs accepted")
	}
	if _, err := svc.RoomAgree(ctx, nil); err == nil {
		t.Fatal("room agree without rooms accepted")
	}
	if _, err := svc.GroupChat(ctx, ""); err == nil {
		t.Fatal("group without id accepted")
	}
}

func TestMsgAuditDecrypt(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	randomKey := []byte("0123456789abcdef0123456789abcdef")
	plain := []byte(`{"msgid":"m1","msgtype":"text","text":{"content":"hello"}}`)

	block, err := aes.NewCipher(randomKey[:32])
	if err != nil {
		t.Fatal(err)
	}
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(append([]byte{}, plain...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, randomKey[:16]).CryptBlocks(ciphertext, padded)

	sealedKey, err := rsa.EncryptPKCS1v15(rand.Reader, &key.PublicKey, randomKey)
	if err != nil {
		t.Fatal(err)
	}
	item := msgaudit.ChatItem{
		Seq: 1, MsgID: "m1", PublicKeyVersion: 1,
		EncryptRandomKey: base64.StdEncoding.EncodeToString(sealedKey),
		EncryptChatMsg:   base64.StdEncoding.EncodeToString(ciphertext),
	}

	svc := msgaudit.New(nil)
	decrypted, err := svc.Decrypt(item, func(version int) (*rsa.PrivateKey, error) {
		if version != 1 {
			t.Fatalf("unexpected version %d", version)
		}
		return key, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) != string(plain) {
		t.Fatalf("decrypted mismatch: %s", decrypted)
	}

	if _, err := svc.Decrypt(item, nil); err == nil {
		t.Fatal("decrypt without keys accepted")
	}
	if _, err := svc.Decrypt(item, func(int) (*rsa.PrivateKey, error) {
		other, genErr := rsa.GenerateKey(rand.Reader, 2048)
		if genErr != nil {
			t.Fatal(genErr)
		}
		return other, nil
	}); err == nil {
		t.Fatal("decrypt with wrong key accepted")
	}
	if _, err := msgaudit.DecryptData([]byte("short"), "AAAA"); err == nil {
		t.Fatal("short random key accepted")
	}
}

func TestMsgAuditEvents(t *testing.T) {
	approval := `<xml><MsgType>event</MsgType><Event><![CDATA[change_external_contact]]></Event>` +
		`<ChangeType><![CDATA[msg_audit_approved]]></ChangeType>` +
		`<UserID><![CDATA[u1]]></UserID><Status><![CDATA[1]]></Status></xml>`
	event, err := msgaudit.ParseApprovalEvent([]byte(approval))
	if err != nil || event.UserID != "u1" {
		t.Fatalf("approval: %v %+v", err, event)
	}
	if _, err := msgaudit.ParseApprovalEvent([]byte(
		`<xml><Event>change_external_contact</Event><ChangeType>other</ChangeType></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}

	notify := `<xml><MsgType>event</MsgType><Event><![CDATA[msgaudit_notify]]></Event>` +
		`<AgentID>2000004</AgentID></xml>`
	notification, err := msgaudit.ParseNotifyEvent([]byte(notify))
	if err != nil || notification.AgentID != 2000004 {
		t.Fatalf("notify: %v %+v", err, notification)
	}
	if _, err := msgaudit.ParseNotifyEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected notify accepted")
	}
}

// TestDecryptDataRejectsShortKey guards the slice bound: a 16..31 byte key used to
// panic instead of returning an error.
func TestDecryptDataRejectsShortKey(t *testing.T) {
	for _, key := range [][]byte{
		{},
		[]byte("0123456789abcdef"),
		[]byte("0123456789abcdef012345678"),
	} {
		if _, err := msgaudit.DecryptData(key, "AAAA"); err == nil {
			t.Fatalf("key of %d bytes accepted", len(key))
		}
	}
}
