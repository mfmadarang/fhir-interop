package validate

import "github.com/mfmadarang/fhir-interop/internal/fhir"

// runs validators over every Patient, Encounter, and Observation in a ParsedBundle and returns all issues found, across all resources
func ValidateBundle(pb *fhir.ParsedBundle) []Issue {
	var issues []Issue

	for _, p := range pb.Patients {
		issues = append(issues, ValidatePatient(p)...)
	}
	for _, e := range pb.Encounters {
		issues = append(issues, ValidateEncounter(e)...)
	}
	for _, o := range pb.Observations {
		issues = append(issues, ValidateObservation(o)...)
	}

	return issues
}
