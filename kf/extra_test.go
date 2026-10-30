package kf

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

type probe struct {
	Method string
	Path   string
	Query  string
	Body   map[string]any
}

func newProbe(t *testing.T, reply func(w http.ResponseWriter, p probe)) (*KF, *[]probe) {
	t.Helper()
	var calls []probe
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		p := probe{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery}
		if body, _ := io.ReadAll(r.Body); len(body) > 0 {
			_ = json.Unmarshal(body, &p.Body)
		}
		calls = append(calls, p)
		reply(w, p)
	}))
	t.Cleanup(server.Close)
	return New(client.New("corp", "secret", client.WithBaseURL(server.URL))), &calls
}

func TestAccountLifecyclePaths(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		if strings.HasSuffix(p.Path, "/add") {
			_, _ = w.Write([]byte(`{"errcode":0,"open_kfid":"KF1"}`))
			return
		}
		if strings.HasSuffix(p.Path, "/list") {
			_, _ = w.Write([]byte(`{"errcode":0,"account_list":[{"open_kfid":"KF1","name":"客服"}]}`))
			return
		}
		if strings.Contains(p.Path, "add_contact_way") {
			_, _ = w.Write([]byte(`{"errcode":0,"url":"https://work.weixin.qq.com/kf/x"}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	})
	ctx := context.Background()
	id, err := c.AddAccount(ctx, " 客服 ", "")
	if err != nil || id != "KF1" {
		t.Fatalf("add: %v %q", err, id)
	}
	accounts, err := c.ListAccounts(ctx, -1, 0)
	if err != nil || len(accounts) != 1 || accounts[0].Name != "客服" {
		t.Fatalf("list: %+v %v", accounts, err)
	}
	if err := c.UpdateAccount(ctx, "KF1", "新名", ""); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteAccount(ctx, "KF1"); err != nil {
		t.Fatal(err)
	}
	link, err := c.AccountLink(ctx, "KF1", "scene1")
	if err != nil || link == "" {
		t.Fatalf("link: %v %q", err, link)
	}
	if (*calls)[1].Body["limit"] != float64(100) {
		t.Fatalf("list limit default: %+v", (*calls)[1].Body)
	}
	if err := c.DeleteAccount(ctx, "  "); err == nil {
		t.Fatal("blank id must fail")
	}
}

func TestEventMessageAndUpgrade(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		switch {
		case strings.Contains(p.Path, "send_msg_on_event"):
			_, _ = w.Write([]byte(`{"errcode":0,"msgid":"M1"}`))
		case strings.Contains(p.Path, "get_upgrade_service_config"):
			_, _ = w.Write([]byte(`{"errcode":0,"member_range":{"userid_list":["a"],"department_id_list":[1]},"groupchat_range":{"chat_id_list":["c1"]}}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	})
	ctx := context.Background()
	id, err := c.SendEventText(ctx, "CODE1", "欢迎")
	if err != nil || id != "M1" {
		t.Fatalf("event text: %v %q", err, id)
	}
	if (*calls)[0].Body["msgtype"] != "text" || (*calls)[0].Body["code"] != "CODE1" {
		t.Fatalf("unexpected body: %+v", (*calls)[0].Body)
	}
	if _, err := c.SendEventText(ctx, "", "x"); err == nil {
		t.Fatal("missing code must fail")
	}

	config, err := c.UpgradeConfig(ctx)
	if err != nil || len(config.MemberRange.UserIDs) != 1 || len(config.GroupChatRange.ChatIDs) != 1 {
		t.Fatalf("upgrade config: %+v %v", config, err)
	}
	if (*calls)[1].Method != http.MethodGet {
		t.Fatalf("upgrade config must use GET: %+v", (*calls)[1])
	}

	if err := c.UpgradeService(ctx, UpgradeRequest{OpenKFID: "k", ExternalUserID: "u", Type: UpgradeToMember}); err == nil {
		t.Fatal("missing member must fail")
	}
	if err := c.UpgradeService(ctx, UpgradeRequest{OpenKFID: "k", ExternalUserID: "u", Type: UpgradeToMember,
		Member: &UpgradeMember{UserID: "zhangsan"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.CancelUpgradeService(ctx, "k", "u"); err != nil {
		t.Fatal(err)
	}
}

func TestStatisticAndCustomer(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		switch {
		case strings.Contains(p.Path, "get_corp_statistic"):
			_, _ = w.Write([]byte(`{"errcode":0,"statistic_list":[{"stat_time":1,"statistic":{"session_cnt":2,"ai_transfer_rate":0.5}}]}`))
		case strings.Contains(p.Path, "get_servicer_statistic"):
			_, _ = w.Write([]byte(`{"errcode":0,"statistic_list":[{"stat_time":1,"statistic":{"session_cnt":1,"first_reply_average_sec":17}}]}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0,"customer_list":[{"external_userid":"wm1","nickname":"张三"}],"invalid_external_userid":["bad"]}`))
		}
	})
	ctx := context.Background()
	corp, err := c.CorpStatistic(ctx, "k", 1, 2)
	if err != nil || len(corp) != 1 || corp[0].Statistic.SessionCount != 2 || corp[0].Statistic.AITransferRate != 0.5 {
		t.Fatalf("corp: %+v %v", corp, err)
	}
	servicer, err := c.ServicerStatistic(ctx, "k", "zhangsan", 1, 2)
	if err != nil || len(servicer) != 1 || servicer[0].Statistic.FirstReplyAverageSec != 17 {
		t.Fatalf("servicer: %+v %v", servicer, err)
	}
	if _, err := c.ServicerStatistic(ctx, "k", "", 1, 2); err == nil {
		t.Fatal("missing servicer must fail")
	}
	if _, err := c.CorpStatistic(ctx, "k", 5, 1); err == nil {
		t.Fatal("reversed range must fail")
	}

	customers, invalid, err := c.BatchCustomers(ctx, []string{"wm1", "bad"}, true)
	if err != nil || len(customers) != 1 || customers[0].Nickname != "张三" || len(invalid) != 1 {
		t.Fatalf("customers: %+v %+v %v", customers, invalid, err)
	}
	last := (*calls)[len(*calls)-1]
	if last.Body["need_enter_session_context"] != float64(1) {
		t.Fatalf("unexpected body: %+v", last.Body)
	}
}

func TestKnowledgeGroupsAndIntents(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		switch {
		case strings.Contains(p.Path, "add_group"):
			_, _ = w.Write([]byte(`{"errcode":0,"group_id":"G1"}`))
		case strings.Contains(p.Path, "add_intent"):
			_, _ = w.Write([]byte(`{"errcode":0,"intent_id":"I1"}`))
		case strings.Contains(p.Path, "list_group"):
			_, _ = w.Write([]byte(`{"errcode":0,"next_cursor":"N","has_more":1,"group_list":[{"group_id":"G1","name":"分组"}]}`))
		case strings.Contains(p.Path, "list_intent"):
			_, _ = w.Write([]byte(`{"errcode":0,"next_cursor":"N","has_more":0,"intent_list":[{"group_id":"G1","intent_id":"I1","question":{"text":{"content":"问"}},"answers":[{"text":{"content":"答"}}]}]}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	})
	ctx := context.Background()
	groupID, err := c.AddKnowledgeGroup(ctx, "分组")
	if err != nil || groupID != "G1" {
		t.Fatalf("add group: %v %q", err, groupID)
	}
	if err := c.UpdateKnowledgeGroup(ctx, "G1", "新名"); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteKnowledgeGroup(ctx, "G1"); err != nil {
		t.Fatal(err)
	}
	groups, err := c.ListKnowledgeGroups(ctx, "", 0, "")
	if err != nil || len(groups.Groups) != 1 || groups.HasMore != 1 {
		t.Fatalf("list group: %+v %v", groups, err)
	}

	intent := KnowledgeIntent{
		GroupID:  "G1",
		Question: KnowledgeTextBody{Text: KnowledgeText{Content: "问"}},
		Answers:  []KnowledgeAnswer{{Text: KnowledgeText{Content: "答"}}},
	}
	intentID, err := c.AddKnowledgeIntent(ctx, intent)
	if err != nil || intentID != "I1" {
		t.Fatalf("add intent: %v %q", err, intentID)
	}
	intent.IntentID = "I1"
	if err := c.UpdateKnowledgeIntent(ctx, intent); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteKnowledgeIntent(ctx, "I1"); err != nil {
		t.Fatal(err)
	}
	intents, err := c.ListKnowledgeIntents(ctx, "", 0, "G1", "")
	if err != nil || len(intents.Intents) != 1 || intents.Intents[0].Question.Text.Content != "问" {
		t.Fatalf("list intent: %+v %v", intents, err)
	}
	if _, err := c.AddKnowledgeIntent(ctx, KnowledgeIntent{GroupID: "G1"}); err == nil {
		t.Fatal("missing question must fail")
	}
	if (*calls)[3].Body["limit"] != float64(500) {
		t.Fatalf("list group limit default: %+v", (*calls)[3].Body)
	}
}

// TestSyncMsgPage covers the documented paging contract: has_more is a uint32
// flag, a page can be empty and still have more, and the limit defaults to 1000.
func TestSyncMsgPage(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		_, _ = w.Write([]byte(`{"errcode":0,"next_cursor":"c2","has_more":1,"msg_list":[` +
			`{"msgid":"m1","open_kfid":"KF1","external_userid":"wm1","send_time":1615478585,` +
			`"origin":3,"msgtype":"text","text":{"content":"hi","menu_id":"101"}},` +
			`{"msgid":"m2","origin":5,"servicer_userid":"zhangsan","msgtype":"image",` +
			`"image":{"media_id":"MEDIA1"}},` +
			`{"msgid":"m3","origin":4,"msgtype":"event","event":{"event_type":"enter_session"}}]}`))
	})
	page, err := c.SyncMsg(context.Background(), SyncMsgRequest{OpenKfID: "KF1", Limit: 5000})
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasMore || page.NextCursor != "c2" || len(page.Messages) != 3 {
		t.Fatalf("page: %+v", page)
	}
	if (*calls)[0].Body["limit"] != float64(1000) {
		t.Fatalf("limit not clamped to the documented maximum: %v", (*calls)[0].Body)
	}
	text := page.Messages[0]
	if text.MsgType != "text" || text.Text == nil || text.Text.Content != "hi" ||
		text.Text.MenuID != "101" || text.SendTime != 1615478585 || text.Origin != OriginCustomer {
		t.Fatalf("text message: %+v", text)
	}
	if page.Messages[1].Image == nil || page.Messages[1].Image.MediaID != "MEDIA1" ||
		page.Messages[1].ServicerUserID != "zhangsan" {
		t.Fatalf("image message: %+v", page.Messages[1])
	}
	if page.Messages[2].Event == nil || len(page.Messages[2].Raw) == 0 {
		t.Fatalf("event payload: %+v", page.Messages[2])
	}
}

func TestSyncMsgDefaultsAndValidation(t *testing.T) {
	c, calls := newProbe(t, func(w http.ResponseWriter, p probe) {
		_, _ = w.Write([]byte(`{"errcode":0,"next_cursor":"c2","has_more":0,"msg_list":[]}`))
	})
	ctx := context.Background()

	page, err := c.SyncMsg(ctx, SyncMsgRequest{OpenKfID: "KF1"})
	if err != nil || page.HasMore || len(page.Messages) != 0 {
		t.Fatalf("empty page: %+v %v", page, err)
	}
	body := (*calls)[0].Body
	if body["limit"] != float64(1000) || body["voice_format"] != float64(0) {
		t.Fatalf("defaults not applied: %v", body)
	}
	if _, ok := body["token"]; ok {
		t.Fatalf("empty token must be omitted: %v", body)
	}

	if _, err := c.SyncMsg(ctx, SyncMsgRequest{OpenKfID: "KF1", VoiceFormat: VoiceFormatSilk}); err != nil {
		t.Fatal(err)
	}
	if (*calls)[1].Body["voice_format"] != float64(1) {
		t.Fatalf("voice format not sent: %v", (*calls)[1].Body)
	}
	if _, err := c.SyncMsg(ctx, SyncMsgRequest{}); err == nil {
		t.Fatal("sync without an account accepted")
	}
	if _, err := c.SyncMsg(ctx, SyncMsgRequest{OpenKfID: "KF1", VoiceFormat: 2}); err == nil {
		t.Fatal("unknown voice format accepted")
	}
}
