package controllers

import (
	"Sixth_world_Suday/internal/admin"
	"Sixth_world_Suday/internal/auth"
	"Sixth_world_Suday/internal/authz"
	"Sixth_world_Suday/internal/block"
	"Sixth_world_Suday/internal/chat"
	"Sixth_world_Suday/internal/media"
	"Sixth_world_Suday/internal/notification"
	"Sixth_world_Suday/internal/profile"
	"Sixth_world_Suday/internal/report"
	searchsvc "Sixth_world_Suday/internal/search"
	"Sixth_world_Suday/internal/session"
	"Sixth_world_Suday/internal/settings"
	"Sixth_world_Suday/internal/upload"
	usersvc "Sixth_world_Suday/internal/user"
	"Sixth_world_Suday/internal/vanityrole"
	"Sixth_world_Suday/internal/ws"
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
	return all
}

func (s *Service) GetPageRoutes() []FSetupRoute {
	return nil
}
