export type SiteRole = "super_admin" | "admin" | "moderator" | "gm";

export const ROLE_GROUPS: { role: SiteRole; label: string }[] = [
    { role: "super_admin", label: "Sysops" },
    { role: "admin", label: "Admins" },
    { role: "moderator", label: "Moderators" },
    { role: "gm", label: "Game Masters" },
];

export function isSiteStaff(role: SiteRole | undefined | null): boolean {
    return role === "super_admin" || role === "admin" || role === "moderator";
}

export type Permission =
    | "ban_user"
    | "manage_roles"
    | "view_admin_panel"
    | "manage_settings"
    | "view_audit_log"
    | "view_stats"
    | "view_users"
    | "delete_any_user"
    | "delete_any_post"
    | "delete_any_comment"
    | "edit_any_post"
    | "edit_any_comment"
    | "manage_vanity_roles"
    | "manage_banned_words"
    | "manage_channels"
    | "reset_password"
    | "lock_files"
    | "manage_events"
    | "manage_maps";

const rolePermissions: Record<string, Permission[]> = {
    super_admin: [
        "ban_user",
        "manage_roles",
        "view_admin_panel",
        "manage_settings",
        "view_audit_log",
        "view_stats",
        "view_users",
        "delete_any_user",
        "delete_any_post",
        "delete_any_comment",
        "edit_any_post",
        "edit_any_comment",
        "manage_vanity_roles",
        "manage_banned_words",
        "manage_channels",
        "reset_password",
        "lock_files",
        "manage_events",
        "manage_maps",
    ],
    admin: [
        "ban_user",
        "manage_roles",
        "view_admin_panel",
        "manage_settings",
        "view_audit_log",
        "view_stats",
        "view_users",
        "delete_any_user",
        "delete_any_post",
        "delete_any_comment",
        "edit_any_post",
        "edit_any_comment",
        "manage_vanity_roles",
        "manage_banned_words",
        "manage_channels",
        "reset_password",
        "lock_files",
        "manage_events",
        "manage_maps",
    ],
    moderator: [
        "delete_any_post",
        "delete_any_comment",
        "edit_any_post",
        "edit_any_comment",
        "view_admin_panel",
        "view_stats",
        "view_users",
        "ban_user",
        "manage_channels",
        "lock_files",
        "manage_events",
        "manage_maps",
    ],
    gm: ["lock_files", "manage_events", "manage_maps"],
};

export function can(role: SiteRole | undefined, perm: Permission): boolean {
    if (!role) {
        return false;
    }
    return rolePermissions[role]?.includes(perm) ?? false;
}

export function canAccessAdmin(role: SiteRole | undefined): boolean {
    return can(role, "view_admin_panel");
}

export function canManageEvents(role: SiteRole | undefined): boolean {
    return can(role, "manage_events");
}

export function canManageMaps(role: SiteRole | undefined): boolean {
    return can(role, "manage_maps");
}
