package store

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// opens a GORM connection to Postgres using DATABASE_URL
func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connecting to postgres: %w", err)
	}
	return db, nil
}

// runs GORM AutoMigrate for all persisted resource types.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&PatientRecord{}, &EncounterRecord{}, &ObservationRecord{})
}