package wedoc

import (
	"context"
	"encoding/json"
	"errors"
	wxclient "github.com/gowsp/go-wecom/client"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// docAPICall records one wedoc request as received by the fake server.
type docAPICall struct {
	Path  string
	Token string
	Body  map[string]any
}

// fakeDocAPI is an httptest replacement for the WeCom origin that serves
// gettoken and delegates every other path to a scripted handler.
type fakeDocAPI struct {
	server *httptest.Server

	mu         sync.Mutex
	tokenCalls int
	tokens     []string
	calls      []docAPICall

	handler func(index int, call docAPICall) map[string]any
}

func newFakeDocAPI(t *testing.T, handler func(index int, call docAPICall) map[string]any) *fakeDocAPI {
	t.Helper()
	fake := &fakeDocAPI{tokens: []string{"access-token-1"}, handler: handler}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.handle))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *fakeDocAPI) handle(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/cgi-bin/gettoken" {
		f.mu.Lock()
		f.tokenCalls++
		call := f.tokenCalls
		token := f.tokens[len(f.tokens)-1]
		if call-1 < len(f.tokens) {
			token = f.tokens[call-1]
		}
		f.mu.Unlock()
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errcode": 0, "errmsg": "ok", "access_token": token, "expires_in": 7200,
		})
		return
	}

	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	call := docAPICall{Path: r.URL.Path, Token: r.URL.Query().Get("access_token"), Body: body}

	f.mu.Lock()
	f.calls = append(f.calls, call)
	index := len(f.calls)
	handler := f.handler
	f.mu.Unlock()

	response := map[string]any{"errcode": 0, "errmsg": "ok"}
	if handler != nil {
		response = handler(index, call)
	}
	_ = json.NewEncoder(w).Encode(response)
}

func (f *fakeDocAPI) counts() (tokenCalls, endpointCalls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tokenCalls, len(f.calls)
}

func (f *fakeDocAPI) recorded() []docAPICall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]docAPICall(nil), f.calls...)
}

func objectField(t *testing.T, value any, key string) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s is %T, want object", key, value)
	}
	return object
}

func arrayField(t *testing.T, value any, key string) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("%s is %T, want array", key, value)
	}
	return array
}

