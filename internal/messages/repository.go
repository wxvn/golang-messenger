package messages

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/postgres"
)

type MessagesRepository struct {
	pool *postgres.Pool
}

func NewMessagesRepository(pool *postgres.Pool) *MessagesRepository {
	return &MessagesRepository{pool: pool}
}

func (r *MessagesRepository) CreateMessage(ctx context.Context, m Message) (Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		INSERT INTO messenger.messages (chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			version,
			chat_id,
			sender_id,
			text,
			created_at,
			deleted_at;
	`

	row := r.pool.QueryRow(ctx, query, m.ChatID, m.SenderID, m.Text)

	var created Message

	err := row.Scan(
		&created.ID,
		&created.Version,
		&created.ChatID,
		&created.SenderID,
		&created.Text,
		&created.CreatedAt,
		&created.DeletedAt,
	)

	if err != nil {
		return Message{}, fmt.Errorf("create message: %w", err)
	}

	return created, nil
}

func (r *MessagesRepository) GetByChatID(ctx context.Context, chatID uuid.UUID, limit, offset *int) ([]Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT
			id,
			version,
			chat_id,
			sender_id,
			text,
			created_at,
			deleted_at
		FROM messenger.messages
		WHERE chat_id = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	args := []any{chatID}
	i := 2

	if limit != nil {
		query += fmt.Sprintf(" LIMIT $%d", i)
		args = append(args, *limit)
		i++
	}

	if offset != nil {
		query += fmt.Sprintf(" OFFSET $%d", i)
		args = append(args, *offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select messages: %w", err)
	}
	defer rows.Close()

	var messages []Message

	for rows.Next() {
		var m Message

		err := rows.Scan(
			&m.ID,
			&m.Version,
			&m.ChatID,
			&m.SenderID,
			&m.Text,
			&m.CreatedAt,
			&m.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}

		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return messages, nil
}

func (r *MessagesRepository) GetByID(ctx context.Context, messageID uuid.UUID) (Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT
			id,
			version,
			chat_id,
			sender_id,
			text,
			created_at,
			deleted_at
		FROM messenger.messages
		WHERE id = $1 AND deleted_at IS NULL;
	`

	var m Message

	err := r.pool.QueryRow(ctx, query, messageID).Scan(
		&m.ID,
		&m.Version,
		&m.ChatID,
		&m.SenderID,
		&m.Text,
		&m.CreatedAt,
		&m.DeletedAt,
	)

	if err != nil {
		return Message{}, fmt.Errorf("get message: %w", err)
	}

	return m, nil
}

func (r *MessagesRepository) UpdateMessage(ctx context.Context, m Message) (Message, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.messages
		SET
			text = $2,
			version = version + 1
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING
			id,
			version,
			chat_id,
			sender_id,
			text,
			created_at,
			deleted_at;
	`

	var updated Message

	err := r.pool.QueryRow(ctx, query, m.ID, m.Text).Scan(
		&updated.ID,
		&updated.Version,
		&updated.ChatID,
		&updated.SenderID,
		&updated.Text,
		&updated.CreatedAt,
		&updated.DeletedAt,
	)

	if err != nil {
		return Message{}, fmt.Errorf("update message: %w", err)
	}

	return updated, nil
}

func (r *MessagesRepository) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.messages
		SET
			deleted_at = now(),
			version = version + 1
		WHERE id = $1 AND deleted_at IS NULL;
	`

	_, err := r.pool.Exec(ctx, query, messageID)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
	}

	return nil
}
