package mail

import "context"

const composeSendPath = "/cgi-bin/exmail/app/compose_send"

// Address is a recipient set. Emails and UserIDs are combined; at least one
// entry must be present in To.
type Address struct {
	Emails  []string `json:"emails,omitempty"`
	UserIDs []string `json:"userids,omitempty"`
}

// Attachment is one base64-encoded attachment.
type Attachment struct {
	FileName string `json:"file_name"`
	Content  string `json:"content"`
}

// Reminders configures the schedule reminder and repeat rule.
type Reminders struct {
	IsRemind             int   `json:"is_remind"`
	RemindBeforeEventMin int   `json:"remind_before_event_mins,omitempty"`
	IsRepeat             int   `json:"is_repeat,omitempty"`
	IsCustomRepeat       int   `json:"is_custom_repeat,omitempty"`
	Timezone             int   `json:"timezone,omitempty"`
	RepeatInterval       int   `json:"repeat_interval,omitempty"`
	RepeatType           int   `json:"repeat_type,omitempty"`
	RepeatDayOfWeek      []int `json:"repeat_day_of_week,omitempty"`
	RepeatDayOfMonth     []int `json:"repeat_day_of_month,omitempty"`
	RepeatWeekOfMonth    []int `json:"repeat_week_of_month,omitempty"`
	RepeatMonthOfYear    []int `json:"repeat_month_of_year,omitempty"`
	RepeatUntil          int64 `json:"repeat_until,omitempty"`
}

// Schedule turns a message into a calendar entry. Method follows iCalendar
// semantics, for example "request".
type Schedule struct {
	Method    string     `json:"method"`
	Location  string     `json:"location,omitempty"`
	StartTime int64      `json:"start_time"`
	EndTime   int64      `json:"end_time,omitempty"`
	Reminders *Reminders `json:"reminders,omitempty"`
}

// MeetingOption configures the meeting created by a meeting mail.
type MeetingOption struct {
	Password              string `json:"password,omitempty"`
	AutoRecord            int    `json:"auto_record,omitempty"`
	EnableWaitingRoom     bool   `json:"enable_waiting_room,omitempty"`
	AllowEnterBeforeHost  bool   `json:"allow_enter_before_host,omitempty"`
	EnterRestraint        int    `json:"enter_restraint,omitempty"`
	EnableScreenWatermark bool   `json:"enable_screen_watermark,omitempty"`
	EnableEnterMute       int    `json:"enable_enter_mute,omitempty"`
	RemindScope           int    `json:"remind_scope,omitempty"`
	WaterMarkType         int    `json:"water_mark_type,omitempty"`
}

// Meeting configures hosts and administrators of a meeting mail.
type Meeting struct {
	Option MeetingOption `json:"option,omitempty"`
	Hosts  List[string]  `json:"hosts,omitempty"`
	Admins List[string]  `json:"meeting_admins,omitempty"`
}

// SendRequest is the body of compose_send. Subject and Content are required;
// To must carry at least one email address or user id.
type SendRequest struct {
	To             Address      `json:"to"`
	Cc             *Address     `json:"cc,omitempty"`
	Bcc            *Address     `json:"bcc,omitempty"`
	Subject        string       `json:"subject"`
	Content        string       `json:"content"`
	AttachmentList []Attachment `json:"attachment_list,omitempty"`
	EnableIDTrans  int          `json:"enable_id_trans,omitempty"`
	Schedule       *Schedule    `json:"schedule,omitempty"`
	Meeting        *Meeting     `json:"meeting,omitempty"`
}

// Send delivers one mail, optionally as a schedule entry or meeting invite.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97445
// 文档: https://developer.work.weixin.qq.com/document/path/97854
// 文档: https://developer.work.weixin.qq.com/document/path/97855
func (c *Mail) Send(ctx context.Context, request SendRequest) error {
	if _, err := required("send mail", "to", joinAddress(request.To)); err != nil {
		return err
	}
	if _, err := required("send mail", "subject", request.Subject); err != nil {
		return err
	}
	return c.call(ctx, "send mail", composeSendPath, request, nil)
}

func joinAddress(address Address) string {
	for _, email := range address.Emails {
		if email != "" {
			return email
		}
	}
	for _, id := range address.UserIDs {
		if id != "" {
			return id
		}
	}
	return ""
}
