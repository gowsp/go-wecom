package contact

import (
	"context"
	"strings"
	"testing"
)

func TestStartBatch(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"jobid":"job-1"}`)
	ctx := context.Background()

	jobID, err := api.SyncMembers(ctx, BatchRequest{MediaID: "media1", ToInvite: true})
	if err != nil || jobID != "job-1" {
		t.Fatalf("sync: %q %v", jobID, err)
	}
	if rec.path != "/cgi-bin/batch/syncuser" || rec.body["to_invite"] != true {
		t.Fatalf("unexpected request %s %v", rec.path, rec.body)
	}
	if _, err := api.ReplaceMembers(ctx, BatchRequest{MediaID: "m"}); err != nil ||
		rec.path != "/cgi-bin/batch/replaceuser" {
		t.Fatalf("replace members: %v %s", err, rec.path)
	}
	if _, err := api.ReplaceDepartments(ctx, BatchRequest{MediaID: "m"}); err != nil ||
		rec.path != "/cgi-bin/batch/replaceparty" {
		t.Fatalf("replace departments: %v %s", err, rec.path)
	}
}

func TestStartBatchValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	if _, err := api.SyncMembers(ctx, BatchRequest{}); err == nil {
		t.Fatal("missing media_id accepted")
	}
	if _, err := api.SyncMembers(ctx, BatchRequest{MediaID: "m", Callback: &BatchCallback{URL: "https://x"}}); err == nil {
		t.Fatal("callback without token accepted")
	}
}

func TestBatchResult(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"status":1,"type":"replace_user","total":3,"percentage":33,`+
		`"result":[{"userid":"u1","action":"update","errcode":0,"errmsg":"ok"}]}`)
	result, err := api.BatchResult(context.Background(), "job-1")
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/batch/getresult" || !strings.Contains(rec.query, "jobid=job-1") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if result.Status != 1 || result.Percentage != 33 || result.Result[0].UserID != "u1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
