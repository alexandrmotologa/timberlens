package parser

import (
	"regexp"
	"strings"
)

// ansiRegex matches standard ANSI color and formatting escape sequences.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes all ANSI escape sequences from a string.
func StripANSI(s string) string {
	if !strings.Contains(s, "\x1b") {
		return s
	}
	return ansiRegex.ReplaceAllString(s, "")
}

// StripANSIBytes removes ANSI escape sequences from a byte slice.
func StripANSIBytes(b []byte) []byte {
	return []byte(StripANSI(string(b)))
}
