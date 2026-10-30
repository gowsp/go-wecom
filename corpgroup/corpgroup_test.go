package corpgroup_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/corpgroup"
)

type recorder struct {
	path string
	body map[string]any
}

func newAPI(t *testing.T, reply string) (*recorder, *client.Client) {
	t.Helper()
	rec := &recorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec.path = r.URL.Path
		if strings.HasSuffix(r.URL.Path, "/gettoken") {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"tk","expires_in":7200}`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &rec.body)
		}
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(server.Close)
	api := client.New("corp", "secret", client.WithBaseURL(server.URL))
	return rec, api
}

func TestSharedCorps(t *testing.T) {
	rec, api := newAPI(t, `{"errcode":0,"ending":1,`+
		`"corp_list":[{"corpid":"c1","corp_name":"下游一","agentid":7}],`+
		`"access_token":"tk","expires_in":7200,"userid":"u1","session_key":"sk"}`)
	svc := corpgroup.New(api)
	ctx := context.Background()

	page, err := svc.SharedCorps(ctx, 7, corpgroup.BusinessChain, "", "", 100)
	if err != nil || len(page.Corps) != 1 || page.Ending != 1 ||
		rec.path != "/cgi-bin/corpgroup/corp/list_app_share_info" {
		t.Fatalf("shared: %v %+v %s", err, page, rec.path)
	}

	token, err := svc.DownstreamToken(ctx, "c1", corpgroup.BusinessChain, 7)
	if err != nil || token.AccessToken != "tk" || rec.path != "/cgi-bin/corpgroup/corp/gettoken" {
		t.Fatalf("token: %v %+v %s", err, token, rec.path)
	}

	userID, sessionKey, err := svc.MiniProgramSession(ctx, "code", "session")
	if err != nil || userID != "u1" || sessionKey != "sk" ||
		rec.path != "/cgi-bin/miniprogram/transfer_session" {
		t.Fatalf("session: %q %q %v %s", userID, sessionKey, err, rec.path)
	}

	if _, err := svc.SharedCorps(ctx, 0, 0, "", "", 0); err == nil {
		t.Fatal("shared without agent accepted")
	}
	if _, err := svc.DownstreamToken(ctx, "", 0, 0); err == nil {
		t.Fatal("token without corpid accepted")
	}
	if _, _, err := svc.MiniProgramSession(ctx, "", ""); err == nil {
		t.Fatal("session without input accepted")
	}
}

func TestChains(t *testing.T) {
	rec, api := newAPI(t, `{"errcode":0,"has_more":true,"next_cursor":"c2",`+
		`"chains":[{"chain_id":"ch1","chain_name":"供应链"}],`+
		`"groups":[{"groupid":2,"chain_name":"一级"}],`+
		`"group_corps":[{"groupid":2,"corpid":"c1","corp_name":"下游"}],`+
		`"user_custom_id":"100","external_userid_info":[{"corpid":"c1","external_userid":"e1"}],`+
		`"invalid_unionid_list":["u9"],"pending_id":"p1",`+
		`"pending_id_list":[{"pending_id":"p1"}],"jobid":"job1","status":3}`)
	svc := corpgroup.New(api)
	ctx := context.Background()

	chains, err := svc.Chains(ctx)
	if err != nil || len(chains) != 1 ||
		rec.path != "/cgi-bin/corpgroup/corp/get_chain_corpinfo_list" {
		t.Fatalf("chains: %v %+v %s", err, chains, rec.path)
	}
	if _, err := svc.ChainGroups(ctx, "ch1", 0); err != nil ||
		rec.path != "/cgi-bin/corpgroup/corp/get_chain_group" {
		t.Fatalf("groups: %v %s", err, rec.path)
	}
	corps, err := svc.ChainCorps(ctx, "ch1", 2, false, "", 0)
	if err != nil || len(corps.Corps) != 1 || !corps.HasMore ||
		rec.path != "/cgi-bin/corpgroup/corp/get_chain_corpinfo" {
		t.Fatalf("corps: %v %+v %s", err, corps, rec.path)
	}
	info, invalid, err := svc.UnionIDToExternalUserID(ctx, "u1", "", "", "")
	if err != nil || len(info) != 1 || len(invalid) != 1 ||
		rec.path != "/cgi-bin/corpgroup/unionid_to_external_userid" {
		t.Fatalf("unionid: %v %+v %v %s", err, info, invalid, rec.path)
	}
	if _, err := svc.PendingID(ctx, "u1", ""); err != nil ||
		rec.path != "/cgi-bin/corpgroup/unionid_to_pending_id" {
		t.Fatalf("pending id: %v %s", err, rec.path)
	}
	ids, err := svc.PendingIDs(ctx, []string{"u1"}, []string{"o1"})
	if err != nil || len(ids) != 1 ||
		rec.path != "/cgi-bin/corpgroup/batch/external_userid_to_pending_id" {
		t.Fatalf("pending ids: %v %+v %s", err, ids, rec.path)
	}
	if id, err := svc.ImportChainContacts(ctx, "ch1", []corpgroup.ContactCorp{{CorpName: "下游"}}); err != nil ||
		id != "job1" || rec.path != "/cgi-bin/corpgroup/import_chain_contact" {
		t.Fatalf("import: %q %v %s", id, err, rec.path)
	}
	if _, err := svc.ImportResult(ctx, "job1"); err != nil ||
		rec.path != "/cgi-bin/corpgroup/getresult" {
		t.Fatalf("result: %v %s", err, rec.path)
	}
	if err := svc.RemoveCorp(ctx, "ch1", "c1", ""); err != nil ||
		rec.path != "/cgi-bin/corpgroup/corp/remove_corp" {
		t.Fatalf("remove: %v %s", err, rec.path)
	}
	custom, err := svc.ChainUserCustomID(ctx, "ch1", "c1", "u1")
	if err != nil || custom != "100" ||
		rec.path != "/cgi-bin/corpgroup/corp/get_chain_user_custom_id" {
		t.Fatalf("custom id: %q %v %s", custom, err, rec.path)
	}
	if _, err := svc.SharedChains(ctx, "c1"); err != nil ||
		rec.path != "/cgi-bin/corpgroup/get_corp_shared_chain_list" {
		t.Fatalf("shared chains: %v %s", err, rec.path)
	}

	if _, err := svc.PendingIDs(ctx, nil, nil); err == nil {
		t.Fatal("empty pending list accepted")
	}
	if err := svc.RemoveCorp(ctx, "ch1", "", ""); err == nil {
		t.Fatal("remove without corp accepted")
	}
	if _, err := svc.ChainUserCustomID(ctx, "", "", ""); err == nil {
		t.Fatal("custom without ids accepted")
	}
	if _, err := svc.SharedChains(ctx, ""); err == nil {
		t.Fatal("shared chains without corpid accepted")
	}
}

func TestRules(t *testing.T) {
	rec, api := newAPI(t, `{"errcode":0,"rule_ids":[1,2],"rule_id":3,`+
		`"rule_info":{"owner_corp_range":{"userids":["u1"]},`+
		`"member_corp_range":{"corpids":["c1"]}}}`)
	svc := corpgroup.New(api)
	ctx := context.Background()

	ids, err := svc.RuleIDs(ctx, "ch1")
	if err != nil || len(ids) != 2 || rec.path != "/cgi-bin/corpgroup/rule/list_ids" {
		t.Fatalf("ids: %v %v %s", err, ids, rec.path)
	}
	rule, err := svc.Rule(ctx, "ch1", 1)
	if err != nil || len(rule.OwnerRange.UserIDs) != 1 ||
		rec.path != "/cgi-bin/corpgroup/rule/get_rule_info" {
		t.Fatalf("rule: %v %+v %s", err, rule, rec.path)
	}
	if id, err := svc.AddRule(ctx, "ch1", rule); err != nil || id != 3 ||
		rec.path != "/cgi-bin/corpgroup/rule/add_rule" {
		t.Fatalf("add: %d %v %s", id, err, rec.path)
	}
	if err := svc.UpdateRule(ctx, "ch1", 1, rule); err != nil ||
		rec.path != "/cgi-bin/corpgroup/rule/modify_rule" {
		t.Fatalf("update: %v %s", err, rec.path)
	}
	if err := svc.DeleteRule(ctx, "ch1", 1); err != nil ||
		rec.path != "/cgi-bin/corpgroup/rule/delete_rule" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	if _, err := svc.RuleIDs(ctx, ""); err == nil {
		t.Fatal("ids without chain accepted")
	}
	if _, err := svc.Rule(ctx, "ch1", 0); err == nil {
		t.Fatal("rule without id accepted")
	}
	if err := svc.UpdateRule(ctx, "ch1", 0, rule); err == nil {
		t.Fatal("update without rule id accepted")
	}
}

func TestChainEvents(t *testing.T) {
	payload := `<xml><CreateTime>1</CreateTime><MsgType>event</MsgType>` +
		`<Event><![CDATA[change_chain]]></Event>` +
		`<ChangeType><![CDATA[create_chain]]></ChangeType>` +
		`<ChainId><![CDATA[ch1]]></ChainId></xml>`
	event, err := corpgroup.ParseChainEvent([]byte(payload))
	if err != nil || event.ChainID != "ch1" || event.ChangeType != corpgroup.ChangeCreateChain {
		t.Fatalf("chain event: %v %+v", err, event)
	}
	if _, err := corpgroup.ParseChainEvent([]byte(
		`<xml><Event>change_chain</Event><ChangeType>other</ChangeType><ChainId>ch1</ChainId></xml>`)); err == nil {
		t.Fatal("unexpected change type accepted")
	}

	job := `<xml><MsgType>event</MsgType><Event><![CDATA[batch_job_result]]></Event>` +
		`<BatchJob><JobId><![CDATA[job1]]></JobId><JobType><![CDATA[import_chain_contact]]></JobType>` +
		`<ErrCode>0</ErrCode><ErrMsg><![CDATA[ok]]></ErrMsg></BatchJob></xml>`
	jobEvent, err := corpgroup.ParseBatchJobEvent([]byte(job))
	if err != nil || jobEvent.JobID != "job1" || jobEvent.JobType != "import_chain_contact" {
		t.Fatalf("job event: %v %+v", err, jobEvent)
	}
	if _, err := corpgroup.ParseBatchJobEvent([]byte(`<xml><Event>other</Event></xml>`)); err == nil {
		t.Fatal("unexpected job event accepted")
	}
}
