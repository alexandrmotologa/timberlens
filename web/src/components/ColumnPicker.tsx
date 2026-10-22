import React, { useState, useRef, useEffect } from 'react';
import { Columns3, Check } from 'lucide-react';
import { ColumnVisibility } from '../types/log';

interface ColumnPickerProps {
  columns: ColumnVisibility;
  onChange: (columns: ColumnVisibility) => void;
}

export const ColumnPicker: React.FC<ColumnPickerProps> = ({ columns, onChange }) => {
  const [isOpen, setIsOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const toggle = (key: keyof ColumnVisibility) => {
    onChange({
      ...columns,
      [key]: !columns[key],
    });
  };

  const columnLabels: { key: keyof ColumnVisibility; label: string }[] = [
    { key: 'timestamp', label: 'Timestamp' },
    { key: 'level', label: 'Severity Level' },
    { key: 'service', label: 'Service Name' },
    { key: 'traceId', label: 'Trace Identifier' },
    { key: 'latency', label: 'Latency (ms)' },
    { key: 'message', label: 'Log Message' },
  ];

  return (
    <div className="relative" ref={menuRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg bg-surfaceHover hover:bg-slate-700/60 text-slate-300 hover:text-slate-100 text-xs font-medium border border-border transition-colors"
        title="Customize visible columns"
      >
        <Columns3 className="w-3.5 h-3.5 text-slate-400" />
        <span className="hidden lg:inline">Columns</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 rounded-xl bg-surface border border-border shadow-2xl p-2 z-50 animate-in fade-in duration-100">
          <div className="text-[11px] font-semibold text-slate-400 px-2 py-1 uppercase tracking-wider border-b border-border/50 mb-1">
            Display Columns
          </div>
          <div className="space-y-1">
            {columnLabels.map(({ key, label }) => {
              const active = columns[key];
              return (
                <button
                  key={key}
                  onClick={() => toggle(key)}
                  className={`w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-colors ${
                    active
                      ? 'bg-sky-500/10 text-sky-400 font-medium'
                      : 'text-slate-400 hover:bg-surfaceHover hover:text-slate-200'
                  }`}
                >
                  <span>{label}</span>
                  {active && <Check className="w-3.5 h-3.5 text-sky-400" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
