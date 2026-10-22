package cluster

import (
	"testing"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestSignatureTrackerClustering(t *testing.T) {
	st := NewSignatureTracker(50)

	// Ingest 3 instances of similar error with different IDs
	rec1 := &model.LogRecord{
		ID:        1,
		Level:     model.LevelError,
		Service:   "payment-api",
		Message:   "Payment failed for user 88319 and transaction 4a5e3f89-9b93-41c3-8f6a-6869a65a31e2",
		Timestamp: time.Now(),
	}
	rec2 := &model.LogRecord{
		ID:        2,
		Level:     model.LevelError,
		Service:   "payment-api",
		Message:   "Payment failed for user 99120 and transaction eb4b3834-3c6f-4d6f-87d3-9bc0a1a0d84f",
		Timestamp: time.Now(),
	}
	rec3 := &model.LogRecord{
		ID:        3,
		Level:     model.LevelError,
		Service:   "payment-api",
		Message:   "Payment failed for user 10223 and transaction 00112233-4455-6677-8899-aabbccddeeff",
		Timestamp: time.Now(),
	}

	st.Record(rec1)
	st.Record(rec2)
	st.Record(rec3)

	clusters := st.TopClusters(10)
	if len(clusters) != 1 {
		t.Fatalf("Expected 1 grouped cluster, got %d", len(clusters))
	}

	if clusters[0].Count != 3 {
		t.Errorf("Expected cluster count 3, got %d", clusters[0].Count)
	}

	if clusters[0].Pattern != "Payment failed for user {num} and transaction {uuid}" {
		t.Errorf("Unexpected pattern: %q", clusters[0].Pattern)
	}
}
