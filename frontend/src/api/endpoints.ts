import {
    apiDelete,
    apiDeleteWithBody,
    apiFetch,
    apiPatch,
    apiPost,
    apiPostFormData,
    apiPut,
    buildQueryString,
} from "./client";
import type {
    AdminStats,
    AdminUserDetail,
    AdminUserListResponse,
    AuditLogListResponse,
    BannedWordRule,
    ChangePasswordPayload,
    ChatMessage,
    ChatMessageListResponse,
    ChatRoom,
    ChatRoomBan,
    ChatRoomMember,
    CreateBannedWordRequest,
    DeleteAccountPayload,
    NotificationListResponse,
    QuickSearchResponse,
    SearchResponse,
    SiteSettings,
    UpdateProfilePayload,
    User,
    UserProfile,
} from "../types/api";

export interface VanityRoleDefinition {
    id: string;
    label: string;
    color: string;
    is_system: boolean;
    sort_order: number;
}

export interface SiteInfo {
    site_name: string;
    site_description: string;
    registration_type: string;
    announcement_banner: string;
    default_theme: string;
    maintenance_mode: boolean;
    maintenance_title: string;
    maintenance_message: string;
    turnstile_enabled: boolean;
    turnstile_site_key: string;
    voice_enabled: boolean;
    email_enabled: boolean;
    max_image_size: number;
    max_video_size: number;
    vanity_roles: VanityRoleDefinition[];
    vanity_role_assignments: Record<string, string[]>;
    rules_page: string;
    version: string;
}

export async function getSiteInfo(): Promise<SiteInfo> {
    return apiFetch<SiteInfo>("/site-info");
}

export async function getStaff(): Promise<User[]> {
    return apiFetch<User[]>("/staff");
}

export async function register(
    username: string,
    email: string,
    password: string,
    displayName: string,
    inviteCode?: string,
    turnstileToken?: string,
): Promise<User> {
    return apiPost<
        User,
        {
            username: string;
            email: string;
            password: string;
            display_name: string;
            invite_code?: string;
            turnstile_token?: string;
        }
    >("/auth/register", {
        username,
        email,
        password,
        display_name: displayName,
        invite_code: inviteCode,
        turnstile_token: turnstileToken,
    });
}

export async function setEmail(email: string): Promise<void> {
    await apiPost<unknown, { email: string }>("/auth/set-email", { email });
}

export async function verifyEmail(token: string): Promise<void> {
    await apiPost<unknown, { token: string }>("/auth/verify-email", { token });
}

export async function resendVerification(): Promise<void> {
    await apiPost<unknown, undefined>("/auth/resend-verification", undefined);
}

export async function login(username: string, password: string, turnstileToken?: string): Promise<User> {
    return apiPost<User, { username: string; password: string; turnstile_token?: string }>("/auth/login", {
        username,
        password,
        turnstile_token: turnstileToken,
    });
}

export async function forgotPassword(username: string, turnstileToken?: string): Promise<void> {
    await apiPost<unknown, { username: string; turnstile_token?: string }>("/auth/forgot-password", {
        username,
        turnstile_token: turnstileToken,
    });
}

export async function resetPassword(token: string, newPassword: string): Promise<void> {
    await apiPost<unknown, { token: string; new_password: string }>("/auth/reset-password", {
        token,
        new_password: newPassword,
    });
}

export async function logout(): Promise<void> {
    await apiPost<unknown, undefined>("/auth/logout", undefined);
}

export async function getMe(): Promise<UserProfile> {
    const session = await apiFetch<{ username: string }>("/auth/session");
    return getUserProfile(session.username);
}

export async function getUserProfile(username: string): Promise<UserProfile> {
    return apiFetch<UserProfile>(`/users/${username}`);
}

export async function updateProfile(payload: UpdateProfilePayload): Promise<{ status: string }> {
    return apiPut<{ status: string }, UpdateProfilePayload>("/auth/profile", payload);
}

export async function updateAppearance(wideLayout: boolean): Promise<void> {
    await apiPut<unknown, { wide_layout: boolean }>("/preferences/appearance", {
        wide_layout: wideLayout,
    });
}

export async function uploadAvatar(file: File): Promise<{ avatar_url: string }> {
    const formData = new FormData();
    formData.append("avatar", file);
    return apiPostFormData<{ avatar_url: string }>("/auth/avatar", formData);
}

export async function getNotifications(params: { limit?: number; offset?: number }): Promise<NotificationListResponse> {
    const qs = buildQueryString({ limit: params.limit ?? 20, offset: params.offset });
    return apiFetch<NotificationListResponse>(`/notifications${qs}`);
}

