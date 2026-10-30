package wedoc

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	createDocPath      = "/cgi-bin/wedoc/create_doc"
	renameDocPath      = "/cgi-bin/wedoc/rename_doc"
	delDocPath         = "/cgi-bin/wedoc/del_doc"
	getDocBaseInfoPath = "/cgi-bin/wedoc/get_doc_base_info"
	docSharePath       = "/cgi-bin/wedoc/doc_share"
	docGetAuthPath     = "/cgi-bin/wedoc/doc_get_auth"
	modDocJoinRulePath = "/cgi-bin/wedoc/mod_doc_join_rule"
	modDocMemberPath   = "/cgi-bin/wedoc/mod_doc_member"
	// docMemberTypeUser marks a document member entry as a user id.
	docMemberTypeUser = 1
)

// DocTypeSmartsheet is the create_doc doc_type for a smart sheet.
const DocTypeSmartsheet = 10

// Document permission levels accepted by ShareDocWithUsers.
const (
	DocAuthRead      = 1
	DocAuthReadWrite = 2
	DocAuthAdmin     = 7
)

// Document is the result of creating a WeCom document.
type Document struct {
	DocID string `json:"docid"`
	URL   string `json:"url"`
}

// CreateDocOptions carries the optional placement of a new document. SpaceID
// and FatherID must be configured together; when both are empty the document
// is created in the application's default space.
type CreateDocOptions struct {
	SpaceID    string
	FatherID   string
	AdminUsers []string
}

// CreateDoc creates one WeCom document and returns its id and url. docType is
// the WeCom doc_type value; the caller decides whether the document is a
// smart sheet or another supported kind. Admin users are optional and may be
// user ids or "userid|userid".
//
// 文档: https://developer.work.weixin.qq.com/document/path/97460
func (c *WeDoc) CreateDoc(ctx context.Context, docType int, docName string, adminUsers ...string) (Document, error) {
	return c.CreateDocWithOptions(ctx, docType, docName, CreateDocOptions{AdminUsers: adminUsers})
}

// CreateDocWithOptions creates a document, optionally inside a space folder.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97460
func (c *WeDoc) CreateDocWithOptions(ctx context.Context, docType int, docName string, options CreateDocOptions) (Document, error) {
	if docType <= 0 {
		return Document{}, errors.New("wecom: create doc: docType must be positive")
	}
	name, err := required("create doc", "docName", docName)
	if err != nil {
		return Document{}, err
	}
	spaceID := strings.TrimSpace(options.SpaceID)
	fatherID := strings.TrimSpace(options.FatherID)
	if (spaceID == "") != (fatherID == "") {
		return Document{}, errors.New("wecom: create doc: spaceId and fatherId must be configured together")
	}
	request := struct {
		SpaceID    string   `json:"spaceid,omitempty"`
		FatherID   string   `json:"fatherid,omitempty"`
		DocType    int      `json:"doc_type"`
		DocName    string   `json:"doc_name"`
		AdminUsers []string `json:"admin_users,omitempty"`
	}{
		SpaceID:    spaceID,
		FatherID:   fatherID,
		DocType:    docType,
		DocName:    name,
		AdminUsers: normalizeUserIDs(options.AdminUsers),
	}
	var response struct {
		DocID string `json:"docid"`
		URL   string `json:"url"`
	}
	if err := c.call(ctx, "create doc", createDocPath, request, &response); err != nil {
		return Document{}, err
	}
	if strings.TrimSpace(response.DocID) == "" {
		return Document{}, errors.New("wecom: create doc response is missing docid")
	}
	return Document{DocID: response.DocID, URL: response.URL}, nil
}

// RenameDoc renames a document created by this application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97736
func (c *WeDoc) RenameDoc(ctx context.Context, docID, newName string) error {
	docID, err := required("rename doc", "docId", docID)
	if err != nil {
		return err
	}
	newName, err = required("rename doc", "newName", newName)
	if err != nil {
		return err
	}
	request := struct {
		DocID   string `json:"docid"`
		NewName string `json:"new_name"`
	}{DocID: docID, NewName: newName}
	return c.call(ctx, "rename doc", renameDocPath, request, nil)
}

