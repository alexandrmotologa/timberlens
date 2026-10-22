# TimberLens: Master Implementation Plan

High-speed local structured log aggregator, JSON/Logfmt live search engine, and monorepo stream multiplexer in Go 1.23 + Charm Bubble Tea + React 19 + Tailwind CSS.

---

## 1. Project Overview & Architecture

Modern microservice, Docker Compose, and monorepo setups output streams from multiple concurrent processes (frontend, API gateway, authentication service, background worker, database). These produce interleaved, high-volume logs with nested JSON, unformatted logfmt, and multi-line stack traces that clutter standard terminal output.

**TimberLens** solves this by providing an ultra-fast, single-binary local observability engine:
- Ingests piped standard input with zero configuration at >120,000 lines/second.
- Auto-detects log formats (JSON Lines, logfmt, Syslog RFC 5424/3164, raw text).
- Stores up to 1,000,000 log records in a lock-free circular memory ring buffer under 25MB RAM using string interning.
- Cleans and collapses external vendor stack traces (Node.js, Go, Python, Java) while keeping user code paths prominent.
- Offers dual interfaces: an interactive Charm Bubble Tea terminal TUI and an embedded React 19 Web Studio with live Server-Sent Events (SSE).
- Includes an integrated demo generator (`timberlens demo`) to simulate realistic multi-service microservice traffic out of the box.

```
+-----------------------------------------------------------------------------------+
|                                    TIMBERLENS                                     |
+-----------------------------------------+-----------------------------------------+
                                          |
        +---------------------------------+---------------------------------+
        |                                                                   |
+-------v-------------------------+                 +-----------------------v-------+
|    Ingestion Subsystem          |                 |    Auto-Format Parser Engine  |
|  * Non-blocking STDIN Reader    |                 |  * JSON Lines (Pino, Zap,...) |
|  * Multi-File Follower (--file) |                 |  * logfmt Tokenizer           |
|  * UDP Syslog Receiver (:514)   |                 |  * Syslog RFC 5424/3164       |
|  * OTLP HTTP Receiver (:4318)   |                 |  * ANSI Color Sanitizer       |
|  * Demo Log Generator (demo)    |                 |  * Field Normalizer (aliases) |
+-------+-------------------------+                 +-----------------------+-------+
        |                                                                   |
        +---------------------------------+---------------------------------+
                                          |
                        +-----------------v-----------------+
                        |   Lock-Free Ring Buffer & Index   |
                        |  * 1,000,000 Record Capacity      |
                        |  * String Interning Pool (<25MB)  |
                        |  * Service & Level Inverted Index |
                        |  * Trace ID Correlator            |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |    Error Clustering & Folding     |
                        |  * Stack Trace Cleaner            |
                        |  * Vendor Frame Collapser         |
                        |  * Error Signature Aggregator     |
                        |  * Velocity & Anomaly Detector    |
                        +-----------------+-----------------+
                                          |
                        +-----------------v-----------------+
                        |    Terminal TUI & Web Studio      |
                        |  * Bubble Tea Virtual Viewport    |
                        |  * Deterministic Service Colors   |
                        |  * React 19 Web Studio (Port 9900)|
                        |  * Live SSE Broadcast Stream      |
                        |  * Trace Waterfall Timeline View  |
                        +-----------------------------------+
```

---

## 2. Directory Layout

