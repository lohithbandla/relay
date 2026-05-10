package users

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents the users table in PostgreSQL.
// Every field maps to a column — GORM infers column names from field names.
type User struct {
	// UUID is safer than auto-increment integers for public-facing IDs.
	// Auto-increment lets attackers enumerate users: /users/1, /users/2...
	// UUIDs are random and unpredictable.
	ID uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`

	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:255" json:"email"`

	// Password is NEVER returned in JSON responses — json:"-" tells
	// Go's JSON encoder to completely skip this field.
	Password string `gorm:"not null" json:"-"`

	// Avatar URL is optional — pointer means it can be NULL in DB
	AvatarURL *string `gorm:"size:500" json:"avatar_url,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// DeletedAt enables soft deletes — records are never truly deleted,
	// just marked with a timestamp. Critical for chat apps (audit trails).
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate is a GORM hook — runs automatically before every INSERT.
// We use it to generate a UUID so the app controls IDs, not the database.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
