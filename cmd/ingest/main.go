package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mfmadarang/fhir-interop/internal/convert"
	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/hl7v2"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/validate"
)

func main() {
	format := flag.String("format", "fhir", `input format: "fhir" (JSON bundles) or "hl7v2" (pipe-delimited messages)`)
	flag.Parse()

	dir := "testdata"
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}

	var glob string
	switch *format {
	case "fhir":
		glob = "*.json"
	case "hl7v2":
		glob = "*.hl7"
	default:
		log.Fatalf("unknown format %q: must be \"fhir\" or \"hl7v2\"", *format)
	}

	db, err := store.Connect()
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, glob))
	if err != nil {
		log.Fatalf("listing files in %s: %v", dir, err)
	}
	if len(files) == 0 {
		log.Fatalf("no %s files found in %s", glob, dir)
	}

	var loaded, skipped int
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("skipping %s: reading file: %v", path, err)
			skipped++
			continue
		}

		var parsed *fhir.ParsedBundle
		switch *format {
		case "fhir":
			parsed, err = fhir.ParseBundle(data)
		case "hl7v2":
			parsed, err = parseHL7v2(data)
		}
		if err != nil {
			log.Printf("skipping %s: parsing: %v", path, err)
			skipped++
			continue
		}

		if issues := validate.ValidateBundle(parsed); len(issues) > 0 {
			log.Printf("%s: %d validation issue(s):", filepath.Base(path), len(issues))
			for _, iss := range issues {
				log.Printf("  - %s", iss)
			}
		}

		if err := store.SaveBundle(db, parsed); err != nil {
			log.Printf("skipping %s: saving bundle: %v", path, err)
			skipped++
			continue
		}

		loaded++
		fmt.Printf("loaded %s (%d patients, %d encounters, %d observations)\n",
			filepath.Base(path), len(parsed.Patients), len(parsed.Encounters), len(parsed.Observations))
	}

	fmt.Printf("\ndone: %d bundle(s) loaded, %d skipped\n", loaded, skipped)
}

func parseHL7v2(data []byte) (*fhir.ParsedBundle, error) {
	msg, err := hl7v2.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing HL7v2 message: %w", err)
	}

	msh := msg.Segment("MSH")
	if msh == nil {
		return nil, fmt.Errorf("message has no MSH segment")
	}
	messageCode := msh.Field(9).Component(1)
	triggerEvent := msh.Field(9).Component(2)

	patient, err := convert.PatientFromADT(msg)
	if err != nil {
		return nil, fmt.Errorf("converting patient: %w", err)
	}
	parsed := &fhir.ParsedBundle{Patients: []fhir.Patient{patient}}

	switch {
	case messageCode == "ADT" && (triggerEvent == "A01" || triggerEvent == "A03"):
		if msg.Segment("PV1") == nil {
			return nil, fmt.Errorf("ADT^%s message has no PV1 segment", triggerEvent)
		}
		encounter, err := convert.EncounterFromADT(msg, patient.ID)
		if err != nil {
			return nil, fmt.Errorf("converting encounter: %w", err)
		}
		parsed.Encounters = []fhir.Encounter{encounter}

	case messageCode == "ORU" && triggerEvent == "R01":
		var encID string
		if msg.Segment("PV1") != nil {
			id, err := convert.EncounterIDFromPV1(msg)
			if err != nil {
				return nil, fmt.Errorf("deriving encounter: %w", err)
			}
			encID = id
		}
		observations, err := convert.ObservationsFromORU(msg, patient.ID, encID)
		if err != nil {
			return nil, fmt.Errorf("converting observations: %w", err)
		}
		parsed.Observations = observations

	default:
		return nil, fmt.Errorf("unsupported message type %s^%s", messageCode, triggerEvent)
	}

	return parsed, nil
}
