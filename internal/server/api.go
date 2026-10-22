package server

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/timberlens/internal/buffer"
	"github.com/alexandrmotologa/timberlens/internal/cluster"
)

// Server provides the HTTP API and SSE broker.
type Server struct {
	ring      *buffer.RingBuffer
	broker    *SSEBroker
	clusters  *cluster.SignatureTracker
	histogram *buffer.TimeHistogram
	mux       *http.ServeMux
}

// NewServer creates a new API server instance.
func NewServer(ring *buffer.RingBuffer, broker *SSEBroker, clusters *cluster.SignatureTracker, histogram *buffer.TimeHistogram) *Server {
	s := &Server{
		ring:      ring,
		broker:    broker,
		clusters:  clusters,
		histogram: histogram,
		mux:       http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("/api/stream", s.broker.ServeHTTP)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/services", s.handleServices)
	s.mux.HandleFunc("/api/traces/", s.handleTraceByID)
	s.mux.HandleFunc("/api/clusters", s.handleClusters)
	s.mux.HandleFunc("/api/histogram", s.handleHistogram)
}

func (s *Server) handleHistogram(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if s.histogram == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	buckets := s.histogram.RecentBuckets()
	writeJSON(w, http.StatusOK, buckets)
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	qStr := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")

	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 1000 {
		limit = 1000
	}

	var q *buffer.Query
	if qStr != "" {
		q = buffer.ParseQuery(qStr)
	}

	records := s.ring.Filter(q, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"total":   s.ring.Total(),
		"count":   len(records),
		"records": records,
	})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	writeJSON(w, http.StatusOK, map[string]any{
		"totalIngested": s.ring.Total(),
		"capacity":      s.ring.Capacity(),
		"bufferedCount": s.ring.Len(),
		"memoryBytes":   mem.Alloc,
		"memoryMb":      float64(mem.Alloc) / 1024 / 1024,
		"services":      s.ring.ServiceCounts(),
		"levels":        s.ring.LevelCounts(),
	})
}

func (s *Server) handleServices(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	svcCounts := s.ring.ServiceCounts()

	type svcInfo struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	list := make([]svcInfo, 0, len(svcCounts))
	for name, count := range svcCounts {
		list = append(list, svcInfo{Name: name, Count: count})
	}

	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleTraceByID(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	traceID := strings.TrimPrefix(r.URL.Path, "/api/traces/")
	traceID = strings.TrimSpace(traceID)

	if traceID == "" {
		http.Error(w, "Missing trace identifier", http.StatusBadRequest)
		return
	}

	records := s.ring.ByTraceID(traceID)
	writeJSON(w, http.StatusOK, map[string]any{
		"traceId": traceID,
		"count":   len(records),
		"records": records,
	})
}

func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	limit := 20
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	clusters := s.clusters.TopClusters(limit)
	writeJSON(w, http.StatusOK, clusters)
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
