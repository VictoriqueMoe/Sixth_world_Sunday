package filehost

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/config"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/logger"
	"Sixth_world_Sunday/internal/repository"
	"Sixth_world_Sunday/internal/settings"
	"Sixth_world_Sunday/internal/upload"
	"Sixth_world_Sunday/internal/ws"

	"github.com/google/uuid"
)

type (
	DownloadInfo struct {
		Path string
		Name string
		Mime string
		Size int64
	}

	Service interface {
		Browse(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*dto.VaultBrowseResponse, error)
		CreateFolder(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (*dto.VaultFolderResponse, error)
		RenameFolder(ctx context.Context, userID, id uuid.UUID, name string) error
		DeleteFolder(ctx context.Context, userID, id uuid.UUID) error
		SetFolderLocked(ctx context.Context, userID, id uuid.UUID, locked bool) error
		Upload(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID, originalName string, size int64, reader io.Reader) (*dto.VaultFileResponse, error)
		OpenForDownload(ctx context.Context, userID, id uuid.UUID) (*DownloadInfo, error)
		RenameFile(ctx context.Context, userID, id uuid.UUID, name string) error
		DeleteFile(ctx context.Context, userID, id uuid.UUID) error
		SetFileLocked(ctx context.Context, userID, id uuid.UUID, locked bool) error
	}

	service struct {
		repo        repository.VaultRepository
		authz       authz.Service
		settingsSvc settings.Service
		hub         *ws.Hub
		vaultDir    string
	}
)

func NewService(repo repository.VaultRepository, authzSvc authz.Service, settingsSvc settings.Service, hub *ws.Hub, vaultDir string) Service {
	return &service{
		repo:        repo,
		authz:       authzSvc,
		settingsSvc: settingsSvc,
		hub:         hub,
		vaultDir:    vaultDir,
	}
}

func (s *service) broadcast(action string) {
	s.hub.Broadcast(ws.Message{
		Type: "vault_changed",
		Data: map[string]any{"action": action},
	})
}

func (s *service) privileged(ctx context.Context, userID uuid.UUID) bool {
	return s.authz.Can(ctx, userID, authz.PermLockFiles)
}

func (s *service) isStaff(ctx context.Context, userID uuid.UUID) bool {
	r, err := s.authz.GetRole(ctx, userID)
	if err != nil {
		return false
	}

	return r.IsSiteStaff()
}

func (s *service) folderAccessible(ctx context.Context, userID uuid.UUID, folderID uuid.UUID) (bool, error) {
	if s.privileged(ctx, userID) {
		return true, nil
	}

	locked, err := s.repo.ChainLocked(ctx, folderID)
	if err != nil {
		return false, err
	}

	return !locked, nil
}

func (s *service) fileAccessible(ctx context.Context, userID uuid.UUID, file *repository.VaultFile) (bool, error) {
	if s.privileged(ctx, userID) {
		return true, nil
	}

	if file.Locked {
		return false, nil
	}

	if file.FolderID != nil {
		locked, err := s.repo.ChainLocked(ctx, *file.FolderID)
		if err != nil {
			return false, err
		}
		return !locked, nil
	}

	return true, nil
}

func (s *service) canManageFolder(ctx context.Context, userID uuid.UUID, folder *repository.VaultFolder) bool {
	if folder.CreatedBy == userID {
		return true
	}

	return s.isStaff(ctx, userID)
}

func (s *service) Browse(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID) (*dto.VaultBrowseResponse, error) {
	priv := s.privileged(ctx, userID)

	resp := &dto.VaultBrowseResponse{CanManageLocks: priv}

	if folderID != nil {
		folder, err := s.repo.GetFolder(ctx, *folderID)
		if err != nil {
			return nil, err
		}
		if folder == nil {
			return nil, ErrFolderNotFound
		}

		if !priv {
			locked, err := s.repo.ChainLocked(ctx, *folderID)
			if err != nil {
				return nil, err
			}
			if locked {
				return nil, ErrFolderNotFound
			}
		}

		chain, err := s.repo.FolderChain(ctx, *folderID)
		if err != nil {
			return nil, err
		}
		crumbs := make([]dto.VaultBreadcrumb, len(chain))
		for i := 0; i < len(chain); i++ {
			crumbs[i] = dto.VaultBreadcrumb{ID: chain[i].ID.String(), Name: chain[i].Name}
		}
		resp.Breadcrumbs = crumbs

		fd := folderToDTO(*folder)
		resp.Folder = &fd
	}

	childFolders, err := s.repo.ListChildFolders(ctx, folderID, priv)
	if err != nil {
		return nil, err
	}
	resp.Folders = make([]dto.VaultFolderResponse, len(childFolders))
	for i := 0; i < len(childFolders); i++ {
		resp.Folders[i] = folderToDTO(childFolders[i])
	}

	files, err := s.repo.ListFolderFiles(ctx, folderID, priv)
	if err != nil {
		return nil, err
	}
	resp.Files = make([]dto.VaultFileResponse, len(files))
	for i := 0; i < len(files); i++ {
		resp.Files[i] = fileToDTO(files[i])
	}

	return resp, nil
}

func (s *service) CreateFolder(ctx context.Context, userID uuid.UUID, name string, parentID *uuid.UUID) (*dto.VaultFolderResponse, error) {
	clean := sanitizeName(name)
	if clean == "" {
		return nil, ErrInvalidName
	}

	if parentID != nil {
		parent, err := s.repo.GetFolder(ctx, *parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrFolderNotFound
		}

		ok, err := s.folderAccessible(ctx, userID, *parentID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrFolderNotFound
		}
	}

	folder := &repository.VaultFolder{
		ID:        uuid.New(),
		ParentID:  parentID,
		Name:      clean,
		CreatedBy: userID,
	}
	if err := s.repo.CreateFolder(ctx, folder); err != nil {
		return nil, err
	}

	s.broadcast("create_folder")

	resp := folderToDTO(*folder)
	return &resp, nil
}

func (s *service) RenameFolder(ctx context.Context, userID, id uuid.UUID, name string) error {
	clean := sanitizeName(name)
	if clean == "" {
		return ErrInvalidName
	}

	folder, err := s.repo.GetFolder(ctx, id)
	if err != nil {
		return err
	}
	if folder == nil {
		return ErrFolderNotFound
	}

	ok, err := s.folderAccessible(ctx, userID, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFolderNotFound
	}

	if !s.canManageFolder(ctx, userID, folder) {
		return ErrForbidden
	}

	if err := s.repo.RenameFolder(ctx, id, clean); err != nil {
		return err
	}

	s.broadcast("rename_folder")
	return nil
}

func (s *service) DeleteFolder(ctx context.Context, userID, id uuid.UUID) error {
	folder, err := s.repo.GetFolder(ctx, id)
	if err != nil {
		return err
	}
	if folder == nil {
		return ErrFolderNotFound
	}

	ok, err := s.folderAccessible(ctx, userID, id)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFolderNotFound
	}

	if !s.canManageFolder(ctx, userID, folder) {
		return ErrForbidden
	}

	stored, err := s.repo.DescendantStoredNames(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteFolder(ctx, id); err != nil {
		return err
	}

	s.removeStored(stored)
	s.broadcast("delete_folder")

	return nil
}

func (s *service) SetFolderLocked(ctx context.Context, userID, id uuid.UUID, locked bool) error {
	if !s.privileged(ctx, userID) {
		return ErrForbidden
	}

	folder, err := s.repo.GetFolder(ctx, id)
	if err != nil {
		return err
	}
	if folder == nil {
		return ErrFolderNotFound
	}

	if err := s.repo.SetFolderLocked(ctx, id, locked); err != nil {
		return err
	}

	action := "unlock_folder"
	if locked {
		action = "lock_folder"
	}
	s.broadcast(action)
	return nil
}

func (s *service) Upload(ctx context.Context, userID uuid.UUID, folderID *uuid.UUID, originalName string, size int64, reader io.Reader) (*dto.VaultFileResponse, error) {
	maxSize := int64(s.settingsSvc.GetInt(ctx, config.SettingMaxGeneralSize))
	if maxSize > 0 && size > maxSize {
		return nil, ErrFileTooLarge
	}

	if folderID != nil {
		folder, err := s.repo.GetFolder(ctx, *folderID)
		if err != nil {
			return nil, err
		}
		if folder == nil {
			return nil, ErrFolderNotFound
		}

		ok, err := s.folderAccessible(ctx, userID, *folderID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrFolderNotFound
		}
	}

	name := sanitizeName(originalName)
	if name == "" {
		name = "file"
	}

	mime, wrapped, err := upload.DetectContentType(reader)
	if err != nil {
		return nil, fmt.Errorf("detect content type: %w", err)
	}

	fileID := uuid.New()
	storedName := fileID.String() + safeExt(name)

	if err := s.writeBytes(storedName, wrapped); err != nil {
		return nil, err
	}

	file := &repository.VaultFile{
		ID:           fileID,
		FolderID:     folderID,
		OriginalName: name,
		StoredName:   storedName,
		Mime:         mime,
		Size:         size,
		UploadedBy:   userID,
	}
	if err := s.repo.CreateFile(ctx, file); err != nil {
		s.removeStored([]string{storedName})
		return nil, err
	}

	s.broadcast("upload")

	resp := fileToDTO(*file)
	return &resp, nil
}

func (s *service) OpenForDownload(ctx context.Context, userID, id uuid.UUID) (*DownloadInfo, error) {
	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}

	ok, err := s.fileAccessible(ctx, userID, file)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFileNotFound
	}

	return &DownloadInfo{
		Path: filepath.Join(s.vaultDir, file.StoredName),
		Name: file.OriginalName,
		Mime: file.Mime,
		Size: file.Size,
	}, nil
}

