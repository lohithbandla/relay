package messages

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	channels := router.Group("/channels")

	channels.Post("/:channelID/messages", handler.SendMessage)
	channels.Get("/:channelID/messages", handler.GetMessages)
}
