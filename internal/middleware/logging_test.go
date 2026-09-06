package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingCapturesStatusAndLogsOnlyMethodPathStatus(t *testing.T) {
	var buf bytes.Buffer
	origWriter := log.Writer()
	origFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(origWriter)
		log.SetFlags(origFlags)
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest(http.MethodGet, "/flags/my-key/evaluate?user=alice&secret=x", nil)
	rec := httptest.NewRecorder()

	Logging(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}

	logged := buf.String()

	if !strings.Contains(logged, "GET") {
		t.Errorf("expected log to contain method GET, got %q", logged)
	}
	if !strings.Contains(logged, "/flags/my-key/evaluate") {
		t.Errorf("expected log to contain path, got %q", logged)
	}
	if !strings.Contains(logged, "404") {
		t.Errorf("expected log to contain status 404, got %q", logged)
	}

	if strings.Contains(logged, "?") {
		t.Errorf("expected no query string in log, got %q", logged)
	}
	if strings.Contains(logged, "user") || strings.Contains(logged, "alice") {
		t.Errorf("expected no user parameter in log, got %q", logged)
	}
	if strings.Contains(logged, "secret") {
		t.Errorf("expected no query values in log, got %q", logged)
	}
}

func TestLoggingDefaultStatusOK(t *testing.T) {
	var buf bytes.Buffer
	origWriter := log.Writer()
	origFlags := log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	defer func() {
		log.SetOutput(origWriter)
		log.SetFlags(origFlags)
	}()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/flags", nil)
	rec := httptest.NewRecorder()

	Logging(handler).ServeHTTP(rec, req)

	logged := buf.String()
	if !strings.Contains(logged, "POST /flags 200") {
		t.Errorf("expected 'POST /flags 200' in log, got %q", logged)
	}
}
