package buffer

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alexandrmotologa/timberlens/internal/model"
)

type OpType int

const (
	OpEqual OpType = iota
	OpNotEqual
	OpGreater
	OpGreaterEqual
	OpLess
	OpLessEqual
	OpContains
)

// Criterion represents a single filter predicate.
type Criterion struct {
	Field    string // e.g. "service", "level", "trace_id", "latency_ms", or "" for full text
	Operator OpType
	StrVal   string
	NumVal   float64
	IsNum    bool
	Negate   bool
}

// Query contains compiled criteria for filtering log records.
type Query struct {
	Raw      string
	Criteria []Criterion
}

// ParseQuery parses a raw query expression into an evaluatable Query AST.
func ParseQuery(raw string) *Query {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &Query{Raw: ""}
	}

	tokens := tokenizeQuery(raw)
	criteria := make([]Criterion, 0, len(tokens))

	for _, token := range tokens {
		crit, ok := parseToken(token)
		if ok {
			criteria = append(criteria, crit)
		}
	}

	return &Query{
		Raw:      raw,
		Criteria: criteria,
	}
}

// Match evaluates whether a LogRecord satisfies the entire query.
func (q *Query) Match(rec *model.LogRecord) bool {
	if q == nil || len(q.Criteria) == 0 {
		return true
	}

	for i := range q.Criteria {
		if !evalCriterion(&q.Criteria[i], rec) {
			return false
		}
	}
	return true
}

func evalCriterion(c *Criterion, rec *model.LogRecord) bool {
	res := evalPositive(c, rec)
	if c.Negate {
		return !res
	}
	return res
}

func evalPositive(c *Criterion, rec *model.LogRecord) bool {
	// Full-text match
	if c.Field == "" {
		target := strings.ToLower(rec.Raw)
		return strings.Contains(target, c.StrVal)
	}

	// Service filter
	if c.Field == "service" || c.Field == "app" {
		switch c.Operator {
		case OpEqual:
			return strings.EqualFold(rec.Service, c.StrVal)
		case OpNotEqual:
			return !strings.EqualFold(rec.Service, c.StrVal)
		case OpContains:
			return strings.Contains(strings.ToLower(rec.Service), c.StrVal)
		}
		return false
	}

	// Level filter
	if c.Field == "level" || c.Field == "lvl" || c.Field == "severity" {
		targetLvl := model.ParseLevel(c.StrVal)
		switch c.Operator {
		case OpEqual:
			return rec.Level == targetLvl
		case OpNotEqual:
			return rec.Level != targetLvl
		case OpGreater:
			return rec.Level > targetLvl
		case OpGreaterEqual:
			return rec.Level >= targetLvl
		case OpLess:
			return rec.Level < targetLvl
		case OpLessEqual:
			return rec.Level <= targetLvl
		}
		return false
	}

	// Trace ID filter
	if c.Field == "trace" || c.Field == "trace_id" || c.Field == "traceid" {
		switch c.Operator {
		case OpEqual:
			return rec.TraceID == c.StrVal
		case OpNotEqual:
			return rec.TraceID != c.StrVal
		case OpContains:
			return strings.Contains(rec.TraceID, c.StrVal)
		}
		return false
	}

	// Numeric field evaluation
	if c.IsNum {
		if val, exists := rec.Numeric[c.Field]; exists {
			switch c.Operator {
			case OpEqual:
				return val == c.NumVal
			case OpNotEqual:
				return val != c.NumVal
			case OpGreater:
				return val > c.NumVal
			case OpGreaterEqual:
				return val >= c.NumVal
			case OpLess:
				return val < c.NumVal
			case OpLessEqual:
				return val <= c.NumVal
			}
		}
		return false
	}

	// Generic field lookup in fields map
	if val, exists := rec.Fields[c.Field]; exists && val != nil {
		str := strings.ToLower(fmt.Sprintf("%v", val))
		switch c.Operator {
		case OpEqual:
			return str == c.StrVal
		case OpNotEqual:
			return str != c.StrVal
		case OpContains:
			return strings.Contains(str, c.StrVal)
		}
	}

	return false
}

func parseToken(token string) (Criterion, bool) {
	negate := false
	if strings.HasPrefix(token, "!") || strings.HasPrefix(token, "-") {
		negate = true
		token = token[1:]
	}

	if token == "" {
		return Criterion{}, false
	}

	// Check operators in order: >=, <=, !=, >, <, :
	operators := []struct {
		sym string
		op  OpType
	}{
		{">=", OpGreaterEqual},
		{"<=", OpLessEqual},
		{"!=", OpNotEqual},
		{">", OpGreater},
		{"<", OpLess},
		{":", OpEqual},
		{"=", OpEqual},
	}

	for _, o := range operators {
		if idx := strings.Index(token, o.sym); idx > 0 {
			field := strings.ToLower(strings.TrimSpace(token[:idx]))
			val := strings.Trim(strings.TrimSpace(token[idx+len(o.sym):]), `"'`)

			crit := Criterion{
				Field:    field,
				Operator: o.op,
				StrVal:   strings.ToLower(val),
				Negate:   negate,
			}

			// Check if numeric
			if num, err := strconv.ParseFloat(val, 64); err == nil {
				crit.NumVal = num
				crit.IsNum = true
			}

			return crit, true
		}
	}

	// Bare word -> full-text search
	cleanVal := strings.ToLower(strings.Trim(token, `"'`))
	return Criterion{
		Field:    "",
		Operator: OpContains,
		StrVal:   cleanVal,
		Negate:   negate,
	}, true
}

func tokenizeQuery(raw string) []string {
	var tokens []string
	var sb strings.Builder
	inQuote := false

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if c == '"' {
			inQuote = !inQuote
			sb.WriteByte(c)
			continue
		}
		if !inQuote && (c == ' ' || c == '\t') {
			if sb.Len() > 0 {
				tokens = append(tokens, sb.String())
				sb.Reset()
			}
			continue
		}
		sb.WriteByte(c)
	}

	if sb.Len() > 0 {
		tokens = append(tokens, sb.String())
	}

	return tokens
}
