package rest

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/mfmadarang/fhir-interop/internal/store"
)

const fhirJSON = "application/fhir+json"

const (
	defaultSearchLimit = 50
	maxSearchLimit     = 200
)

// the bit of the store package the handlers need, as an interface so tests can fake it
type patientStore interface {
	GetPatient(id string) (*store.PatientRecord, error)
	SearchPatients(s store.PatientSearch) ([]*store.PatientRecord, error)
}

type Handler struct {
	patients patientStore
}

func New(p patientStore) *Handler {
	return &Handler{patients: p}
}

// mux for the /fhir endpoints
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /fhir/Patient/{id}", h.readPatient)
	mux.HandleFunc("GET /fhir/Patient", h.searchPatients)
	return mux
}

func (h *Handler) readPatient(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	rec, err := h.patients.GetPatient(id)
	if err != nil {
		http.Error(w, "looking up patient: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if rec == nil {
		http.Error(w, "Patient "+id+" not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", fhirJSON)
	w.Write(rec.Raw)
}

func (h *Handler) searchPatients(w http.ResponseWriter, r *http.Request) {
	recs, err := h.patients.SearchPatients(parsePatientSearch(r))
	if err != nil {
		http.Error(w, "searching patients: "+err.Error(), http.StatusInternalServerError)
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