func (s *service) RenameFile(ctx context.Context, userID, id uuid.UUID, name string) error {
	clean := sanitizeName(name)
	if clean == "" {
		return ErrInvalidName
	}

	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNotFound
	}

	ok, err := s.fileAccessible(ctx, userID, file)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFileNotFound
	}

	if file.UploadedBy != userID && !s.isStaff(ctx, userID) {
		return ErrForbidden
	}

	if err := s.repo.RenameFile(ctx, id, clean); err != nil {
		return err
	}

	s.broadcast("rename_file")
	return nil
}

func (s *service) DeleteFile(ctx context.Context, userID, id uuid.UUID) error {
	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNotFound
	}

	ok, err := s.fileAccessible(ctx, userID, file)
	if err != nil {
		return err
	}
	if !ok {
		return ErrFileNotFound
	}

	if file.UploadedBy != userID && !s.isStaff(ctx, userID) {
		return ErrForbidden
	}

	if err := s.repo.DeleteFile(ctx, id); err != nil {
		return err
	}

	s.removeStored([]string{file.StoredName})
	s.broadcast("delete_file")

	return nil
}

func (s *service) SetFileLocked(ctx context.Context, userID, id uuid.UUID, locked bool) error {
	if !s.privileged(ctx, userID) {
		return ErrForbidden
	}

	file, err := s.repo.GetFile(ctx, id)
	if err != nil {
		return err
	}
	if file == nil {
		return ErrFileNotFound
	}

	if err := s.repo.SetFileLocked(ctx, id, locked); err != nil {
		return err
	}

	action := "unlock_file"
	if locked {
		action = "lock_file"
	}
	s.broadcast(action)
	return nil
}