```
timberlens/
├── cmd/
│   └── timberlens/
│       └── main.go                 # CLI entry point, flags, pipeline orchestration
├── internal/
│   ├── model/
│   │   ├── record.go               # Structured LogRecord, Level, Field types
│   │   └── intern.go               # Thread-safe string interning pool for low memory
│   ├── ingest/
│   │   ├── stdin.go                # High-speed non-blocking STDIN reader
│   │   ├── file.go                 # Continuous file tailing reader
│   │   ├── syslog.go               # UDP syslog listener
│   │   ├── otlp.go                 # OpenTelemetry HTTP logs endpoint
│   │   └── generator.go            # Realistic multi-service demo traffic generator
│   ├── parser/
│   │   ├── detector.go             # Format sniffer (JSON, logfmt, Syslog, plain)
│   │   ├── json.go                 # High-throughput JSON parser with alias mapping
│   │   ├── logfmt.go               # Zero-alloc logfmt tokenizer
│   │   ├── syslog.go               # RFC 5424 & RFC 3164 parser
│   │   └── ansi.go                 # ANSI color stripper and style extractor
│   ├── buffer/
│   │   ├── ring.go                 # Lock-free pre-allocated circular ring buffer
│   │   ├── index.go                # Inverted index for service, level, and trace_id
│   │   └── query.go                # AST and evaluator for query expressions
│   ├── stacktrace/
│   │   ├── cleaner.go              # Stack trace detection and folding engine
│   │   └── rules.go                # Pattern rules for Node, Python, Go, Java
│   ├── cluster/
│   │   └── signature.go            # Error signature grouping and recurrence counting
│   ├── tui/
│   │   ├── app.go                  # Charm Bubble Tea root model and loop
│   │   ├── viewport.go             # Virtualized rows renderer
│   │   ├── details.go              # Structured JSON inspector modal
│   │   ├── styles.go               # Lip Gloss theme, deterministic service colors
│   │   └── keybinds.go             # Navigation and filter hotkeys
│   └── server/
│       ├── api.go                  # HTTP routes (/api/logs, /api/stats, /api/stream)
│       ├── sse.go                  # Server-Sent Events broker with backpressure
│       └── embed.go                # //go:embed distribution for React 19 web assets
├── web/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Header.tsx          # App header, connection status, ingestion rate
│   │   │   ├── LogStreamTable.tsx  # Virtualized live log table with color badges
│   │   │   ├── ServiceFilterBar.tsx# Dynamic service pills with counts and toggles
│   │   │   ├── QueryInput.tsx      # Query bar with presets, syntax hints, history
│   │   │   ├── TraceWaterfall.tsx  # Interactive trace sequence timeline
│   │   │   ├── ErrorClusters.tsx   # Aggregated error signatures card
│   │   │   └── LogDetailModal.tsx  # JSON tree inspector with copy buttons
│   │   ├── hooks/
│   │   │   ├── useLogStream.ts     # SSE connection, buffer maintenance, stats
│   │   │   └── useDebounce.ts      # Input debouncing
│   │   ├── types/
│   │   │   └── log.ts              # TypeScript definitions matching Go models
│   │   ├── utils/
│   │   │   ├── colors.ts           # Consistent hash-based service colors
│   │   │   └── format.ts           # Timestamp and duration formatters
│   │   ├── App.tsx                 # Root application shell
│   │   ├── index.css               # Tailwind CSS rules
│   │   └── main.tsx                # React 19 DOM entry point
│   ├── index.html
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   └── tsconfig.json
├── docs/
│   ├── ARCHITECTURE.md
│   ├── QUERY_SYNTAX.md
│   └── BENCHMARKS.md
├── .gitignore
├── LICENSE
├── go.mod
├── go.sum
└── README.md
```

---

## 3. Detailed Step-by-Step Implementation Roadmap

### Task 1: Repository Setup, Metadata, and Portfolio Sync
- [x] Subtask 1.1: Initialize local git repository, `.gitignore`, and `LICENSE`.
- [x] Subtask 1.2: Author comprehensive documentation files (`README.md`, `ARCHITECTURE.md`, `QUERY_SYNTAX.md`, `BENCHMARKS.md`) with humanizer discipline.
- [x] Subtask 1.3: Create public GitHub repository `alexandrmotologa/timberlens` and apply 8-15 topics across 4 dimensions.
- [x] Subtask 1.4: Execute `mtlg-portfolio-sync` Phase 1 (register DEV badge in both portfolio projects.json files, run sync and compare, commit and push).

### Task 2: Core Data Models & Low-Memory Interning Engine
- [x] Subtask 2.1: Initialize Go module `github.com/alexandrmotologa/timberlens`.
- [x] Subtask 2.2: Implement `internal/model/record.go` defining `LogRecord`, `SeverityLevel`, raw payload representation, structured attributes map, and trace context.
- [x] Subtask 2.3: Implement `internal/model/intern.go` with thread-safe string interning pool for field names, service names, and common string values to keep memory under 25MB.
- [x] Subtask 2.4: Write comprehensive unit tests for `model` and interning efficiency.

### Task 3: Ingestion Subsystem
- [x] Subtask 3.1: Implement `internal/ingest/stdin.go` using a reusable byte buffer scanning STDIN continuously without allocation bottlenecks or pipe drops.
- [x] Subtask 3.2: Implement `internal/ingest/file.go` with multi-file tailing and rotation detection.
- [x] Subtask 3.3: Implement `internal/ingest/syslog.go` for optional UDP/TCP Syslog socket ingestion.
- [x] Subtask 3.4: Implement `internal/ingest/otlp.go` for receiving OpenTelemetry HTTP log payloads.
- [x] Subtask 3.5: Implement `internal/ingest/generator.go` producing realistic multi-service microservice logs (auth, api, payments, worker, postgres) with traces, errors, and stack traces for testing and demonstration.
- [x] Subtask 3.6: Write unit tests verifying ingestion throughput and pipeline handling.

