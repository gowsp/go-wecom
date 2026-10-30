package wedoc

import (
	"strings"
	"testing"
)

const (
	docMemberXML = `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
		`<FromUserName><![CDATA[fromUser]]></FromUserName>` +
		`<CreateTime>1348831860</CreateTime><MsgType><![CDATA[event]]></MsgType>` +
		`<Event><![CDATA[doc_change]]></Event>` +
		`<ChangeType><![CDATA[doc_member_change]]></ChangeType>` +
		`<DocId><![CDATA[docA]]></DocId><DocId><![CDATA[docB]]></DocId></xml>`

	recordXML = `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
		`<FromUserName><![CDATA[fromUser]]></FromUserName>` +
		`<CreateTime>1348831860</CreateTime><MsgType><![CDATA[event]]></MsgType>` +
		`<Event><![CDATA[smart_sheet_change]]></Event>` +
		`<ChangeType><![CDATA[update_record]]></ChangeType>` +
		`<DocId><![CDATA[sheetDoc]]></DocId><SheetId><![CDATA[sheet1]]></SheetId>` +
		`<RecordId><![CDATA[r1]]></RecordId><RecordId><![CDATA[r2]]></RecordId></xml>`
)

func TestParseDocumentEvent(t *testing.T) {
	event, err := ParseDocumentEvent([]byte(docMemberXML))
	if err != nil {
		t.Fatal(err)
	}
	if event.Event != "doc_change" || event.ChangeType != ChangeDocMember {
		t.Fatalf("unexpected event: %+v", event)
	}
	if len(event.DocIDs) != 2 || event.DocIDs[0] != "docA" || event.DocIDs[1] != "docB" {
		t.Fatalf("document ids not preserved: %+v", event.DocIDs)
	}
	if event.CreateTime != 1348831860 {
		t.Fatalf("create time = %d", event.CreateTime)
	}
}

func TestParseDocumentEventForms(t *testing.T) {
	xml := strings.Replace(docMemberXML, ChangeDocMember, ChangeFormComplete, 1)
	xml = strings.Replace(xml, "<DocId><![CDATA[docA]]></DocId><DocId><![CDATA[docB]]></DocId>",
		"<FormId><![CDATA[formA]]></FormId>", 1)
	event, err := ParseDocumentEvent([]byte(xml))
	if err != nil {
		t.Fatal(err)
	}
	if len(event.FormIDs) != 1 || event.FormIDs[0] != "formA" || len(event.DocIDs) != 0 {
		t.Fatalf("form ids not parsed: %+v", event)
	}
}

func TestParseSmartSheetEvent(t *testing.T) {
	event, err := ParseSmartSheetEvent([]byte(recordXML))
	if err != nil {
		t.Fatal(err)
	}
	if event.SheetID != "sheet1" || len(event.RecordIDs) != 2 {
		t.Fatalf("unexpected event: %+v", event)
	}
	fieldXML := strings.Replace(recordXML, ChangeUpdateRecord, ChangeDeleteField, 1)
	fieldXML = strings.Replace(fieldXML,
		"<RecordId><![CDATA[r1]]></RecordId><RecordId><![CDATA[r2]]></RecordId>",
		"<FieldId><![CDATA[f1]]></FieldId>", 1)
	event, err = ParseSmartSheetEvent([]byte(fieldXML))
	if err != nil {
		t.Fatal(err)
	}
	if len(event.FieldIDs) != 1 || event.FieldIDs[0] != "f1" {
		t.Fatalf("field ids not parsed: %+v", event)
	}
}

func TestParseEventRejectsWrongPayload(t *testing.T) {
	cases := map[string]string{
		"malformed":   "<xml>",
		"missing":     `<xml><Event><![CDATA[doc_change]]></Event></xml>`,
		"wrong event": strings.Replace(docMemberXML, "doc_change", "other_event", 1),
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseDocumentEvent([]byte(payload)); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
	if _, err := ParseSmartSheetEvent([]byte(docMemberXML)); err == nil {
		t.Fatal("doc_change accepted by the smart sheet parser")
	}
}

// TestFieldChangeSpellings keeps the official "filed" spellings in place: the
// documentation publishes add_filed/update_filed/delete_filed, so a well meaning
// typo fix here would silently stop matching real callbacks.
func TestFieldChangeSpellings(t *testing.T) {
	cases := []string{ChangeAddField, ChangeUpdateField, ChangeDeleteField}
	want := []string{"add_filed", "update_filed", "delete_filed"}
	for i, change := range cases {
		if change != want[i] {
			t.Fatalf("change type %d = %q, want %q", i, change, want[i])
		}
		payload := strings.Replace(recordXML, ChangeUpdateRecord, change, 1)
		event, err := ParseSmartSheetEvent([]byte(payload))
		if err != nil || event.ChangeType != change {
			t.Fatalf("%s: %+v %v", change, event, err)
		}
	}
}
