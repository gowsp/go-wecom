package mail

import "context"

const (
	publicMailCreatePath = "/cgi-bin/exmail/publicmail/create"
	publicMailUpdatePath = "/cgi-bin/exmail/publicmail/update"
	publicMailDeletePath = "/cgi-bin/exmail/publicmail/delete"
	publicMailGetPath    = "/cgi-bin/exmail/publicmail/get"
	publicMailSearchPath = "/cgi-bin/exmail/publicmail/search"
	authCodeListPath     = "/cgi-bin/exmail/publicmail/get_auth_code_list"
	authCodeDeletePath   = "/cgi-bin/exmail/publicmail/delete_auth_code"
)

// PublicMail is a shared mailbox. Owners come from members, departments and
// tags; at least one of those lists must be set.
type PublicMail struct {
	ID             uint32       `json:"id,omitempty"`
	Email          string       `json:"email,omitempty"`
	Name           string       `json:"name,omitempty"`
	UserIDList     List[string] `json:"userid_list,omitempty"`
	DepartmentList List[int64]  `json:"department_list,omitempty"`
	TagList        List[int64]  `json:"tag_list,omitempty"`
	AliasList      List[string] `json:"alias_list,omitempty"`

	// CreateAuthCode asks the server to issue a client-specific password.
	CreateAuthCode int `json:"create_auth_code,omitempty"`
	// AuthCodeInfo describes that password; it is not returned by reads.
	AuthCodeInfo *AuthCodeInfo `json:"auth_code_info,omitempty"`
}

// AuthCodeInfo is the optional remark attached to a client password.
type AuthCodeInfo struct {
	Remark string `json:"remark,omitempty"`
}

// AuthCode is one client-specific password of a public mailbox.
type AuthCode struct {
	ID          uint32 `json:"auth_code_id"`
	CreateTime  int64  `json:"create_time"`
	LastUseTime int64  `json:"last_use_time"`
	Remark      string `json:"remark"`
}

// CreatePublicMail creates a shared mailbox.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95511
func (c *Mail) CreatePublicMail(ctx context.Context, box PublicMail) error {
	if _, err := required("create public mail", "email", box.Email); err != nil {
		return err
	}
	return c.call(ctx, "create public mail", publicMailCreatePath, box, nil)
}

// UpdatePublicMail changes a shared mailbox. Omitted owner lists stay unchanged.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98000
func (c *Mail) UpdatePublicMail(ctx context.Context, box PublicMail) error {
	if box.ID == 0 {
		return errValue("update public mail", "id")
	}
	return c.call(ctx, "update public mail", publicMailUpdatePath, box, nil)
}

// DeletePublicMail removes a shared mailbox.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98001
func (c *Mail) DeletePublicMail(ctx context.Context, id uint32) error {
	if id == 0 {
		return errValue("delete public mail", "id")
	}
	return c.call(ctx, "delete public mail", publicMailDeletePath, map[string]uint32{"id": id}, nil)
}

// GetPublicMail returns the shared mailboxes with the given ids.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98002
func (c *Mail) GetPublicMail(ctx context.Context, ids []uint32) ([]PublicMail, error) {
	if _, err := requireIDs("get public mail", "id_list", ids); err != nil {
		return nil, err
	}
	var out struct {
		List []PublicMail `json:"list"`
	}
	err := c.call(ctx, "get public mail", publicMailGetPath, map[string][]uint32{"id_list": ids}, &out)
	return out.List, err
}

// SearchPublicMail lists shared mailboxes. fuzzy enables fuzzy matching; email
// narrows the result to one mailbox name or address.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98003
func (c *Mail) SearchPublicMail(ctx context.Context, fuzzy bool, email string) ([]PublicMail, error) {
	query := map[string]string{"email": email}
	if fuzzy {
		query["fuzzy"] = "1"
	} else {
		query["fuzzy"] = "0"
	}
	var out struct {
		List []PublicMail `json:"list"`
	}
	err := c.get(ctx, "search public mail", publicMailSearchPath, query, &out)
	return out.List, err
}

// AuthCodes lists the client-specific passwords of a shared mailbox.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100183
func (c *Mail) AuthCodes(ctx context.Context, id uint32) ([]AuthCode, error) {
	if id == 0 {
		return nil, errValue("list auth codes", "id")
	}
	var out struct {
		List []AuthCode `json:"auth_code_list"`
	}
	err := c.call(ctx, "list auth codes", authCodeListPath, map[string]uint32{"id": id}, &out)
	return out.List, err
}

// DeleteAuthCode removes one client-specific password from a shared mailbox.
//
// 文档: https://developer.work.weixin.qq.com/document/path/100184
func (c *Mail) DeleteAuthCode(ctx context.Context, id, authCodeID uint32) error {
	if id == 0 || authCodeID == 0 {
		return errValue("delete auth code", "id or auth_code_id")
	}
	request := struct {
		ID         uint32 `json:"id"`
		AuthCodeID uint32 `json:"auth_code_id"`
	}{ID: id, AuthCodeID: authCodeID}
	return c.call(ctx, "delete auth code", authCodeDeletePath, request, nil)
}
