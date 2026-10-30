package agent

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

func newAgent(t *testing.T, response string) (*recorder, *Agent) {
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

func TestGetAndListAgents(t *testing.T) {
	rec, api := newAgent(t, `{"errcode":0,"agentid":5,"name":"助手","square_logo_url":"https://x"}`)
	info, err := api.GetAgent(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/agent/get" || !strings.Contains(rec.query, "agentid=5") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if info.Name != "助手" || info.AgentID != 5 {
		t.Fatalf("unexpected agent: %+v", info)
	}

	rec, api = newAgent(t, `{"errcode":0,"agentlist":[{"agentid":1,"name":"A"}]}`)
	list, err := api.ListAgents(context.Background())
	if err != nil || rec.path != "/cgi-bin/agent/list" || len(list) != 1 {
		t.Fatalf("list agents: %v %s %+v", err, rec.path, list)
	}
}

func TestSetAgent(t *testing.T) {
	rec, api := newAgent(t, `{"errcode":0}`)
	info := AgentInfo{AgentID: 5, Name: "新名称", HomeURL: "https://home"}
	if err := api.SetAgent(context.Background(), info); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/agent/set" || rec.body["name"] != "新名称" {
		t.Fatalf("unexpected request %s %v", rec.path, rec.body)
	}
	if _, ok := rec.body["close"]; ok {
		t.Fatal("unset field must be omitted")
	}
	if err := api.SetAgent(context.Background(), AgentInfo{Name: "x"}); err == nil {
		t.Fatal("missing agentid accepted")
	}
}

func TestMenuRoundTrip(t *testing.T) {
	rec, api := newAgent(t, `{"errcode":0,"button":[{"name":"菜单","sub_button":[{"type":"click","name":"点我","key":"k"}]}]}`)
	ctx := context.Background()

	menu := Menu{Button: []MenuButton{{
		Name:      "菜单",
		SubButton: []MenuButton{{Type: "click", Name: "点我", Key: "k"}},
	}}}
	if err := api.CreateMenu(ctx, 5, menu); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/menu/create" || !strings.Contains(rec.query, "agentid=5") {
		t.Fatalf("unexpected create %s?%s", rec.path, rec.query)
	}

	got, err := api.GetMenu(ctx, 5)
	if err != nil || rec.path != "/cgi-bin/menu/get" {
		t.Fatalf("get menu: %v %s", err, rec.path)
	}
	if len(got.Button) != 1 || got.Button[0].SubButton[0].Key != "k" {
		t.Fatalf("unexpected menu: %+v", got)
	}

	if err := api.DeleteMenu(ctx, 5); err != nil || rec.path != "/cgi-bin/menu/delete" {
		t.Fatalf("delete menu: %v %s", err, rec.path)
	}
}

func TestMenuValidation(t *testing.T) {
	_, api := newAgent(t, `{"errcode":0}`)
	ctx := context.Background()
	cases := map[string]struct {
		agentID int
		menu    Menu
	}{
		"no agent":       {0, Menu{Button: []MenuButton{{Type: "click", Name: "a"}}}},
		"no buttons":     {5, Menu{}},
		"too many":       {5, Menu{Button: make([]MenuButton, 4)}},
		"no name":        {5, Menu{Button: []MenuButton{{Type: "click"}}}},
		"bad type":       {5, Menu{Button: []MenuButton{{Type: "unknown", Name: "a"}}}},
		"view no url":    {5, Menu{Button: []MenuButton{{Type: "view", Name: "a"}}}},
		"mini no fields": {5, Menu{Button: []MenuButton{{Type: "miniprogram", Name: "a"}}}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if err := api.CreateMenu(ctx, tc.agentID, tc.menu); err == nil {
				t.Fatal("invalid menu accepted")
			}
		})
	}
	if err := api.CreateMenu(ctx, 5, Menu{Button: []MenuButton{{Type: "view", Name: "a", URL: "https://x"}}}); err != nil {
		t.Fatalf("valid view button rejected: %v", err)
	}
}

func TestWorkbenchData(t *testing.T) {
	rec, api := newAgent(t, `{"errcode":0,"data":[{"type":"webview"}]}`)
	ctx := context.Background()

	payload := json.RawMessage(`[{"type":"webview","url":"https://x"}]`)
	if err := api.SetWorkbenchData(ctx, 5, payload); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/agent/batch_set_workbench_data" || rec.body["agentid"] != float64(5) {
		t.Fatalf("unexpected set %s %v", rec.path, rec.body)
	}
	data, err := api.GetWorkbenchData(ctx, 5)
	if err != nil || rec.path != "/cgi-bin/agent/get_workbench_data" || len(data) == 0 {
		t.Fatalf("get workbench: %v %s %s", err, rec.path, data)
	}
	if err := api.SetWorkbenchData(ctx, 5, json.RawMessage(`{`)); err == nil {
		t.Fatal("invalid JSON accepted")
	}
	if _, err := api.GetWorkbenchData(ctx, 0); err == nil {
		t.Fatal("missing agentid accepted")
	}
}
