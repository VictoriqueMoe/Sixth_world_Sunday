package controllers

import (
	"Sixth_world_Sunday/internal/admin"
	"Sixth_world_Sunday/internal/auth"
	"Sixth_world_Sunday/internal/authz"
	"Sixth_world_Sunday/internal/block"
	"Sixth_world_Sunday/internal/chat"
	"Sixth_world_Sunday/internal/filehost"
	"Sixth_world_Sunday/internal/media"
	"Sixth_world_Sunday/internal/notification"
	"Sixth_world_Sunday/internal/profile"
	"Sixth_world_Sunday/internal/report"
	searchsvc "Sixth_world_Sunday/internal/search"
	"Sixth_world_Sunday/internal/session"
	"Sixth_world_Sunday/internal/settings"
	"Sixth_world_Sunday/internal/upload"
	usersvc "Sixth_world_Sunday/internal/user"
	"Sixth_world_Sunday/internal/vanityrole"
	"Sixth_world_Sunday/internal/ws"
)

type (
	Service struct {
		AuthService         auth.Service
		ProfileService      profile.Service
		NotificationService notification.Service
		AdminService        admin.Service
		AuthzService        authz.Service
		SettingsService     settings.Service
		ChatService         chat.Service
		ReportService       report.Service
		BlockService        block.Service
		UserService         usersvc.Service
		UploadService       upload.Service
		MediaProcessor      *media.Processor
		VanityRoleService   vanityrole.Service
		AuthSession         *session.Manager
		Hub                 *ws.Hub
		SearchService       searchsvc.Service
		FileVaultService    filehost.Service
	}
)

func NewService(
	authService auth.Service,
	profileService profile.Service,
	notificationService notification.Service,
	adminService admin.Service,
	authzService authz.Service,
	settingsService settings.Service,
	chatService chat.Service,
	reportService report.Service,
	blockService block.Service,
	userService usersvc.Service,
	uploadService upload.Service,
	mediaProcessor *media.Processor,
	vanityRoleService vanityrole.Service,
	authSession *session.Manager,
	hub *ws.Hub,
	searchService searchsvc.Service,
	fileVaultService filehost.Service,
) Service {
	return Service{
		AuthService:         authService,
		ProfileService:      profileService,
		NotificationService: notificationService,
		AdminService:        adminService,
		AuthzService:        authzService,
		SettingsService:     settingsService,
		ChatService:         chatService,
		ReportService:       reportService,
		BlockService:        blockService,
		UserService:         userService,
		UploadService:       uploadService,
		MediaProcessor:      mediaProcessor,
		VanityRoleService:   vanityRoleService,
		AuthSession:         authSession,
		Hub:                 hub,
		SearchService:       searchService,
		FileVaultService:    fileVaultService,
	}
}

func (s *Service) GetAPIRoutes() []FSetupRoute {
	var all []FSetupRoute
	all = append(all, s.getAllAuthRoutes()...)
	all = append(all, s.getAllProfileRoutes()...)
	all = append(all, s.getAllNotificationRoutes()...)
	all = append(all, s.getAllAdminRoutes()...)
	all = append(all, s.getAllChatRoutes()...)
	all = append(all, s.getAllReportRoutes()...)
	all = append(all, s.getAllBlockRoutes()...)
	all = append(all, s.getAllUserPreferencesRoutes()...)
	all = append(all, s.getAllSearchRoutes()...)
	all = append(all, s.getAllVaultRoutes()...)
	return all
}

func (s *Service) GetPageRoutes() []FSetupRoute {
	return nil
}
