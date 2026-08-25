package terminology

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateCode_Valid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"resourceType": "Parameters",
			"parameter": []map[string]any{
				{"name": "result", "valueBoolean": true},
				{"name": "display", "valueString": "Heart rate"},
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithBaseURL(srv.URL)
	result, err := client.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid=true")
	}
	if result.Display != "Heart rate" {
		t.Errorf("expected display %q, got %q", "Heart rate", result.Display)
	}
}

func TestValidateCode_Invalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"resourceType": "Parameters",
			"parameter": []map[string]any{
				{"name": "result", "valueBoolean": false},
				{"name": "message", "valueString": "not a valid code"},
			},
		})
	}))
	defer srv.Close()

	client := NewClientWithBaseURL(srv.URL)
	result, err := client.ValidateCode(context.Background(), SystemLOINC, "bogus")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Valid {
		t.Errorf("expected valid=false")
	}
	if result.Message != "not a valid code" {
		t.Errorf("expected message %q, got %q", "not a valid code", result.Message)
	}
}

func TestValidateCode_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClientWithBaseURL(srv.URL)
	_, err := client.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateCode_MalformedResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	client := NewClientWithBaseURL(srv.URL)
	_, err := client.ValidateCode(context.Background(), SystemLOINC, "8867-4")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
