package stacktrace

import (
	"strings"
	"testing"
)

func TestNodeJSStackTraceCleaning(t *testing.T) {
	raw := `Error: connect ECONNREFUSED 127.0.0.1:5432
    at TCPConnectWrap.afterConnect [as oncomplete] (node:net:1494:16)
    at /app/node_modules/pg-pool/index.js:45:11
    at /app/node_modules/pg/lib/client.js:120:5
    at DatabaseService.connect (/app/src/db/connection.ts:32:15)
    at Object.handler (/app/src/routes/users.ts:18:9)
    at /app/node_modules/express/lib/router/layer.js:95:5`

	info := DetectAndClean(raw)
	if info == nil {
		t.Fatalf("Expected stack trace info to be detected")
	}

	if info.Language != "nodejs" {
		t.Errorf("Expected language nodejs, got %s", info.Language)
	}

	if info.HiddenFrames < 4 {
		t.Errorf("Expected at least 4 vendor frames hidden, got %d", info.HiddenFrames)
	}

	folded := FormatFolded(info)
	if !strings.Contains(folded, "library frames hidden") {
		t.Errorf("Expected folded output to contain hidden frame notice, got:\n%s", folded)
	}
	if !strings.Contains(folded, "/app/src/db/connection.ts:32") {
		t.Errorf("Expected folded output to show user code path, got:\n%s", folded)
	}
}

func TestPythonStackTraceCleaning(t *testing.T) {
	raw := `Traceback (most recent call last):
  File "/usr/lib/python3.10/site-packages/urllib3/connectionpool.py", line 703, in urlopen
  File "/usr/lib/python3.10/site-packages/requests/adapters.py", line 489, in send
  File "/workspace/app/services/scraper.py", line 88, in fetch_url
    resp = session.get(url, timeout=5)
ConnectionError: Max retries exceeded`

	info := DetectAndClean(raw)
	if info == nil {
		t.Fatalf("Expected python stack trace info to be detected")
	}

	if info.Language != "python" {
		t.Errorf("Expected language python, got %s", info.Language)
	}

	if info.HiddenFrames < 2 {
		t.Errorf("Expected at least 2 vendor frames hidden, got %d", info.HiddenFrames)
	}

	if !strings.Contains(info.RootCause, "scraper.py:88") {
		t.Errorf("Expected root cause scraper.py:88, got %q", info.RootCause)
	}
}
