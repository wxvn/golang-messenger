package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/wxvn/golang-messenger/internal/jwt"
)

type Service struct {
	repository   *Repository
	tokenManager *jwt.TokenManager
}

func NewService(repo *Repository, tm *jwt.TokenManager) *Service {
	return &Service{repository: repo, tokenManager: tm}
}

func (s *Service) SignUp(ctx context.Context, username, password string) (SignResponse, error) {
	hashPassword, _ := HashPassword(password)

	user := User{Username: username, PasswordHash: string(hashPassword)}

	tokenRaw, tokenHash := GenerateRefreshToken()

	createdUser, err := s.repository.SignUp(ctx, user, tokenHash, time.Now().Add(30*24*time.Hour))
	if err != nil {
		return SignResponse{}, err
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(createdUser.ID.String())
	if err != nil {
		return SignResponse{}, err
	}

	return SignResponse{
		User: createdUser,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: tokenRaw,
		},
	}, nil
}

func (s *Service) SignIn(ctx context.Context, username, password string) (Tokens, error) {
	user, err := s.repository.GetUserByUsername(ctx, username)
	if err != nil {
		return Tokens{}, fmt.Errorf("get user from repository: %w", err)
	}

	if err := CheckPassword(user.PasswordHash, password); err != nil {
		return Tokens{}, fmt.Errorf("get user from repository: %w", err)
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID.String())
	if err != nil {
		return Tokens{}, err
	}
	tokenRaw, tokenHash := GenerateRefreshToken()

	if err := s.repository.SaveRefreshToken(ctx, user.ID, tokenHash, time.Now().Add(30*24*time.Hour)); err != nil {
		return Tokens{}, fmt.Errorf("save token from repository: %w", err)
	}

	return Tokens{
		AccessToken:  accessToken,
		RefreshToken: tokenRaw,
	}, nil

}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	TokenHash := HashRefreshToken(refreshToken)

	revokeAt := time.Now()

	if err := s.repository.RevokeToken(ctx, TokenHash, revokeAt); err != nil {
		return fmt.Errorf("revoke token from repository: %w", err)
	}
	return nil
}

func (s *Service) Refresh(ctx context.Context, tokens Tokens) (Tokens, error) {
	TokenHash := HashRefreshToken(tokens.RefreshToken)

	refreshToken, err := s.repository.GetTokenByHash(ctx, TokenHash)
	if err != nil {
		return Tokens{}, fmt.Errorf("get refresh toke from repository: %w", err)
	}

	if refreshToken.RevokedAt != nil {
		return Tokens{}, fmt.Errorf("token is revoked")
	}

	accesToken, err := s.tokenManager.GenerateAccessToken(refreshToken.UserID.String())
	if err != nil {
		return Tokens{}, err
	}

	tokenRaw, tokenHash := GenerateRefreshToken()

	err = s.repository.SaveRefreshToken(ctx, refreshToken.UserID, tokenHash, time.Now().Add(30*24*time.Hour))
	if err != nil {
		return Tokens{}, fmt.Errorf("save token from repository: %w", err)
	}

	return Tokens{
		AccessToken:  accesToken,
		RefreshToken: tokenRaw,
	}, nil

}
