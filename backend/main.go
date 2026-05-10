package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lohithbandla/relay/internal/channels"
	"github.com/lohithbandla/relay/internal/config"
	"github.com/lohithbandla/relay/internal/database"
	"github.com/lohithbandla/relay/internal/messages"
	"github.com/lohithbandla/relay/internal/middleware"
	redisClient "github.com/lohithbandla/relay/internal/redis"
	"github.com/lohithbandla/relay/internal/servers"
	"github.com/lohithbandla/relay/internal/users"
	ws "github.com/lohithbandla/relay/internal/websocket"
)

func main() {
	cfg := config.Load()

	if err := database.Connect(cfg); err != nil {
		log.Fatalf("[main] Database connection failed: %v", err)
	}

	if err := redisClient.Connect(cfg); err != nil {
		log.Fatalf("[main] Redis connection failed: %v", err)
	}

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
		AppName:     "Discord Backend v1.0",
		IdleTimeout: 5 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		},
	})

	// Global middleware
	app.Use(middleware.Recover())
	app.Use(middleware.RequestLogger())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Server is healthy",
			"env":     cfg.AppEnv,
		})
	})

	// Wiring
	userRepo := users.NewRepository()
	userService := users.NewService(userRepo)
	userHandler := users.NewHandler(userService, cfg)

	channelRepo := channels.NewRepository()
	channelService := channels.NewService(channelRepo)

	serverRepo := servers.NewRepository()
	serverService := servers.NewService(serverRepo, channelRepo)
	serverHandler := servers.NewHandler(serverService, channelService)

	messageRepo := messages.NewRepository()
	messageService := messages.NewService(messageRepo)
	messageHandler := messages.NewHandler(messageService)

	hub := ws.NewHub(messageService)
	go hub.Run()

	wsHandler := ws.NewHandler(hub, cfg)

	// Routes
	api := app.Group("/api/v1")

	// Auth — strict rate limit
	auth := api.Group("/auth", middleware.StrictRateLimiter())
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", userHandler.Login)

	// Protected — JWT + rate limit
	protected := api.Group("",
		middleware.RateLimiter(),
		middleware.Protected(cfg),
	)
	protected.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success":  true,
			"userID":   c.Locals("userID"),
			"username": c.Locals("username"),
		})
	})

	servers.RegisterRoutes(protected, serverHandler)
	messages.RegisterRoutes(protected, messageHandler)
	users.RegisterRoutes(protected, userHandler)
	ws.RegisterRoutes(app, wsHandler)

	// ── Graceful Shutdown ──────────────────────────────────
	// Run server in goroutine so signal handling works
	go func() {
		log.Printf("[server] Starting on port %s in %s mode", cfg.AppPort, cfg.AppEnv)
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Fatalf("[server] Failed to start: %v", err)
		}
	}()

	// Wait for OS signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Printf("[server] Received signal: %s — shutting down gracefully", sig)

	// 10 second timeout for in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown Fiber
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("[server] Forced shutdown: %v", err)
	}

	// Close DB
	if err := database.Close(); err != nil {
		log.Printf("[server] DB close error: %v", err)
	}

	// Close Redis
	if err := redisClient.Close(); err != nil {
		log.Printf("[server] Redis close error: %v", err)
	}

	log.Println("[server] Shutdown complete ✅")
}
