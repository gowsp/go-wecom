package wedoc

import "context"

const docSecurityPath = "/cgi-bin/wedoc/mod_doc_safty_setting"

// Watermark is the visible watermark of a document.
type Watermark struct {
	MarginType      int    `json:"margin_type,omitempty"`
	ShowVisitorName bool   `json:"show_visitor_name,omitempty"`
	ShowText        bool   `json:"show_text,omitempty"`
	Text            string `json:"text,omitempty"`
}

// DocSecurity is the security configuration of a document.
type DocSecurity struct {
	EnableReadonlyCopy bool       `json:"enable_readonly_copy"`
	Watermark          *Watermark `json:"watermark,omitempty"`
}

// SetDocSecurity updates copy protection and watermarking of a document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97782
func (c *WeDoc) SetDocSecurity(ctx context.Context, docID string, security DocSecurity) error {
	docID, err := required("set doc security", "docId", docID)
	if err != nil {
		return err
	}
	body := struct {
		DocID string `json:"docid"`
		DocSecurity
	}{DocID: docID, DocSecurity: security}
	return c.call(ctx, "set doc security", docSecurityPath, body, nil)
}
