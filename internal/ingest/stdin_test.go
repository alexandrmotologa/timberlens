package ingest

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

func TestStdinReaderIngestion(t *testing.T) {
	input := `{"level":"info","service":"web","msg":"hello world"}
{"level":"error","service":"db","msg":"connection lost"}
plain text line without json
`
	var records []*model.LogRecord

	consumer := func(rec *model.LogRecord) {
		records = append(records, rec)
	}

	reader := NewStdinReader(strings.NewReader(input), consumer)
	err := reader.Start()
	if err != nil {
		t.Fatalf("StdinReader.Start failed: %v", err)
	}

	if len(records) != 3 {
		t.Fatalf("Expected 3 records, got %d", len(records))
	}

	if records[0].Service != "web" || records[0].Level != model.LevelInfo {
		t.Errorf("Unexpected record 0: %+v", records[0])
	}
	if records[1].Service != "db" || records[1].Level != model.LevelError {
		t.Errorf("Unexpected record 1: %+v", records[1])
	}
	if records[2].Format != model.FormatPlainText {
		t.Errorf("Unexpected record 2 format: %v", records[2].Format)
	}
}
