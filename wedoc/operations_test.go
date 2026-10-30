package wedoc

import (
	"context"
	wxclient "github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/message"
	"strings"
	"testing"
)

// TestAllAppOperationsReturnAPIErrors exercises every public AppClient
// operation against an endpoint that answers errcode 40013, so each method is
// covered for path validation, error propagation, and the shared error
// summary contract.
func TestAllAppOperationsReturnAPIErrors(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{"errcode": 40013, "errmsg": "permission denied"}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))
	ctx := context.Background()

	operations := map[string]func() error{
		"create doc": func() error {
			_, err := client.CreateDoc(ctx, DocTypeSmartsheet, "文档")
			return err
		},
		"create doc with options": func() error {
			_, err := client.CreateDocWithOptions(ctx, DocTypeSmartsheet, "文档", CreateDocOptions{SpaceID: "s", FatherID: "f"})
			return err
		},
		"create smartsheet": func() error {
			_, err := client.CreateSmartsheet(ctx, "表格")
			return err
		},
		"rename doc": func() error { return client.RenameDoc(ctx, "doc-1", "新名字") },
		"delete doc": func() error { return client.DeleteDoc(ctx, "doc-1") },
		"get doc base info": func() error {
			_, err := client.GetDocBaseInfo(ctx, "doc-1")
			return err
		},
		"share doc": func() error {
			_, err := client.ShareDoc(ctx, "doc-1")
			return err
		},
		"get doc auth": func() error {
			_, err := client.GetDocAuth(ctx, "doc-1")
			return err
		},
		"set doc join rule": func() error {
			return client.SetDocJoinRule(ctx, "doc-1", DocJoinRule{BanShareExternal: boolPointer(true)})
		},
		"share doc with users": func() error {
			return client.ShareDocWithUsers(ctx, "doc-1", []UserPerm{{UserID: "u1", Auth: DocAuthRead}})
		},
		"remove doc members": func() error { return client.RemoveDocMembers(ctx, "doc-1", "u1") },
		"modify doc members": func() error {
			return client.ModifyDocMembers(ctx, "doc-1", []UserPerm{{UserID: "u1", Auth: DocAuthRead}}, []string{"u2"})
		},
		"add sheet": func() error {
			_, err := client.AddSheet(ctx, "doc-1", "表")
			return err
		},
		"update sheet": func() error { return client.UpdateSheet(ctx, "doc-1", "sheet-1", "新表") },
		"delete sheet": func() error { return client.DeleteSheet(ctx, "doc-1", "sheet-1") },
		"get sheets": func() error {
			_, err := client.GetSheets(ctx, "doc-1")
			return err
		},
		"add view": func() error {
			_, err := client.AddView(ctx, "doc-1", "sheet-1", View{ViewTitle: "视图", ViewType: ViewTypeGrid})
			return err
		},
		"update view": func() error {
			_, err := client.UpdateView(ctx, "doc-1", "sheet-1", View{ViewID: "v1", ViewTitle: "新视图"})
			return err
		},
		"delete views": func() error { return client.DeleteViews(ctx, "doc-1", "sheet-1", "v1") },
		"get views": func() error {
			_, err := client.GetViews(ctx, "doc-1", "sheet-1")
			return err
		},
		"add fields": func() error {
			_, err := client.AddFields(ctx, "doc-1", "sheet-1", Field{FieldTitle: "标题", FieldType: FieldTypeText})
			return err
		},
		"update fields": func() error {
			_, err := client.UpdateFields(ctx, "doc-1", "sheet-1", Field{FieldID: "f1", FieldTitle: "新标题", FieldType: FieldTypeText})
			return err
		},
		"delete fields": func() error { return client.DeleteFields(ctx, "doc-1", "sheet-1", "f1") },
		"get fields": func() error {
			_, err := client.GetFields(ctx, "doc-1", "sheet-1")
			return err
		},
		"add field group": func() error {
			_, err := client.AddFieldGroup(ctx, "doc-1", "sheet-1", "组", []string{"f1"})
			return err
		},
		"update field group": func() error {
			_, err := client.UpdateFieldGroup(ctx, "doc-1", "sheet-1", "g1", "组2", []string{"f2"})
			return err
		},
		"delete field groups": func() error { return client.DeleteFieldGroups(ctx, "doc-1", "sheet-1", "g1") },
		"get field groups": func() error {
			_, err := client.GetFieldGroups(ctx, "doc-1", "sheet-1")
			return err
		},
		"add records": func() error {
			_, err := client.AddRecords(ctx, "doc-1", "sheet-1", TextRecord(map[string]string{"标题": "内容"}))
			return err
		},
		"update records": func() error {
			_, err := client.UpdateRecords(ctx, "doc-1", "sheet-1", Record{RecordID: "r1", Values: map[string]any{"标题": TextCell("内容")}})
			return err
		},
		"delete records": func() error { return client.DeleteRecords(ctx, "doc-1", "sheet-1", "r1") },
		"get records": func() error {
			_, err := client.GetRecords(ctx, "doc-1", "sheet-1")
			return err
		},
		"query records": func() error {
			_, err := client.QueryRecords(ctx, "doc-1", "sheet-1", RecordQuery{ViewID: "v1", Limit: 10})
			return err
		},
		"get content privileges": func() error {
			_, err := client.GetContentPrivileges(ctx, "doc-1", ContentRuleExtra, []uint32{1})
			return err
		},
		"update content privilege": func() error {
			return client.UpdateContentPrivilege(ctx, "doc-1", ContentRule{
				Type: ContentRuleExtra, RuleID: 1,
				PrivList: []SheetPrivilege{{SheetID: "sheet-1", Priv: SheetPrivView}},
			})
		},
		"create content privilege rule": func() error {
			_, err := client.CreateContentPrivilegeRule(ctx, "doc-1", "规则")
			return err
		},
		"modify content privilege members": func() error {
			return client.ModifyContentPrivilegeMembers(ctx, "doc-1", 1, []string{"u1"}, nil)
		},
		"delete content privilege rules": func() error {
			return client.DeleteContentPrivilegeRules(ctx, "doc-1", 1)
		},
		"send text": func() error {
			return message.New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002).SendText(ctx, "zhangsan", "hello")
		},
		"send markdown": func() error {
			return message.New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)), 1000002).SendMarkdown(ctx, "zhangsan", "**hello**")
		},
	}

	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			err := operation()
			if err == nil {
				t.Fatal("expected an API error")
			}
			if !strings.Contains(err.Error(), "40013") || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("error does not summarize errcode/errmsg: %v", err)
			}
			for _, forbidden := range []string{"access-token-1", "corp-secret"} {
				if strings.Contains(err.Error(), forbidden) {
					t.Fatalf("error leaks %q: %v", forbidden, err)
				}
			}
		})
	}
}

