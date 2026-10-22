package stacktrace

import (
	"regexp"
	"strings"
)

// Frame patterns for various languages
var (
	// Node.js: "    at Context.<anonymous> (/app/src/auth.ts:25:9)" or "    at /app/node_modules/express/lib/router.js:140:12"
	nodeFrameRegex = regexp.MustCompile(`^\s*at\s+(?:([^\(]+)\s+\((.+):(\d+):(\d+)\)|(.+):(\d+):(\d+))`)

	// Go: "\t/home/user/project/main.go:42 +0x2b"
	goFrameRegex = regexp.MustCompile(`^\s*(.+):(\d+)(?:\s+\+0x[0-9a-fA-F]+)?$`)

	// Python: "  File "/app/services/worker.py", line 45, in process_job"
	pythonFrameRegex = regexp.MustCompile(`^\s*File\s+"([^"]+)",\s+line\s+(\d+),\s+in\s+(.+)$`)

	// Java: "    at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1089)"
	javaFrameRegex = regexp.MustCompile(`^\s*at\s+([a-zA-Z0-9_\.\$]+)\(([a-zA-Z0-9_\.]+\.java):(\d+)\)`)
)

// Vendor path substrings for folding
var vendorKeywords = []string{
	"node_modules",
	"node:",
	"node:internal",
	"internal/modules",
	"internal/process",
	"site-packages",
	"/usr/lib/python",
	"/usr/local/lib/python",
	"/usr/local/go/",
	"/usr/lib/go/",
	"/go/src/runtime/",
	"/go/pkg/mod/",
	"org.springframework.",
	"org.apache.",
	"org.hibernate.",
	"io.netty.",
	"com.sun.",
	"java.lang.",
	"java.util.",
	"jakarta.",
}

// IsVendorPath determines whether a file path or class belongs to third-party runtime/framework code.
func IsVendorPath(path string) bool {
	lower := strings.ToLower(path)
	for _, kw := range vendorKeywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
