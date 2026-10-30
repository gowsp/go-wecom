package meeting

import (
	"context"
	"errors"
	"strings"
)

// PollQuestion is one question of a poll theme.
type PollQuestion struct {
	QuestionType int      `json:"question_type"`
	QuestionDesc string   `json:"question_desc"`
	Options      []string `json:"poll_option,omitempty"`
}

// PollTheme is one poll theme created before or during a meeting.
type PollTheme struct {
	OperatorUserID string         `json:"operator_userid"`
	InstanceID     int            `json:"instance_id,omitempty"`
	MeetingID      string         `json:"meetingid"`
	ThemeID        string         `json:"poll_theme_id,omitempty"`
	Topic          string         `json:"poll_topic,omitempty"`
	Desc           string         `json:"poll_desc,omitempty"`
	IsAnonymous    int            `json:"is_anony,omitempty"`
	Questions      []PollQuestion `json:"poll_questions,omitempty"`
}

// Poll is one running or finished poll of a theme.
type Poll struct {
	PollID      string `json:"poll_id"`
	Topic       string `json:"poll_topic,omitempty"`
	Status      int    `json:"status,omitempty"`
	IsShared    int    `json:"is_shared,omitempty"`
	IsAnonymous int    `json:"is_anony,omitempty"`
}

// PollThemeGroup groups the polls created from one theme.
type PollThemeGroup struct {
	ThemeID string `json:"poll_theme_id"`
	Polls   []Poll `json:"polls_info"`
}

// PollOption is one answer option of a poll detail.
type PollOption struct {
	OptionID   int    `json:"option_id,omitempty"`
	Desc       string `json:"option_desc"`
	VoteNumber int    `json:"option_vote_num,omitempty"`
}

// PollQuestionDetail is one question with its options.
type PollQuestionDetail struct {
	QuestionID   string       `json:"question_id,omitempty"`
	Desc         string       `json:"question_desc"`
	QuestionType int          `json:"question_type"`
	Options      []PollOption `json:"option_info"`
}

// PollDetail is the content and results of one poll.
type PollDetail struct {
	ThemeID     string               `json:"poll_theme_id"`
	Topic       string               `json:"poll_topic"`
	Desc        string               `json:"poll_desc,omitempty"`
	IsAnonymous int                  `json:"is_anony,omitempty"`
	Status      int                  `json:"status,omitempty"`
	IsShared    int                  `json:"is_shared,omitempty"`
	VoteTotal   int                  `json:"vote_total_num,omitempty"`
	Questions   []PollQuestionDetail `json:"poll_question_data"`
}

// CreatePollTheme creates one poll theme and returns its id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98834
func (c *Meeting) CreatePollTheme(ctx context.Context, theme PollTheme) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if theme.MeetingID == "" || strings.TrimSpace(theme.Topic) == "" || len(theme.Questions) == 0 {
		return "", errors.New("wecom: poll theme: meetingid, poll_topic and poll_questions are required")
	}
	var out struct {
		ThemeID string `json:"poll_theme_id"`
	}
	err := c.api.Call(ctx, "meeting poll theme", "/cgi-bin/meeting/poll/create_theme", theme, &out)
	return out.ThemeID, err
}

// UpdatePollTheme changes one poll theme.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98835
func (c *Meeting) UpdatePollTheme(ctx context.Context, theme PollTheme) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(theme.ThemeID) == "" {
		return errors.New("wecom: update poll theme: poll_theme_id is required")
	}
	return c.api.Call(ctx, "meeting update poll theme", "/cgi-bin/meeting/poll/update_theme", theme, nil)
}

