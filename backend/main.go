package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/config"
	"github.com/lohithbandla/relay/internal/database"
	redisClient "github.com/lohithbandla/relay/internal/redis"
)

func main() {
	cfg := config.Load()

	// Initialize PostgreSQL
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("[main] Database connection failed: %v", err)
	}

	// Initialize Redis
	if err := redisClient.Connect(cfg); err != nil {
		log.Fatalf("[main] Redis connection failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName: "Discord Backend v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Server is healthy",
			"env":     cfg.AppEnv,
		})
	})

	log.Printf("[server] Starting on port %s in %s mode", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("[server] Failed to start: %v", err)
	}
}