### Task 4: Auto-Detection & Format Parser Engine
- [x] Subtask 4.1: Implement `internal/parser/ansi.go` to strip ANSI escape codes while extracting terminal color metadata.
- [x] Subtask 4.2: Implement `internal/parser/detector.go` to sniff incoming lines and dispatch to JSON, logfmt, Syslog, or plain text parser.
- [x] Subtask 4.3: Implement `internal/parser/json.go` with field alias normalization (levels: `level`, `lvl`, `severity`; messages: `message`, `msg`; timestamps: `time`, `ts`, `@timestamp`; services: `service`, `app`, `name`; trace: `trace_id`, `traceId`, `correlation_id`).
- [x] Subtask 4.4: Implement `internal/parser/logfmt.go` with zero-allocation key-value scanner supporting quotes and escaped strings.
- [x] Subtask 4.5: Implement `internal/parser/syslog.go` supporting RFC 5424 and RFC 3164 formats.
- [x] Subtask 4.6: Write unit tests covering diverse log formats and edge cases.

### Task 5: High-Capacity Lock-Free Ring Buffer & Query Engine
- [x] Subtask 5.1: Implement `internal/buffer/ring.go` allocating a contiguous slice of 1,000,000 records with atomic sequence indexing for constant memory usage.
- [x] Subtask 5.2: Implement `internal/buffer/index.go` maintaining fast in-memory inverted indices for services, severity levels, and trace IDs.
- [x] Subtask 5.3: Implement `internal/buffer/query.go` supporting free-text, field matches (`service:`, `level:`, `trace_id:`), relational comparisons (`level>=warn`, `latency_ms>500`), and boolean combinations.
- [x] Subtask 5.4: Write unit and benchmark tests asserting <1.5s ingestion for 500,000 lines and sub-millisecond query evaluation.

### Task 6: Stack Trace Cleaning & Error Signature Aggregator
- [x] Subtask 6.1: Implement `internal/stacktrace/rules.go` with regex patterns distinguishing user code from library internals in Node.js, Go, Python, and Java.
- [x] Subtask 6.2: Implement `internal/stacktrace/cleaner.go` to detect multi-line stack traces, fold library frames, and extract the root cause line.
- [x] Subtask 6.3: Implement `internal/cluster/signature.go` to compute deterministic hashes of recurring error messages, tracking occurrence rates and first/last seen timestamps.
- [x] Subtask 6.4: Write unit tests for stack trace folding and signature clustering.

### Task 7: Charm Bubble Tea Terminal TUI
- [x] Subtask 7.1: Implement `internal/tui/styles.go` with Lip Gloss palettes, severity coloring, and deterministic service hash coloring.
- [x] Subtask 7.2: Implement `internal/tui/viewport.go` with virtualized log row tailing, auto-follow, and pause controls.
- [x] Subtask 7.3: Implement `internal/tui/details.go` providing an interactive modal to inspect structured JSON fields and expanded stack traces.
- [x] Subtask 7.4: Implement `internal/tui/app.go` orchestrating keyboard shortcuts (`/`, `f`, `p`, `c`, `Enter`, `1`-`9`, `e`, `q`).
- [x] Subtask 7.5: Verify TUI stability and keyboard navigation.

### Task 8: Embedded HTTP & SSE Server
- [x] Subtask 8.1: Implement `internal/server/sse.go` supporting concurrent SSE connections with client buffering and connection cleanup.
- [x] Subtask 8.2: Implement `internal/server/api.go` exposing `/api/logs` (paginated querying), `/api/stats` (ingestion rate, memory, service counts), `/api/services`, and `/api/traces/{traceId}`.
- [x] Subtask 8.3: Implement `internal/server/embed.go` to serve embedded static files from `web/dist` with fallback to `index.html`.
- [x] Subtask 8.4: Write tests for HTTP endpoints and SSE broadcasting.

