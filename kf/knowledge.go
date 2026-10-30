package kf

import (
	"context"
	"errors"
	"strings"
)

const (
	knowledgeAddGroupPath   = "/cgi-bin/kf/knowledge/add_group"
	knowledgeDelGroupPath   = "/cgi-bin/kf/knowledge/del_group"
	knowledgeModGroupPath   = "/cgi-bin/kf/knowledge/mod_group"
	knowledgeListGroupPath  = "/cgi-bin/kf/knowledge/list_group"
	knowledgeAddIntentPath  = "/cgi-bin/kf/knowledge/add_intent"
	knowledgeDelIntentPath  = "/cgi-bin/kf/knowledge/del_intent"
	knowledgeModIntentPath  = "/cgi-bin/kf/knowledge/mod_intent"
	knowledgeListIntentPath = "/cgi-bin/kf/knowledge/list_intent"
)

// KnowledgeGroup is one knowledge-base group.
type KnowledgeGroup struct {
	GroupID string `json:"group_id"`
	Name    string `json:"name"`
}

// KnowledgeGroups is one page of knowledge-base groups.
type KnowledgeGroups struct {
	NextCursor string           `json:"next_cursor"`
	HasMore    int              `json:"has_more"`
	Groups     []KnowledgeGroup `json:"group_list"`
}

// KnowledgeText wraps the text of a question or answer.
type KnowledgeText struct {
	Content string `json:"content"`
}

// KnowledgeTextBody is the {"text": {...}} envelope used by intents.
type KnowledgeTextBody struct {
	Text KnowledgeText `json:"text"`
}

// KnowledgeQuestions is the {"items": [...]} envelope of similar questions.
type KnowledgeQuestions struct {
	Items []KnowledgeTextBody `json:"items"`
}

// KnowledgeAttachment is one attachment of an answer.
type KnowledgeAttachment struct {
	MsgType string        `json:"msgtype"`
	Image   *MediaContent `json:"image,omitempty"`
}

// KnowledgeAnswer is one answer of a knowledge-base intent.
type KnowledgeAnswer struct {
	Text        KnowledgeText         `json:"text"`
	Attachments []KnowledgeAttachment `json:"attachments,omitempty"`
}

// KnowledgeIntent is one question-and-answer entry.
type KnowledgeIntent struct {
	GroupID          string             `json:"group_id"`
	IntentID         string             `json:"intent_id"`
	Question         KnowledgeTextBody  `json:"question"`
	SimilarQuestions KnowledgeQuestions `json:"similar_questions"`
	Answers          []KnowledgeAnswer  `json:"answers"`
}

// KnowledgeIntents is one page of knowledge-base intents.
type KnowledgeIntents struct {
	NextCursor string            `json:"next_cursor"`
	HasMore    int               `json:"has_more"`
	Intents    []KnowledgeIntent `json:"intent_list"`
}

// AddKnowledgeGroup creates a knowledge-base group and returns its id. The name
// is limited to 12 characters by WeCom.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95971
func (c *KF) AddKnowledgeGroup(ctx context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("wecom: add knowledge group: name is required")
	}
	var out struct {
		GroupID string `json:"group_id"`
	}
	err := c.api.Call(ctx, "add knowledge group", knowledgeAddGroupPath, map[string]string{"name": name}, &out)
	return out.GroupID, err
}

// DeleteKnowledgeGroup removes a knowledge-base group. The system default group
// cannot be deleted.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95971
func (c *KF) DeleteKnowledgeGroup(ctx context.Context, groupID string) error {
	groupID = strings.TrimSpace(groupID)
	if groupID == "" {
		return errors.New("wecom: delete knowledge group: group_id is required")
	}
	return c.api.Call(ctx, "delete knowledge group", knowledgeDelGroupPath, map[string]string{"group_id": groupID}, nil)
}

