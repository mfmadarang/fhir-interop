package rest

import (
	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/validate"
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

func (g *gormStore) SaveBundle(pb *fhir.ParsedBundle) error {
	return store.SaveBundle(g.db, pb)
}

func (g *gormStore) SaveIssues(pb *fhir.ParsedBundle, source string, issues []validate.Issue) error {
	return store.SaveValidationIssues(g.db, pb, source, issues)
}
