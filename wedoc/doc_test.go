package wedoc

import (
	"context"
	"errors"
	wxclient "github.com/gowsp/go-wecom/client"
	"strings"
	"testing"
)

func boolPointer(value bool) *bool { return &value }
func intPointer(value int) *int    { return &value }

func newDocTestClient(t *testing.T, handler func(index int, call docAPICall) map[string]any) (*WeDoc, *fakeDocAPI) {
	t.Helper()
	fake := newFakeDocAPI(t, handler)
	return New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL))), fake
}

func TestRenameDocRequest(t *testing.T) {
	client, fake := newDocTestClient(t, nil)
	if err := client.RenameDoc(context.Background(), " doc-1 ", " 新名称 "); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/rename_doc" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" || calls[0].Body["new_name"] != "新名称" {
		t.Fatalf("unexpected body: %+v", calls[0].Body)
	}
}

func TestGetDocBaseInfoRequest(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"doc_base_info": map[string]any{
				"docid":       "doc-1",
				"doc_name":    "线索表",
				"create_time": 1717071093,
				"modify_time": 1717071100,
				"doc_type":    10,
			},
		}
	})
	info, err := client.GetDocBaseInfo(context.Background(), "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if info.DocID != "doc-1" || info.DocName != "线索表" || info.DocType != 10 || info.CreateTime != 1717071093 {
		t.Fatalf("unexpected info: %+v", info)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/get_doc_base_info" || calls[0].Body["docid"] != "doc-1" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
}

func TestShareDocRequest(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"share_url": "https://doc.example/share/abc",
		}
	})
	url, err := client.ShareDoc(context.Background(), "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://doc.example/share/abc" {
		t.Fatalf("share_url = %q", url)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/doc_share" || calls[0].Body["docid"] != "doc-1" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
}

func TestGetDocAuthRequest(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"access_rule": map[string]any{
				"enable_corp_internal":                true,
				"corp_internal_auth":                  2,
				"enable_corp_external":                false,
				"corp_external_auth":                  1,
				"ban_share_external":                  true,
				"corp_internal_approve_only_by_admin": true,
			},
			"secure_setting": map[string]any{
				"enable_readonly_copy": true,
				"watermark": map[string]any{
					"margin_type": 2, "show_text": true, "text": "Soma",
				},
			},
			"doc_member_list": []any{
				map[string]any{"type": 1, "userid": "admin-1", "auth": 7},
			},
			"co_auth_list": []any{
				map[string]any{"type": 2, "departmentid": 1, "auth": 1},
			},
		}
	})
	auth, err := client.GetDocAuth(context.Background(), "doc-1")
	if err != nil {
		t.Fatal(err)
	}
	if !auth.AccessRule.EnableCorpInternal || auth.AccessRule.CorpInternalAuth != 2 || !auth.AccessRule.BanShareExternal {
		t.Fatalf("unexpected access rule: %+v", auth.AccessRule)
	}
	if !auth.SecureSetting.EnableReadonlyCopy || auth.SecureSetting.Watermark.Text != "Soma" {
		t.Fatalf("unexpected secure setting: %+v", auth.SecureSetting)
	}
	if len(auth.DocMemberList) != 1 || auth.DocMemberList[0].UserID != "admin-1" || auth.DocMemberList[0].Auth != 7 {
		t.Fatalf("unexpected members: %+v", auth.DocMemberList)
	}
	if len(auth.CoAuthList) != 1 || auth.CoAuthList[0].DepartmentID != 1 {
		t.Fatalf("unexpected co-auth: %+v", auth.CoAuthList)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/doc_get_auth" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
}

