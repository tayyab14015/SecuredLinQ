package database

import (
	"fmt"
	"log"
	"time"

	"github.com/securedlinq/backend/internal/config"
	"github.com/securedlinq/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the database
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")
	return db, nil
}

// AutoMigrate runs auto migration for all models
func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	migrator := db.Migrator()

	// Migrate tables one by one to handle errors gracefully
	// Only these tables are used in the system:
	// - drivers, loads, sessions, meeting_rooms, gallery
	tables := []interface{}{
		&models.Driver{},
		&models.Load{},
		&models.Session{},
		&models.MeetingRoom{},
		&models.Gallery{},
	}

	for _, table := range tables {
		// Check if table exists
		if !migrator.HasTable(table) {
			log.Printf("Creating table for %T", table)
			if err := migrator.AutoMigrate(table); err != nil {
				log.Printf("Error creating table %T: %v", table, err)
				return err
			}
		} else {
			log.Printf("Table for %T already exists, attempting to sync schema", table)
			// For existing tables, try to migrate but don't fail on column modification errors
			if err := migrator.AutoMigrate(table); err != nil {
				log.Printf("Warning: Migration issue for existing table %T: %v (continuing...)", table, err)
			}
		}
	}

	return nil
}
