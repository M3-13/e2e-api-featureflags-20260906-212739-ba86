package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"e2e-api-featureflags/internal/store"
)

// seedFlag inserts a flag into the store's unexported map. The store's public
// mutators (Create/Update) belong to the CRUD ticket and are not merged yet, so
// this test seeds the map directly rather than depending on another ticket.
func seedFlag(t *testing.T, s *store.Store, f store.Flag) {
	t.Helper()
	rs := reflect.ValueOf(s).Elem()
	flagsField := rs.FieldByName("flags")
	flags := reflect.NewAt(flagsField.Type(), unsafe.Pointer(flagsField.UnsafeAddr())).Elem()
	flags.SetMapIndex(reflect.ValueOf(f.Key), reflect.ValueOf(f))
}

func TestEvaluateFlagOK(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "f", Enabled: true, RolloutPercent: 100})

	req := httptest.NewRequest(http.MethodGet, "/flags/f/evaluate?user=alice", nil)
	req.SetPathValue("key", "f")
	rr := httptest.NewRecorder()

	EvaluateFlag(s).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var body map[string]bool
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if !body["result"] {
		t.Fatalf("expected result true, got false")
	}
}

func TestEvaluateFlagRepeatable(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "f", Enabled: true, RolloutPercent: 50})

	first := ""
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/flags/f/evaluate?user=alice", nil)
		req.SetPathValue("key", "f")
		rr := httptest.NewRecorder()
		EvaluateFlag(s).ServeHTTP(rr, req)
		if first == "" {
			first = rr.Body.String()
		} else if rr.Body.String() != first {
			t.Fatalf("non-deterministic response: %s then %s", first, rr.Body.String())
		}
	}
}

func TestEvaluateFlagMissingUser(t *testing.T) {
	s := store.New()
	seedFlag(t, s, store.Flag{Key: "f", Enabled: true, RolloutPercent: 100})

	req := httptest.NewRequest(http.MethodGet, "/flags/f/evaluate", nil)
	req.SetPathValue("key", "f")
	rr := httptest.NewRecorder()

	EvaluateFlag(s).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rr.Code)
	}
}

func TestEvaluateFlagUnknownKey(t *testing.T) {
	s := store.New()

	req := httptest.NewRequest(http.MethodGet, "/flags/missing/evaluate?user=alice", nil)
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()

	EvaluateFlag(s).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