export async function markNotificationRead(id: number): Promise<void> {
    await apiPost<unknown, undefined>(`/notifications/${id}/read`, undefined);
}

export async function markAllNotificationsRead(): Promise<void> {
    await apiPost<unknown, undefined>("/notifications/read", undefined);
}

export async function getUnreadCount(): Promise<{ count: number }> {
    return apiFetch<{ count: number }>("/notifications/unread-count");
}

export async function uploadBanner(file: File): Promise<{ banner_url: string }> {
    const formData = new FormData();
    formData.append("banner", file);
    return apiPostFormData<{ banner_url: string }>("/auth/banner", formData);
}

export async function changePassword(payload: ChangePasswordPayload): Promise<{ status: string }> {
    return apiPut<{ status: string }, ChangePasswordPayload>("/auth/password", payload);
}

export async function deleteAccount(payload: DeleteAccountPayload): Promise<{ status: string }> {
    return apiDeleteWithBody<{ status: string }, DeleteAccountPayload>("/auth/account", payload);
}

export async function getOnlineStatus(ids: string[]): Promise<Record<string, boolean>> {
    return apiFetch<Record<string, boolean>>(`/users/online?ids=${ids.join(",")}`);
}

export async function getAdminStats(): Promise<AdminStats> {
    return apiFetch<AdminStats>("/admin/stats");
}

export async function getAdminUsers(params: {
    search?: string;
    limit?: number;
    offset?: number;
}): Promise<AdminUserListResponse> {
    const qs = buildQueryString({ search: params.search, limit: params.limit ?? 20, offset: params.offset });
    return apiFetch<AdminUserListResponse>(`/admin/users${qs}`);
}

export async function getAdminUser(id: string): Promise<AdminUserDetail> {
    return apiFetch<AdminUserDetail>(`/admin/users/${id}`);
}

export async function setUserRole(id: string, role: string): Promise<void> {
    await apiPost<unknown, { role: string }>(`/admin/users/${id}/role`, { role });
}

export async function removeUserRole(id: string, role: string): Promise<void> {
    await apiDeleteWithBody<unknown, { role: string }>(`/admin/users/${id}/role`, { role });
}

export async function banUser(id: string, reason: string): Promise<void> {
    await apiPost<unknown, { reason: string }>(`/admin/users/${id}/ban`, { reason });
}

export async function unbanUser(id: string): Promise<void> {
    await apiPost<unknown, undefined>(`/admin/users/${id}/unban`, undefined);
}

export async function lockUser(id: string, reason: string): Promise<void> {
    await apiPost<unknown, { reason: string }>(`/admin/users/${id}/lock`, { reason });
}

export async function unlockUser(id: string): Promise<void> {
    await apiPost<unknown, undefined>(`/admin/users/${id}/unlock`, undefined);
}

export async function adminDeleteUser(id: string): Promise<void> {
    await apiDelete<unknown>(`/admin/users/${id}`);
}

export async function resetUserPassword(id: string): Promise<{ password: string }> {
    return apiPost<{ password: string }, undefined>(`/admin/users/${id}/reset-password`, undefined);
}

export async function getAdminSettings(): Promise<SiteSettings> {
    return apiFetch<{ settings: SiteSettings }>("/admin/settings").then(r => r.settings);
}

export async function updateAdminSettings(settings: SiteSettings): Promise<void> {
    await apiPut<unknown, { settings: SiteSettings }>("/admin/settings", { settings });
}

export async function uploadOGDefaultImage(file: File): Promise<{ url: string }> {
    const formData = new FormData();
    formData.append("image", file);
    return apiPostFormData<{ url: string }>("/admin/settings/og-image", formData);
}

export async function sendTestEmail(): Promise<void> {
    await apiPost<unknown, undefined>("/admin/settings/test-email", undefined);
}

export async function getAuditLog(params: {
    action?: string;
    limit?: number;
    offset?: number;
}): Promise<AuditLogListResponse> {
    const qs = buildQueryString({ action: params.action, limit: params.limit ?? 50, offset: params.offset });
    return apiFetch<AuditLogListResponse>(`/admin/audit-log${qs}`);
}

interface InviteItem {
    code: string;
    created_by: string;
    used_by?: string;
    used_at?: string;
    created_at: string;
}

interface InviteListResponse {
    invites: InviteItem[];
    total: number;
    limit: number;
    offset: number;
}

export async function createInvite(): Promise<InviteItem> {
    return apiPost<InviteItem, undefined>("/admin/invites", undefined);
}

