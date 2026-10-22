package parser

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// ParseSyslog parses RFC 5424 or RFC 3164 syslog messages.
func ParseSyslog(data []byte, seqID uint64) (*model.LogRecord, error) {
	rawStr := string(data)
	trimmed := bytes.TrimSpace(data)

	record := &model.LogRecord{
		ID:        seqID,
		Format:    model.FormatSyslog,
		FormatStr: model.FormatSyslog.String(),
		Raw:       rawStr,
		Level:     model.LevelInfo,
		LevelStr:  model.LevelInfo.String(),
		Service:   model.Intern("syslog"),
		Timestamp: time.Now(),
		Fields:    make(map[string]any),
		Numeric:   make(map[string]float64),
	}

	if len(trimmed) == 0 || trimmed[0] != '<' {
		record.Message = model.Intern(rawStr)
		return record, nil
	}

	endPri := bytes.IndexByte(trimmed, '>')
	if endPri <= 1 {
		record.Message = model.Intern(rawStr)
		return record, nil
	}

	priStr := string(trimmed[1:endPri])
	pri, err := strconv.Atoi(priStr)
	if err == nil {
		sev := pri & 7
		record.Level = syslogSeverityToModel(sev)
		record.LevelStr = record.Level.String()
		record.Fields["facility"] = pri >> 3
		record.Fields["priority"] = pri
	}

	rest := trimmed[endPri+1:]

	// Check if RFC 5424: starts with version digit like "1 "
	if len(rest) > 2 && rest[0] >= '1' && rest[0] <= '9' && rest[1] == ' ' {
		parseRFC5424(rest[2:], record)
	} else {
		parseRFC3164(rest, record)
	}

	return record, nil
}

func parseRFC5424(data []byte, record *model.LogRecord) {
	// Format: TIMESTAMP HOSTNAME APP-NAME PROCID MSGID [STRUCTURED-DATA] MSG
	parts := strings.SplitN(string(data), " ", 7)
	if len(parts) >= 1 && parts[0] != "-" {
		if t, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
			record.Timestamp = t
		}
	}
	if len(parts) >= 2 && parts[1] != "-" {
		record.Fields["hostname"] = parts[1]
	}
	if len(parts) >= 3 && parts[2] != "-" {
		record.Service = model.Intern(parts[2])
	}
	if len(parts) >= 4 && parts[3] != "-" {
		record.Fields["procid"] = parts[3]
	}
	if len(parts) >= 5 && parts[4] != "-" {
		record.Fields["msgid"] = parts[4]
	}

	msg := ""
	if len(parts) >= 7 {
		msg = parts[6]
	} else if len(parts) >= 6 {
		msg = parts[5]
	}
	record.Message = model.Intern(strings.TrimSpace(msg))
}

func parseRFC3164(data []byte, record *model.LogRecord) {
	// Format: Mmm dd hh:mm:ss HOSTNAME TAG: MSG
	str := string(data)
	if len(str) >= 15 {
		// Attempt parsing timestamp e.g. "Oct 11 22:14:15"
		now := time.Now()
		tsStr := fmtRFC3164Time(str[:15], now.Year())
		if t, err := time.Parse("2006 Jan 2 15:04:05", tsStr); err == nil {
			record.Timestamp = t
			str = strings.TrimSpace(str[15:])
		}
	}

	// Next token is HOSTNAME or TAG:
	parts := strings.SplitN(str, " ", 2)
	if len(parts) == 2 {
		record.Fields["hostname"] = parts[0]
		remaining := parts[1]
		colonIdx := strings.Index(remaining, ":")
		if colonIdx != -1 && colonIdx < 30 {
			record.Service = model.Intern(strings.TrimSpace(remaining[:colonIdx]))
			record.Message = model.Intern(strings.TrimSpace(remaining[colonIdx+1:]))
		} else {
			record.Message = model.Intern(remaining)
		}
	} else {
		record.Message = model.Intern(str)
	}
}

func fmtRFC3164Time(ts string, year int) string {
	return strconv.Itoa(year) + " " + ts
}

func syslogSeverityToModel(sev int) model.SeverityLevel {
	switch sev {
	case 0, 1, 2:
		return model.LevelFatal
	case 3:
		return model.LevelError
	case 4:
		return model.LevelWarn
	case 5, 6:
		return model.LevelInfo
	case 7:
		return model.LevelDebug
	default:
		return model.LevelInfo
	}
}
