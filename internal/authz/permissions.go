package authz

import "Sixth_world_Suday/internal/role"

type Permission string

const (
	PermAll               Permission = "*"
	PermViewAdminPanel    Permission = "view_admin_panel"
	PermViewStats         Permission = "view_stats"
	PermViewAuditLog      Permission = "view_audit_log"
	PermManageSettings    Permission = "manage_settings"
	PermManageRoles       Permission = "manage_roles"
	PermDeleteAnyUser     Permission = "delete_any_user"
	PermBanUser           Permission = "ban_user"
	PermViewUsers         Permission = "view_users"
	PermDeleteAnyPost     Permission = "delete_any_post"
	PermDeleteAnyComment  Permission = "delete_any_comment"
	PermEditAnyPost       Permission = "edit_any_post"
	PermEditAnyComment    Permission = "edit_any_comment"
	PermManageVanityRoles Permission = "manage_vanity_roles"
	PermManageBannedWords Permission = "manage_banned_words"
	PermResetPassword     Permission = "reset_password"
	PermManageChannels    Permission = "manage_channels"
)

var rolePermissions = map[role.Role][]Permission{
	RoleSuperAdmin: {
		PermAll,
	},
	RoleAdmin: {
		PermAll,
	},
	RoleModerator: {
		PermViewAdminPanel,
		PermViewStats,
		PermViewUsers,
		PermDeleteAnyPost,
		PermDeleteAnyComment,
		PermEditAnyPost,
		PermEditAnyComment,
		PermBanUser,
		PermManageChannels,
	},
}