// DocBaseInfo is the basic metadata returned by get_doc_base_info.
type DocBaseInfo struct {
	DocID      string `json:"docid"`
	DocName    string `json:"doc_name"`
	CreateTime uint64 `json:"create_time"`
	ModifyTime uint64 `json:"modify_time"`
	DocType    uint32 `json:"doc_type"`
}

// GetDocBaseInfo returns the basic metadata of a document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97734
func (c *WeDoc) GetDocBaseInfo(ctx context.Context, docID string) (DocBaseInfo, error) {
	docID, err := required("get doc base info", "docId", docID)
	if err != nil {
		return DocBaseInfo{}, err
	}
	request := struct {
		DocID string `json:"docid"`
	}{DocID: docID}
	var response struct {
		Info DocBaseInfo `json:"doc_base_info"`
	}
	if err := c.call(ctx, "get doc base info", getDocBaseInfoPath, request, &response); err != nil {
		return DocBaseInfo{}, err
	}
	return response.Info, nil
}

// ShareDoc returns the share URL of a document created by this application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97733
func (c *WeDoc) ShareDoc(ctx context.Context, docID string) (string, error) {
	docID, err := required("get share url", "docId", docID)
	if err != nil {
		return "", err
	}
	request := struct {
		DocID string `json:"docid"`
	}{DocID: docID}
	var response struct {
		ShareURL string `json:"share_url"`
	}
	if err := c.call(ctx, "get share url", docSharePath, request, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.ShareURL) == "" {
		return "", errors.New("wecom: get share url response is missing share_url")
	}
	return response.ShareURL, nil
}

// DocAccessRule is the document viewing rule returned by doc_get_auth.
type DocAccessRule struct {
	EnableCorpInternal             bool `json:"enable_corp_internal"`
	CorpInternalAuth               int  `json:"corp_internal_auth"`
	EnableCorpExternal             bool `json:"enable_corp_external"`
	CorpExternalAuth               int  `json:"corp_external_auth"`
	CorpInternalApproveOnlyByAdmin bool `json:"corp_internal_approve_only_by_admin"`
	CorpExternalApproveOnlyByAdmin bool `json:"corp_external_approve_only_by_admin"`
	BanShareExternal               bool `json:"ban_share_external"`
}

// DocWatermark is the watermark configuration returned by doc_get_auth.
type DocWatermark struct {
	MarginType      int    `json:"margin_type"`
	ShowVisitorName bool   `json:"show_visitor_name"`
	ShowText        bool   `json:"show_text"`
	Text            string `json:"text"`
}

// DocSecureSetting is the security configuration returned by doc_get_auth.
type DocSecureSetting struct {
	EnableReadonlyCopy    bool         `json:"enable_readonly_copy"`
	Watermark             DocWatermark `json:"watermark"`
	EnableReadonlyComment bool         `json:"enable_readonly_comment"`
}

// DocMember is one document notification-range member.
type DocMember struct {
	Type              int    `json:"type"`
	UserID            string `json:"userid"`
	TmpExternalUserID string `json:"tmp_external_userid"`
	Auth              int    `json:"auth"`
}

// CoAuth grants a department access to a document.
type CoAuth struct {
	Type         int    `json:"type"`
	DepartmentID uint64 `json:"departmentid"`
	Auth         int    `json:"auth"`
}

// DocAuth is the permission description returned by GetDocAuth.
type DocAuth struct {
	AccessRule    DocAccessRule    `json:"access_rule"`
	SecureSetting DocSecureSetting `json:"secure_setting"`
	DocMemberList []DocMember      `json:"doc_member_list"`
	CoAuthList    []CoAuth         `json:"co_auth_list"`
}