func (s *service) writeBytes(storedName string, reader io.Reader) error {
	if err := os.MkdirAll(s.vaultDir, 0755); err != nil {
		return fmt.Errorf("create vault dir: %w", err)
	}

	dst, err := os.Create(filepath.Join(s.vaultDir, storedName))
	if err != nil {
		return fmt.Errorf("create vault file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("write vault file: %w", err)
	}

	return nil
}

func (s *service) removeStored(names []string) {
	for i := 0; i < len(names); i++ {
		path := filepath.Join(s.vaultDir, names[i])
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			logger.Log.Warn().Err(err).Str("path", path).Msg("failed to remove vault file")
		}
	}
}

func folderToDTO(f repository.VaultFolder) dto.VaultFolderResponse {
	var parent *string
	if f.ParentID != nil {
		p := f.ParentID.String()
		parent = &p
	}

	return dto.VaultFolderResponse{
		ID:        f.ID.String(),
		ParentID:  parent,
		Name:      f.Name,
		Locked:    f.Locked,
		CreatedBy: f.CreatedBy.String(),
		CreatedAt: f.CreatedAt,
	}
}

func fileToDTO(f repository.VaultFile) dto.VaultFileResponse {
	var folder *string
	if f.FolderID != nil {
		fid := f.FolderID.String()
		folder = &fid
	}

	return dto.VaultFileResponse{
		ID:         f.ID.String(),
		FolderID:   folder,
		Name:       f.OriginalName,
		Mime:       f.Mime,
		Size:       f.Size,
		Locked:     f.Locked,
		UploadedBy: f.UploadedBy.String(),
		CreatedAt:  f.CreatedAt,
	}
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)

	var b strings.Builder
	count := 0
	for _, r := range name {
		if r == '/' || r == '\\' || r < 0x20 || r == 0x7f {
			continue
		}

		b.WriteRune(r)
		count++

		if count >= 255 {
			break
		}
	}

	return strings.TrimSpace(b.String())
}

func safeExt(name string) string {
	ext := filepath.Ext(name)
	if len(ext) > 16 {
		return ""
	}

	for _, r := range ext {
		if r == '.' {
			continue
		}
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') {
			return ""
		}
	}

	return strings.ToLower(ext)
}
