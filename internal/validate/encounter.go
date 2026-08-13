package validate

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

// validEncounterStatuses is FHIR R4's Encounter.status value set.
var validEncounterStatuses = map[string]bool{
	"planned":          true,
	"arrived":          true,
	"triaged":          true,
	"in-progress":      true,
	"onleave":          true,
	"finished":         true,
	"cancelled":        true,
	"entered-in-error": true,
	"unknown":          true,
}

// checks required fields and value formats on an Encounter resource.
func ValidateEncounter(e fhir.Encounter) []Issue {
	var issues []Issue
	add := func(field, msg string) {
		issues = append(issues, Issue{ResourceType: "Encounter", ResourceID: e.ID, Field: field, Message: msg})
	}

	if e.ID == "" {
		add("id", "missing required id")
	}

	if e.Status == "" {
		add("status", "missing required status")
	} else if !validEncounterStatuses[e.Status] {
		add("status", fmt.Sprintf("invalid status %q", e.Status))
	}

	if e.Subject.Reference == "" {
		add("subject.reference", "missing required subject reference")
	}

	if e.Period.Start != "" && !fhirDateTimeRegex.MatchString(e.Period.Start) {
		add("period.start", fmt.Sprintf("invalid dateTime format %q", e.Period.Start))
	}
	if e.Period.End != "" && !fhirDateTimeRegex.MatchString(e.Period.End) {
		add("period.end", fmt.Sprintf("invalid dateTime format %q", e.Period.End))
	}

	return issues
}
