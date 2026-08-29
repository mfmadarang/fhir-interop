package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebHandlerServesEmbeddedIndex(t *testing.T) {
	h, err := webHandler()
	if err != nil {
		t.Fatalf("webHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /app/: status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "<!DOCTYPE html>") {
		t.Fatalf("GET /app/: body does not look like the browser UI")
	}
}

func TestWebHandlerMissingFileIs404(t *testing.T) {
	h, err := webHandler()
	if err != nil {
		t.Fatalf("webHandler() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/nope.css", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
