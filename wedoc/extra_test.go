package wedoc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gowsp/go-wecom/client"
)

func newDocProbe(t *testing.T, reply func(w http.ResponseWriter, path string, body map[string]any)) (*WeDoc, *[]string) {
	t.Helper()
	var paths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/cgi-bin/gettoken" {
			_, _ = w.Write([]byte(`{"errcode":0,"access_token":"t","expires_in":7200}`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(raw, &body)
		paths = append(paths, r.URL.Path)
		reply(w, r.URL.Path, body)
	}))
	t.Cleanup(server.Close)
	return New(client.New("corp", "secret", client.WithBaseURL(server.URL))), &paths
}

func TestDocumentAndSpreadsheet(t *testing.T) {
	c, paths := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		switch path {
		case documentGetPath:
			_, _ = w.Write([]byte(`{"errcode":0,"version":3,"document":{"blocks":[]}}`))
		case spreadsheetPropertiesPath:
			_, _ = w.Write([]byte(`{"errcode":0,"properties":[{"sheet_id":"s1"}]}`))
		case spreadsheetRangePath:
			if body["range"] != "A1:B2" {
				t.Errorf("unexpected range: %+v", body)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"data":{"result":{"values":[]}}}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0,"data":{"responses":[]}}`))
		}
	})
	ctx := context.Background()

	doc, err := c.GetDocument(ctx, "d1")
	if err != nil || doc.Version != 3 || len(doc.Document) == 0 {
		t.Fatalf("get document: %+v %v", doc, err)
	}
	if err := c.BatchUpdateDocument(ctx, "d1", 3, []map[string]any{{"insert_text": map[string]any{"text": "x"}}}); err != nil {
		t.Fatal(err)
	}
	if err := c.BatchUpdateDocument(ctx, "d1", 0, nil); err == nil {
		t.Fatal("empty requests must fail")
	}
	if _, err := c.SpreadsheetProperties(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SpreadsheetRange(ctx, "d1", "s1", "A1:B2"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.SpreadsheetRange(ctx, "d1", "", "A1:B2"); err == nil {
		t.Fatal("missing sheet must fail")
	}
	if len(*paths) != 4 {
		t.Fatalf("unexpected calls: %v", *paths)
	}
}

func TestSmartdocPagesAndBlocks(t *testing.T) {
	c, _ := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		switch path {
		case smartdocPageHierarchyPath:
			_, _ = w.Write([]byte(`{"errcode":0,"pages":[{"page_id":"p1","title":"首页"}]}`))
		case smartdocAddPagePath, smartdocUpdatePagePath:
			_, _ = w.Write([]byte(`{"errcode":0,"info":{"page_id":"p1","title":"新"}}`))
		case smartdocAddBlocksPath, smartdocUpdateBlocksPath:
			_, _ = w.Write([]byte(`{"errcode":0,"blocks":[{"id":"b1","type":"BLOCK_TYPE_TEXT"}]}`))
		case smartdocBlockListPath:
			if body["limit"] != float64(200) {
				t.Errorf("unexpected limit: %+v", body)
			}
			_, _ = w.Write([]byte(`{"errcode":0,"blocks":[{"id":"b1"}],"has_more":"true","next_start":201}`))
		case smartdocPublishPath:
			_, _ = w.Write([]byte(`{"errcode":0,"share_code":"S","publish_url":"https://doc","version":1}`))
		case smartdocExportTaskPath:
			_, _ = w.Write([]byte(`{"errcode":0,"task_id":"T1"}`))
		case smartdocDatasourcePath:
			_, _ = w.Write([]byte(`{"errcode":0,"ss_docid":"ss1"}`))
		case smartdocAddDatasheetPath, smartdocUpdDatasheetPath:
			_, _ = w.Write([]byte(`{"errcode":0,"info":{"block_id":"b1","title":"表"}}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	})
	ctx := context.Background()

	pages, err := c.PageHierarchy(ctx, "d1")
	if err != nil || len(pages) != 1 || pages[0].Title != "首页" {
		t.Fatalf("pages: %+v %v", pages, err)
	}
	if page, err := c.AddPage(ctx, "d1", Page{Title: "x"}); err != nil || page.PageID != "p1" {
		t.Fatalf("add page: %+v %v", page, err)
	}
	if page, err := c.UpdatePage(ctx, "d1", Page{PageID: "p1", Title: "新"}); err != nil || page.Title != "新" {
		t.Fatalf("update page: %+v %v", page, err)
	}
	if err := c.DeletePage(ctx, "d1", "p1"); err != nil {
		t.Fatal(err)
	}
	if blocks, err := c.AddBlocks(ctx, "d1", "p1", []Block{{Type: "BLOCK_TYPE_TEXT"}}); err != nil || len(blocks) != 1 {
		t.Fatalf("add blocks: %+v %v", blocks, err)
	}
	if _, err := c.UpdateBlocks(ctx, "d1", "p1", []Block{{ID: "b1"}}); err != nil {
		t.Fatal(err)
	}
	if err := c.DeleteBlocks(ctx, "d1", "p1", []string{"b1"}); err != nil {
		t.Fatal(err)
	}
	list, err := c.BlockList(ctx, "d1", "p1", nil, 0, 0)
	if err != nil || len(list.Blocks) != 1 || list.NextStart != 201 {
		t.Fatalf("block list: %+v %v", list, err)
	}
	if task, err := c.ExportDocument(ctx, "d1", ""); err != nil || task != "T1" {
		t.Fatalf("export: %v %q", err, task)
	}
	if ss, err := c.Datasource(ctx, "d1"); err != nil || ss != "ss1" {
		t.Fatalf("datasource: %v %q", err, ss)
	}
	if sheet, err := c.AddDatasheet(ctx, "d1", Datasheet{Title: "表"}); err != nil || sheet.BlockID != "b1" {
		t.Fatalf("add datasheet: %+v %v", sheet, err)
	}
	if sheet, err := c.UpdateDatasheet(ctx, "d1", Datasheet{BlockID: "b1"}); err != nil || sheet.Title != "表" {
		t.Fatalf("update datasheet: %+v %v", sheet, err)
	}
	if err := c.DeleteDatasheet(ctx, "d1", "b1"); err != nil {
		t.Fatal(err)
	}
	result, err := c.PublishDocument(ctx, "d1", 1, nil)
	if err != nil || result.ShareCode != "S" {
		t.Fatalf("publish: %+v %v", result, err)
	}
	if err := c.CancelPublish(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	if err := c.SetPublishSetting(ctx, "d1", 4, []AuthEntry{{Type: 1, UserID: "u"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.AddBlocks(ctx, "d1", "", nil); err == nil {
		t.Fatal("missing page must fail")
	}
}

func TestFormVIPAndUpload(t *testing.T) {
	c, _ := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		switch path {
		case formCreatePath:
			_, _ = w.Write([]byte(`{"errcode":0,"formid":"F1"}`))
		case formInfoPath:
			_, _ = w.Write([]byte(`{"errcode":0,"form_info":{"formid":"F1"}}`))
		case formStatisticPath:
			_, _ = w.Write([]byte(`{"errcode":0,"fill_cnt":2,"fill_user_cnt":2,"unfill_user_cnt":8}`))
		case formAnswerPath:
			_, _ = w.Write([]byte(`{"errcode":0,"answer":{"answer_list":[]}}`))
		case docVIPAddPath, docVIPDelPath:
			_, _ = w.Write([]byte(`{"errcode":0,"succ_userid_list":["a"],"fail_userid_list":["b"]}`))
		case docVIPListPath:
			_, _ = w.Write([]byte(`{"errcode":0,"has_more":false,"userid_list":["a"]}`))
		case imageUploadPath:
			_, _ = w.Write([]byte(`{"errcode":0,"url":"https://img","height":1,"width":1,"size":3}`))
		case fileUploadPath:
			_, _ = w.Write([]byte(`{"errcode":0,"file_url":"https://file"}`))
		case fileUploadInitPath:
			_, _ = w.Write([]byte(`{"errcode":0,"hit_exist":false,"upload_key":"UK"}`))
		case fileUploadFinishPath:
			_, _ = w.Write([]byte(`{"errcode":0,"file_url":"https://file"}`))
		default:
			_, _ = w.Write([]byte(`{"errcode":0}`))
		}
	})
	ctx := context.Background()

	formID, err := c.CreateForm(ctx, "", "", FormInfo{Title: "收集"})
	if err != nil || formID != "F1" {
		t.Fatalf("create form: %v %q", err, formID)
	}
	if err := c.ModifyForm(ctx, "F1", 1, FormInfo{Title: "改"}); err != nil {
		t.Fatal(err)
	}
	if info, err := c.FormInfo(ctx, "F1"); err != nil || len(info) == 0 {
		t.Fatalf("form info: %v %s", err, info)
	}
	if stat, err := c.FormStatistic(ctx, "F1", 1); err != nil || stat.FillCount != 2 {
		t.Fatalf("form stat: %+v %v", stat, err)
	}
	if _, err := c.FormAnswer(ctx, "F1", nil); err != nil {
		t.Fatal(err)
	}
	if result, err := c.AddDocVIP(ctx, []string{"a", "b"}); err != nil || len(result.Succeeded) != 1 {
		t.Fatalf("add vip: %+v %v", result, err)
	}
	if _, err := c.RemoveDocVIP(ctx, []string{"a"}); err != nil {
		t.Fatal(err)
	}
	if list, err := c.ListDocVIP(ctx, "", 0); err != nil || len(list.UserIDs) != 1 {
		t.Fatalf("list vip: %+v %v", list, err)
	}
	if img, err := c.UploadImage(ctx, "d1", []byte("png")); err != nil || img.URL != "https://img" {
		t.Fatalf("upload image: %+v %v", img, err)
	}
	if url, err := c.UploadFile(ctx, "d1", "a.txt", []byte("x")); err != nil || url != "https://file" {
		t.Fatalf("upload file: %v %q", err, url)
	}
	up, err := c.InitChunkUpload(ctx, "d1", "big.bin", 4<<20, []string{"sha"})
	if err != nil || up.UploadKey != "UK" {
		t.Fatalf("init chunk: %+v %v", up, err)
	}
	if err := c.UploadChunk(ctx, up.UploadKey, 1, []byte("part")); err != nil {
		t.Fatal(err)
	}
	if url, err := c.FinishChunkUpload(ctx, up.UploadKey); err != nil || url != "https://file" {
		t.Fatalf("finish chunk: %v %q", err, url)
	}
	if _, err := c.UploadImage(ctx, "d1", nil); err == nil {
		t.Fatal("empty image must fail")
	}
}

func TestSecurityAndWebhookRecords(t *testing.T) {
	c, _ := newDocProbe(t, func(w http.ResponseWriter, path string, body map[string]any) {
		if path == smartsheetWebhookPath {
			if _, ok := body["add_records"]; !ok {
				if _, ok := body["update_records"]; !ok {
					t.Errorf("unexpected webhook body: %+v", body)
				}
			}
			_, _ = w.Write([]byte(`{"errcode":0}`))
			return
		}
		_, _ = w.Write([]byte(`{"errcode":0}`))
	})
	ctx := context.Background()
	if err := c.SetDocSecurity(ctx, "d1", DocSecurity{EnableReadonlyCopy: true, Watermark: &Watermark{Text: "wm"}}); err != nil {
		t.Fatal(err)
	}
	record := WebhookRecord{Values: map[string]json.RawMessage{"f1": json.RawMessage(`[{"type":"text","text":"x"}]`)}}
	if err := c.AddWebhookRecords(ctx, "KEY1", []WebhookRecord{record}); err != nil {
		t.Fatal(err)
	}
	if err := c.UpdateWebhookRecords(ctx, "KEY1", []WebhookRecord{record}); err == nil {
		t.Fatal("missing record id must fail")
	}
	record.RecordID = "r1"
	if err := c.UpdateWebhookRecords(ctx, "KEY1", []WebhookRecord{record}); err != nil {
		t.Fatal(err)
	}
	if err := c.AddWebhookRecords(ctx, " ", nil); err == nil {
		t.Fatal("missing key must fail")
	}
}
