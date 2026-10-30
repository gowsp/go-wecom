package wedoc

import (
	"context"
	"testing"
)

func TestAddSheetIndexRequest(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		return map[string]any{
			"errcode": 0, "errmsg": "ok",
			"properties": map[string]any{"sheet_id": "sheet-1", "title": "线索", "index": 0},
		}
	})
	sheet, err := client.AddSheet(context.Background(), "doc-1", "线索", 0)
	if err != nil {
		t.Fatal(err)
	}
	if sheet.SheetID != "sheet-1" {
		t.Fatalf("unexpected sheet: %+v", sheet)
	}
	calls := fake.recorded()
	properties := objectField(t, calls[0].Body["properties"], "properties")
	if properties["index"] != float64(0) || properties["title"] != "线索" {
		t.Fatalf("unexpected properties: %+v", properties)
	}
}

func TestUpdateAndDeleteSheetRequests(t *testing.T) {
	client, fake := newDocTestClient(t, nil)
	if err := client.UpdateSheet(context.Background(), "doc-1", "sheet-1", "新标题"); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteSheet(context.Background(), "doc-1", "sheet-1"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	if calls[0].Path != "/cgi-bin/wedoc/smartsheet/update_sheet" {
		t.Fatalf("update path = %q", calls[0].Path)
	}
	properties := objectField(t, calls[0].Body["properties"], "properties")
	if properties["sheet_id"] != "sheet-1" || properties["title"] != "新标题" {
		t.Fatalf("unexpected properties: %+v", properties)
	}
	if calls[1].Path != "/cgi-bin/wedoc/smartsheet/delete_sheet" || calls[1].Body["sheet_id"] != "sheet-1" {
		t.Fatalf("unexpected delete: %+v", calls[1])
	}
}

func TestViewRequests(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Path {
		case "/cgi-bin/wedoc/smartsheet/add_view":
			return map[string]any{"errcode": 0, "errmsg": "ok", "view": map[string]any{
				"view_id": "v1", "view_title": "甘特", "view_type": "VIEW_TYPE_GANTT",
			}}
		case "/cgi-bin/wedoc/smartsheet/update_view":
			return map[string]any{"errcode": 0, "errmsg": "ok", "view": map[string]any{
				"view_id": "v1", "view_title": "改名", "view_type": "VIEW_TYPE_GRID",
			}}
		default:
			return map[string]any{"errcode": 0, "errmsg": "ok"}
		}
	})
	ctx := context.Background()
	added, err := client.AddView(ctx, "doc-1", "sheet-1", View{
		ViewTitle:     "甘特",
		ViewType:      ViewTypeGantt,
		GanttProperty: map[string]any{"start_date_field_id": "f1", "end_date_field_id": "f2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if added.ViewID != "v1" || added.ViewType != ViewTypeGantt {
		t.Fatalf("unexpected view: %+v", added)
	}
	updated, err := client.UpdateView(ctx, "doc-1", "sheet-1", View{
		ViewID:    "v1",
		ViewTitle: "改名",
		Property:  map[string]any{"auto_sort": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ViewTitle != "改名" {
		t.Fatalf("unexpected updated view: %+v", updated)
	}
	if err := client.DeleteViews(ctx, "doc-1", "sheet-1", "v1", "v2"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 3 {
		t.Fatalf("recorded %d calls, want 3", len(calls))
	}
	add := calls[0]
	if add.Path != "/cgi-bin/wedoc/smartsheet/add_view" {
		t.Fatalf("add path = %q", add.Path)
	}
	if add.Body["docid"] != "doc-1" || add.Body["sheet_id"] != "sheet-1" ||
		add.Body["view_title"] != "甘特" || add.Body["view_type"] != ViewTypeGantt {
		t.Fatalf("unexpected add body: %+v", add.Body)
	}
	gantt := objectField(t, add.Body["property_gantt"], "property_gantt")
	if gantt["start_date_field_id"] != "f1" {
		t.Fatalf("unexpected gantt property: %+v", gantt)
	}
	update := calls[1]
	if update.Path != "/cgi-bin/wedoc/smartsheet/update_view" || update.Body["view_id"] != "v1" {
		t.Fatalf("unexpected update: %+v", update)
	}
	property := objectField(t, update.Body["property"], "property")
	if property["auto_sort"] != true {
		t.Fatalf("unexpected update property: %+v", property)
	}
	remove := calls[2]
	if remove.Path != "/cgi-bin/wedoc/smartsheet/delete_views" {
		t.Fatalf("delete path = %q", remove.Path)
	}
	ids := arrayField(t, remove.Body["view_ids"], "view_ids")
	if len(ids) != 2 || ids[0] != "v1" {
		t.Fatalf("view_ids = %v", ids)
	}
}

func TestGetViewsPaginates(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Body["offset"] {
		case float64(0):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 3,
				"has_more": true, "next": 2,
				"views": []any{
					map[string]any{"view_id": "v1", "view_title": "一", "view_type": ViewTypeGrid},
					map[string]any{"view_id": "v2", "view_title": "二", "view_type": ViewTypeGrid},
				},
			}
		case float64(2):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 3,
				"has_more": false, "next": 0,
				"views": []any{map[string]any{"view_id": "v3", "view_title": "三", "view_type": ViewTypeKanban}},
			}
		default:
			return map[string]any{"errcode": 40013, "errmsg": "unexpected offset"}
		}
	})
	views, err := client.GetViews(context.Background(), "doc-1", "sheet-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 3 || views[2].ViewID != "v3" {
		t.Fatalf("unexpected views: %+v", views)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	wantOffsets := []float64{0, 2}
	for index, call := range calls {
		if call.Path != "/cgi-bin/wedoc/smartsheet/get_views" {
			t.Fatalf("call %d path = %q", index, call.Path)
		}
		if call.Body["docid"] != "doc-1" || call.Body["sheet_id"] != "sheet-1" {
			t.Fatalf("call %d target = %+v", index, call.Body)
		}
		if call.Body["offset"] != wantOffsets[index] || call.Body["limit"] != float64(maxListPageSize) {
			t.Fatalf("call %d paging = %+v", index, call.Body)
		}
	}
}

func TestUpdateAndDeleteFields(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		if call.Path == "/cgi-bin/wedoc/smartsheet/update_fields" {
			return map[string]any{"errcode": 0, "errmsg": "ok", "fields": call.Body["fields"]}
		}
		return map[string]any{"errcode": 0, "errmsg": "ok"}
	})
	fields, err := client.UpdateFields(context.Background(), "doc-1", "sheet-1", Field{
		FieldID: "f1", FieldTitle: "新标题", FieldType: FieldTypeText,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0].FieldID != "f1" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
	if err := client.DeleteFields(context.Background(), "doc-1", "sheet-1", "f1", "f2"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	if calls[0].Path != "/cgi-bin/wedoc/smartsheet/update_fields" {
		t.Fatalf("update path = %q", calls[0].Path)
	}
	updated := arrayField(t, calls[0].Body["fields"], "fields")
	field := objectField(t, updated[0], "field")
	if field["field_id"] != "f1" || field["field_title"] != "新标题" || field["field_type"] != FieldTypeText {
		t.Fatalf("unexpected field: %+v", field)
	}
	if calls[1].Path != "/cgi-bin/wedoc/smartsheet/delete_fields" {
		t.Fatalf("delete path = %q", calls[1].Path)
	}
	ids := arrayField(t, calls[1].Body["field_ids"], "field_ids")
	if len(ids) != 2 || ids[1] != "f2" {
		t.Fatalf("field_ids = %v", ids)
	}
}

func TestGetFieldsPaginates(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Body["offset"] {
		case float64(0):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 2,
				"fields": []any{map[string]any{"field_id": "f1", "field_title": "标题", "field_type": FieldTypeText}},
			}
		case float64(1):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 2,
				"fields": []any{map[string]any{
					"field_id": "f2", "field_title": "金额", "field_type": FieldTypeNumber,
					"property_number": map[string]any{"decimal_places": 2},
				}},
			}
		default:
			return map[string]any{"errcode": 40013, "errmsg": "unexpected offset"}
		}
	})
	fields, err := client.GetFields(context.Background(), "doc-1", "sheet-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 2 || fields[1].FieldID != "f2" {
		t.Fatalf("unexpected fields: %+v", fields)
	}
	property, ok := fields[1].Property.(map[string]any)
	if !ok || property["decimal_places"] != float64(2) {
		t.Fatalf("unexpected property: %#v", fields[1].Property)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	wantOffsets := []float64{0, 1}
	for index, call := range calls {
		if call.Path != "/cgi-bin/wedoc/smartsheet/get_fields" {
			t.Fatalf("call %d path = %q", index, call.Path)
		}
		if call.Body["offset"] != wantOffsets[index] || call.Body["limit"] != float64(maxListPageSize) {
			t.Fatalf("call %d paging = %+v", index, call.Body)
		}
	}
}

