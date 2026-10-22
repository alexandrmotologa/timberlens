package model

import (
	"strings"
	"time"
)

// SeverityLevel represents standard normalized log severity.
type SeverityLevel uint8

const (
	LevelUnknown SeverityLevel = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the canonical uppercase string representation of the level.
func (l SeverityLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "INFO"
	}
}

// ColorCode returns an ANSI 256-color or 16-color code suitable for terminal TUI.
func (l SeverityLevel) ColorCode() string {
	switch l {
	case LevelTrace:
		return "\033[90m" // Gray
	case LevelDebug:
		return "\033[36m" // Cyan
	case LevelInfo:
		return "\033[32m" // Green
	case LevelWarn:
		return "\033[33m" // Yellow
	case LevelError:
		return "\033[31m" // Red
	case LevelFatal:
		return "\033[35m" // Magenta
	default:
		return "\033[37m" // White
	}
}

// ParseLevel normalizes diverse logging library level representations.
func ParseLevel(val string) SeverityLevel {
	clean := strings.TrimSpace(strings.ToUpper(val))
	switch clean {
	case "TRACE", "TRC", "FINE", "FINER", "FINEST":
		return LevelTrace
	case "DEBUG", "DBG", "DEBU", "D":
		return LevelDebug
	case "INFO", "INF", "NOTICE", "I":
		return LevelInfo
	case "WARN", "WARNING", "WRN", "W":
		return LevelWarn
	case "ERROR", "ERR", "ERRO", "E", "CRIT", "CRITICAL":
		return LevelError
	case "FATAL", "FATL", "EMERG", "EMERGENCY", "PANIC", "ALERT":
		return LevelFatal
	default:
		// Numeric levels common in syslog or Pino
		switch clean {
		case "10":
			return LevelTrace
		case "20":
			return LevelDebug
		case "30":
			return LevelInfo
		case "40":
			return LevelWarn
		case "50":
			return LevelError
		case "60":
			return LevelFatal
		default:
			return LevelInfo
		}
	}
}

// FormatType represents detected log format.
type FormatType uint8

const (
	FormatUnknown FormatType = iota
	FormatJSON
	FormatLogfmt
	FormatSyslog
	FormatPlainText
)

// String returns human readable format name.
func (f FormatType) String() string {
	switch f {
	case FormatJSON:
		return "JSON"
	case FormatLogfmt:
		return "Logfmt"
	case FormatSyslog:
		return "Syslog"
	case FormatPlainText:
		return "PlainText"
	default:
		return "Unknown"
	}
}

// Field represents a key-value attribute extracted from structured logs.
type Field struct {
	Key      string  `json:"key"`
	Value    string  `json:"value"`
	NumValue float64 `json:"numValue,omitempty"`
	IsNum    bool    `json:"isNum,omitempty"`
}

// StackFrame represents a single line/frame inside a stack trace.
type StackFrame struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
	IsVendor bool   `json:"isVendor"`
	Raw      string `json:"raw"`
}

// StackTraceInfo holds parsed stack trace metadata and folded vendor frames.
type StackTraceInfo struct {
	Raw          string       `json:"raw"`
	Language     string       `json:"language"`
	RootCause    string       `json:"rootCause"`
	TotalFrames  int          `json:"totalFrames"`
	HiddenFrames int          `json:"hiddenFrames"`
	Frames       []StackFrame `json:"frames"`
}

// LogRecord is the fully decoded in-memory representation of a log event.
type LogRecord struct {
	ID         uint64            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Level      SeverityLevel     `json:"level"`
	LevelStr   string            `json:"levelStr"`
	Service    string            `json:"service"`
	TraceID    string            `json:"traceId,omitempty"`
	SpanID     string            `json:"spanId,omitempty"`
	Message    string            `json:"message"`
	Raw        string            `json:"raw"`
	Format     FormatType        `json:"format"`
	FormatStr  string            `json:"formatStr"`
	Fields     map[string]any    `json:"fields,omitempty"`
	Numeric    map[string]float64 `json:"numeric,omitempty"`
	StackTrace *StackTraceInfo   `json:"stackTrace,omitempty"`
}

// CompactRecord is the memory-optimized flat representation kept in the circular ring buffer.
type CompactRecord struct {
	SeqID       uint64
	TimestampNs int64
	Level       SeverityLevel
	Format      FormatType
	HasStack    bool
	Service     string // interned pointer
	TraceID     string // interned pointer
	SpanID      string // interned pointer
	Message     string // interned or truncated preview
	RawBytes    []byte // reference or slice in byte arena
}
