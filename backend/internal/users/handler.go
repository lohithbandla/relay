package users

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/auth"
	"github.com/lohithbandla/relay/internal/config"
)

// Handler holds dependencies needed to handle HTTP requests.
type Handler struct {
	service *Service
	cfg     *config.Config
}

// NewHandler creates a new user handler.
func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse JSON body into our DTO struct
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Basic validation — check required fields manually for now
	// (We'll add a proper validator library in a later step)
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Username, email and password are required",
		})
	}

	if len(req.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Password must be at least 6 characters",
		})
	}

	// Delegate to service — handler doesn't know HOW this works
	user, err := h.service.Register(req)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	// Generate JWT for immediate login after registration
	token, err := auth.GenerateToken(user.ID.String(), user.Username, h.cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data": AuthResponse{
			Token: token,
			User:  *user,
		},
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email and password are required",
		})
	}

	// Find user by email
	user, err := h.service.FindByEmail(req.Email)
	if err != nil {
		// Don't say "user not found" — that leaks information about
		// which emails are registered. Always say "invalid credentials".
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	// Verify password against bcrypt hash
	if err := h.service.VerifyPassword(user.Password, req.Password); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid credentials",
		})
	}

	// Generate JWT
	token, err := auth.GenerateToken(user.ID.String(), user.Username, h.cfg)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data": AuthResponse{
			Token: token,
			User:  *user,
		},
	})
}
