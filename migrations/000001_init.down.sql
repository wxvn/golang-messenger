DROP INDEX IF EXISTS auth.idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS auth.idx_refresh_tokens_user_id;

DROP INDEX IF EXISTS messenger.idx_messages_sender_id;
DROP INDEX IF EXISTS messenger.idx_messages_chat_created_at;
DROP INDEX IF EXISTS messenger.idx_chat_members_user_id;

DROP TABLE IF EXISTS auth.refresh_tokens;

DROP TABLE IF EXISTS messenger.messages;
DROP TABLE IF EXISTS messenger.chat_members;
DROP TABLE IF EXISTS messenger.chats;
DROP TABLE IF EXISTS messenger.users;

DROP SCHEMA IF EXISTS auth;
DROP SCHEMA IF EXISTS messenger;

DROP EXTENSION IF EXISTS pgcrypto;
