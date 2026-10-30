package mail

import "testing"

const appEmailXML = `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
	`<FromUserName><![CDATA[sys]]></FromUserName>` +
	`<CreateTime>1668831860</CreateTime><MsgType><![CDATA[event]]></MsgType>` +
	`<Event><![CDATA[app_email_change]]></Event>` +
	`<ChangeType><![CDATA[receive_email]]></ChangeType>` +
	`<Amount><![CDATA[2]]></Amount></xml>`

func TestParseChangeEvent(t *testing.T) {
	event, err := ParseChangeEvent([]byte(appEmailXML))
	if err != nil {
		t.Fatal(err)
	}
	if event.Event != "app_email_change" || event.ChangeType != ChangeReceiveEmail {
		t.Fatalf("unexpected event: %+v", event)
	}
	if event.Amount != 2 || event.FromUser != "sys" {
		t.Fatalf("unexpected payload: %+v", event)
	}
}

func TestParsePublicMailboxEvent(t *testing.T) {
	payload := `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
		`<FromUserName><![CDATA[sys]]></FromUserName>` +
		`<CreateTime>1668831860</CreateTime><MsgType><![CDATA[event]]></MsgType>` +
		`<Event><![CDATA[public_email_change]]></Event>` +
		`<ChangeType><![CDATA[receive_email]]></ChangeType>` +
		`<Id>1</Id><Amount>2</Amount></xml>`
	event, err := ParseChangeEvent([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if event.ID != "1" || event.Amount != 2 {
		t.Fatalf("unexpected payload: %+v", event)
	}
}

func TestParseChangeEventRejectsBadPayload(t *testing.T) {
	for name, payload := range map[string]string{
		"malformed": "<xml>",
		"missing":   `<xml><Event><![CDATA[app_email_change]]></Event></xml>`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseChangeEvent([]byte(payload)); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}
