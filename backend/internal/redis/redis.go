package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/lohithbandla/relay/internal/config"
	"github.com/redis/go-redis/v9"
)

// client is the single shared Redis instance.
// Package-private — only accessible via GetClient().
var client *redis.Client

// Connect initializes the Redis connection.
// Call this ONCE at application startup.
func Connect(cfg *config.Config) error {
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword, // empty string = no auth (local dev)
		DB:       0,                 // Redis has 16 DBs (0-15), use 0 by default

		// Connection pool settings — same concept as PostgreSQL
		PoolSize:     10, // max simultaneous Redis connections
		MinIdleConns: 3,  // keep 3 connections warm at all times
	})

	// Ping Redis to verify the connection is alive
	// context.Background() = no timeout, no cancellation (fine for startup)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("[redis] Connected to Redis successfully")
	return nil
}

// GetClient returns the active Redis client.
// All services that need Redis will call this.
func GetClient() *redis.Client {
	return client
}
