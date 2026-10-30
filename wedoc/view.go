package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	addViewPath     = smartsheetPath + "/add_view"
	updateViewPath  = smartsheetPath + "/update_view"
	deleteViewsPath = smartsheetPath + "/delete_views"
	getViewsPath    = smartsheetPath + "/get_views"
)

// Smart sheet view types.
const (
	ViewTypeGrid     = "VIEW_TYPE_GRID"
	ViewTypeKanban   = "VIEW_TYPE_KANBAN"
	ViewTypeGallery  = "VIEW_TYPE_GALLERY"
	ViewTypeGantt    = "VIEW_TYPE_GANTT"
	ViewTypeCalendar = "VIEW_TYPE_CALENDAR"
)

// View describes one smart sheet view. Property carries the view's
// sort/filter/group/color configuration; GanttProperty and CalendarProperty
// are only used when adding those view types.
type View struct {
	ViewID           string         `json:"view_id,omitempty"`
	ViewTitle        string         `json:"view_title,omitempty"`
	ViewType         string         `json:"view_type,omitempty"`
	Property         map[string]any `json:"property,omitempty"`
	GanttProperty    map[string]any `json:"-"`
	CalendarProperty map[string]any `json:"-"`
}

// AddView creates one view in a sub-table.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99900
func (c *WeDoc) AddView(ctx context.Context, docID, sheetID string, view View) (View, error) {
	docID, err := required("add view", "docId", docID)
	if err != nil {
		return View{}, err
	}
	sheetID, err = required("add view", "sheetId", sheetID)
	if err != nil {
		return View{}, err
	}
	title, err := required("add view", "viewTitle", view.ViewTitle)
	if err != nil {
		return View{}, err
	}
	viewType, err := required("add view", "viewType", view.ViewType)
	if err != nil {
		return View{}, err
	}
	request := struct {
		DocID            string         `json:"docid"`
		SheetID          string         `json:"sheet_id"`
		ViewTitle        string         `json:"view_title"`
		ViewType         string         `json:"view_type"`
		GanttProperty    map[string]any `json:"property_gantt,omitempty"`
		CalendarProperty map[string]any `json:"property_calendar,omitempty"`
	}{
		DocID:            docID,
		SheetID:          sheetID,
		ViewTitle:        title,
		ViewType:         viewType,
		GanttProperty:    view.GanttProperty,
		CalendarProperty: view.CalendarProperty,
	}
	var response struct {
		View View `json:"view"`
	}
	if err := c.call(ctx, "add view", addViewPath, request, &response); err != nil {
		return View{}, err
	}
	return response.View, nil
}

// UpdateView updates a view's title or property configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99902
func (c *WeDoc) UpdateView(ctx context.Context, docID, sheetID string, view View) (View, error) {
	docID, err := required("update view", "docId", docID)
	if err != nil {
		return View{}, err
	}
	sheetID, err = required("update view", "sheetId", sheetID)
	if err != nil {
		return View{}, err
	}
	viewID, err := required("update view", "viewId", view.ViewID)
	if err != nil {
		return View{}, err
	}
	request := struct {
		DocID     string         `json:"docid"`
		SheetID   string         `json:"sheet_id"`
		ViewID    string         `json:"view_id"`
		ViewTitle string         `json:"view_title,omitempty"`
		Property  map[string]any `json:"property,omitempty"`
	}{
		DocID:     docID,
		SheetID:   sheetID,
		ViewID:    viewID,
		ViewTitle: strings.TrimSpace(view.ViewTitle),
		Property:  view.Property,
	}
	var response struct {
		View View `json:"view"`
	}
	if err := c.call(ctx, "update view", updateViewPath, request, &response); err != nil {
		return View{}, err
	}
	return response.View, nil
}

// DeleteViews deletes one or more views from a sub-table.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99901
func (c *WeDoc) DeleteViews(ctx context.Context, docID, sheetID string, viewIDs ...string) error {
	docID, err := required("delete views", "docId", docID)
	if err != nil {
		return err
	}
	sheetID, err = required("delete views", "sheetId", sheetID)
	if err != nil {
		return err
	}
	if len(viewIDs) == 0 {
		return errors.New("wecom: delete views: at least one viewId is required")
	}
	ids := make([]string, 0, len(viewIDs))
	for index, viewID := range viewIDs {
		trimmed := strings.TrimSpace(viewID)
		if trimmed == "" {
			return fmt.Errorf("wecom: delete views: viewIds[%d] is required", index)
		}
		ids = append(ids, trimmed)
	}
	request := struct {
		DocID   string   `json:"docid"`
		SheetID string   `json:"sheet_id"`
		ViewIDs []string `json:"view_ids"`
	}{DocID: docID, SheetID: sheetID, ViewIDs: ids}
	return c.call(ctx, "delete views", deleteViewsPath, request, nil)
}

// GetViews returns every view of a sub-table, walking offset/limit pages.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101155
func (c *WeDoc) GetViews(ctx context.Context, docID, sheetID string) ([]View, error) {
	docID, err := required("get views", "docId", docID)
	if err != nil {
		return nil, err
	}
	sheetID, err = required("get views", "sheetId", sheetID)
	if err != nil {
		return nil, err
	}
	views := make([]View, 0)
	offset := 0
	for {
		request := struct {
			DocID   string `json:"docid"`
			SheetID string `json:"sheet_id"`
			Offset  int    `json:"offset"`
			Limit   int    `json:"limit"`
		}{DocID: docID, SheetID: sheetID, Offset: offset, Limit: maxListPageSize}
		var page struct {
			Total   uint32 `json:"total"`
			HasMore bool   `json:"has_more"`
			Next    uint32 `json:"next"`
			Views   []View `json:"views"`
		}
		if err := c.call(ctx, "get views", getViewsPath, request, &page); err != nil {
			return nil, err
		}
		views = append(views, page.Views...)
		if len(page.Views) == 0 {
			return views, nil
		}
		next := offset + len(page.Views)
		if page.HasMore || page.Next != 0 {
			if int(page.Next) > offset {
				next = int(page.Next)
			}
		} else if page.Total == 0 || next >= int(page.Total) {
			return views, nil
		}
		if next <= offset {
			return nil, errors.New("wecom: get views pagination did not advance")
		}
		offset = next
	}
}