export async function getInvites(params: { limit?: number; offset?: number }): Promise<InviteListResponse> {
    const qs = buildQueryString({ limit: params.limit ?? 50, offset: params.offset });
    return apiFetch<InviteListResponse>(`/admin/invites${qs}`);
}

export async function deleteInvite(code: string): Promise<void> {
    await apiDelete<unknown>(`/admin/invites/${code}`);
}

export async function createChannel(payload: {
    name: string;
    description: string;
    channel_kind: "text" | "voice";
}): Promise<ChatRoom> {
    return apiPost<ChatRoom, typeof payload>("/chat/rooms", payload);
}

export async function setChatRoomMuted(roomId: string, muted: boolean): Promise<{ muted: boolean }> {
    return apiPut<{ muted: boolean }, { muted: boolean }>(`/chat/rooms/${roomId}/mute`, { muted });
}

export async function getChatRoomMembers(roomId: string): Promise<{ members: ChatRoomMember[] }> {
    return apiFetch<{ members: ChatRoomMember[] }>(`/chat/rooms/${roomId}/members`);
}

export async function kickChatRoomMember(roomId: string, userId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/rooms/${roomId}/members/${userId}`);
}

export async function banChatRoomMember(roomId: string, userId: string, reason: string): Promise<void> {
    await apiPost<unknown, { reason: string }>(`/chat/rooms/${roomId}/bans/${userId}`, { reason });
}

export async function unbanChatRoomMember(roomId: string, userId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/rooms/${roomId}/bans/${userId}`);
}

export async function listChatRoomBans(roomId: string): Promise<{ bans: ChatRoomBan[] }> {
    return apiFetch<{ bans: ChatRoomBan[] }>(`/chat/rooms/${roomId}/bans`);
}

export async function listChatRoomBannedWords(roomId: string): Promise<{ rules: BannedWordRule[] }> {
    return apiFetch<{ rules: BannedWordRule[] }>(`/chat/rooms/${roomId}/banned-words`);
}

export async function createChatRoomBannedWord(roomId: string, req: CreateBannedWordRequest): Promise<BannedWordRule> {
    return apiPost<BannedWordRule, CreateBannedWordRequest>(`/chat/rooms/${roomId}/banned-words`, req);
}

export async function updateChatRoomBannedWord(
    roomId: string,
    ruleId: string,
    req: CreateBannedWordRequest,
): Promise<BannedWordRule> {
    return apiPut<BannedWordRule, CreateBannedWordRequest>(`/chat/rooms/${roomId}/banned-words/${ruleId}`, req);
}

export async function deleteChatRoomBannedWord(roomId: string, ruleId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/rooms/${roomId}/banned-words/${ruleId}`);
}

export async function listGlobalBannedWords(): Promise<{ rules: BannedWordRule[] }> {
    return apiFetch<{ rules: BannedWordRule[] }>("/admin/banned-words");
}

export async function createGlobalBannedWord(req: CreateBannedWordRequest): Promise<BannedWordRule> {
    return apiPost<BannedWordRule, CreateBannedWordRequest>("/admin/banned-words", req);
}

export async function updateGlobalBannedWord(ruleId: string, req: CreateBannedWordRequest): Promise<BannedWordRule> {
    return apiPut<BannedWordRule, CreateBannedWordRequest>(`/admin/banned-words/${ruleId}`, req);
}

export async function deleteGlobalBannedWord(ruleId: string): Promise<void> {
    await apiDelete<unknown>(`/admin/banned-words/${ruleId}`);
}

export async function inviteChatRoomMembers(
    roomId: string,
    userIds: string[],
): Promise<{ invited_count: number; skipped_count: number }> {
    return apiPost<{ invited_count: number; skipped_count: number }, { user_ids: string[] }>(
        `/chat/rooms/${roomId}/members`,
        { user_ids: userIds },
    );
}

interface VoiceTokenResponse {
    token: string;
    url: string;
}

export async function getVoiceToken(roomId: string): Promise<VoiceTokenResponse> {
    return apiPost<VoiceTokenResponse, Record<string, never>>(`/chat/rooms/${roomId}/voice/token`, {});
}

export async function forceMuteVoiceParticipant(roomId: string, userId: string, muted: boolean): Promise<void> {
    await apiPost<unknown, { muted: boolean }>(`/chat/rooms/${roomId}/voice/participants/${userId}/mute`, { muted });
}

export async function getUserRooms(): Promise<{ rooms: ChatRoom[] }> {
    return apiFetch<{ rooms: ChatRoom[] }>("/chat/rooms");
}

export async function getRoomMessages(
    roomId: string,
    limit?: number,
    offset?: number,
): Promise<{ messages: ChatMessage[]; total: number }> {
    const qs = buildQueryString({ limit: limit ?? 50, offset });
    return apiFetch<{ messages: ChatMessage[]; total: number }>(`/chat/rooms/${roomId}/messages${qs}`);
}

export async function getRoomMessagesBefore(
    roomId: string,
    before: string,
    limit?: number,
): Promise<{ messages: ChatMessage[] }> {
    const qs = buildQueryString({ before, limit: limit ?? 50 });
    return apiFetch<{ messages: ChatMessage[] }>(`/chat/rooms/${roomId}/messages${qs}`);
}

export async function sendChatMessage(
    roomId: string,
    payload: { body: string; reply_to_id?: string; files?: File[] },
): Promise<ChatMessage> {
    const formData = new FormData();
    formData.append("body", payload.body);
    if (payload.reply_to_id) {
        formData.append("reply_to_id", payload.reply_to_id);
    }
    if (payload.files) {
        for (let i = 0; i < payload.files.length; i++) {
            formData.append("media", payload.files[i]);
        }
    }
    return apiPostFormData<ChatMessage>(`/chat/rooms/${roomId}/messages`, formData);
}

export async function deleteChatRoom(roomId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/rooms/${roomId}`);
}

