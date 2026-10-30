// Package message implements the WeCom application message API.
package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Message types accepted by /cgi-bin/message/send.
const (
	TypeText              = "text"
	TypeImage             = "image"
	TypeVoice             = "voice"
	TypeVideo             = "video"
	TypeFile              = "file"
	TypeTextCard          = "textcard"
	TypeNews              = "news"
	TypeMPNews            = "mpnews"
	TypeMarkdown          = "markdown"
	TypeMiniProgramNotice = "miniprogram_notice"
	TypeTemplateCard      = "template_card"
)

// Text is the text or markdown body.
type Text struct {
	Content string `json:"content"`
}

// Media is the image, voice or file body. MediaID comes from the media upload
// API (api.Media.Upload).
type Media struct {
	MediaID string `json:"media_id"`
}

// Video is the video body.
type Video struct {
	MediaID     string `json:"media_id"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// TextCard is the textcard body.
type TextCard struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	BtnTxt      string `json:"btntxt,omitempty"`
}

// NewsArticle is one article of a news message.
type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
	PicURL      string `json:"picurl,omitempty"`
}

// News is the news body; the official limit is eight articles.
type News struct {
	Articles []NewsArticle `json:"articles"`
}

// MPArticle is one article of an mpnews message.
type MPArticle struct {
	Title            string `json:"title"`
	ThumbMediaID     string `json:"thumb_media_id"`
	Author           string `json:"author,omitempty"`
	ContentSourceURL string `json:"content_source_url,omitempty"`
	Content          string `json:"content"`
	Digest           string `json:"digest,omitempty"`
	ShowCoverPic     int    `json:"show_cover_pic,omitempty"`
}

// MPNews is the mpnews body; the official limit is eight articles.
type MPNews struct {
	Articles []MPArticle `json:"articles"`
}

// MiniProgramContentItem is one row of a miniprogram_notice message.
type MiniProgramContentItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// MiniProgramNotice is the miniprogram_notice body.
type MiniProgramNotice struct {
	AppID             string                   `json:"appid"`
	Page              string                   `json:"page,omitempty"`
	Title             string                   `json:"title"`
	Description       string                   `json:"description,omitempty"`
	EmphasisFirstItem bool                     `json:"emphasis_first_item,omitempty"`
	ContentItem       []MiniProgramContentItem `json:"content_item,omitempty"`
}

// AppMessage is one application message payload. Set exactly one body field; Type is
// inferred from it, or set explicitly when a payload carries several echoes of
// the same data.
type AppMessage struct {
	// Type is the msgtype sent to WeCom. Leave empty to infer it.
	Type string `json:"-"`

	// Safe marks text-like messages as confidential.
	Safe int `json:"safe,omitempty"`
	// EnableIDTrans enables id translation for text-like messages.
	EnableIDTrans int `json:"enable_id_trans,omitempty"`
	// EnableDuplicateCheck enables server-side de-duplication.
	EnableDuplicateCheck int `json:"enable_duplicate_check,omitempty"`
	// DuplicateCheckInterval is the de-duplication window in seconds.
	DuplicateCheckInterval int `json:"duplicate_check_interval,omitempty"`

	Text              *Text              `json:"text,omitempty"`
	Image             *Media             `json:"image,omitempty"`
	Voice             *Media             `json:"voice,omitempty"`
	Video             *Video             `json:"video,omitempty"`
	File              *Media             `json:"file,omitempty"`
	TextCard          *TextCard          `json:"textcard,omitempty"`
	News              *News              `json:"news,omitempty"`
	MPNews            *MPNews            `json:"mpnews,omitempty"`
	Markdown          *Text              `json:"markdown,omitempty"`
	MiniProgramNotice *MiniProgramNotice `json:"miniprogram_notice,omitempty"`
	TemplateCard      json.RawMessage    `json:"template_card,omitempty"`
}

// body validates the message and reports its msgtype.
func (m AppMessage) body() (string, error) {
	type candidate struct {
		name  string
		count int
	}
	candidates := []candidate{
		{TypeText, boolInt(m.Text != nil)},
		{TypeImage, boolInt(m.Image != nil)},
		{TypeVoice, boolInt(m.Voice != nil)},
		{TypeVideo, boolInt(m.Video != nil)},
		{TypeFile, boolInt(m.File != nil)},
		{TypeTextCard, boolInt(m.TextCard != nil)},
		{TypeNews, boolInt(m.News != nil)},
		{TypeMPNews, boolInt(m.MPNews != nil)},
		{TypeMarkdown, boolInt(m.Markdown != nil)},
		{TypeMiniProgramNotice, boolInt(m.MiniProgramNotice != nil)},
		{TypeTemplateCard, boolInt(len(m.TemplateCard) > 0)},
	}

	set := make([]string, 0, len(candidates))
	for _, c := range candidates {
		if c.count > 0 {
			set = append(set, c.name)
		}
	}
	switch {
	case m.Type == "" && len(set) == 0:
		return "", fmt.Errorf("wecom: message body is required")
	case m.Type == "" && len(set) > 1:
		return "", fmt.Errorf("wecom: message has several bodies (%s); set Type",
			strings.Join(set, ", "))
	case m.Type == "":
		m.Type = set[0]
	case !contains(set, m.Type):
		return "", fmt.Errorf("wecom: message type %q does not match its body", m.Type)
	}

	switch m.Type {
	case TypeText:
		return m.Type, requireText("text", m.Text)
	case TypeMarkdown:
		return m.Type, requireText("markdown", m.Markdown)
	case TypeImage:
		return m.Type, requireMedia("image", m.Image)
	case TypeVoice:
		return m.Type, requireMedia("voice", m.Voice)
	case TypeFile:
		return m.Type, requireMedia("file", m.File)
	case TypeVideo:
		if m.Video == nil || strings.TrimSpace(m.Video.MediaID) == "" {
			return "", fmt.Errorf("wecom: video.media_id is required")
		}
	case TypeTextCard:
		if m.TextCard == nil {
			return "", fmt.Errorf("wecom: textcard body is required")
		}
		if strings.TrimSpace(m.TextCard.Title) == "" || strings.TrimSpace(m.TextCard.URL) == "" {
			return "", fmt.Errorf("wecom: textcard.title and textcard.url are required")
		}
	case TypeNews:
		if m.News == nil || len(m.News.Articles) == 0 || len(m.News.Articles) > 8 {
			return "", fmt.Errorf("wecom: news needs 1 to 8 articles")
		}
	case TypeMPNews:
		if m.MPNews == nil || len(m.MPNews.Articles) == 0 || len(m.MPNews.Articles) > 8 {
			return "", fmt.Errorf("wecom: mpnews needs 1 to 8 articles")
		}
	case TypeMiniProgramNotice:
		if m.MiniProgramNotice == nil {
			return "", fmt.Errorf("wecom: miniprogram_notice body is required")
		}
		notice := m.MiniProgramNotice
		if strings.TrimSpace(notice.AppID) == "" || strings.TrimSpace(notice.Title) == "" {
			return "", fmt.Errorf("wecom: miniprogram_notice.appid and title are required")
		}
	}
	return m.Type, nil
}

func requireText(name string, t *Text) error {
	if t == nil || strings.TrimSpace(t.Content) == "" {
		return fmt.Errorf("wecom: %s.content is required", name)
	}
	return nil
}

func requireMedia(name string, m *Media) error {
	if m == nil || strings.TrimSpace(m.MediaID) == "" {
		return fmt.Errorf("wecom: %s.media_id is required", name)
	}
	return nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
