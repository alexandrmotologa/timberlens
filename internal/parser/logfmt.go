package parser

import (
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// ParseLogfmt parses a logfmt formatted string into a normalized LogRecord.
func ParseLogfmt(data []byte, seqID uint64) (*model.LogRecord, error) {
	fields := make(map[string]any)
	numeric := make(map[string]float64)

	rawStr := string(data)
	i := 0
	n := len(data)

	for i < n {
		// Skip spaces
		for i < n && (data[i] == ' ' || data[i] == '\t' || data[i] == '\r' || data[i] == '\n') {
			i++
		}
		if i >= n {
			break
		}

		// Read Key
		keyStart := i
		for i < n && data[i] != '=' && data[i] != ' ' && data[i] != '\t' {
			i++
		}
		if i >= n || data[i] != '=' {
			// Bare word or end
			continue
		}

		key := string(data[keyStart:i])
		i++ // skip '='

		if i >= n {
			fields[key] = ""
			break
		}

		// Read Value
		var val string
		if data[i] == '"' {
			// Quoted string
			i++ // skip opening quote
			var sb strings.Builder
			for i < n {
				if data[i] == '\\' && i+1 < n {
					sb.WriteByte(data[i+1])
					i += 2
					continue
				}
				if data[i] == '"' {
					i++ // skip closing quote
					break
				}
				sb.WriteByte(data[i])
				i++
			}
			val = sb.String()
		} else {
			// Unquoted value until space
			valStart := i
			for i < n && data[i] != ' ' && data[i] != '\t' && data[i] != '\r' && data[i] != '\n' {
				i++
			}
			val = string(data[valStart:i])
		}

		fields[key] = val
		if num, err := strconv.ParseFloat(val, 64); err == nil {
			numeric[key] = num
		}
	}

	record := &model.LogRecord{
		ID:        seqID,
		Format:    model.FormatLogfmt,
		FormatStr: model.FormatLogfmt.String(),
		Raw:       rawStr,
		Fields:    fields,
		Numeric:   numeric,
	}

	// Normalize Level
	if val, ok := findFirstField(fields, levelAliases); ok {
		record.Level = model.ParseLevel(val.(string))
		record.LevelStr = record.Level.String()
	} else {
		record.Level = model.LevelInfo
		record.LevelStr = model.LevelInfo.String()
	}

	// Normalize Message
	if val, ok := findFirstField(fields, messageAliases); ok {
		record.Message = model.Intern(val.(string))
	} else {
		record.Message = model.Intern(rawStr)
	}

	// Normalize Service
	if val, ok := findFirstField(fields, serviceAliases); ok {
		record.Service = model.Intern(val.(string))
	} else {
		record.Service = model.Intern("app")
	}

	// Normalize Timestamp
	if val, ok := findFirstField(fields, timestampAliases); ok {
		record.Timestamp = parseAnyTimestamp(val)
	} else {
		record.Timestamp = time.Now()
	}

	// Normalize Trace ID & Span ID
	if val, ok := findFirstField(fields, traceAliases); ok {
		record.TraceID = model.Intern(val.(string))
	}
	if val, ok := findFirstField(fields, spanAliases); ok {
		record.SpanID = model.Intern(val.(string))
	}

	return record, nil
}
