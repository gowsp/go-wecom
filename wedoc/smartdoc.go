package wedoc

import (
	"context"
	"encoding/json"
)

const (
	smartdocPageHierarchyPath  = "/cgi-bin/wedoc/smartdoc/get_page_hierarchy"
	smartdocAddPagePath        = "/cgi-bin/wedoc/smartdoc/add_page"
	smartdocUpdatePagePath     = "/cgi-bin/wedoc/smartdoc/update_page"
	smartdocDeletePagePath     = "/cgi-bin/wedoc/smartdoc/delete_page"
	smartdocAddBlocksPath      = "/cgi-bin/wedoc/smartdoc/add_blocks"
	smartdocUpdateBlocksPath   = "/cgi-bin/wedoc/smartdoc/update_blocks"
	smartdocDeleteBlocksPath   = "/cgi-bin/wedoc/smartdoc/delete_blocks"
	smartdocBlockListPath      = "/cgi-bin/wedoc/smartdoc/get_block_list"
	smartdocExportTaskPath     = "/cgi-bin/wedoc/smartdoc/export_task"
	smartdocExportResultPath   = "/cgi-bin/wedoc/smartdoc/get_export_result"
	smartdocDatasourcePath     = "/cgi-bin/wedoc/smartdoc/get_smartsheet_info"
	smartdocAddDatasheetPath   = "/cgi-bin/wedoc/smartdoc/add_smartsheet"
	smartdocDelDatasheetPath   = "/cgi-bin/wedoc/smartdoc/delete_smartsheet"
	smartdocUpdDatasheetPath   = "/cgi-bin/wedoc/smartdoc/update_smartsheet"
	smartdocPublishPath        = "/cgi-bin/wedoc/smartdoc/publish"
	smartdocCancelPublishPath  = "/cgi-bin/wedoc/smartdoc/cancel_publish"
	smartdocPublishSettingPath = "/cgi-bin/wedoc/smartdoc/publish_setting"
)

// Page is one page of a smart document.
type Page struct {
	PageID     string `json:"page_id,omitempty"`
	Title      string `json:"title,omitempty"`
	ParentID   string `json:"parent_id,omitempty"`
	AfterID    string `json:"after_id,omitempty"`
	LayoutMode int    `json:"layout_mode,omitempty"`
}

// Block is one content block of a smart document page.
type Block struct {
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Title    string          `json:"title,omitempty"`
	Content  string          `json:"content,omitempty"`
	ParentID string          `json:"parent_id,omitempty"`
	AfterID  string          `json:"after_id,omitempty"`
	Children []string        `json:"children,omitempty"`
	Props    json.RawMessage `json:"props,omitempty"`
}

// BlockList is one page of content blocks.
type BlockList struct {
	Blocks    []Block `json:"blocks"`
	HasMore   string  `json:"has_more"`
	NextStart int     `json:"next_start"`
}

// Datasheet is a data table embedded in a smart document.
type Datasheet struct {
	BlockID string `json:"block_id,omitempty"`
	Title   string `json:"title,omitempty"`
	SheetID string `json:"sheet_id,omitempty"`
	DocID   string `json:"ss_docid,omitempty"`
	AfterID string `json:"after_id,omitempty"`
}

// PublishResult describes the public view of a published smart document.
type PublishResult struct {
	ShareCode   string `json:"share_code"`
	PublishURL  string `json:"publish_url"`
	Version     int    `json:"version"`
	PublishTime int64  `json:"publish_time"`
	Title       string `json:"publish_doc_title"`
}

// AuthEntry selects who may view a published document.
type AuthEntry struct {
	Type         int    `json:"type"`
	UserID       string `json:"userid,omitempty"`
	DepartmentID int64  `json:"departmentid,omitempty"`
}

// PageHierarchy returns every page of a smart document with its parent link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101619
func (c *WeDoc) PageHierarchy(ctx context.Context, docID string) ([]Page, error) {
	docID, err := required("get page hierarchy", "docId", docID)
	if err != nil {
		return nil, err
	}
	var out struct {
		Pages []Page `json:"pages"`
	}
	err = c.call(ctx, "get page hierarchy", smartdocPageHierarchyPath, map[string]string{"docid": docID}, &out)
	return out.Pages, err
}

// AddPage inserts one page and returns it with the assigned id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101620
func (c *WeDoc) AddPage(ctx context.Context, docID string, page Page) (Page, error) {
	docID, err := required("add page", "docId", docID)
	if err != nil {
		return Page{}, err
	}
	body := struct {
		DocID string `json:"docid"`
		Info  Page   `json:"info"`
	}{DocID: docID, Info: page}
	var out struct {
		Info Page `json:"info"`
	}
	err = c.call(ctx, "add page", smartdocAddPagePath, body, &out)
	return out.Info, err
}

