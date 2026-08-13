package convert

import (
	"strings"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/hl7v2"
)

// assembles a PV1 segment with only the fields this project's converter reads populated
func buildPV1(class, location, attending, hospitalService, visitNumber, admitDateTime, dischargeDateTime string) string {
	fields := make([]string, 45)
	fields[0] = "1"
	fields[1] = class
	fields[2] = location
	fields[6] = attending
	fields[9] = hospitalService
	fields[18] = visitNumber
	fields[43] = admitDateTime
	fields[44] = dischargeDateTime
	return "PV1|" + strings.Join(fields, "|")
}

func sampleADT(pv1 string) string {
	return "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20260813120000||ADT^A01|MSG00001|P|2.5\r" +
		"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F|||123 Main St^^Springfield^IL^62704||555-0100|||S||1000001\r" +
		pv1 + "\r"
}

func TestPatientFromADT(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "V001", "20260813080000", "")
	msg, err := hl7v2.Parse([]byte(sampleADT(pv1)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	patient, err := PatientFromADT(msg)
	if err != nil {
		t.Fatalf("PatientFromADT() error = %v", err)
	}

	if patient.ID != "1000001" {
		t.Errorf("ID = %q, want %q", patient.ID, "1000001")
	}
	if patient.Gender != "female" {
		t.Errorf("Gender = %q, want %q", patient.Gender, "female")
	}
	if patient.BirthDate != "1985-03-15" {
		t.Errorf("BirthDate = %q, want %q", patient.BirthDate, "1985-03-15")
	}
	if len(patient.Name) != 1 || patient.Name[0].Family != "Doe" {
		t.Fatalf("Name = %+v, want family Doe", patient.Name)
	}
	if got := patient.Name[0].Given; len(got) < 1 || got[0] != "Jane" {
		t.Errorf("Given = %v, want first element %q", got, "Jane")
	}
	if len(patient.Address) != 1 || patient.Address[0].City != "Springfield" {
		t.Fatalf("Address = %+v, want city Springfield", patient.Address)
	}
	if len(patient.Telecom) != 1 || patient.Telecom[0].Value != "555-0100" {
		t.Fatalf("Telecom = %+v, want home 555-0100", patient.Telecom)
	}
}

func TestEncounterFromADT(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "V001", "20260813080000", "")
	msg, err := hl7v2.Parse([]byte(sampleADT(pv1)))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	encounter, err := EncounterFromADT(msg, "1000001")
	if err != nil {
		t.Fatalf("EncounterFromADT() error = %v", err)
	}

	if encounter.ID != "V001" {
		t.Errorf("ID = %q, want %q", encounter.ID, "V001")
	}
	if encounter.Status != "in-progress" {
		t.Errorf("Status = %q, want %q", encounter.Status, "in-progress")
	}
	if encounter.Class.Code != "IMP" {
		t.Errorf("Class.Code = %q, want %q", encounter.Class.Code, "IMP")
	}
	if encounter.Subject.Reference != "Patient/1000001" {
		t.Errorf("Subject.Reference = %q, want %q", encounter.Subject.Reference, "Patient/1000001")
	}
	if encounter.Period.Start != "2026-08-13T08:00:00Z" {
		t.Errorf("Period.Start = %q, want %q", encounter.Period.Start, "2026-08-13T08:00:00Z")
	}
}

func TestEncounterFromADT_FallsBackToMessageControlID(t *testing.T) {
	pv1 := buildPV1("O", "CLINIC1", "5678^Cruz^Ana^^^Dr", "", "", "", "")
	raw := "MSH|^~\\&|A|B|C|D|20260813120000||ADT^A01|MSG00099|P|2.5\r" +
		"PID|1||1000002^^^MRN||Reyes^Carlos||19700101|M\r" +
		pv1 + "\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	encounter, err := EncounterFromADT(msg, "1000002")
	if err != nil {
		t.Fatalf("EncounterFromADT() error = %v", err)
	}

	if encounter.ID != "encounter-MSG00099" {
		t.Errorf("ID = %q, want %q", encounter.ID, "encounter-MSG00099")
	}
	if encounter.Class.Code != "AMB" {
		t.Errorf("Class.Code = %q, want %q", encounter.Class.Code, "AMB")
	}
}

const sampleORU = "MSH|^~\\&|LAB|FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20260813130000||ORU^R01|MSG00002|P|2.5\r" +
	"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F\r" +
	"OBR|1|ORD001|LABRESULT001|CBC^Complete Blood Count^L|||20260813125000\r" +
	"OBX|1|NM|8867-4^Heart Rate^LN||72|/min|||||F\r" +
	"OBX|2|NM|8480-6^Systolic BP^LN||120|mmHg|||||F\r" +
	"OBX|3|ST|10154-3^Free Text Result^LN||Patient reports feeling well|||||F\r"

