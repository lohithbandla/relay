package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string

	JWTSecret      string
	JWTExpiryHours int
}

// Load reads the .env file and returns a validated Config struct.
// Call this ONCE at application startup.
func Load() *Config {
	// In production, .env won't exist — real env vars are injected.
	// godotenv.Load() silently skips if file is missing, which is correct behavior.
	if err := godotenv.Load(); err != nil {
		log.Println("[config] No .env file found, reading from environment")
	}

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "72"))
	if err != nil {
		log.Fatal("[config] JWT_EXPIRY_HOURS must be a valid integer")
	}

	return &Config{
		AppPort:        getEnv("APP_PORT", "7777"),
		AppEnv:         getEnv("APP_ENV", "development"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "chatuser"),
		DBPassword:     getEnv("DB_PASSWORD", "chatpassword"),
		DBName:         getEnv("DB_NAME", "chatdb"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		RedisHost:      getEnv("REDIS_HOST", "localhost"),
		RedisPort:      getEnv("REDIS_PORT", "6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: jwtExpiry,
	}
}

// getEnv returns the env variable value or a fallback default.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
