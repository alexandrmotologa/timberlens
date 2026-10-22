package cluster

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

var (
	uuidRegex   = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	hexRegex    = regexp.MustCompile(`0x[0-9a-fA-F]+`)
	ipRegex     = regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(?::\d+)?\b`)
	digitsRegex = regexp.MustCompile(`\b\d+\b`)
)

// ErrorCluster represents a group of recurring errors sharing the same normalized signature.
type ErrorCluster struct {
	Signature   string           `json:"signature"`
	Pattern     string           `json:"pattern"`
	RootCause   string           `json:"rootCause,omitempty"`
	Count       int              `json:"count"`
	FirstSeen   time.Time        `json:"firstSeen"`
	LastSeen    time.Time        `json:"lastSeen"`
	Service     string           `json:"service"`
	Sample      *model.LogRecord `json:"sample"`
}

// SignatureTracker groups recurring exceptions.
type SignatureTracker struct {
	mu       sync.RWMutex
	clusters map[string]*ErrorCluster
	maxSize  int
}

// NewSignatureTracker creates a new error signature tracker.
func NewSignatureTracker(maxSize int) *SignatureTracker {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &SignatureTracker{
		clusters: make(map[string]*ErrorCluster),
		maxSize:  maxSize,
	}
}

// Record ingests an error record into the clustering engine if level is WARN, ERROR, or FATAL.
func (st *SignatureTracker) Record(rec *model.LogRecord) {
	if rec.Level < model.LevelWarn {
		return
	}

	st.mu.Lock()
	defer st.mu.Unlock()

	normPattern := normalizeMessage(rec.Message)
	rootCause := ""
	if rec.StackTrace != nil && rec.StackTrace.RootCause != "" {
		rootCause = rec.StackTrace.RootCause
	}

	hashInput := rec.Service + "|" + normPattern + "|" + rootCause
	hasher := sha256.New()
	hasher.Write([]byte(hashInput))
	sig := hex.EncodeToString(hasher.Sum(nil))[:16]

	if cluster, exists := st.clusters[sig]; exists {
		cluster.Count++
		cluster.LastSeen = rec.Timestamp
		cluster.Sample = rec
		return
	}

	if len(st.clusters) >= st.maxSize {
		// Drop least recently updated if capacity reached
		var oldestKey string
		var oldestTime time.Time
		for k, c := range st.clusters {
			if oldestKey == "" || c.LastSeen.Before(oldestTime) {
				oldestKey = k
				oldestTime = c.LastSeen
			}
		}
		delete(st.clusters, oldestKey)
	}

	st.clusters[sig] = &ErrorCluster{
		Signature: sig,
		Pattern:   normPattern,
		RootCause: rootCause,
		Count:     1,
		FirstSeen: rec.Timestamp,
		LastSeen:  rec.Timestamp,
		Service:   rec.Service,
		Sample:    rec,
	}
}

// TopClusters returns clusters sorted by highest occurrence count.
func (st *SignatureTracker) TopClusters(limit int) []*ErrorCluster {
	st.mu.RLock()
	defer st.mu.RUnlock()

	list := make([]*ErrorCluster, 0, len(st.clusters))
	for _, c := range st.clusters {
		list = append(list, c)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})

	if limit > 0 && len(list) > limit {
		list = list[:limit]
	}

	return list
}

func normalizeMessage(msg string) string {
	s := uuidRegex.ReplaceAllString(msg, "{uuid}")
	s = hexRegex.ReplaceAllString(s, "{addr}")
	s = ipRegex.ReplaceAllString(s, "{ip}")
	s = digitsRegex.ReplaceAllString(s, "{num}")
	return strings.TrimSpace(s)
}
