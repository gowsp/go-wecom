package external

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

type recorder struct {
	method string
	path   string
	query  string
	body   map[string]any
}

func newExternal(t *testing.T, response string) (*recorder, *External) {
	t.Helper()
	rec := &recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		rec.method, rec.path, rec.query = r.Method, r.URL.Path, r.URL.RawQuery
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&rec.body)
		}
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)
	return rec, New(client.New("corp", "secret", client.WithBaseURL(server.URL)))
}

func TestFollowUsersAndCustomers(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"follow_user":["u1","u2"],"external_userid":["e1"]}`)
	ctx := context.Background()

	users, err := api.ListFollowUsers(ctx)
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_follow_user_list" || len(users) != 2 {
		t.Fatalf("follow users: %v %s %v", err, rec.path, users)
	}
	customers, err := api.ListCustomers(ctx, "u1")
	if err != nil || rec.path != "/cgi-bin/externalcontact/list" || !strings.Contains(rec.query, "userid=u1") {
		t.Fatalf("list customers: %v %s %s", err, rec.path, rec.query)
	}
	if len(customers) != 1 || customers[0] != "e1" {
		t.Fatalf("unexpected customers: %v", customers)
	}
	if _, err := api.ListCustomers(ctx, " "); err == nil {
		t.Fatal("empty userid accepted")
	}
}

func TestGetCustomer(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"external_contact":{"external_userid":"e1","name":"李四"},`+
		`"follow_user":[{"userid":"u1","remark":"客户","tags":["t1"]}],"next_cursor":"c2"}`)
	detail, err := api.GetCustomer(context.Background(), "e1", "c1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/get" ||
		!strings.Contains(rec.query, "external_userid=e1") || !strings.Contains(rec.query, "cursor=c1") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if detail.Customer.Name != "李四" || detail.FollowUser[0].Remark != "客户" || detail.NextCursor != "c2" {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestBatchCustomers(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"external_contact_list":[{"external_contact":{"external_userid":"e1"}}],"next_cursor":"n"}`)
	page, err := api.GetCustomersByUser(context.Background(), []string{"u1"}, "", 100)
	if err != nil || rec.path != "/cgi-bin/externalcontact/batch/get_by_user" {
		t.Fatalf("batch: %v %s", err, rec.path)
	}
	list, ok := rec.body["userid_list"].([]any)
	if !ok || len(list) != 1 || page.NextCursor != "n" || len(page.Details) != 1 {
		t.Fatalf("unexpected batch: %v %+v", rec.body, page)
	}
	if _, err := api.GetCustomersByUser(context.Background(), nil, "", 0); err == nil {
		t.Fatal("empty user list accepted")
	}
}

func TestRemarkAndProfile(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0}`)
	ctx := context.Background()
	err := api.RemarkCustomer(ctx, RemarkRequest{
		UserID: "u1", ExternalUserID: "e1", Remark: "VIP", RemarkMobiles: []string{"138"},
	})
	if err != nil || rec.path != "/cgi-bin/externalcontact/remark" || rec.body["remark"] != "VIP" {
		t.Fatalf("remark: %v %s %v", err, rec.path, rec.body)
	}
	if err := api.RemarkCustomer(ctx, RemarkRequest{UserID: "u1"}); err == nil {
		t.Fatal("missing external_userid accepted")
	}

	profile := NewExternalProfile([]ExternalAttr{
		{Type: 0, Name: "text", Text: &ExternalAttrText{Value: "v"}},
		{Type: 2, Name: "mini", MiniProgram: &ExternalAttrMini{AppID: "a", PagePath: "p", Title: "t"}},
	})
	if len(profile.Attrs) != 2 || profile.Attrs[1].MiniProgram.Title != "t" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
}

func TestCorpTags(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"tag_group":[{"group_id":"g1","group_name":"组","tag":[{"id":"t1","name":"标签"}]}]}`)
	ctx := context.Background()

	groups, err := api.ListCorpTags(ctx, TagFilter{GroupIDs: []string{"g1"}})
	if err != nil || rec.path != "/cgi-bin/externalcontact/get_corp_tag_list" || len(groups) != 1 {
		t.Fatalf("list tags: %v %s %+v", err, rec.path, groups)
	}
	if groups[0].Tags[0].Name != "标签" {
		t.Fatalf("unexpected tag group: %+v", groups[0])
	}

	if _, err := api.AddCorpTags(ctx, "g1", "", []CorpTag{{Name: "新标签"}}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/add_corp_tag" {
		t.Fatalf("add tags: %v %s", err, rec.path)
	}
	if _, err := api.AddCorpTags(ctx, "", "", nil); err == nil {
		t.Fatal("empty add accepted")
	}
	if err := api.EditCorpTag(ctx, "t1", "改名", 1); err != nil ||
		rec.path != "/cgi-bin/externalcontact/edit_corp_tag" || rec.body["name"] != "改名" {
		t.Fatalf("edit tag: %v %s %v", err, rec.path, rec.body)
	}
	if err := api.DeleteCorpTags(ctx, []string{"t1"}, nil); err != nil ||
		rec.path != "/cgi-bin/externalcontact/del_corp_tag" {
		t.Fatalf("delete tags: %v %s", err, rec.path)
	}
	if err := api.DeleteCorpTags(ctx, nil, nil); err == nil {
		t.Fatal("empty delete accepted")
	}
}

func TestMarkCustomerTags(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0}`)
	ctx := context.Background()
	if err := api.MarkCustomerTags(ctx, "u1", "e1", []string{"t1"}, []string{"t2"}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/externalcontact/mark_tag" ||
		len(rec.body["add_tag"].([]any)) != 1 || len(rec.body["remove_tag"].([]any)) != 1 {
		t.Fatalf("unexpected mark: %s %v", rec.path, rec.body)
	}
	if err := api.MarkCustomerTags(ctx, "u1", "e1", nil, nil); err == nil {
		t.Fatal("empty tag change accepted")
	}
}

