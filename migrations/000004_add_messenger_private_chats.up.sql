CREATE TABLE messenger.private_chats (
    user_low UUID NOT NULL,
    user_high UUID NOT NULL,
    chat_id UUID NOT NULL UNIQUE REFERENCES messenger.chats(id),

    CHECK (user_low < user_high),

    PRIMARY KEY (user_low, user_high)
);
