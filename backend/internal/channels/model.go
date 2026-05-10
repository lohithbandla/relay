package channels

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChannelType distinguishes between text and voice channels.
// Using a string constant prevents magic strings scattered in code.
type ChannelType string

const (
	ChannelTypeText  ChannelType = "text"
	ChannelTypeVoice ChannelType = "voice"
)

// Channel belongs to a Server and holds messages.
type Channel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Name e.g. "general", "announcements"
	Name string `gorm:"not null;size:100" json:"name"`

	Topic string `gorm:"size:500" json:"topic,omitempty"`

	// Type determines behavior — text channels get messages,
	// voice channels get WebRTC connections (later phase)
	Type ChannelType `gorm:"type:varchar(20);default:'text'" json:"type"`

	// ServerID is the foreign key linking back to the parent server
	ServerID uuid.UUID `gorm:"type:uuid;not null;index" json:"server_id"`

	// Position controls display order in the sidebar (like Discord)
	Position int `gorm:"default:0" json:"position"`

	// IsPrivate channels are only visible to specific roles
	IsPrivate bool `gorm:"default:false" json:"is_private"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (c *Channel) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
