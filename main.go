package main

import (
	"log"
	"net/http"
	"os"

	"e2e-api-featureflags/internal/handlers"
	"e2e-api-featureflags/internal/middleware"
	"e2e-api-featureflags/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := store.New()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", handlers.CreateFlag(s))
	mux.HandleFunc("GET /flags", handlers.ListFlags(s))
	mux.HandleFunc("GET /flags/{key}", handlers.GetFlag(s))
	mux.HandleFunc("PUT /flags/{key}", handlers.UpdateFlag(s))
	mux.HandleFunc("DELETE /flags/{key}", handlers.DeleteFlag(s))
	mux.HandleFunc("GET /flags/{key}/evaluate", handlers.EvaluateFlag(s))
	mux.HandleFunc("GET /healthz", handlers.Healthz())

	handler := middleware.Logging(mux)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}
