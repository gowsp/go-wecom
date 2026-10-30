package contact

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

// recorder answers gettoken and records each contact request.
type recorder struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func newContact(t *testing.T, response string) (*recorder, *Contact) {
	t.Helper()
	rec := &recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/cgi-bin/gettoken":
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
		default:
			rec.method = r.Method
			rec.path = r.URL.Path
			rec.query = r.URL.RawQuery
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&rec.body)
			}
			_, _ = w.Write([]byte(response))
		}
	}))
	t.Cleanup(server.Close)
	return rec, New(client.New("corp", "secret", client.WithBaseURL(server.URL)))
}

func TestCreateMember(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"errmsg":"ok"}`)
	member := Member{
		UserID: "zhangsan", Name: "张三", Mobile: "13800000000",
		Department: []int{1, 2}, MainDepartment: 1,
		ExtAttr: &ExtAttr{Attrs: []ExtAttrItem{NewExtAttrText("爱好", "阅读"), NewExtAttrWeb("主页", "https://x", "主页")}},
	}
	if err := api.CreateMember(context.Background(), member); err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodPost || rec.path != "/cgi-bin/user/create" {
		t.Fatalf("unexpected request %s %s", rec.method, rec.path)
	}
	if rec.body["userid"] != "zhangsan" || rec.body["name"] != "张三" {
		t.Fatalf("unexpected body: %v", rec.body)
	}
	if _, ok := rec.body["email"]; ok {
		t.Fatal("omitted field must not be sent")
	}
	attrs := rec.body["extattr"].(map[string]any)["attrs"].([]any)
	if len(attrs) != 2 {
		t.Fatalf("extattr not sent: %v", rec.body["extattr"])
	}
}

func TestCreateMemberValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	cases := map[string]Member{
		"no userid":     {Name: "n", Department: []int{1}},
		"no name":       {UserID: "u", Department: []int{1}},
		"no department": {UserID: "u", Name: "n"},
	}
	for name, member := range cases {
		t.Run(name, func(t *testing.T) {
			if err := api.CreateMember(ctx, member); err == nil {
				t.Fatal("invalid member accepted")
			}
		})
	}
}

func TestGetMember(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"userid":"zhangsan","name":"张三","department":[1],"status":1}`)
	member, err := api.GetMember(context.Background(), "zhangsan")
	if err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodGet || rec.path != "/cgi-bin/user/get" {
		t.Fatalf("unexpected request %s %s", rec.method, rec.path)
	}
	if !strings.Contains(rec.query, "userid=zhangsan") {
		t.Fatalf("query = %s", rec.query)
	}
	if member.Name != "张三" || member.Status != 1 || member.UserID != "zhangsan" {
		t.Fatalf("unexpected member: %+v", member)
	}
}

func TestListDepartmentMembers(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"userlist":[{"userid":"a","name":"A","department":[3]}]}`)
	members, err := api.ListDepartmentMembers(context.Background(), 3, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.query, "department_id=3") || !strings.Contains(rec.query, "fetch_child=1") {
		t.Fatalf("query = %s", rec.query)
	}
	if len(members) != 1 || members[0].UserID != "a" {
		t.Fatalf("unexpected members: %+v", members)
	}
}

func TestMemberIDPagination(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"next_cursor":"next","dept_user":[{"userid":"a","department":1}]}`)
	page, err := api.ListMemberIDs(context.Background(), "", 100)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/user/list_id" || rec.body["limit"] != float64(100) {
		t.Fatalf("unexpected request %v %v", rec.path, rec.body)
	}
	if page.NextCursor != "next" || len(page.DeptUsers) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func TestConvertAndLookup(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"openid":"o1","userid":"u1"}`)
	ctx := context.Background()

	openID, err := api.ConvertToOpenID(ctx, "u1")
	if err != nil || openID != "o1" || rec.path != "/cgi-bin/user/convert_to_openid" {
		t.Fatalf("convert to openid: %q %v %s", openID, err, rec.path)
	}
	userID, err := api.ConvertToUserID(ctx, "o1")
	if err != nil || userID != "u1" || rec.path != "/cgi-bin/user/convert_to_userid" {
		t.Fatalf("convert to userid: %q %v %s", userID, err, rec.path)
	}
	if _, err := api.GetUserIDByMobile(ctx, "13800000000"); err != nil ||
		rec.path != "/cgi-bin/user/getuserid" || rec.body["mobile"] != "13800000000" {
		t.Fatalf("get userid by mobile: %v %s %v", err, rec.path, rec.body)
	}
	if _, err := api.GetUserIDByEmail(ctx, "a@b.com", EmailTypePersonal); err != nil ||
		rec.path != "/cgi-bin/user/get_userid_by_email" ||
		rec.body["email_type"] != float64(2) {
		t.Fatalf("get userid by email: %v %s %v", err, rec.path, rec.body)
	}
}

func TestDeleteAndInvite(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"invaliduser":["x"]}`)
	ctx := context.Background()

	if err := api.DeleteMember(ctx, "u1"); err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodGet || !strings.Contains(rec.query, "userid=u1") {
		t.Fatalf("unexpected delete request %s %s", rec.method, rec.query)
	}
	if err := api.BatchDeleteMembers(ctx, []string{"a", "b"}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/user/batchdelete" {
		t.Fatalf("path = %s", rec.path)
	}
	list, ok := rec.body["useridlist"].([]any)
	if !ok || len(list) != 2 {
		t.Fatalf("useridlist not sent: %v", rec.body)
	}

	result, err := api.InviteMembers(ctx, []string{"a"}, []string{"1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/batch/invite" || result.InvalidUsers[0] != "x" {
		t.Fatalf("unexpected invite: %s %+v", rec.path, result)
	}
	if err := api.BatchDeleteMembers(ctx, nil); err == nil {
		t.Fatal("empty batch delete accepted")
	}
	if _, err := api.InviteMembers(ctx, nil, nil, nil); err == nil {
		t.Fatal("empty invite accepted")
	}
}

func TestAuthSuccessAndQRCode(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"join_qrcode":"https://qr"}`)
	ctx := context.Background()

	if err := api.AuthSuccess(ctx, "u1"); err != nil || rec.path != "/cgi-bin/user/authsucc" {
		t.Fatalf("auth success: %v %s", err, rec.path)
	}
	code, err := api.JoinQRCode(ctx, 2)
	if err != nil || code != "https://qr" || !strings.Contains(rec.query, "size_type=2") {
		t.Fatalf("join qr: %q %v %s", code, err, rec.query)
	}
}

func TestAPIErrorSurfaces(t *testing.T) {
	_, api := newContact(t, `{"errcode":60111,"errmsg":"user not found"}`)
	_, err := api.GetMember(context.Background(), "nobody")
	if err == nil || !strings.Contains(err.Error(), "60111") {
		t.Fatalf("errcode not surfaced: %v", err)
	}
}
