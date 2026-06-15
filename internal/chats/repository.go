package chats

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/postgres"
)

type ChatsRepository struct {
	pool *postgres.Pool
}

func NewChatsRepository(pool *postgres.Pool) *ChatsRepository {
	return &ChatsRepository{pool: pool}
}

func (r *ChatsRepository) CreateGroupChat(ctx context.Context, chat *Chat, memberIDs []uuid.UUID) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Chat{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO messenger.chats (type, name, owner_id)
        VALUES ($1, $2, $3)
        RETURNING id, version, type, created_at, name, owner_id, deleted_at;
    `

	var created Chat

	err = tx.QueryRow(ctx, query, chat.Type, chat.Name, chat.OwnerID).Scan(
		&created.ID,
		&created.Version,
		&created.Type,
		&created.CreatedAt,
		&created.Name,
		&created.OwnerID,
		&created.DeletedAt,
	)
	if err != nil {
		return Chat{}, fmt.Errorf("create chat: %w", err)
	}

	memberQuery := `
        INSERT INTO messenger.chat_members (chat_id, user_id)
        VALUES ($1, $2)
    `

	for _, userID := range memberIDs {
		_, err = tx.Exec(ctx, memberQuery, created.ID, userID)
		if err != nil {
			return Chat{}, fmt.Errorf("insert member %s: %w", userID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Chat{}, fmt.Errorf("commit tx: %w", err)
	}

	return created, nil
}

func (r *ChatsRepository) AddMember(ctx context.Context, chatID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO messenger.chat_members (chat_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING;
	`

	_, err := r.pool.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	return nil
}

func (r *ChatsRepository) RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		DELETE FROM messenger.chat_members
		WHERE chat_id = $1
		  AND user_id = $2;
	`

	_, err := r.pool.Exec(ctx, query, chatID, userID)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}

	return nil
}

func (r *ChatsRepository) GetByID(ctx context.Context, chatID uuid.UUID) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, type, created_at, name, owner_id, deleted_at
		FROM messenger.chats
		WHERE id = $1;
	`

	row := r.pool.QueryRow(ctx, query, chatID)

	var chat Chat

	err := row.Scan(
		&chat.ID, &chat.Version, &chat.Type, &chat.CreatedAt,
		&chat.Name, &chat.OwnerID, &chat.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Chat{}, errs.ErrNotFound
		}

		return Chat{}, fmt.Errorf("get chat: %w", err)
	}

	return chat, nil
}

func (r *ChatsRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT chats.id, chats.version, chats.type, chats.created_at, chats.name, chats.owner_id, chats.deleted_at
        FROM messenger.chats chats
        INNER JOIN messenger.chat_members members ON members.chat_id = chats.id
        WHERE members.user_id = $1 AND chats.deleted_at IS NULL
        ORDER BY chats.created_at DESC;
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("select chats: %w", err)
	}
	defer rows.Close()

	var chats []Chat

	for rows.Next() {
		var chat Chat

		err := rows.Scan(
			&chat.ID, &chat.Version, &chat.Type, &chat.CreatedAt,
			&chat.Name, &chat.OwnerID, &chat.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan chats: %w", err)
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return chats, nil
}

func (r *ChatsRepository) IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
        SELECT EXISTS (
            SELECT 1
            FROM messenger.chat_members cm
            INNER JOIN messenger.chats c ON cm.chat_id = c.id
            WHERE cm.chat_id = $1
              AND cm.user_id = $2
              AND c.deleted_at IS NULL
        );
    `

	var exists bool
	err := r.pool.QueryRow(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check member: %w", err)
	}

	return exists, nil
}

func (r *ChatsRepository) GetOrCreatePrivateChat(ctx context.Context, userA, userB uuid.UUID) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	if userA == userB {
		return Chat{}, fmt.Errorf("cannot create private chat with same user")
	}

	if bytes.Compare(userA[:], userB[:]) > 0 {
		userA, userB = userB, userA
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Chat{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var chat Chat

	err = tx.QueryRow(ctx, `
		SELECT
			c.id, c.version, c.type, c.created_at, c.name, c.owner_id, c.deleted_at
		FROM messenger.private_chats pc
		JOIN messenger.chats c ON c.id = pc.chat_id
		WHERE pc.user_low = $1
		  AND pc.user_high = $2
		  AND c.deleted_at IS NULL
	`,
		userA,
		userB,
	).Scan(
		&chat.ID,
		&chat.Version,
		&chat.Type,
		&chat.CreatedAt,
		&chat.Name,
		&chat.OwnerID,
		&chat.DeletedAt,
	)

	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return Chat{}, err
		}
		return chat, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return Chat{}, fmt.Errorf("find private chat: %w", err)
	}

	err = tx.QueryRow(ctx, `
		INSERT INTO messenger.chats(type)
		VALUES ('private')
		RETURNING id, version, type, created_at, name, owner_id, deleted_at
	`).Scan(
		&chat.ID,
		&chat.Version,
		&chat.Type,
		&chat.CreatedAt,
		&chat.Name,
		&chat.OwnerID,
		&chat.DeletedAt,
	)

	if err != nil {
		return Chat{}, fmt.Errorf("create chat: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO messenger.private_chats (user_low, user_high, chat_id)
		VALUES ($1, $2, $3)
	`,
		userA,
		userB,
		chat.ID,
	)

	if err != nil {
		return Chat{}, fmt.Errorf("create private chat: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO messenger.chat_members(chat_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`,
		chat.ID,
		userA,
		userB,
	)

	if err != nil {
		return Chat{}, fmt.Errorf("insert members: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Chat{}, fmt.Errorf("commit tx: %w", err)
	}

	return chat, nil
}

func (r *ChatsRepository) UpdateGroupChat(ctx context.Context, chat Chat) (Chat, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.chats
		SET
			version = version + 1,
			name = $3,
			owner_id = $4
		WHERE id = $1
		  AND version = $2
		  AND deleted_at IS NULL
		RETURNING
			id,
			version,
			type,
			created_at,
			name,
			owner_id,
			deleted_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		chat.ID,
		chat.Version,
		chat.Name,
		chat.OwnerID,
	)

	var updated Chat

	err := row.Scan(
		&updated.ID,
		&updated.Version,
		&updated.Type,
		&updated.CreatedAt,
		&updated.Name,
		&updated.OwnerID,
		&updated.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Chat{}, errs.ErrConflict
		}

		return Chat{}, fmt.Errorf("update chat: %w", err)
	}

	return updated, nil
}

func (r *ChatsRepository) DeleteChat(ctx context.Context, chat Chat) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.chats
		SET
			version = version + 1,
			deleted_at = NOW()
		WHERE id = $1
		  AND version = $2
	`

	tag, err := r.pool.Exec(ctx, query, chat.ID, chat.Version)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return errs.ErrConflict
	}

	return nil
}

func (r *ChatsRepository) GetChatMembers(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT user_id
		FROM messenger.chat_members
		WHERE chat_id = $1
	`

	rows, err := r.pool.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("get chat members: %w", err)
	}
	defer rows.Close()

	var members []uuid.UUID

	for rows.Next() {
		var userID uuid.UUID

		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}

		members = append(members, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return members, nil
}
