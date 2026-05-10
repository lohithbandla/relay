package websocket

import (
	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/lohithbandla/relay/internal/auth"
	"github.com/lohithbandla/relay/internal/config"
)

// Handler manages WebSocket upgrade and connection setup.
type Handler struct {
	hub *Hub
	cfg *config.Config
}

// NewHandler creates a new WebSocket handler.
func NewHandler(hub *Hub, cfg *config.Config) *Handler {
	return &Handler{hub: hub, cfg: cfg}
}

// HandleConnection upgrades HTTP to WebSocket and starts the client pumps.
// URL: /api/v1/ws/:channelID?token=<jwt>
func (h *Handler) HandleConnection(c *fiberws.Conn) {
	// Extract channelID from URL params
	channelIDStr := c.Params("channelID")
	channelID, err := uuid.Parse(channelIDStr)
	if err != nil {
		c.Close()
		return
	}

	// WebSocket connections can't send custom headers easily from browsers
	// So we accept the JWT as a query parameter for WebSocket auth
	token := c.Query("token")
	if token == "" {
		c.Close()
		return
	}

	// Validate JWT
	claims, err := auth.ValidateToken(token, h.cfg)
	if err != nil {
		c.Close()
		return
	}

	// Create client and register with hub
	client := NewClient(c, claims.UserID, claims.Username, channelID, h.hub)
	h.hub.register <- client

	// Start write pump in a goroutine — handles outbound messages
	// We need it in a goroutine because WritePump blocks on its for/select loop
	go client.WritePump()

	// ReadPump runs in the current goroutine — blocks until client disconnects
	// When ReadPump returns, the defer in it unregisters the client
	client.ReadPump()
}

// UpgradeMiddleware checks if the request is a valid WebSocket upgrade.
// Fiber requires this check before the actual WebSocket handler.
func UpgradeMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
