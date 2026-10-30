package external

import (
	"context"
	"strings"
	"testing"
)

func TestAcquisitionLinks(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n","link_id_list":["L1"],`+
		`"link":{"link_id":"L1","link_name":"链接","url":"https://work.weixin.qq.com/ca/x","create_time":10},`+
		`"range":{"user_list":["u1"],"department_list":[2]},`+
		`"priority_option":{"priority_type":2,"priority_userid_list":["u2"]}}`)
	ctx := context.Background()

	page, err := api.ListAcquisitionLinks(ctx, "", 100)
	if err != nil || rec.path != "/cgi-bin/externalcontact/customer_acquisition/list_link" ||
		len(page.LinkIDs) != 1 || page.NextCursor != "n" {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}

	detail, err := api.GetAcquisitionLink(ctx, "L1")
	if err != nil || rec.path != "/cgi-bin/externalcontact/customer_acquisition/get" {
		t.Fatalf("get: %v %s", err, rec.path)
	}
	if detail.Link.URL == "" || detail.Range.UserList[0] != "u1" || detail.PriorityOption.PriorityType != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}

	created, err := api.CreateAcquisitionLink(ctx, AcquisitionLink{
		LinkName: "链接", Range: &AcquisitionRange{UserList: []string{"u1"}}, SkipVerify: true,
	})
	if err != nil || created.LinkID != "L1" ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/create_link" {
		t.Fatalf("create: %v %+v %s", err, created, rec.path)
	}
	if rec.body["skip_verify"] != true {
		t.Fatalf("skip_verify not sent: %v", rec.body)
	}

	if err := api.UpdateAcquisitionLink(ctx, AcquisitionLink{LinkID: "L1", LinkName: "改名"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/update_link" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := api.DeleteAcquisitionLink(ctx, "L1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/delete_link" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := api.CreateAcquisitionLink(ctx, AcquisitionLink{Range: &AcquisitionRange{UserList: []string{"u"}}}); err == nil {
		t.Fatal("missing link name accepted")
	}
	if _, err := api.CreateAcquisitionLink(ctx, AcquisitionLink{LinkName: "x"}); err == nil {
		t.Fatal("missing range accepted")
	}
	if err := api.UpdateAcquisitionLink(ctx, AcquisitionLink{}); err == nil {
		t.Fatal("update without id accepted")
	}
	if err := api.DeleteAcquisitionLink(ctx, " "); err == nil {
		t.Fatal("delete without id accepted")
	}
}

func TestAcquisitionCustomersQuotaAndUsage(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n",`+
		`"customer_list":[{"external_userid":"e1","userid":"u1","chat_status":0,"state":"A"}],`+
		`"total":1000,"balance":500,"quota_list":[{"expire_date":10,"balance":200}],`+
		`"click_link_customer_cnt":9,"new_customer_cnt":3,`+
		`"userid":"u1","external_userid":"e1","chat_info":{"recv_msg_cnt":3,"link_id":"L1"}}`)
	ctx := context.Background()

	customers, err := api.ListAcquisitionCustomers(ctx, "L1", "", 100)
	if err != nil || len(customers.Customers) != 1 || customers.Customers[0].State != "A" ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/customer" {
		t.Fatalf("customers: %v %+v %s", err, customers, rec.path)
	}
	if rec.body["link_id"] != "L1" {
		t.Fatalf("link not sent: %v", rec.body)
	}

	quota, err := api.AcquisitionQuota(ctx)
	if err != nil || quota.Balance != 500 || quota.QuotaList[0].Balance != 200 ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition_quota" {
		t.Fatalf("quota: %v %+v %s", err, quota, rec.path)
	}

	usage, err := api.AcquisitionUsage(ctx, "L1", 1, 2)
	if err != nil || usage.ClickLinkCustomerCnt != 9 || usage.NewCustomerCnt != 3 ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/statistic" {
		t.Fatalf("usage: %v %+v %s", err, usage, rec.path)
	}

	info, err := api.AcquisitionChatInfo(ctx, "ck")
	if err != nil || info.ChatInfo.RecvMsgCnt != 3 ||
		rec.path != "/cgi-bin/externalcontact/customer_acquisition/get_chat_info" {
		t.Fatalf("chat info: %v %+v %s", err, info, rec.path)
	}
	if rec.body["chat_key"] != "ck" {
		t.Fatalf("chat key not sent: %v", rec.body)
	}

	if _, err := api.ListAcquisitionCustomers(ctx, "", "", 0); err == nil {
		t.Fatal("missing link id accepted")
	}
	if _, err := api.AcquisitionChatInfo(ctx, ""); err == nil {
		t.Fatal("missing chat key accepted")
	}
}

