package validate

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

// FHIR R4's Observation.status value set.
var validObservationStatuses = map[string]bool{
	"registered":       true,
	"preliminary":      true,
	"final":            true,
	"amended":          true,
	"corrected":        true,
	"cancelled":        true,
	"entered-in-error": true,
	"unknown":          true,
}

// checks required fields, value formats, and the value[x] cardinality constraint on an Observation resource
func ValidateObservation(o fhir.Observation) []Issue {
	var issues []Issue
	add := func(field, msg string) {
		issues = append(issues, Issue{ResourceType: "Observation", ResourceID: o.ID, Field: field, Message: msg})
	}

	if o.ID == "" {
		add("id", "missing required id")
	}

	if o.Status == "" {
		add("status", "missing required status")
	} else if !validObservationStatuses[o.Status] {
		add("status", fmt.Sprintf("invalid status %q", o.Status))
	}

	if len(o.Code.Coding) == 0 && o.Code.Text == "" {
		add("code", "missing required code (no coding entries and no text)")
	}

	if o.Subject.Reference == "" {
		add("subject.reference", "missing required subject reference")
	}

	if o.EffectiveDateTime != "" && !fhirDateTimeRegex.MatchString(o.EffectiveDateTime) {
		add("effectiveDateTime", fmt.Sprintf("invalid dateTime format %q", o.EffectiveDateTime))
	}

	valueCount := 0
	if o.ValueQuantity != nil {
		valueCount++
	}
	if o.ValueString != nil {
		valueCount++
	}
	if o.ValueCodeableConcept != nil {
		valueCount++
	}
	if valueCount > 1 {
		add("value[x]", "more than one value[x] field is set; exactly one is expected")
	}

	return issues
}
