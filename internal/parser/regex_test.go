package parser

import (
	"testing"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestRegexParserCustomFormat(t *testing.T) {
	pattern := `^\[(?P<time>[^\]]+)\] \[(?P<level>[A-Z]+)\] \[(?P<service>[^\]]+)\] (?P<msg>.*)$`
	rp, err := NewRegexParser(pattern)
	if err != nil {
		t.Fatalf("Failed to compile regex: %v", err)
	}

	line := []byte("[2026-10-22 14:30:00] [ERROR] [auth-gateway] Connection to Redis failed")
	rec, ok := rp.Parse(line, 1)
	if !ok {
		t.Fatalf("Expected line to match custom regex")
	}

	if rec.Level != model.LevelError {
		t.Errorf("Expected LevelError, got %v", rec.Level)
	}
	if rec.Service != "auth-gateway" {
		t.Errorf("Expected Service auth-gateway, got %q", rec.Service)
	}
	if rec.Message != "Connection to Redis failed" {
		t.Errorf("Expected message 'Connection to Redis failed', got %q", rec.Message)
	}
}
