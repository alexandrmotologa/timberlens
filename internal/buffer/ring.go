package buffer

import (
	"sync"
	"sync/atomic"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// RingBuffer is a thread-safe circular memory buffer for log records.
type RingBuffer struct {
	capacity uint64
	records  []*model.LogRecord
	head     uint64 // Total number of pushed items
	mu       sync.RWMutex

	// Aggregate statistics
	services    map[string]int
	levelCounts [7]int
}

// NewRingBuffer allocates a new ring buffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 500000
	}
	return &RingBuffer{
		capacity: uint64(capacity),
		records:  make([]*model.LogRecord, capacity),
		services: make(map[string]int),
	}
}

// Push adds a record to the circular buffer, overwriting the oldest entry if full.
func (rb *RingBuffer) Push(rec *model.LogRecord) uint64 {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	idx := rb.head % rb.capacity
	old := rb.records[idx]

	// If overwriting an existing record, decrement previous stats
	if old != nil {
		if count, ok := rb.services[old.Service]; ok && count > 0 {
			rb.services[old.Service] = count - 1
			if rb.services[old.Service] == 0 {
				delete(rb.services, old.Service)
			}
		}
		if int(old.Level) < len(rb.levelCounts) && rb.levelCounts[old.Level] > 0 {
			rb.levelCounts[old.Level]--
		}
	}

	rb.records[idx] = rec
	rb.head++

	// Increment new stats
	if rec.Service != "" {
		rb.services[rec.Service]++
	}
	if int(rec.Level) < len(rb.levelCounts) {
		rb.levelCounts[rec.Level]++
	}

	return rb.head
}

// Total returns total number of records ingested since startup.
func (rb *RingBuffer) Total() uint64 {
	return atomic.LoadUint64(&rb.head)
}

// Len returns the current number of valid records held in memory.
func (rb *RingBuffer) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.head < rb.capacity {
		return int(rb.head)
	}
	return int(rb.capacity)
}

// Capacity returns the maximum buffer capacity.
func (rb *RingBuffer) Capacity() int {
	return int(rb.capacity)
}

// Recent returns up to `limit` latest records in chronological order.
func (rb *RingBuffer) Recent(limit int) []*model.LogRecord {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	total := rb.head
	if total == 0 {
		return nil
	}

	available := total
	if available > rb.capacity {
		available = rb.capacity
	}

	count := uint64(limit)
	if count > available || count <= 0 {
		count = available
	}

	result := make([]*model.LogRecord, count)
	startSeq := total - count

	for i := uint64(0); i < count; i++ {
		seq := startSeq + i
		idx := seq % rb.capacity
		result[i] = rb.records[idx]
	}

	return result
}

// GetByID finds a record by its sequence ID if it still exists in the ring buffer.
func (rb *RingBuffer) GetByID(id uint64) *model.LogRecord {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.head == 0 || id == 0 || id > rb.head {
		return nil
	}

	// Check if already overwritten
	if rb.head > rb.capacity && id <= (rb.head-rb.capacity) {
		return nil
	}

	idx := (id - 1) % rb.capacity
	rec := rb.records[idx]
	if rec != nil && rec.ID == id {
		return rec
	}
	return nil
}

// ServiceCounts returns a snapshot of active log counts per service.
func (rb *RingBuffer) ServiceCounts() map[string]int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make(map[string]int, len(rb.services))
	for k, v := range rb.services {
		result[k] = v
	}
	return result
}

// LevelCounts returns a snapshot of active counts per severity level.
func (rb *RingBuffer) LevelCounts() map[string]int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	return map[string]int{
		"TRACE": rb.levelCounts[model.LevelTrace],
		"DEBUG": rb.levelCounts[model.LevelDebug],
		"INFO":  rb.levelCounts[model.LevelInfo],
		"WARN":  rb.levelCounts[model.LevelWarn],
		"ERROR": rb.levelCounts[model.LevelError],
		"FATAL": rb.levelCounts[model.LevelFatal],
	}
}

// Filter scans recent records in reverse chronological order and returns those matching the query.
func (rb *RingBuffer) Filter(q *Query, limit int) []*model.LogRecord {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	total := rb.head
	if total == 0 {
		return nil
	}

	available := total
	if available > rb.capacity {
		available = rb.capacity
	}

	if limit <= 0 || limit > int(available) {
		limit = int(available)
	}

	matches := make([]*model.LogRecord, 0, limit)

	// Scan from most recent to oldest
	for i := uint64(0); i < available; i++ {
		seq := total - 1 - i
		idx := seq % rb.capacity
		rec := rb.records[idx]
		if rec == nil {
			continue
		}

		if q == nil || q.Match(rec) {
			matches = append(matches, rec)
			if len(matches) >= limit {
				break
			}
		}
	}

	// Reverse to return in chronological order
	for i, j := 0, len(matches)-1; i < j; i, j = i+1, j-1 {
		matches[i], matches[j] = matches[j], matches[i]
	}

	return matches
}

// ByTraceID returns all records associated with the given trace identifier in chronological order.
func (rb *RingBuffer) ByTraceID(traceID string) []*model.LogRecord {
	if traceID == "" {
		return nil
	}
	q := &Query{
		Criteria: []Criterion{
			{Field: "trace_id", Operator: OpEqual, StrVal: traceID},
		},
	}
	return rb.Filter(q, 1000)
}

