package wecom_test

import (
	"context"
	"testing"

	"github.com/gowsp/go-wecom/wedrive"
)

func TestWeDriveSpaces(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"spaceid":"sp1","space_share_url":"https://share",`+
		`"space_info":{"spaceid":"sp1","space_name":"项目空间",`+
		`"auth_list":{"auth_info":[{"type":1,"userid":"u1","auth":3}]}}}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	id, err := svc.CreateSpace(ctx, "项目空间", 0, []wedrive.AuthInfo{
		{Type: wedrive.MemberUser, UserID: "u1", Auth: wedrive.AuthEdit},
	})
	if err != nil || id != "sp1" || rec.path != "/cgi-bin/wedrive/space_create" {
		t.Fatalf("create: %q %v %s", id, err, rec.path)
	}
	authList, ok := rec.body["auth_info"].([]any)
	if !ok || len(authList) != 1 {
		t.Fatalf("auth_info not sent: %v", rec.body)
	}

	if err := svc.RenameSpace(ctx, "sp1", "新名字"); err != nil ||
		rec.path != "/cgi-bin/wedrive/space_rename" {
		t.Fatalf("rename: %v %s", err, rec.path)
	}

	info, err := svc.GetSpace(ctx, "sp1")
	if err != nil || info.SpaceName != "项目空间" || rec.path != "/cgi-bin/wedrive/space_info" {
		t.Fatalf("get space: %v %+v %s", err, info, rec.path)
	}
	if info.AuthList == nil || info.AuthList.AuthInfo[0].Auth != wedrive.AuthEdit {
		t.Fatalf("auth list: %+v", info.AuthList)
	}

	if _, err := svc.GetNewSpaceInfo(ctx, "sp1"); err != nil ||
		rec.path != "/cgi-bin/wedrive/new_space_info" {
		t.Fatalf("new space info: %v %s", err, rec.path)
	}

	if err := svc.AddSpaceMembers(ctx, "sp1", []wedrive.AuthInfo{{Type: 2, DepartmentID: 3, Auth: 1}}); err != nil ||
		rec.path != "/cgi-bin/wedrive/space_acl_add" {
		t.Fatalf("add members: %v %s", err, rec.path)
	}
	if err := svc.RemoveSpaceMembers(ctx, "sp1", []wedrive.AuthInfo{{Type: 2, DepartmentID: 3}}); err != nil ||
		rec.path != "/cgi-bin/wedrive/space_acl_del" {
		t.Fatalf("remove members: %v %s", err, rec.path)
	}

	watermark := true
	if err := svc.SetSpaceSetting(ctx, wedrive.SpaceSetting{SpaceID: "sp1", EnableWatermark: &watermark}); err != nil ||
		rec.path != "/cgi-bin/wedrive/space_setting" {
		t.Fatalf("setting: %v %s", err, rec.path)
	}
	if rec.body["enable_watermark"] != true {
		t.Fatalf("watermark not sent: %v", rec.body)
	}
	if _, ok := rec.body["ban_share_external"]; ok {
		t.Fatalf("unset boolean must be omitted: %v", rec.body)
	}

	link, err := svc.SpaceShareLink(ctx, "sp1")
	if err != nil || link != "https://share" || rec.path != "/cgi-bin/wedrive/space_share" {
		t.Fatalf("share link: %q %v %s", link, err, rec.path)
	}

	if err := svc.DismissSpace(ctx, "sp1"); err != nil ||
		rec.path != "/cgi-bin/wedrive/space_dismiss" {
		t.Fatalf("dismiss: %v %s", err, rec.path)
	}
}

