import { SeverityLevel } from '../types/log';

const serviceColors = [
  { bg: 'bg-sky-500/10', text: 'text-sky-400', border: 'border-sky-500/20', hex: '#38BDF8' },
  { bg: 'bg-purple-500/10', text: 'text-purple-400', border: 'border-purple-500/20', hex: '#A855F7' },
  { bg: 'bg-pink-500/10', text: 'text-pink-400', border: 'border-pink-500/20', hex: '#EC4899' },
  { bg: 'bg-emerald-500/10', text: 'text-emerald-400', border: 'border-emerald-500/20', hex: '#10B981' },
  { bg: 'bg-orange-500/10', text: 'text-orange-400', border: 'border-orange-500/20', hex: '#F97316' },
  { bg: 'bg-amber-500/10', text: 'text-amber-400', border: 'border-amber-500/20', hex: '#EAB308' },
  { bg: 'bg-indigo-500/10', text: 'text-indigo-400', border: 'border-indigo-500/20', hex: '#6366F1' },
  { bg: 'bg-teal-500/10', text: 'text-teal-400', border: 'border-teal-500/20', hex: '#14B8A6' },
  { bg: 'bg-rose-500/10', text: 'text-rose-400', border: 'border-rose-500/20', hex: '#F43F5E' },
  { bg: 'bg-lime-500/10', text: 'text-lime-400', border: 'border-lime-500/20', hex: '#84CC16' },
];

export function getServiceColor(service: string) {
  if (!service) {
    return { bg: 'bg-slate-500/10', text: 'text-slate-400', border: 'border-slate-500/20', hex: '#94A3B8' };
  }
  let hash = 0;
  for (let i = 0; i < service.length; i++) {
    hash = (hash << 5) - hash + service.charCodeAt(i);
    hash |= 0;
  }
  const idx = Math.abs(hash) % serviceColors.length;
  return serviceColors[idx];
}

export function getLevelBadge(level: SeverityLevel | string) {
  switch (level?.toUpperCase()) {
    case 'FATAL':
      return 'bg-pink-500/20 text-pink-400 border border-pink-500/30 font-semibold';
    case 'ERROR':
      return 'bg-rose-500/20 text-rose-400 border border-rose-500/30 font-semibold';
    case 'WARN':
      return 'bg-amber-500/20 text-amber-400 border border-amber-500/30 font-medium';
    case 'INFO':
      return 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20';
    case 'DEBUG':
      return 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/20';
    case 'TRACE':
      return 'bg-slate-500/15 text-slate-400 border border-slate-500/20';
    default:
      return 'bg-slate-500/15 text-slate-400 border border-slate-500/20';
  }
}