// PollList returns the poll themes and polls of one meeting.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98836
func (c *Meeting) PollList(ctx context.Context, operatorUserID string, instanceID int, meetingID string) ([]PollThemeGroup, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(meetingID) == "" {
		return nil, errors.New("wecom: poll list: meetingid is required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID}
	var out struct {
		Themes []PollThemeGroup `json:"polls_theme_info"`
	}
	err := c.api.Call(ctx, "meeting poll list", "/cgi-bin/meeting/poll/get_poll_list", request, &out)
	return out.Themes, err
}

// PollThemeInfo is the definition of one poll theme.
type PollThemeInfo struct {
	Topic       string               `json:"poll_topic"`
	Desc        string               `json:"poll_desc,omitempty"`
	IsAnonymous int                  `json:"is_anony,omitempty"`
	Questions   []PollQuestionDetail `json:"poll_question_data"`
}

// PollTheme reads one poll theme.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98837
func (c *Meeting) PollTheme(ctx context.Context, operatorUserID string, instanceID int, meetingID, themeID string) (PollThemeInfo, error) {
	if err := check(c); err != nil {
		return PollThemeInfo{}, err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(themeID) == "" {
		return PollThemeInfo{}, errors.New("wecom: poll theme info: meetingid and poll_theme_id are required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
		ThemeID        string `json:"poll_theme_id"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID, ThemeID: themeID}
	var out PollThemeInfo
	err := c.api.Call(ctx, "meeting poll theme info", "/cgi-bin/meeting/poll/get_theme_info", request, &out)
	return out, err
}

// Poll returns the detail and results of one poll.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98838
func (c *Meeting) Poll(ctx context.Context, operatorUserID string, instanceID int, meetingID, pollID string) (PollDetail, error) {
	if err := check(c); err != nil {
		return PollDetail{}, err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(pollID) == "" {
		return PollDetail{}, errors.New("wecom: poll detail: meetingid and poll_id are required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
		PollID         string `json:"poll_id"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID, PollID: pollID}
	var out PollDetail
	err := c.api.Call(ctx, "meeting poll detail", "/cgi-bin/meeting/poll/get_poll_detail", request, &out)
	return out, err
}

// DeletePoll removes one poll, or a whole theme when pollID is empty.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98839
func (c *Meeting) DeletePoll(ctx context.Context, operatorUserID string, instanceID int, meetingID, themeID, pollID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(themeID) == "" {
		return errors.New("wecom: delete poll: meetingid and poll_theme_id are required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
		ThemeID        string `json:"poll_theme_id"`
		PollID         string `json:"poll_id,omitempty"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID, ThemeID: themeID, PollID: pollID}
	return c.api.Call(ctx, "meeting delete poll", "/cgi-bin/meeting/poll/delete", request, nil)
}

// StartPoll launches one poll theme and returns the poll id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98840
func (c *Meeting) StartPoll(ctx context.Context, operatorUserID string, instanceID int, meetingID, themeID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(themeID) == "" {
		return "", errors.New("wecom: start poll: meetingid and poll_theme_id are required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
		ThemeID        string `json:"poll_theme_id"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID, ThemeID: themeID}
	var out struct {
		PollID string `json:"poll_id"`
	}
	err := c.api.Call(ctx, "meeting start poll", "/cgi-bin/meeting/poll/start", request, &out)
	return out.PollID, err
}

// FinishPoll ends one running poll.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98841
func (c *Meeting) FinishPoll(ctx context.Context, operatorUserID string, instanceID int, meetingID, themeID, pollID string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(meetingID) == "" || strings.TrimSpace(pollID) == "" {
		return errors.New("wecom: finish poll: meetingid and poll_id are required")
	}
	request := struct {
		OperatorUserID string `json:"operator_userid"`
		InstanceID     int    `json:"instance_id,omitempty"`
		MeetingID      string `json:"meetingid"`
		ThemeID        string `json:"poll_theme_id,omitempty"`
		PollID         string `json:"poll_id"`
	}{OperatorUserID: operatorUserID, InstanceID: instanceID, MeetingID: meetingID, ThemeID: themeID, PollID: pollID}
	return c.api.Call(ctx, "meeting finish poll", "/cgi-bin/meeting/poll/finish", request, nil)
}
