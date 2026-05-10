package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/auth"
	"github.com/lohithbandla/relay/internal/config"
)

// Protected is a middleware that validates JWT tokens.
// Attach it to any route group that requires authentication.
func Protected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWT is sent in the Authorization header as: "Bearer <token>"
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization header required",
			})
		}

		// Split "Bearer <token>" into ["Bearer", "<token>"]
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authorization format must be: Bearer <token>",
			})
		}

		tokenStr := parts[1]

		// Validate the token — checks signature + expiry
		claims, err := auth.ValidateToken(tokenStr, cfg)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired token",
			})
		}

		// Store claims in context so handlers can access them
		// without re-parsing the token
		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)

		// Pass control to the next handler
		return c.Next()
	}
}
