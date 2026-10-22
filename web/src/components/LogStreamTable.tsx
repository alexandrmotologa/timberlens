import React, { useRef, useEffect } from 'react';
import { GitCommit, Layers, ChevronRight, AlertCircle } from 'lucide-react';
import { LogRecord } from '../types/log';
import { getServiceColor, getLevelBadge } from '../utils/colors';
import { formatTimestamp } from '../utils/format';

interface LogStreamTableProps {
  logs: LogRecord[];
  isPaused: boolean;
  onSelectLog: (log: LogRecord) => void;
  onSelectTrace: (traceId: string) => void;
}

export const LogStreamTable: React.FC<LogStreamTableProps> = ({
  logs,
  isPaused,
  onSelectLog,
  onSelectTrace,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const autoScrollRef = useRef(true);

  // Auto-scroll on new logs if user was near bottom
  useEffect(() => {
    if (!autoScrollRef.current || isPaused) return;

    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs, isPaused]);

  const handleScroll = () => {
    if (!containerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = containerRef.current;
    // If within 40px of bottom, keep auto-scroll on
    autoScrollRef.current = scrollHeight - scrollTop - clientHeight < 50;
  };

  if (logs.length === 0) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-500">
        <div className="w-12 h-12 rounded-xl bg-surface border border-border flex items-center justify-center mb-3 text-slate-400">
          <Layers className="w-6 h-6" />
        </div>
        <div className="font-medium text-slate-300 mb-1">Waiting for Log Stream</div>
        <div className="text-xs text-slate-500 max-w-sm">
          Pipe your application or Docker Compose logs into TimberLens to begin analyzing:
        </div>
        <code className="mt-3 px-3 py-1.5 rounded-lg bg-surface border border-border font-mono text-xs text-sky-400">
          docker compose logs -f | timberlens
        </code>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      onScroll={handleScroll}
      className="flex-1 overflow-y-auto font-mono text-xs divide-y divide-border/60 bg-[#090D16]"
    >
      {logs.map((log) => {
        const svcColor = getServiceColor(log.service);
        const levelBadge = getLevelBadge(log.levelStr);

        return (
          <div
            key={log.id}
            onClick={() => onSelectLog(log)}
            className="group px-4 py-2 hover:bg-surface/80 flex items-start space-x-3 cursor-pointer transition-colors border-l-2 border-transparent hover:border-sky-500"
          >
            {/* Timestamp */}
            <div className="text-slate-500 shrink-0 select-none w-20">
              {formatTimestamp(log.timestamp)}
            </div>

            {/* Severity Pill */}
            <div className="shrink-0 w-16">
              <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] uppercase font-mono ${levelBadge}`}>
                {log.levelStr}
              </span>
            </div>

            {/* Service Name */}
            <div className="shrink-0 w-28 truncate">
              <span
                className={`inline-block px-2 py-0.5 rounded text-[11px] truncate max-w-full font-medium border ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}
                title={log.service}
              >
                {log.service}
              </span>
            </div>

            {/* Trace ID */}
            {log.traceId ? (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onSelectTrace(log.traceId!);
                }}
                className="shrink-0 flex items-center space-x-1 px-1.5 py-0.5 rounded bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-400 border border-indigo-500/30 text-[10px] transition-colors"
                title={`Correlate all logs for trace ${log.traceId}`}
              >
                <GitCommit className="w-3 h-3" />
                <span className="font-mono">{log.traceId.slice(0, 8)}</span>
              </button>
            ) : (
              <div className="shrink-0 w-14" />
            )}

            {/* Message & Payload Preview */}
            <div className="flex-1 min-w-0 text-slate-300 break-words group-hover:text-slate-100 flex items-center flex-wrap gap-2">
              <span>{log.message || log.raw}</span>

              {/* Stack trace indicator */}
              {log.stackTrace && (
                <span className="inline-flex items-center space-x-1 px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20 text-[10px]">
                  <AlertCircle className="w-3 h-3 text-rose-400" />
                  <span>
                    Stack trace (-{log.stackTrace.hiddenFrames} vendor frames)
                  </span>
                </span>
              )}

              {/* Latency preview */}
              {log.numeric && log.numeric['latency_ms'] !== undefined && (
                <span className="inline-block px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 text-[10px]">
                  {log.numeric['latency_ms'].toFixed(1)}ms
                </span>
              )}
            </div>

            {/* Inspect Chevron */}
            <div className="opacity-0 group-hover:opacity-100 text-slate-500 shrink-0 self-center">
              <ChevronRight className="w-4 h-4" />
            </div>
          </div>
        );
      })}
    </div>
  );
};
