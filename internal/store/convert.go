package store

import (
	"fmt"
	"strings"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

func patientToRecord(p fhir.Patient) (PatientRecord, error) {
	raw, err := toJSON(p)
	if err != nil {
		return PatientRecord{}, fmt.Errorf("marshaling patient %s: %w", p.ID, err)
	}

	rec := PatientRecord{
		ID:        p.ID,
		Gender:    p.Gender,
		BirthDate: p.BirthDate,
		Raw:       raw,
	}
	if len(p.Name) > 0 {
		rec.FamilyName = p.Name[0].Family
		if len(p.Name[0].Given) > 0 {
			rec.GivenName = p.Name[0].Given[0]
		}
	}
	return rec, nil
}

func encounterToRecord(e fhir.Encounter) (EncounterRecord, error) {
	raw, err := toJSON(e)
	if err != nil {
		return EncounterRecord{}, fmt.Errorf("marshaling encounter %s: %w", e.ID, err)
	}
	return EncounterRecord{
		ID:          e.ID,
		PatientID:   referenceID(e.Subject.Reference),
		Status:      e.Status,
		PeriodStart: e.Period.Start,
		PeriodEnd:   e.Period.End,
		Raw:         raw,
	}, nil
}

func observationToRecord(o fhir.Observation) (ObservationRecord, error) {
	raw, err := toJSON(o)
	if err != nil {
		return ObservationRecord{}, fmt.Errorf("marshaling observation %s: %w", o.ID, err)
	}
	rec := ObservationRecord{
		ID:                o.ID,
		PatientID:         referenceID(o.Subject.Reference),
		EncounterID:       referenceID(o.Encounter.Reference),
		Status:            o.Status,
		EffectiveDateTime: o.EffectiveDateTime,
		Raw:               raw,
	}
	if len(o.Code.Coding) > 0 {
		rec.CodeSystem = o.Code.Coding[0].System
		rec.CodeValue = o.Code.Coding[0].Code
	}
	return rec, nil
}

// so it can be stored as a plain foreign key column
func referenceID(ref string) string {
	if ref == "" {
		return ""
	}
	if idx := strings.LastIndex(ref, "/"); idx != -1 {
		return ref[idx+1:]
	}
	if idx := strings.LastIndex(ref, ":"); idx != -1 {
		return ref[idx+1:]
	}
	return ref
}