package channels

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

// CreateChannel inserts a new channel into the database.
func (r *Repository) CreateChannel(channel *Channel) error {
	return r.db.Create(channel).Error
}

// FindByServerID returns all channels belonging to a server.
func (r *Repository) FindByServerID(serverID uuid.UUID) ([]Channel, error) {
	var channelList []Channel
	err := r.db.
		Where("server_id = ? AND deleted_at IS NULL", serverID).
		Order("position ASC").
		Find(&channelList).Error
	return channelList, err
}

// FindByID fetches a single channel by UUID.
func (r *Repository) FindByID(id uuid.UUID) (*Channel, error) {
	var channel Channel
	err := r.db.Where("id = ?", id).First(&channel).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}
