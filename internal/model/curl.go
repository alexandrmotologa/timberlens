package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var (
	httpMethodRegex = regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+([/\w\.\-\?\=\&\%]+)`)
)

// GenerateCurl synthesizes a runnable cURL command if the log record contains HTTP request details.
func GenerateCurl(rec *LogRecord) (string, bool) {
	if rec == nil {
		return "", false
	}

	method := ""
	path := ""

	// 1. Check explicit fields
	if m, ok := rec.Fields["method"]; ok && m != nil {
		method = strings.ToUpper(fmt.Sprintf("%v", m))
	}
	if p, ok := rec.Fields["path"]; ok && p != nil {
		path = fmt.Sprintf("%v", p)
	} else if u, ok := rec.Fields["url"]; ok && u != nil {
		path = fmt.Sprintf("%v", u)
	}

	// 2. Infer from message if not in fields
	if method == "" || path == "" {
		match := httpMethodRegex.FindStringSubmatch(rec.Message)
		if len(match) >= 3 {
			if method == "" {
				method = match[1]
			}
			if path == "" {
				path = match[2]
			}
		}
	}

	// If still no method or path, cannot generate cURL
	if method == "" || path == "" {
		return "", false
	}

	// Default host if path is relative
	urlStr := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		host := "localhost:3000"
		if h, ok := rec.Fields["host"]; ok && h != nil {
			host = fmt.Sprintf("%v", h)
		}
		urlStr = fmt.Sprintf("http://%s%s", host, path)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("curl -X %s %q", method, urlStr))

	// Headers
	hasContentType := false
	if headers, ok := rec.Fields["headers"]; ok && headers != nil {
		if hMap, ok := headers.(map[string]any); ok {
			for k, v := range hMap {
				sb.WriteString(fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %v", k, v)))
				if strings.EqualFold(k, "Content-Type") {
					hasContentType = true
				}
			}
		}
	}

	// Body / Payload
	var bodyStr string
	for _, bodyKey := range []string{"body", "payload", "request_body", "data"} {
		if b, exists := rec.Fields[bodyKey]; exists && b != nil {
			switch val := b.(type) {
			case string:
				bodyStr = val
			case map[string]any, []any:
				if jsonBytes, err := json.Marshal(val); err == nil {
					bodyStr = string(jsonBytes)
				}
			default:
				bodyStr = fmt.Sprintf("%v", val)
			}
			break
		}
	}

	if bodyStr != "" {
		if !hasContentType && (strings.HasPrefix(bodyStr, "{") || strings.HasPrefix(bodyStr, "[")) {
			sb.WriteString(" -H \"Content-Type: application/json\"")
		}
		sb.WriteString(fmt.Sprintf(" --data %q", bodyStr))
	}

	return sb.String(), true
}
