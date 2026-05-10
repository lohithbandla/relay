package users

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// Service handles all business logic for users.
// It sits between handlers and the repository.
type Service struct {
	repo *Repository
}

// NewService creates a new user service with its dependency (repository).
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Register validates input, hashes password, and creates the user.
// Returns the created user or a descriptive error.
func (s *Service) Register(req RegisterRequest) (*User, error) {
	// Business rule: no duplicate emails
	if s.repo.ExistsByEmail(req.Email) {
		return nil, errors.New("email already registered")
	}

	// Business rule: no duplicate usernames
	if s.repo.ExistsByUsername(req.Username) {
		return nil, errors.New("username already taken")
	}

	// Hash the password — cost factor 12 is the industry standard balance
	// between security and performance. Higher = slower = safer but more CPU.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, errors.New("failed to process password")
	}

	user := &User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

// VerifyPassword checks a plain password against its bcrypt hash.
// Used during login — returns nil if match, error if not.
func (s *Service) VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}

// FindByEmail exposes the repository method to the handler layer.
func (s *Service) FindByEmail(email string) (*User, error) {
	return s.repo.FindByEmail(email)
}

// FindByID exposes the repository method to the handler layer.
func (s *Service) FindByID(id string) (*User, error) {
	return s.repo.FindByID(id)
}