func TestAddSheetRequest(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"properties": map[string]any{"sheet_id": "sheet-1", "title": "线索"},
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	sheet, err := client.AddSheet(context.Background(), " doc-1 ", " 线索 ")
	if err != nil {
		t.Fatal(err)
	}
	if sheet.SheetID != "sheet-1" || sheet.Title != "线索" {
		t.Fatalf("unexpected sheet: %+v", sheet)
	}

	calls := fake.recorded()
	if len(calls) != 1 {
		t.Fatalf("recorded %d calls, want 1", len(calls))
	}
	if calls[0].Path != "/cgi-bin/wedoc/smartsheet/add_sheet" {
		t.Fatalf("path = %q", calls[0].Path)
	}
	if calls[0].Token != "access-token-1" {
		t.Fatalf("token = %q", calls[0].Token)
	}
	if got := calls[0].Body["docid"]; got != "doc-1" {
		t.Fatalf("docid = %v", got)
	}
	properties := objectField(t, calls[0].Body["properties"], "properties")
	if properties["title"] != "线索" {
		t.Fatalf("properties.title = %v", properties["title"])
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
}

func TestGetSheetsListsSheetList(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"sheet_list": []any{
				map[string]any{"sheet_id": "s1", "title": "表1"},
				map[string]any{"sheet_id": "s2", "title": "表2"},
			},
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	sheets, err := client.GetSheets(context.Background(), " doc-1 ")
	if err != nil {
		t.Fatal(err)
	}
	if len(sheets) != 2 || sheets[0].SheetID != "s1" || sheets[1].Title != "表2" {
		t.Fatalf("unexpected sheets: %+v", sheets)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/smartsheet/get_sheet" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" {
		t.Fatalf("docid = %v", calls[0].Body["docid"])
	}
	if _, ok := calls[0].Body["sheet_id"]; ok {
		t.Fatalf("get_sheet must list by docid only: %+v", calls[0].Body)
	}
}

func TestAddFieldRequestAndPropertyKey(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		fields := arrayField(t, call.Body["fields"], "fields")
		created := make([]any, 0, len(fields))
		for _, raw := range fields {
			field := objectField(t, raw, "field")
			created = append(created, map[string]any{
				"field_id":    "field-" + field["field_title"].(string),
				"field_title": field["field_title"],
				"field_type":  field["field_type"],
			})
		}
		return map[string]any{"errcode": 0, "errmsg": "ok", "fields": created}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	field, err := client.AddField(context.Background(), "doc-1", "sheet-1", Field{
		FieldTitle: "标题",
		FieldType:  FieldTypeText,
	})
	if err != nil {
		t.Fatal(err)
	}
	if field.FieldID != "field-标题" || field.FieldTitle != "标题" || field.FieldType != FieldTypeText {
		t.Fatalf("unexpected field: %+v", field)
	}

	number := Field{
		FieldTitle: "金额",
		FieldType:  FieldTypeNumber,
		Property:   map[string]any{"precision": 2},
	}
	if _, err := client.AddField(context.Background(), "doc-1", "sheet-1", number); err != nil {
		t.Fatal(err)
	}

	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	if calls[0].Path != "/cgi-bin/wedoc/smartsheet/add_fields" || calls[0].Body["sheet_id"] != "sheet-1" {
		t.Fatalf("unexpected first call: %+v", calls[0])
	}
	textFields := arrayField(t, calls[0].Body["fields"], "fields")
	textField := objectField(t, textFields[0], "text field")
	if textField["field_title"] != "标题" || textField["field_type"] != FieldTypeText {
		t.Fatalf("unexpected text field: %+v", textField)
	}
	for key := range textField {
		if strings.HasPrefix(key, "property") {
			t.Fatalf("text field unexpectedly carried %s", key)
		}
	}

	numberFields := arrayField(t, calls[1].Body["fields"], "fields")
	numberField := objectField(t, numberFields[0], "number field")
	if numberField["field_title"] != "金额" || numberField["field_type"] != FieldTypeNumber {
		t.Fatalf("unexpected number field: %+v", numberField)
	}
	property := objectField(t, numberField["property_number"], "property_number")
	if property["precision"] != float64(2) {
		t.Fatalf("property_number = %+v", property)
	}
}

func TestGetFieldsRequest(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"fields": []any{
				map[string]any{"field_id": "f1", "field_title": "标题", "field_type": FieldTypeText},
				map[string]any{"field_id": "f2", "field_title": "金额", "field_type": FieldTypeNumber},
			},
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	fields, err := client.GetFields(context.Background(), "doc-1", "sheet-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 || fields[0].FieldID != "f1" || fields[1].FieldType != FieldTypeNumber {
		t.Fatalf("unexpected fields: %+v", fields)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/smartsheet/get_fields" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" || calls[0].Body["sheet_id"] != "sheet-1" {
		t.Fatalf("unexpected body: %+v", calls[0].Body)
	}
}

func TestAddRecordsRequest(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		records := arrayField(t, call.Body["records"], "records")
		values := objectField(t, objectField(t, records[0], "record")["values"], "values")
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"records": []any{
				map[string]any{"record_id": "rec-1", "values": values},
			},
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	records, err := client.AddRecords(context.Background(), "doc-1", "sheet-1",
		TextRecord(map[string]string{"标题": "第一条"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].RecordID != "rec-1" {
		t.Fatalf("unexpected records: %+v", records)
	}

	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/smartsheet/add_records" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" || calls[0].Body["sheet_id"] != "sheet-1" {
		t.Fatalf("unexpected target: %+v", calls[0].Body)
	}
	if calls[0].Body["key_type"] != "CELL_VALUE_KEY_TYPE_FIELD_TITLE" {
		t.Fatalf("key_type = %v", calls[0].Body["key_type"])
	}
	sent := arrayField(t, calls[0].Body["records"], "records")
	values := objectField(t, objectField(t, sent[0], "record")["values"], "values")
	cells := arrayField(t, values["标题"], "标题")
	cell := objectField(t, cells[0], "cell")
	if cell["type"] != "text" || cell["text"] != "第一条" {
		t.Fatalf("unexpected cell: %+v", cell)
	}
}

func TestUpdateRecordsRequest(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{"errcode": 0, "errmsg": "ok", "records": call.Body["records"]}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	updated, err := client.UpdateRecords(context.Background(), "doc-1", "sheet-1", Record{
		RecordID: "rec-1",
		Values:   map[string]any{"标题": TextCell("改后")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated) != 1 || updated[0].RecordID != "rec-1" {
		t.Fatalf("unexpected records: %+v", updated)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/smartsheet/update_records" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["key_type"] != "CELL_VALUE_KEY_TYPE_FIELD_TITLE" {
		t.Fatalf("key_type = %v", calls[0].Body["key_type"])
	}
	records := arrayField(t, calls[0].Body["records"], "records")
	record := objectField(t, records[0], "record")
	if record["record_id"] != "rec-1" {
		t.Fatalf("record_id = %v", record["record_id"])
	}
	values := objectField(t, record["values"], "values")
	cells := arrayField(t, values["标题"], "标题")
	if objectField(t, cells[0], "cell")["text"] != "改后" {
		t.Fatalf("unexpected values: %+v", values)
	}
}

func TestDeleteRecordsRequest(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	if err := client.DeleteRecords(context.Background(), "doc-1", "sheet-1", " rec-1 ", "rec-2"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/smartsheet/delete_records" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if _, ok := calls[0].Body["key_type"]; ok {
		t.Fatal("delete_records must not send key_type")
	}
	ids := arrayField(t, calls[0].Body["record_ids"], "record_ids")
	if len(ids) != 2 || ids[0] != "rec-1" || ids[1] != "rec-2" {
		t.Fatalf("record_ids = %v", ids)
	}
}

func TestGetRecordsPaginatesUntilDone(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		page := map[string]any{"errcode": 0, "errmsg": "ok", "total": 3}
		switch call.Body["offset"] {
		case float64(0):
			page["records"] = []any{recordJSON("rec-1")}
			page["has_more"] = true
			page["next"] = 1
		case float64(1):
			page["records"] = []any{recordJSON("rec-2")}
			page["has_more"] = true
			page["next"] = 2
		case float64(2):
			page["records"] = []any{recordJSON("rec-3")}
			page["has_more"] = false
			page["next"] = 0
		default:
			page["errcode"] = 40013
			page["errmsg"] = "unexpected page"
		}
		return page
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	records, err := client.QueryRecords(context.Background(), "doc-1", "sheet-1", RecordQuery{
		ViewID:            "view-1",
		RecordIDs:         []string{"rec-1", "rec-2", "rec-3"},
		KeyType:           CellValueKeyTypeFieldTitle,
		FieldTitles:       []string{"标题"},
		Sort:              []Sort{{FieldTitle: "标题", Desc: true}},
		Limit:             5,
		Ver:               160,
		NeedFormulaResult: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 || records[0].RecordID != "rec-1" || records[2].RecordID != "rec-3" {
		t.Fatalf("unexpected records: %+v", records)
	}

	calls := fake.recorded()
	if len(calls) != 3 {
		t.Fatalf("recorded %d calls, want 3", len(calls))
	}
	wantOffsets := []float64{0, 1, 2}
	for index, call := range calls {
		if call.Path != "/cgi-bin/wedoc/smartsheet/get_records" {
			t.Fatalf("call %d path = %q", index, call.Path)
		}
		if call.Body["docid"] != "doc-1" || call.Body["sheet_id"] != "sheet-1" {
			t.Fatalf("call %d target = %+v", index, call.Body)
		}
		if call.Body["view_id"] != "view-1" {
			t.Fatalf("call %d view_id = %v", index, call.Body["view_id"])
		}
		if call.Body["offset"] != wantOffsets[index] || call.Body["limit"] != float64(5) {
			t.Fatalf("call %d paging = %+v", index, call.Body)
		}
		if call.Body["key_type"] != CellValueKeyTypeFieldTitle {
			t.Fatalf("call %d key_type = %v", index, call.Body["key_type"])
		}
		if call.Body["ver"] != float64(160) || call.Body["need_formula_result"] != true {
			t.Fatalf("call %d query flags = %+v", index, call.Body)
		}
		ids := arrayField(t, call.Body["record_ids"], "record_ids")
		if len(ids) != 3 {
			t.Fatalf("call %d record_ids = %v", index, ids)
		}
		titles := arrayField(t, call.Body["field_titles"], "field_titles")
		if len(titles) != 1 || titles[0] != "标题" {
			t.Fatalf("call %d field_titles = %v", index, titles)
		}
		sorts := arrayField(t, call.Body["sort"], "sort")
		sort := objectField(t, sorts[0], "sort")
		if sort["field_title"] != "标题" || sort["desc"] != true {
			t.Fatalf("call %d sort = %+v", index, sort)
		}
	}
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
}

func TestGetRecordsFallsBackToTotal(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		switch call.Body["offset"] {
		case float64(0):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 3,
				"records":  []any{recordJSON("rec-1"), recordJSON("rec-2")},
				"has_more": false,
			}
		case float64(2):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 3,
				"records":  []any{recordJSON("rec-3")},
				"has_more": false,
			}
		default:
			return map[string]any{"errcode": 40013, "errmsg": "unexpected offset"}
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	records, err := client.GetRecords(context.Background(), "doc-1", "sheet-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 3 {
		t.Fatalf("unexpected records: %+v", records)
	}
	_, endpointCalls := fake.counts()
	if endpointCalls != 2 {
		t.Fatalf("endpoint called %d times, want 2", endpointCalls)
	}
}

func TestSmartSheetTokenInvalidRetriesOnce(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		if index == 1 {
			return map[string]any{"errcode": 42001, "errmsg": "access token expired"}
		}
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"properties": map[string]any{"sheet_id": "sheet-1", "title": "线索"},
		}
	})
	fake.tokens = []string{"token-1", "token-2"}
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	if _, err := client.AddSheet(context.Background(), "doc-1", "线索"); err != nil {
		t.Fatalf("retry after token refresh: %v", err)
	}
	tokenCalls, endpointCalls := fake.counts()
	if tokenCalls != 2 {
		t.Fatalf("gettoken called %d times, want 2", tokenCalls)
	}
	if endpointCalls != 2 {
		t.Fatalf("endpoint called %d times, want 2", endpointCalls)
	}
	calls := fake.recorded()
	if calls[0].Token != "token-1" || calls[1].Token != "token-2" {
		t.Fatalf("retry did not use the refreshed token: %v", []string{calls[0].Token, calls[1].Token})
	}
}

func TestSmartSheetErrorSummaryHidesSecretsAndContent(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 40013,
			"errmsg":  "invalid sheet corp-secret",
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	_, err := client.AddRecords(context.Background(), "doc-1", "sheet-1",
		TextRecord(map[string]string{"标题": "private-cell"}))
	if err == nil {
		t.Fatal("expected an API error")
	}
	var apiErr *wxclient.APIError
	if !errors.As(err, &apiErr) || apiErr.Errcode != 40013 {
		t.Fatalf("expected APIError 40013, got %v", err)
	}
	if !strings.Contains(err.Error(), "40013") || !strings.Contains(err.Error(), "invalid sheet") {
		t.Fatalf("error does not summarize errcode and errmsg: %v", err)
	}
	if !strings.Contains(err.Error(), "[redacted]") {
		t.Fatalf("error did not redact the corp secret: %v", err)
	}
	for _, forbidden := range []string{"corp-secret", "access-token-1", "private-cell"} {
		if strings.Contains(err.Error(), forbidden) {
			t.Fatalf("error leaks %q: %v", forbidden, err)
		}
	}
}

func TestSmartSheetRequiredArguments(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))
	ctx := context.Background()

	tests := []struct {
		name string
		call func() error
		want string
	}{
		{"add sheet without doc", func() error { _, err := client.AddSheet(ctx, " ", "线索"); return err }, "docId"},
		{"add sheet without title", func() error { _, err := client.AddSheet(ctx, "doc-1", " "); return err }, "title"},
		{"get sheets without doc", func() error { _, err := client.GetSheets(ctx, " "); return err }, "docId"},
		{"get fields without sheet", func() error { _, err := client.GetFields(ctx, "doc-1", ""); return err }, "sheetId"},
		{"add fields without fields", func() error {
			_, err := client.AddFields(ctx, "doc-1", "sheet-1")
			return err
		}, "at least one field"},
		{"add field without title", func() error {
			_, err := client.AddField(ctx, "doc-1", "sheet-1", Field{FieldType: FieldTypeText})
			return err
		}, "fieldTitle"},
		{"add field without type", func() error {
			_, err := client.AddField(ctx, "doc-1", "sheet-1", Field{FieldTitle: "标题"})
			return err
		}, "fieldType"},
		{"add records without records", func() error {
			_, err := client.AddRecords(ctx, "doc-1", "sheet-1")
			return err
		}, "at least one record"},
		{"update records without id", func() error {
			_, err := client.UpdateRecords(ctx, "doc-1", "sheet-1", Record{Values: map[string]any{}})
			return err
		}, "recordId"},
		{"delete records without ids", func() error {
			return client.DeleteRecords(ctx, "doc-1", "sheet-1")
		}, "at least one recordId"},
		{"get records negative max", func() error {
			_, err := client.QueryRecords(ctx, "doc-1", "sheet-1", RecordQuery{Limit: -1})
			return err
		}, "limit"},
		{"create doc without name", func() error {
			_, err := client.CreateDoc(ctx, DocTypeSmartsheet, " ")
			return err
		}, "docName"},
		{"create doc without type", func() error {
			_, err := client.CreateDoc(ctx, 0, "新文档")
			return err
		}, "docType"},
		{"create smartsheet without name", func() error {
			_, err := client.CreateSmartsheet(ctx, " ")
			return err
		}, "docName"},
		{"delete doc without id", func() error {
			return client.DeleteDoc(ctx, " ")
		}, "docId"},
		{"share doc without users", func() error {
			return client.ShareDocWithUsers(ctx, "doc-1", nil)
		}, "at least one user"},
		{"share doc without user id", func() error {
			return client.ShareDocWithUsers(ctx, "doc-1", []UserPerm{{Auth: DocAuthAdmin}})
		}, "userId"},
		{"share doc with invalid auth", func() error {
			return client.ShareDocWithUsers(ctx, "doc-1", []UserPerm{{UserID: "zhangsan", Auth: 3}})
		}, "auth must be 1, 2, or 7"},
		{"modify doc members without change", func() error {
			return client.ModifyDocMembers(ctx, "doc-1", nil, nil)
		}, "at least one member change"},
		{"remove doc members without ids", func() error {
			return client.RemoveDocMembers(ctx, "doc-1")
		}, "at least one userId"},
		{"rename doc without name", func() error {
			return client.RenameDoc(ctx, "doc-1", " ")
		}, "newName"},
		{"get doc base info without doc", func() error {
			_, err := client.GetDocBaseInfo(ctx, " ")
			return err
		}, "docId"},
		{"share doc without doc", func() error {
			_, err := client.ShareDoc(ctx, " ")
			return err
		}, "docId"},
		{"get doc auth without doc", func() error {
			_, err := client.GetDocAuth(ctx, " ")
			return err
		}, "docId"},
		{"set doc join rule without doc", func() error {
			return client.SetDocJoinRule(ctx, " ", DocJoinRule{})
		}, "docId"},
		{"create doc with only space", func() error {
			_, err := client.CreateDocWithOptions(ctx, DocTypeSmartsheet, "文档", CreateDocOptions{SpaceID: "space-1"})
			return err
		}, "spaceId and fatherId"},
		{"update sheet without sheet", func() error {
			return client.UpdateSheet(ctx, "doc-1", " ", "标题")
		}, "sheetId"},
		{"delete sheet without sheet", func() error {
			return client.DeleteSheet(ctx, "doc-1", " ")
		}, "sheetId"},
		{"add view without title", func() error {
			_, err := client.AddView(ctx, "doc-1", "sheet-1", View{ViewType: ViewTypeGrid})
			return err
		}, "viewTitle"},
		{"add view without type", func() error {
			_, err := client.AddView(ctx, "doc-1", "sheet-1", View{ViewTitle: "视图"})
			return err
		}, "viewType"},
		{"update view without id", func() error {
			_, err := client.UpdateView(ctx, "doc-1", "sheet-1", View{})
			return err
		}, "viewId"},
		{"delete views without ids", func() error {
			return client.DeleteViews(ctx, "doc-1", "sheet-1")
		}, "at least one viewId"},
		{"update fields without id", func() error {
			_, err := client.UpdateFields(ctx, "doc-1", "sheet-1", Field{FieldType: FieldTypeText})
			return err
		}, "fieldId"},
		{"update fields without type", func() error {
			_, err := client.UpdateFields(ctx, "doc-1", "sheet-1", Field{FieldID: "f1"})
			return err
		}, "fieldType"},
		{"delete fields without ids", func() error {
			return client.DeleteFields(ctx, "doc-1", "sheet-1")
		}, "at least one fieldId"},
		{"add field group without name", func() error {
			_, err := client.AddFieldGroup(ctx, "doc-1", "sheet-1", " ", nil)
			return err
		}, "name"},
		{"update field group without id", func() error {
			_, err := client.UpdateFieldGroup(ctx, "doc-1", "sheet-1", " ", "组", nil)
			return err
		}, "fieldGroupId"},
		{"delete field groups without ids", func() error {
			return client.DeleteFieldGroups(ctx, "doc-1", "sheet-1")
		}, "at least one fieldGroupId"},
		{"get content privileges with bad type", func() error {
			_, err := client.GetContentPrivileges(ctx, "doc-1", 3, nil)
			return err
		}, "ruleType"},
		{"update content privilege without rule id", func() error {
			return client.UpdateContentPrivilege(ctx, "doc-1", ContentRule{Type: ContentRuleExtra})
		}, "ruleId"},
		{"update content privilege with bad priv", func() error {
			return client.UpdateContentPrivilege(ctx, "doc-1", ContentRule{
				Type: ContentRuleExtra, RuleID: 1,
				PrivList: []SheetPrivilege{{SheetID: "sheet-1", Priv: 9}},
			})
		}, "priv must be"},
		{"create content privilege rule without name", func() error {
			_, err := client.CreateContentPrivilegeRule(ctx, "doc-1", " ")
			return err
		}, "name"},
		{"modify content privilege members without rule", func() error {
			return client.ModifyContentPrivilegeMembers(ctx, "doc-1", 0, []string{"u1"}, nil)
		}, "ruleId"},
		{"delete content privilege rules without ids", func() error {
			return client.DeleteContentPrivilegeRules(ctx, "doc-1")
		}, "at least one ruleId"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.call()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected error containing %q, got %v", test.want, err)
			}
		})
	}
	if tokenCalls, endpointCalls := fake.counts(); tokenCalls != 0 || endpointCalls != 0 {
		t.Fatalf("invalid arguments reached the API: %d token, %d endpoint calls", tokenCalls, endpointCalls)
	}
}

func TestCreateSmartsheetRequest(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"docid": "doc-1", "url": "https://doc.example/doc-1",
		}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	docID, err := client.CreateSmartsheet(context.Background(), " 线索表 ")
	if err != nil {
		t.Fatal(err)
	}
	if docID != "doc-1" {
		t.Fatalf("docID = %q, want doc-1", docID)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/create_doc" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["doc_type"] != float64(DocTypeSmartsheet) || calls[0].Body["doc_name"] != "线索表" {
		t.Fatalf("unexpected body: %+v", calls[0].Body)
	}
	if _, ok := calls[0].Body["admin_users"]; ok {
		t.Fatalf("admin_users must stay optional: %+v", calls[0].Body)
	}
}

