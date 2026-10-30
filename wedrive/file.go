package wedrive

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

// File types reported by the file endpoints. CreateFile accepts the folder and
// document kinds only; a plain file arrives through an upload.
const (
	FileTypeFolder = 1
	FileTypeFile   = 2
	FileTypeDoc    = 3
	FileTypeSheet  = 4
	FileTypeForm   = 5
	FileTypeSlide  = 6
)

// Sort orders accepted by ListFiles.
const (
	SortByNameAsc  = 1
	SortByNameDesc = 2
	SortBySizeAsc  = 3
	SortBySizeDesc = 4
	SortByTimeAsc  = 5
	SortByTimeDesc = 6
)

// File is one micro disk entry.
type File struct {
	FileID     string `json:"fileid"`
	FileName   string `json:"file_name"`
	SpaceID    string `json:"spaceid,omitempty"`
	FatherID   string `json:"fatherid,omitempty"`
	FileSize   int64  `json:"file_size,omitempty"`
	CTime      int64  `json:"ctime,omitempty"`
	MTime      int64  `json:"mtime,omitempty"`
	FileType   int    `json:"file_type,omitempty"`
	FileStatus int    `json:"file_status,omitempty"`
	SHA        string `json:"sha,omitempty"`
	MD5        string `json:"md5,omitempty"`
	URL        string `json:"url,omitempty"`
}

// FileList is one page of directory entries.
type FileList struct {
	HasMore   bool   `json:"has_more"`
	NextStart int    `json:"next_start"`
	Items     []File `json:"-"`
}

// fileListEnvelope decodes the nested {"file_list":{"item":[...]}} shape.
type fileListEnvelope struct {
	HasMore   bool `json:"has_more"`
	NextStart int  `json:"next_start"`
	List      struct {
		Items []File `json:"item"`
	} `json:"file_list"`
}

// ListFiles returns one directory page. fatherID is the directory to list and is
// the space id for the root. start is 0 for the first page and then the previous
// page's NextStart; limit is capped at 1000 and sortBy takes one of the SortBy*
// constants.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93657
func (c *WeDrive) ListFiles(ctx context.Context, spaceID, fatherID string, sortBy, start, limit int) (FileList, error) {
	if err := check(c); err != nil {
		return FileList{}, err
	}
	spaceID = strings.TrimSpace(spaceID)
	if spaceID == "" {
		return FileList{}, errors.New("wecom: list files: spaceid is required")
	}
	if strings.TrimSpace(fatherID) == "" {
		return FileList{}, errors.New("wecom: list files: fatherid is required, use the spaceid for the root")
	}
	if sortBy < SortByNameAsc || sortBy > SortByTimeDesc {
		return FileList{}, errors.New("wecom: list files: sort_type must be between 1 and 6")
	}
	if start < 0 {
		return FileList{}, errors.New("wecom: list files: start must not be negative")
	}
	if limit <= 0 || limit > 1000 {
		return FileList{}, errors.New("wecom: list files: limit must be between 1 and 1000")
	}
	request := struct {
		SpaceID  string `json:"spaceid"`
		FatherID string `json:"fatherid,omitempty"`
		SortType int    `json:"sort_type,omitempty"`
		Start    int    `json:"start,omitempty"`
		Limit    int    `json:"limit,omitempty"`
	}{SpaceID: spaceID, FatherID: fatherID, SortType: sortBy, Start: start, Limit: limit}
	var out fileListEnvelope
	if err := c.api.Call(ctx, "wedrive list files", "/cgi-bin/wedrive/file_list", request, &out); err != nil {
		return FileList{}, err
	}
	return FileList{HasMore: out.HasMore, NextStart: out.NextStart, Items: out.List.Items}, nil
}

