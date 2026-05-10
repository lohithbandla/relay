package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lohithbandla/relay/internal/config"
)

// Claims defines what we embed inside the JWT token.
// Keep it minimal — tokens are sent with every request.
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	// RegisteredClaims handles expiry, issued-at, etc. automatically
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for a user.
// Called after successful register or login.
func GenerateToken(userID, username string, cfg *config.Config) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// Token expires based on config — default 72 hours
			ExpiresAt: jwt.NewNumericDate(
				time.Now().UTC().Add(time.Duration(cfg.JWTExpiryHours) * time.Hour),
			),
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		},
	}

	// HS256 = HMAC-SHA256 — symmetric signing using our secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return signed, nil
}

// ValidateToken parses and validates a JWT string.
// Returns the claims if valid, error if expired or tampered.
func ValidateToken(tokenStr string, cfg *config.Config) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		// This function is called by the parser to get the signing key.
		// We verify the algorithm first to prevent algorithm confusion attacks.
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(cfg.JWTSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