func TestCreateSmartsheetSharesOwners(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		if call.Path == "/cgi-bin/wedoc/create_doc" {
			return map[string]any{"errcode": 0, "errmsg": "ok", "docid": "doc-1"}
		}
		return map[string]any{"errcode": 0, "errmsg": "ok"}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	docID, err := client.CreateSmartsheet(context.Background(), "线索表")
	if err != nil {
		t.Fatal(err)
	}
	if docID != "doc-1" {
		t.Fatalf("docID = %q, want doc-1", docID)
	}
	if err := client.ShareDocWithUsers(context.Background(), docID, []UserPerm{{UserID: "admin-1", Auth: DocAuthAdmin}, {UserID: "admin-2", Auth: DocAuthAdmin}}); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	if calls[1].Path != "/cgi-bin/wedoc/mod_doc_member" {
		t.Fatalf("share path = %q", calls[1].Path)
	}
	if calls[1].Body["docid"] != "doc-1" {
		t.Fatalf("share docid = %v", calls[1].Body["docid"])
	}
	members := arrayField(t, calls[1].Body["update_file_member_list"], "update_file_member_list")
	if len(members) != 2 {
		t.Fatalf("members = %v", members)
	}
	first := objectField(t, members[0], "member")
	if first["type"] != float64(1) || first["auth"] != float64(DocAuthAdmin) || first["userid"] != "admin-1" {
		t.Fatalf("unexpected first member: %+v", first)
	}
	second := objectField(t, members[1], "member")
	if second["userid"] != "admin-2" {
		t.Fatalf("unexpected second member: %+v", second)
	}
	// Creating the document and sharing it must reuse one access token.
	if tokenCalls, _ := fake.counts(); tokenCalls != 1 {
		t.Fatalf("gettoken called %d times, want 1", tokenCalls)
	}
}