export async function markChatRoomRead(roomId: string): Promise<void> {
    await apiPost<unknown, Record<string, never>>(`/chat/rooms/${roomId}/read`, {});
}

export async function updateChatRoomNickname(roomId: string, nickname: string): Promise<ChatRoomMember> {
    return apiPut<ChatRoomMember, { nickname: string }>(`/chat/rooms/${roomId}/me`, { nickname });
}

export async function setChatRoomMemberNickname(
    roomId: string,
    userId: string,
    nickname: string,
): Promise<ChatRoomMember> {
    return apiPut<ChatRoomMember, { nickname: string }>(`/chat/rooms/${roomId}/members/${userId}/nickname`, {
        nickname,
    });
}

export async function unlockChatRoomMemberNickname(roomId: string, userId: string): Promise<ChatRoomMember> {
    return apiDelete<ChatRoomMember>(`/chat/rooms/${roomId}/members/${userId}/nickname`);
}

export async function setChatRoomMemberTimeout(
    roomId: string,
    userId: string,
    amount: number,
    unit: string,
): Promise<ChatRoomMember> {
    return apiPut<ChatRoomMember, { amount: number; unit: string }>(`/chat/rooms/${roomId}/members/${userId}/timeout`, {
        amount,
        unit,
    });
}

export async function clearChatRoomMemberTimeout(roomId: string, userId: string): Promise<ChatRoomMember> {
    return apiDelete<ChatRoomMember>(`/chat/rooms/${roomId}/members/${userId}/timeout`);
}

export async function uploadChatRoomAvatar(roomId: string, file: File): Promise<ChatRoomMember> {
    const formData = new FormData();
    formData.append("avatar", file);
    return apiPostFormData<ChatRoomMember>(`/chat/rooms/${roomId}/me/avatar`, formData);
}

export async function clearChatRoomAvatar(roomId: string): Promise<ChatRoomMember> {
    return apiDelete<ChatRoomMember>(`/chat/rooms/${roomId}/me/avatar`);
}

