package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/postgres"
)

type UserRepositoy struct {
	pool *postgres.Pool
}

func NewUserRepositoy(pool *postgres.Pool) *UserRepositoy {
	return &UserRepositoy{pool: pool}
}

func (r *UserRepositoy) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT id, version, username, password_hash, created_at, deleted_at
	FROM messenger.users
	WHERE id=$1 AND deleted_at IS NULL;
	`

	row := r.pool.QueryRow(ctx, query, userID)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Version,
		&user.Username,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, errs.ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	return user, nil
}

func (r *UserRepositoy) GetUsers(
	ctx context.Context,
	username *string,
	limit, offset *int,
) ([]User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, username, created_at
		FROM messenger.users
		WHERE deleted_at IS NULL
	`

	args := []any{}
	i := 1

	if username != nil {
		query += fmt.Sprintf(" AND username = $%d", i)
		args = append(args, *username)
		i++
	}

	query += " ORDER BY created_at DESC"

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
		return nil, fmt.Errorf("select users: %w", err)
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User

		err := rows.Scan(
			&u.ID,
			&u.Version,
			&u.Username,
			&u.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan users: %w", err)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	return users, nil
}

func (r *UserRepositoy) UpdateUser(ctx context.Context, user User) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.users
		SET
			version = version + 1,
			username = $3,
			password_hash = $4
		WHERE id = $1
		  AND version = $2
		  AND deleted_at IS NULL
		RETURNING
			id,
			version,
			username,
			password_hash,
			created_at,
			deleted_at;
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		user.ID,
		user.Version,
		user.Username,
		user.PasswordHash,
	)

	var updated User

	err := row.Scan(
		&updated.ID,
		&updated.Version,
		&updated.Username,
		&updated.PasswordHash,
		&updated.CreatedAt,
		&updated.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, errs.ErrConflict
		}
		return User{}, fmt.Errorf("update user: %w", err)
	}

	return updated, nil
}

func (r *UserRepositoy) DeleteUser(ctx context.Context, user User) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE messenger.users
		SET
			version = version + 1,
			deleted_at = NOW()
		WHERE id = $1
		  AND version = $2
		  AND deleted_at IS NULL;
	`

	cmdTag, err := r.pool.Exec(ctx, query, user.ID, user.Version)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return errs.ErrConflict
	}

	return nil
}
