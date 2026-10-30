package contact

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Department is one department. DepartmentLeader holds userids of department
// leaders and appears on reads, not on writes.
type Department struct {
	ID               int      `json:"id"`
	Name             string   `json:"name,omitempty"`
	NameEn           string   `json:"name_en,omitempty"`
	ParentID         int      `json:"parentid,omitempty"`
	Order            int      `json:"order,omitempty"`
	DepartmentLeader []string `json:"department_leader,omitempty"`
}

// CreateDepartment adds one department under ParentID and returns its id (0 when
// the server assigns one and the response omits it).
//
// 文档: https://developer.work.weixin.qq.com/document/path/90205
func (c *Contact) CreateDepartment(ctx context.Context, department Department) (int, error) {
	if strings.TrimSpace(department.Name) == "" {
		return 0, errors.New("wecom: create department: name is required")
	}
	if department.ID < 0 || department.ParentID < 0 {
		return 0, errors.New("wecom: create department: ids must not be negative")
	}
	var out struct {
		ID int `json:"id"`
	}
	err := c.call(ctx, "contact create department", "/cgi-bin/department/create", department, &out)
	if err != nil {
		return 0, err
	}
	if out.ID != 0 {
		return out.ID, nil
	}
	return department.ID, nil
}

// UpdateDepartment changes one department; omitted fields stay unchanged.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90206
func (c *Contact) UpdateDepartment(ctx context.Context, department Department) error {
	if department.ID <= 0 {
		return errors.New("wecom: update department: id is required")
	}
	return c.call(ctx, "contact update department", "/cgi-bin/department/update", department, nil)
}

// DeleteDepartment removes one department and moves its members to the parent.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90207
func (c *Contact) DeleteDepartment(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("wecom: delete department: id is required")
	}
	return c.get(ctx, "contact delete department", "/cgi-bin/department/delete",
		url.Values{"id": {itoa(id)}}, nil)
}

// ListDepartments returns departments; id 0 returns the whole tree.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90208
func (c *Contact) ListDepartments(ctx context.Context, id int) ([]Department, error) {
	query := url.Values{}
	if id > 0 {
		query.Set("id", itoa(id))
	}
	var out struct {
		Department []Department `json:"department"`
	}
	err := c.get(ctx, "contact list departments", "/cgi-bin/department/list", query, &out)
	return out.Department, err
}

// ListSubDepartmentIDs returns the id list below one department.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95350
func (c *Contact) ListSubDepartmentIDs(ctx context.Context, id int) ([]Department, error) {
	query := url.Values{}
	if id > 0 {
		query.Set("id", itoa(id))
	}
	var out struct {
		DepartmentID []Department `json:"department_id"`
	}
	err := c.get(ctx, "contact list sub department ids", "/cgi-bin/department/simplelist", query, &out)
	return out.DepartmentID, err
}

// GetDepartment reads one department.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95351
func (c *Contact) GetDepartment(ctx context.Context, id int) (Department, error) {
	if id <= 0 {
		return Department{}, errors.New("wecom: get department: id is required")
	}
	var out struct {
		Department Department `json:"department"`
	}
	err := c.get(ctx, "contact get department", "/cgi-bin/department/get",
		url.Values{"id": {itoa(id)}}, &out)
	return out.Department, err
}
