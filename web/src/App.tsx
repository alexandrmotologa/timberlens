import { useState, useEffect, useMemo } from 'react';
import { Header } from './components/Header';
import { ServiceFilterBar } from './components/ServiceFilterBar';
import { QueryInput } from './components/QueryInput';
import { LogStreamTable } from './components/LogStreamTable';
import { TraceWaterfall } from './components/TraceWaterfall';
import { TraceCompareModal } from './components/TraceCompareModal';
import { ErrorClusters } from './components/ErrorClusters';
import { LogDetailModal } from './components/LogDetailModal';
import { VelocityHistogram } from './components/VelocityHistogram';
import { useLogStream } from './hooks/useLogStream';
import { LogRecord, SystemStats, ColumnVisibility, HistogramBucket } from './types/log';
import { exportHtmlSnapshot } from './utils/snapshot';

const DEFAULT_COLUMNS: ColumnVisibility = {
  timestamp: true,
  level: true,
  service: true,
  traceId: true,
  latency: true,
  message: true,
};

export default function App() {
  const [query, setQuery] = useState('');
  const [selectedService, setSelectedService] = useState('');
  const [stats, setStats] = useState<SystemStats | null>(null);

  // Column visibility state with localStorage persistence
  const [columns, setColumns] = useState<ColumnVisibility>(() => {
    try {
      const saved = localStorage.getItem('timberlens_columns');
      return saved ? JSON.parse(saved) : DEFAULT_COLUMNS;
    } catch {
      return DEFAULT_COLUMNS;
    }
  });

  const handleColumnsChange = (newCols: ColumnVisibility) => {
    setColumns(newCols);
    try {
      localStorage.setItem('timberlens_columns', JSON.stringify(newCols));
    } catch {
      // ignore
    }
  };

  // Modals & Drawers state
  const [selectedLog, setSelectedLog] = useState<LogRecord | null>(null);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [showClusters, setShowClusters] = useState(false);
  const [showTraceCompare, setShowTraceCompare] = useState(false);
  const [compareTraceA, setCompareTraceA] = useState('');
  const [compareTraceB, setCompareTraceB] = useState('');

  // Compute effective query combining search bar + service filter
  const effectiveQuery = selectedService
    ? query
      ? `service:${selectedService} ${query}`
      : `service:${selectedService}`
    : query;

  const { logs, isConnected, isPaused, linesPerSec, togglePause, clearLogs } =
    useLogStream({ query: effectiveQuery, maxBuffer: 2000 });

  // Extract distinct recent trace IDs for quick comparison
  const recentTraces = useMemo(() => {
    const set = new Set<string>();
    for (let i = logs.length - 1; i >= 0; i--) {
      const tid = logs[i].traceId;
      if (tid) {
        set.add(tid);
        if (set.size >= 12) break;
      }
    }
    return Array.from(set);
  }, [logs]);

  // Poll system stats every 2 seconds
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('q')) setQuery(params.get('q')!);
    if (params.get('service')) setSelectedService(params.get('service')!);
    if (params.get('trace')) setSelectedTraceId(params.get('trace')!);
    if (params.get('clusters') === 'true') setShowClusters(true);

    const fetchStats = async () => {
      try {
        const res = await fetch('/api/stats');
        if (res.ok) {
          const data: SystemStats = await res.json();
          setStats(data);
        }
      } catch {
        // ignore fetch stats error
      }
    };

    fetchStats();
    const interval = setInterval(fetchStats, 2000);
    return () => clearInterval(interval);
  }, []);

  // Auto-select error log if ?modal=error requested, or deep-link trace compare
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('compare') === 'true') {
      setShowTraceCompare(true);
      setCompareTraceA(params.get('traceA') || 'trc-b914fd1e');
      setCompareTraceB(params.get('traceB') || 'trc-c821ea34');
    }

    if (params.get('modal') === 'error' && !selectedLog) {
      if (logs.length > 0) {
        const err = logs.find((l) => l.levelStr === 'ERROR' || l.stackTrace);
        if (err) setSelectedLog(err);
      } else {
        setSelectedLog({
          id: 1042,
          timestamp: new Date().toISOString(),
          level: 5,
          levelStr: 'ERROR',
          service: 'payment-worker',
          traceId: 'trc-b914fd1e',
          spanId: 'spn-4f91a2',
          message: 'Unhandled database connection refusal during transaction commit',
          raw: '{"error":"ECONNREFUSED","latency_ms":1250,"level":"ERROR","msg":"Unhandled database connection refusal during transaction commit","retryCount":3,"service":"payment-worker","trace_id":"trc-b914fd1e"}',
          format: 1,
          formatStr: 'JSON',
          fields: {
            method: 'POST',
            path: '/v1/checkout/charge',
            host: 'api.timberlens.local',
            headers: {
              'content-type': 'application/json',
              'x-request-id': 'req-98421a',
            },
            body: {
              amount: 4999,
              currency: 'USD',
              customer_id: 'cust_88231',
            },
            error: 'ECONNREFUSED',
            latency_ms: 1250,
            retryCount: 3,
          },
          numeric: { latency_ms: 1250 },
          stackTrace: {
            raw: 'Error: connect ECONNREFUSED 10.0.12.44:5432\n    at TCPConnectWrap.afterConnect (node:net:1494:16)\n    at /app/node_modules/pg-pool/index.js:45:11\n    at /app/node_modules/pg/lib/client.js:120:5\n    at DatabasePool.acquire (/app/src/db/pool.ts:48:19)\n    at OrderRepository.createOrder (/app/src/repositories/orders.ts:88:27)\n    at CheckoutHandler.execute (/app/src/handlers/checkout.ts:34:12)',
            language: 'Node.js',
            rootCause: 'Error: connect ECONNREFUSED 10.0.12.44:5432',
            totalFrames: 6,
            hiddenFrames: 3,
            frames: [
              { file: 'node:net', line: 1494, function: 'TCPConnectWrap.afterConnect', isVendor: true, raw: '' },
              { file: '/app/node_modules/pg-pool/index.js', line: 45, isVendor: true, raw: '' },
              { file: '/app/node_modules/pg/lib/client.js', line: 120, isVendor: true, raw: '' },
              { file: '/app/src/db/pool.ts', line: 48, function: 'DatabasePool.acquire', isVendor: false, raw: '' },
              { file: '/app/src/repositories/orders.ts', line: 88, function: 'OrderRepository.createOrder', isVendor: false, raw: '' },
              { file: '/app/src/handlers/checkout.ts', line: 34, function: 'CheckoutHandler.execute', isVendor: false, raw: '' },
            ],
          },
        });
      }
    }
  }, [logs, selectedLog]);

  const handleStartTraceCompare = (traceId: string) => {
    setCompareTraceA(traceId);
    // Find another recent trace if available
    const other = recentTraces.find((t) => t !== traceId) || '';
    setCompareTraceB(other);
    setShowTraceCompare(true);
  };

  const handleExportSnapshot = () => {
    exportHtmlSnapshot(logs, stats, effectiveQuery);
  };

  const handleHistogramIntervalSelect = (bucket: HistogramBucket) => {
    // If bucket had errors, add level>=warn to query
    if (bucket.error > 0) {
      setQuery((prev) => (prev ? `${prev} level>=error` : 'level>=error'));
    }
  };

  return (
    <div className="h-screen flex flex-col bg-[#090D16] text-slate-100 overflow-hidden font-sans">
      {/* 1. App Header */}
      <Header
        stats={stats}
        linesPerSec={linesPerSec}
        isConnected={isConnected}
        isPaused={isPaused}
        columns={columns}
        onColumnsChange={handleColumnsChange}
        onTogglePause={togglePause}
        onClear={clearLogs}
        onOpenClusters={() => setShowClusters(true)}
        onOpenTraceCompare={() => setShowTraceCompare(true)}
        onExportSnapshot={handleExportSnapshot}
      />

      {/* 2. Realtime Traffic Velocity Histogram */}
      <VelocityHistogram onSelectInterval={handleHistogramIntervalSelect} />

      {/* 3. Microservice Filter Pills */}
      <ServiceFilterBar
        services={stats?.services || {}}
        selectedService={selectedService}
        onSelectService={(svc) => setSelectedService(svc)}
      />

      {/* 4. Query Bar */}
      <QueryInput value={query} onChange={(q) => setQuery(q)} />

      {/* 5. Live Stream Table */}
      <LogStreamTable
        logs={logs}
        isPaused={isPaused}
        columns={columns}
        searchQuery={query}
        onSelectLog={(log) => setSelectedLog(log)}
        onSelectTrace={(traceId) => setSelectedTraceId(traceId)}
      />

      {/* 6. Modals & Drawers */}
      <TraceWaterfall
        traceId={selectedTraceId}
        onClose={() => setSelectedTraceId(null)}
        onSelectLog={(log) => setSelectedLog(log)}
      />

      <TraceCompareModal
        isOpen={showTraceCompare}
        initialTraceA={compareTraceA}
        initialTraceB={compareTraceB}
        recentTraces={recentTraces}
        onClose={() => setShowTraceCompare(false)}
        onSelectLog={(log) => {
          setSelectedLog(log);
          setShowTraceCompare(false);
        }}
      />

      <ErrorClusters
        isOpen={showClusters}
        onClose={() => setShowClusters(false)}
        onApplyFilter={(q) => setQuery(q)}
      />

      <LogDetailModal
        log={selectedLog}
        onClose={() => setSelectedLog(null)}
        onSelectTrace={(traceId) => setSelectedTraceId(traceId)}
        onCompareTrace={handleStartTraceCompare}
      />
    </div>
  );
}
