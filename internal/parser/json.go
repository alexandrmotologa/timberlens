package parser

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// Level alias candidates
var levelAliases = []string{"level", "lvl", "severity", "log_level", "logLevel", "log.level"}

// Message alias candidates
var messageAliases = []string{"message", "msg", "text", "description", "log"}

// Timestamp alias candidates
var timestampAliases = []string{"time", "timestamp", "@timestamp", "ts", "datetime", "date"}

// Service alias candidates
var serviceAliases = []string{"service", "app", "component", "name", "application", "service.name"}

// Trace ID alias candidates
var traceAliases = []string{"trace_id", "traceId", "trace", "correlation_id", "correlationId", "requestId", "reqId", "request_id"}

// Span ID alias candidates
var spanAliases = []string{"span_id", "spanId", "span"}

// Stack trace alias candidates
var stackAliases = []string{"stack", "stacktrace", "stack_trace", "exception", "err.stack", "error.stack"}

// ParseJSON takes a JSON line byte slice and decodes it into a normalized LogRecord.
func ParseJSON(data []byte, seqID uint64) (*model.LogRecord, error) {
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return nil, err
	}

	record := &model.LogRecord{
		ID:        seqID,
		Format:    model.FormatJSON,
		FormatStr: model.FormatJSON.String(),
		Raw:       string(data),
		Fields:    rawMap,
		Numeric:   make(map[string]float64),
	}

	// 1. Extract Level
	if val, ok := findFirstField(rawMap, levelAliases); ok {
		levelStr := fmt.Sprintf("%v", val)
		record.Level = model.ParseLevel(levelStr)
		record.LevelStr = record.Level.String()
	} else {
		record.Level = model.LevelInfo
		record.LevelStr = model.LevelInfo.String()
	}

	// 2. Extract Message
	if val, ok := findFirstField(rawMap, messageAliases); ok {
		record.Message = model.Intern(fmt.Sprintf("%v", val))
	}

	// 3. Extract Service Name
	if val, ok := findFirstField(rawMap, serviceAliases); ok {
		record.Service = model.Intern(fmt.Sprintf("%v", val))
	} else {
		record.Service = model.Intern("app")
	}

	// 4. Extract Timestamp
	if val, ok := findFirstField(rawMap, timestampAliases); ok {
		record.Timestamp = parseAnyTimestamp(val)
	} else {
		record.Timestamp = time.Now()
	}

	// 5. Extract Trace ID & Span ID
	if val, ok := findFirstField(rawMap, traceAliases); ok {
		record.TraceID = model.Intern(fmt.Sprintf("%v", val))
	}
	if val, ok := findFirstField(rawMap, spanAliases); ok {
		record.SpanID = model.Intern(fmt.Sprintf("%v", val))
	}

	// 6. Extract Stack Trace if present
	if val, ok := findFirstField(rawMap, stackAliases); ok {
		stackStr := fmt.Sprintf("%v", val)
		if len(stackStr) > 20 {
			record.StackTrace = &model.StackTraceInfo{
				Raw: stackStr,
			}
		}
	}

	// 7. Extract Numeric fields for filtering (e.g. latency_ms, status, duration)
	for k, v := range rawMap {
		if num, ok := toFloat(v); ok {
			record.Numeric[k] = num
		}
	}

	return record, nil
}

func findFirstField(m map[string]any, keys []string) (any, bool) {
	for _, key := range keys {
		if val, exists := m[key]; exists && val != nil {
			return val, true
		}
		// Also support dot notation for 1 level (e.g. "err.stack")
		if strings.Contains(key, ".") {
			parts := strings.SplitN(key, ".", 2)
			if sub, exists := m[parts[0]]; exists {
				if subMap, ok := sub.(map[string]any); ok {
					if subVal, subExists := subMap[parts[1]]; subExists && subVal != nil {
						return subVal, true
					}
				}
			}
		}
	}
	return nil, false
}

func parseAnyTimestamp(val any) time.Time {
	switch v := val.(type) {
	case string:
		// Try RFC3339Nano
		if t, err := time.Parse(time.RFC3339Nano, v); err == nil {
			return t
		}
		// Try RFC3339
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t
		}
		// Try ISO8601 variant without timezone
		if t, err := time.Parse("2006-01-02 15:04:05.999999999", v); err == nil {
			return t
		}
		if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			return t
		}
		// Try string numeric epoch
		if num, err := strconv.ParseInt(v, 10, 64); err == nil {
			return epochToTime(num)
		}
	case float64:
		return epochToTime(int64(v))
	case int64:
		return epochToTime(v)
	case int:
		return epochToTime(int64(v))
	}
	return time.Now()
}

func epochToTime(epoch int64) time.Time {
	// If epoch is in seconds (e.g. 10 digits: ~1.7e9)
	if epoch < 10000000000 {
		return time.Unix(epoch, 0)
	}
	// If epoch is in milliseconds (e.g. 13 digits)
	if epoch < 10000000000000 {
		return time.UnixMilli(epoch)
	}
	// If epoch is in microseconds (e.g. 16 digits)
	if epoch < 10000000000000000 {
		return time.UnixMicro(epoch)
	}
	// Nanoseconds
	return time.Unix(0, epoch)
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}
