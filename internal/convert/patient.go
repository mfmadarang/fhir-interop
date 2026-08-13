package convert

import (
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/hl7v2"
)

// builds a fhir.Patient from the PID segment of an HL7v2 ADT message.
// Patient.ID is set to PID-3 component 1 (the MRN in this project's test data)
func PatientFromADT(msg *hl7v2.Message) (fhir.Patient, error) {
	pid := msg.Segment("PID")
	if pid == nil {
		return fhir.Patient{}, fmt.Errorf("convert: message has no PID segment")
	}

	mrn := pid.Field(3).Component(1)
	if mrn == "" {
		return fhir.Patient{}, fmt.Errorf("convert: PID-3 (patient identifier) is empty")
	}

	patient := fhir.Patient{
		ResourceType: "Patient",
		ID:           mrn,
		Gender:       mapGender(pid.Field(8).String()),
	}

	identifier := fhir.Identifier{Value: mrn, Use: "usual"}
	if system := pid.Field(3).Component(4); system != "" {
		identifier.System = system
	}
	patient.Identifier = []fhir.Identifier{identifier}

	family := pid.Field(5).Component(1)
	given := pid.Field(5).Component(2)
	if family != "" || given != "" {
		name := fhir.HumanName{Family: family}
		if given != "" {
			name.Given = []string{given}
		}
		if middle := pid.Field(5).Component(3); middle != "" {
			name.Given = append(name.Given, middle)
		}
		patient.Name = []fhir.HumanName{name}
	}

	if dob := pid.Field(7).String(); dob != "" {
		birthDate, err := hl7DateToFHIR(dob)
		if err != nil {
			return fhir.Patient{}, fmt.Errorf("convert: PID-7 (birth date): %w", err)
		}
		if len(birthDate) > 10 {
			birthDate = birthDate[0:10] // Patient.birthDate is date-only.
		}
		patient.BirthDate = birthDate
	}

	if addr := pid.Field(11); addr.String() != "" {
		var lines []string
		if line := addr.Component(1); line != "" {
			lines = append(lines, line)
		}
		patient.Address = []fhir.Address{{
			Line:       lines,
			City:       addr.Component(3),
			State:      addr.Component(4),
			PostalCode: addr.Component(5),
			Country:    addr.Component(6),
		}}
	}

	var telecom []fhir.ContactPoint
	if home := pid.Field(13).String(); home != "" {
		telecom = append(telecom, fhir.ContactPoint{System: "phone", Value: home, Use: "home"})
	}
	if work := pid.Field(14).String(); work != "" {
		telecom = append(telecom, fhir.ContactPoint{System: "phone", Value: work, Use: "work"})
	}
	if len(telecom) > 0 {
		patient.Telecom = telecom
	}

	return patient, nil
}
