package contact

import (
	"context"
	"strings"
	"testing"
)

func TestCreateAndListRules(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"rule_ids":[10001],"rules":[{"rule_id":10001,"rule_type":1,"range":{"userid":["u1"]}}]}`)
	ctx := context.Background()

	ids, err := api.CreateRules(ctx, []Rule{{
		Type:  RuleTypeHide,
		Range: RuleRange{UserIDs: []string{"u1"}, PartyIDs: []int{2}, TagIDs: []int{3}},
	}})
	if err != nil || len(ids) != 1 || ids[0] != 10001 {
		t.Fatalf("create rules: %v %v", ids, err)
	}
	if rec.path != "/cgi-bin/contactrule/create" {
		t.Fatalf("path = %s", rec.path)
	}
	rules, ok := rec.body["rules"].([]any)
	if !ok || len(rules) != 1 {
		t.Fatalf("rules not sent: %v", rec.body)
	}
	first := rules[0].(map[string]any)
	if first["rule_type"] != float64(1) {
		t.Fatalf("unexpected rule: %v", first)
	}

	list, err := api.ListRules(ctx)
	if err != nil || rec.path != "/cgi-bin/contactrule/list" || len(list) != 1 {
		t.Fatalf("list rules: %v %s %+v", err, rec.path, list)
	}
	if list[0].ID != 10001 || list[0].Range.UserIDs[0] != "u1" {
		t.Fatalf("unexpected rule: %+v", list[0])
	}
}

func TestUpdateAndDeleteRules(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()

	if err := api.UpdateRules(ctx, []Rule{{ID: 10001, Range: RuleRange{UserIDs: []string{"u1"}}}}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/contactrule/update" {
		t.Fatalf("path = %s", rec.path)
	}
	if err := api.DeleteRules(ctx, []int{10001, 10002}); err != nil {
		t.Fatal(err)
	}
	ids, ok := rec.body["rule_ids"].([]any)
	if !ok || len(ids) != 2 || rec.path != "/cgi-bin/contactrule/delete" {
		t.Fatalf("unexpected delete: %s %v", rec.path, rec.body)
	}
}

func TestRuleValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	cases := map[string]func() error{
		"create empty": func() error { _, err := api.CreateRules(ctx, nil); return err },
		"create no type": func() error {
			_, err := api.CreateRules(ctx, []Rule{{Range: RuleRange{UserIDs: []string{"u"}}}})
			return err
		},
		"create no range": func() error { _, err := api.CreateRules(ctx, []Rule{{Type: 1}}); return err },
		"update no id":    func() error { return api.UpdateRules(ctx, []Rule{{Range: RuleRange{UserIDs: []string{"u"}}}}) },
		"delete empty":    func() error { return api.DeleteRules(ctx, nil) },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if fn() == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestExportJobs(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"jobid":"job1","status":2,"data_list":[{"url":"https://x","size":10,"md5":"m"}]}`)
	ctx := context.Background()
	key := strings.Repeat("a", 43)

	if _, err := api.ExportMembers(ctx, ExportRequest{EncodingAESKey: key}); err != nil ||
		rec.path != "/cgi-bin/export/simple_user" {
		t.Fatalf("export members: %v %s", err, rec.path)
	}
	if _, err := api.ExportMemberDetails(ctx, ExportRequest{EncodingAESKey: key, BlockSize: 1000000}); err != nil ||
		rec.path != "/cgi-bin/export/user" || rec.body["block_size"] != float64(1000000) {
		t.Fatalf("export details: %v %s %v", err, rec.path, rec.body)
	}
	if _, err := api.ExportDepartments(ctx, ExportRequest{EncodingAESKey: key}); err != nil ||
		rec.path != "/cgi-bin/export/department" {
		t.Fatalf("export departments: %v %s", err, rec.path)
	}
	if _, err := api.ExportTagMembers(ctx, 5, ExportRequest{EncodingAESKey: key}); err != nil ||
		rec.path != "/cgi-bin/export/taguser" || rec.body["tagid"] != float64(5) {
		t.Fatalf("export tag: %v %s %v", err, rec.path, rec.body)
	}

	result, err := api.ExportResult(ctx, "job1")
	if err != nil || rec.path != "/cgi-bin/export/get_result" {
		t.Fatalf("export result: %v %s", err, rec.path)
	}
	if result.Status != 2 || result.DataList[0].MD5 != "m" || result.DataList[0].Size != 10 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestExportValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	if _, err := api.ExportMembers(ctx, ExportRequest{EncodingAESKey: "short"}); err == nil {
		t.Fatal("short aes key accepted")
	}
	if _, err := api.ExportTagMembers(ctx, 0, ExportRequest{EncodingAESKey: strings.Repeat("a", 43)}); err == nil {
		t.Fatal("missing tagid accepted")
	}
	if _, err := api.ExportResult(ctx, " "); err == nil {
		t.Fatal("missing jobid accepted")
	}
}
