package contact

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func TestCreateDepartment(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"id":42}`)
	id, err := api.CreateDepartment(context.Background(), Department{Name: "研发部", ParentID: 1, Order: 10})
	if err != nil {
		t.Fatal(err)
	}
	if rec.method != http.MethodPost || rec.path != "/cgi-bin/department/create" {
		t.Fatalf("unexpected request %s %s", rec.method, rec.path)
	}
	if id != 42 || rec.body["name"] != "研发部" || rec.body["parentid"] != float64(1) {
		t.Fatalf("unexpected result %d %v", id, rec.body)
	}
	if _, ok := rec.body["department_leader"]; ok {
		t.Fatal("read-only field must not be sent")
	}
}

func TestCreateDepartmentFallsBackToExplicitID(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	id, err := api.CreateDepartment(context.Background(), Department{ID: 7, Name: "x"})
	if err != nil || id != 7 {
		t.Fatalf("id = %d, err = %v", id, err)
	}
}

func TestDepartmentValidation(t *testing.T) {
	_, api := newContact(t, `{"errcode":0}`)
	ctx := context.Background()
	if _, err := api.CreateDepartment(ctx, Department{}); err == nil {
		t.Fatal("department without name accepted")
	}
	if err := api.UpdateDepartment(ctx, Department{Name: "x"}); err == nil {
		t.Fatal("update without id accepted")
	}
	if err := api.DeleteDepartment(ctx, 0); err == nil {
		t.Fatal("delete without id accepted")
	}
	if _, err := api.GetDepartment(ctx, 0); err == nil {
		t.Fatal("get without id accepted")
	}
}

func TestListDepartments(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"department":[{"id":1,"name":"root","parentid":0}]}`)
	list, err := api.ListDepartments(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/department/list" || strings.Contains(rec.query, "id=") {
		t.Fatalf("unexpected request %s?%s", rec.path, rec.query)
	}
	if len(list) != 1 || list[0].Name != "root" {
		t.Fatalf("unexpected list: %+v", list)
	}

	if _, err := api.ListDepartments(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rec.query, "id=1") {
		t.Fatalf("query = %s", rec.query)
	}
}

func TestListSubDepartmentIDs(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"department_id":[{"id":3,"parentid":1,"order":5}]}`)
	list, err := api.ListSubDepartmentIDs(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/department/simplelist" || len(list) != 1 || list[0].ID != 3 {
		t.Fatalf("unexpected result %s %+v", rec.path, list)
	}
}

func TestGetAndUpdateDepartment(t *testing.T) {
	rec, api := newContact(t, `{"errcode":0,"department":{"id":5,"name":"A","department_leader":["u1"]}}`)
	department, err := api.GetDepartment(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/department/get" || department.Name != "A" ||
		len(department.DepartmentLeader) != 1 {
		t.Fatalf("unexpected department %+v", department)
	}

	if err := api.UpdateDepartment(context.Background(), Department{ID: 5, Name: "B"}); err != nil {
		t.Fatal(err)
	}
	if rec.path != "/cgi-bin/department/update" || rec.body["name"] != "B" {
		t.Fatalf("unexpected update %s %v", rec.path, rec.body)
	}
}
