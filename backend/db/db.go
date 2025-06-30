package db

import (
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
			Colorful:                  !cfg.IsProduction,
		},
	)

	// Configure postgres with UTC timezone
	// Note: Timezone should be specified in the connection string, not as a separate field
	connStr := cfg.DatabaseUrl
	if connStr != "" && !strings.Contains(connStr, "timezone=") {
		// Only append timezone if not already specified in the connection string
		if strings.Contains(connStr, "?") {
			connStr += "&timezone=UTC"
		} else {
			connStr += "?timezone=UTC"
		}
	}

	pgConfig := postgres.Config{
		DSN:                  connStr,
		PreferSimpleProtocol: true,
	}

	// Open connection with custom config
	db, err := gorm.Open(postgres.New(pgConfig), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool settings optimized for Neon's free tier
	sqlDB.SetMaxIdleConns(5)                   // Reduce to save resources
	sqlDB.SetMaxOpenConns(20)                  // Reduce to avoid hitting connection limits
	sqlDB.SetConnMaxLifetime(10 * time.Minute) // Allow longer-lived connections

	DB = db
	return db, nil
}