func TestFieldGroupRequests(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Path {
		case "/cgi-bin/wedoc/smartsheet/add_field_group", "/cgi-bin/wedoc/smartsheet/update_field_group":
			return map[string]any{"errcode": 0, "errmsg": "ok", "field_group": map[string]any{
				"field_group_id": "g1", "name": "基础", "children": []any{map[string]any{"field_id": "f1"}},
			}}
		default:
			return map[string]any{"errcode": 0, "errmsg": "ok"}
		}
	})
	ctx := context.Background()
	added, err := client.AddFieldGroup(ctx, "doc-1", "sheet-1", "基础", []string{"f1", "f2"})
	if err != nil {
		t.Fatal(err)
	}
	if added.FieldGroupID != "g1" {
		t.Fatalf("unexpected group: %+v", added)
	}
	updated, err := client.UpdateFieldGroup(ctx, "doc-1", "sheet-1", "g1", "基础2", []string{"f2"})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "基础" {
		t.Fatalf("unexpected group: %+v", updated)
	}
	if err := client.DeleteFieldGroups(ctx, "doc-1", "sheet-1", "g1", "g2"); err != nil {
		t.Fatal(err)
	}
	calls := fake.recorded()
	if len(calls) != 3 {
		t.Fatalf("recorded %d calls, want 3", len(calls))
	}
	add := calls[0]
	if add.Path != "/cgi-bin/wedoc/smartsheet/add_field_group" || add.Body["name"] != "基础" {
		t.Fatalf("unexpected add: %+v", add)
	}
	children := arrayField(t, add.Body["children"], "children")
	child := objectField(t, children[1], "child")
	if child["field_id"] != "f2" {
		t.Fatalf("children = %v", children)
	}
	update := calls[1]
	if update.Path != "/cgi-bin/wedoc/smartsheet/update_field_group" || update.Body["field_group_id"] != "g1" {
		t.Fatalf("unexpected update: %+v", update)
	}
	if update.Body["name"] != "基础2" {
		t.Fatalf("unexpected update name: %+v", update.Body)
	}
	remove := calls[2]
	if remove.Path != "/cgi-bin/wedoc/smartsheet/delete_field_groups" {
		t.Fatalf("delete path = %q", remove.Path)
	}
	ids := arrayField(t, remove.Body["field_group_ids"], "field_group_ids")
	if len(ids) != 2 {
		t.Fatalf("field_group_ids = %v", ids)
	}
}

