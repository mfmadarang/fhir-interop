package convert

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/hl7v2"
)

// builds one fhir.Observation per OBX segment in an HL7v2 ORU^R01 message. patientID and encounterID identify the
// already-converted Patient and Encounter these observations belong
// to; encounterID may be empty if there's no associated encounter

// Only OBX value type "NM" (numeric) maps to valueQuantity; everything
// else is stored as valueString from the raw OBX-5 text. This covers
// this project's test data, not the full HL7v2 value type table.
func ObservationsFromORU(msg *hl7v2.Message, patientID, encounterID string) ([]fhir.Observation, error) {
	if pid := msg.Segment("PID"); pid == nil {
		return nil, fmt.Errorf("convert: message has no PID segment")
	}

	var effectiveDateTime string
	if obr := msg.Segment("OBR"); obr != nil {
		if raw := obr.Field(7).String(); raw != "" {
			dt, err := hl7DateToFHIR(raw)
			if err != nil {
				return nil, fmt.Errorf("convert: OBR-7 (observation date/time): %w", err)
			}
			effectiveDateTime = dt
		}
	}
	if effectiveDateTime == "" {
		// No OBR, or OBR didn't carry a date: fall back to the message
		// timestamp so the observation isn't left with no timing at all.
		if msh := msg.Segment("MSH"); msh != nil {
			if raw := msh.Field(7).String(); raw != "" {
				if dt, err := hl7DateToFHIR(raw); err == nil {
					effectiveDateTime = dt
				}
			}
		}
	}

	obxSegments := msg.SegmentsOf("OBX")
	if len(obxSegments) == 0 {
		return nil, fmt.Errorf("convert: message has no OBX segments")
	}

	observations := make([]fhir.Observation, 0, len(obxSegments))
	for _, obx := range obxSegments {
		setID := obx.Field(1).String()
		if setID == "" {
			return nil, fmt.Errorf("convert: OBX segment missing Set ID (OBX-1)")
		}

		code := obx.Field(3)
		observation := fhir.Observation{
			ResourceType: "Observation",
			ID:           fmt.Sprintf("%s-obs-%s", patientID, setID),
			Status:       mapObservationStatus(obx.Field(11).String()),
			Code: fhir.CodeableConcept{
				Coding: []fhir.Coding{{
					System:  mapCodingSystem(code.Component(3)),
					Code:    code.Component(1),
					Display: code.Component(2),
				}},
				Text: code.Component(2),
			},
			Subject:           fhir.Reference{Reference: "Patient/" + patientID, Type: "Patient"},
			EffectiveDateTime: effectiveDateTime,
		}
		if encounterID != "" {
			observation.Encounter = fhir.Reference{Reference: "Encounter/" + encounterID, Type: "Encounter"}
		}

		valueType := obx.Field(2).String()
		rawValue := obx.Field(5).String()
		switch {
		case valueType == "NM":
			val, err := parseFloat(rawValue, fmt.Sprintf("OBX-5 (set ID %s)", setID))
			if err != nil {
				return nil, err
			}
			observation.ValueQuantity = &fhir.Quantity{Value: val, Unit: obx.Field(6).String()}
		case rawValue != "":
			v := rawValue
			observation.ValueString = &v
		}

		observations = append(observations, observation)
	}

	return observations, nil
}
