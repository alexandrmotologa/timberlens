# Query Syntax Guide

TimberLens provides an expressive filter syntax for analyzing structured log streams.

## Free-Text Search

Any bare word matches anywhere in the raw log line or structured message:

```
database connection failed
```

To search for phrases containing whitespace, enclose terms in double quotes:

```
"connection reset by peer"
```

## Field Filters

Field queries match key-value attributes extracted during ingestion:

| Filter | Example | Description |
|---|---|---|
| `service:` | `service:auth-api` | Exact match on service or component name |
| `level:` | `level:error` | Exact match on log severity |
| `trace:` / `trace_id:` | `trace_id:c4b8109f` | Filter by distributed trace identifier |
| `path:` | `path:"/v1/checkout"` | Match HTTP request path |
| `status:` | `status:500` | Match HTTP status code |

## Severity Comparisons

Severity levels support inequality operators:

- `level:error`: Matches ERROR and FATAL
- `level>=warn`: Matches WARN, ERROR, and FATAL
- `level<=info`: Matches DEBUG and INFO
- `level!=debug`: Excludes DEBUG entries

Valid levels in ascending order: `DEBUG` < `INFO` < `WARN` < `ERROR` < `FATAL`.

## Numeric Comparisons

Fields containing numeric values (such as latencies, byte sizes, status codes) support relational operators:

- `latency_ms>500`: Requests taking longer than 500ms
- `duration<50`: Fast operations under 50ms
- `status>=400`: All HTTP client and server errors

## Logical Combinations

Separate multiple expressions with whitespace to perform an implicit `AND`:

```
service:billing level:error latency_ms>1000
```

Negate expressions using a leading exclamation mark or minus sign:

```
service:worker !level:debug
```
