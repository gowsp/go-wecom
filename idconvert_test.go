package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/idconvert"
)

func TestIDConvert(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"corpid":"wwcorp",`+
		`"userid_list":[{"open_userid":"o1","userid":"u1"}],"invalid_open_userid_list":["o9"],`+
		`"external_userid":"e1",`+
		`"results":[{"tmp_external_userid":"t1","external_userid":"e1"}],`+
		`"invalid_tmp_external_userid_list":["t9"]}`)
	svc := idconvert.New(api)
	ctx := context.Background()

	users, invalid, err := svc.UserIDs(ctx, []string{"o1"}, 100001)
	if err != nil || len(users) != 1 || users[0].UserID != "u1" || len(invalid) != 1 ||
		rec.path != "/cgi-bin/batch/openuserid_to_userid" {
		t.Fatalf("userids: %v %+v %v %s", err, users, invalid, rec.path)
	}

	corpID, err := svc.CorpID(ctx, "open", 100001)
	if err != nil || corpID != "wwcorp" || rec.path != "/cgi-bin/corp/opencorpid_to_corpid" {
		t.Fatalf("corpid: %q %v %s", corpID, err, rec.path)
	}

	externalID, err := svc.ExternalUserID(ctx, "svc", 100001)
	if err != nil || externalID != "e1" ||
		rec.path != "/cgi-bin/externalcontact/from_service_external_userid" {
		t.Fatalf("external: %q %v %s", externalID, err, rec.path)
	}

	results, invalidTmp, err := svc.TmpExternalUserIDs(ctx, 1, 2, []string{"t1"})
	if err != nil || len(results) != 1 || len(invalidTmp) != 1 ||
		rec.path != "/cgi-bin/idconvert/convert_tmp_external_userid" {
		t.Fatalf("tmp: %v %+v %v %s", err, results, invalidTmp, rec.path)
	}

	if _, _, err := svc.UserIDs(ctx, nil, 0); err == nil {
		t.Fatal("empty user list accepted")
	}
	if _, err := svc.CorpID(ctx, " ", 0); err == nil {
		t.Fatal("empty corpid accepted")
	}
	if _, err := svc.ExternalUserID(ctx, "", 0); err == nil {
		t.Fatal("empty external id accepted")
	}
	if _, _, err := svc.TmpExternalUserIDs(ctx, 1, 1, nil); err == nil {
		t.Fatal("empty tmp list accepted")
	}
}
