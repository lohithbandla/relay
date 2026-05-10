package users

// RegisterRequest defines the expected JSON body for POST /auth/register
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest defines the expected JSON body for POST /auth/login
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse is what we return after successful register or login.
// Notice: no password, just the token and safe user fields.
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
