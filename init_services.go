package main

import (
	"context"
	"os"
	"path/filepath"

	"Sixth_world_Suday/internal/admin"
	"Sixth_world_Suday/internal/auth"
	"Sixth_world_Suday/internal/authz"
	blocksvc "Sixth_world_Suday/internal/block"
	"Sixth_world_Suday/internal/chat"
	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/contentfilter"
	slursrule "Sixth_world_Suday/internal/contentfilter/rules/slurs"
	"Sixth_world_Suday/internal/email"
	"Sixth_world_Suday/internal/livekit"
	"Sixth_world_Suday/internal/logger"
	"Sixth_world_Suday/internal/media"
	"Sixth_world_Suday/internal/notification"
	"Sixth_world_Suday/internal/profile"
	"Sixth_world_Suday/internal/report"
	"Sixth_world_Suday/internal/repository"
	searchsvc "Sixth_world_Suday/internal/search"
	"Sixth_world_Suday/internal/session"
	"Sixth_world_Suday/internal/settings"
	"Sixth_world_Suday/internal/upload"
	"Sixth_world_Suday/internal/user"
	"Sixth_world_Suday/internal/vanityrole"
	"Sixth_world_Suday/internal/ws"
)

func initServices(repos *repository.Repositories, settingsSvc settings.Service) *services {
	uploadDir := settingsSvc.Get(context.Background(), config.SettingUploadDir)
	for _, sub := range []string{"avatars", "banners", "media"} {
		if err := os.MkdirAll(filepath.Join(uploadDir, sub), 0755); err != nil {
			logger.Log.Fatal().Err(err).Msgf("failed to create %s directory", sub)
		}
	}

	sessionMgr := session.NewManager(repos.Session, settingsSvc)
	mediaProc := media.NewProcessor(4)
	uploadSvc := upload.NewService(settingsSvc, mediaProc)
	authzSvc := authz.NewService(repos.Role, repos.User)

	contentFilter := contentfilter.New(
		slursrule.New(),
	)
	userSvc := user.NewService(repos.User, repos.Role, authzSvc)
	hub := ws.NewHub()

	emailSvc := email.NewService(settingsSvc)
	blockSvc := blocksvc.NewService(repos.Block, authzSvc)
	notifSvc := notification.NewService(repos.Notification, repos.User, hub, emailSvc)
	reportSvc := report.NewService(repos.Report, repos.Role, repos.User, notifSvc, settingsSvc)
	livekitSvc := livekit.NewService(settingsSvc)
	chatSvc := chat.NewService(repos.Chat, repos.User, repos.Role, repos.VanityRole, repos.ChatRoomBan, repos.ChatBannedWord, repos.AuditLog, authzSvc, notifSvc, blockSvc, uploadSvc, settingsSvc, mediaProc, hub, livekitSvc, contentFilter)
	vanityRoleSvc := vanityrole.NewService(repos.VanityRole)
	searchSvc := searchsvc.NewService(repos.Search, repos.Chat)

	authSvc := auth.NewService(userSvc, sessionMgr, settingsSvc, repos.Invite, repos.User, repos.AuditLog, repos.PasswordReset, repos.EmailVerification, emailSvc, contentFilter)

	return &services{
		settings:      settingsSvc,
		auth:          authSvc,
		profile:       profile.NewService(repos.User, authzSvc, uploadSvc, settingsSvc, contentFilter, hub, authSvc),
		notification:  notifSvc,
		admin:         admin.NewService(repos.User, repos.Role, repos.Stats, repos.AuditLog, repos.Invite, repos.VanityRole, authzSvc, settingsSvc, sessionMgr, uploadSvc, hub, chatSvc, emailSvc),
		authz:         authzSvc,
		chat:          chatSvc,
		report:        reportSvc,
		block:         blockSvc,
		email:         emailSvc,
		session:       sessionMgr,
		upload:        uploadSvc,
		hub:           hub,
		mediaProc:     mediaProc,
		contentFilter: contentFilter,
		vanityRole:    vanityRoleSvc,
		search:        searchSvc,
		user:          userSvc,
	}
}
