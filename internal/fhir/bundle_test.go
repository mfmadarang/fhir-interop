package fhir

import (
	"os"
	"testing"
)

func TestParseBundle_MinimalFixutre(t *testing.T) {
	data, err := os.ReadFile("../../testdata/sample_minimal.json")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}

	parsed, err := ParseBundle(data)
	if err != nil {
		t.Fatalf("ParseBundle returned error: %v", err)
	}

	if len(parsed.Patients) != 1 {
		t.Errorf("expected 1 patient, got %d", len(parsed.Patients))
	}
	p := parsed.Patients[0]
	if p.ID != "patient-1" {
		t.Errorf("expected patient ID %q, got %q", "patient-1", p.ID)
	}
	if len(p.Name) != 1 || p.Name[0].Family != "Dela Cruz" {
		t.Errorf("expected patient name: %+v", p.Name)
	}

	if len(parsed.Encounters) != 1 {
		t.Errorf("expected 1 encounter, got %d", len(parsed.Encounters))
	}
	if parsed.Encounters[0].Status != "finished" {
		t.Errorf("expected encounter status %q, got %q", "finished", parsed.Encounters[0].Status)
	}

	if len(parsed.Observations) != 1 {
		t.Errorf("expected 1 observation, got %d", len(parsed.Observations))
	}
	o := parsed.Observations[0]
	if o.ValueQuantity == nil {
		t.Fatalf("expected ValueQuantity to be set")
	}

	if o.ValueQuantity.Value != 72 {
		t.Errorf("expected observation value 72, got %v", o.ValueQuantity.Value)
	}

	if len(parsed.Other) != 1 || parsed.Other[0] != "Condition" {
		t.Errorf("expected Other to contain [\"Condition\"], got %+v", parsed.Other)
	}
}

func TestParseBundle_RejectsNonBundle(t *testing.T) {
	_, err := ParseBundle([]byte(`{"resourceType":"Patient"}`))
	if err == nil {
		t.Fatal("expected an error for a non-Bundle resourceType, got nil")
	}
}
