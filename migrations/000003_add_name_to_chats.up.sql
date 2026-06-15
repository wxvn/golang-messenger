ALTER TABLE messenger.chats
ADD COLUMN name TEXT
    CHECK (
        name IS NULL
        OR char_length(name) BETWEEN 1 AND 100
    );
