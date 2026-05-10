package servers

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/lohithbandla/relay/internal/channels"
)

type Service struct {
	repo        *Repository
	channelRepo *channels.Repository
}

func NewService(repo *Repository, channelRepo *channels.Repository) *Service {
	return &Service{repo: repo, channelRepo: channelRepo}
}

// CreateServer creates a server, auto-generates an invite code,
// adds the creator as owner, and creates a default "general" channel.
func (s *Service) CreateServer(req CreateServerRequest, ownerID uuid.UUID) (*Server, error) {
	server := &Server{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
		InviteCode:  generateInviteCode(),
	}

	if err := s.repo.CreateServer(server); err != nil {
		return nil, errors.New("failed to create server")
	}

	// Auto-add creator as owner member
	member := &ServerMember{
		ServerID: server.ID,
		UserID:   ownerID,
		Role:     "owner",
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, errors.New("failed to add owner as member")
	}

	// Every server gets a "general" channel by default — just like Discord
	defaultChannel := &channels.Channel{
		Name:     "general",
		Topic:    "General discussion",
		Type:     channels.ChannelTypeText,
		ServerID: server.ID,
		Position: 0,
	}
	if err := s.channelRepo.CreateChannel(defaultChannel); err != nil {
		return nil, errors.New("failed to create default channel")
	}

	return server, nil
}

// GetUserServers returns all servers a user belongs to.
func (s *Service) GetUserServers(userID uuid.UUID) ([]Server, error) {
	return s.repo.FindServersByUserID(userID)
}

// JoinServer adds a user to a server using an invite code.
func (s *Service) JoinServer(inviteCode string, userID uuid.UUID) (*Server, error) {
	server, err := s.repo.FindByInviteCode(inviteCode)
	if err != nil {
		return nil, errors.New("invalid invite code")
	}

	// Idempotent join — joining twice is not an error
	if s.repo.IsMember(server.ID, userID) {
		return server, nil
	}

	member := &ServerMember{
		ServerID: server.ID,
		UserID:   userID,
		Role:     "member",
	}
	if err := s.repo.AddMember(member); err != nil {
		return nil, errors.New("failed to join server")
	}

	return server, nil
}

// generateInviteCode creates a random 8-character alphanumeric code.
func generateInviteCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, 8)
	for i := range code {
		code[i] = charset[r.Intn(len(charset))]
	}
	return string(code)
}
