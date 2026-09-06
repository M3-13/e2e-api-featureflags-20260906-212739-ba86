package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"e2e-api-featureflags/internal/store"
)

func newStore() *store.Store {
	return store.New()
}

func assertErrorBody(t *testing.T, rr *httptest.ResponseRecorder) {
	t.Helper()
	var got map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("error body is not valid JSON: %v (body=%q)", err, rr.Body.String())
	}
	if _, ok := got["error"]; !ok {
		t.Fatalf("error body missing 'error' key: %v", got)
	}
	if len(got) != 1 {
		t.Fatalf("error body has unexpected keys: %v", got)
	}
}

func TestCreateFlag201(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	body := `{"key":"myflag","enabled":true,"description":"desc","rollout_percent":50}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d want %d (body=%s)", rr.Code, http.StatusCreated, rr.Body.String())
	}
	var got store.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Key != "myflag" || !got.Enabled || got.Description != "desc" || got.RolloutPercent != 50 {
		t.Fatalf("body: %+v", got)
	}
}

func TestCreateFlagDefaults(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"myflag"}`))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d", rr.Code)
	}
	var got store.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Enabled || got.Description != "" || got.RolloutPercent != 0 {
		t.Fatalf("defaults: %+v", got)
	}
}

func TestCreateFlagEmptyKey(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":""}`))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestCreateFlagInvalidJSON(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{invalid`))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestCreateFlagInvalidKeyCharacters(t *testing.T) {
	for _, key := range []string{"bad key", "key!", "key:1", "key/1", "ke@y", "key[]"} {
		s := newStore()
		h := CreateFlag(s)
		body := `{"key":` + jsonString(key) + `}`
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
		rr := httptest.NewRecorder()
		h(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("key=%q status: got %d want 400", key, rr.Code)
		}
		assertErrorBody(t, rr)
	}
}

func TestCreateFlagKeyTooLong(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	key := strings.Repeat("a", 129)
	body := `{"key":` + jsonString(key) + `}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestCreateFlagKeyMaxLengthOK(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	key := strings.Repeat("a", 128)
	body := `{"key":` + jsonString(key) + `}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status: got %d want 201", rr.Code)
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestCreateFlagTooManyFlags(t *testing.T) {
	s := newStore()
	for i := 0; i < 1000; i++ {
		if err := s.Create(store.Flag{Key: fmt.Sprintf("flag-%d", i)}); err != nil {
			t.Fatalf("seed %d: %v", i, err)
		}
	}
	h := CreateFlag(s)
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"overflow"}`))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestUpdateFlagInvalidKeyCharacters(t *testing.T) {
	s := newStore()
	h := UpdateFlag(s)
	req := httptest.NewRequest(http.MethodPut, "/flags/bad!key", strings.NewReader(`{"enabled":true}`))
	req.SetPathValue("key", "bad!key")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestUpdateFlagKeyTooLong(t *testing.T) {
	s := newStore()
	h := UpdateFlag(s)
	key := strings.Repeat("a", 129)
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(`{"enabled":true}`))
	req.SetPathValue("key", key)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestCreateFlagRolloutOutOfRange(t *testing.T) {
	for _, p := range []int{-1, 101} {
		s := newStore()
		h := CreateFlag(s)
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(
			`{"key":"myflag","rollout_percent":`+jsonNumber(p)+`}`,
		))
		rr := httptest.NewRecorder()
		h(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent=%d status: got %d want 400", p, rr.Code)
		}
		assertErrorBody(t, rr)
	}
}

func jsonNumber(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func TestCreateFlagDuplicate(t *testing.T) {
	s := newStore()
	if err := s.Create(store.Flag{Key: "myflag"}); err != nil {
		t.Fatalf("seed: %v", err)
	}
	h := CreateFlag(s)
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(`{"key":"myflag"}`))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status: got %d want 409", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestListFlags(t *testing.T) {
	s := newStore()
	s.Create(store.Flag{Key: "a"})
	s.Create(store.Flag{Key: "b"})
	h := ListFlags(s)
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var got []store.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len: got %d want 2", len(got))
	}
	if got[0].Key != "a" || got[1].Key != "b" {
		t.Fatalf("order: %v", got)
	}
}

func TestListFlagsEmpty(t *testing.T) {
	s := newStore()
	h := ListFlags(s)
	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("empty list body: got %q want []", rr.Body.String())
	}
}

func TestGetFlag(t *testing.T) {
	s := newStore()
	s.Create(store.Flag{Key: "myflag", Enabled: true})
	h := GetFlag(s)
	req := httptest.NewRequest(http.MethodGet, "/flags/myflag", nil)
	req.SetPathValue("key", "myflag")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var got store.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Key != "myflag" || !got.Enabled {
		t.Fatalf("body: %+v", got)
	}
}

func TestGetFlagNotFound(t *testing.T) {
	s := newStore()
	h := GetFlag(s)
	req := httptest.NewRequest(http.MethodGet, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestUpdateFlag(t *testing.T) {
	s := newStore()
	s.Create(store.Flag{Key: "myflag", Enabled: false})
	h := UpdateFlag(s)
	body := `{"enabled":true,"description":"updated","rollout_percent":80}`
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(body))
	req.SetPathValue("key", "myflag")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status: got %d", rr.Code)
	}
	var got store.Flag
	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Key != "myflag" || !got.Enabled || got.Description != "updated" || got.RolloutPercent != 80 {
		t.Fatalf("body: %+v", got)
	}
}

func TestUpdateFlagNotFound(t *testing.T) {
	s := newStore()
	h := UpdateFlag(s)
	req := httptest.NewRequest(http.MethodPut, "/flags/missing", strings.NewReader(`{"enabled":true}`))
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestUpdateFlagRolloutInvalid(t *testing.T) {
	s := newStore()
	s.Create(store.Flag{Key: "myflag"})
	h := UpdateFlag(s)
	req := httptest.NewRequest(http.MethodPut, "/flags/myflag", strings.NewReader(`{"rollout_percent":101}`))
	req.SetPathValue("key", "myflag")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestDeleteFlag(t *testing.T) {
	s := newStore()
	s.Create(store.Flag{Key: "myflag"})
	h := DeleteFlag(s)
	req := httptest.NewRequest(http.MethodDelete, "/flags/myflag", nil)
	req.SetPathValue("key", "myflag")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status: got %d want 204", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("delete body: got %q want empty", rr.Body.String())
	}
	if _, ok := s.Get("myflag"); ok {
		t.Fatal("Delete: flag still exists")
	}
}

func TestDeleteFlagNotFound(t *testing.T) {
	s := newStore()
	h := DeleteFlag(s)
	req := httptest.NewRequest(http.MethodDelete, "/flags/missing", nil)
	req.SetPathValue("key", "missing")
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status: got %d want 404", rr.Code)
	}
	assertErrorBody(t, rr)
}

func TestCreateFlagBodyTooLarge(t *testing.T) {
	s := newStore()
	h := CreateFlag(s)
	big := strings.Repeat("a", maxBodyBytes+1)
	body := `{"key":"x","description":"` + big + `"}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d want 400", rr.Code)
	}
	assertErrorBody(t, rr)
}
