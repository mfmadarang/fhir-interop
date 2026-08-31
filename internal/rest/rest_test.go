package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/store"
)

// fakeStore is an in-memory patientStore for the handler tests.
type fakeStore struct {
	byID       map[string]*store.PatientRecord
	lastSearch store.PatientSearch
	searchOut  []*store.PatientRecord
}

func (f *fakeStore) GetPatient(id string) (*store.PatientRecord, error) {
	return f.byID[id], nil
}

func (f *fakeStore) SearchPatients(s store.PatientSearch) ([]*store.PatientRecord, error) {
	f.lastSearch = s
	return f.searchOut, nil
}

func patient(id, family string) *store.PatientRecord {
	raw := `{"resourceType":"Patient","id":"` + id + `","name":[{"family":"` + family + `"}]}`
	return &store.PatientRecord{ID: id, FamilyName: family, Raw: store.JSON(raw)}
}

func TestReadPatientFound(t *testing.T) {
	f := &fakeStore{byID: map[string]*store.PatientRecord{"p1": patient("p1", "Reyes")}}
	srv := New(f).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient/p1", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != fhirJSON {
		t.Fatalf("Content-Type = %q, want %q", ct, fhirJSON)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not JSON: %v", err)
	}
	if got["id"] != "p1" || got["resourceType"] != "Patient" {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestReadPatientNotFound(t *testing.T) {
	f := &fakeStore{byID: map[string]*store.PatientRecord{}}
	srv := New(f).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestSearchPatientsBundle(t *testing.T) {
	f := &fakeStore{searchOut: []*store.PatientRecord{patient("p1", "Reyes"), patient("p2", "Reyes")}}
	srv := New(f).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient?family=Reyes&gender=female&_count=10", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var b bundle
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("body is not a Bundle: %v", err)
	}
	if b.ResourceType != "Bundle" || b.Type != "searchset" {
		t.Fatalf("got resourceType=%q type=%q", b.ResourceType, b.Type)
	}
	if b.Total != 2 || len(b.Entry) != 2 {
		t.Fatalf("got total=%d entries=%d, want 2/2", b.Total, len(b.Entry))
	}

	// the query params should have been passed through to the store
	if f.lastSearch.Family != "Reyes" || f.lastSearch.Gender != "female" || f.lastSearch.Limit != 10 {
		t.Fatalf("search params not forwarded: %+v", f.lastSearch)
	}
}

func TestSearchPatientsEmpty(t *testing.T) {
	f := &fakeStore{searchOut: nil}
	srv := New(f).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var b bundle
	json.Unmarshal(rec.Body.Bytes(), &b)
	if b.Total != 0 || len(b.Entry) != 0 {
		t.Fatalf("got total=%d entries=%d, want 0/0", b.Total, len(b.Entry))
	}
	if f.lastSearch.Limit != defaultSearchLimit {
		t.Fatalf("default limit = %d, want %d", f.lastSearch.Limit, defaultSearchLimit)
	}
}

func TestSearchLimitCap(t *testing.T) {
	f := &fakeStore{}
	srv := New(f).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=9999", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if f.lastSearch.Limit != maxSearchLimit {
		t.Fatalf("limit = %d, want capped at %d", f.lastSearch.Limit, maxSearchLimit)
	}
}
