package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func RequestLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] | ${status} | ${latency} | ${ip} | ${method} ${path}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
		Output:     os.Stdout,
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
	})
}
