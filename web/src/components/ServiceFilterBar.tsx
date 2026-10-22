import React from 'react';
import { Layers } from 'lucide-react';
import { getServiceColor } from '../utils/colors';

interface ServiceFilterBarProps {
  services: Record<string, number>;
  selectedService: string;
  onSelectService: (service: string) => void;
}

export const ServiceFilterBar: React.FC<ServiceFilterBarProps> = ({
  services,
  selectedService,
  onSelectService,
}) => {
  const serviceList = Object.entries(services).sort((a, b) => b[1] - a[1]);

  return (
    <div className="bg-surface/50 border-b border-border px-4 py-2 flex items-center space-x-2 overflow-x-auto scrollbar-none text-xs select-none">
      <div className="flex items-center space-x-1.5 text-slate-400 mr-2 shrink-0">
        <Layers className="w-3.5 h-3.5" />
        <span className="font-medium text-[11px] uppercase tracking-wider">Services:</span>
      </div>

      <button
        onClick={() => onSelectService('')}
        className={`px-2.5 py-1 rounded-md transition-all shrink-0 font-medium ${
          selectedService === ''
            ? 'bg-sky-500 text-white font-semibold shadow-sm'
            : 'bg-surface hover:bg-surfaceHover text-slate-400 hover:text-slate-200 border border-border'
        }`}
      >
        All Services
      </button>

      {serviceList.map(([name, count]) => {
        const colors = getServiceColor(name);
        const isSelected = selectedService.toLowerCase() === name.toLowerCase();

        return (
          <button
            key={name}
            onClick={() => onSelectService(isSelected ? '' : name)}
            className={`flex items-center space-x-1.5 px-2.5 py-1 rounded-md transition-all shrink-0 border ${
              isSelected
                ? 'bg-slate-800 text-white border-sky-400 ring-1 ring-sky-400/50'
                : `${colors.bg} ${colors.text} ${colors.border} hover:opacity-90`
            }`}
          >
            <span
              className="w-1.5 h-1.5 rounded-full"
              style={{ backgroundColor: colors.hex }}
            />
            <span className="font-mono">{name}</span>
            <span className="text-[10px] opacity-75 font-mono">({count})</span>
          </button>
        );
      })}
    </div>
  );
};