// TestAllAppOperationsRetryInvalidToken is a narrower retry test: it proves the
// shared call path retries once for both document and smart sheet families.
// Token retry is implemented centrally in call(), which every operation uses.
func TestAllAppOperationsRetryInvalidToken(t *testing.T) {
	failures := map[string]int{
		"/cgi-bin/wedoc/doc_share":                           1,
		"/cgi-bin/wedoc/smartsheet/get_records":              1,
		"/cgi-bin/wedoc/smartsheet/content_priv/create_rule": 1,
	}
	// The handler fails the first request on each path, then succeeds.
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		if failures[call.Path] > 0 {
			failures[call.Path]--
			return map[string]any{"errcode": 42001, "errmsg": "expired"}
		}
		switch call.Path {
		case "/cgi-bin/wedoc/doc_share":
			return map[string]any{"errcode": 0, "errmsg": "ok", "share_url": "https://doc.example/1"}
		case "/cgi-bin/wedoc/smartsheet/get_records":
			return map[string]any{"errcode": 0, "errmsg": "ok", "records": []any{}, "has_more": false}
		case "/cgi-bin/wedoc/smartsheet/content_priv/create_rule":
			return map[string]any{"errcode": 0, "errmsg": "ok", "rule_id": 3}
		default:
			return map[string]any{"errcode": 0, "errmsg": "ok"}
		}
	})
	fake.tokens = []string{"token-1", "token-2", "token-3", "token-4"}
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))
	ctx := context.Background()

	if _, err := client.ShareDoc(ctx, "doc-1"); err != nil {
		t.Fatalf("share doc retry: %v", err)
	}
	if _, err := client.GetRecords(ctx, "doc-1", "sheet-1"); err != nil {
		t.Fatalf("get records retry: %v", err)
	}
	if _, err := client.CreateContentPrivilegeRule(ctx, "doc-1", "规则"); err != nil {
		t.Fatalf("create rule retry: %v", err)
	}
	tokenCalls, endpointCalls := fake.counts()
	// One initial token plus one refresh per failed operation.
	if tokenCalls != 4 || endpointCalls != 6 {
		t.Fatalf("token calls = %d, endpoint calls = %d, want 4 and 6", tokenCalls, endpointCalls)
	}
}
