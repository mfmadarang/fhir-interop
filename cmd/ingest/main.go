package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/validate"
)

func main() {
	dir := "testdata"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	db, err := store.Connect()
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		log.Fatalf("running migrations: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		log.Fatalf("listing bundle files in %s: %v", dir, err)
	}
	if len(files) == 0 {
		log.Fatalf("no .json files found in %s", dir)
	}

	var loaded, skipped int
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("skipping %s: reading file: %v", path, err)
			skipped++
			continue
		}

		parsed, err := fhir.ParseBundle(data)
		if err != nil {
			log.Printf("skipping %s: parsing bundle: %v", path, err)
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
