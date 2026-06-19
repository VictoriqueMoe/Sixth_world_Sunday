export type SiteRole = "super_admin" | "admin" | "moderator";

export const ROLE_GROUPS: { role: SiteRole; label: string }[] = [
    { role: "super_admin", label: "Sysops" },
    { role: "admin", label: "Admins" },
    { role: "moderator", label: "Moderators" },
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
    | "reset_password";

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
    ],
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
