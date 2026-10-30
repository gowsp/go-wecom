package kf

import "testing"

const authChangeXML = `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
	`<FromUserName><![CDATA[sys]]></FromUserName>` +
	`<CreateTime>1348831860</CreateTime><MsgType><![CDATA[event]]></MsgType>` +
	`<Event><![CDATA[kf_account_auth_change]]></Event>` +
	`<AuthAddOpenKfId><![CDATA[wkA]]></AuthAddOpenKfId>` +
	`<AuthDelOpenKfId><![CDATA[wkB]]></AuthDelOpenKfId></xml>`

func TestParseAuthChangeEvent(t *testing.T) {
	event, err := ParseAuthChangeEvent([]byte(authChangeXML))
	if err != nil {
		t.Fatal(err)
	}
	if event.Event != EventKFAuthChange {
		t.Fatalf("unexpected event: %+v", event)
	}
	if len(event.AuthAddOpenKFID) != 1 || event.AuthAddOpenKFID[0] != "wkA" {
		t.Fatalf("added ids not parsed: %+v", event)
	}
	if len(event.AuthDelOpenKFID) != 1 || event.AuthDelOpenKFID[0] != "wkB" {
		t.Fatalf("removed ids not parsed: %+v", event)
	}
}

func TestParseAuthChangeEventRejectsBadPayload(t *testing.T) {
	if _, err := ParseAuthChangeEvent([]byte("<xml>")); err == nil {
		t.Fatal("malformed payload accepted")
	}
	if _, err := ParseAuthChangeEvent([]byte(`<xml><Event><![CDATA[other]]></Event></xml>`)); err == nil {
		t.Fatal("unexpected event accepted")
	}
}
