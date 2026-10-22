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
      if (params.get('traceA')) setCompareTraceA(params.get('traceA')!);
      if (params.get('traceB')) setCompareTraceB(params.get('traceB')!);
    }

    if (params.get('modal') === 'error' && !selectedLog && logs.length > 0) {
      const err = logs.find((l) => l.levelStr === 'ERROR' || l.stackTrace);
      if (err) setSelectedLog(err);
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
