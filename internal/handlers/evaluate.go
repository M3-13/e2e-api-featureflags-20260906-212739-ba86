package handlers

import (
	"net/http"

	"e2e-api-featureflags/internal/store"
)

func EvaluateFlag(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotImplemented, "not implemented")
	}
}
