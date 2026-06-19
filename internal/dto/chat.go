package dto

import "github.com/google/uuid"

type (
	CreateGroupRoomRequest struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		ChannelKind string   `json:"channel_kind"`
		Tags        []string `json:"tags"`
	}

	UpdateRoomRequest struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	ReorderChannelsRequest struct {
		RoomIDs []uuid.UUID `json:"room_ids"`
	}

	ChatCategoryResponse struct {
		ID        uuid.UUID `json:"id"`
		Name      string    `json:"name"`
		Position  int       `json:"position"`
		IsBuiltin bool      `json:"is_builtin"`
		Kind      string    `json:"kind,omitempty"`
	}

	InviteMembersRequest struct {
		UserIDs []uuid.UUID `json:"user_ids"`
	}

	InviteMembersResponse struct {
		InvitedCount int `json:"invited_count"`
		SkippedCount int `json:"skipped_count"`
	}

	SendMessageRequest struct {
		Body      string     `json:"body"`
		ReplyToID *uuid.UUID `json:"reply_to_id,omitempty"`
	}

	ChatMessageReplyPreview struct {
		ID          uuid.UUID `json:"id"`
		SenderID    uuid.UUID `json:"sender_id"`
		SenderName  string    `json:"sender_name"`
		BodyPreview string    `json:"body_preview"`
	}

	ChatRoomResponse struct {
		ID                uuid.UUID      `json:"id"`
		Name              string         `json:"name"`
		Description       string         `json:"description"`
		Type              string         `json:"type"`
		ChannelKind       string         `json:"channel_kind"`
		CategoryID        string         `json:"category_id,omitempty"`
		Position          int            `json:"position"`
		IsPublic          bool           `json:"is_public"`
		IsRP              bool           `json:"is_rp"`
		IsSystem          bool           `json:"is_system"`
		SystemKind        string         `json:"system_kind,omitempty"`
		Tags              []string       `json:"tags"`
		ViewerRole        string         `json:"viewer_role,omitempty"`
		ViewerMuted       bool           `json:"viewer_muted"`
		ViewerGhost       bool           `json:"viewer_ghost"`
		IsMember          bool           `json:"is_member"`
		MemberCount       int            `json:"member_count"`
		HotScore          int            `json:"hot_score"`
		Members           []UserResponse `json:"members"`
		CreatedAt         string         `json:"created_at"`
		LastMessageAt     string         `json:"last_message_at,omitempty"`
		ArchivedAt        string         `json:"archived_at,omitempty"`
		Unread            bool           `json:"unread"`
		VoiceCount        int            `json:"voice_count"`
		VoiceParticipants []uuid.UUID    `json:"voice_participants"`
	}

	VoicePresenceEvent struct {
		RoomID       uuid.UUID   `json:"room_id"`
		Participants []uuid.UUID `json:"participants"`
		Count        int         `json:"count"`
	}

	VoiceTokenResponse struct {
		Token string `json:"token"`
		URL   string `json:"url"`
	}

	VoiceMuteRequest struct {
		Muted bool `json:"muted"`
	}

	ChatRoomMemberResponse struct {
		User            UserResponse `json:"user"`
		Role            string       `json:"role"`
		JoinedAt        string       `json:"joined_at"`
		Nickname        string       `json:"nickname"`
		NicknameLocked  bool         `json:"nickname_locked"`
		MemberAvatarURL string       `json:"member_avatar_url"`
		TimeoutUntil    string       `json:"timeout_until,omitempty"`
		TimeoutByStaff  bool         `json:"timeout_set_by_staff"`
		Presence        string       `json:"presence,omitempty"`
		Ghost           bool         `json:"ghost"`
	}

	ChatMessageResponse struct {
		ID                    uuid.UUID                `json:"id"`
		RoomID                uuid.UUID                `json:"room_id"`
		Sender                UserResponse             `json:"sender"`
		SenderNickname        string                   `json:"sender_nickname,omitempty"`
		SenderMemberAvatarURL string                   `json:"sender_member_avatar_url,omitempty"`
		Body                  string                   `json:"body"`
		IsSystem              bool                     `json:"is_system"`
		CreatedAt             string                   `json:"created_at"`
		Media                 []PostMediaResponse      `json:"media,omitempty"`
		ReplyTo               *ChatMessageReplyPreview `json:"reply_to,omitempty"`
		Pinned                bool                     `json:"pinned"`
		PinnedAt              *string                  `json:"pinned_at,omitempty"`
		PinnedBy              *uuid.UUID               `json:"pinned_by,omitempty"`
		EditedAt              *string                  `json:"edited_at,omitempty"`
		Reactions             []ReactionGroup          `json:"reactions"`
	}

	EditMessageRequest struct {
		Body string `json:"body"`
	}

	ReactionGroup struct {
		Emoji         string   `json:"emoji"`
		Count         int      `json:"count"`
		ViewerReacted bool     `json:"viewer_reacted"`
		DisplayNames  []string `json:"display_names"`
	}

	UpdateMemberProfileRequest struct {
		Nickname string `json:"nickname"`
	}

	SetMemberTimeoutRequest struct {
		Amount int    `json:"amount"`
		Unit   string `json:"unit"`
	}

	AddReactionRequest struct {
		Emoji string `json:"emoji"`
	}

	ChatRoomListResponse struct {
		Rooms []ChatRoomResponse `json:"rooms"`
		Total int                `json:"total"`
	}

	ChatMessageListResponse struct {
		Messages []ChatMessageResponse `json:"messages"`
		Total    int                   `json:"total"`
		Limit    int                   `json:"limit"`
		Offset   int                   `json:"offset"`
	}

	ChatRoomBanResponse struct {
		User      UserResponse  `json:"user"`
		BannedBy  *UserResponse `json:"banned_by,omitempty"`
		Reason    string        `json:"reason"`
		CreatedAt string        `json:"created_at"`
	}

	BanMemberRequest struct {
		Reason string `json:"reason"`
	}

	BannedWordRuleResponse struct {
		ID            string  `json:"id"`
		Scope         string  `json:"scope"`
		RoomID        *string `json:"room_id,omitempty"`
		Pattern       string  `json:"pattern"`
		MatchMode     string  `json:"match_mode"`
		CaseSensitive bool    `json:"case_sensitive"`
		Action        string  `json:"action"`
		CreatedByID   *string `json:"created_by_id,omitempty"`
		CreatedByName string  `json:"created_by_name,omitempty"`
		CreatedAt     string  `json:"created_at"`
	}

	CreateBannedWordRequest struct {
		Pattern       string `json:"pattern"`
		MatchMode     string `json:"match_mode"`
		CaseSensitive bool   `json:"case_sensitive"`
		Action        string `json:"action"`
	}

	UpdateBannedWordRequest struct {
		Pattern       string `json:"pattern"`
		MatchMode     string `json:"match_mode"`
		CaseSensitive bool   `json:"case_sensitive"`
		Action        string `json:"action"`
	}

	BannedWordRuleListResponse struct {
		Rules []BannedWordRuleResponse `json:"rules"`
	}

	ChatRoomBanListResponse struct {
		Bans []ChatRoomBanResponse `json:"bans"`
	}
)
