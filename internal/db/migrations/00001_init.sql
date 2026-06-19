-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- +goose StatementEnd

-- Identity & access -------------------------------------------------------

-- +goose StatementBegin
CREATE TABLE users (
    id                       uuid PRIMARY KEY,
    username                 text NOT NULL UNIQUE,
    password_hash            text NOT NULL,
    display_name             text NOT NULL,
    created_at               timestamptz NOT NULL DEFAULT now(),
    bio                      text NOT NULL DEFAULT '',
    avatar_url               text NOT NULL DEFAULT '',
    banner_url               text NOT NULL DEFAULT '',
    favourite_character      text NOT NULL DEFAULT '',
    gender                   text NOT NULL DEFAULT '',
    pronoun_subject          text NOT NULL DEFAULT '',
    pronoun_possessive       text NOT NULL DEFAULT '',
    banned_at                timestamptz,
    banned_by                uuid REFERENCES users(id) ON DELETE SET NULL,
    ban_reason               text NOT NULL DEFAULT '',
    locked_at                timestamptz,
    locked_by                uuid REFERENCES users(id) ON DELETE SET NULL,
    lock_reason              text NOT NULL DEFAULT '',
    social_twitter           text NOT NULL DEFAULT '',
    social_discord           text NOT NULL DEFAULT '',
    social_waifulist         text NOT NULL DEFAULT '',
    social_tumblr            text NOT NULL DEFAULT '',
    social_github            text NOT NULL DEFAULT '',
    website                  text NOT NULL DEFAULT '',
    banner_position          double precision NOT NULL DEFAULT 0,
    dms_enabled              boolean NOT NULL DEFAULT true,
    email                    text NOT NULL DEFAULT '',
    email_public             boolean NOT NULL DEFAULT false,
    email_verified           boolean NOT NULL DEFAULT false,
    verify_grace_until       timestamptz NOT NULL DEFAULT now(),
    dob                      text NOT NULL DEFAULT '',
    dob_public               boolean NOT NULL DEFAULT false,
    email_notifications      boolean NOT NULL DEFAULT true,
    play_message_sound       boolean NOT NULL DEFAULT true,
    play_notification_sound  boolean NOT NULL DEFAULT true,
    home_page                text NOT NULL DEFAULT 'landing',
    default_profile_tab      text NOT NULL DEFAULT '',
    theme                    text NOT NULL DEFAULT '',
    font                     text NOT NULL DEFAULT '',
    wide_layout              boolean NOT NULL DEFAULT false,
    ip                       text,
    search_vector            tsvector GENERATED ALWAYS AS (
        to_tsvector('english',
            coalesce(display_name, '') || ' ' || coalesce(username, '') || ' ' || coalesce(bio, ''))
    ) STORED
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_users_created_at ON users (created_at DESC);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_users_search_vector ON users USING gin (search_vector);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_users_display_name_trgm ON users USING gin (display_name gin_trgm_ops);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE user_roles (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role    text NOT NULL,
    PRIMARY KEY (user_id, role)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_user_roles_role ON user_roles (role);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE sessions (
    token      text PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_sessions_user_id ON sessions (user_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE password_reset_tokens (
    token_hash text PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE email_verification_tokens (
    token_hash text PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE blocks (
    blocker_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_id, blocked_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_blocks_blocked_id ON blocks (blocked_id);
-- +goose StatementEnd

-- Vanity roles ------------------------------------------------------------

-- +goose StatementBegin
CREATE TABLE vanity_roles (
    id         text PRIMARY KEY,
    label      text NOT NULL,
    color      text NOT NULL,
    is_system  boolean NOT NULL DEFAULT false,
    sort_order integer NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE user_vanity_roles (
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vanity_role_id text NOT NULL REFERENCES vanity_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, vanity_role_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_user_vanity_roles_role ON user_vanity_roles (vanity_role_id);
-- +goose StatementEnd

-- Chat / channels ---------------------------------------------------------

-- +goose StatementBegin
CREATE TABLE chat_rooms (
    id              uuid PRIMARY KEY,
    name            text NOT NULL,
    description     text NOT NULL DEFAULT '',
    type            text NOT NULL DEFAULT 'group',
    is_public       boolean NOT NULL DEFAULT false,
    is_rp           boolean NOT NULL DEFAULT false,
    is_system       boolean NOT NULL DEFAULT false,
    system_kind     text,
    created_by      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      timestamptz NOT NULL DEFAULT now(),
    last_message_at timestamptz,
    archived_at     timestamptz
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_rooms_type ON chat_rooms (type);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_rooms_system_kind ON chat_rooms (system_kind);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_rooms_last_message_at ON chat_rooms (last_message_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_room_members (
    room_id              uuid NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id              uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role                 text NOT NULL DEFAULT 'member',
    ghost                boolean NOT NULL DEFAULT false,
    muted                boolean NOT NULL DEFAULT false,
    joined_at            timestamptz NOT NULL DEFAULT now(),
    left_at              timestamptz,
    last_read_at         timestamptz,
    nickname             text NOT NULL DEFAULT '',
    nickname_locked      boolean NOT NULL DEFAULT false,
    avatar_url           text NOT NULL DEFAULT '',
    timeout_until        timestamptz,
    timeout_set_by_staff boolean NOT NULL DEFAULT false,
    PRIMARY KEY (room_id, user_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_room_members_user_id ON chat_room_members (user_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_room_members_active ON chat_room_members (room_id, left_at);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_room_tags (
    room_id uuid NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    tag     text NOT NULL,
    PRIMARY KEY (room_id, tag)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_room_tags_tag ON chat_room_tags (tag);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_messages (
    id            uuid PRIMARY KEY,
    room_id       uuid NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body          text NOT NULL,
    reply_to_id   uuid REFERENCES chat_messages(id) ON DELETE SET NULL,
    is_system     boolean NOT NULL DEFAULT false,
    created_at    timestamptz NOT NULL DEFAULT now(),
    pinned_at     timestamptz,
    pinned_by     uuid REFERENCES users(id) ON DELETE SET NULL,
    edited_at     timestamptz,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(body, ''))) STORED
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_messages_room_created ON chat_messages (room_id, created_at DESC, id DESC);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_messages_reply_to ON chat_messages (reply_to_id);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_messages_pinned ON chat_messages (room_id, pinned_at);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_messages_search_vector ON chat_messages USING gin (search_vector);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_messages_body_trgm ON chat_messages USING gin (body gin_trgm_ops);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_message_media (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message_id    uuid NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    media_url     text NOT NULL,
    media_type    text NOT NULL,
    thumbnail_url text NOT NULL DEFAULT '',
    sort_order    integer NOT NULL DEFAULT 0
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_message_media_message ON chat_message_media (message_id, sort_order);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_message_reactions (
    message_id uuid NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji      text NOT NULL,
    PRIMARY KEY (message_id, user_id, emoji)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_message_reactions_emoji ON chat_message_reactions (message_id, emoji);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_room_bans (
    room_id    uuid NOT NULL REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    banned_by  uuid REFERENCES users(id) ON DELETE SET NULL,
    reason     text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, user_id)
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_room_bans_user_id ON chat_room_bans (user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE chat_banned_words (
    id             uuid PRIMARY KEY,
    scope          text NOT NULL,
    room_id        uuid REFERENCES chat_rooms(id) ON DELETE CASCADE,
    pattern        text NOT NULL,
    match_mode     text NOT NULL,
    case_sensitive boolean NOT NULL DEFAULT false,
    action         text NOT NULL,
    created_by     uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at     timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_chat_banned_words_scope_room ON chat_banned_words (scope, room_id);
-- +goose StatementEnd

-- Platform ----------------------------------------------------------------

-- +goose StatementBegin
CREATE TABLE reports (
    id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    reporter_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type        text NOT NULL,
    target_id          text NOT NULL,
    context_id         text,
    reason             text NOT NULL,
    status             text NOT NULL DEFAULT 'open',
    resolved_by        uuid REFERENCES users(id) ON DELETE SET NULL,
    resolution_comment text,
    created_at         timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_reports_status ON reports (status);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_reports_created_at ON reports (created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE audit_log (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    actor_id    uuid REFERENCES users(id) ON DELETE SET NULL,
    action      text NOT NULL,
    target_type text NOT NULL,
    target_id   text NOT NULL,
    details     text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_audit_log_action ON audit_log (action);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_audit_log_created_at ON audit_log (created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE invites (
    code       text PRIMARY KEY,
    created_by uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_by    uuid REFERENCES users(id) ON DELETE SET NULL,
    used_at    timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_invites_created_at ON invites (created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE notifications (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id        uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type           text NOT NULL,
    reference_id   uuid NOT NULL,
    reference_type text NOT NULL,
    actor_id       uuid REFERENCES users(id) ON DELETE CASCADE,
    message        text,
    read           boolean NOT NULL DEFAULT false,
    created_at     timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_notifications_user_read ON notifications (user_id, read);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX idx_notifications_dedupe ON notifications (user_id, type, reference_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE site_settings (
    key        text PRIMARY KEY,
    value      text NOT NULL,
    updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS site_settings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS invites;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS chat_banned_words;
DROP TABLE IF EXISTS chat_room_bans;
DROP TABLE IF EXISTS chat_message_reactions;
DROP TABLE IF EXISTS chat_message_media;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS chat_room_tags;
DROP TABLE IF EXISTS chat_room_members;
DROP TABLE IF EXISTS chat_rooms;
DROP TABLE IF EXISTS user_vanity_roles;
DROP TABLE IF EXISTS vanity_roles;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS email_verification_tokens;
DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
