package validate

import (
	"context"
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
)

type CodeValidator interface {
	ValidateCode(ctx context.Context, system, code string) (*terminology.Result, error)
}

func ValidateObservationTerminology(ctx context.Context, client CodeValidator, o fhir.Observation) []Issue {
	var issues []Issue
	add := func(field, msg string) {
		issues = append(issues, Issue{ResourceType: "Observation", ResourceID: o.ID, Field: field, Message: msg})
	}

	for _, c := range o.Code.Coding {
		if c.System != terminology.SystemLOINC && c.System != terminology.SystemSNOMED {
			continue
		}
		if c.Code == "" {
			continue
		}

		result, err := client.ValidateCode(ctx, c.System, c.Code)
		if err != nil {
			add("code", fmt.Sprintf("could not verify code %q against %s (marked unverified): %v", c.Code, c.System, err))
			continue
		}
		if !result.Valid {
			msg := fmt.Sprintf("code %q not valid in %s", c.Code, c.System)
			if result.Message != "" {
				msg = fmt.Sprintf("%s: %s", msg, result.Message)
			}
			add("code", msg)
		}
	}

	return issues
}
