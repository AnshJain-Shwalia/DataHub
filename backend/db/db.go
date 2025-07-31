package db

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() (*gorm.DB, error) {
	cfg := config.LoadConfig()

	// Configure custom logger with 300ms slow threshold
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             300 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		},
	)

	// Configure postgres with UTC timezone and connection timeout
	// Note: Timezone should be specified in the connection string, not as a separate field
	connStr := cfg.DatabaseUrl
	if connStr != "" {
		// Add timezone if not present
		if !strings.Contains(connStr, "timezone=") {
			if strings.Contains(connStr, "?") {
				connStr += "&timezone=UTC"
			} else {
				connStr += "?timezone=UTC"
			}
		}
		// Add connection timeout if not present
		if !strings.Contains(connStr, "connect_timeout=") {
			if strings.Contains(connStr, "?") {
				connStr += "&connect_timeout=5"
			} else {
				connStr += "?connect_timeout=5"
			}
		}
	}

	pgConfig := postgres.Config{
		DSN:                  connStr,
		PreferSimpleProtocol: true,
	}

	// Create context with 10 second timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("Attempting to open database connection...")
	// Open connection with custom config
	db, err := gorm.Open(postgres.New(pgConfig), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Database connection opened successfully")

	// Test the connection with timeout
	log.Println("Getting database instance...")
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// Ping database with timeout to ensure connection is established
	log.Println("Pinging database with 10 second timeout...")
	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database within 10 seconds: %v", err)
	}
	log.Println("Database ping successful")

	// Connection pool settings optimized for Neon's free tier
	sqlDB.SetMaxIdleConns(5)                   // Reduce to save resources
	sqlDB.SetMaxOpenConns(20)                  // Reduce to avoid hitting connection limits
	sqlDB.SetConnMaxLifetime(10 * time.Minute) // Allow longer-lived connections

	DB = db
	return db, nil
}
