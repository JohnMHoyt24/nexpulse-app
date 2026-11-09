package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	handler := New()

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
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestHealthResponseBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	New().ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var got healthResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}

	want := healthResponse{Status: "ok", Service: "nexpulse"}
	if got != want {
		t.Errorf("body = %+v, want %+v", got, want)
	}
}
