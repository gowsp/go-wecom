// Package webhook implements the WeCom group message push webhook
// (/cgi-bin/webhook/*). A webhook is addressed by its key and does not use a
// corp access token.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the public WeCom API origin.
const DefaultBaseURL = "https://qyapi.weixin.qq.com"

const defaultTimeout = 15 * time.Second

// Webhook posts messages to one group robot.
type Webhook struct {
	key     string
	baseURL string
	client  *http.Client
}

// Option customizes a Webhook.
type Option func(*Webhook)

// WithBaseURL overrides the API origin, for example in tests.
func WithBaseURL(baseURL string) Option {
	return func(w *Webhook) {
		if trimmed := strings.TrimSpace(baseURL); trimmed != "" {
			w.baseURL = strings.TrimRight(trimmed, "/")
		}
	}
}

// WithHTTPClient overrides the HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(w *Webhook) {
		if client != nil {
			w.client = client
		}
	}
}

// New creates a webhook for one group robot key.
func New(key string, options ...Option) *Webhook {
	w := &Webhook{key: strings.TrimSpace(key), baseURL: DefaultBaseURL, client: &http.Client{Timeout: defaultTimeout}}
	for _, option := range options {
		if option != nil {
			option(w)
		}
	}
	return w
}

// Message is the body of one group robot message. Only the field matching
// MsgType is sent.
type Message struct {
	MsgType      string        `json:"msgtype"`
	Text         *Text         `json:"text,omitempty"`
	Markdown     *Text         `json:"markdown,omitempty"`
	Image        *Image        `json:"image,omitempty"`
	News         *News         `json:"news,omitempty"`
	File         *File         `json:"file,omitempty"`
	TemplateCard *TemplateCard `json:"template_card,omitempty"`
}

// Text is a plain text or markdown body.
type Text struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// Image references an uploaded image media id.
type Image struct {
	Base64 string `json:"base64"`
	MD5    string `json:"md5"`
}

// News is a card-style article list.
type News struct {
	Articles []Article `json:"articles"`
}

// Article is one news entry.
type Article struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

// File references an uploaded file media id.
type File struct {
	MediaID string `json:"media_id"`
}

// TemplateCard is the interactive card body.
type TemplateCard struct {
	CardType              string              `json:"card_type"`
	Source                *Source             `json:"source,omitempty"`
	MainTitle             *MainTitle          `json:"main_title,omitempty"`
	EmphasisContent       *EmphasisContent    `json:"emphasis_content,omitempty"`
	SubTitleText          string              `json:"sub_title_text,omitempty"`
	HorizontalContentList []HorizontalContent `json:"horizontal_content_list,omitempty"`
	JumpList              []Jump              `json:"jump_list,omitempty"`
	CardAction            *CardAction         `json:"card_action,omitempty"`
}

// Source describes the card source.
type Source struct {
	IconURL   string `json:"icon_url,omitempty"`
	Desc      string `json:"desc,omitempty"`
	DescColor int    `json:"desc_color,omitempty"`
}

// MainTitle is the card title.
type MainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// EmphasisContent is the highlighted number of a card.
type EmphasisContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc,omitempty"`
}

// HorizontalContent is one key-value row.
type HorizontalContent struct {
	KeyName string `json:"keyname"`
	Value   string `json:"value,omitempty"`
}

// Jump is one link of a card.
type Jump struct {
	Type  int    `json:"type"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title"`
}

// CardAction is the click target of the whole card.
type CardAction struct {
	Type int    `json:"type"`
	URL  string `json:"url,omitempty"`
}

// Send delivers one message to the group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99110
func (w *Webhook) Send(ctx context.Context, message Message) error {
	if w == nil || strings.TrimSpace(w.key) == "" {
		return errors.New("wecom: webhook key is required")
	}
	if strings.TrimSpace(message.MsgType) == "" {
		return errors.New("wecom: webhook message: msgtype is required")
	}
	body, err := json.Marshal(message)
	if err != nil {
		return errors.New("wecom: webhook message: encode failed")
	}
	_, err = w.post(ctx, "/cgi-bin/webhook/send", nil, body, "application/json")
	return err
}

// UploadMedia stores one temporary material for a webhook and returns its media
// id. The image and file message types require it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/99110
func (w *Webhook) UploadMedia(ctx context.Context, mediaType, name string, data []byte) (string, error) {
	if w == nil || strings.TrimSpace(w.key) == "" {
		return "", errors.New("wecom: webhook key is required")
	}
	if len(data) == 0 {
		return "", errors.New("wecom: webhook upload_media: content is empty")
	}
	query := url.Values{"type": {mediaType}}
	query.Set("key", w.key)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("media", name)
	if err != nil {
		return "", errors.New("wecom: webhook upload_media: encode failed")
	}
	if _, err := part.Write(data); err != nil {
		return "", errors.New("wecom: webhook upload_media: encode failed")
	}
	if err := writer.Close(); err != nil {
		return "", errors.New("wecom: webhook upload_media: encode failed")
	}

	raw, err := w.post(ctx, "/cgi-bin/webhook/upload_media", query, body.Bytes(), writer.FormDataContentType())
	if err != nil {
		return "", err
	}
	var out struct {
		MediaID string `json:"media_id"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", errors.New("wecom: webhook upload_media: decode failed")
	}
	return out.MediaID, nil
}

func (w *Webhook) post(ctx context.Context, path string, query url.Values, body []byte, contentType string) ([]byte, error) {
	values := url.Values{}
	for key, entries := range query {
		for _, entry := range entries {
			values.Add(key, entry)
		}
	}
	values.Set("key", w.key)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.baseURL+path+"?"+values.Encode(), bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("wecom: webhook request failed")
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wecom: webhook %s failed", path)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("wecom: webhook %s failed", path)
	}
	var envelope struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.Errcode != 0 {
		return nil, fmt.Errorf("wecom: webhook %s rejected with errcode %d", path, envelope.Errcode)
	}
	return raw, nil
}
