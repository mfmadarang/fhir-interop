package graph

import (
	"context"
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/store"
	"gorm.io/gorm"
)

type Resolver struct {
	DB *gorm.DB
}

func (r *queryResolver) Patient(ctx context.Context, id string) (*store.PatientRecord, error) {
	return store.GetPatient(r.DB, id)
}

func (r *queryResolver) Patients(ctx context.Context, first *int, after *string) (*PatientConnection, error) {
	f := 20
	if first != nil {
		f = *first
	}

	afterID := ""
	if after != nil {
		id, err := decodeCursor(*after)
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %w", err)
		}
		afterID = id
	}

	recs, hasNextPage, err := store.ListPatientsCursor(r.DB, f, afterID)
	if err != nil {
		return nil, err
	}

	edges := make([]*PatientEdge, len(recs))
	for i, rec := range recs {
		edges[i] = &PatientEdge{Node: rec, Cursor: encodeCursor(rec.ID)}
	}

	var endCursor *string
	if len(edges) > 0 {
		c := edges[len(edges)-1].Cursor
		endCursor = &c
	}

	return &PatientConnection{
		Edges: edges,
		PageInfo: &PageInfo{
			HasNextPage: hasNextPage,
			EndCursor:   endCursor,
		},
	}, nil
}

func (r *queryResolver) EncountersByPatient(ctx context.Context, patientID string) ([]*store.EncounterRecord, error) {
	return store.ListEncountersByPatient(r.DB, patientID)
}

func (r *queryResolver) ObservationsByPatient(ctx context.Context, patientID string) ([]*store.ObservationRecord, error) {
	return store.ListObservationsByPatient(r.DB, patientID)
}

func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