func TestGroupChats(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"group_chat_list":[{"chat_id":"c1","status":0}],"next_cursor":"n"}`)
	ctx := context.Background()

	page, err := api.ListGroupChats(ctx, GroupChatFilter{
		StatusFilter: 1, OwnerFilter: &OwnerFilter{UserIDList: []string{"u1"}},
	}, "", 50)
	if err != nil || rec.path != "/cgi-bin/externalcontact/groupchat/list" || len(page.Chats) != 1 {
		t.Fatalf("list chats: %v %s %+v", err, rec.path, page)
	}
	if rec.body["status_filter"] != float64(1) {
		t.Fatalf("filter not sent: %v", rec.body)
	}

	rec, api = newExternal(t, `{"errcode":0,"group_chat":{"chat_id":"c1","owner":"u1","member_list":[{"userid":"e1","type":2,"join_scene":2}]}}`)
	chat, err := api.GetGroupChat(ctx, "c1", true)
	if err != nil || rec.path != "/cgi-bin/externalcontact/groupchat/get" {
		t.Fatalf("get chat: %v %s", err, rec.path)
	}
	if chat.Owner != "u1" || chat.MemberList[0].JoinScene != 2 || rec.body["need_name"] != true {
		t.Fatalf("unexpected chat: %+v (body %v)", chat, rec.body)
	}

	rec, api = newExternal(t, `{"errcode":0,"chat_id":"c1"}`)
	chatID, err := api.ConvertOpenGID(ctx, "open1")
	if err != nil || chatID != "c1" || rec.path != "/cgi-bin/externalcontact/opengid_to_chatid" {
		t.Fatalf("opengid: %q %v %s", chatID, err, rec.path)
	}
}

func TestJoinWays(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"config_id":"cfg1","join_way":{"config_id":"cfg1","chat_id_list":["c1"]}}`)
	ctx := context.Background()

	id, err := api.AddJoinWay(ctx, GroupChatJoinWay{ChatIDList: []string{"c1"}})
	if err != nil || id != "cfg1" || rec.path != "/cgi-bin/externalcontact/groupchat/add_join_way" {
		t.Fatalf("add join way: %q %v %s", id, err, rec.path)
	}
	if _, err := api.GetJoinWay(ctx, "cfg1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/groupchat/get_join_way" {
		t.Fatalf("get join way: %v %s", err, rec.path)
	}
	if err := api.UpdateJoinWay(ctx, GroupChatJoinWay{ConfigID: "cfg1"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/groupchat/update_join_way" {
		t.Fatalf("update join way: %v %s", err, rec.path)
	}
	if err := api.DeleteJoinWay(ctx, "cfg1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/groupchat/del_join_way" {
		t.Fatalf("delete join way: %v %s", err, rec.path)
	}
	if _, err := api.AddJoinWay(ctx, GroupChatJoinWay{}); err == nil {
		t.Fatal("empty chat list accepted")
	}
}

func TestContactWays(t *testing.T) {
	rec, api := newExternal(t, `{"errcode":0,"config_id":"w1","qr_code":"https://qr","contact_way":{"config_id":"w1"}}`)
	ctx := context.Background()

	way, err := api.AddContactWay(ctx, ContactWay{Type: 1, Scene: 2, User: []string{"u1"}})
	if err != nil || way.ConfigID != "w1" || way.QRCode != "https://qr" {
		t.Fatalf("add contact way: %+v %v", way, err)
	}
	if rec.path != "/cgi-bin/externalcontact/add_contact_way" {
		t.Fatalf("path = %s", rec.path)
	}
	if _, err := api.GetContactWay(ctx, "w1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/get_contact_way" ||
		!strings.Contains(rec.query, "config_id=w1") {
		t.Fatalf("get contact way: %v %s %s", err, rec.path, rec.query)
	}
	if err := api.UpdateContactWay(ctx, ContactWay{ConfigID: "w1"}); err != nil ||
		rec.path != "/cgi-bin/externalcontact/update_contact_way" {
		t.Fatalf("update contact way: %v %s", err, rec.path)
	}
	if err := api.DeleteContactWay(ctx, "w1"); err != nil ||
		rec.path != "/cgi-bin/externalcontact/del_contact_way" {
		t.Fatalf("delete contact way: %v %s", err, rec.path)
	}
	if _, err := api.AddContactWay(ctx, ContactWay{}); err == nil {
		t.Fatal("missing type/scene accepted")
	}
}
