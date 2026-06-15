package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/postgres"
)

type Repository struct {
	pool *postgres.Pool
}

func NewRepository(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) SignUp(ctx context.Context, u User, tokenHash string, expiresAt time.Time) (User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	userQuery := `
		INSERT INTO messenger.users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id, version, username, created_at;
`

	var createdUser User
	err = tx.QueryRow(ctx, userQuery, u.Username, u.PasswordHash).Scan(
		&createdUser.ID, &createdUser.Version, &createdUser.Username, &createdUser.CreatedAt,
	)
	if err != nil {
		return User{}, fmt.Errorf("insert user: %w", err)
	}

	tokenQuery := `
		INSERT INTO auth.refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, tokenQuery, createdUser.ID, tokenHash, expiresAt)
	if err != nil {
		return User{}, fmt.Errorf("insert refresh token: %w", err)
	}

	return createdUser, tx.Commit(ctx)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (User, error) {
	ctx, canel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer canel()

	query := `
	SELECT id, version, username, password_hash, created_at, deleted_at
	FROM messenger.users
	WHERE username=$1;
	`

	row := r.pool.QueryRow(ctx, query, username)

	var user User
	err := row.Scan(&user.ID, &user.Version, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, errs.ErrNotFound
		}
		return User{}, fmt.Errorf("scan user: %w", err)
	}

	return user, nil
}

func (r *Repository) SaveRefreshToken(ctx context.Context, userID uuid.UUID, hashToken string, expiresAt time.Time) error {
	ctx, cansel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cansel()

	query := `
		INSERT INTO auth.refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := r.pool.Exec(ctx, query, userID, hashToken, expiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (r *Repository) RevokeToken(ctx context.Context, tokenHash string, revokedAt time.Time) error {

	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		UPDATE auth.refresh_tokens
		SET revoked_at = $1
		WHERE token_hash = $2
		  AND revoked_at IS NULL
	`

	cmdTag, err := r.pool.Exec(
		ctx,
		query,
		revokedAt,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return errs.ErrConflict
	}

	return nil
}

func (r *Repository) GetTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
		SELECT id, version, user_id, token_hash, created_at, expires_at, revoked_at
		FROM auth.refresh_tokens
		WHERE token_hash=$1;
	`
	row := r.pool.QueryRow(ctx, query, tokenHash)

	var token RefreshToken
	err := row.Scan(&token.ID, &token.Version, &token.UserID, &token.TokenHash, &token.CreatedAt, &token.ExpiresAt, &token.RevokedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RefreshToken{}, errs.ErrNotFound
		}
		return RefreshToken{}, fmt.Errorf("scan error: %w", err)
	}

	return token, nil
}
