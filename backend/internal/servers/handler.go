package servers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/lohithbandla/relay/internal/channels"
)

type Handler struct {
	service        *Service
	channelService *channels.Service
}

func NewHandler(service *Service, channelService *channels.Service) *Handler {
	return &Handler{service: service, channelService: channelService}
}

// CreateServer handles POST /api/v1/servers
func (h *Handler) CreateServer(c *fiber.Ctx) error {
	// Extract userID injected by JWT middleware
	ownerID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	var req CreateServerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Server name is required",
		})
	}

	server, err := h.service.CreateServer(req, ownerID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    server,
	})
}

// GetMyServers handles GET /api/v1/servers
func (h *Handler) GetMyServers(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	serverList, err := h.service.GetUserServers(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch servers",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    serverList,
	})
}

// JoinServer handles POST /api/v1/servers/join
func (h *Handler) JoinServer(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals("userID").(string))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid user ID",
		})
	}

	var body struct {
		InviteCode string `json:"invite_code"`
	}
	if err := c.BodyParser(&body); err != nil || body.InviteCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "invite_code is required",
		})
	}

	server, err := h.service.JoinServer(body.InviteCode, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    server,
	})
}

// CreateChannel handles POST /api/v1/servers/:serverID/channels
func (h *Handler) CreateChannel(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("serverID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid server ID",
		})
	}

	var req channels.CreateChannelRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	channel, err := h.channelService.CreateChannel(req, serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    channel,
	})
}

// GetChannels handles GET /api/v1/servers/:serverID/channels
func (h *Handler) GetChannels(c *fiber.Ctx) error {
	serverID, err := uuid.Parse(c.Params("serverID"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid server ID",
		})
	}

	channelList, err := h.channelService.GetServerChannels(serverID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch channels",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    channelList,
	})
}
