package validate

import (
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

func TestValidatePatient(t *testing.T) {
	valid := fhir.Patient{
		ID:        "patient-1",
		Name:      []fhir.HumanName{{Family: "Dela Cruz", Given: []string{"Juan"}}},
		Gender:    "male",
		BirthDate: "1990-05-14",
	}
	if issues := ValidatePatient(valid); len(issues) != 0 {
		t.Errorf("expected no issues for valid patient, got %v", issues)
	}

	invalid := fhir.Patient{
		Gender:    "man",
		BirthDate: "14-05-1990",
	}
	issues := ValidatePatient(invalid)
	wantFields := map[string]bool{"id": false, "name": false, "gender": false, "birthDate": false}
	for _, iss := range issues {
		wantFields[iss.Field] = true
	}
	for field, found := range wantFields {
		if !found {
			t.Errorf("expected an issue on field %q, got %v", field, issues)
		}
	}
}

func TestValidateEncounter(t *testing.T) {
	valid := fhir.Encounter{
		ID:      "encounter-1",
		Status:  "finished",
		Subject: fhir.Reference{Reference: "urn:uuid:patient-1"},
		Period:  fhir.Period{Start: "2026-01-10T08:00:00Z", End: "2026-01-10T08:30:00Z"},
	}
	if issues := ValidateEncounter(valid); len(issues) != 0 {
		t.Errorf("expected no issues for valid encounter, got %v", issues)
	}

	invalid := fhir.Encounter{Status: "in-transit"}
	issues := ValidateEncounter(invalid)
	wantFields := map[string]bool{"id": false, "status": false, "subject.reference": false}
	for _, iss := range issues {
		wantFields[iss.Field] = true
	}
	for field, found := range wantFields {
		if !found {
			t.Errorf("expected an issue on field %q, got %v", field, issues)
		}
	}
}

func TestValidateObservation(t *testing.T) {
	valid := fhir.Observation{
		ID:                "observation-1",
		Status:            "final",
		Code:              fhir.CodeableConcept{Coding: []fhir.Coding{{System: "http://loinc.org", Code: "8867-4"}}},
		Subject:           fhir.Reference{Reference: "urn:uuid:patient-1"},
		EffectiveDateTime: "2026-01-10T08:05:00Z",
		ValueQuantity:     &fhir.Quantity{Value: 72, Unit: "beats/minute"},
	}
	if issues := ValidateObservation(valid); len(issues) != 0 {
		t.Errorf("expected no issues for valid observation, got %v", issues)
	}

	valueStr := "positive"
	multiValue := valid
	multiValue.ValueString = &valueStr // now both ValueQuantity and ValueString are set
	issues := ValidateObservation(multiValue)
	found := false
	for _, iss := range issues {
		if iss.Field == "value[x]" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a value[x] cardinality issue, got %v", issues)
	}

	missing := fhir.Observation{Status: "final"}
	issues = ValidateObservation(missing)
	wantFields := map[string]bool{"id": false, "code": false, "subject.reference": false}
	for _, iss := range issues {
		wantFields[iss.Field] = true
	}
	for field, foundField := range wantFields {
		if !foundField {
			t.Errorf("expected an issue on field %q, got %v", field, issues)
		}
	}
}

func TestValidateBundle_AggregatesAcrossResourceTypes(t *testing.T) {
	pb := &fhir.ParsedBundle{
		Patients:     []fhir.Patient{{}},     // missing id, name
		Encounters:   []fhir.Encounter{{}},   // missing id, status, subject
		Observations: []fhir.Observation{{}}, // missing id, status, code, subject
	}

	issues := ValidateBundle(pb)
	if len(issues) == 0 {
		t.Fatal("expected issues from all three resource types, got none")
	}

	seen := map[string]bool{}
	for _, iss := range issues {
		seen[iss.ResourceType] = true
	}
	for _, rt := range []string{"Patient", "Encounter", "Observation"} {
		if !seen[rt] {
			t.Errorf("expected at least one issue for resourceType %q", rt)
		}
	}
}