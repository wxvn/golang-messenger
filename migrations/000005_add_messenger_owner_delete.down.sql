DROP INDEX IF EXISTS idx_chats_owner_id;
DROP INDEX IF EXISTS idx_chats_deleted_at;

ALTER TABLE messenger.chats
DROP COLUMN IF EXISTS owner_id,
DROP COLUMN IF EXISTS deleted_at;