func TestSetDocJoinRuleRequest(t *testing.T) {
	client, fake := newDocTestClient(t, nil)
	rule := DocJoinRule{
		EnableCorpInternal:             boolPointer(true),
		CorpInternalAuth:               intPointer(1),
		EnableCorpExternal:             boolPointer(false),
		CorpExternalAuth:               intPointer(2),
		CorpInternalApproveOnlyByAdmin: boolPointer(true),
		CorpExternalApproveOnlyByAdmin: boolPointer(false),
		BanShareExternal:               boolPointer(true),
		UpdateCoAuthList:               boolPointer(true),
		CoAuthList:                     []CoAuth{},
	}
	if err := client.SetDocJoinRule(context.Background(), "doc-1", rule); err != nil {
		t.Fatal(err)
	}
	if err := client.SetDocJoinRule(context.Background(), "doc-1", DocJoinRule{BanShareExternal: boolPointer(false)}); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 2 || calls[0].Path != "/cgi-bin/wedoc/mod_doc_join_rule" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	body := calls[0].Body
	if body["docid"] != "doc-1" || body["enable_corp_internal"] != true || body["corp_internal_auth"] != float64(1) {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body["enable_corp_external"] != false || body["corp_external_auth"] != float64(2) {
		t.Fatalf("unexpected body: %+v", body)
	}
	if body["corp_internal_approve_only_by_admin"] != true || body["update_co_auth_list"] != true {
		t.Fatalf("unexpected body: %+v", body)
	}
	if list, ok := body["co_auth_list"].([]any); !ok || len(list) != 0 {
		t.Fatalf("co_auth_list = %#v, want empty array", body["co_auth_list"])
	}
	if _, ok := calls[1].Body["enable_corp_internal"]; ok {
		t.Fatalf("nil fields must stay absent: %+v", calls[1].Body)
	}
	if calls[1].Body["ban_share_external"] != false {
		t.Fatalf("explicit false must be sent: %+v", calls[1].Body)
	}
}

func TestCreateDocWithOptionsRequest(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"docid": "doc-1", "url": "https://doc.example/doc-1",
		}
	})
	document, err := client.CreateDocWithOptions(context.Background(), DocTypeSmartsheet, " 线索表 ", CreateDocOptions{
		SpaceID:    " space-1 ",
		FatherID:   " father-1 ",
		AdminUsers: []string{" admin-1 ", "", "admin-1", "admin-2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if document.DocID != "doc-1" || document.URL != "https://doc.example/doc-1" {
		t.Fatalf("unexpected document: %+v", document)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/create_doc" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	body := calls[0].Body
	if body["spaceid"] != "space-1" || body["fatherid"] != "father-1" {
		t.Fatalf("unexpected space: %+v", body)
	}
	if body["doc_type"] != float64(DocTypeSmartsheet) || body["doc_name"] != "线索表" {
		t.Fatalf("unexpected document fields: %+v", body)
	}
	admins := arrayField(t, body["admin_users"], "admin_users")
	if len(admins) != 2 || admins[0] != "admin-1" || admins[1] != "admin-2" {
		t.Fatalf("admin_users = %v", admins)
	}
}

func TestModifyAndRemoveDocMembers(t *testing.T) {
	client, fake := newDocTestClient(t, nil)
	if err := client.ModifyDocMembers(context.Background(), "doc-1",
		[]UserPerm{{UserID: " admin-1 ", Auth: DocAuthAdmin}},
		[]string{" user-2 "}); err != nil {
		t.Fatal(err)
	}
	if err := client.RemoveDocMembers(context.Background(), "doc-1", "user-3", "user-4"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	for index, call := range calls {
		if call.Path != "/cgi-bin/wedoc/mod_doc_member" || call.Body["docid"] != "doc-1" {
			t.Fatalf("call %d = %+v", index, call)
		}
	}
	update := arrayField(t, calls[0].Body["update_file_member_list"], "update_file_member_list")
	member := objectField(t, update[0], "member")
	if member["type"] != float64(1) || member["auth"] != float64(7) || member["userid"] != "admin-1" {
		t.Fatalf("unexpected update member: %+v", member)
	}
	remove := arrayField(t, calls[0].Body["del_file_member_list"], "del_file_member_list")
	removed := objectField(t, remove[0], "removed")
	if removed["type"] != float64(1) || removed["userid"] != "user-2" {
		t.Fatalf("unexpected removed member: %+v", removed)
	}
	if _, ok := removed["auth"]; ok {
		t.Fatalf("removed member must not carry auth: %+v", removed)
	}
	if _, ok := calls[1].Body["update_file_member_list"]; ok {
		t.Fatalf("remove-only call unexpectedly updated members: %+v", calls[1].Body)
	}
	onlyRemove := arrayField(t, calls[1].Body["del_file_member_list"], "del_file_member_list")
	if len(onlyRemove) != 2 {
		t.Fatalf("unexpected remove list: %v", onlyRemove)
	}
}

func TestDocRequestTokenRetry(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		if index == 1 {
			return map[string]any{"errcode": 40014, "errmsg": "invalid access token"}
		}
		return map[string]any{"errcode": 0, "errmsg": "ok"}
	})
	fake.tokens = []string{"token-1", "token-2"}
	if err := client.RenameDoc(context.Background(), "doc-1", "新名称"); err != nil {
		t.Fatalf("retry after token refresh: %v", err)
	}
	tokenCalls, endpointCalls := fake.counts()
	if tokenCalls != 2 || endpointCalls != 2 {
		t.Fatalf("token calls = %d, endpoint calls = %d, want 2 each", tokenCalls, endpointCalls)
	}
	calls := fake.recorded()
	if calls[0].Token != "token-1" || calls[1].Token != "token-2" {
		t.Fatalf("retry tokens = %v", []string{calls[0].Token, calls[1].Token})
	}
}

func TestDocErrorsHideSecrets(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{"errcode": 40013, "errmsg": "invalid corp-secret"}
	})
	_ = fake
	_, err := client.ShareDoc(context.Background(), "doc-1")
	if err == nil {
		t.Fatal("expected an API error")
	}
	var apiErr *wxclient.APIError
	if !errors.As(err, &apiErr) || apiErr.Errcode != 40013 {
		t.Fatalf("expected APIError 40013, got %v", err)
	}
	for _, forbidden := range []string{"corp-secret", "access-token-1"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaks %q: %v", forbidden, err)
		}
	}
}
