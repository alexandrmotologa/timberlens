import { useState, useEffect } from 'react';
import { Header } from './components/Header';
import { ServiceFilterBar } from './components/ServiceFilterBar';
import { QueryInput } from './components/QueryInput';
import { LogStreamTable } from './components/LogStreamTable';
import { TraceWaterfall } from './components/TraceWaterfall';
import { ErrorClusters } from './components/ErrorClusters';
import { LogDetailModal } from './components/LogDetailModal';
import { useLogStream } from './hooks/useLogStream';
import { LogRecord, SystemStats } from './types/log';

export default function App() {
  const [query, setQuery] = useState('');
  const [selectedService, setSelectedService] = useState('');
  const [stats, setStats] = useState<SystemStats | null>(null);

  // Modals & Drawers state
  const [selectedLog, setSelectedLog] = useState<LogRecord | null>(null);
  const [selectedTraceId, setSelectedTraceId] = useState<string | null>(null);
  const [showClusters, setShowClusters] = useState(false);

  // Compute effective query combining search bar + service filter
  const effectiveQuery = selectedService
    ? query
      ? `service:${selectedService} ${query}`
      : `service:${selectedService}`
    : query;

  const { logs, isConnected, isPaused, linesPerSec, togglePause, clearLogs } =
    useLogStream({ query: effectiveQuery, maxBuffer: 1500 });

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

  // Auto-select error log if ?modal=error requested
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get('modal') === 'error' && logs.length > 0 && !selectedLog) {
      const err = logs.find((l) => l.levelStr === 'ERROR' || l.stackTrace);
      if (err) setSelectedLog(err);
    }
  }, [logs, selectedLog]);

  return (
    <div className="h-screen flex flex-col bg-[#090D16] text-slate-100 overflow-hidden font-sans">
      {/* 1. App Header */}
      <Header
        stats={stats}
        linesPerSec={linesPerSec}
        isConnected={isConnected}
        isPaused={isPaused}
        onTogglePause={togglePause}
        onClear={clearLogs}
        onOpenClusters={() => setShowClusters(true)}
      />

      {/* 2. Microservice Filter Pills */}
      <ServiceFilterBar
        services={stats?.services || {}}
        selectedService={selectedService}
        onSelectService={(svc) => setSelectedService(svc)}
      />

      {/* 3. Query Bar */}
      <QueryInput value={query} onChange={(q) => setQuery(q)} />

      {/* 4. Live Stream Table */}
      <LogStreamTable
        logs={logs}
        isPaused={isPaused}
        onSelectLog={(log) => setSelectedLog(log)}
        onSelectTrace={(traceId) => setSelectedTraceId(traceId)}
      />

      {/* 5. Modals & Drawers */}
      <TraceWaterfall
        traceId={selectedTraceId}
        onClose={() => setSelectedTraceId(null)}
        onSelectLog={(log) => setSelectedLog(log)}
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
      />
    </div>
  );
}
