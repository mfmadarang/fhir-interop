// persists parsed and validated FHIR resources to Postgres via GORM
package store

import (
	"time"
)

// PatientRecord is the persisted form of a fhir.Patient
type PatientRecord struct {
	ID         string `gorm:"primaryKey"`
	FamilyName string
	GivenName  string
	Gender     string
	BirthDate  string
	Raw        JSON
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (PatientRecord) TableName() string { return "patients" }

// EncounterRecord is the persisted form of a fhir.Encounter
type EncounterRecord struct {
	ID          string `gorm:"primaryKey"`
	PatientID   string `gorm:"index"`
	Status      string
	PeriodStart string
	PeriodEnd   string
	Raw         JSON
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (EncounterRecord) TableName() string { return "encounters" }

// ObservationRecord is the persisted form of a fhir.Observation
type ObservationRecord struct {
	ID                string `gorm:"primaryKey"`
	PatientID         string `gorm:"index"`
	EncounterID       string `gorm:"index"`
	Status            string
	CodeSystem        string
	CodeValue         string
	EffectiveDateTime string
	Raw               JSON
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (ObservationRecord) TableName() string { return "observations" }