// UpdatePage renames or moves one page.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101621
func (c *WeDoc) UpdatePage(ctx context.Context, docID string, page Page) (Page, error) {
	docID, err := required("update page", "docId", docID)
	if err != nil {
		return Page{}, err
	}
	if _, err := required("update page", "page_id", page.PageID); err != nil {
		return Page{}, err
	}
	body := struct {
		DocID string `json:"docid"`
		Info  Page   `json:"info"`
	}{DocID: docID, Info: page}
	var out struct {
		Info Page `json:"info"`
	}
	err = c.call(ctx, "update page", smartdocUpdatePagePath, body, &out)
	return out.Info, err
}

// DeletePage removes one page and everything below it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101622
func (c *WeDoc) DeletePage(ctx context.Context, docID, pageID string) error {
	docID, err := required("delete page", "docId", docID)
	if err != nil {
		return err
	}
	if pageID, err = required("delete page", "page_id", pageID); err != nil {
		return err
	}
	return c.call(ctx, "delete page", smartdocDeletePagePath,
		map[string]string{"docid": docID, "page_id": pageID}, nil)
}

// AddBlocks appends content blocks to one page and returns them with ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101623
func (c *WeDoc) AddBlocks(ctx context.Context, docID, pageID string, blocks []Block) ([]Block, error) {
	if err := requireDocPage("add blocks", docID, pageID); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, errEmpty("add blocks", "blocks")
	}
	body := struct {
		DocID  string  `json:"docid"`
		PageID string  `json:"page_id"`
		Blocks []Block `json:"blocks"`
	}{DocID: docID, PageID: pageID, Blocks: blocks}
	var out struct {
		Blocks []Block `json:"blocks"`
	}
	err := c.call(ctx, "add blocks", smartdocAddBlocksPath, body, &out)
	return out.Blocks, err
}

// UpdateBlocks replaces the content of existing blocks.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101624
func (c *WeDoc) UpdateBlocks(ctx context.Context, docID, pageID string, blocks []Block) ([]Block, error) {
	if err := requireDocPage("update blocks", docID, pageID); err != nil {
		return nil, err
	}
	if len(blocks) == 0 {
		return nil, errEmpty("update blocks", "blocks")
	}
	body := struct {
		DocID  string  `json:"docid"`
		PageID string  `json:"page_id"`
		Blocks []Block `json:"blocks"`
	}{DocID: docID, PageID: pageID, Blocks: blocks}
	var out struct {
		Blocks []Block `json:"blocks"`
	}
	err := c.call(ctx, "update blocks", smartdocUpdateBlocksPath, body, &out)
	return out.Blocks, err
}

// DeleteBlocks removes content blocks from one page.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101625
func (c *WeDoc) DeleteBlocks(ctx context.Context, docID, pageID string, ids []string) error {
	if err := requireDocPage("delete blocks", docID, pageID); err != nil {
		return err
	}
	if len(ids) == 0 {
		return errEmpty("delete blocks", "ids")
	}
	body := struct {
		DocID  string   `json:"docid"`
		PageID string   `json:"page_id"`
		IDs    []string `json:"ids"`
	}{DocID: docID, PageID: pageID, IDs: ids}
	return c.call(ctx, "delete blocks", smartdocDeleteBlocksPath, body, nil)
}

// BlockList returns content blocks of one page. An empty ids list returns every
// block; limit defaults to 200.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101626
func (c *WeDoc) BlockList(ctx context.Context, docID, pageID string, ids []string, start, limit int) (BlockList, error) {
	if err := requireDocPage("get block list", docID, pageID); err != nil {
		return BlockList{}, err
	}
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	body := struct {
		DocID  string   `json:"docid"`
		PageID string   `json:"page_id"`
		IDs    []string `json:"ids,omitempty"`
		Start  int      `json:"start,omitempty"`
		Limit  int      `json:"limit"`
	}{DocID: docID, PageID: pageID, IDs: ids, Start: start, Limit: limit}
	var out BlockList
	err := c.call(ctx, "get block list", smartdocBlockListPath, body, &out)
	return out, err
}

// ExportDocument starts an asynchronous Markdown export and returns the task
// id. An empty pageID exports the whole document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101627
func (c *WeDoc) ExportDocument(ctx context.Context, docID, pageID string) (string, error) {
	docID, err := required("export document", "docId", docID)
	if err != nil {
		return "", err
	}
	body := struct {
		DocID       string `json:"docid"`
		ContentType int    `json:"content_type"`
		PageID      string `json:"page_id,omitempty"`
	}{DocID: docID, ContentType: 1, PageID: pageID}
	var out struct {
		TaskID string `json:"task_id"`
	}
	err = c.call(ctx, "export document", smartdocExportTaskPath, body, &out)
	return out.TaskID, err
}

