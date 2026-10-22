import React, { useEffect, useState } from 'react';
import { X, AlertTriangle, RefreshCw, Filter, Clock } from 'lucide-react';
import { ErrorCluster } from '../types/log';
import { getServiceColor } from '../utils/colors';
import { formatFullDate } from '../utils/format';

interface ErrorClustersProps {
  isOpen: boolean;
  onClose: () => void;
  onApplyFilter: (query: string) => void;
}

export const ErrorClusters: React.FC<ErrorClustersProps> = ({
  isOpen,
  onClose,
  onApplyFilter,
}) => {
  const [clusters, setClusters] = useState<ErrorCluster[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchClusters = () => {
    setLoading(true);
    fetch('/api/clusters')
      .then((res) => res.json())
      .then((data) => {
        if (Array.isArray(data)) {
          setClusters(data);
        }
      })
      .catch((err) => console.error('Failed to load clusters:', err))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (isOpen) {
      fetchClusters();
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-150">
      <div className="w-full max-w-3xl bg-surface border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
        {/* Header */}
        <div className="p-4 border-b border-border flex items-center justify-between bg-surface">
          <div className="flex items-center space-x-2">
            <div className="p-1.5 rounded-lg bg-amber-500/10 text-amber-400 border border-amber-500/20">
              <AlertTriangle className="w-4 h-4" />
            </div>
            <div>
              <div className="text-sm font-semibold text-slate-100">Recurring Error Signatures</div>
              <div className="text-xs text-slate-400">
                Clusters similar errors by normalizing variable values, memory addresses, and UUIDs
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={fetchClusters}
              className="p-1.5 rounded-lg text-slate-400 hover:text-slate-200 hover:bg-surfaceHover border border-border transition-colors"
              title="Refresh error clusters"
            >
              <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
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
        <div className="flex-1 overflow-y-auto p-4 space-y-3 font-mono text-xs">
          {clusters.length === 0 ? (
            <div className="p-8 text-center text-slate-500">
              {loading ? 'Analyzing error clusters...' : 'No recurring errors detected yet.'}
            </div>
          ) : (
            clusters.map((c) => {
              const svcColor = getServiceColor(c.service);

              return (
                <div
                  key={c.signature}
                  className="p-3.5 rounded-lg bg-[#090D16] border border-border hover:border-amber-500/40 transition-all space-y-2.5"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                      <span className="px-2 py-0.5 rounded-full bg-rose-500/15 text-rose-400 font-bold border border-rose-500/30 text-[11px]">
                        {c.count} occurrences
                      </span>
                      <span className={`px-2 py-0.5 rounded font-medium border ${svcColor.bg} ${svcColor.text} ${svcColor.border}`}>
                        {c.service}
                      </span>
                    </div>

                    <button
                      onClick={() => {
                        onApplyFilter(`service:${c.service} level>=warn`);
                        onClose();
                      }}
                      className="flex items-center space-x-1 px-2.5 py-1 rounded bg-sky-500/10 hover:bg-sky-500/20 text-sky-400 border border-sky-500/30 text-[11px] transition-colors"
                    >
                      <Filter className="w-3 h-3" />
                      <span>Filter Stream</span>
                    </button>
                  </div>

                  <div className="text-slate-200 font-medium break-words">
                    {c.pattern}
                  </div>

                  {c.rootCause && (
                    <div className="text-amber-400/90 text-[11px] flex items-center space-x-1.5">
                      <span className="text-slate-500">Origin:</span>
                      <span className="font-semibold underline decoration-amber-500/30">{c.rootCause}</span>
                    </div>
                  )}

                  <div className="flex items-center space-x-4 text-[11px] text-slate-500 pt-1 border-t border-border/40">
                    <span className="flex items-center space-x-1">
                      <Clock className="w-3 h-3" />
                      <span>First seen: {formatFullDate(c.firstSeen)}</span>
                    </span>
                    <span>•</span>
                    <span>Last seen: {formatFullDate(c.lastSeen)}</span>
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
