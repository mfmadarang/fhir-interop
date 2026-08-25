package validate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

type fakeValidator struct {
	results map[string]*terminology.Result
	errs    map[string]error
}

func (f *fakeValidator) ValidateCode(ctx context.Context, system, code string) (*terminology.Result, error) {
	key := system + "|" + code
	if err, ok := f.errs[key]; ok {
		return nil, err
	}
	if r, ok := f.results[key]; ok {
		return r, nil
	}
	return &terminology.Result{Valid: true}, nil
}

func TestValidateObservationTerminology_Valid(t *testing.T) {
	fv := &fakeValidator{
		results: map[string]*terminology.Result{
			terminology.SystemLOINC + "|8867-4": {Valid: true, Display: "Heart rate"},
		},
	}
	o := fhir.Observation{
		ID:   "obs-1",
		Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "8867-4"}}},
	}
	issues := ValidateObservationTerminology(context.Background(), fv, o)
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %v", issues)
	}
}

func TestValidateObservationTerminology_Invalid(t *testing.T) {
	fv := &fakeValidator{
		results: map[string]*terminology.Result{
			terminology.SystemLOINC + "|bogus": {Valid: false, Message: "not a valid code"},
		},
	}
	o := fhir.Observation{
		ID:   "obs-2",
		Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "bogus"}}},
	}
	issues := ValidateObservationTerminology(context.Background(), fv, o)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Field != "code" {
		t.Errorf("expected field %q, got %q", "code", issues[0].Field)
	}
}

func TestValidateObservationTerminology_SkipsOtherSystems(t *testing.T) {
	fv := &fakeValidator{}
	o := fhir.Observation{
		ID: "obs-3",
		Code: fhir.CodeableConcept{Coding: []fhir.Coding{
			{System: "http://terminology.hl7.org/CodeSystem/observation-category", Code: "vital-signs"},
		}},
	}
	issues := ValidateObservationTerminology(context.Background(), fv, o)
	if len(issues) != 0 {
		t.Errorf("expected no issues for non-LOINC/SNOMED system, got %v", issues)
	}
}

func TestValidateObservationTerminology_ServerError(t *testing.T) {
	fv := &fakeValidator{
		errs: map[string]error{
			terminology.SystemLOINC + "|8867-4": errors.New("timeout"),
		},
	}
	o := fhir.Observation{
		ID:   "obs-4",
		Code: fhir.CodeableConcept{Coding: []fhir.Coding{{System: terminology.SystemLOINC, Code: "8867-4"}}},
	}
	issues := ValidateObservationTerminology(context.Background(), fv, o)
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Message, "unverified") {
		t.Errorf("expected message to mention unverified, got %q", issues[0].Message)
	}
}
