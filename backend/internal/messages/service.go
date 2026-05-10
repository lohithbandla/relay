package messages

import (
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// SendMessage validates and persists a new message.
func (s *Service) SendMessage(req SendMessageRequest, channelID, senderID uuid.UUID) (*Message, error) {
	if req.Content == "" {
		return nil, errors.New("message content cannot be empty")
	}

	if len(req.Content) > 2000 {
		return nil, errors.New("message content exceeds 2000 characters")
	}

	// Default to text if not specified
	if req.Type == "" {
		req.Type = MessageTypeText
	}

	message := &Message{
		Content:   req.Content,
		Type:      req.Type,
		ChannelID: channelID,
		SenderID:  senderID,
	}

	if err := s.repo.CreateMessage(message); err != nil {
		return nil, errors.New("failed to send message")
	}

	return message, nil
}

// GetMessages returns paginated messages for a channel.
func (s *Service) GetMessages(channelID uuid.UUID, limit, offset int) (*PaginatedMessages, error) {
	// Enforce sane pagination limits
	if limit <= 0 || limit > 100 {
		limit = 50 // default page size
	}
	if offset < 0 {
		offset = 0
	}

	msgs, total, err := s.repo.GetChannelMessages(GetMessagesQuery{
		ChannelID: channelID,
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, errors.New("failed to fetch messages")
	}

	return &PaginatedMessages{
		Messages: msgs,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		HasMore:  int64(offset+limit) < total,
	}, nil
}
