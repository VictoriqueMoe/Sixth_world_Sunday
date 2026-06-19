import { Suspense, useEffect } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { useSiteInfo } from "./hooks/useSiteInfo";
import { useAuth } from "./hooks/useAuth";
import { canAccessAdmin } from "./utils/permissions";
import { ensureNotificationPermission } from "./utils/notifications";
import { Header } from "./components/layout/Header/Header";
import { VoiceProvider } from "./context/VoiceContext";
import { CanonicalTag } from "./components/CanonicalTag/CanonicalTag";
import { ProtectedRoute } from "./components/ProtectedRoute/ProtectedRoute";
import { LockBanner } from "./components/LockBanner/LockBanner";
import { VerifyEmailBanner } from "./components/VerifyEmailBanner/VerifyEmailBanner";
import { MaintenancePage } from "./pages/maintenance/MaintenancePage";
import { linkify } from "./utils/linkify";
import {
    AdminAuditLog,
    AdminBannedWords,
    AdminContentRules,
    AdminDashboard,
    AdminInvites,
    AdminLayout,
    AdminReports,
    AdminRulesPage,
    AdminSettings,
    AdminUserDetail,
    AdminUsers,
    AdminVanityRoles,
    ForgotPasswordPage,
    LoginPage,
    NotFoundPage,
    NotificationsPage,
    ChannelsLayout,
    ProfilePage,
    ResetPasswordPage,
    SetEmailPage,
    RulesPage,
    SearchPage,
    SettingsPage,
    UsersPage,
    VerifyEmailPage,
} from "./pages/lazyPages";

function HomePage() {
    const { user } = useAuth();
    return <Navigate to={user ? "/channels" : "/login"} replace />;
}

function AnnouncementBanner() {
    const siteInfo = useSiteInfo();
    const banner = siteInfo.announcement_banner ?? "";

    if (!banner) {
        return null;
    }

    return <div className="announcement-banner">{linkify(banner, "ab")}</div>;
}

function RouteFallback() {
    return <div className="loading">Loading...</div>;
}

function AppLayout() {
    const siteInfo = useSiteInfo();
    const { user, loading: authLoading } = useAuth();

    useEffect(() => {
        if (user) {
            ensureNotificationPermission().catch(() => {});
        }
    }, [user]);

    if (authLoading) {
        return null;
    }

    if (siteInfo.maintenance_mode && !canAccessAdmin(user?.role)) {
        return (
            <MaintenancePage title={siteInfo.maintenance_title ?? ""} message={siteInfo.maintenance_message ?? ""} />
        );
    }

    return (
        <VoiceProvider>
            <div className="app-layout">
                <CanonicalTag />
                <div className="app-main">
                    <Header />
                    <LockBanner />
                    <VerifyEmailBanner />
                    <AnnouncementBanner />
                    <main className="main-content">
                        <Suspense fallback={<RouteFallback />}>
                            <Routes>
                                <Route path="/" element={<HomePage />} />
                                <Route path="/login" element={<LoginPage />} />
                                <Route path="/forgot-password" element={<ForgotPasswordPage />} />
                                <Route path="/reset-password" element={<ResetPasswordPage />} />
                                <Route path="/set-email" element={<SetEmailPage />} />
                                <Route path="/verify-email" element={<VerifyEmailPage />} />

                                <Route element={<ProtectedRoute />}>
                                    <Route path="/channels" element={<ChannelsLayout />}>
                                        <Route path=":roomId" element={null} />
                                    </Route>
                                    <Route path="/rules" element={<RulesPage />} />
                                    <Route path="/search" element={<SearchPage />} />
                                    <Route path="/users" element={<UsersPage />} />
                                    <Route path="/user/:username" element={<ProfilePage />} />
                                    <Route path="/notifications" element={<NotificationsPage />} />
                                    <Route path="/settings" element={<SettingsPage />} />
                                </Route>

                                <Route element={<ProtectedRoute permission="view_admin_panel" />}>
                                    <Route path="/admin" element={<AdminLayout />}>
                                        <Route index element={<AdminDashboard />} />
                                        <Route path="users" element={<AdminUsers />} />
                                        <Route path="users/:id" element={<AdminUserDetail />} />
                                        <Route path="invites" element={<AdminInvites />} />
                                        <Route path="settings" element={<AdminSettings />} />
                                        <Route path="reports" element={<AdminReports />} />
                                        <Route path="content-rules" element={<AdminContentRules />} />
                                        <Route path="rules" element={<AdminRulesPage />} />
                                        <Route path="banned-words" element={<AdminBannedWords />} />
                                        <Route path="audit-log" element={<AdminAuditLog />} />
                                        <Route path="vanity-roles" element={<AdminVanityRoles />} />
                                    </Route>
                                </Route>

                                <Route path="*" element={<NotFoundPage />} />
                            </Routes>
                        </Suspense>
                    </main>
                </div>
            </div>
        </VoiceProvider>
    );
}

export default function App() {
    return (
        <BrowserRouter>
            <AppLayout />
        </BrowserRouter>
    );
}
