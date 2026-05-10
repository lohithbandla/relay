package channels

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

// CreateChannel creates a new channel inside a server.
func (s *Service) CreateChannel(req CreateChannelRequest, serverID uuid.UUID) (*Channel, error) {
	if req.Name == "" {
		return nil, errors.New("channel name is required")
	}

	// Default to text channel if not specified
	if req.Type == "" {
		req.Type = ChannelTypeText
	}

	channel := &Channel{
		Name:      req.Name,
		Topic:     req.Topic,
		Type:      req.Type,
		ServerID:  serverID,
		IsPrivate: req.IsPrivate,
	}

	if err := s.repo.CreateChannel(channel); err != nil {
		return nil, errors.New("failed to create channel")
	}

	return channel, nil
}

// GetServerChannels returns all channels for a server.
func (s *Service) GetServerChannels(serverID uuid.UUID) ([]Channel, error) {
	return s.repo.FindByServerID(serverID)
}
