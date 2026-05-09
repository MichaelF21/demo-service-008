package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRootHandler_ReturnsServiceGreeting(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	rootHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content-type: want application/json, got %q", got)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["service"] != serviceName {
		t.Errorf("service field: want %q, got %q", serviceName, body["service"])
	}
	if !strings.Contains(body["message"], serviceName) {
		t.Errorf("message: want contains %q, got %q", serviceName, body["message"])
	}
}

func TestHealthHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Errorf("body: want contains 'ok', got %q", rec.Body.String())
	}
}

func TestReadyHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	readyHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d", rec.Code)
	}
}

func TestNewMux_AllRoutesRegistered(t *testing.T) {
	mux := newMux()
	for _, path := range []string{"/", "/healthz", "/readyz", "/metrics"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code >= 500 {
			t.Errorf("%s: server error %d", path, rec.Code)
		}
	}
}
