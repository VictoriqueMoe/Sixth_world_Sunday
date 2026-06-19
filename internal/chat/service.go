package chat

import (
	"Sixth_world_Suday/internal/authz"
	"Sixth_world_Suday/internal/block"
	"Sixth_world_Suday/internal/contentfilter"
	"Sixth_world_Suday/internal/dto"
	"Sixth_world_Suday/internal/livekit"
	"Sixth_world_Suday/internal/media"
	"Sixth_world_Suday/internal/notification"
	"Sixth_world_Suday/internal/repository"
	"Sixth_world_Suday/internal/role"
	"Sixth_world_Suday/internal/settings"
	"Sixth_world_Suday/internal/upload"
	"Sixth_world_Suday/internal/ws"
	"context"
	"io"

	"github.com/google/uuid"
)

type (
	Service interface {
		EnsureSystemRooms(ctx context.Context) error
		SyncSystemRoomMembership(ctx context.Context, userID uuid.UUID, newRole role.Role) error

		CreateGroupRoom(ctx context.Context, creatorID uuid.UUID, req dto.CreateGroupRoomRequest) (*dto.ChatRoomResponse, error)
		UpdateRoom(ctx context.Context, roomID, userID uuid.UUID, req dto.UpdateRoomRequest) (*dto.ChatRoomResponse, error)
		ReorderChannels(ctx context.Context, categoryID uuid.UUID, roomIDs []uuid.UUID) error
		ListCategories(ctx context.Context) ([]dto.ChatCategoryResponse, error)
		ListRooms(ctx context.Context, userID uuid.UUID) (*dto.ChatRoomListResponse, error)
		ListUserGroupRooms(ctx context.Context, userID uuid.UUID, search string, isRPOnly bool, tag, role string, includeArchived bool, limit, offset int) (*dto.ChatRoomListResponse, error)
		ArchiveStale(ctx context.Context) (int, error)
		GetMessages(ctx context.Context, userID, roomID uuid.UUID, limit, offset int) (*dto.ChatMessageListResponse, error)
		GetMessagesBefore(ctx context.Context, userID, roomID uuid.UUID, before string, limit int) (*dto.ChatMessageListResponse, error)

		SendMessage(ctx context.Context, senderID, roomID uuid.UUID, req dto.SendMessageRequest, files []FileUpload) (*dto.ChatMessageResponse, error)
		GetRoomsByUser(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
		DeleteChat(ctx context.Context, roomID, userID uuid.UUID) error
		SetRoomMuted(ctx context.Context, roomID, userID uuid.UUID, muted bool) error
		IsRoomMuted(ctx context.Context, roomID, userID uuid.UUID) (bool, error)
		KickMember(ctx context.Context, hostID, roomID, targetID uuid.UUID) error
		InviteMembers(ctx context.Context, hostID, roomID uuid.UUID, userIDs []uuid.UUID) (*dto.InviteMembersResponse, error)
		SetMemberTimeout(ctx context.Context, roomID, actorID, targetID uuid.UUID, req dto.SetMemberTimeoutRequest) (*dto.ChatRoomMemberResponse, error)
		ClearMemberTimeout(ctx context.Context, roomID, actorID, targetID uuid.UUID) (*dto.ChatRoomMemberResponse, error)
		GetMembers(ctx context.Context, viewerID, roomID uuid.UUID) ([]dto.ChatRoomMemberResponse, error)
		MarkRead(ctx context.Context, roomID, userID uuid.UUID) error

		SetRoomNickname(ctx context.Context, roomID, userID uuid.UUID, nickname string) (*dto.ChatRoomMemberResponse, error)
		SetRoomAvatar(ctx context.Context, roomID, userID uuid.UUID, contentType string, fileSize int64, reader io.Reader) (*dto.ChatRoomMemberResponse, error)
		ClearRoomAvatar(ctx context.Context, roomID, userID uuid.UUID) (*dto.ChatRoomMemberResponse, error)
		SetMemberNicknameAsMod(ctx context.Context, roomID, actorID, targetID uuid.UUID, nickname string) (*dto.ChatRoomMemberResponse, error)
		UnlockMemberNickname(ctx context.Context, roomID, actorID, targetID uuid.UUID) (*dto.ChatRoomMemberResponse, error)
		PinMessage(ctx context.Context, messageID, userID uuid.UUID) error
		UnpinMessage(ctx context.Context, messageID, userID uuid.UUID) error
		ListPinnedMessages(ctx context.Context, roomID, viewerID uuid.UUID) (*dto.ChatMessageListResponse, error)
		AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
		RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
		DeleteMessage(ctx context.Context, messageID, actorID uuid.UUID) error
		EditMessage(ctx context.Context, messageID, actorID uuid.UUID, body string) (*dto.ChatMessageResponse, error)

		BanMember(ctx context.Context, actorID, roomID, targetID uuid.UUID, reason string) error
		UnbanMember(ctx context.Context, actorID, roomID, targetID uuid.UUID) error
		ListRoomBans(ctx context.Context, actorID, roomID uuid.UUID) ([]dto.ChatRoomBanResponse, error)

		ListRoomBannedWords(ctx context.Context, actorID, roomID uuid.UUID) ([]dto.BannedWordRuleResponse, error)
		CreateRoomBannedWord(ctx context.Context, actorID, roomID uuid.UUID, req dto.CreateBannedWordRequest) (*dto.BannedWordRuleResponse, error)
		UpdateRoomBannedWord(ctx context.Context, actorID, roomID, ruleID uuid.UUID, req dto.UpdateBannedWordRequest) (*dto.BannedWordRuleResponse, error)
		DeleteRoomBannedWord(ctx context.Context, actorID, roomID, ruleID uuid.UUID) error

		ListGlobalBannedWords(ctx context.Context, actorID uuid.UUID) ([]dto.BannedWordRuleResponse, error)
		CreateGlobalBannedWord(ctx context.Context, actorID uuid.UUID, req dto.CreateBannedWordRequest) (*dto.BannedWordRuleResponse, error)
		UpdateGlobalBannedWord(ctx context.Context, actorID, ruleID uuid.UUID, req dto.UpdateBannedWordRequest) (*dto.BannedWordRuleResponse, error)
		DeleteGlobalBannedWord(ctx context.Context, actorID, ruleID uuid.UUID) error

		VoiceEnabled() bool
		MintVoiceToken(ctx context.Context, roomID, userID uuid.UUID) (token, url string, err error)
		ForceMuteVoice(ctx context.Context, roomID, actorID, targetID uuid.UUID, muted bool) error
		HandleVoiceWebhook(ctx context.Context, authHeader string, body []byte) error
		VoiceParticipants(roomID uuid.UUID) []uuid.UUID
		VoiceCount(roomID uuid.UUID) int
		ReconcilePresence(ctx context.Context) (int, error)
	}

	service struct {
		*core
		*systemService
		*reactionsService
		*membersService
		*roomsService
		*messagesService
		*moderationService
		*voiceService
	}
)

func NewService(
	chatRepo repository.ChatRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	vanityRoleRepo repository.VanityRoleRepository,
	banRepo repository.ChatRoomBanRepository,
	bannedWordRepo repository.ChatBannedWordRepository,
	auditRepo repository.AuditLogRepository,
	authzSvc authz.Service,
	notifSvc notification.Service,
	blockSvc block.Service,
	uploadSvc upload.Service,
	settingsSvc settings.Service,
	mediaProc *media.Processor,
	hub *ws.Hub,
	livekitSvc livekit.Service,
	contentFilter *contentfilter.Manager,
) Service {
	c := &core{
		chatRepo:        chatRepo,
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		vanityRoleRepo:  vanityRoleRepo,
		banRepo:         banRepo,
		bannedWordRepo:  bannedWordRepo,
		auditRepo:       auditRepo,
		authzSvc:        authzSvc,
		notifSvc:        notifSvc,
		blockSvc:        blockSvc,
		settingsSvc:     settingsSvc,
		uploadSvc:       uploadSvc,
		uploader:        media.NewUploader(uploadSvc, settingsSvc, mediaProc),
		hub:             hub,
		livekitSvc:      livekitSvc,
		contentFilter:   contentFilter,
		bannedWordsRule: contentfilter.NewChatBannedWordsRule(bannedWordRepo),
		voiceMuted:      make(map[string]map[uuid.UUID]struct{}),
	}

	svs := &service{
		core:              c,
		systemService:     &systemService{core: c},
		reactionsService:  &reactionsService{core: c},
		roomsService:      &roomsService{core: c},
		moderationService: &moderationService{core: c},
		voiceService:      newVoiceService(c),
	}
	svs.membersService = &membersService{core: c, parent: svs}
	svs.messagesService = &messagesService{core: c, parent: svs}

	return svs
}
