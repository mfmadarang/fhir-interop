package store

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// opens a GORM connection to Postgres using the given DSN
func Connect(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("database URL is empty")
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
