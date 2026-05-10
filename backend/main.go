package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/config"
	"github.com/lohithbandla/relay/internal/database"
	"github.com/lohithbandla/relay/internal/middleware"
	redisClient "github.com/lohithbandla/relay/internal/redis"
	"github.com/lohithbandla/relay/internal/users"
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

	// Run migrations — pass all models here as you add them
	if err := database.Migrate(&users.User{}); err != nil {
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

	// Wire up dependencies manually — this is called Pure DI (no DI framework)
	// Order matters: repo → service → handler → routes
	userRepo := users.NewRepository()
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService, cfg)

	// All API routes live under /api/v1
	api := app.Group("/api/v1")
	users.RegisterRoutes(api, userHandler)

	// Protected route group — all routes here require a valid JWT
	protected := api.Group("", middleware.Protected(cfg))
	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success":  true,
			"userID":   c.Locals("userID"),
			"username": c.Locals("username"),
		})
	})

	log.Printf("[server] Starting on port %s in %s mode", cfg.AppPort, cfg.AppEnv)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatalf("[server] Failed to start: %v", err)
	}
}
