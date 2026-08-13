package hl7v2

import "testing"

const sampleADT = "MSH|^~\\&|SENDING_APP|SENDING_FACILITY|RECEIVING_APP|RECEIVING_FACILITY|20260813120000||ADT^A01|MSG00001|P|2.5\r" +
	"PID|1||1000001^^^MRN||Doe^Jane^A||19850315|F|||123 Main St^^Springfield^IL^62704||555-0100|||S||1000001\r" +
	"PV1|1|I|WARD1^101^1||||1234^Smith^Robert^^^Dr|||MED\r"

func TestParse_ADT_A01(t *testing.T) {
	msg, err := Parse([]byte(sampleADT))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	msh := msg.Segment("MSH")
	if msh == nil {
		t.Fatal("MSH segment not found")
	}
	if got := msh.Field(9).Component(1); got != "ADT" {
		t.Errorf("MSH-9.1 = %q, want %q", got, "ADT")
	}
	if got := msh.Field(9).Component(2); got != "A01" {
		t.Errorf("MSH-9.2 = %q, want %q", got, "A01")
	}

	pid := msg.Segment("PID")
	if pid == nil {
		t.Fatal("PID segment not found")
	}
	if got := pid.Field(5).Component(1); got != "Doe" {
		t.Errorf("PID-5.1 (family name) = %q, want %q", got, "Doe")
	}
	if got := pid.Field(5).Component(2); got != "Jane" {
		t.Errorf("PID-5.2 (given name) = %q, want %q", got, "Jane")
	}
	if got := pid.Field(7); got != "19850315" {
		t.Errorf("PID-7 (DOB) = %q, want %q", got, "19850315")
	}
	if got := pid.Field(8); got != "F" {
		t.Errorf("PID-8 (sex) = %q, want %q", got, "F")
	}

	pv1 := msg.Segment("PV1")
	if pv1 == nil {
		t.Fatal("PV1 segment not found")
	}
	if got := pv1.Field(2); got != "I" {
		t.Errorf("PV1-2 (patient class) = %q, want %q", got, "I")
	}
}

func TestParse_RepeatingSegments(t *testing.T) {
	raw := "MSH|^~\\&|A|B|C|D|20260813120000||ORU^R01|MSG00002|P|2.5\r" +
		"PID|1||2000001^^^MRN||Cruz^Ana||19900101|F\r" +
		"OBX|1|NM|8867-4^Heart Rate^LN||72|/min\r" +
		"OBX|2|NM|8480-6^Systolic BP^LN||120|mmHg\r"

	msg, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	obxs := msg.SegmentsOf("OBX")
	if len(obxs) != 2 {
		t.Fatalf("got %d OBX segments, want 2", len(obxs))
	}
	if got := obxs[0].Field(5); got != "72" {
		t.Errorf("first OBX-5 (value) = %q, want %q", got, "72")
	}
	if got := obxs[1].Field(5); got != "120" {
		t.Errorf("second OBX-5 (value) = %q, want %q", got, "120")
	}
}

func TestParse_Errors(t *testing.T) {
	cases := map[string]string{
		"segment before MSH": "PID|1||123\r",
		"segment too short":  "MSH|^~\\&\rXY\r",
		"empty input":        "",
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(raw)); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
