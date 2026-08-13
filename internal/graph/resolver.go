package graph

import (
	"context"

	"github.com/mfmadarang/fhir-interop/internal/store"
	"gorm.io/gorm"
)

type Resolver struct {
	DB *gorm.DB
}

func (r *queryResolver) Patient(ctx context.Context, id string) (*store.PatientRecord, error) {
	return store.GetPatient(r.DB, id)
}

func (r *queryResolver) Patients(ctx context.Context, limit *int, offset *int) ([]*store.PatientRecord, error) {
	l := 20
	if limit != nil {
		l = *limit
	}
	o := 0
	if offset != nil {
		o = *offset
	}
	return store.ListPatients(r.DB, l, o)
}

func (r *queryResolver) EncountersByPatient(ctx context.Context, patientID string) ([]*store.EncounterRecord, error) {
	return store.ListEncountersByPatient(r.DB, patientID)
}

func (r *queryResolver) ObservationsByPatient(ctx context.Context, patientID string) ([]*store.ObservationRecord, error) {
	return store.ListObservationsByPatient(r.DB, patientID)
}

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
