package messages

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

// CreateMessage inserts a new message into the database.
func (r *Repository) CreateMessage(message *Message) error {
	return r.db.Create(message).Error
}

// GetChannelMessages fetches paginated messages for a channel.
// We JOIN with users table to get sender info in one query.
// Messages are ordered newest first — client reverses for display.
func (r *Repository) GetChannelMessages(query GetMessagesQuery) ([]MessageResponse, int64, error) {
	var total int64

	// Count total messages in this channel for pagination metadata
	r.db.Model(&Message{}).
		Where("channel_id = ? AND deleted_at IS NULL", query.ChannelID).
		Count(&total)

	// Raw struct to scan JOIN result into
	type row struct {
		// Message fields
		ID        uuid.UUID   `gorm:"column:id"`
		Content   string      `gorm:"column:content"`
		Type      MessageType `gorm:"column:type"`
		ChannelID uuid.UUID   `gorm:"column:channel_id"`
		SenderID  uuid.UUID   `gorm:"column:sender_id"`
		CreatedAt interface{} `gorm:"column:created_at"`
		UpdatedAt interface{} `gorm:"column:updated_at"`
		EditedAt  interface{} `gorm:"column:edited_at"`

		// Sender fields (from JOIN)
		SenderUsername  string  `gorm:"column:sender_username"`
		SenderAvatarURL *string `gorm:"column:sender_avatar_url"`
	}

	var rows []row

	err := r.db.Raw(`
		SELECT 
			m.id, m.content, m.type, m.channel_id, m.sender_id,
			m.created_at, m.updated_at, m.edited_at,
			u.username AS sender_username,
			u.avatar_url AS sender_avatar_url
		FROM messages m
		JOIN users u ON u.id = m.sender_id
		WHERE m.channel_id = ? AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`, query.ChannelID, query.Limit, query.Offset).Scan(&rows).Error

	if err != nil {
		return nil, 0, err
	}

	// Map raw rows into our response struct
	result := make([]MessageResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, MessageResponse{
			Message: Message{
				ID:        r.ID,
				Content:   r.Content,
				Type:      r.Type,
				ChannelID: r.ChannelID,
				SenderID:  r.SenderID,
			},
			Sender: MessageSender{
				ID:        r.SenderID,
				Username:  r.SenderUsername,
				AvatarURL: r.SenderAvatarURL,
			},
		})
	}

	return result, total, nil
}

// FindByID fetches a single message — used for edit/delete later
func (r *Repository) FindByID(id uuid.UUID) (*Message, error) {
	var message Message
	err := r.db.Where("id = ?", id).First(&message).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}
