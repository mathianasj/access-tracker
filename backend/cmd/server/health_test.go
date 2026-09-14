package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockDBPingable struct {
	pingError error
}

func (m *mockDBPingable) Ping(ctx context.Context) error {
	return m.pingError
}

func TestHealthHandlerHealthy(t *testing.T) {
	db := &mockDBPingable{pingError: nil}
	
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	
	handler := healthHandlerWrapper(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", response["status"])
	}

	if response["timestamp"] == "" {
		t.Error("expected timestamp to be set")
	}

	if _, err := time.Parse(time.RFC3339, response["timestamp"]); err != nil {
		t.Errorf("timestamp not in RFC3339 format: %v", err)
	}
}

func TestHealthHandlerUnhealthy(t *testing.T) {
	db := &mockDBPingable{pingError: context.DeadlineExceeded}
	
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	
	handler := healthHandlerWrapper(db)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "unhealthy" {
		t.Errorf("expected status 'unhealthy', got '%s'", response["status"])
	}
}

func TestHealthHandlerContentType(t *testing.T) {
	db := &mockDBPingable{pingError: nil}
	
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	
	handler := healthHandlerWrapper(db)
	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func TestResponseWriter(t *testing.T) {
	w := httptest.NewRecorder()

	rw := &responseWriter{ResponseWriter: w, status: 0}

	if rw.status != 0 {
		t.Errorf("expected initial status 0, got %d", rw.status)
	}

	rw.WriteHeader(http.StatusOK)
	if rw.status != http.StatusOK {
		t.Errorf("expected status 200, got %d", rw.status)
	}

	n, err := rw.Write([]byte("test"))
	if n != 4 {
		t.Errorf("expected 4 bytes written, got %d", n)
	}
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func healthHandlerWrapper(db Pingable) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		response := struct {
			Status    string `json:"status"`
			Timestamp string `json:"timestamp"`
		}{
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if err := db.Ping(ctx); err != nil {
			response.Status = "unhealthy"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(response)
			return
		}

		response.Status = "healthy"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

type Pingable interface {
	Ping(ctx context.Context) error
}
