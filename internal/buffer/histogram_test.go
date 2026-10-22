package buffer

import (
	"testing"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestTimeHistogramAggregation(t *testing.T) {
	hist := NewTimeHistogram(1, 60)

	now := time.Now()
	rec1 := &model.LogRecord{
		ID:        1,
		Level:     model.LevelInfo,
		Timestamp: now,
	}
	rec2 := &model.LogRecord{
		ID:        2,
		Level:     model.LevelError,
		Timestamp: now,
	}
	rec3 := &model.LogRecord{
		ID:        3,
		Level:     model.LevelWarn,
		Timestamp: now,
	}

	hist.Record(rec1)
	hist.Record(rec2)
	hist.Record(rec3)

	buckets := hist.RecentBuckets()
	if len(buckets) != 60 {
		t.Fatalf("Expected 60 buckets, got %d", len(buckets))
	}

	// Last bucket should contain our 3 records
	last := buckets[59]
	if last.Total != 3 {
		t.Errorf("Expected last bucket total 3, got %d", last.Total)
	}
	if last.Info != 1 || last.Warn != 1 || last.Error != 1 {
		t.Errorf("Expected 1 info, 1 warn, 1 error; got info=%d warn=%d error=%d", last.Info, last.Warn, last.Error)
	}
}
