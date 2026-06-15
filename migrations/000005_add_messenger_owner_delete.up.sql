ALTER TABLE messenger.chats
ADD COLUMN owner_id UUID REFERENCES messenger.users(id) ON DELETE SET NULL,
ADD COLUMN deleted_at TIMESTAMPTZ;


CREATE INDEX idx_chats_owner_id ON messenger.chats(owner_id);

CREATE INDEX idx_chats_deleted_at ON messenger.chats(deleted_at) WHERE deleted_at IS NOT NULL;
