package controllers

import (
	"errors"
	"fmt"
	"os"

	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/controllers/utils"
	"Sixth_world_Sunday/internal/dto"
	"Sixth_world_Sunday/internal/filehost"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

func (s *Service) getAllVaultRoutes() []FSetupRoute {
	return []FSetupRoute{
		s.setupVaultBrowse,
		s.setupVaultCreateFolder,
		s.setupVaultRenameFolder,
		s.setupVaultDeleteFolder,
		s.setupVaultLockFolder,
		s.setupVaultUnlockFolder,
		s.setupVaultUpload,
		s.setupVaultDownload,
		s.setupVaultRenameFile,
		s.setupVaultDeleteFile,
		s.setupVaultLockFile,
		s.setupVaultUnlockFile,
	}
}

func (s *Service) setupVaultBrowse(r fiber.Router) {
	r.Get("/files/contents", s.vaultBrowse)
}

func (s *Service) setupVaultCreateFolder(r fiber.Router) {
	r.Post("/files/folders", s.vaultCreateFolder)
}

func (s *Service) setupVaultRenameFolder(r fiber.Router) {
	r.Patch("/files/folders/:id", s.vaultRenameFolder)
}

func (s *Service) setupVaultDeleteFolder(r fiber.Router) {
	r.Delete("/files/folders/:id", s.vaultDeleteFolder)
}

func (s *Service) setupVaultLockFolder(r fiber.Router) {
	r.Post("/files/folders/:id/lock", s.requirePerm(authz.PermLockFiles), s.vaultLockFolder)
}

func (s *Service) setupVaultUnlockFolder(r fiber.Router) {
	r.Post("/files/folders/:id/unlock", s.requirePerm(authz.PermLockFiles), s.vaultUnlockFolder)
}

func (s *Service) setupVaultUpload(r fiber.Router) {
	r.Post("/files/upload", s.vaultUpload)
}

func (s *Service) setupVaultDownload(r fiber.Router) {
	r.Get("/files/items/:id/download", s.vaultDownload)
}

func (s *Service) setupVaultRenameFile(r fiber.Router) {
	r.Patch("/files/items/:id", s.vaultRenameFile)
}

func (s *Service) setupVaultDeleteFile(r fiber.Router) {
	r.Delete("/files/items/:id", s.vaultDeleteFile)
}

func (s *Service) setupVaultLockFile(r fiber.Router) {
	r.Post("/files/items/:id/lock", s.requirePerm(authz.PermLockFiles), s.vaultLockFile)
}

func (s *Service) setupVaultUnlockFile(r fiber.Router) {
	r.Post("/files/items/:id/unlock", s.requirePerm(authz.PermLockFiles), s.vaultUnlockFile)
}

func optionalFolderQuery(ctx fiber.Ctx) (*uuid.UUID, bool) {
	q := ctx.Query("folder")
	if q == "" {
		return nil, true
	}

	id, err := uuid.Parse(q)
	if err != nil {
		_ = utils.BadRequest(ctx, "invalid folder id")
		return nil, false
	}

	return &id, true
}

func parseOptionalID(ctx fiber.Ctx, raw *string) (*uuid.UUID, bool) {
	if raw == nil || *raw == "" {
		return nil, true
	}

	id, err := uuid.Parse(*raw)
	if err != nil {
		_ = utils.BadRequest(ctx, "invalid id")
		return nil, false
	}

	return &id, true
}

func (s *Service) vaultBrowse(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	folderID, ok := optionalFolderQuery(ctx)
	if !ok {
		return nil
	}

	res, err := s.FileVaultService.Browse(ctx.Context(), userID, folderID)
	if err != nil {
		return handleVaultError(ctx, err)
	}

	return ctx.JSON(res)
}

func (s *Service) vaultCreateFolder(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	req, ok := utils.BindJSON[dto.CreateVaultFolderRequest](ctx)
	if !ok {
		return nil
	}

	parentID, ok := parseOptionalID(ctx, req.ParentID)
	if !ok {
		return nil
	}

	res, err := s.FileVaultService.CreateFolder(ctx.Context(), userID, req.Name, parentID)
	if err != nil {
		return handleVaultError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (s *Service) vaultRenameFolder(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	req, ok := utils.BindJSON[dto.RenameVaultRequest](ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.RenameFolder(ctx.Context(), userID, id, req.Name); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) vaultDeleteFolder(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.DeleteFolder(ctx.Context(), userID, id); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) vaultLockFolder(ctx fiber.Ctx) error {
	return s.setFolderLocked(ctx, true)
}

func (s *Service) vaultUnlockFolder(ctx fiber.Ctx) error {
	return s.setFolderLocked(ctx, false)
}

func (s *Service) setFolderLocked(ctx fiber.Ctx, locked bool) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.SetFolderLocked(ctx.Context(), userID, id, locked); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) vaultUpload(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return utils.BadRequest(ctx, "file is required")
	}

	var folderID *uuid.UUID
	if v := ctx.FormValue("folderId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return utils.BadRequest(ctx, "invalid folder id")
		}
		folderID = &id
	}

	src, err := fileHeader.Open()
	if err != nil {
		return utils.BadRequest(ctx, "failed to read file")
	}
	defer src.Close()

	res, err := s.FileVaultService.Upload(ctx.Context(), userID, folderID, fileHeader.Filename, fileHeader.Size, src)
	if err != nil {
		return handleVaultError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (s *Service) vaultDownload(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	info, err := s.FileVaultService.OpenForDownload(ctx.Context(), userID, id)
	if err != nil {
		return handleVaultError(ctx, err)
	}

	f, err := os.Open(info.Path)
	if err != nil {
		return utils.InternalError(ctx, "failed to open file", err)
	}

	ctx.Set(fiber.HeaderContentType, info.Mime)
	ctx.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", info.Name))

	return ctx.SendStream(f, int(info.Size))
}

func (s *Service) vaultRenameFile(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	req, ok := utils.BindJSON[dto.RenameVaultRequest](ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.RenameFile(ctx.Context(), userID, id, req.Name); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) vaultDeleteFile(ctx fiber.Ctx) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.DeleteFile(ctx.Context(), userID, id); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func (s *Service) vaultLockFile(ctx fiber.Ctx) error {
	return s.setFileLocked(ctx, true)
}

func (s *Service) vaultUnlockFile(ctx fiber.Ctx) error {
	return s.setFileLocked(ctx, false)
}

func (s *Service) setFileLocked(ctx fiber.Ctx, locked bool) error {
	userID := utils.UserID(ctx)

	id, ok := utils.ParseID(ctx)
	if !ok {
		return nil
	}

	if err := s.FileVaultService.SetFileLocked(ctx.Context(), userID, id, locked); err != nil {
		return handleVaultError(ctx, err)
	}

	return utils.OK(ctx)
}

func handleVaultError(ctx fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, filehost.ErrFolderNotFound):
		return utils.NotFound(ctx, "folder not found")
	case errors.Is(err, filehost.ErrFileNotFound):
		return utils.NotFound(ctx, "file not found")
	case errors.Is(err, filehost.ErrForbidden):
		return utils.Forbidden(ctx, "you cannot modify this item")
	case errors.Is(err, filehost.ErrInvalidName):
		return utils.BadRequest(ctx, "invalid name")
	case errors.Is(err, filehost.ErrFileTooLarge):
		return utils.TooLarge(ctx, "file too large")
	default:
		return utils.InternalError(ctx, "file vault error", err)
	}
}
