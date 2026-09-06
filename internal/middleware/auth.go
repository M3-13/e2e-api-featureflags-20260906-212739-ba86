package middleware

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

func Auth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				writeUnauthorized(w)
				return
			}

			provided, ok := bearerToken(r)
			if !ok || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				writeUnauthorized(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", false
	}
	return strings.TrimPrefix(auth, prefix), true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
