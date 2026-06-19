package chat

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"Sixth_world_Suday/internal/config"
	"Sixth_world_Suday/internal/dto"
	"Sixth_world_Suday/internal/ws"

	"github.com/google/uuid"
)

type roomsService struct {
	*core
}

func (r *roomsService) CreateGroupRoom(ctx context.Context, creatorID uuid.UUID, req dto.CreateGroupRoomRequest) (*dto.ChatRoomResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrMissingFields
	}

	channelKind := strings.TrimSpace(req.ChannelKind)
	if channelKind == "" {
		channelKind = "text"
	}
	if channelKind != "text" && channelKind != "voice" {
		return nil, ErrInvalidChannelKind
	}

	if err := r.filterTexts(ctx, name, req.Description); err != nil {
		return nil, err
	}
	if len(name) > 80 {
		name = name[:80]
	}
	description := strings.TrimSpace(req.Description)
	if len(description) > 500 {
		description = description[:500]
	}
	tags := sanitizeTags(req.Tags)

	roomID := uuid.New()
	if err := r.chatRepo.CreateRoom(ctx, roomID, name, description, "group", true, false, channelKind, creatorID); err != nil {
		return nil, fmt.Errorf("create group room: %w", err)
	}
	if len(tags) > 0 {
		if err := r.chatRepo.AddRoomTags(ctx, roomID, tags); err != nil {
			return nil, fmt.Errorf("add room tags: %w", err)
		}
	}
	if err := r.chatRepo.AddMemberWithRole(ctx, roomID, creatorID, "host", false); err != nil {
		return nil, fmt.Errorf("add creator to group: %w", err)
	}

	resp, err := r.buildRoomResponse(ctx, roomID, creatorID)
	if err != nil {
		return nil, err
	}

	r.hub.Broadcast(ws.Message{Type: "channel_created", Data: resp})

	return resp, nil
}

