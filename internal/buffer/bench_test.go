package buffer

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func BenchmarkRingBufferPush(b *testing.B) {
	rb := NewRingBuffer(500000)
	rec := &model.LogRecord{
		ID:        1,
		Timestamp: time.Now(),
		Level:     model.LevelInfo,
		LevelStr:  "INFO",
		Service:   "api-gateway",
		Message:   "HTTP request 200 OK GET /api/v1/users",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Push(rec)
	}
}

func BenchmarkQueryEvaluation(b *testing.B) {
	rb := NewRingBuffer(100000)
	for i := 1; i <= 100000; i++ {
		lvl := model.LevelInfo
		if i%20 == 0 {
			lvl = model.LevelError
		}
		rb.Push(&model.LogRecord{
			ID:        uint64(i),
			Timestamp: time.Now(),
			Level:     lvl,
			LevelStr:  lvl.String(),
			Service:   fmt.Sprintf("svc-%d", i%5),
			Message:   fmt.Sprintf("Transaction %d processed with status", i),
			Raw:       fmt.Sprintf("Transaction %d processed with status", i),
			Numeric:   map[string]float64{"latency_ms": float64(i % 500)},
		})
	}

	q := ParseQuery("service:svc-1 level:error latency_ms>250")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.Filter(q, 100)
	}
}
