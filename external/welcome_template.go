package external

import (
	"context"
	"errors"
	"strings"
)

// WelcomeTemplate is the reusable greeting for customers joining a group chat.
type WelcomeTemplate struct {
	TemplateID  string                 `json:"template_id,omitempty"`
	Text        *TextContent           `json:"text,omitempty"`
	Image       *ImageAttachment       `json:"image,omitempty"`
	Link        *LinkAttachment        `json:"link,omitempty"`
	MiniProgram *MiniProgramAttachment `json:"miniprogram,omitempty"`
	File        *MediaAttachment       `json:"file,omitempty"`
	Video       *MediaAttachment       `json:"video,omitempty"`
	AgentID     int                    `json:"agentid,omitempty"`
	Notify      int                    `json:"notify,omitempty"`
}

// AddWelcomeTemplate creates one group welcome template and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92366
func (c *External) AddWelcomeTemplate(ctx context.Context, template WelcomeTemplate) (string, error) {
	if err := template.validate(); err != nil {
		return "", err
	}
	var out struct {
		TemplateID string `json:"template_id"`
	}
	err := c.call(ctx, "external add welcome template",
		"/cgi-bin/externalcontact/group_welcome_template/add", template, &out)
	return out.TemplateID, err
}

// UpdateWelcomeTemplate replaces one group welcome template.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92366
func (c *External) UpdateWelcomeTemplate(ctx context.Context, template WelcomeTemplate) error {
	if strings.TrimSpace(template.TemplateID) == "" {
		return errors.New("wecom: update welcome template: template_id is required")
	}
	if err := template.validate(); err != nil {
		return err
	}
	return c.call(ctx, "external update welcome template",
		"/cgi-bin/externalcontact/group_welcome_template/edit", template, nil)
}

// GetWelcomeTemplate reads one group welcome template.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92366
func (c *External) GetWelcomeTemplate(ctx context.Context, templateID string) (WelcomeTemplate, error) {
	if strings.TrimSpace(templateID) == "" {
		return WelcomeTemplate{}, errors.New("wecom: get welcome template: template_id is required")
	}
	request := struct {
		TemplateID string `json:"template_id"`
	}{TemplateID: templateID}
	// The endpoint returns the text/image/link blocks at the top level, so the
	// whole response body decodes into the template.
	var out WelcomeTemplate
	err := c.call(ctx, "external get welcome template",
		"/cgi-bin/externalcontact/group_welcome_template/get", request, &out)
	if err != nil {
		return WelcomeTemplate{}, err
	}
	out.TemplateID = templateID
	return out, nil
}

// DeleteWelcomeTemplate removes one group welcome template. agentid is required
// when templates belong to several applications.
//
// 文档: https://developer.work.weixin.qq.com/document/path/92366
func (c *External) DeleteWelcomeTemplate(ctx context.Context, templateID string, agentID int) error {
	if strings.TrimSpace(templateID) == "" {
		return errors.New("wecom: delete welcome template: template_id is required")
	}
	request := struct {
		TemplateID string `json:"template_id"`
		AgentID    int    `json:"agentid,omitempty"`
	}{TemplateID: templateID, AgentID: agentID}
	return c.call(ctx, "external delete welcome template",
		"/cgi-bin/externalcontact/group_welcome_template/del", request, nil)
}

func (t WelcomeTemplate) validate() error {
	if t.Text == nil && t.Image == nil && t.Link == nil && t.MiniProgram == nil &&
		t.File == nil && t.Video == nil {
		return errors.New("wecom: welcome template: at least one content block is required")
	}
	return nil
}
