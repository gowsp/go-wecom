package external

import (
	"context"
	"strings"
	"testing"
)

func TestProductAlbum(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"product_id":"p1","next_cursor":"n",`+
		`"product":{"product_id":"p1","description":"商品","price":100},`+
		`"product_list":[{"product_id":"p1","description":"商品","price":100}]}`)
	ctx := context.Background()

	id, err := api.AddProduct(ctx, Product{
		Description: "商品", Price: 100,
		Attachments: []ProductAttachment{{Type: "image", Image: &MediaAttachment{MediaID: "m"}}},
	})
	if err != nil || id != "p1" || rec.path != "/cgi-bin/externalcontact/add_product_album" {
		t.Fatalf("add: %q %v %s", id, err, rec.path)
	}

	product, err := api.GetProduct(ctx, "p1")
	if err != nil || product.Description != "商品" ||
		rec.path != "/cgi-bin/externalcontact/get_product_album" {
		t.Fatalf("get: %v %+v %s", err, product, rec.path)
	}

	page, err := api.ListProducts(ctx, "", 50)
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_product_album_list" ||
		len(page.Products) != 1 || page.NextCursor != "n" {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}

	if err := api.UpdateProduct(ctx, Product{ProductID: "p1", Description: "改", Price: 2}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/update_product_album" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := api.DeleteProducts(ctx, []string{"p1"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/delete_product_album" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := api.AddProduct(ctx, Product{Price: 1}); err == nil {
		t.Fatal("product without description accepted")
	}
	if err := api.UpdateProduct(ctx, Product{Description: "x"}); err == nil {
		t.Fatal("update without id accepted")
	}
	if err := api.DeleteProducts(ctx, nil); err == nil {
		t.Fatal("empty delete accepted")
	}
}

func TestInterceptRules(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"rule_id":"r1",`+
		`"rule":{"rule_id":"r1","rule_name":"规则","word_list":["敏感词"]},`+
		`"rule_list":[{"rule_id":"r1","rule_name":"规则"}]}`)
	ctx := context.Background()

	id, err := api.AddInterceptRule(ctx, InterceptRule{
		RuleName: "规则", WordList: []string{"敏感词"}, InterceptType: 1,
		ApplicableRange: &InterceptRange{UserList: []string{"u1"}},
	})
	if err != nil || id != "r1" || rec.path != "/cgi-bin/externalcontact/add_intercept_rule" {
		t.Fatalf("add: %q %v %s", id, err, rec.path)
	}

	rule, err := api.GetInterceptRule(ctx, "r1")
	if err != nil || rule.RuleName != "规则" ||
		rec.path != "/cgi-bin/externalcontact/get_intercept_rule" {
		t.Fatalf("get: %v %+v %s", err, rule, rec.path)
	}

	rules, err := api.ListInterceptRules(ctx)
	if err != nil || len(rules) != 1 || rec.path != "/cgi-bin/externalcontact/get_intercept_rule_list" {
		t.Fatalf("list: %v %+v %s", err, rules, rec.path)
	}

	if err := api.UpdateInterceptRule(ctx, InterceptRule{RuleID: "r1", RuleName: "改"},
		&InterceptRange{UserList: []string{"u2"}}, nil); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/update_intercept_rule" {
		t.Fatalf("path = %s", rec.path)
	}
	if _, ok := rec.body["add_applicable_range"]; !ok {
		t.Fatalf("range not sent: %v", rec.body)
	}
	if _, ok := rec.body["remove_applicable_range"]; ok {
		t.Fatalf("empty range must be omitted: %v", rec.body)
	}

	if err := api.DeleteInterceptRule(ctx, "r1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/del_intercept_rule" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := api.AddInterceptRule(ctx, InterceptRule{RuleName: "x"}); err == nil {
		t.Fatal("rule without words accepted")
	}
	if err := api.UpdateInterceptRule(ctx, InterceptRule{}, nil, nil); err == nil {
		t.Fatal("update without id accepted")
	}
}

func TestUploadAttachmentAndServedContacts(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"type":"image","media_id":"m1","created_at":10,`+
		`"next_cursor":"n","info_list":[{"is_customer":true,"external_userid":"e1","follow_userid":"u1","add_time":5}]}`)
	ctx := context.Background()

	uploaded, err := api.UploadAttachment(ctx, "image", "a.png", []byte("data"))
	if err != nil || uploaded.MediaID != "m1" || rec.path != "/cgi-bin/media/upload_attachment" {
		t.Fatalf("upload: %v %+v %s", err, uploaded, rec.path)
	}
	if !strings.Contains(rec.query, "type=image") {
		t.Fatalf("query = %s", rec.query)
	}
	if _, err := api.UploadAttachment(ctx, "doc", "a", []byte("x")); err == nil {
		t.Fatal("unsupported media type accepted")
	}

	page, err := api.ListServedContacts(ctx, "", 100)
	if err != nil || len(page.Info) != 1 || !page.Info[0].IsCustomer ||
		rec.path != "/cgi-bin/externalcontact/contact_list" {
		t.Fatalf("served contacts: %v %+v %s", err, page, rec.path)
	}
}

func TestStatistics(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"behavior_data":[{"stat_time":1,"chat_cnt":2}],`+
		`"total":1,"next_offset":1,"items":[{"owner":"u1","data":{"chat_total":3}}]}`)
	ctx := context.Background()

	data, err := api.UserBehaviorData(ctx, []string{"u1"}, nil, 100, 200)
	if err != nil || len(data) != 1 || data[0].ChatCnt != 2 ||
		rec.path != "/cgi-bin/externalcontact/get_user_behavior_data" {
		t.Fatalf("behavior: %v %+v %s", err, data, rec.path)
	}
	if rec.body["start_time"] != float64(100) || rec.body["end_time"] != float64(200) {
		t.Fatalf("window not sent: %v", rec.body)
	}

	page, err := api.GroupChatStatistic(ctx, GroupChatStatOptions{
		DayBeginTime: 10, OwnerFilter: []string{"u1"}, Limit: 100,
	})
	if err != nil || page.Total != 1 || rec.path != "/cgi-bin/externalcontact/groupchat/statistic" {
		t.Fatalf("group stat: %v %+v %s", err, page, rec.path)
	}
	owner := rec.body["owner_filter"].(map[string]any)
	if owner["userid_list"].([]any)[0] != "u1" {
		t.Fatalf("owner filter not sent: %v", rec.body)
	}

	if _, err := api.GroupChatStatisticByDay(ctx, GroupChatStatOptions{DayBeginTime: 10}); err == nil {
		t.Fatal("by-day without end accepted")
	}
	if _, err := api.UserBehaviorData(ctx, nil, nil, 1, 2); err == nil {
		t.Fatal("empty subject accepted")
	}
	if _, err := api.UserBehaviorData(ctx, []string{"u1"}, nil, 0, 0); err == nil {
		t.Fatal("invalid window accepted")
	}
}
