package servers

import (
	"github.com/google/uuid"
	"github.com/lohithbandla/relay/internal/database"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository() *Repository {
	return &Repository{db: database.GetDB()}
}

// CreateServer inserts a new server into the database.
func (r *Repository) CreateServer(server *Server) error {
	return r.db.Create(server).Error
}

// AddMember adds a user to a server's member list.
func (r *Repository) AddMember(member *ServerMember) error {
	return r.db.Create(member).Error
}

// FindByID fetches a server by its UUID.
func (r *Repository) FindByID(id uuid.UUID) (*Server, error) {
	var server Server
	err := r.db.Where("id = ?", id).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// FindByInviteCode fetches a server using its invite code.
func (r *Repository) FindByInviteCode(code string) (*Server, error) {
	var server Server
	err := r.db.Where("invite_code = ?", code).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// FindServersByUserID returns all servers a user belongs to.
func (r *Repository) FindServersByUserID(userID uuid.UUID) ([]Server, error) {
	var serverList []Server
	err := r.db.
		Joins("JOIN server_members ON server_members.server_id = servers.id").
		Where("server_members.user_id = ? AND servers.deleted_at IS NULL", userID).
		Find(&serverList).Error
	return serverList, err
}

// IsMember checks if a user is already in a server.
func (r *Repository) IsMember(serverID, userID uuid.UUID) bool {
	var count int64
	r.db.Model(&ServerMember{}).
		Where("server_id = ? AND user_id = ?", serverID, userID).
		Count(&count)
	return count > 0
}