func (r *roomsService) UpdateRoom(ctx context.Context, roomID, userID uuid.UUID, req dto.UpdateRoomRequest) (*dto.ChatRoomResponse, error) {
	row, err := r.chatRepo.GetRoomByID(ctx, roomID, userID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if row == nil {
		return nil, ErrRoomNotFound
	}
	if row.IsSystem {
		return nil, ErrSystemRoom
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrMissingFields
	}

	description := strings.TrimSpace(req.Description)
	if err := r.filterTexts(ctx, name, description); err != nil {
		return nil, err
	}
	if len(name) > 80 {
		name = name[:80]
	}
	if len(description) > 500 {
		description = description[:500]
	}

	if err := r.chatRepo.UpdateRoom(ctx, roomID, name, description); err != nil {
		return nil, fmt.Errorf("update room: %w", err)
	}

	resp, err := r.buildRoomResponse(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}

	r.hub.Broadcast(ws.Message{Type: "channel_updated", Data: resp})

	return resp, nil
}

func (r *roomsService) ReorderChannels(ctx context.Context, categoryID uuid.UUID, roomIDs []uuid.UUID) error {
	if len(roomIDs) == 0 {
		return nil
	}

	if err := r.chatRepo.ReorderChannels(ctx, categoryID, roomIDs); err != nil {
		return fmt.Errorf("reorder channels: %w", err)
	}

	ids := make([]string, len(roomIDs))
	for i := 0; i < len(roomIDs); i++ {
		ids[i] = roomIDs[i].String()
	}

	r.hub.Broadcast(ws.Message{
		Type: "channels_reordered",
		Data: map[string]interface{}{
			"category_id": categoryID.String(),
			"room_ids":    ids,
		},
	})

	return nil
}

func (r *roomsService) ListCategories(ctx context.Context) ([]dto.ChatCategoryResponse, error) {
	rows, err := r.chatRepo.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	cats := make([]dto.ChatCategoryResponse, 0, len(rows))
	for i := 0; i < len(rows); i++ {
		cats = append(cats, dto.ChatCategoryResponse{
			ID:        rows[i].ID,
			Name:      rows[i].Name,
			Position:  rows[i].Position,
			IsBuiltin: rows[i].IsBuiltin,
			Kind:      rows[i].Kind,
		})
	}

	return cats, nil
}

func (r *roomsService) ListUserGroupRooms(ctx context.Context, userID uuid.UUID, search string, isRPOnly bool, tag, roleFilter string, includeArchived bool, limit, offset int) (*dto.ChatRoomListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if roleFilter != "host" && roleFilter != "member" {
		roleFilter = ""
	}

	tag = strings.ToLower(strings.TrimSpace(tag))
	rows, total, err := r.chatRepo.ListUserGroupRooms(ctx, userID, search, isRPOnly, tag, roleFilter, includeArchived, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list user group rooms: %w", err)
	}

	rooms := make([]dto.ChatRoomResponse, 0, len(rows))
	for i := range rows {
		rooms = append(rooms, r.rowToResponse(rows[i]))
	}
	return &dto.ChatRoomListResponse{Rooms: rooms, Total: total}, nil
}

func (r *roomsService) SetRoomMuted(ctx context.Context, roomID, userID uuid.UUID, muted bool) error {
	room, err := r.chatRepo.GetRoomByID(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}

	canAccess, err := r.canAccessChannel(ctx, userID, room)
	if err != nil {
		return err
	}
	if !canAccess {
		return ErrNotMember
	}

	if err := r.chatRepo.EnsureMember(ctx, roomID, userID); err != nil {
		return fmt.Errorf("ensure member: %w", err)
	}

	if err := r.chatRepo.SetMuted(ctx, roomID, userID, muted); err != nil {
		return fmt.Errorf("set muted: %w", err)
	}
	return nil
}

func (r *roomsService) IsRoomMuted(ctx context.Context, roomID, userID uuid.UUID) (bool, error) {
	return r.chatRepo.IsMuted(ctx, roomID, userID)
}

func (r *roomsService) ListRooms(ctx context.Context, userID uuid.UUID) (*dto.ChatRoomListResponse, error) {
	siteRole, err := r.authzSvc.GetRole(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get site role: %w", err)
	}

	rows, err := r.chatRepo.ListAllChannels(ctx, userID, siteRole.IsSiteStaff())
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}

	rooms := make([]dto.ChatRoomResponse, 0, len(rows))
	for i := 0; i < len(rows); i++ {
		rooms = append(rooms, r.rowToResponse(rows[i]))
	}

	return &dto.ChatRoomListResponse{Rooms: rooms, Total: len(rooms)}, nil
}

func (r *roomsService) ArchiveStale(ctx context.Context) (int, error) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	ids, err := r.chatRepo.ArchiveStaleGroupRooms(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("archive stale chat rooms: %w", err)
	}
	return len(ids), nil
}

// purgeRoomMessages validates the room (must exist, be a group, and not a system
// room), deletes all of its messages, and asynchronously removes the backing media
// files. Shared by DeleteChat (which then drops the room) and TruncateChat (which keeps it).
func (r *roomsService) purgeRoomMessages(ctx context.Context, roomID, userID uuid.UUID) error {
	row, err := r.chatRepo.GetRoomByID(ctx, roomID, userID)
	if err != nil {
		return fmt.Errorf("get room: %w", err)
	}
	if row == nil {
		return ErrRoomNotFound
	}
	if row.IsSystem {
		return ErrSystemRoom
	}

	mediaURLs, _ := r.chatRepo.ListRoomMediaURLs(ctx, roomID)

	if err := r.chatRepo.DeleteMessages(ctx, roomID); err != nil {
		return fmt.Errorf("delete messages: %w", err)
	}

	if len(mediaURLs) > 0 {
		go func(urls []string) {
			for i := 0; i < len(urls); i++ {
				_ = r.uploadSvc.Delete(urls[i])
			}
		}(mediaURLs)
	}

	return nil
}

func (r *roomsService) DeleteChat(ctx context.Context, roomID, userID uuid.UUID) error {
	if err := r.purgeRoomMessages(ctx, roomID, userID); err != nil {
		return err
	}

	if err := r.chatRepo.DeleteRoom(ctx, roomID); err != nil {
		return fmt.Errorf("delete room: %w", err)
	}

	r.hub.Broadcast(ws.Message{
		Type: "channel_deleted",
		Data: map[string]interface{}{
			"room_id": roomID,
		},
	})

	return nil
}

