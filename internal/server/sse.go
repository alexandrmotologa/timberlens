package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/buffer"
	"github.com/alexandrmotologa/timberlens/internal/model"
)

type clientConn struct {
	ch    chan *model.LogRecord
	query *buffer.Query
}

// SSEBroker handles broadcasting log records to connected web browsers via Server-Sent Events.
type SSEBroker struct {
	mu      sync.RWMutex
	clients map[*clientConn]struct{}
}

// NewSSEBroker initializes an SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[*clientConn]struct{}),
	}
}

// Broadcast sends a new log record to all connected clients non-blockingly.
func (b *SSEBroker) Broadcast(rec *model.LogRecord) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		if client.query != nil && !client.query.Match(rec) {
			continue
		}

		select {
		case client.ch <- rec:
		default:
			// Client channel full, drop frame to maintain backpressure safety
		}
	}
}

// ServeHTTP handles the SSE streaming connection.
func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	queryStr := r.URL.Query().Get("q")
	var q *buffer.Query
	if queryStr != "" {
		q = buffer.ParseQuery(queryStr)
	}

	client := &clientConn{
		ch:    make(chan *model.LogRecord, 256),
		query: q,
	}

	b.mu.Lock()
	b.clients[client] = struct{}{}
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, client)
		b.mu.Unlock()
	}()

	// Send initial connected event
	fmt.Fprintf(w, "event: ready\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		case rec := <-client.ch:
			data, err := json.Marshal(rec)
			if err == nil {
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}
