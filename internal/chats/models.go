package chats

import (
	"time"

	"github.com/google/uuid"
)

type ChatType string

const (
	TypePrivate ChatType = "private"
	TypeGroup   ChatType = "group"
)

type Chat struct {
	ID        uuid.UUID  `json:"id"`
	Version   int64      `json:"version"`
	Type      ChatType   `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	Name      *string    `json:"name"`
	OwnerID   *uuid.UUID `json:"owner_id"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type Member struct {
	ChatID uuid.UUID `json:"chat_id"`
	UserID uuid.UUID `json:"user_id"`
}

type CreateChatRequest struct {
	ChatName  *string     `json:"chat_name"`
	MemberIDs []uuid.UUID `json:"member_ids"`
}

type PatchChat struct {
	ChatName        *string     `json:"chat_name"`
	OwnerID         *uuid.UUID  `json:"owner_id"`
	AddMemberIDs    []uuid.UUID `json:"add_member_ids"`
	RemoveMemberIDs []uuid.UUID `json:"remove_member_ids"`
}
