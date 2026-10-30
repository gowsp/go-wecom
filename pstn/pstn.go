// Package pstn implements the WeCom emergency voice call API (/cgi-bin/pstncc).
package pstn

import (
	"context"
	"errors"
	"strings"

	"github.com/gowsp/go-wecom/client"
)

// PSTN places emergency voice calls with one shared transport.
type PSTN struct{ api *client.Client }

// New creates an emergency-call client.
func New(api *client.Client) *PSTN { return &PSTN{api: api} }

// CallState is the result of placing one call. Code 0 means the call was placed;
// CallID is then used to query the answer state.
type CallState struct {
	Code   int    `json:"code"`
	CallID string `json:"callid"`
	UserID string `json:"userid"`
}

// Call places emergency voice calls to the given members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91627
func (c *PSTN) Call(ctx context.Context, calleeUserIDs []string) ([]CallState, error) {
	if c == nil || c.api == nil {
		return nil, errors.New("wecom: pstn client is unavailable")
	}
	if len(calleeUserIDs) == 0 {
		return nil, errors.New("wecom: pstn call: callee_userid is empty")
	}
	request := struct {
		Callee []string `json:"callee_userid"`
	}{Callee: calleeUserIDs}
	var out struct {
		States []CallState `json:"states"`
	}
	err := c.api.Call(ctx, "pstn call", "/cgi-bin/pstncc/call", request, &out)
	return out.States, err
}

// AnswerState is the state of one placed call. IsTalked 1 means answered.
type AnswerState struct {
	IsTalked int   `json:"istalked"`
	CallTime int64 `json:"calltime"`
	TalkTime int   `json:"talktime"`
	Reason   int   `json:"reason"`
}

// CallState reports whether one member answered the call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91628
func (c *PSTN) AnswerState(ctx context.Context, calleeUserID, callID string) (AnswerState, error) {
	if c == nil || c.api == nil {
		return AnswerState{}, errors.New("wecom: pstn client is unavailable")
	}
	if strings.TrimSpace(calleeUserID) == "" || strings.TrimSpace(callID) == "" {
		return AnswerState{}, errors.New("wecom: pstn state: callee_userid and callid are required")
	}
	request := struct {
		Callee string `json:"callee_userid"`
		CallID string `json:"callid"`
	}{Callee: calleeUserID, CallID: callID}
	var out AnswerState
	err := c.api.Call(ctx, "pstn state", "/cgi-bin/pstncc/getstates", request, &out)
	return out, err
}
