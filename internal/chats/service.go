package chats

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
)

type ChatsService struct {
	repository *ChatsRepository
}

func NewChatsService(repo *ChatsRepository) *ChatsService {
	return &ChatsService{
		repository: repo,
	}
}

func (s *ChatsService) GetChat(ctx context.Context, chatID uuid.UUID, userID uuid.UUID) (Chat, error) {
	ok, err := s.repository.IsMember(ctx, chatID, userID)
	if err != nil {
		return Chat{}, err
	}

	if !ok {
		return Chat{}, nil
	}

	return s.repository.GetByID(ctx, chatID)
}

func (s *ChatsService) GetChats(ctx context.Context, userID uuid.UUID) ([]Chat, error) {
	return s.repository.GetByUserID(ctx, userID)
}

func (s *ChatsService) CreateChatGroup(ctx context.Context, creatorID uuid.UUID, chatName *string, memberIDs []uuid.UUID) (Chat, error) {
	if chatName == nil || len(*chatName) == 0 {
		return Chat{}, fmt.Errorf("group chat name is required")
	}

	finalMembers := append(memberIDs, creatorID)

	uniqueMembers := make(map[uuid.UUID]struct{})
	for _, id := range finalMembers {
		uniqueMembers[id] = struct{}{}
	}

	var members []uuid.UUID
	for id := range uniqueMembers {
		members = append(members, id)
	}

	chat := Chat{
		Name:    chatName,
		Type:    TypeGroup,
		OwnerID: &creatorID,
	}

	created, err := s.repository.CreateGroupChat(ctx, &chat, members)
	if err != nil {
		return Chat{}, fmt.Errorf("create chat from repository: %w", err)
	}

	return created, nil
}

func (s *ChatsService) CanManage(ctx context.Context, chatID, userID uuid.UUID) (bool, error) {
	chat, err := s.repository.GetByID(ctx, chatID)
	if err != nil {
		return false, err
	}

	if chat.OwnerID != nil {
		return *chat.OwnerID == userID, nil
	}

	return s.repository.IsMember(ctx, chatID, userID)
}

func (s *ChatsService) UpdateGroupChat(ctx context.Context, patchChat PatchChat, userID uuid.UUID, chatID uuid.UUID) (Chat, error) {
	chat, err := s.repository.GetByID(ctx, chatID)
	if err != nil {
		return Chat{}, fmt.Errorf("get chat: %w", err)
	}

	if chat.Type != TypeGroup {
		return Chat{}, fmt.Errorf("chat is not group")
	}

	if chat.OwnerID != nil {
		if *chat.OwnerID != userID {
			return Chat{}, errs.ErrForbidden
		}
	} else {
		isMember, err := s.repository.IsMember(ctx, chatID, userID)
		if err != nil {
			return Chat{}, fmt.Errorf("check membership: %w", err)
		}
		if !isMember {
			return Chat{}, errs.ErrForbidden
		}
	}

	if patchChat.ChatName != nil {
		chat.Name = patchChat.ChatName
	}

	if patchChat.OwnerID != nil {
		chat.OwnerID = patchChat.OwnerID
	}

	updatedChat, err := s.repository.UpdateGroupChat(ctx, chat)
	if err != nil {
		return Chat{}, fmt.Errorf("update chat: %w", err)
	}

	for _, u := range patchChat.AddMemberIDs {
		if err := s.repository.AddMember(ctx, chatID, u); err != nil {
			return Chat{}, fmt.Errorf("add member: %w", err)
		}
	}

	for _, u := range patchChat.RemoveMemberIDs {
		if updatedChat.OwnerID != nil && u == *updatedChat.OwnerID {
			continue
		}

		if err := s.repository.RemoveMember(ctx, chatID, u); err != nil {
			return Chat{}, fmt.Errorf("remove member: %w", err)
		}
	}

	return updatedChat, nil
}

func (s *ChatsService) DeleteChat(ctx context.Context, userID uuid.UUID, chatID uuid.UUID) error {
	chat, err := s.repository.GetByID(ctx, chatID)
	if err != nil {
		return fmt.Errorf("get chat from repostiry: %w", err)
	}

	if chat.OwnerID != nil {
		if *chat.OwnerID != userID {
			return errs.ErrForbidden
		}
		return nil
	}

	isMember, err := s.repository.IsMember(ctx, chatID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return errs.ErrForbidden
	}

	if err := s.repository.DeleteChat(ctx, chat); err != nil {
		return fmt.Errorf("delte chat from repository: %w", err)
	}

	return nil
}
