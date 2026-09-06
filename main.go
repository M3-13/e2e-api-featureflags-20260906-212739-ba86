package main

import (
	"log"
	"net"
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

	bindAddr := os.Getenv("BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "127.0.0.1"
	}

	apiToken := os.Getenv("FLAG_API_TOKEN")

	s := store.New()

	auth := middleware.Auth(apiToken)

	mux := http.NewServeMux()
	mux.Handle("POST /flags", auth(handlers.CreateFlag(s)))
	mux.Handle("GET /flags", auth(handlers.ListFlags(s)))
	mux.Handle("GET /flags/{key}", auth(handlers.GetFlag(s)))
	mux.Handle("PUT /flags/{key}", auth(handlers.UpdateFlag(s)))
	mux.Handle("DELETE /flags/{key}", auth(handlers.DeleteFlag(s)))
	mux.Handle("GET /flags/{key}/evaluate", auth(handlers.EvaluateFlag(s)))
	mux.Handle("GET /healthz", handlers.Healthz())

	handler := middleware.Logging(mux)

	addr := net.JoinHostPort(bindAddr, port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