// UploadFile stores one small file inline (base64).
//
// 文档: https://developer.work.weixin.qq.com/document/path/97880
func (c *WeDrive) UploadFile(ctx context.Context, spaceID, fatherID, ticket, name, base64Content string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	if err := checkLocation(spaceID, ticket, "upload file"); err != nil {
		return "", err
	}
	if strings.TrimSpace(name) == "" || strings.TrimSpace(base64Content) == "" {
		return "", errors.New("wecom: upload file: file_name and content are required")
	}
	request := struct {
		SpaceID  string `json:"spaceid"`
		FatherID string `json:"fatherid,omitempty"`
		Ticket   string `json:"selected_ticket,omitempty"`
		Name     string `json:"file_name"`
		Content  string `json:"file_base64_content"`
	}{SpaceID: spaceID, FatherID: fatherID, Ticket: ticket, Name: name, Content: base64Content}
	var out struct {
		FileID string `json:"fileid"`
	}
	err := c.api.Call(ctx, "wedrive upload file", "/cgi-bin/wedrive/file_upload", request, &out)
	return out.FileID, err
}

// UploadSession is the result of starting a chunked upload. HitExist reports that
// the server already stored an identical file: the upload is complete and
// UploadKey is empty, so no block needs to be sent.
type UploadSession struct {
	HitExist  bool   `json:"hit_exist"`
	UploadKey string `json:"upload_key"`
	FileID    string `json:"fileid"`
}

// UploadInitRequest starts one chunked upload. Set either SpaceID (with an
// optional FatherID) or Ticket from the file picker, never both.
type UploadInitRequest struct {
	SpaceID  string
	FatherID string
	Ticket   string
	FileName string
	Size     int64
	BlockSHA []string
	// SkipPushCard suppresses the WeCom card that is pushed when the file is
	// created. Cards are pushed by default.
	SkipPushCard bool
}

// UploadInit starts a chunked upload. BlockSHA holds one cumulative SHA per 2 MiB
// block, in order; BlockSHA computes it. Send each block with UploadBlock (index
// starts at 1) and finish with UploadFinish.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98004
func (c *WeDrive) UploadInit(ctx context.Context, request UploadInitRequest) (UploadSession, error) {
	if err := check(c); err != nil {
		return UploadSession{}, err
	}
	if err := checkLocation(request.SpaceID, request.Ticket, "upload init"); err != nil {
		return UploadSession{}, err
	}
	if strings.TrimSpace(request.FileName) == "" || request.Size <= 0 || len(request.BlockSHA) == 0 {
		return UploadSession{}, errors.New("wecom: upload init: file_name, size and block_sha are required")
	}
	body := struct {
		SpaceID      string   `json:"spaceid,omitempty"`
		FatherID     string   `json:"fatherid,omitempty"`
		Ticket       string   `json:"selected_ticket,omitempty"`
		Name         string   `json:"file_name"`
		Size         int64    `json:"size"`
		BlockSHA     []string `json:"block_sha"`
		SkipPushCard bool     `json:"skip_push_card,omitempty"`
	}{
		SpaceID: request.SpaceID, FatherID: request.FatherID, Ticket: request.Ticket,
		Name: request.FileName, Size: request.Size, BlockSHA: request.BlockSHA,
		SkipPushCard: request.SkipPushCard,
	}
	var out UploadSession
	if err := c.api.Call(ctx, "wedrive upload init", "/cgi-bin/wedrive/file_upload_init", body, &out); err != nil {
		return UploadSession{}, err
	}
	if !out.HitExist && out.UploadKey == "" {
		return UploadSession{}, errors.New("wecom: upload init: response carried neither upload_key nor hit_exist")
	}
	return out, nil
}

// UploadBlock sends one chunk of a started upload. index is 1-based and must match
// the position of the block SHA passed to UploadInit.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98004
func (c *WeDrive) UploadBlock(ctx context.Context, uploadKey string, index int, base64Content string) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(uploadKey) == "" || index <= 0 || strings.TrimSpace(base64Content) == "" {
		return errors.New("wecom: upload block: upload_key, index and content are required")
	}
	request := struct {
		UploadKey string `json:"upload_key"`
		Index     int    `json:"index"`
		Content   string `json:"file_base64_content"`
	}{UploadKey: uploadKey, Index: index, Content: base64Content}
	return c.api.Call(ctx, "wedrive upload block", "/cgi-bin/wedrive/file_upload_part", request, nil)
}

