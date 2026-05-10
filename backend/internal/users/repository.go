package users

import (
	"github.com/lohithbandla/relay/internal/database"
	"github.com/lohithbandla/relay/internal/servers"
	"gorm.io/gorm"
)

// Repository handles all database operations for users.
// Using a struct with methods (not standalone functions) makes it
// easy to mock in tests later.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new user repository.
// We inject the DB here — this is called Dependency Injection.
func NewRepository() *Repository {
	return &Repository{
		db: database.GetDB(),
	}
}

// CreateUser inserts a new user into the database.
func (r *Repository) CreateUser(user *User) error {
	return r.db.Create(user).Error
}

// FindByEmail fetches a user by email — used during login.
func (r *Repository) FindByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID fetches a user by their UUID — used in JWT middleware later.
func (r *Repository) FindByID(id string) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if an email is already registered.
func (r *Repository) ExistsByEmail(email string) bool {
	var count int64
	r.db.Model(&User{}).Where("email = ?", email).Count(&count)
	return count > 0
}

// ExistsByUsername checks if a username is already taken.
func (r *Repository) ExistsByUsername(username string) bool {
	var count int64
	r.db.Model(&User{}).Where("username = ?", username).Count(&count)
	return count > 0
}

// GetServerMembers returns all members of a server.
func (r *Repository) GetServerMembers(serverID string) ([]servers.ServerMember, error) {
	var members []servers.ServerMember
	err := r.db.Where("server_id = ?", serverID).Find(&members).Error
	return members, err
}
