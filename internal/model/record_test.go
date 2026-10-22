package model

import (
	"testing"
	"time"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected SeverityLevel
	}{
		{"DEBUG", LevelDebug},
		{"dbg", LevelDebug},
		{"info", LevelInfo},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"err", LevelError},
		{"fatal", LevelFatal},
		{"panic", LevelFatal},
		{"30", LevelInfo},
		{"50", LevelError},
		{"unknown", LevelInfo},
	}

	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, expected %v", tt.input, got, tt.expected)
		}
	}
}

func TestStringInterner(t *testing.T) {
	interner := NewStringInterner(100)

	s1 := interner.Intern("service_auth")
	s2 := interner.Intern("service_auth")

	if s1 != s2 {
		t.Errorf("Expected identical strings, got %q and %q", s1, s2)
	}

	if interner.Len() != 1 {
		t.Errorf("Expected pool length 1, got %d", interner.Len())
	}
}

func TestLogRecordCreation(t *testing.T) {
	rec := LogRecord{
		ID:        1,
		Timestamp: time.Now(),
		Level:     LevelError,
		LevelStr:  LevelError.String(),
		Service:   "api-gateway",
		Message:   "connection refused",
		Format:    FormatJSON,
		FormatStr: FormatJSON.String(),
	}

	if rec.LevelStr != "ERROR" {
		t.Errorf("Expected LevelStr ERROR, got %q", rec.LevelStr)
	}
	if rec.FormatStr != "JSON" {
		t.Errorf("Expected FormatStr JSON, got %q", rec.FormatStr)
	}
}
