package servers

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler) {
	s := router.Group("/servers")

	s.Post("/", handler.CreateServer)
	s.Get("/", handler.GetMyServers)
	s.Post("/join", handler.JoinServer)
	s.Post("/:serverID/channels", handler.CreateChannel)
	s.Get("/:serverID/channels", handler.GetChannels)
}
