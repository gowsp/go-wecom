package wedoc

import (
	"context"
	"encoding/json"
)

const (
	formCreatePath    = "/cgi-bin/wedoc/create_form"
	formModifyPath    = "/cgi-bin/wedoc/modify_form"
	formInfoPath      = "/cgi-bin/wedoc/get_form_info"
	formStatisticPath = "/cgi-bin/wedoc/get_form_statistic"
	formAnswerPath    = "/cgi-bin/wedoc/get_form_answer"
)

// FormInfo is the definition of a collection form. The question list and the
// settings are passed through as raw JSON because WeCom supports many question
// types.
type FormInfo struct {
	Title     string          `json:"form_title,omitempty"`
	Desc      string          `json:"form_desc,omitempty"`
	Header    string          `json:"form_header,omitempty"`
	Questions json.RawMessage `json:"form_question,omitempty"`
	Setting   json.RawMessage `json:"form_setting,omitempty"`
}

// FormStatistic is the answer count of one collection form.
type FormStatistic struct {
	FillCount     int `json:"fill_cnt"`
	FillUserCount int `json:"fill_user_cnt"`
	UnfillCount   int `json:"unfill_user_cnt"`
}

// CreateForm creates one collection form and returns its id. Both spaceID and
// fatherID select the target folder and may be empty for the default folder.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97462
func (c *WeDoc) CreateForm(ctx context.Context, spaceID, fatherID string, info FormInfo) (string, error) {
	body := struct {
		SpaceID  string   `json:"spaceid,omitempty"`
		FatherID string   `json:"fatherid,omitempty"`
		FormInfo FormInfo `json:"form_info"`
	}{SpaceID: spaceID, FatherID: fatherID, FormInfo: info}
	var out struct {
		FormID string `json:"formid"`
	}
	err := c.call(ctx, "create form", formCreatePath, body, &out)
	return out.FormID, err
}

// ModifyForm changes one collection form. oper selects the operation supported
// by WeCom (for example 1 to modify). Full question definitions must be sent.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97816
func (c *WeDoc) ModifyForm(ctx context.Context, formID string, oper int, info FormInfo) error {
	formID, err := required("modify form", "formid", formID)
	if err != nil {
		return err
	}
	body := struct {
		Oper     int      `json:"oper"`
		FormID   string   `json:"formid"`
		FormInfo FormInfo `json:"form_info"`
	}{Oper: oper, FormID: formID, FormInfo: info}
	return c.call(ctx, "modify form", formModifyPath, body, nil)
}

// FormInfo returns the definition of one collection form.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97817
func (c *WeDoc) FormInfo(ctx context.Context, formID string) (json.RawMessage, error) {
	formID, err := required("get form info", "formid", formID)
	if err != nil {
		return nil, err
	}
	var out struct {
		Info json.RawMessage `json:"form_info"`
	}
	err = c.call(ctx, "get form info", formInfoPath, map[string]string{"formid": formID}, &out)
	return out.Info, err
}

// FormStatistic returns how many answers a collection form received.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97818
func (c *WeDoc) FormStatistic(ctx context.Context, formID string, reqType int) (FormStatistic, error) {
	formID, err := required("get form statistic", "formid", formID)
	if err != nil {
		return FormStatistic{}, err
	}
	body := struct {
		FormID  string `json:"repeated_id"`
		ReqType int    `json:"req_type"`
	}{FormID: formID, ReqType: reqType}
	var out FormStatistic
	err = c.call(ctx, "get form statistic", formStatisticPath, body, &out)
	return out, err
}

// FormAnswer returns the answers of a collection form.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97819
func (c *WeDoc) FormAnswer(ctx context.Context, formID string, answerIDs []int64) (json.RawMessage, error) {
	formID, err := required("get form answer", "formid", formID)
	if err != nil {
		return nil, err
	}
	body := struct {
		FormID    string  `json:"repeated_id"`
		AnswerIDs []int64 `json:"answer_ids,omitempty"`
	}{FormID: formID, AnswerIDs: answerIDs}
	var out json.RawMessage
	err = c.call(ctx, "get form answer", formAnswerPath, body, &out)
	return out, err
}
