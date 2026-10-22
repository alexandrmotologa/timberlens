import React, { useEffect, useState } from 'react';
import { X, GitCommit, Clock, ArrowRight } from 'lucide-react';
import { LogRecord } from '../types/log';
import { getServiceColor, getLevelBadge } from '../utils/colors';
import { formatTimestamp } from '../utils/format';

interface TraceWaterfallProps {
  traceId: string | null;
  onClose: () => void;
  onSelectLog: (log: LogRecord) => void;
}

export const TraceWaterfall: React.FC<TraceWaterfallProps> = ({
  traceId,
  onClose,
  onSelectLog,
}) => {
  const [logs, setLogs] = useState<LogRecord[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!traceId) return;

    let ignore = false;
    setLoading(true);

    fetch(`/api/traces/${encodeURIComponent(traceId)}`)
      .then((res) => res.json())
      .then((data) => {
        if (!ignore && data.records) {
          setLogs(data.records);
        }
      })
      .catch((err) => console.error('Failed to load trace:', err))
      .finally(() => setLoading(false));

    return () => {
      ignore = true;
    };
  }, [traceId]);

  if (!traceId) return null;

  // Calculate total timeline window
  const timestamps = logs.map((l) => new Date(l.timestamp).getTime()).filter((t) => !isNaN(t));
  const minTime = timestamps.length ? Math.min(...timestamps) : 0;
  const maxTime = timestamps.length ? Math.max(...timestamps) : 0;
  const totalDurationMs = Math.max(maxTime - minTime, 1);

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex justify-end animate-in fade-in duration-150">
      <div className="w-full max-w-2xl bg-surface border-l border-border h-full flex flex-col shadow-2xl animate-in slide-in-from-right duration-200">
        {/* Header */}
        <div className="p-4 border-b border-border flex items-center justify-between bg-surface">
          <div className="flex items-center space-x-2">
            <div className="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <GitCommit className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <span className="text-sm font-semibold text-slate-100">Trace Waterfall</span>
                <span className="font-mono text-xs text-indigo-400 bg-indigo-500/10 px-2 py-0.5 rounded border border-indigo-500/20">
                  {traceId}
                </span>
              </div>
              <div className="text-xs text-slate-400 flex items-center space-x-3 mt-0.5">
                <span className="flex items-center space-x-1">
                  <Clock className="w-3 h-3 text-slate-500" />
                  <span>Duration: ~{totalDurationMs}ms</span>
                </span>
                <span>•</span>
                <span>{logs.length} correlated events</span>
              </div>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-surfaceHover border border-transparent hover:border-border transition-colors"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Trace Flow Sequence */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3 font-mono text-xs">
          {loading ? (
            <div className="p-8 text-center text-slate-500">Loading trace events...</div>
          ) : logs.length === 0 ? (
            <div className="p-8 text-center text-slate-500">No events found for this trace ID.</div>
          ) : (
            logs.map((log, idx) => {
              const logTime = new Date(log.timestamp).getTime();
              const offsetMs = Math.max(0, logTime - minTime);
              const offsetPercent = totalDurationMs > 0 ? (offsetMs / totalDurationMs) * 85 : 0;
              const svcColor = getServiceColor(log.service);

              return (
                <div
                  key={log.id}
                  onClick={() => onSelectLog(log)}
                  className="p-3 rounded-lg bg-[#090D16] border border-border hover:border-sky-500/50 cursor-pointer transition-all space-y-2 group"
                >
                  <div className="flex items-center justify-between text-[11px]">
                    <div className="flex items-center space-x-2">
                      <span className="text-slate-500">#{idx + 1}</span>
                      <span className={`px-1.5 py-0.5 rounded text-[10px] ${getLevelBadge(log.levelStr)}`}>
                        {log.levelStr}
                      </span>
                      <span className={`px-2 py-0.5 rounded font-medium border ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}>
                        {log.service}
                      </span>
                    </div>

                    <div className="text-slate-500 flex items-center space-x-2">
                      <span>+{offsetMs}ms</span>
                      <span>{formatTimestamp(log.timestamp)}</span>
                    </div>
                  </div>

                  {/* Relative Offset Timeline Bar */}
                  <div className="w-full bg-slate-800/60 h-1.5 rounded-full overflow-hidden relative">
                    <div
                      className="absolute top-0 bottom-0 w-8 rounded-full"
                      style={{
                        left: `${offsetPercent}%`,
                        backgroundColor: svcColor.hex,
                      }}
                    />
                  </div>

                  {/* Message */}
                  <div className="text-slate-300 group-hover:text-slate-100 break-words flex items-center justify-between">
                    <span>{log.message || log.raw}</span>
                    <ArrowRight className="w-3.5 h-3.5 text-slate-500 group-hover:text-sky-400 shrink-0 ml-2" />
                  </div>
                </div>
              );
            })
          )}
        </div>
      </div>
    </div>
  );
};
