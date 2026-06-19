import type { SiteRole } from "../utils/permissions";

export interface User {
    id: string;
    username: string;
    display_name: string;
    avatar_url?: string;
    role?: SiteRole;
    banned?: boolean;
    ban_reason?: string;
    locked?: boolean;
    lock_reason?: string;
}

export interface UserProfile {
    id: string;
    username: string;
    display_name: string;
    bio: string;
    avatar_url: string;
    banner_url: string;
    banner_position: number;
    favourite_character: string;
    gender: string;
    pronoun_subject: string;
    pronoun_possessive: string;
    role?: SiteRole;
    online: boolean;
    social_twitter: string;
    social_discord: string;
    social_waifulist: string;
    social_tumblr: string;
    social_github: string;
    website: string;
    dms_enabled: boolean;
    secrets: string[];
    dob?: string;
    dob_public?: boolean;
    email?: string;
    email_public?: boolean;
    created_at: string;
    banned?: boolean;
    ban_reason?: string;
    locked?: boolean;
    lock_reason?: string;
    private?: UserPrivateFields;
}

export interface UserPrivateFields {
    email_verified?: boolean;
    verify_grace_until?: string;
    email_notifications?: boolean;
    play_message_sound?: boolean;
    play_notification_sound?: boolean;
    wide_layout?: boolean;
}

export interface UpdateProfilePayload {
    display_name: string;
    bio: string;
    avatar_url: string;
    banner_url: string;
    banner_position: number;
    favourite_character: string;
    gender: string;
    pronoun_subject: string;
    pronoun_possessive: string;
    social_twitter: string;
    social_discord: string;
    social_waifulist: string;
    social_tumblr: string;
    social_github: string;
    website: string;
    dms_enabled: boolean;
    dob: string;
    dob_public: boolean;
    email: string;
    email_public: boolean;
    email_notifications: boolean;
    play_message_sound: boolean;
    play_notification_sound: boolean;
}

export interface ChangePasswordPayload {
    old_password: string;
    new_password: string;
}

export interface DeleteAccountPayload {
    password: string;
}

export interface PostMedia {
    id: number;
    media_url: string;
    media_type: "image" | "video";
    thumbnail_url?: string;
    sort_order: number;
}

export type NotificationType =
    | "chat_message"
    | "report"
    | "report_resolved"
    | "mention"
    | "comment_liked"
    | "content_edited"
    | "chat_mention"
    | "chat_reaction"
    | "chat_room_message"
    | "chat_room_invite"
    | "chat_reply"
    | "chat_room_banned"
    | "chat_room_kicked"
    | "chat_room_unbanned"
    | "content_shared";

export interface Notification {
    id: number;
    type: NotificationType;
    reference_id: string;
    reference_type: string;
    actor: User;
    message?: string;
    read: boolean;
    created_at: string;
    count: number;
}

export interface NotificationListResponse {
    notifications: Notification[];
    total: number;
    limit: number;
    offset: number;
}

export interface WSMessage {
    type: string;
    data: unknown;
}

export interface AdminUserItem {
    id: string;
    username: string;
    display_name: string;
    avatar_url: string;
    role?: SiteRole;
    banned: boolean;
    locked: boolean;
    created_at: string;
}

export interface AdminUserListResponse {
    users: AdminUserItem[];
    total: number;
    limit: number;
    offset: number;
}

export interface AdminUserDetail extends AdminUserItem {
    email?: string;
    ip?: string;
    ban_reason?: string;
    banned_at?: string;
    banned_by?: User;
    lock_reason?: string;
    locked_at?: string;
}

export interface AdminStats {
    total_users: number;
    total_messages: number;
    total_rooms: number;
    new_users_24h: number;
    new_users_7d: number;
    new_users_30d: number;
    new_messages_24h: number;
    new_messages_7d: number;
    new_messages_30d: number;
    most_active_users: {
        id: string;
        username: string;
        display_name: string;
        avatar_url: string;
        action_count: number;
    }[];
}

export interface AuditLogEntry {
    id: number;
    actor_id: string;
    actor_name: string;
    action: string;
    target_type: string;
    target_id: string;
    details: string;
    created_at: string;
}

export interface AuditLogListResponse {
    entries: AuditLogEntry[];
    total: number;
    limit: number;
    offset: number;
}

export interface SiteSettings {
    [key: string]: string;
}

export interface ChatCategory {
    id: string;
    name: string;
    position: number;
    is_builtin: boolean;
    kind?: string;
}

export interface ChatRoom {
    id: string;
    name: string;
    description: string;
    type: "dm" | "group";
    channel_kind: "text" | "voice";
    category_id?: string;
    position: number;
    is_public: boolean;
    is_rp: boolean;
    is_system: boolean;
    system_kind?: string;
    tags: string[];
    viewer_role?: string;
    viewer_muted: boolean;
    viewer_ghost: boolean;
    is_member: boolean;
    member_count: number;
    hot_score: number;
    members: User[];
    created_at: string;
    last_message_at?: string;
    archived_at?: string;
    unread?: boolean;
    voice_count?: number;
    voice_participants?: string[];
}

export interface ChatRoomMember {
    user: User;
    role: string;
    joined_at: string;
    nickname: string;
    member_avatar_url: string;
    nickname_locked: boolean;
    timeout_until?: string;
    timeout_set_by_staff?: boolean;
    presence?: "active" | "idle" | "";
    ghost?: boolean;
}

export interface ChatRoomBan {
    user: User;
    banned_by?: User;
    reason: string;
    created_at: string;
}

export type BannedWordMatchMode = "substring" | "whole_word" | "regex";
export type BannedWordAction = "delete" | "kick";

export interface BannedWordRule {
    id: string;
    scope: "global" | "room";
    room_id?: string;
    pattern: string;
    match_mode: BannedWordMatchMode;
    case_sensitive: boolean;
    action: BannedWordAction;
    created_by_id?: string;
    created_by_name?: string;
    created_at: string;
}

export interface CreateBannedWordRequest {
    pattern: string;
    match_mode: BannedWordMatchMode;
    case_sensitive: boolean;
    action: BannedWordAction;
}

export interface ChatMessageReplyPreview {
    id: string;
    sender_id: string;
    sender_name: string;
    body_preview: string;
}

export interface ReactionGroup {
    emoji: string;
    count: number;
    viewer_reacted: boolean;
    display_names: string[];
}

export interface ChatMessage {
    id: string;
    room_id: string;
    sender: User;
    body: string;
    is_system: boolean;
    created_at: string;
    media?: PostMedia[];
    reply_to?: ChatMessageReplyPreview;
    pinned: boolean;
    pinned_at?: string;
    pinned_by?: string;
    edited_at?: string;
    reactions: ReactionGroup[];
    sender_nickname?: string;
    sender_member_avatar_url?: string;
}

export interface ChatMessageListResponse {
    messages: ChatMessage[];
    total: number;
}

export type SearchEntityType = "chat_message" | "user";

export interface SearchResultAuthor {
    id: string | null;
    username: string;
    display_name: string;
    avatar_url: string;
}

export interface SearchResult {
    type: SearchEntityType;
    id: string;
    parent_id: string | null;
    parent_title: string | null;
    title: string;
    snippet: string;
    url: string;
    author: SearchResultAuthor;
    created_at: string;
}

export interface SearchResponse {
    results: SearchResult[];
    total: number;
}

export interface QuickSearchResponse {
    results: SearchResult[];
}
