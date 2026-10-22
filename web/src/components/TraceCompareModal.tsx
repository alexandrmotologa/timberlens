import React, { useState, useEffect } from 'react';
import { X, GitCompare, ArrowRight, AlertTriangle, Clock, Server } from 'lucide-react';
import { LogRecord } from '../types/log';
import { getServiceColor, getLevelBadge } from '../utils/colors';
import { formatTimestamp } from '../utils/format';

interface TraceCompareModalProps {
  isOpen: boolean;
  initialTraceA?: string;
  initialTraceB?: string;
  recentTraces: string[];
  onClose: () => void;
  onSelectLog: (log: LogRecord) => void;
}

interface AnalyzedTrace {
  traceId: string;
  records: LogRecord[];
  durationMs: number;
  services: string[];
  errorCount: number;
}

export const TraceCompareModal: React.FC<TraceCompareModalProps> = ({
  isOpen,
  initialTraceA = '',
  initialTraceB = '',
  recentTraces,
  onClose,
  onSelectLog,
}) => {
  const [traceIdA, setTraceIdA] = useState(initialTraceA);
  const [traceIdB, setTraceIdB] = useState(initialTraceB);
  const [analyzedA, setAnalyzedA] = useState<AnalyzedTrace | null>(null);
  const [analyzedB, setAnalyzedB] = useState<AnalyzedTrace | null>(null);
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  useEffect(() => {
    if (initialTraceA) setTraceIdA(initialTraceA);
    if (initialTraceB) setTraceIdB(initialTraceB);
  }, [initialTraceA, initialTraceB]);

  const analyzeRecords = (traceId: string, records: LogRecord[]): AnalyzedTrace => {
    if (!records || records.length === 0) {
      return {
        traceId,
        records: [],
        durationMs: 0,
        services: [],
        errorCount: 0,
      };
    }

    const timestamps = records
      .map((r) => new Date(r.timestamp).getTime())
      .filter((t) => !isNaN(t));

    const minTime = timestamps.length ? Math.min(...timestamps) : 0;
    const maxTime = timestamps.length ? Math.max(...timestamps) : 0;
    const durationMs = maxTime > minTime ? maxTime - minTime : 0;

    const services = Array.from(new Set(records.map((r) => r.service)));
    const errorCount = records.filter(
      (r) => r.levelStr === 'ERROR' || r.levelStr === 'FATAL'
    ).length;

    return {
      traceId,
      records,
      durationMs,
      services,
      errorCount,
    };
  };

  const handleCompare = async () => {
    if (!traceIdA.trim() || !traceIdB.trim()) {
      setErrorMsg('Please enter both Trace IDs to compare.');
      return;
    }
    setErrorMsg(null);
    setLoading(true);

    try {
      const [resA, resB] = await Promise.all([
        fetch(`/api/traces/${encodeURIComponent(traceIdA.trim())}`),
        fetch(`/api/traces/${encodeURIComponent(traceIdB.trim())}`),
      ]);

      const dataA = resA.ok ? await resA.json() : { records: [] };
      const dataB = resB.ok ? await resB.json() : { records: [] };

      setAnalyzedA(analyzeRecords(traceIdA.trim(), dataA.records || []));
      setAnalyzedB(analyzeRecords(traceIdB.trim(), dataB.records || []));
    } catch {
      setErrorMsg('Failed to fetch trace comparison data.');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  const durationDelta =
    analyzedA && analyzedB ? analyzedB.durationMs - analyzedA.durationMs : 0;
  const isBRegression = durationDelta > 0;

  return (
    <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-150">
      <div className="w-full max-w-6xl bg-surface border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
        {/* Header */}
        <div className="p-4 border-b border-border flex items-center justify-between bg-surface">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
              <GitCompare className="w-4 h-4" />
            </div>
            <div>
              <h2 className="font-bold text-slate-100 text-sm">Trace Diff & Comparison</h2>
              <p className="text-xs text-slate-400">
                Compare execution paths, latency shifts, and anomalous service divergence side-by-side.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-surfaceHover border border-border transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Trace Inputs Bar */}
        <div className="p-4 bg-[#090D16] border-b border-border flex flex-wrap items-center gap-4">
          <div className="flex-1 min-w-[240px]">
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Baseline Trace (Trace A):
            </label>
            <input
              type="text"
              value={traceIdA}
              onChange={(e) => setTraceIdA(e.target.value)}
              placeholder="e.g. req-abc-123"
              className="w-full px-3 py-1.5 rounded-lg bg-surface border border-border text-slate-100 font-mono text-xs focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="text-slate-500 font-bold hidden sm:block pt-5">vs</div>

          <div className="flex-1 min-w-[240px]">
            <label className="block text-[11px] font-mono text-slate-400 mb-1">
              Comparison Trace (Trace B):
            </label>
            <input
              type="text"
              value={traceIdB}
              onChange={(e) => setTraceIdB(e.target.value)}
              placeholder="e.g. req-xyz-789"
              className="w-full px-3 py-1.5 rounded-lg bg-surface border border-border text-slate-100 font-mono text-xs focus:outline-none focus:border-indigo-500"
            />
          </div>

          <div className="pt-5">
            <button
              onClick={handleCompare}
              disabled={loading || !traceIdA || !traceIdB}
              className="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-medium text-xs transition-colors flex items-center space-x-1.5"
            >
              <span>{loading ? 'Comparing...' : 'Compare Traces'}</span>
            </button>
          </div>
        </div>

        {/* Quick select recent traces pills */}
        {recentTraces.length > 0 && (
          <div className="px-4 py-2 bg-surface/50 border-b border-border flex items-center space-x-2 overflow-x-auto text-[11px]">
            <span className="text-slate-500 shrink-0 font-medium">Recent traces:</span>
            {recentTraces.slice(0, 6).map((t) => (
              <button
                key={t}
                onClick={() => {
                  if (!traceIdA) setTraceIdA(t);
                  else if (!traceIdB) setTraceIdB(t);
                  else setTraceIdB(t);
                }}
                className="px-2 py-0.5 rounded bg-surfaceHover border border-border text-slate-300 hover:text-indigo-400 font-mono text-[10px] shrink-0"
              >
                {t.slice(0, 10)}
              </button>
            ))}
          </div>
        )}

        {/* Error Notification */}
        {errorMsg && (
          <div className="p-3 bg-rose-500/10 border-b border-rose-500/20 text-rose-400 text-xs flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        {/* Comparison Summary Banner */}
        {analyzedA && analyzedB && (
          <div className="p-4 bg-surface border-b border-border grid grid-cols-1 md:grid-cols-3 gap-4 text-xs font-mono">
            {/* Latency Comparison */}
            <div className="p-3 rounded-lg bg-[#090D16] border border-border">
              <div className="text-slate-400 flex items-center space-x-1 mb-1">
                <Clock className="w-3.5 h-3.5 text-indigo-400" />
                <span>Execution Duration Delta</span>
              </div>
              <div className="flex items-baseline space-x-2">
                <span className="text-lg font-bold text-slate-100">
                  {analyzedA.durationMs}ms
                </span>
                <ArrowRight className="w-3 h-3 text-slate-500" />
                <span className="text-lg font-bold text-slate-100">
                  {analyzedB.durationMs}ms
                </span>
                <span
                  className={`text-xs font-bold px-1.5 py-0.5 rounded ${
                    isBRegression
                      ? 'bg-rose-500/10 text-rose-400 border border-rose-500/30'
                      : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                  }`}
                >
                  {isBRegression ? `+${durationDelta}ms slower` : `${durationDelta}ms faster`}
                </span>
              </div>
            </div>

            {/* Service Hops */}
            <div className="p-3 rounded-lg bg-[#090D16] border border-border">
              <div className="text-slate-400 flex items-center space-x-1 mb-1">
                <Server className="w-3.5 h-3.5 text-sky-400" />
                <span>Service Footprint</span>
              </div>
              <div className="flex items-baseline space-x-2">
                <span className="font-bold text-slate-100">
                  {analyzedA.services.length} services
                </span>
                <ArrowRight className="w-3 h-3 text-slate-500" />
                <span className="font-bold text-slate-100">
                  {analyzedB.services.length} services
                </span>
              </div>
              <div className="text-[10px] text-slate-400 truncate mt-1">
                {analyzedB.services.filter((s) => !analyzedA.services.includes(s)).length > 0
                  ? `New services in B: ${analyzedB.services
                      .filter((s) => !analyzedA.services.includes(s))
                      .join(', ')}`
                  : 'Identical service set'}
              </div>
            </div>

            {/* Error Status */}
            <div className="p-3 rounded-lg bg-[#090D16] border border-border">
              <div className="text-slate-400 flex items-center space-x-1 mb-1">
                <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
                <span>Error Count Comparison</span>
              </div>
              <div className="flex items-baseline space-x-2">
                <span
                  className={analyzedA.errorCount > 0 ? 'text-rose-400 font-bold' : 'text-emerald-400'}
                >
                  {analyzedA.errorCount} errors
                </span>
                <ArrowRight className="w-3 h-3 text-slate-500" />
                <span
                  className={analyzedB.errorCount > 0 ? 'text-rose-400 font-bold' : 'text-emerald-400'}
                >
                  {analyzedB.errorCount} errors
                </span>
              </div>
            </div>
          </div>
        )}

        {/* Side-by-Side Waterfall List */}
        <div className="flex-1 overflow-y-auto grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border bg-[#090D16] font-mono text-xs">
          {/* Left Column: Trace A */}
          <div className="p-4 space-y-2 overflow-y-auto">
            <div className="font-bold text-slate-300 pb-2 border-b border-border flex items-center justify-between">
              <span>Trace A ({analyzedA?.records.length || 0} events)</span>
              <span className="text-[10px] text-indigo-400">{traceIdA}</span>
            </div>

            {!analyzedA ? (
              <div className="text-center py-12 text-slate-500">
                Enter Trace A ID to display execution events.
              </div>
            ) : analyzedA.records.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No log records found for Trace A in memory.
              </div>
            ) : (
              analyzedA.records.map((log) => {
                const svcColor = getServiceColor(log.service);
                return (
                  <div
                    key={log.id}
                    onClick={() => onSelectLog(log)}
                    className="p-2 rounded border border-border/70 hover:border-sky-500 bg-surface/50 cursor-pointer space-y-1 transition-colors"
                  >
                    <div className="flex items-center justify-between text-[10px]">
                      <span className="text-slate-400">{formatTimestamp(log.timestamp)}</span>
                      <span className={`px-1.5 py-0.2 rounded ${getLevelBadge(log.levelStr)}`}>
                        {log.levelStr}
                      </span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span
                        className={`px-1.5 py-0.2 rounded border text-[10px] ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}
                      >
                        {log.service}
                      </span>
                    </div>
                    <div className="text-slate-300 text-[11px] truncate">
                      {log.message || log.raw}
                    </div>
                  </div>
                );
              })
            )}
          </div>

          {/* Right Column: Trace B */}
          <div className="p-4 space-y-2 overflow-y-auto">
            <div className="font-bold text-slate-300 pb-2 border-b border-border flex items-center justify-between">
              <span>Trace B ({analyzedB?.records.length || 0} events)</span>
              <span className="text-[10px] text-indigo-400">{traceIdB}</span>
            </div>

            {!analyzedB ? (
              <div className="text-center py-12 text-slate-500">
                Enter Trace B ID to display execution events.
              </div>
            ) : analyzedB.records.length === 0 ? (
              <div className="text-center py-12 text-slate-500">
                No log records found for Trace B in memory.
              </div>
            ) : (
              analyzedB.records.map((log) => {
                const svcColor = getServiceColor(log.service);
                return (
                  <div
                    key={log.id}
                    onClick={() => onSelectLog(log)}
                    className="p-2 rounded border border-border/70 hover:border-sky-500 bg-surface/50 cursor-pointer space-y-1 transition-colors"
                  >
                    <div className="flex items-center justify-between text-[10px]">
                      <span className="text-slate-400">{formatTimestamp(log.timestamp)}</span>
                      <span className={`px-1.5 py-0.2 rounded ${getLevelBadge(log.levelStr)}`}>
                        {log.levelStr}
                      </span>
                    </div>
                    <div className="flex items-center space-x-1.5">
                      <span
                        className={`px-1.5 py-0.2 rounded border text-[10px] ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}
                      >
                        {log.service}
                      </span>
                    </div>
                    <div className="text-slate-300 text-[11px] truncate">
                      {log.message || log.raw}
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