func TestGetFieldGroupsPaginates(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Body["offset"] {
		case float64(0):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 2,
				"has_more": true, "next": 1,
				"field_groups": []any{map[string]any{"field_group_id": "g1", "name": "一"}},
			}
		case float64(1):
			return map[string]any{
				"errcode": 0, "errmsg": "ok", "total": 2,
				"has_more": false, "next": 0,
				"field_groups": []any{map[string]any{"field_group_id": "g2", "name": "二"}},
			}
		default:
			return map[string]any{"errcode": 40013, "errmsg": "unexpected offset"}
		}
	})
	groups, err := client.GetFieldGroups(context.Background(), "doc-1", "sheet-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 2 || groups[1].FieldGroupID != "g2" {
		t.Fatalf("unexpected groups: %+v", groups)
	}
	calls := fake.recorded()
	if len(calls) != 2 {
		t.Fatalf("recorded %d calls, want 2", len(calls))
	}
	if calls[0].Body["offset"] != float64(0) || calls[1].Body["offset"] != float64(1) {
		t.Fatalf("unexpected offsets: %v %v", calls[0].Body["offset"], calls[1].Body["offset"])
	}
}

func TestContentPrivilegeRequests(t *testing.T) {
	client, fake := newDocTestClient(t, func(index int, call docAPICall) map[string]any {
		switch call.Path {
		case "/cgi-bin/wedoc/smartsheet/content_priv/get_sheet_priv":
			return map[string]any{
				"errcode": 0, "errmsg": "ok",
				"rule_list": []any{map[string]any{
					"rule_id": 2, "type": 2, "name": "成员",
					"priv_list": []any{map[string]any{
						"sheet_id": "sheet-1", "priv": 2,
						"field_priv": map[string]any{
							"field_range_type": 2,
							"field_rule_list": []any{map[string]any{
								"field_id": "f1", "field_type": FieldTypeText,
								"can_edit": true, "can_insert": true, "can_view": true,
							}},
						},
						"record_priv": map[string]any{"record_range_type": 1},
					}},
				}},
			}
		case "/cgi-bin/wedoc/smartsheet/content_priv/create_rule":
			return map[string]any{"errcode": 0, "errmsg": "ok", "rule_id": 7}
		default:
			return map[string]any{"errcode": 0, "errmsg": "ok"}
		}
	})
	ctx := context.Background()
	rules, err := client.GetContentPrivileges(ctx, "doc-1", ContentRuleExtra, []uint32{2})
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 1 || rules[0].RuleID != 2 || len(rules[0].PrivList) != 1 {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	privilege := rules[0].PrivList[0]
	if privilege.Priv != SheetPrivEdit || privilege.FieldPriv == nil ||
		privilege.FieldPriv.FieldRuleList[0].CanEdit != true {
		t.Fatalf("unexpected privilege: %+v", privilege)
	}

	err = client.UpdateContentPrivilege(ctx, "doc-1", ContentRule{
		RuleID: 2, Type: ContentRuleExtra, Name: "编辑者",
		PrivList: []SheetPrivilege{{
			SheetID:         "sheet-1",
			Priv:            SheetPrivEdit,
			CanInsertRecord: boolPointer(true),
			FieldPriv: &FieldPrivilege{
				FieldRangeType: FieldRangePartial,
				FieldRuleList: []FieldRule{{
					FieldID: "f1", FieldType: FieldTypeText, CanView: true,
				}},
			},
			RecordPriv: &RecordPrivilege{RecordRangeType: RecordRangeAll},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	ruleID, err := client.CreateContentPrivilegeRule(ctx, "doc-1", "成员")
	if err != nil {
		t.Fatal(err)
	}
	if ruleID != 7 {
		t.Fatalf("rule id = %d", ruleID)
	}
	if err := client.ModifyContentPrivilegeMembers(ctx, "doc-1", 7,
		[]string{"user-1"}, []string{"user-2"}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteContentPrivilegeRules(ctx, "doc-1", 7, 8); err != nil {
		t.Fatal(err)
	}

	calls := fake.recorded()
	if len(calls) != 5 {
		t.Fatalf("recorded %d calls, want 5", len(calls))
	}
	get := calls[0]
	if get.Path != "/cgi-bin/wedoc/smartsheet/content_priv/get_sheet_priv" ||
		get.Body["type"] != float64(ContentRuleExtra) {
		t.Fatalf("unexpected get: %+v", get)
	}
	ruleIDs := arrayField(t, get.Body["rule_id_list"], "rule_id_list")
	if len(ruleIDs) != 1 || ruleIDs[0] != float64(2) {
		t.Fatalf("rule_id_list = %v", ruleIDs)
	}

	update := calls[1]
	if update.Path != "/cgi-bin/wedoc/smartsheet/content_priv/update_sheet_priv" {
		t.Fatalf("update path = %q", update.Path)
	}
	if update.Body["rule_id"] != float64(2) || update.Body["name"] != "编辑者" {
		t.Fatalf("unexpected update: %+v", update.Body)
	}
	privList := arrayField(t, update.Body["priv_list"], "priv_list")
	priv := objectField(t, privList[0], "priv")
	if priv["sheet_id"] != "sheet-1" || priv["priv"] != float64(SheetPrivEdit) ||
		priv["can_insert_record"] != true {
		t.Fatalf("unexpected priv: %+v", priv)
	}
	fieldPriv := objectField(t, priv["field_priv"], "field_priv")
	if fieldPriv["field_range_type"] != float64(FieldRangePartial) {
		t.Fatalf("unexpected field_priv: %+v", fieldPriv)
	}

	create := calls[2]
	if create.Path != "/cgi-bin/wedoc/smartsheet/content_priv/create_rule" || create.Body["name"] != "成员" {
		t.Fatalf("unexpected create: %+v", create)
	}

	modify := calls[3]
	if modify.Path != "/cgi-bin/wedoc/smartsheet/content_priv/mod_rule_member" ||
		modify.Body["rule_id"] != float64(7) {
		t.Fatalf("unexpected modify: %+v", modify)
	}
	add := objectField(t, modify.Body["add_member_range"], "add_member_range")
	addUsers := arrayField(t, add["userid_list"], "userid_list")
	if len(addUsers) != 1 || addUsers[0] != "user-1" {
		t.Fatalf("unexpected add range: %+v", add)
	}
	del := objectField(t, modify.Body["del_member_range"], "del_member_range")
	delUsers := arrayField(t, del["userid_list"], "userid_list")
	if len(delUsers) != 1 || delUsers[0] != "user-2" {
		t.Fatalf("unexpected del range: %+v", del)
	}

	remove := calls[4]
	if remove.Path != "/cgi-bin/wedoc/smartsheet/content_priv/delete_rule" {
		t.Fatalf("delete path = %q", remove.Path)
	}
	removeIDs := arrayField(t, remove.Body["rule_id_list"], "rule_id_list")
	if len(removeIDs) != 2 {
		t.Fatalf("rule_id_list = %v", removeIDs)
	}
}
