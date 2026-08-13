package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApiKeyMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"valid key", "Bearer secret123", http.StatusOK},
		{"wrong key", "Bearer wrong", http.StatusUnauthorized},
		{"missing header", "", http.StatusUnauthorized},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mw := APIKeyMiddleware("secret123")(next)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/query", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			mw.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
