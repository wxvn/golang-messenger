package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID  `json:"id"`
	Version   int64      `json:"version"`
	ChatID    uuid.UUID  `json:"chat_id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	Text      string     `json:"text"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type SendMessageRequest struct {
	ChatID   *uuid.UUID `json:"chat_id"`
	ToUserID *uuid.UUID `json:"to_user_id"`
	Text     string     `json:"text"`
}

type OutgoingMessage struct {
	Type      string    `json:"type"`
	ChatID    uuid.UUID `json:"chat_id"`
	MessageID uuid.UUID `json:"message_id"`
	SenderID  uuid.UUID `json:"sender_id"`
	Text      string    `json:"text"`
}