func TestParseAcquisitionEvent(t *testing.T) {
	cases := map[string]string{
		AcqBalanceLow: `<xml><ToUserName>toUser</ToUserName><CreateTime>1</CreateTime>` +
			`<MsgType>event</MsgType><Event>customer_acquisition</Event>` +
			`<ChangeType>balance_low</ChangeType></xml>`,
		AcqOpenProfile: `<xml><MsgType>event</MsgType><Event>customer_acquisition</Event>` +
			`<ChangeType>open_profile</ChangeType><LinkId>ca1</LinkId><State>S1</State></xml>`,
		AcqQuotaExpireSoon: `<xml><MsgType>event</MsgType><Event>customer_acquisition</Event>` +
			`<ChangeType>quota_expire_soon</ChangeType><ExpireTime>100</ExpireTime>` +
			`<ExpireQuotaNum>50</ExpireQuotaNum></xml>`,
		AcqMessageFromCustomer: `<xml><MsgType>event</MsgType><Event>customer_acquisition</Event>` +
			`<ChangeType>message_from_customer</ChangeType><UserID>u1</UserID>` +
			`<ExternalUserID>e1</ExternalUserID><LinkId>ca1</LinkId></xml>`,
	}
	for want, payload := range cases {
		event, err := ParseAcquisitionEvent([]byte(payload))
		if err != nil {
			t.Fatalf("%s: %v", want, err)
		}
		if event.ChangeType != want {
			t.Fatalf("change type = %q, want %q", event.ChangeType, want)
		}
	}
	event, err := ParseAcquisitionEvent([]byte(cases[AcqMessageFromCustomer]))
	if err != nil || event.UserID != "u1" || event.ExternalUserID != "e1" || event.LinkID != "ca1" {
		t.Fatalf("unexpected event: %+v", event)
	}

	for name, payload := range map[string]string{
		"malformed":    "<xml>",
		"wrong event":  `<xml><Event>other</Event><ChangeType>balance_low</ChangeType></xml>`,
		"unknown type": `<xml><Event>customer_acquisition</Event><ChangeType>unknown</ChangeType></xml>`,
		"missing type": `<xml><Event>customer_acquisition</Event></xml>`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseAcquisitionEvent([]byte(payload)); err == nil {
				t.Fatal("invalid event accepted")
			}
		})
	}
	if !strings.Contains(cases[AcqQuotaExpireSoon], "quota_expire_soon") {
		t.Fatal("fixture mismatch")
	}
}

func TestMomentRecords(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"next_cursor":"n",`+
		`"moment_list":[{"moment_id":"m1","creator":"u1","text":{"content":"hi"}}],`+
		`"task_list":[{"userid":"u1","publish_status":1}],`+
		`"customer_list":[{"userid":"u1","external_userid":"e1"}],`+
		`"comment_list":[{"external_userid":"e1","create_time":10}],`+
		`"like_list":[{"userid":"u1","create_time":11}]}`)
	ctx := context.Background()

	records, err := api.ListMomentRecords(ctx, MomentQuery{Creator: "u1", Limit: 10})
	if err != nil || len(records.Moments) != 1 || records.NextCursor != "n" ||
		rec.path != "/cgi-bin/externalcontact/get_moment_list" {
		t.Fatalf("records: %v %+v %s", err, records, rec.path)
	}

	tasks, err := api.ListMomentTasks(ctx, "m1", "", 10)
	if err != nil || len(tasks.Tasks) != 1 || tasks.Tasks[0].PublishStatus != 1 ||
		rec.path != "/cgi-bin/externalcontact/get_moment_task" {
		t.Fatalf("tasks: %v %+v %s", err, tasks, rec.path)
	}

	receivers, err := api.ListMomentReceivers(ctx, "m1", "u1", "", 100)
	if err != nil || len(receivers.CustomerList) != 1 ||
		rec.path != "/cgi-bin/externalcontact/get_moment_customer_list" {
		t.Fatalf("receivers: %v %+v %s", err, receivers, rec.path)
	}

	sent, err := api.ListMomentSendResults(ctx, "m1", "u1", "", 100)
	if err != nil || len(sent.CustomerList) != 1 ||
		rec.path != "/cgi-bin/externalcontact/get_moment_send_result" {
		t.Fatalf("send results: %v %+v %s", err, sent, rec.path)
	}

	comments, err := api.GetMomentComments(ctx, "m1", "u1")
	if err != nil || len(comments.Comments) != 1 || len(comments.Likes) != 1 ||
		rec.path != "/cgi-bin/externalcontact/get_moment_comments" {
		t.Fatalf("comments: %v %+v %s", err, comments, rec.path)
	}

	if _, err := api.ListMomentTasks(ctx, "", "", 0); err == nil {
		t.Fatal("missing moment id accepted")
	}
	if _, err := api.ListMomentReceivers(ctx, "m1", "", "", 0); err == nil {
		t.Fatal("missing userid accepted")
	}
	if _, err := api.GetMomentComments(ctx, "m1", ""); err == nil {
		t.Fatal("missing userid accepted")
	}
}
