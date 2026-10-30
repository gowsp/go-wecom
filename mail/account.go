package mail

import "context"

const (
	emailAliasPath       = "/cgi-bin/exmail/app/get_email_alias"
	updateEmailAliasPath = "/cgi-bin/exmail/app/update_email_alias"
	accountActivePath    = "/cgi-bin/exmail/account/act_email"
	userOptionGetPath    = "/cgi-bin/exmail/useroption/get"
	userOptionUpdatePath = "/cgi-bin/exmail/useroption/update"
)

// Account status values accepted by SetAccountActive.
const (
	AccountEnable  = 1
	AccountDisable = 2
)

// User option types accepted by UserOptions and UpdateUserOptions.
const (
	OptionForceSecureLogin = 1
	OptionIMAPSMTP         = 2
	OptionPOPSMTP          = 3
	OptionSecureLogin      = 4
)

// EmailAlias describes the application mailbox address and its aliases.
type EmailAlias struct {
	Email     string   `json:"email"`
	AliasList []string `json:"alias_list"`
}

// UserOption is one mailbox feature switch. Value is "1" for enabled and "0"
// for disabled.
type UserOption struct {
	Type  int    `json:"type"`
	Value string `json:"value"`
}

// EmailAlias returns the mailbox account bound to the calling application.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97991
func (c *Mail) EmailAlias(ctx context.Context) (EmailAlias, error) {
	var out EmailAlias
	err := c.call(ctx, "get app email alias", emailAliasPath, nil, &out)
	return out, err
}

// UpdateEmailAlias changes the application mailbox address.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97373
func (c *Mail) UpdateEmailAlias(ctx context.Context, newEmail string) error {
	newEmail, err := required("update app email alias", "new_email", newEmail)
	if err != nil {
		return err
	}
	return c.call(ctx, "update app email alias", updateEmailAliasPath, map[string]string{"new_email": newEmail}, nil)
}

// SetAccountActive enables or disables a member mailbox or a public mailbox.
// userID takes precedence when both are provided.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95512
func (c *Mail) SetAccountActive(ctx context.Context, userID string, publicMailID uint32, accountType int) error {
	if userID == "" && publicMailID == 0 {
		return errPair("activate mailbox", "userid or publicemail_id")
	}
	if accountType != AccountEnable && accountType != AccountDisable {
		return errValue("activate mailbox", "type")
	}
	request := struct {
		UserID       string `json:"userid,omitempty"`
		PublicMailID uint32 `json:"publicemail_id,omitempty"`
		Type         int    `json:"type"`
	}{UserID: userID, PublicMailID: publicMailID, Type: accountType}
	return c.call(ctx, "activate mailbox", accountActivePath, request, nil)
}

// UserOptions reads mailbox feature switches for one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/95513
func (c *Mail) UserOptions(ctx context.Context, userID string, types []int) ([]UserOption, error) {
	userID, err := required("get user options", "userid", userID)
	if err != nil {
		return nil, err
	}
	if _, err := requireIDs("get user options", "type", types); err != nil {
		return nil, err
	}
	request := struct {
		UserID string `json:"userid"`
		Type   []int  `json:"type"`
	}{UserID: userID, Type: types}
	var out struct {
		Option List[UserOption] `json:"option"`
	}
	err = c.call(ctx, "get user options", userOptionGetPath, request, &out)
	return out.Option.List, err
}

// UpdateUserOptions changes mailbox feature switches for one member.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98008
func (c *Mail) UpdateUserOptions(ctx context.Context, userID string, options []UserOption) error {
	userID, err := required("update user options", "userid", userID)
	if err != nil {
		return err
	}
	if _, err := requireIDs("update user options", "option", options); err != nil {
		return err
	}
	request := struct {
		UserID string           `json:"userid"`
		Option List[UserOption] `json:"option"`
	}{UserID: userID, Option: newList(options)}
	return c.call(ctx, "update user options", userOptionUpdatePath, request, nil)
}