### Task 9: Modern React 19 Web Studio
- [x] Subtask 9.1: Initialize React 19 + TypeScript + Vite + Tailwind CSS project in `web/`.
- [x] Subtask 9.2: Implement `useLogStream` hook with auto-reconnecting EventSource, circular ring buffer on client, and query filtering.
- [x] Subtask 9.3: Build `Header` component showing live ingestion speed (lines/sec), memory consumption, and connection pill.
- [x] Subtask 9.4: Build `ServiceFilterBar` with deterministic color tags, log counts, and toggle filters.
- [x] Subtask 9.5: Build `QueryInput` with instant search, syntax hints, and query presets.
- [x] Subtask 9.6: Build `LogStreamTable` virtualized table rendering severity badges, service pills, timestamps, messages, and stack trace fold indicators.
- [x] Subtask 9.7: Build `TraceWaterfall` drawer rendering distributed request flow across services.
- [x] Subtask 9.8: Build `LogDetailModal` with structured JSON explorer and 1-click clipboard copy.
- [x] Subtask 9.9: Build `ErrorClusters` card summarizing top recurring exceptions.
- [x] Subtask 9.10: Compile production build into `web/dist/` ready for Go embedding.

### Task 10: CLI Integration & End-to-End Verification
- [x] Subtask 10.1: Assemble `cmd/timberlens/main.go` handling command-line flags (`--web-port`, `--no-tui`, `--buffer-size`, `--file`, `--syslog`, `--otlp`, `demo`).
- [x] Subtask 10.2: Verify complete pipeline by running `timberlens demo` and checking both TUI and Web Studio on `http://localhost:9900`.
- [x] Subtask 10.3: Execute comprehensive benchmark tests (`go test -bench=. ./...`).

### Task 11: Visual Assets & Portfolio Sync Phase 2
- [x] Subtask 11.1: Capture authentic screenshots of both the Terminal TUI and the React 19 Web Studio in action.
- [x] Subtask 11.2: Generate animated demo GIF (`docs/images/timberlens_demo.gif`) and embed into `README.md`.
- [x] Subtask 11.3: Execute `mtlg-portfolio-sync` Phase 2 (copy visual assets, update `project_images_map.json`, update `projects.json` to LIVE, recompile portfolios, preview and broadcast to X).

---

## 4. Phase 2: Advanced Usability & Productivity Roadmap (10 Features)

### Task 12: Custom Regex Parser & HTTP cURL Generator
- [x] Subtask 12.1: Implement named-group regex parser in `internal/parser/regex.go` supporting `--format-regex`.
- [x] Subtask 12.2: Implement HTTP request attribute detection and cURL command synthesis in `internal/model/curl.go`.
- [x] Subtask 12.3: Unit tests for custom regex parsing and cURL synthesis.

### Task 13: Time-Series Velocity & Anomaly Histogram
- [x] Subtask 13.1: Implement rolling time-bucketed aggregation in `internal/buffer/histogram.go` (60 buckets for total, info, warn, error).
- [x] Subtask 13.2: Expose `/api/histogram` endpoint in `internal/server/api.go`.
- [x] Subtask 13.3: Unit tests verifying bucket rotation and error spike tracking.

### Task 14: Terminal TUI Split-View, Vim Navigation & Mouse Support
- [x] Subtask 14.1: Implement side-by-side Split View in `internal/tui/app.go` (`v` / `Tab` toggles live stream on left, live inspector on right).
- [x] Subtask 14.2: Implement Vim-style search navigation (`n` next match, `N` previous match).
- [x] Subtask 14.3: Add mouse support (`tea.WithMouseCellMotion()`) for wheel scrolling and row click selection.
- [x] Subtask 14.4: Add `c` shortcut to copy cURL replay command.

### Task 15: React 19 Web Studio Usability Suite
- [x] Subtask 15.1: Build `VelocityHistogram.tsx` component in Header with interactive stacked bars and zoom filter.
- [x] Subtask 15.2: Build 1-click "Copy as cURL" button in `LogDetailModal.tsx` and row actions.
- [x] Subtask 15.3: Build `TraceCompareModal.tsx` for side-by-side diffing of two traces.
- [x] Subtask 15.4: Build `SnapshotExporter.ts` generating a self-contained, offline interactive HTML report.
- [x] Subtask 15.5: Build `AlertManager.tsx` with Web Audio procedural synthesizer chimes + Notification API.
- [x] Subtask 15.6: Build `ColumnPicker.tsx` allowing customizable visible columns in the log stream table.
- [x] Subtask 15.7: Build dynamic keyword and regex text highlighting in log messages.

### Task 16: Verification, Visual Assets & Portfolio Re-Sync
- [x] Subtask 16.1: Run all unit and benchmark tests across Go and Web.
- [x] Subtask 16.2: Rebuild embedded production bundle into `bin/timberlens.exe`.
- [x] Subtask 16.3: Capture updated screenshots showcasing Split View TUI, Velocity Histogram, Trace Compare, and Snapshot Exporter.
- [x] Subtask 16.4: Regenerate demo GIF and update documentation.
- [x] Subtask 16.5: Sync portfolio sites and commit to GitHub.

