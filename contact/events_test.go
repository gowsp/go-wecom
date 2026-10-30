package contact

import "testing"

const createUserXML = `<xml><ToUserName><![CDATA[toUser]]></ToUserName>` +
	`<FromUserName><![CDATA[sys]]></FromUserName><CreateTime>1403610513</CreateTime>` +
	`<MsgType><![CDATA[event]]></MsgType><Event><![CDATA[change_contact]]></Event>` +
	`<ChangeType>create_user</ChangeType><UserID><![CDATA[zhangsan]]></UserID>` +
	`<Name><![CDATA[张三]]></Name><Department><![CDATA[1,2,3]]></Department>` +
	`<MainDepartment>1</MainDepartment><IsLeaderInDept><![CDATA[1,0,0]]></IsLeaderInDept>` +
	`<DirectLeader><![CDATA[lisi,wangwu]]></DirectLeader>` +
	`<Position><![CDATA[产品经理]]></Position><Status>1</Status></xml>`

func TestParseChangeEventMember(t *testing.T) {
	event, err := ParseChangeEvent([]byte(createUserXML))
	if err != nil {
		t.Fatal(err)
	}
	if event.ChangeType != ChangeCreateUser || event.UserID != "zhangsan" {
		t.Fatalf("unexpected event: %+v", event)
	}
	if len(event.Department) != 3 || event.Department[2] != 3 {
		t.Fatalf("department not split: %+v", event.Department)
	}
	if len(event.IsLeaderInDept) != 3 || event.IsLeaderInDept[0] != 1 {
		t.Fatalf("leader flags not split: %+v", event.IsLeaderInDept)
	}
	if len(event.DirectLeader) != 2 || event.DirectLeader[1] != "wangwu" {
		t.Fatalf("direct leaders not split: %+v", event.DirectLeader)
	}
}

func TestParseChangeEventDepartmentAndTag(t *testing.T) {
	party := `<xml><MsgType>event</MsgType><Event>change_contact</Event>` +
		`<ChangeType>update_party</ChangeType><Id>2</Id><Name>研发</Name>` +
		`<ParentId>1</ParentId><Order>4</Order></xml>`
	event, err := ParseChangeEvent([]byte(party))
	if err != nil {
		t.Fatal(err)
	}
	if event.ChangeType != ChangeUpdateParty || event.ID != 2 || event.ParentID != 1 || event.Order != 4 {
		t.Fatalf("unexpected department event: %+v", event)
	}

	tag := `<xml><MsgType>event</MsgType><Event>change_contact</Event>` +
		`<ChangeType>update_tag</ChangeType><TagId>1</TagId>` +
		`<AddUserItems><![CDATA[zhangsan,lisi]]></AddUserItems>` +
		`<DelUserItems><![CDATA[wangwu]]></DelUserItems>` +
		`<AddPartyItems><![CDATA[1,2]]></AddPartyItems></xml>`
	event, err = ParseChangeEvent([]byte(tag))
	if err != nil {
		t.Fatal(err)
	}
	if event.TagID != 1 || len(event.AddUsers) != 2 || event.DelUsers[0] != "wangwu" ||
		len(event.AddParties) != 2 {
		t.Fatalf("unexpected tag event: %+v", event)
	}
}

func TestParseChangeEventRejects(t *testing.T) {
	for name, payload := range map[string]string{
		"malformed":   "<xml>",
		"wrong type":  `<xml><Event>change_contact</Event><ChangeType></ChangeType></xml>`,
		"wrong event": `<xml><Event>other</Event><ChangeType>create_user</ChangeType></xml>`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseChangeEvent([]byte(payload)); err == nil {
				t.Fatal("invalid payload accepted")
			}
		})
	}
}

func TestParseBatchJobEvent(t *testing.T) {
	payload := `<xml><ToUserName>corp</ToUserName><FromUserName>sys</FromUserName>` +
		`<CreateTime>1425284517</CreateTime><MsgType>event</MsgType>` +
		`<Event><![CDATA[batch_job_result]]></Event>` +
		`<BatchJob><JobId><![CDATA[job1]]></JobId><JobType><![CDATA[sync_user]]></JobType>` +
		`<ErrCode>0</ErrCode><ErrMsg><![CDATA[ok]]></ErrMsg></BatchJob></xml>`
	event, err := ParseBatchJobEvent([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	if event.Job.JobID != "job1" || event.Job.JobType != "sync_user" || event.Job.ErrCode != 0 {
		t.Fatalf("unexpected job: %+v", event.Job)
	}
	if _, err := ParseBatchJobEvent([]byte(`<xml><Event>batch_job_result</Event></xml>`)); err == nil {
		t.Fatal("job without id accepted")
	}
}
