package rest

import (
	"github.com/mfmadarang/fhir-interop/internal/store"
	"gorm.io/gorm"
)

// gormStore is the real patientStore, backed by the store package + a GORM DB.
type gormStore struct {
	db *gorm.DB
}

// NewGormStore returns a patientStore that reads from Postgres via GORM.
func NewGormStore(db *gorm.DB) *gormStore {
	return &gormStore{db: db}
}

func (g *gormStore) GetPatient(id string) (*store.PatientRecord, error) {
	return store.GetPatient(g.db, id)
}

func (g *gormStore) SearchPatients(s store.PatientSearch) ([]*store.PatientRecord, error) {
	return store.SearchPatients(g.db, s)
}
