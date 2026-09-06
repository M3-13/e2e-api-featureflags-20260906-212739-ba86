package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"e2e-api-featureflags/internal/store"
)

const maxBodyBytes = 1 << 20 // 1 MB

// decodeBody limits the request body and decodes it as JSON. On failure it
// writes a 400 error and returns false.
func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusBadRequest, "request body too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
		}
		return false
	}
	return true
}

func validRollout(p int) bool {
	return p >= 0 && p <= 100
}

func CreateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var f store.Flag
		if !decodeBody(w, r, &f) {
			return
		}
		if f.Key == "" {
			writeError(w, http.StatusBadRequest, "key must not be empty")
			return
		}
		if !validRollout(f.RolloutPercent) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		if err := s.Create(f); err != nil {
			if errors.Is(err, store.ErrDuplicate) {
				writeError(w, http.StatusConflict, "flag already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		writeJSON(w, http.StatusCreated, f)
	}
}

func ListFlags(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, s.List())
	}
}

func GetFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		f, ok := s.Get(key)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, f)
	}
}

func UpdateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		var f store.Flag
		if !decodeBody(w, r, &f) {
			return
		}
		if !validRollout(f.RolloutPercent) {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		f.Key = key
		updated, ok := s.Update(key, f)
		if !ok {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		writeJSON(w, http.StatusOK, updated)
	}
}

func DeleteFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")
		if !s.Delete(key) {
			writeError(w, http.StatusNotFound, "flag not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
