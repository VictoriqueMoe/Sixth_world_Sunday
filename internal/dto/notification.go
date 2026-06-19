package dto

import "github.com/google/uuid"

type (
	NotificationType string

	NotifyParams struct {
		RecipientID   uuid.UUID
		Type          NotificationType
		ReferenceID   uuid.UUID
		ReferenceType string
		ActorID       uuid.UUID
		Message       string
		EmailSubject  string
		EmailBody     string
	}

	NotificationResponse struct {
		ID            int              `json:"id"`
		Type          NotificationType `json:"type"`
		ReferenceID   uuid.UUID        `json:"reference_id"`
		ReferenceType string           `json:"reference_type"`
		Actor         UserResponse     `json:"actor"`
		Message       string           `json:"message,omitempty"`
		Read          bool             `json:"read"`
		CreatedAt     string           `json:"created_at"`
		Count         int              `json:"count"`
	}

	NotificationListResponse struct {
		Notifications []NotificationResponse `json:"notifications"`
		Total         int                    `json:"total"`
		Limit         int                    `json:"limit"`
		Offset        int                    `json:"offset"`
	}

	UnreadCountResponse struct {
		Count int `json:"count"`
	}
)

const (
	NotifChatRoomMessage  NotificationType = "chat_room_message"
	NotifChatMention      NotificationType = "chat_mention"
	NotifChatReply        NotificationType = "chat_reply"
	NotifChatReaction     NotificationType = "chat_reaction"
	NotifChatRoomInvite   NotificationType = "chat_room_invite"
	NotifChatRoomBanned   NotificationType = "chat_room_banned"
	NotifChatRoomKicked   NotificationType = "chat_room_kicked"
	NotifChatRoomUnbanned NotificationType = "chat_room_unbanned"
	NotifReport           NotificationType = "report"
	NotifReportResolved   NotificationType = "report_resolved"
)
