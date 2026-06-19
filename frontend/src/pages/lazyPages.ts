import { type ComponentType, lazy } from "react";

// React.lazy only accepts modules with a `default` export, but our pages use
// named exports. `named` does that adapter so each page below fits on one line.
// The `any` mirrors React's own `lazy<T extends ComponentType<any>>` signature —
// you can't satisfy that constraint with `never`, `unknown`, or `object` because
// of how props variance interacts with class components inside ComponentType.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function named<T extends ComponentType<any>, K extends string>(loader: () => Promise<Record<K, T>>, name: K) {
    return lazy(() => loader().then(m => ({ default: m[name] })));
}

//  Auth
export const LoginPage = named(() => import("./auth/LoginPage"), "LoginPage");
export const ForgotPasswordPage = named(() => import("./auth/ForgotPasswordPage"), "ForgotPasswordPage");
export const ResetPasswordPage = named(() => import("./auth/ResetPasswordPage"), "ResetPasswordPage");
export const SetEmailPage = named(() => import("./auth/SetEmailPage"), "SetEmailPage");
export const VerifyEmailPage = named(() => import("./auth/VerifyEmailPage"), "VerifyEmailPage");

//  Profile
export const ProfilePage = named(() => import("./profile/ProfilePage"), "ProfilePage");
export const SettingsPage = named(() => import("./profile/SettingsPage"), "SettingsPage");

//  Admin
export const AdminLayout = named(() => import("./admin/AdminLayout"), "AdminLayout");
export const AdminDashboard = named(() => import("./admin/AdminDashboard"), "AdminDashboard");
export const AdminUsers = named(() => import("./admin/AdminUsers"), "AdminUsers");
export const AdminUserDetail = named(() => import("./admin/AdminUserDetail"), "AdminUserDetail");
export const AdminSettings = named(() => import("./admin/AdminSettings"), "AdminSettings");
export const AdminAuditLog = named(() => import("./admin/AdminAuditLog"), "AdminAuditLog");
export const AdminInvites = named(() => import("./admin/AdminInvites"), "AdminInvites");
export const AdminReports = named(() => import("./admin/AdminReports"), "AdminReports");
export const AdminContentRules = named(() => import("./admin/AdminContentRules"), "AdminContentRules");
export const AdminVanityRoles = named(() => import("./admin/AdminVanityRoles"), "AdminVanityRoles");
export const AdminBannedWords = named(() => import("./admin/AdminBannedWords"), "AdminBannedWords");
export const AdminRulesPage = named(() => import("./admin/AdminRulesPage"), "AdminRulesPage");

//  Rules
export const RulesPage = named(() => import("./rules/RulesPage"), "RulesPage");

//  Users
export const UsersPage = named(() => import("./users/UsersPage"), "UsersPage");

//  Notifications
export const NotificationsPage = named(() => import("./notifications/NotificationsPage"), "NotificationsPage");

//  Channels
export const ChannelsLayout = named(() => import("./channels/ChannelsLayout"), "ChannelsLayout");

//  Not Found
export const NotFoundPage = named(() => import("./notfound/NotFoundPage"), "NotFoundPage");

//  Search
export const SearchPage = named(() => import("./search/SearchPage"), "SearchPage");
