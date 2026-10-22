package parser

import (
	"bytes"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// ParseLine processes a single line from the stream, auto-detecting prefix, format, and fields.
func ParseLine(line []byte, seqID uint64) (*model.LogRecord, error) {
	// Clean line endings
	line = bytes.TrimRight(line, "\r\n")
	if len(line) == 0 {
		return &model.LogRecord{
			ID:        seqID,
			Timestamp: time.Now(),
			Level:     model.LevelInfo,
			LevelStr:  model.LevelInfo.String(),
			Format:    model.FormatPlainText,
			FormatStr: model.FormatPlainText.String(),
			Raw:       "",
			Message:   "",
		}, nil
	}

	// 1. Strip ANSI escape sequences
	cleanBytes := StripANSIBytes(line)

	// 2. Check for Docker Compose / Foreman prefix (e.g. "api_1  | " or "web.1 | ")
	servicePrefix, payloadBytes := extractPipePrefix(cleanBytes)

	// 3. Detect Format of the payload
	format := DetectFormat(payloadBytes)

	var record *model.LogRecord
	var err error

	switch format {
	case model.FormatJSON:
		record, err = ParseJSON(payloadBytes, seqID)
	case model.FormatLogfmt:
		record, err = ParseLogfmt(payloadBytes, seqID)
	case model.FormatSyslog:
		record, err = ParseSyslog(payloadBytes, seqID)
	default:
		record = parsePlainText(payloadBytes, seqID)
	}

	if err != nil {
		// Fallback to plain text on parse failure
		record = parsePlainText(payloadBytes, seqID)
	}

	// Preserve original raw line
	record.Raw = string(cleanBytes)

	// If service was detected from pipe prefix and record has generic service, override
	if servicePrefix != "" && (record.Service == "app" || record.Service == "syslog" || record.Service == "") {
		record.Service = model.Intern(servicePrefix)
	}

	return record, nil
}

// extractPipePrefix checks for "service_name | " pattern common in docker-compose, honcho, foreman
func extractPipePrefix(b []byte) (string, []byte) {
	// Look for " | " within first 40 characters
	idx := bytes.Index(b, []byte(" | "))
	if idx > 0 && idx < 40 {
		prefix := strings.TrimSpace(string(b[:idx]))
		// Ensure prefix looks like a service name (no spaces)
		if !strings.Contains(prefix, " ") && len(prefix) > 0 {
			// Strip standard container suffix like "_1" or "-1" for cleaner service grouping if desired,
			// or keep exact name. Keeping exact service base name:
			cleanService := strings.TrimRight(prefix, "0123456789-_.")
			if cleanService == "" {
				cleanService = prefix
			}
			return cleanService, bytes.TrimSpace(b[idx+3:])
		}
	}
	return "", b
}

func parsePlainText(data []byte, seqID uint64) *model.LogRecord {
	str := string(data)
	level := detectLevelFromText(str)

	return &model.LogRecord{
		ID:        seqID,
		Timestamp: time.Now(),
		Level:     level,
		LevelStr:  level.String(),
		Service:   model.Intern("app"),
		Message:   model.Intern(str),
		Raw:       str,
		Format:    model.FormatPlainText,
		FormatStr: model.FormatPlainText.String(),
		Fields:    make(map[string]any),
		Numeric:   make(map[string]float64),
	}
}

func detectLevelFromText(s string) model.SeverityLevel {
	upper := strings.ToUpper(s)
	if strings.Contains(upper, "FATAL") || strings.Contains(upper, "PANIC") {
		return model.LevelFatal
	}
	if strings.Contains(upper, "ERROR") || strings.Contains(upper, "ERR:") || strings.Contains(upper, "EXCEPTION") {
		return model.LevelError
	}
	if strings.Contains(upper, "WARN") || strings.Contains(upper, "WARNING") {
		return model.LevelWarn
	}
	if strings.Contains(upper, "DEBUG") || strings.Contains(upper, "DBG") {
		return model.LevelDebug
	}
	if strings.Contains(upper, "TRACE") {
		return model.LevelTrace
	}
	return model.LevelInfo
}
