package parser

import (
	"bytes"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// DetectFormat inspects the raw line to determine its structured format.
func DetectFormat(line []byte) model.FormatType {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return model.FormatPlainText
	}

	// JSON Lines: starts with '{' and ends with '}'
	if trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}' {
		return model.FormatJSON
	}

	// Syslog RFC 3164 / 5424: starts with '<' followed by priority digits and '>'
	if trimmed[0] == '<' {
		idx := bytes.IndexByte(trimmed, '>')
		if idx > 1 && idx < 6 {
			allDigits := true
			for i := 1; i < idx; i++ {
				if trimmed[i] < '0' || trimmed[i] > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				return model.FormatSyslog
			}
		}
	}

	// Logfmt: check for key=value patterns
	if isLogfmt(trimmed) {
		return model.FormatLogfmt
	}

	return model.FormatPlainText
}

// isLogfmt checks if the line contains at least 2 distinct key=value patterns.
func isLogfmt(b []byte) bool {
	equalsCount := 0
	inQuote := false

	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == '"' && (i == 0 || b[i-1] != '\\') {
			inQuote = !inQuote
			continue
		}
		if !inQuote && c == '=' {
			// Check if preceded by a valid identifier character
			if i > 0 && isValidIdentChar(b[i-1]) {
				equalsCount++
				if equalsCount >= 2 {
					return true
				}
			}
		}
	}
	return false
}

func isValidIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.'
}
