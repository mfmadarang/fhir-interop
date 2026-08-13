// maps parsed HL7v2 messages onto this project's
// internal/fhir resource structs. It only implements the mappings
// needed for ADT^A01 (admit) and ORU^R01 (results) messages; other
// trigger events are not yet supported
package convert

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

// converts an HL7v2 TS (timestamp) value into a FHIR date or dateTime string
// Only date and, if present, hour/minute/second are handled; any
// timezone offset in the source value is dropped rather than parsed.
func hl7DateToFHIR(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if len(raw) < 8 {
		return "", fmt.Errorf("convert: invalid HL7 date/time %q: too short", raw)
	}

	datePart := raw[0:8]
	if _, err := time.Parse("20060102", datePart); err != nil {
		return "", fmt.Errorf("convert: invalid HL7 date %q: %w", datePart, err)
	}
	date := fmt.Sprintf("%s-%s-%s", datePart[0:4], datePart[4:6], datePart[6:8])

	if len(raw) < 14 {
		return date, nil
	}

	timePart := raw[8:14]
	if _, err := time.Parse("150405", timePart); err != nil {
		// Malformed time component; fall back to date-only rather
		// than failing the whole conversion over it.
		return date, nil
	}
	clock := fmt.Sprintf("%s:%s:%s", timePart[0:2], timePart[2:4], timePart[4:6])

	offset := "Z"
	if rest := raw[14:]; len(rest) >= 5 && (rest[0] == '+' || rest[0] == '-') {
		sign := string(rest[0])
		hhmm := rest[1:5]
		if _, err := time.Parse("1504", hhmm); err == nil {
			offset = sign + hhmm[0:2] + ":" + hhmm[2:4]
		}
	}

	return fmt.Sprintf("%sT%s%s", date, clock, offset), nil
}

// maps an HL7v2 table 0001 (Administrative Sex) code from PID-8 to a FHIR Patient.gender code
func mapGender(code string) string {
	switch code {
	case "M":
		return "male"
	case "F":
		return "female"
	case "O", "A":
		// "Ambiguous" has no FHIR equivalent; "other" is the closest fit.
		return "other"
	default:
		// Includes "N" (not applicable) and anything unrecognized.
		return "unknown"
	}
}

// maps an HL7v2 table 0004 (Patient Class) code from PV1-2 to a FHIR encounter-class Coding, using the FHIR ActEncounterCode
func mapPatientClass(code string) fhir.Coding {
	const actCodeSystem = "http://terminology.hl7.org/CodeSystem/v3-ActCode"
	switch code {
	case "I":
		return fhir.Coding{System: actCodeSystem, Code: "IMP", Display: "inpatient encounter"}
	case "O":
		return fhir.Coding{System: actCodeSystem, Code: "AMB", Display: "ambulatory"}
	case "E":
		return fhir.Coding{System: actCodeSystem, Code: "EMER", Display: "emergency"}
	default:
		return fhir.Coding{Code: code}
	}
}

// maps an HL7v2 coding system abbreviation (component 3 of a CE field, e.g. OBX-3) to the URI FHIR expects. Only LOINC ("LN")
// is mapped, since it's the only system this project's test data uses.
func mapCodingSystem(abbrev string) string {
	if abbrev == "LN" {
		return "http://loinc.org"
	}
	return ""
}

// maps an HL7v2 table 0085 (Observation Result Status) code from OBX-11 to a FHIR Observation.status code. An empty
// input defaults to "final"
func mapObservationStatus(code string) string {
	switch code {
	case "", "F":
		return "final"
	case "P":
		return "preliminary"
	case "C":
		return "corrected"
	case "X":
		return "cancelled"
	default:
		return "unknown"
	}
}

func parseFloat(raw, context string) (float64, error) {
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("convert: %s: invalid numeric value %q: %w", context, raw, err)
	}
	return v, nil
}

func mapADTStatus(triggerEvent string) (string, error) {
	switch triggerEvent {
	case "A01":
		return "in-progress", nil
	case "A03":
		return "finished", nil
	default:
		return "", fmt.Errorf("convert: unsupported ADT trigger event %q", triggerEvent)
	}
}
