package model

import (
	"sync"
)

// StringInterner maintains an in-memory string deduplication pool.
type StringInterner struct {
	mu    sync.RWMutex
	pool  map[string]string
	maxCap int
}

// Global default string interner.
var DefaultInterner = NewStringInterner(50000)

// NewStringInterner creates a new string interning pool with a maximum capacity.
func NewStringInterner(maxCap int) *StringInterner {
	if maxCap <= 0 {
		maxCap = 50000
	}
	return &StringInterner{
		pool:   make(map[string]string, 1024),
		maxCap: maxCap,
	}
}

// Intern returns a canonical interned version of s.
func (si *StringInterner) Intern(s string) string {
	if s == "" {
		return ""
	}
	// Fast read lock path
	si.mu.RLock()
	if canonical, ok := si.pool[s]; ok {
		si.mu.RUnlock()
		return canonical
	}
	si.mu.RUnlock()

	// Upgrade to write lock
	si.mu.Lock()
	defer si.mu.Unlock()

	// Double check
	if canonical, ok := si.pool[s]; ok {
		return canonical
	}

	// Prevent unbounded pool growth
	if len(si.pool) >= si.maxCap {
		return s
	}

	si.pool[s] = s
	return s
}

// Len returns the current number of unique strings in the pool.
func (si *StringInterner) Len() int {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return len(si.pool)
}

// Reset clears the pool.
func (si *StringInterner) Reset() {
	si.mu.Lock()
	defer si.mu.Unlock()
	si.pool = make(map[string]string, 1024)
}

// Intern uses the global default pool to intern string s.
func Intern(s string) string {
	return DefaultInterner.Intern(s)
}
