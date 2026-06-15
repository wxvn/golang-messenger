package messages

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/wxvn/golang-messenger/internal/chats"
	"github.com/wxvn/golang-messenger/internal/models"
)

type ChatsRepo interface {
	GetOrCreatePrivateChat(ctx context.Context, user1, user2 uuid.UUID) (chats.Chat, error)
	IsMember(ctx context.Context, chatID, userID uuid.UUID) (bool, error)
	GetChatMembers(ctx context.Context, chatID uuid.UUID) ([]uuid.UUID, error)
}

type MessagesService struct {
	repo      *MessagesRepository
	chatsRepo ChatsRepo
}

func NewMessagesService(repo *MessagesRepository, chatsRepo ChatsRepo) *MessagesService {
	return &MessagesService{
		repo:      repo,
		chatsRepo: chatsRepo,
	}
}

func (s *MessagesService) SendMessage(ctx context.Context, req models.SendMessageRequest, senderID uuid.UUID) (models.Message, []uuid.UUID, error) {

	var chatID uuid.UUID

	switch {
	case req.ChatID != nil:
		chatID = *req.ChatID

	case req.ToUserID != nil:
		chat, err := s.chatsRepo.GetOrCreatePrivateChat(ctx, senderID, *req.ToUserID)
		if err != nil {
			return models.Message{}, nil, err
		}
		chatID = chat.ID

	default:
		return models.Message{}, nil, fmt.Errorf("chat_id or to_user_id required")
	}

	ok, err := s.chatsRepo.IsMember(ctx, chatID, senderID)
	if err != nil {
		return models.Message{}, nil, err
	}
	if !ok {
		return models.Message{}, nil, fmt.Errorf("access denied")
	}

	localMsg := Message{
		ChatID:   chatID,
		SenderID: senderID,
		Text:     req.Text,
	}

	createdLocal, err := s.repo.CreateMessage(ctx, localMsg)
	if err != nil {
		return models.Message{}, nil, err
	}

	members, err := s.chatsRepo.GetChatMembers(ctx, chatID)
	if err != nil {
		return models.Message{}, nil, err
	}

	createdPublic := models.Message{
		ID:        createdLocal.ID,
		Version:   createdLocal.Version,
		ChatID:    createdLocal.ChatID,
		SenderID:  createdLocal.SenderID,
		Text:      createdLocal.Text,
		CreatedAt: createdLocal.CreatedAt,
		DeletedAt: createdLocal.DeletedAt,
	}

	return createdPublic, members, nil
}

func (s *MessagesService) GetMessages(ctx context.Context, chatID uuid.UUID, userID uuid.UUID, limit, offset *int) ([]models.Message, error) {
	ok, err := s.chatsRepo.IsMember(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("access denied")
	}

	localMsgs, err := s.repo.GetByChatID(ctx, chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	publicMsgs := make([]models.Message, len(localMsgs))
	for i, m := range localMsgs {
		publicMsgs[i] = models.Message{
			ID:        m.ID,
			Version:   m.Version,
			ChatID:    m.ChatID,
			SenderID:  m.SenderID,
			Text:      m.Text,
			CreatedAt: m.CreatedAt,
			DeletedAt: m.DeletedAt,
		}
	}

	return publicMsgs, nil
}

func (s *MessagesService) UpdateMessage(ctx context.Context, req PatchMessage, userID uuid.UUID, messageID uuid.UUID) (Message, error) {
	msg, err := s.repo.GetByID(ctx, messageID)
	if err != nil {
		return Message{}, err
	}

	if msg.SenderID != userID {
		return Message{}, fmt.Errorf("access denied")
	}

	if req.Text != nil {
		msg.Text = *req.Text
	}

	updated, err := s.repo.UpdateMessage(ctx, msg)
	if err != nil {
		return Message{}, err
	}

	return Message{
		ID:        updated.ID,
		Version:   updated.Version,
		ChatID:    updated.ChatID,
		SenderID:  updated.SenderID,
		Text:      updated.Text,
		CreatedAt: updated.CreatedAt,
		DeletedAt: updated.DeletedAt,
	}, nil
}

func (s *MessagesService) DeleteMessage(ctx context.Context, userID uuid.UUID, messageID uuid.UUID) error {

	msg, err := s.repo.GetByID(ctx, messageID)
	if err != nil {
		return err
	}

	if msg.SenderID != userID {
		return fmt.Errorf("access denied")
	}

	return s.repo.DeleteMessage(ctx, messageID)
}