export async function deleteChatMessage(messageId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/messages/${messageId}`);
}

export async function editChatMessage(messageId: string, body: string): Promise<ChatMessage> {
    return apiPatch<ChatMessage, { body: string }>(`/chat/messages/${messageId}`, { body });
}

export async function pinChatMessage(messageId: string): Promise<void> {
    await apiPost<unknown, Record<string, never>>(`/chat/messages/${messageId}/pin`, {});
}

export async function unpinChatMessage(messageId: string): Promise<void> {
    await apiDelete<unknown>(`/chat/messages/${messageId}/pin`);
}

export async function getChatRoomPinnedMessages(roomId: string): Promise<ChatMessageListResponse> {
    return apiFetch<ChatMessageListResponse>(`/chat/rooms/${roomId}/pins`);
}

export async function addChatMessageReaction(messageId: string, emoji: string): Promise<void> {
    await apiPost<unknown, { emoji: string }>(`/chat/messages/${messageId}/reactions`, { emoji });
}

export async function removeChatMessageReaction(messageId: string, emoji: string): Promise<void> {
    await apiDelete<unknown>(`/chat/messages/${messageId}/reactions/${encodeURIComponent(emoji)}`);
}

export async function createReport(
    targetType: string,
    targetId: string,
    reason: string,
    contextId?: string,
): Promise<void> {
    await apiPost<unknown, { target_type: string; target_id: string; context_id?: string; reason: string }>("/report", {
        target_type: targetType,
        target_id: targetId,
        context_id: contextId,
        reason,
    });
}

export interface ReportItem {
    id: number;
    reporter_name: string;
    reporter_avatar: string;
    target_type: string;
    target_id: string;
    context_id?: string;
    reason: string;
    status: string;
    resolved_by?: string;
    created_at: string;
}

interface ReportListResponse {
    reports: ReportItem[];
    total: number;
    limit: number;
    offset: number;
}

export async function getReports(
    status: string = "open",
    limit: number = 50,
    offset: number = 0,
): Promise<ReportListResponse> {
    const qs = buildQueryString({ status, limit, offset });
    return apiFetch<ReportListResponse>(`/admin/reports${qs}`);
}

export async function resolveReport(id: number, comment: string): Promise<void> {
    await apiPost<unknown, { comment: string }>(`/admin/reports/${id}/resolve`, { comment });
}

export async function getRules(page: string): Promise<{ page: string; rules: string }> {
    return apiFetch<{ page: string; rules: string }>(`/rules/${page}`);
}

export async function searchUsers(query: string): Promise<User[]> {
    return apiFetch<User[]>(`/users/search?q=${encodeURIComponent(query)}`);
}

export interface PublicUser extends User {
    online: boolean;
}

export async function listUsersPublic(): Promise<PublicUser[]> {
    return apiFetch<PublicUser[]>("/users");
}

export async function blockUser(id: string): Promise<void> {
    await apiPost<unknown, undefined>(`/users/${id}/block`, undefined);
}

export async function unblockUser(id: string): Promise<void> {
    await apiDelete(`/users/${id}/block`);
}

interface BlockStatus {
    blocking: boolean;
    blocked_by: boolean;
}

export async function getBlockStatus(id: string): Promise<BlockStatus> {
    return apiFetch<BlockStatus>(`/users/${id}/block-status`);
}

interface BlockedUserItem {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string;
    blocked_at: string;
}

export async function getBlockedUsers(): Promise<{ users: BlockedUserItem[] }> {
    return apiFetch<{ users: BlockedUserItem[] }>("/blocked-users");
}

export async function getVanityRoles(): Promise<VanityRoleDefinition[]> {
    return apiFetch<VanityRoleDefinition[]>("/admin/vanity-roles");
}

export async function createVanityRole(data: {
    label: string;
    color: string;
    sort_order: number;
}): Promise<VanityRoleDefinition> {
    return apiPost<VanityRoleDefinition, typeof data>("/admin/vanity-roles", data);
}

export async function updateVanityRole(
    id: string,
    data: { label: string; color: string; sort_order: number },
): Promise<void> {
    await apiPut<unknown, typeof data>(`/admin/vanity-roles/${id}`, data);
}

export async function deleteVanityRole(id: string): Promise<void> {
    await apiDelete(`/admin/vanity-roles/${id}`);
}

interface VanityRoleUsersResponse {
    users: { id: string; username: string; display_name: string; avatar_url: string }[];
    total: number;
    limit: number;
    offset: number;
}

export async function getVanityRoleUsers(
    id: string,
    params: { search?: string; limit?: number; offset?: number },
): Promise<VanityRoleUsersResponse> {
    const parts: string[] = [];
    if (params.search) {
        parts.push(`search=${encodeURIComponent(params.search)}`);
    }
    parts.push(`limit=${params.limit ?? 20}`);
    if (params.offset) {
        parts.push(`offset=${params.offset}`);
    }
    return apiFetch<VanityRoleUsersResponse>(`/admin/vanity-roles/${id}/users?${parts.join("&")}`);
}

export async function assignVanityRole(roleId: string, userId: string): Promise<void> {
    await apiPost<unknown, { user_id: string }>(`/admin/vanity-roles/${roleId}/users`, { user_id: userId });
}

export async function unassignVanityRole(roleId: string, userId: string): Promise<void> {
    await apiDelete(`/admin/vanity-roles/${roleId}/users/${userId}`);
}

export async function quickSearch(q: string, perType: number = 3): Promise<QuickSearchResponse> {
    const qs = buildQueryString({ q, perType });
    return apiFetch<QuickSearchResponse>(`/search/quick${qs}`);
}

export async function searchSite(
    q: string,
    types?: string,
    limit: number = 20,
    offset: number = 0,
    room?: string,
): Promise<SearchResponse> {
    const qs = buildQueryString({ q, types: types ?? "", limit, offset, room });
    return apiFetch<SearchResponse>(`/search${qs}`);
}
