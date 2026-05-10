package websocket

import (
	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
)

// RegisterRoutes mounts the WebSocket endpoint.
// Note: WebSocket routes are NOT under the JWT middleware group
// because auth is handled via query param token inside the handler.
func RegisterRoutes(app *fiber.App, handler *Handler) {
	app.Get("/ws/:channelID",
		UpgradeMiddleware(),
		fiberws.New(handler.HandleConnection),
	)
}