// UploadFinish completes a chunked upload and returns the new file id.
//
// 文档: https://developer.work.weixin.qq.com/document/path/98004
func (c *WeDrive) UploadFinish(ctx context.Context, uploadKey string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	uploadKey = strings.TrimSpace(uploadKey)
	if uploadKey == "" {
		return "", errors.New("wecom: upload finish: upload_key is required")
	}
	var out struct {
		FileID string `json:"fileid"`
	}
	err := c.api.Call(ctx, "wedrive upload finish", "/cgi-bin/wedrive/file_upload_finish",
		map[string]string{"upload_key": uploadKey}, &out)
	return out.FileID, err
}

// Download is a temporary download address plus the cookie it requires. Three
// same-typed strings would be easy to transpose, so they travel as one value.
type Download struct {
	URL         string `json:"download_url"`
	CookieName  string `json:"cookie_name"`
	CookieValue string `json:"cookie_value"`
}

// DownloadFile returns a temporary download URL and the cookie that must be sent
// with it.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97881
func (c *WeDrive) DownloadFile(ctx context.Context, fileID, ticket string) (Download, error) {
	if err := check(c); err != nil {
		return Download{}, err
	}
	if err := checkLocation(fileID, ticket, "download file"); err != nil {
		return Download{}, err
	}
	request := struct {
		FileID string `json:"fileid"`
		Ticket string `json:"selected_ticket,omitempty"`
	}{FileID: fileID, Ticket: ticket}
	var out Download
	err := c.api.Call(ctx, "wedrive download file", "/cgi-bin/wedrive/file_download", request, &out)
	return out, err
}

// CreateFile creates a folder or document and returns its id and URL.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97882
func (c *WeDrive) CreateFile(ctx context.Context, spaceID, fatherID string, fileType int, name string) (string, string, error) {
	if err := check(c); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(spaceID) == "" || strings.TrimSpace(name) == "" {
		return "", "", errors.New("wecom: create file: spaceid and file_name are required")
	}
	request := struct {
		SpaceID  string `json:"spaceid"`
		FatherID string `json:"fatherid,omitempty"`
		FileType int    `json:"file_type"`
		Name     string `json:"file_name"`
	}{SpaceID: spaceID, FatherID: fatherID, FileType: fileType, Name: name}
	var out struct {
		FileID string `json:"fileid"`
		URL    string `json:"url"`
	}
	err := c.api.Call(ctx, "wedrive create file", "/cgi-bin/wedrive/file_create", request, &out)
	return out.FileID, out.URL, err
}

// RenameFile changes the name of one entry.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97883
func (c *WeDrive) RenameFile(ctx context.Context, fileID, newName string) (File, error) {
	if err := check(c); err != nil {
		return File{}, err
	}
	if strings.TrimSpace(fileID) == "" || strings.TrimSpace(newName) == "" {
		return File{}, errors.New("wecom: rename file: fileid and new_name are required")
	}
	request := struct {
		FileID  string `json:"fileid"`
		NewName string `json:"new_name"`
	}{FileID: fileID, NewName: newName}
	var out struct {
		File File `json:"file"`
	}
	err := c.api.Call(ctx, "wedrive rename file", "/cgi-bin/wedrive/file_rename", request, &out)
	return out.File, err
}

// MoveFiles moves entries into one folder.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97884
func (c *WeDrive) MoveFiles(ctx context.Context, fatherID string, fileIDs []string, replace bool) ([]File, error) {
	if err := check(c); err != nil {
		return nil, err
	}
	if strings.TrimSpace(fatherID) == "" || len(fileIDs) == 0 {
		return nil, errors.New("wecom: move files: fatherid and fileid are required")
	}
	request := struct {
		FatherID string   `json:"fatherid"`
		Replace  bool     `json:"replace,omitempty"`
		FileIDs  []string `json:"fileid"`
	}{FatherID: fatherID, Replace: replace, FileIDs: fileIDs}
	var out fileListEnvelope
	if err := c.api.Call(ctx, "wedrive move files", "/cgi-bin/wedrive/file_move", request, &out); err != nil {
		return nil, err
	}
	return out.List.Items, nil
}

// DeleteFiles removes entries.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97885
func (c *WeDrive) DeleteFiles(ctx context.Context, fileIDs []string) error {
	if err := check(c); err != nil {
		return err
	}
	if len(fileIDs) == 0 {
		return errors.New("wecom: delete files: fileid is empty")
	}
	return c.api.Call(ctx, "wedrive delete files", "/cgi-bin/wedrive/file_delete",
		map[string][]string{"fileid": fileIDs}, nil)
}

