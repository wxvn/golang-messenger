CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE SCHEMA IF NOT EXISTS messenger;
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE messenger.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version BIGINT NOT NULL DEFAULT 1,

    username VARCHAR(40) NOT NULL UNIQUE
        CHECK (char_length(username) BETWEEN 2 AND 40),

    password_hash TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE messenger.chats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version BIGINT NOT NULL DEFAULT 1,

    type TEXT NOT NULL
        CHECK (type IN ('private', 'group')),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE messenger.chat_members (
    chat_id UUID NOT NULL REFERENCES messenger.chats(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES messenger.users(id) ON DELETE CASCADE,

    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messenger.messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version BIGINT NOT NULL DEFAULT 1,

    chat_id UUID NOT NULL REFERENCES messenger.chats(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES messenger.users(id) ON DELETE RESTRICT,

    text TEXT NOT NULL
        CHECK (char_length(text) BETWEEN 1 AND 1000),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version BIGINT NOT NULL DEFAULT 1,

    user_id UUID NOT NULL REFERENCES messenger.users(id) ON DELETE CASCADE,

    token_hash TEXT NOT NULL UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ
);

CREATE INDEX idx_chat_members_user_id
    ON messenger.chat_members(user_id);

CREATE INDEX idx_messages_chat_created_at
    ON messenger.messages(chat_id, created_at);

CREATE INDEX idx_messages_sender_id
    ON messenger.messages(sender_id);

CREATE INDEX idx_refresh_tokens_user_id
    ON auth.refresh_tokens(user_id);

CREATE INDEX idx_refresh_tokens_token_hash
    ON auth.refresh_tokens(token_hash);
