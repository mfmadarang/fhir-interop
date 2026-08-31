package rest

import (
	"github.com/mfmadarang/fhir-interop/internal/store"
	"gorm.io/gorm"
)

// real patientStore, backed by the store package and a GORM DB
type gormStore struct {
	db *gorm.DB
}

func NewGormStore(db *gorm.DB) *gormStore {
	return &gormStore{db: db}
}

func (g *gormStore) GetPatient(id string) (*store.PatientRecord, error) {
	return store.GetPatient(g.db, id)
}

func (g *gormStore) SearchPatients(s store.PatientSearch) ([]*store.PatientRecord, error) {
	return store.SearchPatients(g.db, s)
}
