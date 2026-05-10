package servers

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Server struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"not null;size:100" json:"name"`
	Description string         `gorm:"size:500" json:"description,omitempty"`
	IconURL     *string        `gorm:"size:500" json:"icon_url,omitempty"`
	OwnerID     uuid.UUID      `gorm:"type:uuid;not null" json:"owner_id"`
	InviteCode  string         `gorm:"uniqueIndex;size:20" json:"invite_code"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Only keep Members here — channels are fetched separately via their own repo
	Members []ServerMember `gorm:"foreignKey:ServerID" json:"members,omitempty"`
}

type ServerMember struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ServerID uuid.UUID `gorm:"type:uuid;not null;index" json:"server_id"`
	UserID   uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Role     string    `gorm:"size:20;default:'member'" json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

func (s *Server) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (sm *ServerMember) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == uuid.Nil {
		sm.ID = uuid.New()
	}
	sm.JoinedAt = time.Now().UTC()
	return nil
}
