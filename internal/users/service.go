package users

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
	"github.com/wxvn/golang-messenger/internal/security"
)

type UserService struct {
	repository UserRepository
	cache      UserCache
}

func NewUserService(repository *UserRepository, cache UserCache) *UserService {
	return &UserService{
		repository: *repository,
		cache:      cache,
	}
}

func (s *UserService) GetUser(ctx context.Context, userID uuid.UUID) (User, error) {
	user, err := s.cache.Get(ctx, userID)
	if err == nil {
		return user, nil
	}

	user, err = s.repository.GetUser(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("get user from repository: %w", err)
	}

	_ = s.cache.Set(ctx, user)

	return user, nil
}

func (s *UserService) GetUsers(ctx context.Context, username *string, limit *int, offset *int) ([]User, error) {

	if limit != nil && *limit < 0 {
		return nil, fmt.Errorf(
			"limit must be non-negative: %w",
			errs.ErrInvalidArgument,
		)
	}

	if offset != nil && *offset < 0 {
		return nil, fmt.Errorf(
			"offset must be non-negative: %w",
			errs.ErrInvalidArgument,
		)
	}

	users, err := s.repository.GetUsers(ctx, username, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get users from repository: %w", err)
	}

	return users, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {

	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user from repository: %w", err)
	}

	if err := s.repository.DeleteUser(ctx, user); err != nil {
		return fmt.Errorf("delete user from repository: %w", err)
	}

	_ = s.cache.Delete(ctx, userID)

	return nil
}

func (s *UserService) UpdateUser(ctx context.Context, userID uuid.UUID, patchUser PatchUser) (User, error) {

	user, err := s.repository.GetUser(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("get user from repository: %w", err)
	}

	if patchUser.Username != nil {
		user.Username = *patchUser.Username
	}

	if patchUser.PasswordHash != nil {
		hash, err := security.HashPassword(*patchUser.PasswordHash)
		if err != nil {
			return User{}, fmt.Errorf("hash password: %w", err)
		}

		user.PasswordHash = hash
	}

	updatedUser, err := s.repository.UpdateUser(ctx, user)
	if err != nil {
		return User{}, fmt.Errorf("update user from repository: %w", err)
	}

	_ = s.cache.Delete(ctx, userID)

	return updatedUser, nil
}
