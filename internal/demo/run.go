package demo

import (
	"context"
	"fmt"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/terminology"
	"github.com/mfmadarang/fhir-interop/internal/validate"
	"gorm.io/gorm"
)

const simulatedFailureCode = "9999-DEMO"

type demoCodeValidator struct {
	inner validate.CodeValidator
}

func (v *demoCodeValidator) ValidateCode(ctx context.Context, system, code string) (*terminology.Result, error) {
	if code == simulatedFailureCode {
		return nil, fmt.Errorf("simulated terminology server timeout")
	}
	return v.inner.ValidateCode(ctx, system, code)
}

func RunPipeline(job *Job, db *gorm.DB, client validate.CodeValidator, data []byte) {
	defer close(job.Done)

	job.Events <- StageEvent{Stage: StageParse, Status: StatusRunning}
	pb, err := fhir.ParseBundle(data)
	if err != nil {
		job.Events <- StageEvent{Stage: StageParse, Status: StatusError, Message: err.Error()}
		return
	}
	job.Events <- StageEvent{Stage: StageParse, Status: StatusDone}

	job.Events <- StageEvent{Stage: StageValidate, Status: StatusRunning}
	issues := validate.ValidateBundle(pb)
	msg := ""
	if len(issues) > 0 {
		msg = fmt.Sprintf("%d issue(s) found", len(issues))
	}
	job.Events <- StageEvent{Stage: StageValidate, Status: StatusDone, Message: msg}

	job.Events <- StageEvent{Stage: StageTerminology, Status: StatusRunning}
	wrapped := &demoCodeValidator{inner: client}
	runTerminologyChecks(job, wrapped, pb)
	job.Events <- StageEvent{Stage: StageTerminology, Status: StatusDone}

	job.Events <- StageEvent{Stage: StagePersist, Status: StatusRunning}
	if err := store.SaveBundle(db, pb); err != nil {
		job.Events <- StageEvent{Stage: StagePersist, Status: StatusError, Message: err.Error()}
		return
	}
	job.Events <- StageEvent{Stage: StagePersist, Status: StatusDone}
}

func runTerminologyChecks(job *Job, client validate.CodeValidator, pb *fhir.ParsedBundle) {
	ctx := context.Background()

	for _, o := range pb.Observations {
		for _, c := range o.Code.Coding {
			if c.System != terminology.SystemLOINC && c.System != terminology.SystemSNOMED {
				continue
			}
			if c.Code == "" {
				continue
			}

			result, err := client.ValidateCode(ctx, c.System, c.Code)
			simulated := c.Code == simulatedFailureCode

			var ev TerminologyEvent
			switch {
			case err != nil:
				ev = TerminologyEvent{
					ObservationID: o.ID,
					System:        c.System,
					Code:          c.Code,
					Display:       c.Display,
					Result:        TerminologyUnverified,
					Message:       err.Error(),
					Simulated:     simulated,
				}
			case !result.Valid:
				ev = TerminologyEvent{
					ObservationID: o.ID,
					System:        c.System,
					Code:          c.Code,
					Display:       result.Display,
					Result:        TerminologyInvalid,
					Message:       result.Message,
				}
			default:
				ev = TerminologyEvent{
					ObservationID: o.ID,
					System:        c.System,
					Code:          c.Code,
					Display:       result.Display,
					Result:        TerminologyValid,
				}
			}

			job.Events <- ev
		}
	}
}
