package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/buffer"
	"github.com/alexandrmotologa/timberlens/internal/cluster"
	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestAPIRoutes(t *testing.T) {
	ring := buffer.NewRingBuffer(100)
	broker := NewSSEBroker()
	clusters := cluster.NewSignatureTracker(100)
	srv := NewServer(ring, broker, clusters)

	// Ingest sample records
	rec1 := &model.LogRecord{
		ID:        1,
		Level:     model.LevelInfo,
		Service:   "auth-api",
		Message:   "token generated",
		Raw:       "token generated",
		TraceID:   "trc-xyz",
		Timestamp: time.Now(),
	}
	rec2 := &model.LogRecord{
		ID:        2,
		Level:     model.LevelError,
		Service:   "payment-api",
		Message:   "card declined",
		Raw:       "card declined",
		TraceID:   "trc-xyz",
		Timestamp: time.Now(),
	}
	ring.Push(rec1)
	ring.Push(rec2)
	clusters.Record(rec2)

	// 1. Test /api/logs
	req1 := httptest.NewRequest(http.MethodGet, "/api/logs?q=service:auth-api", nil)
	w1 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from /api/logs, got %d", w1.Code)
	}

	var logsResp struct {
		Count   int                `json:"count"`
		Records []*model.LogRecord `json:"records"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &logsResp); err != nil {
		t.Fatalf("Failed to parse logs response: %v", err)
	}
	if logsResp.Count != 1 || logsResp.Records[0].Service != "auth-api" {
		t.Errorf("Expected 1 match for auth-api, got %d", logsResp.Count)
	}

	// 2. Test /api/stats
	req2 := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	w2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from /api/stats, got %d", w2.Code)
	}

	// 3. Test /api/traces/trc-xyz
	req3 := httptest.NewRequest(http.MethodGet, "/api/traces/trc-xyz", nil)
	w3 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w3, req3)

	if w3.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK from /api/traces/, got %d", w3.Code)
	}

	var traceResp struct {
		Count   int                `json:"count"`
		Records []*model.LogRecord `json:"records"`
	}
	if err := json.Unmarshal(w3.Body.Bytes(), &traceResp); err != nil {
		t.Fatalf("Failed to parse trace response: %v", err)
	}
	if traceResp.Count != 2 {
		t.Errorf("Expected 2 records for trace trc-xyz, got %d", traceResp.Count)
	}
}