// GetDocAuth returns the viewing rules, member permissions, and security
// settings of a document created by this application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97461
func (c *WeDoc) GetDocAuth(ctx context.Context, docID string) (DocAuth, error) {
	docID, err := required("get doc auth", "docId", docID)
	if err != nil {
		return DocAuth{}, err
	}
	request := struct {
		DocID string `json:"docid"`
	}{DocID: docID}
	var response DocAuth
	if err := c.call(ctx, "get doc auth", docGetAuthPath, request, &response); err != nil {
		return DocAuth{}, err
	}
	return response, nil
}

// DocJoinRule is an optional-override document join rule. Nil fields are left
// unchanged; a non-nil field overwrites the current value.
type DocJoinRule struct {
	EnableCorpInternal             *bool    `json:"enable_corp_internal,omitempty"`
	CorpInternalAuth               *int     `json:"corp_internal_auth,omitempty"`
	EnableCorpExternal             *bool    `json:"enable_corp_external,omitempty"`
	CorpExternalAuth               *int     `json:"corp_external_auth,omitempty"`
	CorpInternalApproveOnlyByAdmin *bool    `json:"corp_internal_approve_only_by_admin,omitempty"`
	CorpExternalApproveOnlyByAdmin *bool    `json:"corp_external_approve_only_by_admin,omitempty"`
	BanShareExternal               *bool    `json:"ban_share_external,omitempty"`
	UpdateCoAuthList               *bool    `json:"update_co_auth_list,omitempty"`
	CoAuthList                     []CoAuth `json:"co_auth_list,omitempty"`
}

// SetDocJoinRule updates the document join rule. When UpdateCoAuthList is true
// the CoAuthList replaces the current department list; an empty list clears it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97778
func (c *WeDoc) SetDocJoinRule(ctx context.Context, docID string, rule DocJoinRule) error {
	docID, err := required("set doc join rule", "docId", docID)
	if err != nil {
		return err
	}
	request := map[string]any{"docid": docID}
	if rule.EnableCorpInternal != nil {
		request["enable_corp_internal"] = *rule.EnableCorpInternal
	}
	if rule.CorpInternalAuth != nil {
		request["corp_internal_auth"] = *rule.CorpInternalAuth
	}
	if rule.EnableCorpExternal != nil {
		request["enable_corp_external"] = *rule.EnableCorpExternal
	}
	if rule.CorpExternalAuth != nil {
		request["corp_external_auth"] = *rule.CorpExternalAuth
	}
	if rule.CorpInternalApproveOnlyByAdmin != nil {
		request["corp_internal_approve_only_by_admin"] = *rule.CorpInternalApproveOnlyByAdmin
	}
	if rule.CorpExternalApproveOnlyByAdmin != nil {
		request["corp_external_approve_only_by_admin"] = *rule.CorpExternalApproveOnlyByAdmin
	}
	if rule.BanShareExternal != nil {
		request["ban_share_external"] = *rule.BanShareExternal
	}
	if rule.UpdateCoAuthList != nil {
		request["update_co_auth_list"] = *rule.UpdateCoAuthList
	}
	if rule.CoAuthList != nil || (rule.UpdateCoAuthList != nil && *rule.UpdateCoAuthList) {
		list := rule.CoAuthList
		if list == nil {
			list = []CoAuth{}
		}
		request["co_auth_list"] = list
	}
	return c.call(ctx, "set doc join rule", modDocJoinRulePath, request, nil)
}

// CreateSmartsheet creates a smart sheet document (doc_type 10) and returns
// its docid. Sharing is deliberately controlled by the caller through
// ShareDocWithUsers; the document client does not own operator configuration.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97460
func (c *WeDoc) CreateSmartsheet(ctx context.Context, name string) (string, error) {
	document, err := c.CreateDoc(ctx, DocTypeSmartsheet, name)
	if err != nil {
		return "", err
	}
	return document.DocID, nil
}

// DeleteDoc deletes one WeCom document.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97735
func (c *WeDoc) DeleteDoc(ctx context.Context, docID string) error {
	docID, err := required("delete doc", "docId", docID)
	if err != nil {
		return err
	}
	request := struct {
		DocID string `json:"docid"`
	}{DocID: docID}
	return c.call(ctx, "delete doc", delDocPath, request, nil)
}

