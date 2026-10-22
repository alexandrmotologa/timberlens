import React from 'react';
import {
  Play,
  Pause,
  Trash2,
  Activity,
  Zap,
  Cpu,
  AlertTriangle,
  Download,
  GitCompare,
} from 'lucide-react';
import { SystemStats, ColumnVisibility } from '../types/log';
import { formatNumber } from '../utils/format';
import { ColumnPicker } from './ColumnPicker';
import { AudioSettings } from './AudioSettings';

interface HeaderProps {
  stats: SystemStats | null;
  linesPerSec: number;
  isConnected: boolean;
  isPaused: boolean;
  columns: ColumnVisibility;
  onColumnsChange: (columns: ColumnVisibility) => void;
  onTogglePause: () => void;
  onClear: () => void;
  onOpenClusters: () => void;
  onOpenTraceCompare: () => void;
  onExportSnapshot: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  stats,
  linesPerSec,
  isConnected,
  isPaused,
  columns,
  onColumnsChange,
  onTogglePause,
  onClear,
  onOpenClusters,
  onOpenTraceCompare,
  onExportSnapshot,
}) => {
  return (
    <header className="h-14 border-b border-border bg-surface px-4 flex items-center justify-between select-none">
      {/* Brand & Status */}
      <div className="flex items-center space-x-3">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 rounded-lg bg-sky-500/10 border border-sky-500/30 flex items-center justify-center text-sky-400">
            <Activity className="w-4 h-4 text-sky-400" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <span className="font-bold text-slate-100 text-sm tracking-tight">TimberLens</span>
              <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-sky-500/10 text-sky-400 border border-sky-500/20">
                v1.1
              </span>
            </div>
            <div className="text-[11px] text-slate-400 flex items-center space-x-1.5">
              <span
                className={`w-1.5 h-1.5 rounded-full ${
                  isConnected
                    ? isPaused
                      ? 'bg-amber-400'
                      : 'bg-emerald-400 animate-pulse'
                    : 'bg-rose-500'
                }`}
              />
              <span>
                {!isConnected ? 'Disconnected' : isPaused ? 'Stream Paused' : 'Live Streaming'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Metrics Center */}
      <div className="hidden lg:flex items-center space-x-6 text-xs text-slate-300">
        <div className="flex items-center space-x-2">
          <Zap className="w-3.5 h-3.5 text-amber-400" />
          <span className="text-slate-400">Velocity:</span>
          <span className="font-mono font-medium text-slate-100">{linesPerSec} lines/s</span>
        </div>

        <div className="flex items-center space-x-2">
          <Cpu className="w-3.5 h-3.5 text-sky-400" />
          <span className="text-slate-400">RAM:</span>
          <span className="font-mono font-medium text-slate-100">
            {stats ? `${stats.memoryMb.toFixed(1)} MB` : '--'}
          </span>
        </div>

        <div className="flex items-center space-x-2">
          <span className="text-slate-400">Ingested:</span>
          <span className="font-mono font-medium text-slate-100">
            {stats ? formatNumber(stats.totalIngested) : '0'}
          </span>
        </div>
      </div>

      {/* Control Actions */}
      <div className="flex items-center space-x-1 sm:space-x-2">
        {/* Trace Comparison */}
        <button
          onClick={onOpenTraceCompare}
          className="flex items-center space-x-1 px-2 sm:px-2.5 py-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 text-xs font-medium border border-border transition-colors"
          title="Compare two traces side-by-side"
        >
          <GitCompare className="w-3.5 h-3.5 text-indigo-400" />
          <span className="hidden md:inline">Trace Diff</span>
        </button>

        {/* Error Signatures */}
        <button
          onClick={onOpenClusters}
          className="flex items-center space-x-1 px-2 sm:px-2.5 py-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 text-xs font-medium border border-border transition-colors"
          title="Inspect recurring error signatures"
        >
          <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
          <span className="hidden md:inline">Signatures</span>
        </button>

        {/* Offline HTML Snapshot Export */}
        <button
          onClick={onExportSnapshot}
          className="flex items-center space-x-1 px-2 sm:px-2.5 py-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 text-xs font-medium border border-border transition-colors"
          title="Export self-contained standalone HTML snapshot"
        >
          <Download className="w-3.5 h-3.5 text-sky-400" />
          <span className="hidden md:inline">Snapshot</span>
        </button>

        {/* Column Customizer */}
        <ColumnPicker columns={columns} onChange={onColumnsChange} />

        {/* Audio & Desktop Alerts Settings */}
        <AudioSettings />

        {/* Stream Pause / Resume */}
        <button
          onClick={onTogglePause}
          className={`flex items-center space-x-1 px-2 sm:px-2.5 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
            isPaused
              ? 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border-emerald-500/30'
              : 'bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 border-border'
          }`}
          title={isPaused ? 'Resume live log stream' : 'Pause live log stream'}
        >
          {isPaused ? <Play className="w-3.5 h-3.5" /> : <Pause className="w-3.5 h-3.5" />}
          <span className="hidden sm:inline">{isPaused ? 'Resume' : 'Pause'}</span>
        </button>

        {/* Clear Buffer */}
        <button
          onClick={onClear}
          className="p-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-400 hover:text-slate-200 border border-border transition-colors"
          title="Clear screen buffer"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      </div>
    </header>
  );
};
