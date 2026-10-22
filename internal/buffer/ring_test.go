package buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestRingBufferPushAndWrap(t *testing.T) {
	rb := NewRingBuffer(5)

	for i := 1; i <= 10; i++ {
		rb.Push(&model.LogRecord{
			ID:        uint64(i),
			Service:   "test-svc",
			Level:     model.LevelInfo,
			Message:   fmt.Sprintf("msg-%d", i),
			Timestamp: time.Now(),
		})
	}

	if rb.Total() != 10 {
		t.Errorf("Expected Total 10, got %d", rb.Total())
	}
	if rb.Len() != 5 {
		t.Errorf("Expected Len 5, got %d", rb.Len())
	}

	recent := rb.Recent(5)
	if len(recent) != 5 {
		t.Fatalf("Expected 5 recent records, got %d", len(recent))
	}
	if recent[0].ID != 6 || recent[4].ID != 10 {
		t.Errorf("Expected IDs 6..10, got first=%d, last=%d", recent[0].ID, recent[4].ID)
	}

	// Record 1 should be overwritten
	if rb.GetByID(1) != nil {
		t.Errorf("Expected record 1 to be overwritten and return nil")
	}
	// Record 8 should still exist
	if rb.GetByID(8) == nil {
		t.Errorf("Expected record 8 to exist")
	}
}

func TestQueryFiltering(t *testing.T) {
	rb := NewRingBuffer(100)

	rb.Push(&model.LogRecord{
		ID:      1,
		Service: "auth",
		Level:   model.LevelInfo,
		Message: "user logged in",
		Raw:     "user logged in",
		TraceID: "trc-1",
		Numeric: map[string]float64{"latency_ms": 45.0},
	})

	rb.Push(&model.LogRecord{
		ID:      2,
		Service: "api",
		Level:   model.LevelError,
		Message: "database connection timeout",
		Raw:     "database connection timeout",
		TraceID: "trc-1",
		Numeric: map[string]float64{"latency_ms": 650.0},
	})

	rb.Push(&model.LogRecord{
		ID:      3,
		Service: "worker",
		Level:   model.LevelWarn,
		Message: "queue backlog high",
		Raw:     "queue backlog high",
		TraceID: "trc-2",
		Numeric: map[string]float64{"latency_ms": 120.0},
	})

	// Test 1: service:api
	q1 := ParseQuery("service:api")
	res1 := rb.Filter(q1, 10)
	if len(res1) != 1 || res1[0].ID != 2 {
		t.Fatalf("Expected 1 match for service:api, got %d", len(res1))
	}

	// Test 2: level:error
	q2 := ParseQuery("level:error")
	res2 := rb.Filter(q2, 10)
	if len(res2) != 1 || res2[0].ID != 2 {
		t.Fatalf("Expected 1 match for level:error, got %d", len(res2))
	}

	// Test 3: level>=warn
	q3 := ParseQuery("level>=warn")
	res3 := rb.Filter(q3, 10)
	if len(res3) != 2 {
		t.Fatalf("Expected 2 matches for level>=warn, got %d", len(res3))
	}

	// Test 4: latency_ms>500
	q4 := ParseQuery("latency_ms>500")
	res4 := rb.Filter(q4, 10)
	if len(res4) != 1 || res4[0].ID != 2 {
		t.Fatalf("Expected 1 match for latency_ms>500, got %d", len(res4))
	}

	// Test 5: Full-text substring
	q5 := ParseQuery("backlog")
	res5 := rb.Filter(q5, 10)
	if len(res5) != 1 || res5[0].ID != 3 {
		t.Fatalf("Expected 1 match for 'backlog', got %d", len(res5))
	}

	// Test 6: Trace correlation
	traceMatches := rb.ByTraceID("trc-1")
	if len(traceMatches) != 2 {
		t.Fatalf("Expected 2 records for trace trc-1, got %d", len(traceMatches))
	}
}
