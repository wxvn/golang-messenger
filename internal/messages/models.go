package messages

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	Version   int64
	ChatID    uuid.UUID
	SenderID  uuid.UUID
	Text      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type PatchMessage struct {
	Text *string `json:"text"`
}
