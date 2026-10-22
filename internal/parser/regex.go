package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

// RegexParser extracts structured log records using user-defined named regex groups.
type RegexParser struct {
	re        *regexp.Regexp
	groupMap  map[string]int
	timeGroup int
	lvlGroup  int
	svcGroup  int
	msgGroup  int
	trcGroup  int
}

// NewRegexParser compiles a regex with named capture groups.
func NewRegexParser(pattern string) (*RegexParser, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	rp := &RegexParser{
		re:        re,
		groupMap:  make(map[string]int),
		timeGroup: -1,
		lvlGroup:  -1,
		svcGroup:  -1,
		msgGroup:  -1,
		trcGroup:  -1,
	}

	for idx, name := range re.SubexpNames() {
		if name == "" {
			continue
		}
		rp.groupMap[name] = idx
		lower := strings.ToLower(name)
		switch lower {
		case "time", "timestamp", "ts", "datetime":
			rp.timeGroup = idx
		case "level", "lvl", "severity":
			rp.lvlGroup = idx
		case "service", "app", "component":
			rp.svcGroup = idx
		case "msg", "message", "text":
			rp.msgGroup = idx
		case "trace", "trace_id", "traceid":
			rp.trcGroup = idx
		}
	}

	return rp, nil
}

// Parse extracts a LogRecord from line using the compiled regex.
func (rp *RegexParser) Parse(line []byte, seqID uint64) (*model.LogRecord, bool) {
	clean := StripANSI(string(line))
	matches := rp.re.FindStringSubmatch(clean)
	if len(matches) == 0 {
		return nil, false
	}

	rec := &model.LogRecord{
		ID:        seqID,
		Timestamp: time.Now(),
		Level:     model.LevelInfo,
		LevelStr:  model.LevelInfo.String(),
		Service:   model.Intern("app"),
		Raw:       clean,
		Format:    model.FormatPlainText,
		FormatStr: "CustomRegex",
		Fields:    make(map[string]any),
		Numeric:   make(map[string]float64),
	}

	if rp.timeGroup != -1 && rp.timeGroup < len(matches) {
		val := matches[rp.timeGroup]
		rec.Timestamp = parseAnyTimestamp(val)
	}

	if rp.lvlGroup != -1 && rp.lvlGroup < len(matches) {
		rec.Level = model.ParseLevel(matches[rp.lvlGroup])
		rec.LevelStr = rec.Level.String()
	}

	if rp.svcGroup != -1 && rp.svcGroup < len(matches) {
		rec.Service = model.Intern(matches[rp.svcGroup])
	}

	if rp.msgGroup != -1 && rp.msgGroup < len(matches) {
		rec.Message = model.Intern(matches[rp.msgGroup])
	} else {
		rec.Message = model.Intern(clean)
	}

	if rp.trcGroup != -1 && rp.trcGroup < len(matches) {
		rec.TraceID = model.Intern(matches[rp.trcGroup])
	}

	// Capture all other named groups into Fields and Numeric
	for name, idx := range rp.groupMap {
		if idx < len(matches) {
			val := matches[idx]
			rec.Fields[name] = val
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				rec.Numeric[name] = num
			}
		}
	}

	return rec, true
}
