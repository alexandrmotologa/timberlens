package parser

import (
	"testing"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestParseJSONPino(t *testing.T) {
	line := []byte(`{"level":30,"time":1698000000000,"pid":123,"hostname":"srv1","name":"auth-svc","msg":"User logged in successfully","trace_id":"trc-987"}`)
	rec, err := ParseLine(line, 1)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if rec.Format != model.FormatJSON {
		t.Errorf("Expected FormatJSON, got %v", rec.Format)
	}
	if rec.Level != model.LevelInfo {
		t.Errorf("Expected LevelInfo (30), got %v", rec.Level)
	}
	if rec.Service != "auth-svc" {
		t.Errorf("Expected Service auth-svc, got %q", rec.Service)
	}
	if rec.Message != "User logged in successfully" {
		t.Errorf("Expected Message 'User logged in successfully', got %q", rec.Message)
	}
	if rec.TraceID != "trc-987" {
		t.Errorf("Expected TraceID 'trc-987', got %q", rec.TraceID)
	}
}

func TestParseJSONZap(t *testing.T) {
	line := []byte(`{"level":"error","ts":1698000000.123,"caller":"db/client.go:42","msg":"connection timeout","service":"billing","latency_ms":540.2}`)
	rec, err := ParseLine(line, 2)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if rec.Level != model.LevelError {
		t.Errorf("Expected LevelError, got %v", rec.Level)
	}
	if rec.Service != "billing" {
		t.Errorf("Expected Service billing, got %q", rec.Service)
	}
	if rec.Numeric["latency_ms"] != 540.2 {
		t.Errorf("Expected latency_ms 540.2, got %v", rec.Numeric["latency_ms"])
	}
}

func TestParseLogfmt(t *testing.T) {
	line := []byte(`time="2026-10-22T10:00:00Z" level=warn service=cache msg="cache miss for key" key="usr:44" duration_ms=12.5`)
	rec, err := ParseLine(line, 3)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if rec.Format != model.FormatLogfmt {
		t.Errorf("Expected FormatLogfmt, got %v", rec.Format)
	}
	if rec.Level != model.LevelWarn {
		t.Errorf("Expected LevelWarn, got %v", rec.Level)
	}
	if rec.Service != "cache" {
		t.Errorf("Expected Service cache, got %q", rec.Service)
	}
	if rec.Numeric["duration_ms"] != 12.5 {
		t.Errorf("Expected duration_ms 12.5, got %v", rec.Numeric["duration_ms"])
	}
}

func TestDockerComposePrefix(t *testing.T) {
	line := []byte(`api_1  | {"level":"info","msg":"request completed","status":200}`)
	rec, err := ParseLine(line, 4)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if rec.Service != "api" {
		t.Errorf("Expected Service 'api', got %q", rec.Service)
	}
	if rec.Message != "request completed" {
		t.Errorf("Expected Message 'request completed', got %q", rec.Message)
	}
}

func TestANSIStripping(t *testing.T) {
	line := []byte("\033[31m[ERROR]\033[0m failed to connect to database")
	rec, err := ParseLine(line, 5)
	if err != nil {
		t.Fatalf("ParseLine failed: %v", err)
	}

	if rec.Level != model.LevelError {
		t.Errorf("Expected LevelError, got %v", rec.Level)
	}
	if rec.Raw != "[ERROR] failed to connect to database" {
		t.Errorf("Expected ANSI stripped raw, got %q", rec.Raw)
	}
}
