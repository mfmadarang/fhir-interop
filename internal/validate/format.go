package validate

import "regexp"

// matches FHIR date format YYYY, YYYY-MM, or YYYY-MM-DD
var fhirDateRegex = regexp.MustCompile(`^\d{4}(-\d{2}(-\d{2})?)?$`)

// matches FHIR dateTime values: a date, optionally followed by a time and timezone offset ("2026-01-10T-9:00:00Z")
var fhirDateTimeRegex = regexp.MustCompile(`^\d{4}(-\d{2}(-\d{2}(T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2}))?)?)?$`)