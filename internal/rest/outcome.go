package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mfmadarang/fhir-interop/internal/validate"
)

// a FHIR OperationOutcome, used for every error and for validate results
type operationOutcome struct {
	ResourceType string         `json:"resourceType"`
	Issue        []outcomeIssue `json:"issue"`
}

type outcomeIssue struct {
	Severity   string       `json:"severity"`
	Code       string       `json:"code"`
	Details    *outcomeText `json:"details,omitempty"`
	Expression []string     `json:"expression,omitempty"`
}

type outcomeText struct {
	Text string `json:"text"`
}

func newOutcome(severity, code, msg string) operationOutcome {
	return operationOutcome{
		ResourceType: "OperationOutcome",
		Issue:        []outcomeIssue{{Severity: severity, Code: code, Details: &outcomeText{Text: msg}}},
	}
}

func writeOutcome(w http.ResponseWriter, status int, oo operationOutcome) {
	w.Header().Set("Content-Type", fhirJSON)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(oo)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeOutcome(w, status, newOutcome("error", code, msg))
}

// one outcome issue per validation issue, with the resource and field as the expression
func outcomeFromIssues(issues []validate.Issue) operationOutcome {
	oo := operationOutcome{ResourceType: "OperationOutcome", Issue: make([]outcomeIssue, 0, len(issues))}
	for _, iss := range issues {
		expr := iss.ResourceType
		if iss.Field != "" {
			expr += "." + iss.Field
		}
		id := iss.ResourceID
		if id == "" {
			id = "(no id)"
		}
		oo.Issue = append(oo.Issue, outcomeIssue{
			Severity:   "error",
			Code:       "invalid",
			Details:    &outcomeText{Text: iss.ResourceType + "/" + id + ": " + iss.Message},
			Expression: []string{expr},
		})
	}
	return oo
}
