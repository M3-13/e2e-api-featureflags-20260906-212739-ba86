package handlers

import (
	"errors"
	"net/http"

	"e2e-api-featureflags/internal/store"
)

func EvaluateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		user := r.URL.Query().Get("user")
		if user == "" {
			writeError(w, http.StatusBadRequest, "missing or empty user parameter")
			return
		}

		result, err := s.Evaluate(key, user)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				writeError(w, http.StatusNotFound, "flag not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]bool{"result": result})
	}
}
