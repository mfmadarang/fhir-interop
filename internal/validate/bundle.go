package validate

import (
	"context"
	"sync"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
)

const terminologyWorkerCount = 10

func ValidateBundle(pb *fhir.ParsedBundle) []Issue {
	var issues []Issue

	for _, p := range pb.Patients {
		issues = append(issues, ValidatePatient(p)...)
	}
	for _, e := range pb.Encounters {
		issues = append(issues, ValidateEncounter(e)...)
	}
	for _, o := range pb.Observations {
		issues = append(issues, ValidateObservation(o)...)
	}

	return issues
}

func ValidateBundleTerminology(ctx context.Context, client CodeValidator, pb *fhir.ParsedBundle) []Issue {
	observations := pb.Observations
	if len(observations) == 0 {
		return nil
	}

	workers := terminologyWorkerCount
	if len(observations) < workers {
		workers = len(observations)
	}

	jobs := make(chan fhir.Observation)
	results := make(chan []Issue)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for o := range jobs {
				results <- ValidateObservationTerminology(ctx, client, o)
			}
		}()
	}

	go func() {
		for _, o := range observations {
			jobs <- o
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var issues []Issue
	for r := range results {
		issues = append(issues, r...)
	}

	return issues
}
