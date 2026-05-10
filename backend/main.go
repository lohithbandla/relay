package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/channels"
	"github.com/lohithbandla/relay/internal/config"
	"github.com/lohithbandla/relay/internal/database"
	"github.com/lohithbandla/relay/internal/messages"
	"github.com/lohithbandla/relay/internal/middleware"
	redisClient "github.com/lohithbandla/relay/internal/redis"
	"github.com/lohithbandla/relay/internal/servers"
	"github.com/lohithbandla/relay/internal/users"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("[main] Database connection failed: %v", err)
	}

	if err := redisClient.Connect(cfg); err != nil {
		log.Fatalf("[main] Redis connection failed: %v", err)
	}

	// Single migrate call with ALL models
	if err := database.Migrate(
		&users.User{},
		&servers.Server{},
		&servers.ServerMember{},
		&channels.Channel{},
		&messages.Message{},
	); err != nil {
		log.Fatalf("[main] Migration failed: %v", err)
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

	// User wiring
	userRepo := users.NewRepository()
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService, cfg)

	// Channel wiring
	channelRepo := channels.NewRepository()
	channelService := channels.NewService(channelRepo)

	// Server wiring
	serverRepo := servers.NewRepository()
	serverService := servers.NewService(serverRepo, channelRepo)
	serverHandler := servers.NewHandler(serverService, channelService)

	// Message wiring
	messageRepo := messages.NewRepository()
	messageService := messages.NewService(messageRepo)
	messageHandler := messages.NewHandler(messageService)

	// Public routes
	api := app.Group("/api/v1")
	users.RegisterRoutes(api, userHandler)

	// Protected routes — JWT required for everything below
	protected := api.Group("", middleware.Protected(cfg))

	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success":  true,
			"userID":   c.Locals("userID"),
			"username": c.Locals("username"),
		})
	})

	servers.RegisterRoutes(protected, serverHandler)
	messages.RegisterRoutes(protected, messageHandler)

	log.Printf("[server] Starting on port %s in %s mode", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("[server] Failed to start: %v", err)
	}
}