// GetFile reads one entry.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97886
func (c *WeDrive) GetFile(ctx context.Context, fileID string) (File, error) {
	if err := check(c); err != nil {
		return File{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return File{}, errors.New("wecom: get file: fileid is required")
	}
	var out struct {
		File File `json:"file_info"`
	}
	err := c.api.Call(ctx, "wedrive get file", "/cgi-bin/wedrive/file_info",
		map[string]string{"fileid": fileID}, &out)
	return out.File, err
}

// AddFileMembers grants access to a file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/93658
func (c *WeDrive) AddFileMembers(ctx context.Context, fileID string, auth []AuthInfo) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(fileID) == "" || len(auth) == 0 {
		return errors.New("wecom: add file members: fileid and auth_info are required")
	}
	if err := checkAuthInfo(auth, true); err != nil {
		return errors.New("wecom: add file members: " + err.Error())
	}
	request := struct {
		FileID   string     `json:"fileid"`
		AuthInfo []AuthInfo `json:"auth_info"`
	}{FileID: fileID, AuthInfo: auth}
	return c.api.Call(ctx, "wedrive add file members", "/cgi-bin/wedrive/file_acl_add", request, nil)
}

// RemoveFileMembers revokes access to a file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97888
func (c *WeDrive) RemoveFileMembers(ctx context.Context, fileID string, auth []AuthInfo) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(fileID) == "" || len(auth) == 0 {
		return errors.New("wecom: remove file members: fileid and auth_info are required")
	}
	if err := checkAuthInfo(auth, false); err != nil {
		return errors.New("wecom: remove file members: " + err.Error())
	}
	request := struct {
		FileID   string     `json:"fileid"`
		AuthInfo []AuthInfo `json:"auth_info"`
	}{FileID: fileID, AuthInfo: auth}
	return c.api.Call(ctx, "wedrive remove file members", "/cgi-bin/wedrive/file_acl_del", request, nil)
}

// SetFileShareScope changes who may open the share link.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97889
func (c *WeDrive) SetFileShareScope(ctx context.Context, fileID string, scope, auth int) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(fileID) == "" {
		return errors.New("wecom: file share scope: fileid is required")
	}
	if scope < 1 || scope > 5 {
		return errors.New("wecom: file share scope: auth_scope must be between 1 and 5")
	}
	// auth is optional: leaving it zero keeps the current permission instead of
	// resetting it, which is what the endpoint documents.
	request := struct {
		FileID    string `json:"fileid"`
		AuthScope int    `json:"auth_scope"`
		Auth      int    `json:"auth,omitempty"`
	}{FileID: fileID, AuthScope: scope, Auth: auth}
	return c.api.Call(ctx, "wedrive file share scope", "/cgi-bin/wedrive/file_setting", request, nil)
}

// FileShareLink returns the share link of one file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97890
func (c *WeDrive) FileShareLink(ctx context.Context, fileID string) (string, error) {
	if err := check(c); err != nil {
		return "", err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return "", errors.New("wecom: file share link: fileid is required")
	}
	var out struct {
		URL string `json:"share_url"`
	}
	err := c.api.Call(ctx, "wedrive file share link", "/cgi-bin/wedrive/file_share",
		map[string]string{"fileid": fileID}, &out)
	return out.URL, err
}

// FilePermission is the sharing and security configuration of one file.
type FilePermission struct {
	ShareRange struct {
		EnableCorpInternal             bool `json:"enable_corp_internal"`
		CorpInternalAuth               int  `json:"corp_internal_auth"`
		EnableCorpExternal             bool `json:"enable_corp_external"`
		CorpExternalAuth               int  `json:"corp_external_auth"`
		CorpInternalApproveOnlyByAdmin bool `json:"corp_internal_approve_only_by_admin"`
		CorpExternalApproveOnlyByAdmin bool `json:"corp_external_approve_only_by_admin"`
	} `json:"share_range"`
	SecureSetting struct {
		EnableReadonlyCopy    bool `json:"enable_readonly_copy"`
		ModifyOnlyByAdmin     bool `json:"modify_only_by_admin"`
		EnableReadonlyComment bool `json:"enable_readonly_comment"`
		BanShareExternal      bool `json:"ban_share_external"`
	} `json:"secure_setting"`
	InheritFatherAuth struct {
		AuthList []AuthInfo `json:"auth_list"`
		Inherit  bool       `json:"inherit"`
	} `json:"inherit_father_auth"`
	// FileMemberList is returned for documents: the members of the containing
	// folder plus the other granted entries.
	FileMemberList []AuthInfo    `json:"file_member_list"`
	Watermark      FileWatermark `json:"watermark"`
}