// UpdateKnowledgeGroup renames a knowledge-base group.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95971
func (c *KF) UpdateKnowledgeGroup(ctx context.Context, groupID, name string) error {
	groupID = strings.TrimSpace(groupID)
	name = strings.TrimSpace(name)
	if groupID == "" || name == "" {
		return errors.New("wecom: update knowledge group: group_id and name are required")
	}
	return c.api.Call(ctx, "update knowledge group", knowledgeModGroupPath,
		map[string]string{"group_id": groupID, "name": name}, nil)
}

// ListKnowledgeGroups returns knowledge-base groups page by page. An empty
// groupID lists every group; limit defaults to 500 and is capped at 1000.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95971
func (c *KF) ListKnowledgeGroups(ctx context.Context, cursor string, limit int, groupID string) (KnowledgeGroups, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	request := struct {
		Cursor  string `json:"cursor,omitempty"`
		Limit   int    `json:"limit"`
		GroupID string `json:"group_id,omitempty"`
	}{Cursor: cursor, Limit: limit, GroupID: groupID}
	var out KnowledgeGroups
	err := c.api.Call(ctx, "list knowledge groups", knowledgeListGroupPath, request, &out)
	return out, err
}

// AddKnowledgeIntent creates a question-and-answer entry and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95972
func (c *KF) AddKnowledgeIntent(ctx context.Context, intent KnowledgeIntent) (string, error) {
	if strings.TrimSpace(intent.GroupID) == "" {
		return "", errors.New("wecom: add knowledge intent: group_id is required")
	}
	if _, err := knowledgeIntentBody("add knowledge intent", intent); err != nil {
		return "", err
	}
	var out struct {
		IntentID string `json:"intent_id"`
	}
	err := c.api.Call(ctx, "add knowledge intent", knowledgeAddIntentPath, intent, &out)
	return out.IntentID, err
}

// DeleteKnowledgeIntent removes one question-and-answer entry.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95972
func (c *KF) DeleteKnowledgeIntent(ctx context.Context, intentID string) error {
	intentID = strings.TrimSpace(intentID)
	if intentID == "" {
		return errors.New("wecom: delete knowledge intent: intent_id is required")
	}
	return c.api.Call(ctx, "delete knowledge intent", knowledgeDelIntentPath, map[string]string{"intent_id": intentID}, nil)
}

// UpdateKnowledgeIntent replaces the provided parts of an entry. Each provided
// part is overwritten completely, so it must carry all of its fields.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95972
func (c *KF) UpdateKnowledgeIntent(ctx context.Context, intent KnowledgeIntent) error {
	intent.IntentID = strings.TrimSpace(intent.IntentID)
	if intent.IntentID == "" {
		return errors.New("wecom: update knowledge intent: intent_id is required")
	}
	if _, err := knowledgeIntentBody("update knowledge intent", intent); err != nil {
		return err
	}
	return c.api.Call(ctx, "update knowledge intent", knowledgeModIntentPath, intent, nil)
}

// ListKnowledgeIntents returns question-and-answer entries page by page. An
// empty groupID or intentID widens the query; limit defaults to 500.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95972
func (c *KF) ListKnowledgeIntents(ctx context.Context, cursor string, limit int, groupID, intentID string) (KnowledgeIntents, error) {
	if limit <= 0 || limit > 1000 {
		limit = 500
	}
	request := struct {
		Cursor   string `json:"cursor,omitempty"`
		Limit    int    `json:"limit"`
		GroupID  string `json:"group_id,omitempty"`
		IntentID string `json:"intent_id,omitempty"`
	}{Cursor: cursor, Limit: limit, GroupID: groupID, IntentID: intentID}
	var out KnowledgeIntents
	err := c.api.Call(ctx, "list knowledge intents", knowledgeListIntentPath, request, &out)
	return out, err
}

func knowledgeIntentBody(op string, intent KnowledgeIntent) (struct{}, error) {
	if strings.TrimSpace(intent.Question.Text.Content) == "" {
		return struct{}{}, errors.New("wecom: " + op + ": question is required")
	}
	if len(intent.Answers) == 0 {
		return struct{}{}, errors.New("wecom: " + op + ": answers are required")
	}
	return struct{}{}, nil
}
