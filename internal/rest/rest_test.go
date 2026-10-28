package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/fhir"
	"github.com/mfmadarang/fhir-interop/internal/store"
	"github.com/mfmadarang/fhir-interop/internal/validate"
)

// fakeStore is an in-memory patientStore for the handler tests.
type fakeStore struct {
	byID       map[string]*store.PatientRecord
	lastSearch store.PatientSearch
	searchOut  []*store.PatientRecord
	saved      []*fhir.ParsedBundle
	saveErr    error
	cleared    []string
}

func (f *fakeStore) SaveIssues(pb *fhir.ParsedBundle, source string, issues []validate.Issue) error {
	f.cleared = append(f.cleared, source)
	return nil
}

func (f *fakeStore) SaveBundle(pb *fhir.ParsedBundle) error {
	f.saved = append(f.saved, pb)
	return f.saveErr
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
	srv := New(f, f, nil, nil).Routes()

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
	srv := New(f, f, nil, nil).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestSearchPatientsBundle(t *testing.T) {
	f := &fakeStore{searchOut: []*store.PatientRecord{patient("p1", "Reyes"), patient("p2", "Reyes")}}
	srv := New(f, f, nil, nil).Routes()

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
	srv := New(f, f, nil, nil).Routes()

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
	srv := New(f, f, nil, nil).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient?_count=9999", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if f.lastSearch.Limit != maxSearchLimit {
		t.Fatalf("limit = %d, want capped at %d", f.lastSearch.Limit, maxSearchLimit)
	}
}

const goodPatient = `{"resourceType":"Patient","id":"p1","name":[{"family":"Reyes"}],"gender":"female","birthDate":"1990-01-01"}`
const badPatient = `{"resourceType":"Patient","id":"p2","name":[{"family":"Reyes"}],"gender":"robot"}`

func post(srv http.Handler, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, strings.NewReader(body)))
	return rec
}

func outcomeOf(t *testing.T, rec *httptest.ResponseRecorder) operationOutcome {
	t.Helper()
	var oo operationOutcome
	if err := json.Unmarshal(rec.Body.Bytes(), &oo); err != nil || oo.ResourceType != "OperationOutcome" {
		t.Fatalf("body is not an OperationOutcome: %s", rec.Body.String())
	}
	return oo
}

func TestNotFoundIsOperationOutcome(t *testing.T) {
	f := &fakeStore{byID: map[string]*store.PatientRecord{}}
	rec := httptest.NewRecorder()
	New(f, f, nil, nil).Routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient/nope", nil))

	oo := outcomeOf(t, rec)
	if rec.Code != http.StatusNotFound || oo.Issue[0].Code != "not-found" {
		t.Fatalf("got status %d, issue %+v", rec.Code, oo.Issue[0])
	}
}

func TestValidateGood(t *testing.T) {
	f := &fakeStore{}
	rec := post(New(f, f, nil, nil).Routes(), "/fhir/$validate", goodPatient)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if oo := outcomeOf(t, rec); oo.Issue[0].Severity != "information" {
		t.Fatalf("severity = %q, want information", oo.Issue[0].Severity)
	}
	if len(f.saved) != 0 {
		t.Fatal("validate should not save anything")
	}
}

func TestValidateBad(t *testing.T) {
	f := &fakeStore{}
	rec := post(New(f, f, nil, nil).Routes(), "/fhir/$validate", badPatient)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	oo := outcomeOf(t, rec)
	if oo.Issue[0].Severity != "error" || oo.Issue[0].Expression[0] != "Patient.gender" {
		t.Fatalf("unexpected issue: %+v", oo.Issue[0])
	}
	if len(f.saved) != 0 {
		t.Fatal("validate should not save anything")
	}
}

func TestCreateSaves(t *testing.T) {
	f := &fakeStore{}
	rec := post(New(f, f, nil, nil).Routes(), "/fhir", goodPatient)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if len(f.saved) != 1 || len(f.saved[0].Patients) != 1 {
		t.Fatalf("expected one saved patient, got %+v", f.saved)
	}
	if len(f.cleared) != 2 {
		t.Fatalf("expected old issues cleared for both sources, got %v", f.cleared)
	}
}

func TestCreateBundle(t *testing.T) {
	f := &fakeStore{}
	body := `{"resourceType":"Bundle","type":"transaction","entry":[{"resource":` + goodPatient + `}]}`
	rec := post(New(f, f, nil, nil).Routes(), "/fhir", body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", rec.Code, rec.Body.String())
	}
	if len(f.saved) != 1 || len(f.saved[0].Patients) != 1 {
		t.Fatalf("expected one saved patient, got %+v", f.saved)
	}
}

func TestCreateRejectsInvalid(t *testing.T) {
	f := &fakeStore{}
	rec := post(New(f, f, nil, nil).Routes(), "/fhir", badPatient)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", rec.Code)
	}
	outcomeOf(t, rec)
	if len(f.saved) != 0 {
		t.Fatal("invalid resource should not be saved")
	}
}

func TestCreateBadBody(t *testing.T) {
	f := &fakeStore{}
	srv := New(f, f, nil, nil).Routes()

	for name, body := range map[string]string{
		"not json":         "hello",
		"no resourceType":  `{"id":"x"}`,
		"unsupported type": `{"resourceType":"Medication"}`,
	} {
		rec := post(srv, "/fhir", body)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: status = %d, want 400", name, rec.Code)
		}
		outcomeOf(t, rec)
	}
}

func TestCreateSaveError(t *testing.T) {
	f := &fakeStore{saveErr: errors.New("db down")}
	rec := post(New(f, f, nil, nil).Routes(), "/fhir", goodPatient)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	outcomeOf(t, rec)
}

func TestPostGuard(t *testing.T) {
	f := &fakeStore{}
	deny := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		})
	}
	srv := New(f, f, nil, deny).Routes()

	for _, path := range []string{"/fhir", "/fhir/$validate"} {
		if rec := post(srv, path, goodPatient); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: status = %d, want 401", path, rec.Code)
		}
	}
	if len(f.saved) != 0 {
		t.Fatal("blocked request should not save")
	}

	// reads stay open
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fhir/Patient", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", rec.Code)
	}
}
