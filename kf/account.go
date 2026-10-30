package kf

import (
	"context"
	"errors"
	"strings"
)

const maxAccountPage = 100

// Account is one customer-service account.
type Account struct {
	OpenKFID        string `json:"open_kfid"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	ManagePrivilege bool   `json:"manage_privilege"`
}

// ListAccounts returns customer-service accounts page by page. limit defaults to
// 100 and is capped at 100.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94661
func (c *KF) ListAccounts(ctx context.Context, offset, limit int) ([]Account, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > maxAccountPage {
		limit = maxAccountPage
	}
	var out struct {
		Accounts []Account `json:"account_list"`
	}
	err := c.api.Call(ctx, "customer service account list", "/cgi-bin/kf/account/list",
		map[string]int{"offset": offset, "limit": limit}, &out)
	return out.Accounts, err
}

// AddAccount creates a customer-service account. mediaID is the avatar media id
// uploaded through the material API and may be empty.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94662
func (c *KF) AddAccount(ctx context.Context, name, mediaID string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("wecom: add customer service account: name is required")
	}
	request := map[string]string{"name": name}
	if mediaID = strings.TrimSpace(mediaID); mediaID != "" {
		request["media_id"] = mediaID
	}
	var out struct {
		OpenKFID string `json:"open_kfid"`
	}
	err := c.api.Call(ctx, "customer service account add", "/cgi-bin/kf/account/add", request, &out)
	return out.OpenKFID, err
}

// DeleteAccount removes one customer-service account.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94663
func (c *KF) DeleteAccount(ctx context.Context, openKfID string) error {
	return c.accountUpdate(ctx, "customer service account del", "/cgi-bin/kf/account/del", openKfID, "", "")
}

// UpdateAccount changes the name or avatar of a customer-service account. An
// empty argument leaves that field unchanged.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94664
func (c *KF) UpdateAccount(ctx context.Context, openKfID, name, mediaID string) error {
	return c.accountUpdate(ctx, "customer service account update", "/cgi-bin/kf/account/update", openKfID, name, mediaID)
}

func (c *KF) accountUpdate(ctx context.Context, op, path, openKfID, name, mediaID string) error {
	openKfID = strings.TrimSpace(openKfID)
	if openKfID == "" {
		return errors.New("wecom: " + op + ": open_kfid is required")
	}
	request := map[string]string{"open_kfid": openKfID}
	if name = strings.TrimSpace(name); name != "" {
		request["name"] = name
	}
	if mediaID = strings.TrimSpace(mediaID); mediaID != "" {
		request["media_id"] = mediaID
	}
	return c.api.Call(ctx, op, path, request, nil)
}

// AccountLink returns a contact link that opens one customer-service account.
// scene is echoed back in the enter-session event and may be empty.
//
// 文档: https://developer.work.weixin.qq.com/document/path/94665
func (c *KF) AccountLink(ctx context.Context, openKfID, scene string) (string, error) {
	openKfID = strings.TrimSpace(openKfID)
	if openKfID == "" {
		return "", errors.New("wecom: customer service account link: open_kfid is required")
	}
	request := map[string]string{"open_kfid": openKfID}
	if scene = strings.TrimSpace(scene); scene != "" {
		request["scene"] = scene
	}
	var out struct {
		URL string `json:"url"`
	}
	err := c.api.Call(ctx, "customer service account link", "/cgi-bin/kf/add_contact_way", request, &out)
	return out.URL, err
}
