# Architecture and Internals

TimberLens is engineered for low latency and bounded memory consumption when ingesting unbounded log streams from local development environments.

## Core Subsystems

### 1. Ingestion Subsystem (`internal/ingest`)
The ingestion pipeline reads from standard input (`os.Stdin`) using a reusable byte buffer rather than standard `bufio.Scanner` to avoid default buffer allocation limits and string copying overhead.
- Supports lines up to 1MB in length without crashing or dropping packets.
- Employs a non-blocking ingestion ring channel that applies backpressure without blocking incoming pipes if system consumers fall behind.
- Optional network listeners accept RFC 5424/3164 Syslog over UDP/TCP as well as OpenTelemetry logs over HTTP (`/v1/logs`).

### 2. Auto-Detection and Parser Engine (`internal/parser`)
The parser operates without schema pre-registration:
- **Sniffing heuristic**: Inspects first non-whitespace characters. If lines begin with `{` and terminate with `}`, they pass directly to the zero-allocation JSON field extractor.
- **Logfmt parsing**: Key-value pairs (`foo=bar baz="hello world"`) are extracted via a single-pass scanner that handles quoted strings and escaped values.
- **Field normalization**: Disparate logging formats (Pino, Winston, Zap, Logrus, Slog, Structlog) use varying field names. TimberLens normalizes aliases into standard properties:
  - Severity: `level`, `lvl`, `severity`, `log_level` -> normalized to `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`.
  - Message: `message`, `msg`, `text`, `description`.
  - Timestamp: `time`, `timestamp`, `@timestamp`, `ts` -> parsed to UNIX nanoseconds.
  - Service: `service`, `app`, `component`, `application`.
  - Trace ID: `trace_id`, `traceId`, `trace`, `correlation_id`, `requestId`.
  - Latency: `latency`, `latency_ms`, `duration`, `elapsed`, `response_time_ms`.
- **ANSI stripping**: Strips escape codes from messages while recording service color hints.

### 3. Ring Buffer and In-Memory Index (`internal/buffer`)
- Ring buffer allocates a contiguous fixed-size circular slice of log record entries (default 1,000,000 items).
- Uses atomic sequence counters to write new entries at `head % capacity`, naturally wrapping around and overwriting older entries without dynamic slice resizes or garbage collection sweeps.
- String fields are interned into a deduplicated string table to preserve memory.
- In-memory index tracks offsets for fast lookup by service name, log level, and trace identifiers.

### 4. Stack Trace Cleaner and Folding Engine (`internal/stacktrace`)
Runtime stack traces often overwhelm logs with internal framework frames. The folding engine:
- Detects multi-line stack trace blocks (Node.js/V8, Go runtime, Python tracebacks, Java JVM).
- Distinguishes user code paths from known library paths (`node_modules/`, `vendor/`, `site-packages/`, `/usr/local/go/`, standard libraries).
- Collapses external frames into a single fold summary while keeping the immediate user origin frame visible.

### 5. Terminal TUI (`internal/tui`)
Built with Charm's Bubble Tea and Lip Gloss:
- Viewport virtualization renders only visible rows on screen, ensuring fluid 60fps scrolling even when receiving tens of thousands of lines per second.
- Custom keybindings allow quick pausing, following, search filtering, and JSON modal inspection.

### 6. Web Studio and SSE Server (`internal/server` & `web/`)
- Embedded HTTP server built with Go `net/http`.
- Server-Sent Events (SSE) endpoint broadcasts newly ingested log events to browser clients in microsecond batches.
- React 19 frontend features virtualized list rendering, interactive search filter bar, and distributed trace inspection.
- The web assets are embedded into the Go binary at compile time via `//go:embed`.
