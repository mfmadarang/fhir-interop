package obs

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mfmadarang/fhir-interop/internal/config"
)

func TestHealthz(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "ok" {
		t.Fatalf("body = %q, want ok", rec.Body.String())
	}
}

func TestMetricsHandlerServesText(t *testing.T) {
	m := NewMetrics()

	// run one request through the middleware so there's something to report
	wrapped := m.Middleware("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))

	rec := httptest.NewRecorder()
	m.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "http_requests_total") {
		t.Fatalf("metrics output missing http_requests_total:\n%s", body)
	}
	if !strings.Contains(body, `route="/test"`) || !strings.Contains(body, `status="418"`) {
		t.Fatalf("metrics output missing the labels from the test request:\n%s", body)
	}
}

func TestRequestLoggerSkipsQuietPaths(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := RequestLogger(logger, []string{"/healthz"}, next)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if buf.Len() != 0 {
		t.Fatalf("expected no log line for skipped path, got: %s", buf.String())
	}

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/query", nil))
	if !strings.Contains(buf.String(), "path=/query") {
		t.Fatalf("expected a log line for /query, got: %s", buf.String())
	}
}

func TestRequestLoggerPreservesFlush(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))

	flushed := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("handler did not receive an http.Flusher")
		}
		f.Flush()
		flushed = true
	})
	h := RequestLogger(logger, nil, next)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/demo/stream", nil))
	if !flushed {
		t.Fatal("expected the handler to flush")
	}
}

func TestNewLoggerFormat(t *testing.T) {
	jsonLog := NewLogger(config.Config{LogLevel: "info", LogFormat: "json"})
	if jsonLog == nil {
		t.Fatal("NewLogger returned nil for json format")
	}
	textLog := NewLogger(config.Config{LogLevel: "debug", LogFormat: "text"})
	if textLog == nil {
		t.Fatal("NewLogger returned nil for text format")
	}
}
