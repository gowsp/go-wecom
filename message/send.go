package message

import "context"

// SendText sends one plain-text application message to touser, which may be a
// user id, several ids joined by "|", or "@all".
//
// 文档: https://developer.work.weixin.qq.com/document/path/90236
func (c *Message) SendText(ctx context.Context, toUser, content string) error {
	return c.Send(ctx, toUser, AppMessage{Text: &Text{Content: content}})
}

// SendMarkdown sends one markdown application message to touser, which may be a
// user id, several ids joined by "|", or "@all".
//
// 文档: https://developer.work.weixin.qq.com/document/path/90236
func (c *Message) SendMarkdown(ctx context.Context, toUser, content string) error {
	return c.Send(ctx, toUser, AppMessage{Markdown: &Text{Content: content}})
}
