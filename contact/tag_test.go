package contact

import (
	"context"
	"strings"
	"testing"
)

func TestTagLifecycle(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"tagid":12}`)
	ctx := context.Background()

	id, err := api.CreateTag(ctx, "UI")
	if err != nil || id != 12 || rec.path != "/cgi-bin/tag/create" || rec.body["tagname"] != "UI" {
		t.Fatalf("create tag: %d %v %s %v", id, err, rec.path, rec.body)
	}
	if err := api.UpdateTag(ctx, 12, "UX"); err != nil ||
		rec.path != "/cgi-bin/tag/update" || rec.body["tagid"] != float64(12) {
		t.Fatalf("update tag: %v %s %v", err, rec.path, rec.body)
	}
	if err := api.DeleteTag(ctx, 12); err != nil ||
		rec.path != "/cgi-bin/tag/delete" || !strings.Contains(rec.query, "tagid=12") {
		t.Fatalf("delete tag: %v %s %s", err, rec.path, rec.query)
	}
}

func TestTagValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	cases := map[string]func() error{
		"create no name": func() error { _, err := api.CreateTag(ctx, " "); return err },
		"update no id":   func() error { return api.UpdateTag(ctx, 0, "x") },
		"update no name": func() error { return api.UpdateTag(ctx, 1, "") },
		"delete no id":   func() error { return api.DeleteTag(ctx, 0) },
		"get no id":      func() error { _, err := api.GetTag(ctx, 0); return err },
		"add no members": func() error { _, err := api.AddTagMembers(ctx, 1, nil, nil); return err },
	}
	for name, fn := range cases {
		t.Run(name, func(t *testing.T) {
			if fn() == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}

func TestGetTagAndMembers(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"tagname":"乒乓","userlist":[{"userid":"u1","name":"李四"}],"partylist":[2]}`)
	detail, err := api.GetTag(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/tag/get" || !strings.Contains(rec.query, "tagid=3") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if detail.Name != "乒乓" || len(detail.Users) != 1 || detail.Parties[0] != 2 {
		t.Fatalf("unexpected detail: %+v", detail)
	}
}

func TestTagMemberChanges(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"invalidlist":["bad"],"invalidparty":[9]}`)
	ctx := context.Background()

	result, err := api.AddTagMembers(ctx, 3, []string{"u1"}, []int{2})
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/tag/addtagusers" || result.InvalidUsers[0] != "bad" || result.InvalidParties[0] != 9 {
		t.Fatalf("unexpected add: %s %+v", rec.path, result)
	}
	if _, err := api.DeleteTagMembers(ctx, 3, []string{"u1"}, nil); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/tag/deltagusers" {
		t.Fatalf("path = %s", rec.path)
	}
}

func TestListTags(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"taglist":[{"tagid":1,"tagname":"A"}]}`)
	tags, err := api.ListTags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/tag/list" || len(tags) != 1 || tags[0].Name != "A" {
		t.Fatalf("unexpected tags: %s %+v", rec.path, tags)
	}
}