// UserPerm grants one user access to a document. Auth is DocAuthRead,
// DocAuthReadWrite, or DocAuthAdmin.
type UserPerm struct {
	UserID string
	Auth   int
}

// docMember is the wire form of one mod_doc_member update entry. type 1
// addresses a user id.
type docMember struct {
	Type   int    `json:"type"`
	Auth   int    `json:"auth"`
	UserID string `json:"userid"`
}

// docMemberRef is the wire form of one mod_doc_member delete entry.
type docMemberRef struct {
	Type   int    `json:"type"`
	UserID string `json:"userid"`
}

// ShareDocWithUsers grants document permissions through mod_doc_member. Each
// user id is added with the requested auth level; the call is a single
// request regardless of the number of users.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97781
func (c *WeDoc) ShareDocWithUsers(ctx context.Context, docID string, users []UserPerm) error {
	docID, err := required("share doc", "docId", docID)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return errors.New("wecom: share doc: at least one user is required")
	}
	members, err := buildDocMembers("share doc", users)
	if err != nil {
		return err
	}
	request := struct {
		DocID   string      `json:"docid"`
		Members []docMember `json:"update_file_member_list"`
	}{DocID: docID, Members: members}
	return c.call(ctx, "share doc", modDocMemberPath, request, nil)
}

// RemoveDocMembers removes users from a document through mod_doc_member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97781
func (c *WeDoc) RemoveDocMembers(ctx context.Context, docID string, userIDs ...string) error {
	docID, err := required("remove doc members", "docId", docID)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return errors.New("wecom: remove doc members: at least one userId is required")
	}
	members := make([]docMemberRef, 0, len(userIDs))
	for index, userID := range userIDs {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			return fmt.Errorf("wecom: remove doc members: userIds[%d] is required", index)
		}
		members = append(members, docMemberRef{Type: docMemberTypeUser, UserID: trimmed})
	}
	request := struct {
		DocID   string         `json:"docid"`
		Members []docMemberRef `json:"del_file_member_list"`
	}{DocID: docID, Members: members}
	return c.call(ctx, "remove doc members", modDocMemberPath, request, nil)
}

// ModifyDocMembers adds or updates and removes document members in one call.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97781
func (c *WeDoc) ModifyDocMembers(ctx context.Context, docID string, update []UserPerm, remove []string) error {
	docID, err := required("modify doc members", "docId", docID)
	if err != nil {
		return err
	}
	if len(update) == 0 && len(remove) == 0 {
		return errors.New("wecom: modify doc members: at least one member change is required")
	}
	updateMembers, err := buildDocMembers("modify doc members", update)
	if err != nil {
		return err
	}
	removeMembers := make([]docMemberRef, 0, len(remove))
	for index, userID := range remove {
		trimmed := strings.TrimSpace(userID)
		if trimmed == "" {
			return fmt.Errorf("wecom: modify doc members: remove[%d] is required", index)
		}
		removeMembers = append(removeMembers, docMemberRef{Type: docMemberTypeUser, UserID: trimmed})
	}
	request := struct {
		DocID  string         `json:"docid"`
		Update []docMember    `json:"update_file_member_list,omitempty"`
		Remove []docMemberRef `json:"del_file_member_list,omitempty"`
	}{DocID: docID, Update: updateMembers, Remove: removeMembers}
	return c.call(ctx, "modify doc members", modDocMemberPath, request, nil)
}

func buildDocMembers(op string, users []UserPerm) ([]docMember, error) {
	members := make([]docMember, 0, len(users))
	for index, user := range users {
		userID := strings.TrimSpace(user.UserID)
		if userID == "" {
			return nil, fmt.Errorf("wecom: %s: users[%d].userId is required", op, index)
		}
		switch user.Auth {
		case DocAuthRead, DocAuthReadWrite, DocAuthAdmin:
		default:
			return nil, fmt.Errorf("wecom: %s: users[%d].auth must be 1, 2, or 7", op, index)
		}
		members = append(members, docMember{Type: docMemberTypeUser, Auth: user.Auth, UserID: userID})
	}
	return members, nil
}
