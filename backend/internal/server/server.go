// Package server holds the HTTP layer of the Nexpulse API: the router,
// the request handlers, and small helpers for writing responses.
package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Pinger is anything that can check database connectivity.
// *pgxpool.Pool satisfies it, and so can a fake in tests.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server holds the dependencies that the HTTP handlers need.
type Server struct {
	db Pinger
}

// New builds the router with every API route registered.
//
// Routes use Go 1.22+ ServeMux patterns ("METHOD /path"), so a request with
// the wrong method gets 405 Method Not Allowed and an unknown path gets 404
// without any extra code.
func New(db Pinger) http.Handler {
	s := &Server{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	return mux
}

// healthResponse is the JSON body returned by GET /healthz.
type healthResponse struct {
	Status   string `json:"status"`   // "ok" or "degraded"
	Service  string `json:"service"`  // always "nexpulse"
	Database string `json:"database"` // "ok" or "unreachable"
}

// healthHandler reports whether the API and its database are up.
//
// It responds 200 when the database answers a ping, and 503 Service
// Unavailable with status "degraded" when it does not, so load balancers and
// uptime checkers can tell the instance is unhealthy.
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	// Cap the ping so a hung database can't make the health check hang too.
	// Deriving from r.Context() also stops the ping if the client disconnects.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// Assume healthy, then downgrade if the ping fails.
	resp := healthResponse{Status: "ok", Service: "nexpulse", Database: "ok"}
	status := http.StatusOK

	if err := s.db.Ping(ctx); err != nil {
		// Log the real error for operators, but keep it out of the public
		// response body.
		log.Printf("health check: database ping failed: %v", err)
		resp.Status = "degraded"
		resp.Database = "unreachable"
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, resp)
}

// writeJSON sends v as a JSON response with the given status code.
//
// Headers must be set before WriteHeader is called, and once WriteHeader has
// run the status can't be changed, so an encoding error can only be logged.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}
