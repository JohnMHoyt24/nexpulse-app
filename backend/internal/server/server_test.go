package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// fakePinger lets tests decide whether the "database" is up.
// A nil err simulates a healthy database; any other value simulates a failure.
type fakePinger struct {
	err error
}

// Ping returns the error the test configured, without touching a real database.
func (f fakePinger) Ping(ctx context.Context) error {
	return f.err
}

// TestRoutes checks that the router sends requests to the right place and
// rejects the wrong method or an unknown path with the right status code.
func TestRoutes(t *testing.T) {
	handler := New(fakePinger{})

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"health check returns OK", http.MethodGet, "/healthz", http.StatusOK},
		{"wrong method is rejected", http.MethodPost, "/healthz", http.StatusMethodNotAllowed},
		{"unknown path returns 404", http.MethodGet, "/does-not-exist", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptest builds an in-memory request and records the response,
			// so the handler runs without opening a real network port.
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

// TestHealthCheck checks the /healthz status code and JSON body for both a
// reachable and an unreachable database.
func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name       string
		dbErr      error
		wantStatus int
		wantBody   healthResponse
	}{
		{
			name:       "database reachable",
			dbErr:      nil,
			wantStatus: http.StatusOK,
			wantBody:   healthResponse{Status: "ok", Service: "nexpulse", Database: "ok"},
		},
		{
			name:       "database unreachable",
			dbErr:      errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   healthResponse{Status: "degraded", Service: "nexpulse", Database: "unreachable"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()

			New(fakePinger{err: tt.dbErr}).ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var got healthResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("could not decode response body: %v", err)
			}
			if got != tt.wantBody {
				t.Errorf("body = %+v, want %+v", got, tt.wantBody)
			}
		})
	}
}
