package chat

import (
	"context"
	"fmt"
	"time"

	"Sixth_world_Suday/internal/dto"
	"Sixth_world_Suday/internal/ws"

	"github.com/google/uuid"
)

type reactionsService struct {
	*core
}

func (r *reactionsService) PinMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	msg, err := r.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}
	if msg == nil {
		return ErrRoomNotFound
	}

	canMod, err := r.canModerateRoom(ctx, msg.RoomID, userID)
	if err != nil {
		return err
	}
	if !canMod {
		return ErrNotHost
	}

	if err := r.chatRepo.PinMessage(ctx, messageID, userID); err != nil {
		return fmt.Errorf("pin message: %w", err)
	}

	r.broadcastToRoomMembers(ctx, msg.RoomID, ws.Message{
		Type: "chat_message_pinned",
		Data: map[string]interface{}{
			"room_id":    msg.RoomID,
			"message_id": messageID,
			"pinned_at":  time.Now().UTC().Format(time.RFC3339),
			"pinned_by":  userID,
		},
	})
	return nil
}

func (r *reactionsService) UnpinMessage(ctx context.Context, messageID, userID uuid.UUID) error {
	msg, err := r.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}
	if msg == nil {
		return ErrRoomNotFound
	}
	if msg.PinnedAt == nil {
		return ErrMessageNotPinned
	}

	canMod, err := r.canModerateRoom(ctx, msg.RoomID, userID)
	if err != nil {
		return err
	}
	if !canMod {
		return ErrNotHost
	}

	if err := r.chatRepo.UnpinMessage(ctx, messageID); err != nil {
		return fmt.Errorf("unpin message: %w", err)
	}

	r.broadcastToRoomMembers(ctx, msg.RoomID, ws.Message{
		Type: "chat_message_unpinned",
		Data: map[string]interface{}{
			"room_id":    msg.RoomID,
			"message_id": messageID,
		},
	})
	return nil
}

func (r *reactionsService) ListPinnedMessages(ctx context.Context, roomID, viewerID uuid.UUID) (*dto.ChatMessageListResponse, error) {
	room, err := r.chatRepo.GetRoomByID(ctx, roomID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("get room: %w", err)
	}

	canAccess, err := r.canAccessChannel(ctx, viewerID, room)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, ErrNotMember
	}

	rows, err := r.chatRepo.ListPinnedMessages(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("list pinned messages: %w", err)
	}

	messages := r.hydrateMessageRows(ctx, viewerID, rows)
	return &dto.ChatMessageListResponse{
		Messages: messages,
		Total:    len(messages),
	}, nil
}

func (r *reactionsService) resolveMemberDisplayName(ctx context.Context, roomID, userID uuid.UUID) string {
	user, err := r.userRepo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return ""
	}
	name := user.DisplayName
	if name == "" {
		name = user.Username
	}

	rows, _ := r.chatRepo.GetRoomMembersDetailed(ctx, roomID)
	for _, mr := range rows {
		if mr.UserID == userID {
			if mr.Nickname != "" {
				name = mr.Nickname
			}
			break
		}
	}
	return name
}

func (r *reactionsService) AddReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if err := validateEmoji(emoji); err != nil {
		return err
	}

	msg, err := r.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}
	if msg == nil {
		return ErrRoomNotFound
	}

	room, err := r.chatRepo.GetRoomByID(ctx, msg.RoomID, userID)
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

	if err := r.checkSenderTimeout(ctx, msg.RoomID, userID); err != nil {
		return err
	}

	inserted, err := r.chatRepo.AddReaction(ctx, messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}
	if !inserted {
		return nil
	}

	displayName := r.resolveMemberDisplayName(ctx, msg.RoomID, userID)
	count, _ := r.chatRepo.CountReactions(ctx, messageID, emoji)

	r.broadcastToRoomMembers(ctx, msg.RoomID, ws.Message{
		Type: "chat_reaction_added",
		Data: map[string]interface{}{
			"room_id":      msg.RoomID,
			"message_id":   messageID,
			"emoji":        emoji,
			"user_id":      userID,
			"display_name": displayName,
			"count":        count,
		},
	})

	if msg.SenderID != userID && !msg.IsSystem && !r.hub.IsUserViewing(msg.RoomID, msg.SenderID) {
		muted, _ := r.chatRepo.IsMuted(ctx, msg.RoomID, msg.SenderID)
		if !muted {
			r.sideEffectsWG.Add(1)
			go r.notifyReaction(msg.RoomID, messageID, msg.SenderID, userID, emoji)
		}
	}

	return nil
}

func (r *reactionsService) notifyReaction(roomID, messageID, recipientID, actorID uuid.UUID, emoji string) {
	defer r.sideEffectsWG.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_ = r.notifSvc.Notify(ctx, dto.NotifyParams{
		RecipientID:   recipientID,
		ActorID:       actorID,
		Type:          dto.NotifChatReaction,
		ReferenceID:   roomID,
		ReferenceType: fmt.Sprintf("chat_message:%s", messageID),
		Message:       fmt.Sprintf("reacted %s to your message", emoji),
	})
}

func (r *reactionsService) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if err := validateEmoji(emoji); err != nil {
		return err
	}

	msg, err := r.chatRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("get message: %w", err)
	}
	if msg == nil {
		return ErrRoomNotFound
	}

	room, err := r.chatRepo.GetRoomByID(ctx, msg.RoomID, userID)
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

	deleted, err := r.chatRepo.RemoveReaction(ctx, messageID, userID, emoji)
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	if !deleted {
		return nil
	}

	displayName := r.resolveMemberDisplayName(ctx, msg.RoomID, userID)
	count, _ := r.chatRepo.CountReactions(ctx, messageID, emoji)

	r.broadcastToRoomMembers(ctx, msg.RoomID, ws.Message{
		Type: "chat_reaction_removed",
		Data: map[string]interface{}{
			"room_id":      msg.RoomID,
			"message_id":   messageID,
			"emoji":        emoji,
			"user_id":      userID,
			"display_name": displayName,
			"count":        count,
		},
	})
	return nil
}
