import React, { useState, useEffect } from 'react';
import { Search, X, HelpCircle } from 'lucide-react';

interface QueryInputProps {
  value: string;
  onChange: (query: string) => void;
}

const PRESETS = [
  { label: 'All Logs', query: '' },
  { label: 'Errors & Warnings', query: 'level>=warn' },
  { label: 'Slow Latency (>100ms)', query: 'latency_ms>100' },
  { label: 'HTTP 5xx Server Errors', query: 'status>=500' },
  { label: 'Exceptions & Stacks', query: 'stack' },
];

export const QueryInput: React.FC<QueryInputProps> = ({ value, onChange }) => {
  const [localVal, setLocalVal] = useState(value);
  const [showHelp, setShowHelp] = useState(false);

  useEffect(() => {
    setLocalVal(value);
  }, [value]);

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      onChange(localVal);
    }
  };

  const handleClear = () => {
    setLocalVal('');
    onChange('');
  };

  return (
    <div className="bg-surface px-4 py-3 border-b border-border space-y-2 select-none">
      {/* Search Input Bar */}
      <div className="relative flex items-center">
        <div className="absolute left-3 text-slate-400 pointer-events-none">
          <Search className="w-4 h-4" />
        </div>

        <input
          type="text"
          value={localVal}
          onChange={(e) => setLocalVal(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Filter logs by substring or expressions (e.g. service:api level:error latency_ms>250 trace_id:abc)..."
          className="w-full bg-[#090D16] border border-border focus:border-sky-500 rounded-lg pl-9 pr-20 py-2 text-xs font-mono text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-sky-500 transition-all"
        />

        <div className="absolute right-3 flex items-center space-x-2">
          {localVal && (
            <button
              onClick={handleClear}
              className="text-slate-400 hover:text-slate-200 p-0.5"
              title="Clear filter"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}

          <button
            onClick={() => setShowHelp(!showHelp)}
            className="text-slate-400 hover:text-sky-400 transition-colors"
            title="Query Syntax Guide"
          >
            <HelpCircle className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Preset Pills */}
      <div className="flex items-center space-x-2 text-[11px] overflow-x-auto scrollbar-none">
        <span className="text-slate-500 font-medium shrink-0">Presets:</span>
        {PRESETS.map((p) => {
          const isActive = value === p.query;
          return (
            <button
              key={p.label}
              onClick={() => {
                setLocalVal(p.query);
                onChange(p.query);
              }}
              className={`px-2 py-0.5 rounded border transition-colors shrink-0 ${
                isActive
                  ? 'bg-sky-500/20 text-sky-400 border-sky-500/40 font-semibold'
                  : 'bg-surfaceHover hover:bg-slate-700/60 text-slate-400 hover:text-slate-200 border-border'
              }`}
            >
              {p.label}
            </button>
          );
        })}
      </div>

      {/* Query Syntax Help Dropdown */}
      {showHelp && (
        <div className="mt-2 p-3 rounded-lg bg-[#090D16] border border-border text-xs text-slate-300 grid grid-cols-1 md:grid-cols-2 gap-2 animate-in fade-in duration-150">
          <div>
            <div className="font-semibold text-sky-400 mb-1">Field Comparisons</div>
            <div className="font-mono text-[11px] text-slate-400 space-y-0.5">
              <div><span className="text-sky-300">service:auth</span> — exact service match</div>
              <div><span className="text-sky-300">level:error</span> — matches ERROR severity</div>
              <div><span className="text-sky-300">level&gt;=warn</span> — WARN, ERROR, and FATAL</div>
              <div><span className="text-sky-300">trace_id:trc-123</span> — exact trace filter</div>
            </div>
          </div>
          <div>
            <div className="font-semibold text-sky-400 mb-1">Numeric &amp; Substring</div>
            <div className="font-mono text-[11px] text-slate-400 space-y-0.5">
              <div><span className="text-sky-300">latency_ms&gt;500</span> — operations &gt; 500ms</div>
              <div><span className="text-sky-300">"connection timeout"</span> — phrase search</div>
              <div><span className="text-sky-300">!level:debug</span> — negation filter</div>
              <div><span className="text-sky-300">service:api level:err 503</span> — combined AND</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
