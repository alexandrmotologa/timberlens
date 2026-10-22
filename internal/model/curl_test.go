package model

import (
	"strings"
	"testing"
)

func TestGenerateCurlFromFields(t *testing.T) {
	rec := &LogRecord{
		Message: "Incoming request",
		Fields: map[string]any{
			"method": "POST",
			"path":   "/api/v1/checkout/charge",
			"body":   map[string]any{"amount": 99.5, "currency": "USD"},
		},
	}

	curl, ok := GenerateCurl(rec)
	if !ok {
		t.Fatalf("Expected cURL command to be generated")
	}

	if !strings.Contains(curl, "curl -X POST") {
		t.Errorf("Expected curl -X POST, got %s", curl)
	}
	if !strings.Contains(curl, "/api/v1/checkout/charge") {
		t.Errorf("Expected path in curl command, got %s", curl)
	}
	if !strings.Contains(curl, "--data") {
		t.Errorf("Expected --data payload in curl command, got %s", curl)
	}
}

func TestGenerateCurlInferredFromMessage(t *testing.T) {
	rec := &LogRecord{
		Message: "200 OK GET /api/v1/health",
		Fields:  map[string]any{},
	}

	curl, ok := GenerateCurl(rec)
	if !ok {
		t.Fatalf("Expected cURL to be inferred from message")
	}

	if !strings.Contains(curl, "curl -X GET") || !strings.Contains(curl, "/api/v1/health") {
		t.Errorf("Unexpected curl command: %s", curl)
	}
}
