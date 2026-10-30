package wedoc

import (
	"context"
	"encoding/json"
)

const (
	documentBatchUpdatePath = "/cgi-bin/wedoc/document/batch_update"
	documentGetPath         = "/cgi-bin/wedoc/document/get"
)

// DocumentData is the content of a document. The nested structure differs per
// document type, so it is returned as raw JSON.
type DocumentData struct {
	Version  int             `json:"version"`
	Document json.RawMessage `json:"document"`
}

// BatchUpdateDocument applies a list of edit requests to a document. Each entry
// keeps the WeCom request shape, for example {"insert_text": {...}}; version is
// the document version returned by GetDocument and 0 lets the server decide.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97626
func (c *WeDoc) BatchUpdateDocument(ctx context.Context, docID string, version int, requests []map[string]any) error {
	docID, err := required("batch update document", "docId", docID)
	if err != nil {
		return err
	}
	if len(requests) == 0 {
		return errEmpty("batch update document", "requests")
	}
	body := struct {
		DocID    string           `json:"docid"`
		Version  int              `json:"verison,omitempty"`
		Requests []map[string]any `json:"requests"`
	}{DocID: docID, Version: version, Requests: requests}
	return c.call(ctx, "batch update document", documentBatchUpdatePath, body, nil)
}

// GetDocument returns the content and version of a document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/101161
func (c *WeDoc) GetDocument(ctx context.Context, docID string) (DocumentData, error) {
	docID, err := required("get document", "docId", docID)
	if err != nil {
		return DocumentData{}, err
	}
	var out DocumentData
	err = c.call(ctx, "get document", documentGetPath, map[string]string{"docid": docID}, &out)
	return out, err
}