func (r *roomsService) TruncateChat(ctx context.Context, roomID, userID uuid.UUID) error {
	if err := r.purgeRoomMessages(ctx, roomID, userID); err != nil {
		return err
	}

	r.hub.Broadcast(ws.Message{
		Type: "chat_truncated",
		Data: map[string]interface{}{
			"room_id": roomID,
		},
	})

	return nil
}

func (r *roomsService) buildRoomResponse(ctx context.Context, roomID, viewerID uuid.UUID) (*dto.ChatRoomResponse, error) {
	row, err := r.chatRepo.GetRoomByID(ctx, roomID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}
	if row == nil {
		return nil, ErrNotMember
	}

	members, count, err := r.getRoomMemberResponses(ctx, roomID, viewerID)
	if err != nil {
		return nil, err
	}

	resp := r.rowToResponse(*row)
	resp.Members = members
	resp.MemberCount = count
	return &resp, nil
}

func (r *roomsService) SetRoomNickname(ctx context.Context, roomID, userID uuid.UUID, nickname string) (*dto.ChatRoomMemberResponse, error) {
	if err := r.filterTexts(ctx, nickname); err != nil {
		return nil, err
	}

	if err := r.ensureAccess(ctx, roomID, userID); err != nil {
		return nil, err
	}

	locked, err := r.effectiveLocked(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, ErrNicknameLocked
	}

	nickname = strings.TrimSpace(nickname)
	if len(nickname) > 32 {
		nickname = nickname[:32]
	}

	if err := r.chatRepo.SetMemberNickname(ctx, roomID, userID, nickname); err != nil {
		return nil, fmt.Errorf("set member nickname: %w", err)
	}

	name, possessive := r.nameAndPossessive(ctx, userID)
	if name != "" {
		if nickname == "" {
			r.postRoomActionMessage(ctx, roomID, userID, fmt.Sprintf("%s cleared %s alias.", name, possessive))
		} else {
			r.postRoomActionMessage(ctx, roomID, userID, fmt.Sprintf("%s changed %s alias.", name, possessive))
		}
	}

	return r.broadcastAndBuildMember(ctx, roomID, userID)
}

func (r *roomsService) SetRoomAvatar(ctx context.Context, roomID, userID uuid.UUID, contentType string, fileSize int64, reader io.Reader) (*dto.ChatRoomMemberResponse, error) {
	if err := r.ensureAccess(ctx, roomID, userID); err != nil {
		return nil, err
	}

	locked, err := r.effectiveLocked(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, ErrNicknameLocked
	}

	maxSize := int64(r.settingsSvc.GetInt(ctx, config.SettingMaxImageSize))
	subDir := fmt.Sprintf("chat-avatars/%s", roomID.String())
	avatarURL, err := r.uploadSvc.SaveImage(ctx, subDir, userID, fileSize, maxSize, reader)
	if err != nil {
		return nil, err
	}

	if err := r.chatRepo.SetMemberAvatar(ctx, roomID, userID, avatarURL); err != nil {
		return nil, fmt.Errorf("set member avatar: %w", err)
	}

	return r.broadcastAndBuildMember(ctx, roomID, userID)
}

func (r *roomsService) ClearRoomAvatar(ctx context.Context, roomID, userID uuid.UUID) (*dto.ChatRoomMemberResponse, error) {
	if err := r.ensureAccess(ctx, roomID, userID); err != nil {
		return nil, err
	}

	locked, err := r.effectiveLocked(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if locked {
		return nil, ErrNicknameLocked
	}

	rows, err := r.chatRepo.GetRoomMembersDetailed(ctx, roomID)
	if err == nil {
		for _, row := range rows {
			if row.UserID == userID && row.MemberAvatarURL != "" {
				_ = r.uploadSvc.Delete(row.MemberAvatarURL)
				break
			}
		}
	}

	if err := r.chatRepo.SetMemberAvatar(ctx, roomID, userID, ""); err != nil {
		return nil, fmt.Errorf("clear member avatar: %w", err)
	}

	return r.broadcastAndBuildMember(ctx, roomID, userID)
}
