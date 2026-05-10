package users

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	redispkg "github.com/lohithbandla/relay/internal/redis"
)

// GetPresence handles GET /api/v1/servers/:serverID/presence
// Returns which members of a server are currently online.
func (h *Handler) GetPresence(c *fiber.Ctx) error {
	serverID := c.Params("serverID")
	if serverID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Server ID required",
		})
	}

	// Parse serverID
	_, err := uuid.Parse(serverID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid server ID",
		})
	}

	// Get all members of this server
	members, err := h.service.GetServerMembers(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch members",
		})
	}

	// Extract user IDs
	userIDs := make([]string, len(members))
	for i, m := range members {
		userIDs[i] = m.UserID.String()
	}

	// Check which ones are online in Redis
	onlineIDs, err := redispkg.GetOnlineUsers(context.Background(), userIDs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch presence",
		})
	}

	// Build response map
	presence := make(map[string]bool)
	for _, id := range userIDs {
		presence[id] = false
	}
	for _, id := range onlineIDs {
		presence[id] = true
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"online_count": len(onlineIDs),
			"presence":     presence,
		},
	})
}
