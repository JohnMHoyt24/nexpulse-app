package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

//Pinger is anything that can check database connectivity.
//*pgxpool.Pool satsifies it, and so can a fake in tests.

type Pinger interface {
	Ping(ctx context.Context) error
}

type Server struct {
	db Pinger
}

// New builds the router with every API route registered.
func New(db Pinger) http.Handler {
	s := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	return mux
}

type healthResponse struct {
	Status   string `json:"status"`
	Service  string `json:"service"`
	Database string `json:"database"`
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	resp := healthResponse{Status: "ok", Service: "nexpulse", Database: "ok"}
	status := http.StatusOK

	if err := s.db.Ping(ctx); err != nil {
		log.Printf("health check: database ping failed: %v", err)
		resp.Status = "degraded"
		resp.Database = "unreachable"
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, resp)
}

// writeJSON sends v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
