package external

import (
	"context"
	"errors"
	"strings"
)

// ProductAttachment is one image of a product album entry.
type ProductAttachment struct {
	Type  string           `json:"type"`
	Image *MediaAttachment `json:"image,omitempty"`
}

// Product is one entry of the product album attached to a member profile.
type Product struct {
	ProductID   string              `json:"product_id,omitempty"`
	Description string              `json:"description"`
	Price       int                 `json:"price"`
	ProductSN   string              `json:"product_sn,omitempty"`
	CreateTime  int64               `json:"create_time,omitempty"`
	Attachments []ProductAttachment `json:"attachments"`
}

// ProductPage is one page of the product album.
type ProductPage struct {
	NextCursor string    `json:"next_cursor"`
	Products   []Product `json:"product_list"`
}

// AddProduct creates one product album entry and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95096
func (c *External) AddProduct(ctx context.Context, product Product) (string, error) {
	if strings.TrimSpace(product.Description) == "" {
		return "", errors.New("wecom: add product: description is required")
	}
	if len(product.Attachments) == 0 {
		return "", errors.New("wecom: add product: at least one attachment is required")
	}
	var out struct {
		ProductID string `json:"product_id"`
	}
	err := c.call(ctx, "external add product",
		"/cgi-bin/externalcontact/add_product_album", product, &out)
	return out.ProductID, err
}

// UpdateProduct replaces one product album entry.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95096
func (c *External) UpdateProduct(ctx context.Context, product Product) error {
	if strings.TrimSpace(product.ProductID) == "" {
		return errors.New("wecom: update product: product_id is required")
	}
	return c.call(ctx, "external update product",
		"/cgi-bin/externalcontact/update_product_album", product, nil)
}

// DeleteProducts removes product album entries.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95096
func (c *External) DeleteProducts(ctx context.Context, productIDs []string) error {
	if len(productIDs) == 0 {
		return errors.New("wecom: delete products: product_id is empty")
	}
	return c.call(ctx, "external delete products",
		"/cgi-bin/externalcontact/delete_product_album",
		map[string][]string{"product_id": productIDs}, nil)
}

// GetProduct reads one product album entry.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95096
func (c *External) GetProduct(ctx context.Context, productID string) (Product, error) {
	if strings.TrimSpace(productID) == "" {
		return Product{}, errors.New("wecom: get product: product_id is required")
	}
	var out struct {
		Product Product `json:"product"`
	}
	err := c.call(ctx, "external get product",
		"/cgi-bin/externalcontact/get_product_album",
		map[string]string{"product_id": productID}, &out)
	return out.Product, err
}

// ListProducts returns one page of the product album.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95096
func (c *External) ListProducts(ctx context.Context, cursor string, limit int) (ProductPage, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out ProductPage
	err := c.call(ctx, "external list products",
		"/cgi-bin/externalcontact/get_product_album_list", request, &out)
	return out, err
}

// InterceptRange is the range an intercept rule applies to.
type InterceptRange struct {
	UserList       []string `json:"user_list,omitempty"`
	DepartmentList []int    `json:"department_list,omitempty"`
}

// InterceptRule is one chat sensitive-word rule.
type InterceptRule struct {
	RuleID          string          `json:"rule_id,omitempty"`
	RuleName        string          `json:"rule_name,omitempty"`
	WordList        []string        `json:"word_list,omitempty"`
	SemanticsList   []int           `json:"semantics_list,omitempty"`
	InterceptType   int             `json:"intercept_type,omitempty"`
	ApplicableRange *InterceptRange `json:"applicable_range,omitempty"`
	CreateTime      int64           `json:"create_time,omitempty"`
}

// AddInterceptRule creates one intercept rule and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95097
func (c *External) AddInterceptRule(ctx context.Context, rule InterceptRule) (string, error) {
	if strings.TrimSpace(rule.RuleName) == "" {
		return "", errors.New("wecom: add intercept rule: rule_name is required")
	}
	if len(rule.WordList) == 0 && len(rule.SemanticsList) == 0 {
		return "", errors.New("wecom: add intercept rule: word_list or semantics_list is required")
	}
	var out struct {
		RuleID string `json:"rule_id"`
	}
	err := c.call(ctx, "external add intercept rule",
		"/cgi-bin/externalcontact/add_intercept_rule", rule, &out)
	return out.RuleID, err
}

