package contact

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
)

// Member is one contact member. Fields mirror the official payload; optional
// fields are omitted when zero so an update does not clear them.
type Member struct {
	UserID           string          `json:"userid"`
	Name             string          `json:"name,omitempty"`
	Alias            string          `json:"alias,omitempty"`
	Mobile           string          `json:"mobile,omitempty"`
	Department       []int           `json:"department,omitempty"`
	Order            []int           `json:"order,omitempty"`
	Position         string          `json:"position,omitempty"`
	Gender           string          `json:"gender,omitempty"`
	Email            string          `json:"email,omitempty"`
	BizMail          string          `json:"biz_mail,omitempty"`
	DirectLeader     []string        `json:"direct_leader,omitempty"`
	IsLeaderInDept   []int           `json:"is_leader_in_dept,omitempty"`
	MainDepartment   int             `json:"main_department,omitempty"`
	ExtAttr          *ExtAttr        `json:"extattr,omitempty"`
	Telephone        string          `json:"telephone,omitempty"`
	Address          string          `json:"address,omitempty"`
	ExternalPosition string          `json:"external_position,omitempty"`
	ExternalProfile  json.RawMessage `json:"external_profile,omitempty"`
	Enable           int             `json:"enable,omitempty"`

	// Read-only fields returned by GetMember and the detail list.
	Avatar      string `json:"avatar,omitempty"`
	ThumbAvatar string `json:"thumb_avatar,omitempty"`
	Status      int    `json:"status,omitempty"`
	QRCode      string `json:"qr_code,omitempty"`
}

// MemberSummary is one entry of the department member list.
type MemberSummary struct {
	UserID     string `json:"userid"`
	Name       string `json:"name"`
	Department []int  `json:"department"`
}

// ExtAttr is the custom attribute block of a member.
type ExtAttr struct {
	Attrs []ExtAttrItem `json:"attrs"`
}

// ExtAttrItem is one custom attribute. Type 0 uses Text, type 1 uses Web.
type ExtAttrItem struct {
	Type int          `json:"type"`
	Name string       `json:"name"`
	Text *ExtAttrText `json:"text,omitempty"`
	Web  *ExtAttrWeb  `json:"web,omitempty"`
}

// ExtAttrText is a text custom attribute.
type ExtAttrText struct {
	Value string `json:"value"`
}

// ExtAttrWeb is a link custom attribute.
type ExtAttrWeb struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// NewExtAttrText builds a text custom attribute. Attributes must first exist in
// the admin console; unknown names are ignored by WeCom.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100067
func NewExtAttrText(name, value string) ExtAttrItem {
	return ExtAttrItem{Type: 0, Name: name, Text: &ExtAttrText{Value: value}}
}

// NewExtAttrWeb builds a link custom attribute.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100067
func NewExtAttrWeb(name, url, title string) ExtAttrItem {
	return ExtAttrItem{Type: 1, Name: name, Web: &ExtAttrWeb{URL: url, Title: title}}
}

// InviteResult reports members that could not be invited.
type InviteResult struct {
	InvalidUsers []string `json:"invaliduser"`
	InvalidParty []string `json:"invalidparty"`
	InvalidTag   []string `json:"invalidtag"`
}

// MemberIDPage is one page of the member id list.
type MemberIDPage struct {
	NextCursor string      `json:"next_cursor"`
	DeptUsers  []MemberRef `json:"dept_user"`
}

// MemberRef pairs a member id with its main department.
type MemberRef struct {
	UserID     string `json:"userid"`
	Department int    `json:"department"`
}

// CreateMember adds one member. Name and Department are required.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90195
func (c *Contact) CreateMember(ctx context.Context, member Member) error {
	if strings.TrimSpace(member.UserID) == "" {
		return errors.New("wecom: create member: userid is required")
	}
	if strings.TrimSpace(member.Name) == "" {
		return errors.New("wecom: create member: name is required")
	}
	if len(member.Department) == 0 {
		return errors.New("wecom: create member: department is required")
	}
	return c.call(ctx, "contact create member", "/cgi-bin/user/create", member, nil)
}

// GetMember reads one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90196
func (c *Contact) GetMember(ctx context.Context, userID string) (Member, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Member{}, errors.New("wecom: get member: userid is required")
	}
	var out Member
	err := c.get(ctx, "contact get member", "/cgi-bin/user/get",
		url.Values{"userid": {userID}}, &out)
	out.UserID = userID
	return out, err
}

// UpdateMember changes one member; omitted fields stay unchanged.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90197
func (c *Contact) UpdateMember(ctx context.Context, member Member) error {
	if strings.TrimSpace(member.UserID) == "" {
		return errors.New("wecom: update member: userid is required")
	}
	return c.call(ctx, "contact update member", "/cgi-bin/user/update", member, nil)
}

// DeleteMember removes one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90198
func (c *Contact) DeleteMember(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("wecom: delete member: userid is required")
	}
	return c.get(ctx, "contact delete member", "/cgi-bin/user/delete",
		url.Values{"userid": {userID}}, nil)
}

// BatchDeleteMembers removes up to 200 members.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90199
func (c *Contact) BatchDeleteMembers(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return errors.New("wecom: batch delete members: useridlist is empty")
	}
	if len(userIDs) > 200 {
		return errors.New("wecom: batch delete members: at most 200 user ids")
	}
	request := struct {
		UserIDList []string `json:"useridlist"`
	}{UserIDList: userIDs}
	return c.call(ctx, "contact batch delete members", "/cgi-bin/user/batchdelete", request, nil)
}

