package store

import (
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

func TestPatientToRecord(t *testing.T) {
	p := fhir.Patient{
		ID:        "patient-1",
		Gender:    "male",
		BirthDate: "1990-05-14",
		Name:      []fhir.HumanName{{Family: "Dela Cruz", Given: []string{"Juan", "Miguel"}}},
	}

	rec, err := patientToRecord(p)
	if err != nil {
		t.Fatalf("patientToRecord returned error: %v", err)
	}
	if rec.ID != "patient-1" {
		t.Errorf("expected ID %q, got %q", "patient-1", rec.ID)
	}
	if rec.FamilyName != "Dela Cruz" {
		t.Errorf("expected FamilyName %q, got %q", "Dela Cruz", rec.FamilyName)
	}
	if rec.GivenName != "Juan" {
		t.Errorf("expected GivenName %q (first given name), got %q", "Juan", rec.GivenName)
	}
	if len(rec.Raw) == 0 {
		t.Error("expected Raw to contain marshaled JSON, got empty")
	}
}

func TestEncounterToRecord_ResolvesReferences(t *testing.T) {
	e := fhir.Encounter{
		ID:      "encounter-1",
		Status:  "finished",
		Subject: fhir.Reference{Reference: "urn:uuid:patient-1"},
		Period:  fhir.Period{Start: "2026-01-10T08:00:00Z", End: "2026-01-10T08:30:00Z"},
	}

	rec, err := encounterToRecord(e)
	if err != nil {
		t.Fatalf("encounterToRecord returned error: %v", err)
	}
	if rec.PatientID != "patient-1" {
		t.Errorf("expected PatientID %q, got %q", "patient-1", rec.PatientID)
	}
}

func TestObservationToRecord_ExtractsCodeAndReferences(t *testing.T) {
	o := fhir.Observation{
		ID:        "observation-1",
		Status:    "final",
		Subject:   fhir.Reference{Reference: "Patient/patient-1"},
		Encounter: fhir.Reference{Reference: "Encounter/encounter-1"},
		Code: fhir.CodeableConcept{
			Coding: []fhir.Coding{{System: "http://loinc.org", Code: "8867-4", Display: "Heart rate"}},
		},
	}

	rec, err := observationToRecord(o)
	if err != nil {
		t.Fatalf("observationToRecord returned error: %v", err)
	}
	if rec.PatientID != "patient-1" {
		t.Errorf("expected PatientID %q, got %q", "patient-1", rec.PatientID)
	}
	if rec.EncounterID != "encounter-1" {
		t.Errorf("expected EncounterID %q, got %q", "encounter-1", rec.EncounterID)
	}
	if rec.CodeSystem != "http://loinc.org" || rec.CodeValue != "8867-4" {
		t.Errorf("expected code loinc/8867-4, got %s/%s", rec.CodeSystem, rec.CodeValue)
	}
}

func TestReferenceID(t *testing.T) {
	cases := map[string]string{
		"urn:uuid:patient-1": "patient-1",
		"Patient/patient-1":  "patient-1",
		"patient-1":          "patient-1",
		"":                   "",
	}
	for input, want := range cases {
		if got := referenceID(input); got != want {
			t.Errorf("referenceID(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestJSON_ValueAndScanRoundTrip(t *testing.T) {
	original := JSON(`{"resourceType":"Patient","id":"patient-1"}`)

	val, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	var scanned JSON
	if err := scanned.Scan(val); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}
	if string(scanned) != string(original) {
		t.Errorf("round trip mismatch: got %q, want %q", scanned, original)
	}
}

func TestJSON_ScanNil(t *testing.T) {
	var j JSON
	if err := j.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) returned error: %v", err)
	}
	if j != nil {
		t.Errorf("expected nil JSON after scanning nil, got %v", j)
	}
}