// UpdateInterceptRule changes one intercept rule. Use AddRange and RemoveRange
// to adjust the applicable range instead of replacing it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95097
func (c *External) UpdateInterceptRule(ctx context.Context, rule InterceptRule, addRange, removeRange *InterceptRange) error {
	if strings.TrimSpace(rule.RuleID) == "" {
		return errors.New("wecom: update intercept rule: rule_id is required")
	}
	request := struct {
		RuleID                string          `json:"rule_id"`
		RuleName              string          `json:"rule_name,omitempty"`
		WordList              []string        `json:"word_list,omitempty"`
		ExtraRule             *InterceptExtra `json:"extra_rule,omitempty"`
		InterceptType         int             `json:"intercept_type,omitempty"`
		AddApplicableRange    *InterceptRange `json:"add_applicable_range,omitempty"`
		RemoveApplicableRange *InterceptRange `json:"remove_applicable_range,omitempty"`
	}{
		RuleID: rule.RuleID, RuleName: rule.RuleName, WordList: rule.WordList,
		InterceptType:      rule.InterceptType,
		AddApplicableRange: addRange, RemoveApplicableRange: removeRange,
	}
	if len(rule.SemanticsList) > 0 {
		request.ExtraRule = &InterceptExtra{SemanticsList: rule.SemanticsList}
	}
	return c.call(ctx, "external update intercept rule",
		"/cgi-bin/externalcontact/update_intercept_rule", request, nil)
}

// InterceptExtra carries the semantic rule ids of an intercept rule.
type InterceptExtra struct {
	SemanticsList []int `json:"semantics_list,omitempty"`
}

// DeleteInterceptRule removes one intercept rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95097
func (c *External) DeleteInterceptRule(ctx context.Context, ruleID string) error {
	if strings.TrimSpace(ruleID) == "" {
		return errors.New("wecom: delete intercept rule: rule_id is required")
	}
	return c.call(ctx, "external delete intercept rule",
		"/cgi-bin/externalcontact/del_intercept_rule",
		map[string]string{"rule_id": ruleID}, nil)
}

// GetInterceptRule reads one intercept rule.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95097
func (c *External) GetInterceptRule(ctx context.Context, ruleID string) (InterceptRule, error) {
	if strings.TrimSpace(ruleID) == "" {
		return InterceptRule{}, errors.New("wecom: get intercept rule: rule_id is required")
	}
	var out struct {
		Rule InterceptRule `json:"rule"`
	}
	err := c.call(ctx, "external get intercept rule",
		"/cgi-bin/externalcontact/get_intercept_rule",
		map[string]string{"rule_id": ruleID}, &out)
	return out.Rule, err
}

// ListInterceptRules returns every intercept rule of the enterprise.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95097
func (c *External) ListInterceptRules(ctx context.Context) ([]InterceptRule, error) {
	var out struct {
		Rules []InterceptRule `json:"rule_list"`
	}
	err := c.call(ctx, "external list intercept rules",
		"/cgi-bin/externalcontact/get_intercept_rule_list", nil, &out)
	return out.Rules, err
}

// UploadedAttachment is one uploaded temporary resource.
type UploadedAttachment struct {
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt int64  `json:"created_at"`
}

// UploadAttachment stores one attachment used by moments and profiles. mediaType
// is "image", "video" or "file"; the official size limits differ per type.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95098
func (c *External) UploadAttachment(ctx context.Context, mediaType, filename string, data []byte) (UploadedAttachment, error) {
	switch mediaType {
	case "image", "video", "file":
	default:
		return UploadedAttachment{}, errors.New("wecom: upload attachment: unsupported type " + mediaType)
	}
	if len(data) == 0 {
		return UploadedAttachment{}, errors.New("wecom: upload attachment: content is empty")
	}
	if strings.TrimSpace(filename) == "" {
		return UploadedAttachment{}, errors.New("wecom: upload attachment: file name is required")
	}
	var out UploadedAttachment
	err := c.api.Upload(ctx, "external upload attachment", "/cgi-bin/media/upload_attachment",
		map[string][]string{"type": {mediaType}}, "media", filename, data, &out)
	return out, err
}

// ServedContact is one external contact the service has reached.
type ServedContact struct {
	IsCustomer     bool   `json:"is_customer"`
	Name           string `json:"name,omitempty"`
	TmpOpenID      string `json:"tmp_openid,omitempty"`
	ExternalUserID string `json:"external_userid,omitempty"`
	FollowUserID   string `json:"follow_userid"`
	ChatName       string `json:"chat_name,omitempty"`
	AddTime        int64  `json:"add_time"`
}

// ServedContactPage is one page of served contacts.
type ServedContactPage struct {
	NextCursor string          `json:"next_cursor"`
	Info       []ServedContact `json:"info_list"`
}

// ListServedContacts returns the external contacts the enterprise has served,
// including those that never became customers.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99434
func (c *External) ListServedContacts(ctx context.Context, cursor string, limit int) (ServedContactPage, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out ServedContactPage
	err := c.call(ctx, "external list served contacts",
		"/cgi-bin/externalcontact/contact_list", request, &out)
	return out, err
}
