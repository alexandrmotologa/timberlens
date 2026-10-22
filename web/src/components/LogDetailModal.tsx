import React, { useState } from 'react';
import { X, Copy, Check, Terminal, Code2, AlertTriangle, Eye, EyeOff } from 'lucide-react';
import { LogRecord } from '../types/log';
import { getServiceColor, getLevelBadge } from '../utils/colors';
import { formatFullDate } from '../utils/format';

interface LogDetailModalProps {
  log: LogRecord | null;
  onClose: () => void;
  onSelectTrace: (traceId: string) => void;
}

export const LogDetailModal: React.FC<LogDetailModalProps> = ({
  log,
  onClose,
  onSelectTrace,
}) => {
  const [copied, setCopied] = useState(false);
  const [copiedRaw, setCopiedRaw] = useState(false);
  const [expandStack, setExpandStack] = useState(false);
  const [activeTab, setActiveTab] = useState<'structured' | 'raw'>('structured');

  if (!log) return null;

  const handleCopyJSON = () => {
    navigator.clipboard.writeText(JSON.stringify(log.fields || log, null, 2));
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleCopyRaw = () => {
    navigator.clipboard.writeText(log.raw);
    setCopiedRaw(true);
    setTimeout(() => setCopiedRaw(false), 2000);
  };

  const svcColor = getServiceColor(log.service);

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-150">
      <div className="w-full max-w-4xl bg-surface border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        {/* Header */}
        <div className="p-4 border-b border-border flex items-center justify-between bg-surface">
          <div className="flex items-center space-x-3">
            <span className={`px-2 py-0.5 rounded text-xs font-mono ${getLevelBadge(log.levelStr)}`}>
              {log.levelStr}
            </span>
            <span className={`px-2.5 py-0.5 rounded font-mono text-xs border font-medium ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}>
              {log.service}
            </span>
            <span className="text-xs font-mono text-slate-400">
              {formatFullDate(log.timestamp)}
            </span>
            <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-400 font-mono">
              {log.formatStr}
            </span>
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={handleCopyJSON}
              className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 text-xs font-medium border border-border transition-colors"
            >
              {copied ? <Check className="w-3.5 h-3.5 text-emerald-400" /> : <Copy className="w-3.5 h-3.5" />}
              <span>{copied ? 'Copied' : 'Copy JSON'}</span>
            </button>

            <button
              onClick={onClose}
              className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-surfaceHover border border-border transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4 font-mono text-xs">
          {/* Message Bar */}
          <div className="p-3 rounded-lg bg-[#090D16] border border-border text-slate-200 break-words font-sans text-sm">
            <div className="text-[11px] font-mono text-slate-500 uppercase tracking-wider mb-1">
              Message
            </div>
            {log.message || log.raw}
          </div>

          {/* Trace Context if available */}
          {log.traceId && (
            <div className="flex items-center justify-between p-3 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-300">
              <div className="flex items-center space-x-3">
                <span className="text-slate-400">Trace ID:</span>
                <span className="font-bold">{log.traceId}</span>
                {log.spanId && (
                  <>
                    <span className="text-slate-500">•</span>
                    <span className="text-slate-400">Span ID:</span>
                    <span>{log.spanId}</span>
                  </>
                )}
              </div>

              <button
                onClick={() => {
                  onSelectTrace(log.traceId!);
                  onClose();
                }}
                className="px-2.5 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white font-medium text-[11px] transition-colors"
              >
                Inspect Waterfall Flow
              </button>
            </div>
          )}

          {/* Stack Trace Section */}
          {log.stackTrace && (
            <div className="rounded-lg bg-[#090D16] border border-rose-500/30 overflow-hidden">
              <div className="px-3 py-2 bg-rose-500/10 border-b border-rose-500/20 flex items-center justify-between">
                <div className="flex items-center space-x-2 text-rose-400 font-semibold">
                  <AlertTriangle className="w-4 h-4" />
                  <span>Stack Trace ({log.stackTrace.language})</span>
                  {log.stackTrace.hiddenFrames > 0 && !expandStack && (
                    <span className="text-slate-400 font-normal text-[11px]">
                      — {log.stackTrace.hiddenFrames} framework frames folded
                    </span>
                  )}
                </div>

                <button
                  onClick={() => setExpandStack(!expandStack)}
                  className="flex items-center space-x-1 px-2 py-1 rounded bg-surface text-slate-300 hover:text-white border border-border text-[11px] transition-colors"
                >
                  {expandStack ? <EyeOff className="w-3.5 h-3.5" /> : <Eye className="w-3.5 h-3.5" />}
                  <span>{expandStack ? 'Fold Vendor Frames' : 'Expand All Frames'}</span>
                </button>
              </div>

              <div className="p-3 text-[11px] text-slate-300 overflow-x-auto space-y-1">
                {expandStack ? (
                  <pre className="text-rose-300 whitespace-pre-wrap">{log.stackTrace.raw}</pre>
                ) : (
                  log.stackTrace.frames.map((frame, i) => {
                    if (frame.isVendor) {
                      // Check if preceding or next frame was vendor to avoid repeating fold labels
                      const prevVendor = i > 0 && log.stackTrace!.frames[i - 1].isVendor;
                      if (prevVendor) return null;
                      return (
                        <div key={i} className="text-slate-500 py-0.5 italic">
                          [+ library frames hidden]
                        </div>
                      );
                    }

                    return (
                      <div key={i} className="text-sky-300 flex items-center space-x-2 py-0.5">
                        <span className="text-amber-400 font-semibold">↳</span>
                        <span className="font-bold">{frame.file}:{frame.line}</span>
                        {frame.function && <span className="text-slate-400">({frame.function})</span>}
                      </div>
                    );
                  })
                )}
              </div>
            </div>
          )}

          {/* Structured Fields & Raw Toggle */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <button
                  onClick={() => setActiveTab('structured')}
                  className={`flex items-center space-x-1.5 px-3 py-1 rounded-md text-xs font-medium transition-colors ${
                    activeTab === 'structured'
                      ? 'bg-sky-500 text-white'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <Code2 className="w-3.5 h-3.5" />
                  <span>Structured Fields</span>
                </button>
                <button
                  onClick={() => setActiveTab('raw')}
                  className={`flex items-center space-x-1.5 px-3 py-1 rounded-md text-xs font-medium transition-colors ${
                    activeTab === 'raw'
                      ? 'bg-sky-500 text-white'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <Terminal className="w-3.5 h-3.5" />
                  <span>Raw Text</span>
                </button>
              </div>

              {activeTab === 'raw' && (
                <button
                  onClick={handleCopyRaw}
                  className="flex items-center space-x-1 text-slate-400 hover:text-slate-200 text-[11px]"
                >
                  {copiedRaw ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                  <span>{copiedRaw ? 'Copied' : 'Copy'}</span>
                </button>
              )}
            </div>

            {activeTab === 'structured' ? (
              <div className="p-3 rounded-lg bg-[#090D16] border border-border overflow-x-auto text-[11px] text-slate-300">
                <pre>{JSON.stringify(log.fields || {}, null, 2)}</pre>
              </div>
            ) : (
              <div className="p-3 rounded-lg bg-[#090D16] border border-border overflow-x-auto text-[11px] text-slate-400 break-all">
                {log.raw}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
