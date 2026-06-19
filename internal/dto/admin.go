package dto

import (
	"Sixth_world_Suday/internal/role"

	"github.com/google/uuid"
)

type (
	AdminUserItem struct {
		ID          uuid.UUID `json:"id"`
		Username    string    `json:"username"`
		DisplayName string    `json:"display_name"`
		AvatarURL   string    `json:"avatar_url"`
		Role        role.Role `json:"role,omitempty"`
		Banned      bool      `json:"banned"`
		Locked      bool      `json:"locked"`
		CreatedAt   string    `json:"created_at"`
	}

	AdminUserListResponse struct {
		Users  []AdminUserItem `json:"users"`
		Total  int             `json:"total"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}

	AdminUserDetailResponse struct {
		AdminUserItem
		Email                  string        `json:"email,omitempty"`
		IP                     string        `json:"ip,omitempty"`
		BanReason              string        `json:"ban_reason,omitempty"`
		BannedAt               string        `json:"banned_at,omitempty"`
		BannedBy               *UserResponse `json:"banned_by,omitempty"`
		LockReason             string        `json:"lock_reason,omitempty"`
		LockedAt               string        `json:"locked_at,omitempty"`
	}

	AdminStatsResponse struct {
		TotalUsers      int              `json:"total_users"`
		TotalMessages   int              `json:"total_messages"`
		TotalRooms      int              `json:"total_rooms"`
		NewUsers24h     int              `json:"new_users_24h"`
		NewUsers7d      int              `json:"new_users_7d"`
		NewUsers30d     int              `json:"new_users_30d"`
		NewMessages24h  int              `json:"new_messages_24h"`
		NewMessages7d   int              `json:"new_messages_7d"`
		NewMessages30d  int              `json:"new_messages_30d"`
		MostActiveUsers []MostActiveUser `json:"most_active_users"`
	}

	MostActiveUser struct {
		ID          uuid.UUID `json:"id"`
		Username    string    `json:"username"`
		DisplayName string    `json:"display_name"`
		AvatarURL   string    `json:"avatar_url"`
		ActionCount int       `json:"action_count"`
	}

	AuditLogEntryResponse struct {
		ID         int       `json:"id"`
		ActorID    uuid.UUID `json:"actor_id"`
		ActorName  string    `json:"actor_name"`
		Action     string    `json:"action"`
		TargetType string    `json:"target_type"`
		TargetID   string    `json:"target_id"`
		Details    string    `json:"details"`
		CreatedAt  string    `json:"created_at"`
	}

	AuditLogListResponse struct {
		Entries []AuditLogEntryResponse `json:"entries"`
		Total   int                     `json:"total"`
		Limit   int                     `json:"limit"`
		Offset  int                     `json:"offset"`
	}

	SettingsResponse struct {
		Settings map[string]string `json:"settings"`
	}

	UpdateSettingsRequest struct {
		Settings map[string]string `json:"settings"`
	}

	SetRoleRequest struct {
		Role string `json:"role"`
	}

	BanUserRequest struct {
		Reason string `json:"reason"`
	}

	LockUserRequest struct {
		Reason string `json:"reason"`
	}

	AdminResetPasswordResponse struct {
		Password string `json:"password"`
	}

	InviteResponse struct {
		Code      string     `json:"code"`
		CreatedBy uuid.UUID  `json:"created_by"`
		UsedBy    *uuid.UUID `json:"used_by,omitempty"`
		UsedAt    *string    `json:"used_at,omitempty"`
		CreatedAt string     `json:"created_at"`
	}

	InviteListResponse struct {
		Invites []InviteResponse `json:"invites"`
		Total   int              `json:"total"`
		Limit   int              `json:"limit"`
		Offset  int              `json:"offset"`
	}

	VanityRoleResponse struct {
		ID        string `json:"id"`
		Label     string `json:"label"`
		Color     string `json:"color"`
		IsSystem  bool   `json:"is_system"`
		SortOrder int    `json:"sort_order"`
	}

	VanityRoleUsersResponse struct {
		Users  []VanityRoleUserItem `json:"users"`
		Total  int                  `json:"total"`
		Limit  int                  `json:"limit"`
		Offset int                  `json:"offset"`
	}

	VanityRoleUserItem struct {
		ID          uuid.UUID `json:"id"`
		Username    string    `json:"username"`
		DisplayName string    `json:"display_name"`
		AvatarURL   string    `json:"avatar_url"`
	}

	CreateVanityRoleRequest struct {
		Label     string `json:"label"`
		Color     string `json:"color"`
		SortOrder int    `json:"sort_order"`
	}

	UpdateVanityRoleRequest struct {
		Label     string `json:"label"`
		Color     string `json:"color"`
		SortOrder int    `json:"sort_order"`
	}

	AssignVanityRoleRequest struct {
		UserID string `json:"user_id"`
	}

	SiteInfoResponse struct {
		SiteName              string               `json:"site_name"`
		SiteDescription       string               `json:"site_description"`
		RegistrationType      string               `json:"registration_type"`
		AnnouncementBanner    string               `json:"announcement_banner"`
		DefaultTheme          string               `json:"default_theme"`
		MaintenanceMode       bool                 `json:"maintenance_mode"`
		MaintenanceTitle      string               `json:"maintenance_title"`
		MaintenanceMessage    string               `json:"maintenance_message"`
		TurnstileEnabled      bool                 `json:"turnstile_enabled"`
		TurnstileSiteKey      string               `json:"turnstile_site_key"`
		VoiceEnabled          bool                 `json:"voice_enabled"`
		EmailEnabled          bool                 `json:"email_enabled"`
		MaxImageSize          int                  `json:"max_image_size"`
		MaxVideoSize          int                  `json:"max_video_size"`
		VanityRoles           []SiteInfoVanityRole `json:"vanity_roles"`
		VanityRoleAssignments map[string][]string  `json:"vanity_role_assignments"`
		RulesPage             string               `json:"rules_page"`
		Version               string               `json:"version"`
	}

	SiteInfoVanityRole struct {
		ID        string `json:"id"`
		Label     string `json:"label"`
		Color     string `json:"color"`
		IsSystem  bool   `json:"is_system"`
		SortOrder int    `json:"sort_order"`
	}
)
