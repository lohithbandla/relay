package database

import (
	"fmt"
	"log"
	"time"

	"github.com/lohithbandla/relay/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the single shared database instance across the app.
// We keep it package-level here but only expose it via GetDB().
// This prevents accidental overwrites from other packages.
var db *gorm.DB

// Connect initializes the PostgreSQL connection using config values.
// Call this ONCE at application startup, before any routes are registered.
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// Configure GORM logger based on environment
	gormLogger := logger.Default.LogMode(logger.Silent)
	if cfg.AppEnv == "development" {
		// In dev, log every SQL query — helps you understand what GORM generates
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		// Skips inserting zero values — important for partial updates
		NowFunc: func() time.Time {
			return time.Now().UTC() // always store UTC in DB
		},
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying sql.DB to configure connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Connection pool settings — critical for production
	sqlDB.SetMaxOpenConns(25)                 // max simultaneous DB connections
	sqlDB.SetMaxIdleConns(10)                 // connections kept alive when idle
	sqlDB.SetConnMaxLifetime(5 * time.Minute) // recycle connections every 5 min

	log.Println("[database] Connected to PostgreSQL successfully")
	return nil
}

// GetDB returns the active GORM instance.
// All repositories will call this to get the DB handle.
func GetDB() *gorm.DB {
	return db
}

// Migrate runs auto-migration for all models.
// Call this after Connect() at startup.
// AutoMigrate creates tables, adds missing columns — never deletes columns.
func Migrate(models ...interface{}) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	log.Println("[database] Migration complete")
	return nil
}

// Close gracefully closes the database connection pool.
func Close() error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	log.Println("[database] Closing connection pool")
	return sqlDB.Close()
}
