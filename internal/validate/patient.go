package validate

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

// FHIR R4's AdministrativeGender value set
var validGenders = map[string]bool{
	"male":    true,
	"female":  true,
	"other":   true,
	"unknown": true,
}

func ValidatePatient(p fhir.Patient) []Issue {
	var issues []Issue
	add := func(field, msg string) {
		issues = append(issues, Issue{ResourceType: "Patient", ResourceID: p.ID, Field: field, Message: msg})
	}

	if p.ID == "" {
		add("id", "missing required id")
	}

	if len(p.Name) == 0 {
		add("name", "missing name")
	} else if p.Name[0].Family == "" {
		add("name[0].family", "missing family name")
	}

	if p.Gender != "" && !validGenders[p.Gender] {
		add("gender", fmt.Sprintf("invalid gender %q, expected one of male|female|other|unknown", p.Gender))
	}

	if p.BirthDate != "" && !fhirDateRegex.MatchString(p.BirthDate) {
		add("birthDate", fmt.Sprintf("invalid date format %q, expected YYYY, YYYY-MM, or YYYY-MM-DD", p.BirthDate))
	}

	return issues
}
