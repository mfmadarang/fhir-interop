package convert

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/hl7v2"
)

// derives Encounter ID from PV1-19, falling back to MSH-10
func encounterID(pv1, msh hl7v2.Segment) (string, error) {
	if id := pv1.Field(19).String(); id != "" {
		return id, nil
	}
	controlID := msh.Field(10).String()
	if controlID == "" {
		return "", fmt.Errorf("convert: cannot determine encounter ID: PV1-19 and MSH-10 both empty")
	}
	return "encounter-" + controlID, nil
}

// builds a fhir.Encounter from an ADT message's PV1 segment
// ADT only. see EncounterIDFromPV1 for ORU
func EncounterFromADT(msg *hl7v2.Message, patientID string) (fhir.Encounter, error) {
	pv1 := msg.Segment("PV1")
	if pv1 == nil {
		return fhir.Encounter{}, fmt.Errorf("convert: message has no PV1 segment")
	}
	msh := msg.Segment("MSH")
	if msh == nil {
		return fhir.Encounter{}, fmt.Errorf("convert: message has no MSH segment")
	}

	status, err := mapADTStatus(msh.Field(9).Component(2))
	if err != nil {
		return fhir.Encounter{}, err
	}

	id, err := encounterID(*pv1, *msh)
	if err != nil {
		return fhir.Encounter{}, err
	}

	encounter := fhir.Encounter{
		ResourceType: "Encounter",
		ID:           id,
		Status:       status,
		Class:        mapPatientClass(pv1.Field(2).String()),
		Subject:      fhir.Reference{Reference: "Patient/" + patientID, Type: "Patient"},
	}

	if admit := pv1.Field(44).String(); admit != "" {
		start, err := hl7DateToFHIR(admit)
		if err != nil {
			return fhir.Encounter{}, fmt.Errorf("convert: PV1-44 (admit date/time): %w", err)
		}
		encounter.Period.Start = start
	}
	if discharge := pv1.Field(45).String(); discharge != "" {
		end, err := hl7DateToFHIR(discharge)
		if err != nil {
			return fhir.Encounter{}, fmt.Errorf("convert: PV1-45 (discharge date/time): %w", err)
		}
		encounter.Period.End = end
	}

	return encounter, nil
}

// derives Encounter ID from PV1 without persisting an Encounter
// used by ORU: avoids overwriting ADT-set status/period on upsert
func EncounterIDFromPV1(msg *hl7v2.Message) (string, error) {
	pv1 := msg.Segment("PV1")
	if pv1 == nil {
		return "", fmt.Errorf("convert: message has no PV1 segment")
	}
	msh := msg.Segment("MSH")
	if msh == nil {
		return "", fmt.Errorf("convert: message has no MSH segment")
	}
	return encounterID(*pv1, *msh)
}
