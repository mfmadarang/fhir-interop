package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/validate"
)

const fhirJSON = "application/fhir+json"

const (
	defaultSearchLimit = 50
	maxSearchLimit     = 200

	maxBodyBytes = 10 << 20
)

// the bit of the store package the handlers need, as an interface so tests can fake it
type patientStore interface {
	GetPatient(id string) (*store.PatientRecord, error)
	SearchPatients(s store.PatientSearch) ([]*store.PatientRecord, error)
}

// saves parsed resources, same idea as patientStore
type resourceWriter interface {
	SaveBundle(pb *fhir.ParsedBundle) error
}

type Handler struct {
	patients patientStore
	writer   resourceWriter
	codes    validate.CodeValidator // nil skips the terminology check
	guard    func(http.Handler) http.Handler
}

// guard wraps the POST routes (the API key check), nil means no check
func New(p patientStore, w resourceWriter, codes validate.CodeValidator, guard func(http.Handler) http.Handler) *Handler {
	if guard == nil {
		guard = func(next http.Handler) http.Handler { return next }
	}
	return &Handler{patients: p, writer: w, codes: codes, guard: guard}
}

// mux for the /fhir endpoints
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /fhir/Patient/{id}", h.readPatient)
	mux.HandleFunc("GET /fhir/Patient", h.searchPatients)
	mux.Handle("POST /fhir/$validate", h.guard(http.HandlerFunc(h.validate)))
	mux.Handle("POST /fhir", h.guard(http.HandlerFunc(h.create)))
	return mux
}

func (h *Handler) readPatient(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	rec, err := h.patients.GetPatient(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "exception", "looking up patient: "+err.Error())
		return
	}
	if rec == nil {
		writeError(w, http.StatusNotFound, "not-found", "Patient "+id+" not found")
		return
	}

	w.Header().Set("Content-Type", fhirJSON)
	w.Write(rec.Raw)
}

func (h *Handler) searchPatients(w http.ResponseWriter, r *http.Request) {
	recs, err := h.patients.SearchPatients(parsePatientSearch(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "exception", "searching patients: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", fhirJSON)
	json.NewEncoder(w).Encode(buildSearchBundle(recs))
}

// pulls the supported search params (family, given, gender, birthdate, _count) off the query string
func parsePatientSearch(r *http.Request) store.PatientSearch {
	q := r.URL.Query()

	limit := defaultSearchLimit
	if raw := q.Get("_count"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	return store.PatientSearch{
		Family:    q.Get("family"),
		Given:     q.Get("given"),
		Gender:    q.Get("gender"),
		BirthDate: q.Get("birthdate"),
		Limit:     limit,
	}
}

// a FHIR searchset Bundle; each entry's resource is the raw stored Patient JSON
type bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Total        int           `json:"total"`
	Entry        []bundleEntry `json:"entry"`
}

type bundleEntry struct {
	Resource json.RawMessage `json:"resource"`
}

func buildSearchBundle(recs []*store.PatientRecord) bundle {
	b := bundle{
		ResourceType: "Bundle",
		Type:         "searchset",
		Total:        len(recs),
		Entry:        make([]bundleEntry, 0, len(recs)),
	}
	for _, rec := range recs {
		b.Entry = append(b.Entry, bundleEntry{Resource: json.RawMessage(rec.Raw)})
	}
	return b
}

// reads the posted body and parses it, writing the error response itself if that fails
func (h *Handler) readBody(w http.ResponseWriter, r *http.Request) (*fhir.ParsedBundle, bool) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid", "reading body: "+err.Error())
		return nil, false
	}
	parsed, err := fhir.ParseResource(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid", err.Error())
		return nil, false
	}
	return parsed, true
}

// structural checks, plus the terminology check when a code validator is set
func (h *Handler) check(r *http.Request, parsed *fhir.ParsedBundle) []validate.Issue {
	issues := validate.ValidateBundle(parsed)
	if h.codes != nil {
		issues = append(issues, validate.ValidateBundleTerminology(r.Context(), h.codes, parsed)...)
	}
	return issues
}

// POST /fhir/$validate: checks the resource and says what is wrong, saves nothing
func (h *Handler) validate(w http.ResponseWriter, r *http.Request) {
	parsed, ok := h.readBody(w, r)
	if !ok {
		return
	}

	issues := h.check(r, parsed)
	if len(issues) == 0 {
		writeOutcome(w, http.StatusOK, newOutcome("information", "informational", "no issues found"))
		return
	}
	writeOutcome(w, http.StatusOK, outcomeFromIssues(issues))
}

// POST /fhir: validates, then saves. a resource with issues is rejected
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	parsed, ok := h.readBody(w, r)
	if !ok {
		return
	}

	if issues := h.check(r, parsed); len(issues) > 0 {
		writeOutcome(w, http.StatusUnprocessableEntity, outcomeFromIssues(issues))
		return
	}

	if err := h.writer.SaveBundle(parsed); err != nil {
		writeError(w, http.StatusInternalServerError, "exception", "saving: "+err.Error())
		return
	}
	writeOutcome(w, http.StatusCreated, newOutcome("information", "informational", "saved"))
}
