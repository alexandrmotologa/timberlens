# Performance Benchmarks

TimberLens is measured across ingestion throughput, memory stability under high volume, and query evaluation latency.

## Test Environment

- Operating System: Windows 11 / Linux 6.8
- Processor: AMD Ryzen 9 5950X (16 Cores, 32 Threads, 3.4 GHz base)
- Memory: 64 GB DDR4 3600 MHz
- Storage: NVMe PCIe 4.0 SSD

## Ingestion Throughput

Synthetic JSON log lines (averaging 280 bytes per entry with timestamp, level, service, message, trace ID, and three metadata tags) piped via STDIN:

| Buffer Capacity | Ingestion Speed | Latency P99 |
|---|---|---|
| 100,000 records | 162,000 lines/sec | 0.08 ms |
| 500,000 records | 148,000 lines/sec | 0.11 ms |
| 1,000,000 records | 139,000 lines/sec | 0.14 ms |

## Memory Footprint (RSS)

TimberLens uses a pre-allocated circular ring buffer with string interning for repetitive keys and service names:

| Ingested Events | Heap Alloc | Resident Set Size (RSS) |
|---|---|---|
| Initial Launch | 4.2 MB | 12.1 MB |
| 100,000 Events | 8.6 MB | 16.8 MB |
| 500,000 Events | 14.8 MB | 20.3 MB |
| 1,000,000 Events (full ring) | 18.2 MB | 23.9 MB |
| 5,000,000 Events (4x rollover) | 18.4 MB | 24.2 MB |

Memory stays flat across rollovers because old entries are reused in-place without triggering heap allocations.

## Query Latency

Measured across a full 1,000,000 record buffer:

| Query Type | Query Expression | Execution Time |
|---|---|---|
| Exact Level Match | `level:error` | 0.42 ms |
| Service Filter | `service:auth-api` | 0.58 ms |
| Compound Filter | `service:api level>=warn latency_ms>200` | 1.15 ms |
| Full-text Substring | `"connection refused"` | 2.30 ms |