// FileWatermark is the watermark state reported by GetFilePermission. The two
// force flags are read-only: they record an administrator decision.
type FileWatermark struct {
	Text              string `json:"text"`
	MarginType        int    `json:"margin_type"`
	ShowVisitorName   bool   `json:"show_visitor_name"`
	ForceByAdmin      bool   `json:"force_by_admin"`
	ShowText          bool   `json:"show_text"`
	ForceBySpaceAdmin bool   `json:"force_by_space_admin"`
}

// GetFilePermission reads the sharing and security configuration of one file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97891
func (c *WeDrive) GetFilePermission(ctx context.Context, fileID string) (FilePermission, error) {
	if err := check(c); err != nil {
		return FilePermission{}, err
	}
	fileID = strings.TrimSpace(fileID)
	if fileID == "" {
		return FilePermission{}, errors.New("wecom: file permission: fileid is required")
	}
	var out FilePermission
	err := c.api.Call(ctx, "wedrive file permission", "/cgi-bin/wedrive/get_file_permission",
		map[string]string{"fileid": fileID}, &out)
	return out, err
}

// Watermark changes the watermark of one file. Every field is optional and a nil
// field leaves the current value alone, so an explicit false stays expressible:
// plain bool fields with omitempty silently dropped it, and an empty Text used to
// wipe the watermark text instead of keeping it.
type Watermark struct {
	Text            *string `json:"text,omitempty"`
	MarginType      *int    `json:"margin_type,omitempty"`
	ShowVisitorName *bool   `json:"show_visitor_name,omitempty"`
	ShowText        *bool   `json:"show_text,omitempty"`
}

// SetFileWatermark updates the watermark of one file.
//
// 文档: https://developer.work.weixin.qq.com/document/path/97892
func (c *WeDrive) SetFileWatermark(ctx context.Context, fileID string, watermark Watermark) error {
	if err := check(c); err != nil {
		return err
	}
	if strings.TrimSpace(fileID) == "" {
		return errors.New("wecom: file watermark: fileid is required")
	}
	request := struct {
		FileID    string    `json:"fileid"`
		Watermark Watermark `json:"watermark"`
	}{FileID: fileID, Watermark: watermark}
	return c.api.Call(ctx, "wedrive file watermark", "/cgi-bin/wedrive/file_secure_setting", request, nil)
}

// checkAuthInfo validates the member entries shared by the space and file ACL
// endpoints. Removing a member needs an identity but no permission, while adding
// one needs the permission it grants.
func checkAuthInfo(entries []AuthInfo, needAuth bool) error {
	for index, entry := range entries {
		position := "auth_info[" + strconv.Itoa(index) + "]: "
		if entry.Type != 1 && entry.Type != 2 {
			return errors.New(position + "type must be 1 (user) or 2 (department)")
		}
		if entry.Type == 1 && strings.TrimSpace(entry.UserID) == "" {
			return errors.New(position + "userid is required for a user")
		}
		if entry.Type == 2 && entry.DepartmentID <= 0 {
			return errors.New(position + "departmentid is required for a department")
		}
		if needAuth && entry.Auth <= 0 {
			return errors.New(position + "auth is required")
		}
	}
	return nil
}

// checkLocation enforces the documented "either a space or a picker ticket, and
// only one of them" rule that the upload and download endpoints share.
func checkLocation(spaceID, ticket, op string) error {
	hasSpace := strings.TrimSpace(spaceID) != ""
	hasTicket := strings.TrimSpace(ticket) != ""
	switch {
	case hasSpace && hasTicket:
		return errors.New("wecom: " + op + ": set either spaceid or selected_ticket, not both")
	case !hasSpace && !hasTicket:
		return errors.New("wecom: " + op + ": spaceid or selected_ticket is required")
	}
	return nil
}