// ListDepartmentMembers returns the simple member list of one department.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90200
func (c *Contact) ListDepartmentMembers(ctx context.Context, departmentID int, fetchChild bool) ([]MemberSummary, error) {
	query := url.Values{
		"department_id": {itoa(departmentID)},
		"fetch_child":   {boolString(fetchChild)},
	}
	var out struct {
		UserList []MemberSummary `json:"userlist"`
	}
	err := c.get(ctx, "contact list department members",
		"/cgi-bin/user/simplelist", query, &out)
	return out.UserList, err
}

// ListDepartmentMemberDetails returns the full member list of one department.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90201
func (c *Contact) ListDepartmentMemberDetails(ctx context.Context, departmentID int, fetchChild bool) ([]Member, error) {
	query := url.Values{
		"department_id": {itoa(departmentID)},
		"fetch_child":   {boolString(fetchChild)},
	}
	var out struct {
		UserList []Member `json:"userlist"`
	}
	err := c.get(ctx, "contact list department member details",
		"/cgi-bin/user/list", query, &out)
	return out.UserList, err
}

// ConvertToOpenID maps a userid to an openid.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90202
func (c *Contact) ConvertToOpenID(ctx context.Context, userID string) (string, error) {
	if strings.TrimSpace(userID) == "" {
		return "", errors.New("wecom: convert to openid: userid is required")
	}
	var out struct {
		OpenID string `json:"openid"`
	}
	err := c.call(ctx, "contact convert to openid", "/cgi-bin/user/convert_to_openid",
		map[string]string{"userid": userID}, &out)
	return out.OpenID, err
}

// ConvertToUserID maps an openid back to a userid.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90202
func (c *Contact) ConvertToUserID(ctx context.Context, openID string) (string, error) {
	if strings.TrimSpace(openID) == "" {
		return "", errors.New("wecom: convert to userid: openid is required")
	}
	var out struct {
		UserID string `json:"userid"`
	}
	err := c.call(ctx, "contact convert to userid", "/cgi-bin/user/convert_to_userid",
		map[string]string{"openid": openID}, &out)
	return out.UserID, err
}

// AuthSuccess reports that the member passed second-factor verification.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90203
func (c *Contact) AuthSuccess(ctx context.Context, userID string) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return errors.New("wecom: auth success: userid is required")
	}
	return c.get(ctx, "contact auth success", "/cgi-bin/user/authsucc",
		url.Values{"userid": {userID}}, nil)
}

// InviteMembers invites members by userid, department or tag.
//
// 文档: https://developer.work.weixin.qq.com/document/path/90975
func (c *Contact) InviteMembers(ctx context.Context, users, parties, tags []string) (InviteResult, error) {
	if len(users) == 0 && len(parties) == 0 && len(tags) == 0 {
		return InviteResult{}, errors.New("wecom: invite members: user, party or tag is required")
	}
	request := struct {
		User  []string `json:"user,omitempty"`
		Party []string `json:"party,omitempty"`
		Tag   []string `json:"tag,omitempty"`
	}{User: users, Party: parties, Tag: tags}
	var out InviteResult
	err := c.call(ctx, "contact invite members", "/cgi-bin/batch/invite", request, &out)
	return out, err
}

// JoinQRCode returns the QR code that lets people join the enterprise.
//
// 文档: https://developer.work.weixin.qq.com/document/path/91714
func (c *Contact) JoinQRCode(ctx context.Context, sizeType int) (string, error) {
	var out struct {
		QRCode string `json:"join_qrcode"`
	}
	query := url.Values{}
	if sizeType > 0 {
		query.Set("size_type", itoa(sizeType))
	}
	err := c.get(ctx, "contact join qr code", "/cgi-bin/corp/get_join_qrcode", query, &out)
	return out.QRCode, err
}

// GetUserIDByMobile resolves a mobile number to a userid.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95402
func (c *Contact) GetUserIDByMobile(ctx context.Context, mobile string) (string, error) {
	mobile = strings.TrimSpace(mobile)
	if mobile == "" {
		return "", errors.New("wecom: get userid by mobile: mobile is required")
	}
	var out struct {
		UserID string `json:"userid"`
	}
	err := c.call(ctx, "contact get userid by mobile", "/cgi-bin/user/getuserid",
		map[string]string{"mobile": mobile}, &out)
	return out.UserID, err
}

// Email types accepted by GetUserIDByEmail.
const (
	EmailTypeEnterprise = 1
	EmailTypePersonal   = 2
)

// GetUserIDByEmail resolves an email address to a userid. emailType selects the
// enterprise or personal mailbox.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95895
func (c *Contact) GetUserIDByEmail(ctx context.Context, email string, emailType int) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", errors.New("wecom: get userid by email: email is required")
	}
	request := struct {
		Email     string `json:"email"`
		EmailType int    `json:"email_type,omitempty"`
	}{Email: email, EmailType: emailType}
	var out struct {
		UserID string `json:"userid"`
	}
	err := c.call(ctx, "contact get userid by email", "/cgi-bin/user/get_userid_by_email", request, &out)
	return out.UserID, err
}

// ListMemberIDs returns one page of member ids, newest first. An empty cursor
// starts the walk; stop when NextCursor is empty.
//
// 文档: https://developer.work.weixin.qq.com/document/path/96067
func (c *Contact) ListMemberIDs(ctx context.Context, cursor string, limit int) (MemberIDPage, error) {
	request := struct {
		Cursor string `json:"cursor,omitempty"`
		Limit  int    `json:"limit,omitempty"`
	}{Cursor: cursor, Limit: limit}
	var out MemberIDPage
	err := c.call(ctx, "contact list member ids", "/cgi-bin/user/list_id", request, &out)
	return out, err
}