// ExportResult polls one export task. The body keeps the WeCom shape because the
// finished payload carries the Markdown content.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101627
func (c *WeDoc) ExportResult(ctx context.Context, taskID string) (json.RawMessage, error) {
	taskID, err := required("get export result", "task_id", taskID)
	if err != nil {
		return nil, err
	}
	var out json.RawMessage
	err = c.call(ctx, "get export result", smartdocExportResultPath, map[string]string{"task_id": taskID}, &out)
	return out, err
}

// Datasource returns the smart sheet behind a smart document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101628
func (c *WeDoc) Datasource(ctx context.Context, docID string) (string, error) {
	docID, err := required("get datasource", "docId", docID)
	if err != nil {
		return "", err
	}
	var out struct {
		DocID string `json:"ss_docid"`
	}
	err = c.call(ctx, "get datasource", smartdocDatasourcePath, map[string]string{"docid": docID}, &out)
	return out.DocID, err
}

// AddDatasheet inserts a data table into a smart document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101629
func (c *WeDoc) AddDatasheet(ctx context.Context, docID string, sheet Datasheet) (Datasheet, error) {
	docID, err := required("add datasheet", "docId", docID)
	if err != nil {
		return Datasheet{}, err
	}
	body := struct {
		DocID string    `json:"docid"`
		Info  Datasheet `json:"info"`
	}{DocID: docID, Info: sheet}
	var out struct {
		Info Datasheet `json:"info"`
	}
	err = c.call(ctx, "add datasheet", smartdocAddDatasheetPath, body, &out)
	return out.Info, err
}

// DeleteDatasheet removes a data table from a smart document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101630
func (c *WeDoc) DeleteDatasheet(ctx context.Context, docID, blockID string) error {
	docID, err := required("delete datasheet", "docId", docID)
	if err != nil {
		return err
	}
	if blockID, err = required("delete datasheet", "block_id", blockID); err != nil {
		return err
	}
	return c.call(ctx, "delete datasheet", smartdocDelDatasheetPath,
		map[string]string{"docid": docID, "block_id": blockID}, nil)
}

// UpdateDatasheet renames or moves a data table inside a smart document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101631
func (c *WeDoc) UpdateDatasheet(ctx context.Context, docID string, sheet Datasheet) (Datasheet, error) {
	docID, err := required("update datasheet", "docId", docID)
	if err != nil {
		return Datasheet{}, err
	}
	if _, err := required("update datasheet", "block_id", sheet.BlockID); err != nil {
		return Datasheet{}, err
	}
	body := struct {
		DocID string    `json:"docid"`
		Info  Datasheet `json:"info"`
	}{DocID: docID, Info: sheet}
	var out struct {
		Info Datasheet `json:"info"`
	}
	err = c.call(ctx, "update datasheet", smartdocUpdDatasheetPath, body, &out)
	return out.Info, err
}

// PublishDocument publishes a smart document and returns its public URL.
// publishRange selects who may view it; authList narrows the range.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101616
func (c *WeDoc) PublishDocument(ctx context.Context, docID string, publishRange int, authList []AuthEntry) (PublishResult, error) {
	docID, err := required("publish document", "docId", docID)
	if err != nil {
		return PublishResult{}, err
	}
	body := struct {
		DocID        string      `json:"docid"`
		PublishRange int         `json:"publish_range"`
		AuthList     []AuthEntry `json:"auth_list,omitempty"`
	}{DocID: docID, PublishRange: publishRange, AuthList: authList}
	var out PublishResult
	err = c.call(ctx, "publish document", smartdocPublishPath, body, &out)
	return out, err
}

// CancelPublish removes the public view of a smart document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101617
func (c *WeDoc) CancelPublish(ctx context.Context, docID string) error {
	docID, err := required("cancel publish", "docId", docID)
	if err != nil {
		return err
	}
	return c.call(ctx, "cancel publish", smartdocCancelPublishPath, map[string]string{"docid": docID}, nil)
}

// SetPublishSetting changes the visible range of a published document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101618
func (c *WeDoc) SetPublishSetting(ctx context.Context, docID string, publishRange int, authList []AuthEntry) error {
	docID, err := required("set publish setting", "docId", docID)
	if err != nil {
		return err
	}
	body := struct {
		DocID        string      `json:"docid"`
		PublishRange int         `json:"publish_range"`
		AuthList     []AuthEntry `json:"auth_list,omitempty"`
	}{DocID: docID, PublishRange: publishRange, AuthList: authList}
	return c.call(ctx, "set publish setting", smartdocPublishSettingPath, body, nil)
}

func requireDocPage(op, docID, pageID string) error {
	if _, err := required(op, "docId", docID); err != nil {
		return err
	}
	if _, err := required(op, "page_id", pageID); err != nil {
		return err
	}
	return nil
}