func TestWeDriveValidation(t *testing.T) {
	_, api := newOffice(t, `{"errcode":0}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	if _, err := svc.CreateSpace(ctx, " ", 0, nil); err == nil {
		t.Fatal("empty space name accepted")
	}
	if err := svc.RenameSpace(ctx, "sp1", ""); err == nil {
		t.Fatal("rename without name accepted")
	}
	if err := svc.DismissSpace(ctx, ""); err == nil {
		t.Fatal("dismiss without id accepted")
	}
	if _, err := svc.GetSpace(ctx, " "); err == nil {
		t.Fatal("get without id accepted")
	}
	if _, err := svc.GetNewSpaceInfo(ctx, ""); err == nil {
		t.Fatal("new space info without id accepted")
	}
	if err := svc.AddSpaceMembers(ctx, "sp1", nil); err == nil {
		t.Fatal("add without members accepted")
	}
	if err := svc.RemoveSpaceMembers(ctx, "", nil); err == nil {
		t.Fatal("remove without ids accepted")
	}
	if err := svc.SetSpaceSetting(ctx, wedrive.SpaceSetting{}); err == nil {
		t.Fatal("setting without id accepted")
	}
	if _, err := svc.SpaceShareLink(ctx, ""); err == nil {
		t.Fatal("share link without id accepted")
	}
}

func TestWeDriveFiles(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"fileid":"f1","upload_key":"uk","url":"https://doc",`+
		`"has_more":true,"next_start":5,`+
		`"file_list":{"item":[{"fileid":"f1","file_name":"a.txt","file_size":10}]},`+
		`"file":{"fileid":"f1","file_name":"b.txt"},"file_info":{"fileid":"f1","file_name":"a.txt"},`+
		`"download_url":"https://dl","cookie_name":"ck","cookie_value":"cv","share_url":"https://share"}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	page, err := svc.ListFiles(ctx, "sp1", "root", wedrive.SortByNameAsc, 0, 100)
	if err != nil || rec.path != "/cgi-bin/wedrive/file_list" || !page.HasMore ||
		page.NextStart != 5 || len(page.Items) != 1 {
		t.Fatalf("list: %v %+v %s", err, page, rec.path)
	}

	id, err := svc.UploadFile(ctx, "sp1", "root", "", "a.txt", "YmFzZTY0")
	if err != nil || id != "f1" || rec.path != "/cgi-bin/wedrive/file_upload" {
		t.Fatalf("upload: %q %v %s", id, err, rec.path)
	}

	session, err := svc.UploadInit(ctx, wedrive.UploadInitRequest{
		SpaceID: "sp1", FatherID: "root", FileName: "big.bin", Size: 10, BlockSHA: []string{"sha1"},
	})
	if err != nil || session.UploadKey != "uk" || session.FileID != "f1" ||
		rec.path != "/cgi-bin/wedrive/file_upload_init" {
		t.Fatalf("upload init: %+v %v %s", session, err, rec.path)
	}
	if err := svc.UploadBlock(ctx, "uk", 1, "YmxvY2s="); err != nil ||
		rec.path != "/cgi-bin/wedrive/file_upload_part" {
		t.Fatalf("upload block: %v %s", err, rec.path)
	}
	if rec.body["index"] != float64(1) || rec.body["upload_key"] != "uk" {
		t.Fatalf("block payload: %v", rec.body)
	}
	if fileID, err := svc.UploadFinish(ctx, "uk"); err != nil || fileID != "f1" ||
		rec.path != "/cgi-bin/wedrive/file_upload_finish" {
		t.Fatalf("upload finish: %q %v %s", fileID, err, rec.path)
	}

	download, err := svc.DownloadFile(ctx, "f1", "")
	if err != nil || download.URL != "https://dl" || download.CookieName != "ck" ||
		download.CookieValue != "cv" || rec.path != "/cgi-bin/wedrive/file_download" {
		t.Fatalf("download: %+v %v %s", download, err, rec.path)
	}

	newID, docURL, err := svc.CreateFile(ctx, "sp1", "root", wedrive.FileTypeDoc, "文档")
	if err != nil || newID != "f1" || docURL != "https://doc" ||
		rec.path != "/cgi-bin/wedrive/file_create" {
		t.Fatalf("create: %q %q %v %s", newID, docURL, err, rec.path)
	}

	renamed, err := svc.RenameFile(ctx, "f1", "b.txt")
	if err != nil || renamed.FileName != "b.txt" || rec.path != "/cgi-bin/wedrive/file_rename" {
		t.Fatalf("rename: %v %+v %s", err, renamed, rec.path)
	}

	moved, err := svc.MoveFiles(ctx, "dest", []string{"f1"}, true)
	if err != nil || len(moved) != 1 || rec.path != "/cgi-bin/wedrive/file_move" {
		t.Fatalf("move: %v %+v %s", err, moved, rec.path)
	}

	if err := svc.DeleteFiles(ctx, []string{"f1"}); err != nil ||
		rec.path != "/cgi-bin/wedrive/file_delete" {
		t.Fatalf("delete: %v %s", err, rec.path)
	}

	file, err := svc.GetFile(ctx, "f1")
	if err != nil || file.FileName != "a.txt" || rec.path != "/cgi-bin/wedrive/file_info" {
		t.Fatalf("info: %v %+v %s", err, file, rec.path)
	}
}

func TestWeDriveFilePermissions(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"share_url":"https://share",`+
		`"share_range":{"enable_corp_internal":true,"corp_internal_auth":1},`+
		`"secure_setting":{"ban_share_external":true},`+
		`"inherit_father_auth":{"inherit":true,"auth_list":[{"type":1,"userid":"u8","auth":1}]},`+
		`"file_member_list":[{"type":1,"userid":"u9","auth":1}],`+
		`"watermark":{"text":"WM","margin_type":1,"force_by_admin":true}}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	if err := svc.AddFileMembers(ctx, "f1", []wedrive.AuthInfo{{Type: 1, UserID: "u1", Auth: 1}}); err != nil ||
		rec.path != "/cgi-bin/wedrive/file_acl_add" {
		t.Fatalf("add members: %v %s", err, rec.path)
	}
	if err := svc.RemoveFileMembers(ctx, "f1", []wedrive.AuthInfo{{Type: 2, DepartmentID: 3}}); err != nil ||
		rec.path != "/cgi-bin/wedrive/file_acl_del" {
		t.Fatalf("remove members: %v %s", err, rec.path)
	}
	rec.body = nil
	if err := svc.SetFileShareScope(ctx, "f1", 1, 0); err != nil ||
		rec.path != "/cgi-bin/wedrive/file_setting" {
		t.Fatalf("share scope: %v %s", err, rec.path)
	}
	// auth is optional: zero keeps the current permission, so it is not sent.
	if _, ok := rec.body["auth"]; ok {
		t.Fatalf("zero auth must be omitted: %v", rec.body)
	}
	if err := svc.SetFileShareScope(ctx, "f1", 4, 1); err != nil {
		t.Fatalf("share scope with auth: %v", err)
	}
	if rec.body["auth"] != float64(1) || rec.body["auth_scope"] != float64(4) {
		t.Fatalf("share scope body: %v", rec.body)
	}
	link, err := svc.FileShareLink(ctx, "f1")
	if err != nil || link != "https://share" || rec.path != "/cgi-bin/wedrive/file_share" {
		t.Fatalf("share link: %q %v %s", link, err, rec.path)
	}
	permission, err := svc.GetFilePermission(ctx, "f1")
	if err != nil || !permission.ShareRange.EnableCorpInternal ||
		!permission.SecureSetting.BanShareExternal ||
		rec.path != "/cgi-bin/wedrive/get_file_permission" {
		t.Fatalf("permission: %v %+v %s", err, permission, rec.path)
	}
	if permission.Watermark.Text != "WM" || !permission.Watermark.ForceByAdmin {
		t.Fatalf("watermark not read: %+v", permission.Watermark)
	}
	if !permission.InheritFatherAuth.Inherit || len(permission.InheritFatherAuth.AuthList) != 1 {
		t.Fatalf("inherit auth not read: %+v", permission.InheritFatherAuth)
	}
	if len(permission.FileMemberList) != 1 || permission.FileMemberList[0].UserID != "u9" {
		t.Fatalf("file members not read: %+v", permission.FileMemberList)
	}
	rec.body = nil
	watermarkText := "机密"
	hideName := false
	if err := svc.SetFileWatermark(ctx, "f1", wedrive.Watermark{
		Text: &watermarkText, ShowVisitorName: &hideName,
	}); err != nil || rec.path != "/cgi-bin/wedrive/file_secure_setting" {
		t.Fatalf("watermark: %v %s", err, rec.path)
	}
	block := rec.body["watermark"].(map[string]any)
	if block["text"] != "机密" {
		t.Fatalf("watermark text not sent: %v", block)
	}
	// An explicit false must travel; omitempty used to drop it.
	if value, ok := block["show_visitor_name"]; !ok || value != false {
		t.Fatalf("explicit false not sent: %v", block)
	}
	if _, ok := block["margin_type"]; ok {
		t.Fatalf("unset margin_type must be omitted: %v", block)
	}

	if _, err := svc.ListFiles(ctx, "", "", 0, 0, 0); err == nil {
		t.Fatal("list without space accepted")
	}
	if _, err := svc.ListFiles(ctx, "sp1", "", 1, 0, 100); err == nil {
		t.Fatal("list without fatherid accepted")
	}
	if _, err := svc.ListFiles(ctx, "sp1", "root", 0, 0, 100); err == nil {
		t.Fatal("list without a sort order accepted")
	}
	if _, err := svc.ListFiles(ctx, "sp1", "root", 7, 0, 100); err == nil {
		t.Fatal("list with an unknown sort order accepted")
	}
	if _, err := svc.ListFiles(ctx, "sp1", "root", 1, 0, 1001); err == nil {
		t.Fatal("list above the page cap accepted")
	}
	if _, err := svc.ListFiles(ctx, "sp1", "root", 1, 0, 0); err == nil {
		t.Fatal("list without a limit accepted")
	}
	if _, err := svc.UploadFile(ctx, "sp1", "", "", "", ""); err == nil {
		t.Fatal("upload without content accepted")
	}
	if _, err := svc.UploadInit(ctx, wedrive.UploadInitRequest{SpaceID: "sp1", FileName: "x"}); err == nil {
		t.Fatal("upload init without size accepted")
	}
	if err := svc.UploadBlock(ctx, "", 0, ""); err == nil {
		t.Fatal("block without input accepted")
	}
	if _, err := svc.UploadFinish(ctx, ""); err == nil {
		t.Fatal("finish without key accepted")
	}
	if _, err := svc.DownloadFile(ctx, "", ""); err == nil {
		t.Fatal("download without id accepted")
	}
	if _, _, err := svc.CreateFile(ctx, "sp1", "", 0, ""); err == nil {
		t.Fatal("create without name accepted")
	}
	if _, err := svc.RenameFile(ctx, "f1", ""); err == nil {
		t.Fatal("rename without name accepted")
	}
	if _, err := svc.MoveFiles(ctx, "", nil, false); err == nil {
		t.Fatal("move without target accepted")
	}
	if err := svc.DeleteFiles(ctx, nil); err == nil {
		t.Fatal("delete without ids accepted")
	}
	if _, err := svc.GetFile(ctx, ""); err == nil {
		t.Fatal("get without id accepted")
	}
	if err := svc.AddFileMembers(ctx, "f1", nil); err == nil {
		t.Fatal("add members without list accepted")
	}
	if err := svc.SetFileShareScope(ctx, "", 0, 0); err == nil {
		t.Fatal("share scope without id accepted")
	}
	if _, err := svc.FileShareLink(ctx, ""); err == nil {
		t.Fatal("share link without id accepted")
	}
	if _, err := svc.GetFilePermission(ctx, ""); err == nil {
		t.Fatal("permission without id accepted")
	}
	if err := svc.SetFileWatermark(ctx, "", wedrive.Watermark{}); err == nil {
		t.Fatal("watermark without id accepted")
	}
}

func TestWeDriveVIPAndCapacity(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"has_more":true,"next_cursor":"c2","userid_list":["u1"],`+
		`"is_pro":true,"total_vip_acct_num":10,"use_vip_acct_num":5,"pro_expire_time":100,`+
		`"total_capacity_for_all":1024,"total_capacity_for_vip":512}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	if err := svc.AddVIPAccounts(ctx, []string{"u1"}); err != nil ||
		rec.path != "/cgi-bin/wedrive/vip/batch_add" {
		t.Fatalf("add vip: %v %s", err, rec.path)
	}
	if err := svc.RemoveVIPAccounts(ctx, []string{"u1"}); err != nil ||
		rec.path != "/cgi-bin/wedrive/vip/batch_del" {
		t.Fatalf("remove vip: %v %s", err, rec.path)
	}
	page, err := svc.ListVIPAccounts(ctx, "", 2)
	if err != nil || len(page.UserIDs) != 1 || page.NextCursor != "c2" ||
		rec.path != "/cgi-bin/wedrive/vip/list" {
		t.Fatalf("list vip: %v %+v %s", err, page, rec.path)
	}

	pro, err := svc.ProInfo(ctx)
	if err != nil || !pro.IsPro || pro.TotalVIPNumber != 10 ||
		rec.path != "/cgi-bin/wedrive/mng_pro_info" {
		t.Fatalf("pro info: %v %+v %s", err, pro, rec.path)
	}
	capacity, err := svc.Capacity(ctx)
	if err != nil || capacity.TotalForAll != 1024 ||
		rec.path != "/cgi-bin/wedrive/mng_capacity" {
		t.Fatalf("capacity: %v %+v %s", err, capacity, rec.path)
	}

	if err := svc.AddVIPAccounts(ctx, nil); err == nil {
		t.Fatal("empty vip list accepted")
	}
	if err := svc.RemoveVIPAccounts(ctx, nil); err == nil {
		t.Fatal("empty vip removal accepted")
	}
}

func TestWeDriveEvent(t *testing.T) {
	capacityEvent := `<xml><CreateTime>1</CreateTime><MsgType>event</MsgType>` +
		`<Event><![CDATA[wedrive_insufficient_capacity]]></Event></xml>`
	event, err := wedrive.ParseEvent([]byte(capacityEvent))
	if err != nil || event.Name != "wedrive_insufficient_capacity" {
		t.Fatalf("capacity event: %v %+v", err, event)
	}

	spaceEvent := `<xml><MsgType>event</MsgType><Event><![CDATA[wedrive_space_change]]></Event>` +
		`<ChangeType><![CDATA[dismiss_space]]></ChangeType>` +
		`<SpaceId><![CDATA[sp1]]></SpaceId><SpaceId><![CDATA[sp2]]></SpaceId></xml>`
	event, err = wedrive.ParseEvent([]byte(spaceEvent))
	if err != nil || len(event.SpaceIDs) != 2 || event.ChangeType != wedrive.ChangeDismissSpace {
		t.Fatalf("space event: %v %+v", err, event)
	}

	fileEvent := `<xml><MsgType>event</MsgType><Event><![CDATA[wedrive_file_change]]></Event>` +
		`<ChangeType><![CDATA[add_file]]></ChangeType><FileId><![CDATA[f1]]></FileId></xml>`
	event, err = wedrive.ParseEvent([]byte(fileEvent))
	if err != nil || event.FileID != "f1" {
		t.Fatalf("file event: %v %+v", err, event)
	}

	for name, payload := range map[string]string{
		"unexpected event": `<xml><Event>other</Event></xml>`,
		"unknown change": `<xml><Event>wedrive_space_change</Event>` +
			`<ChangeType>unknown</ChangeType><SpaceId>sp1</SpaceId></xml>`,
		"space without id": `<xml><Event>wedrive_space_change</Event>` +
			`<ChangeType>dismiss_space</ChangeType></xml>`,
		"file without id": `<xml><Event>wedrive_file_change</Event></xml>`,
		"malformed":       `<xml>`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := wedrive.ParseEvent([]byte(payload)); err == nil {
				t.Fatal("invalid event accepted")
			}
		})
	}
}

func TestWeDriveUploadHitExist(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"hit_exist":true,"fileid":"f1"}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	session, err := svc.UploadInit(ctx, wedrive.UploadInitRequest{
		SpaceID: "sp1", FileName: "a.txt", Size: 10, BlockSHA: []string{"sha"},
	})
	if err != nil || !session.HitExist || session.FileID != "f1" ||
		rec.path != "/cgi-bin/wedrive/file_upload_init" {
		t.Fatalf("hit exist: %+v %v %s", session, err, rec.path)
	}
	if session.UploadKey != "" {
		t.Fatalf("hit exist must not carry an upload key: %q", session.UploadKey)
	}
}

func TestWeDriveUploadMissingKey(t *testing.T) {
	rec, api := newOffice(t, `{"errcode":0,"fileid":"f1"}`)
	svc := wedrive.New(api)
	ctx := context.Background()

	if _, err := svc.UploadInit(ctx, wedrive.UploadInitRequest{
		SpaceID: "sp1", FileName: "a.txt", Size: 10, BlockSHA: []string{"sha"},
	}); err == nil {
		t.Fatalf("response without upload_key accepted: %s", rec.path)
	}
}
