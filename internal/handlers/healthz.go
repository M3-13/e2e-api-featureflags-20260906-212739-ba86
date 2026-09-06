package handlers

import (
	"net/http"

	"e2e-api-featureflags/internal/version"
)

func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version.Version})
	}
}