func TestCreateSmartsheetShareFailureReturnsDocID(t *testing.T) {
	fake := newFakeDocAPI(t, func(index int, call docAPICall) map[string]any {
		if call.Path == "/cgi-bin/wedoc/create_doc" {
			return map[string]any{"errcode": 0, "errmsg": "ok", "docid": "doc-1"}
		}
		return map[string]any{"errcode": 40013, "errmsg": "member denied"}
	})
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	docID, err := client.CreateSmartsheet(context.Background(), "线索表")
	if err != nil {
		t.Fatal(err)
	}
	if err = client.ShareDocWithUsers(context.Background(), docID, []UserPerm{{UserID: "admin-1", Auth: DocAuthAdmin}}); err == nil || !strings.Contains(err.Error(), "40013") {
		t.Fatalf("expected sharing to fail: %v", err)
	}
	if docID != "doc-1" {
		t.Fatalf("docID = %q, want the created document id", docID)
	}
}

func TestDeleteDocRequest(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	if err := client.DeleteDoc(context.Background(), " doc-1 "); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/del_doc" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" {
		t.Fatalf("docid = %v", calls[0].Body["docid"])
	}
}

func TestShareDocWithUsersRequest(t *testing.T) {
	fake := newFakeDocAPI(t, nil)
	client := New(wxclient.New("corp-id", "corp-secret", wxclient.WithBaseURL(fake.server.URL)))

	err := client.ShareDocWithUsers(context.Background(), "doc-1", []UserPerm{
		{UserID: " reader ", Auth: DocAuthRead},
		{UserID: "writer", Auth: DocAuthReadWrite},
		{UserID: "owner", Auth: DocAuthAdmin},
	})
	if err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 1 || calls[0].Path != "/cgi-bin/wedoc/mod_doc_member" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
	if calls[0].Body["docid"] != "doc-1" {
		t.Fatalf("docid = %v", calls[0].Body["docid"])
	}
	members := arrayField(t, calls[0].Body["update_file_member_list"], "update_file_member_list")
	if len(members) != 3 {
		t.Fatalf("members = %v", members)
	}
	wantAuth := []float64{1, 2, 7}
	wantUsers := []string{"reader", "writer", "owner"}
	for index, raw := range members {
		member := objectField(t, raw, "member")
		if member["type"] != float64(1) {
			t.Fatalf("member %d type = %v", index, member["type"])
		}
		if member["auth"] != wantAuth[index] || member["userid"] != wantUsers[index] {
			t.Fatalf("member %d = %+v", index, member)
		}
	}
}

func TestTextCellHelpers(t *testing.T) {
	if got := TextCell("hello"); !reflect.DeepEqual(got, []TextValue{{Type: "text", Text: "hello"}}) {
		t.Fatalf("TextCell = %+v", got)
	}
	if got := LinkCell("Soma", "https://soma.example"); !reflect.DeepEqual(got,
		[]TextValue{{Type: "url", Text: "Soma", Link: "https://soma.example"}}) {
		t.Fatalf("LinkCell = %+v", got)
	}
	record := TextRecord(map[string]string{"标题": "hello", "备注": "world"})
	if record.RecordID != "" || len(record.Values) != 2 {
		t.Fatalf("unexpected record: %+v", record)
	}
	encoded, err := json.Marshal(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"标题":[{"type":"text","text":"hello"}]`) {
		t.Fatalf("unexpected record JSON: %s", encoded)
	}
}

func recordJSON(recordID string) map[string]any {
	return map[string]any{
		"record_id": recordID,
		"values": map[string]any{
			"标题": []any{map[string]any{"type": "text", "text": recordID}},
		},
	}
}
