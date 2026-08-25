package validate

import (
	"context"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

func TestValidateBundleTerminology_MultipleObservations(t *testing.T) {
	fv := &fakeValidator{
		results: map[string]*terminology.Result{
			terminology.SystemLOINC + "|1": {Valid: false, Message: "bad"},
			terminology.SystemLOINC + "|2": {Valid: true},
			terminology.SystemLOINC + "|3": {Valid: false, Message: "bad"},
		},
	}
	pb := &fhir.ParsedBundle{
		Observations: []fhir.Observation{
			{ID: "o1", Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "1"}}}},
			{ID: "o2", Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "2"}}}},
			{ID: "o3", Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "3"}}}},
		},
	}
	issues := ValidateBundleTerminology(context.Background(), fv, pb)
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateBundleTerminology_Empty(t *testing.T) {
	fv := &fakeValidator{}
	pb := &fhir.ParsedBundle{}
	issues := ValidateBundleTerminology(context.Background(), fv, pb)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}