func TestObservationsFromORU(t *testing.T) {
	msg, err := hl7v2.Parse([]byte(sampleORU))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	obs, err := ObservationsFromORU(msg, "1000001", "V001")
	if err != nil {
		t.Fatalf("ObservationsFromORU() error = %v", err)
	}
	if len(obs) != 3 {
		t.Fatalf("got %d observations, want 3", len(obs))
	}

	hr := obs[0]
	if hr.ID != "1000001-obs-1" {
		t.Errorf("ID = %q, want %q", hr.ID, "1000001-obs-1")
	}
	if hr.Code.Coding[0].System != "http://loinc.org" || hr.Code.Coding[0].Code != "8867-4" {
		t.Errorf("Code.Coding[0] = %+v, want LOINC 8867-4", hr.Code.Coding[0])
	}
	if hr.ValueQuantity == nil || hr.ValueQuantity.Value != 72 || hr.ValueQuantity.Unit != "/min" {
		t.Errorf("ValueQuantity = %+v, want 72 /min", hr.ValueQuantity)
	}
	if hr.Status != "final" {
		t.Errorf("Status = %q, want %q", hr.Status, "final")
	}
	if hr.Subject.Reference != "Patient/1000001" {
		t.Errorf("Subject.Reference = %q, want %q", hr.Subject.Reference, "Patient/1000001")
	}
	if hr.Encounter.Reference != "Encounter/V001" {
		t.Errorf("Encounter.Reference = %q, want %q", hr.Encounter.Reference, "Encounter/V001")
	}
	if hr.EffectiveDateTime != "2026-08-13T12:50:00Z" {
		t.Errorf("EffectiveDateTime = %q, want %q", hr.EffectiveDateTime, "2026-08-13T12:50:00Z")
	}

	freeText := obs[2]
	if freeText.ValueString == nil || *freeText.ValueString != "Patient reports feeling well" {
		t.Errorf("ValueString = %v, want %q", freeText.ValueString, "Patient reports feeling well")
	}
	if freeText.ValueQuantity != nil {
		t.Errorf("ValueQuantity = %+v, want nil for ST type", freeText.ValueQuantity)
	}
}

func TestObservationsFromORU_FallsBackToMessageTimestampWithoutOBR(t *testing.T) {
	raw := "MSH|^~\\&|LAB|FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20260813130000||ORU^R01|MSG00004|P|2.5\r" +
		"PID|1||1000003^^^MRN||Santos^Maria||19950620|F\r" +
		"OBX|1|NM|8867-4^Heart Rate^LN||80|/min|||||F\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	obs, err := ObservationsFromORU(msg, "1000003", "")
	if err != nil {
		t.Fatalf("ObservationsFromORU() error = %v", err)
	}
	if len(obs) != 1 {
		t.Fatalf("got %d observations, want 1", len(obs))
	}
	if obs[0].EffectiveDateTime != "2026-08-13T13:00:00Z" {
		t.Errorf("EffectiveDateTime = %q, want %q (fallback to MSH-7)", obs[0].EffectiveDateTime, "2026-08-13T13:00:00Z")
	}
	if obs[0].Encounter.Reference != "" {
		t.Errorf("Encounter.Reference = %q, want empty when encounterID is empty", obs[0].Encounter.Reference)
	}
}

func TestObservationsFromORU_NoPID(t *testing.T) {
	raw := "MSH|^~\\&|A|B|C|D|20260813120000||ORU^R01|MSG00003|P|2.5\r" +
		"OBX|1|NM|8867-4^Heart Rate^LN||72|/min\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if _, err := ObservationsFromORU(msg, "1000001", ""); err == nil {
		t.Fatal("expected error for missing PID, got nil")
	}
}

func TestEncounterFromADT_Discharge(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "V001", "20260813080000", "20260814100000")
	raw := "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20260814100000||ADT^A03|MSG00005|P|2.5\r" +
		"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F\r" +
		pv1 + "\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	encounter, err := EncounterFromADT(msg, "1000001")
	if err != nil {
		t.Fatalf("EncounterFromADT() error = %v", err)
	}

	if encounter.Status != "finished" {
		t.Errorf("Status = %q, want %q", encounter.Status, "finished")
	}
	if encounter.Period.End != "2026-08-14T10:00:00Z" {
		t.Errorf("Period.End = %q, want %q", encounter.Period.End, "2026-08-14T10:00:00Z")
	}
}

func TestEncounterFromADT_UnsupportedTriggerEvent(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "V001", "", "")
	raw := "MSH|^~\\&|A|B|C|D|20260814100000||ADT^A02|MSG00006|P|2.5\r" +
		"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F\r" +
		pv1 + "\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if _, err := EncounterFromADT(msg, "1000001"); err == nil {
		t.Fatal("expected error for unsupported trigger event A02, got nil")
	}
}

func TestEncounterIDFromPV1(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "V001", "", "")
	raw := "MSH|^~\\&|LAB|FACILITY|C|D|20260813130000||ORU^R01|MSG00010|P|2.5\r" +
		"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F\r" +
		pv1 + "\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	id, err := EncounterIDFromPV1(msg)
	if err != nil {
		t.Fatalf("EncounterIDFromPV1() error = %v", err)
	}
	if id != "V001" {
		t.Errorf("id = %q, want %q", id, "V001")
	}
}

func TestEncounterIDFromPV1_FallsBackToMessageControlID(t *testing.T) {
	pv1 := buildPV1("I", "WARD1^101^1", "1234^Smith^Robert^^^Dr", "MED", "", "", "")
	raw := "MSH|^~\\&|LAB|FACILITY|C|D|20260813130000||ORU^R01|MSG00011|P|2.5\r" +
		"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F\r" +
		pv1 + "\r"

	msg, err := hl7v2.Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	id, err := EncounterIDFromPV1(msg)
	if err != nil {
		t.Fatalf("EncounterIDFromPV1() error = %v", err)
	}
	if id != "encounter-MSG00011" {
		t.Errorf("id = %q, want %q", id, "encounter-MSG00011")
	}
}
