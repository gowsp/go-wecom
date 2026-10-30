// Package wecom exposes the WeCom CGI API set: a shared authenticated
// transport plus one client per API namespace. Business rules stay in the
// application layer.
package wecom

import (
	"github.com/gowsp/go-wecom/advanced"
	"github.com/gowsp/go-wecom/agent"
	"github.com/gowsp/go-wecom/appchat"
	"github.com/gowsp/go-wecom/approval"
	"github.com/gowsp/go-wecom/auth"
	"github.com/gowsp/go-wecom/chatdata"
	"github.com/gowsp/go-wecom/checkin"
	"github.com/gowsp/go-wecom/client"
	"github.com/gowsp/go-wecom/contact"
	"github.com/gowsp/go-wecom/corpgroup"
	"github.com/gowsp/go-wecom/dial"
	"github.com/gowsp/go-wecom/external"
	"github.com/gowsp/go-wecom/externalpay"
	"github.com/gowsp/go-wecom/hr"
	"github.com/gowsp/go-wecom/idconvert"
	"github.com/gowsp/go-wecom/invoice"
	"github.com/gowsp/go-wecom/journal"
	"github.com/gowsp/go-wecom/kf"
	"github.com/gowsp/go-wecom/live"
	"github.com/gowsp/go-wecom/mail"
	"github.com/gowsp/go-wecom/media"
	"github.com/gowsp/go-wecom/meeting"
	"github.com/gowsp/go-wecom/meetingroom"
	"github.com/gowsp/go-wecom/message"
	"github.com/gowsp/go-wecom/miniapppay"
	"github.com/gowsp/go-wecom/msgaudit"
	"github.com/gowsp/go-wecom/pstn"
	"github.com/gowsp/go-wecom/schedule"
	"github.com/gowsp/go-wecom/security"
	"github.com/gowsp/go-wecom/todo"
	"github.com/gowsp/go-wecom/wedoc"
	"github.com/gowsp/go-wecom/wedrive"
)

// API is the aggregate CGI API surface. Client is the shared transport used by
// every field; each field is named after its WeCom API namespace.
type API struct {
	Client      *client.Client
	Advanced    *advanced.Advanced
	Agent       *agent.Agent
	Approval    *approval.Approval
	Auth        *auth.Auth
	ChatData    *chatdata.ChatData
	Checkin     *checkin.Checkin
	Contact     *contact.Contact
	CorpGroup   *corpgroup.CorpGroup
	Dial        *dial.Dial
	External    *external.External
	ExternalPay *externalpay.ExternalPay
	HR          *hr.HR
	IDConvert   *idconvert.IDConvert
	IPRange     *IPRange
	Invoice     *invoice.Invoice
	Journal     *journal.Journal
	Live        *live.Live
	MeetingRoom *meetingroom.MeetingRoom
	Meeting     *meeting.Meeting
	PSTN        *pstn.PSTN
	Schedule    *schedule.Schedule
	Security    *security.Security
	Todo        *todo.Todo
	Message     *message.Message
	AppChat     *appchat.AppChat
	KF          *kf.KF
	Mail        *mail.Mail
	Media       *media.Media
	MiniAppPay  *miniapppay.MiniAppPay
	MsgAudit    *msgaudit.MsgAudit
	WeDoc       *wedoc.WeDoc
	WeDrive     *wedrive.WeDrive
}

// New creates the CGI API set from corp credentials. A non-positive agentID
// disables application messages but keeps the document and customer-service
// APIs usable.
//
// Every field is constructed here; merchant payments are not part of this set
// because they use a merchant certificate and a separate client — build one with
// payment.New.
func New(corpID, corpSecret string, agentID int, options ...client.Option) *API {
	c := client.New(corpID, corpSecret, options...)
	return &API{
		Client:      c,
		Advanced:    advanced.New(c),
		Agent:       agent.New(c),
		Approval:    approval.New(c),
		Auth:        auth.New(c),
		ChatData:    chatdata.New(c),
		Checkin:     checkin.New(c),
		Contact:     contact.New(c),
		CorpGroup:   corpgroup.New(c),
		Dial:        dial.New(c),
		External:    external.New(c),
		ExternalPay: externalpay.New(c),
		HR:          hr.New(c),
		IDConvert:   idconvert.New(c),
		IPRange:     NewIPRange(c),
		Invoice:     invoice.New(c),
		Journal:     journal.New(c),
		Live:        live.New(c),
		MeetingRoom: meetingroom.New(c),
		Meeting:     meeting.New(c),
		MiniAppPay:  miniapppay.New(c),
		MsgAudit:    msgaudit.New(c),
		PSTN:        pstn.New(c),
		Schedule:    schedule.New(c),
		Security:    security.New(c),
		Todo:        todo.New(c),
		Message:     message.New(c, agentID),
		AppChat:     appchat.New(c, agentID),
		KF:          kf.New(c),
		Mail:        mail.New(c),
		Media:       media.New(c),
		WeDoc:       wedoc.New(c),
		WeDrive:     wedrive.New(c),
	}
}
