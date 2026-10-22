package buffer

import (
	"sync"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// HistogramBucket represents aggregated log traffic in a single time slice.
type HistogramBucket struct {
	Timestamp int64 `json:"timestamp"` // Unix timestamp in seconds
	Total     int   `json:"total"`
	Info      int   `json:"info"`
	Warn      int   `json:"warn"`
	Error     int   `json:"error"`
}

// TimeHistogram maintains a rolling window of time-bucketed counts.
type TimeHistogram struct {
	mu         sync.RWMutex
	bucketSec  int64
	numBuckets int
	buckets    map[int64]*HistogramBucket
}

// NewTimeHistogram creates a rolling histogram.
func NewTimeHistogram(bucketSec, numBuckets int) *TimeHistogram {
	if bucketSec <= 0 {
		bucketSec = 1
	}
	if numBuckets <= 0 {
		numBuckets = 60
	}
	return &TimeHistogram{
		bucketSec:  int64(bucketSec),
		numBuckets: numBuckets,
		buckets:    make(map[int64]*HistogramBucket, numBuckets),
	}
}

// Record increments the counter for the record's timestamp.
func (th *TimeHistogram) Record(rec *model.LogRecord) {
	if rec == nil {
		return
	}

	sec := rec.Timestamp.Unix()
	bucketKey := (sec / th.bucketSec) * th.bucketSec

	th.mu.Lock()
	defer th.mu.Unlock()

	b, exists := th.buckets[bucketKey]
	if !exists {
		b = &HistogramBucket{Timestamp: bucketKey}
		th.buckets[bucketKey] = b

		// Prune older than retention window
		cutoff := bucketKey - (int64(th.numBuckets) * th.bucketSec)
		for k := range th.buckets {
			if k < cutoff {
				delete(th.buckets, k)
			}
		}
	}

	b.Total++
	switch {
	case rec.Level >= model.LevelError:
		b.Error++
	case rec.Level == model.LevelWarn:
		b.Warn++
	default:
		b.Info++
	}
}

// RecentBuckets returns an ordered list of buckets for the requested window.
func (th *TimeHistogram) RecentBuckets() []HistogramBucket {
	th.mu.RLock()
	defer th.mu.RUnlock()

	now := time.Now().Unix()
	latestKey := (now / th.bucketSec) * th.bucketSec
	startKey := latestKey - (int64(th.numBuckets-1) * th.bucketSec)

	result := make([]HistogramBucket, th.numBuckets)
	for i := 0; i < th.numBuckets; i++ {
		ts := startKey + int64(i)*th.bucketSec
		if b, ok := th.buckets[ts]; ok {
			result[i] = *b
		} else {
			result[i] = HistogramBucket{Timestamp: ts}
		}
	}

	return result
}
