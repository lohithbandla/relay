package messages

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageType allows future expansion — text, image, system messages
type MessageType string

const (
	MessageTypeText   MessageType = "text"
	MessageTypeImage  MessageType = "image"
	MessageTypeSystem MessageType = "system"
)

// Message represents a single chat message inside a channel.
type Message struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	// Content is the actual text of the message
	Content string `gorm:"type:text;not null" json:"content"`

	Type MessageType `gorm:"type:varchar(20);default:'text'" json:"type"`

	// ChannelID links this message to a specific channel
	ChannelID uuid.UUID `gorm:"type:uuid;not null;index" json:"channel_id"`

	// SenderID links to the user who sent this message
	SenderID uuid.UUID `gorm:"type:uuid;not null;index" json:"sender_id"`

	// EditedAt is nil until the message is edited
	EditedAt *time.Time `json:"edited_at,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Sender is populated via JOIN — not a DB column
	// This lets us return username without a second API call
	Sender *MessageSender `gorm:"-" json:"sender,omitempty"`
}

// MessageSender is a lightweight struct for sender info.
// We don't embed the full User model to avoid exposing sensitive fields.
type MessageSender struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
