import React, { useEffect, useState } from 'react';
import { HistogramBucket } from '../types/log';
import { AlertCircle } from 'lucide-react';

interface VelocityHistogramProps {
  onSelectInterval?: (bucket: HistogramBucket) => void;
}

export const VelocityHistogram: React.FC<VelocityHistogramProps> = ({ onSelectInterval }) => {
  const [buckets, setBuckets] = useState<HistogramBucket[]>([]);
  const [hoveredBucket, setHoveredBucket] = useState<{
    bucket: HistogramBucket;
    x: number;
    y: number;
  } | null>(null);

  useEffect(() => {
    let isMounted = true;

    const fetchHistogram = async () => {
      try {
        const res = await fetch('/api/histogram');
        if (res.ok) {
          const data: HistogramBucket[] = await res.json();
          if (isMounted) {
            setBuckets(data || []);
          }
        }
      } catch {
        // histogram fetch error ignored
      }
    };

    fetchHistogram();
    const interval = setInterval(fetchHistogram, 1500);
    return () => {
      isMounted = false;
      clearInterval(interval);
    };
  }, []);

  if (!buckets || buckets.length === 0) {
    return null;
  }

  // Calculate max volume for scale
  const maxTotal = Math.max(10, ...buckets.map((b) => b.total));

  // Check if any bucket has an error spike (more than 3 errors and > 30% of total)
  const hasSpike = buckets.some((b) => b.error > 3 && b.error / Math.max(1, b.total) > 0.3);

  return (
    <div className="relative border-b border-border/70 bg-[#090D16]/95 px-4 py-1.5 flex items-center justify-between select-none">
      <div className="flex items-center space-x-2 text-[11px] text-slate-400 shrink-0 mr-4">
        <span className="font-semibold text-slate-300">Traffic Velocity</span>
        <span className="text-[10px] text-slate-500">(Last 60s)</span>
        {hasSpike && (
          <span className="inline-flex items-center space-x-1 text-rose-400 bg-rose-500/10 px-1.5 py-0.5 rounded text-[10px] border border-rose-500/20 animate-pulse">
            <AlertCircle className="w-3 h-3" />
            <span>Anomaly Detected</span>
          </span>
        )}
      </div>

      {/* Histogram Bars Container */}
      <div className="flex-1 flex items-end h-8 space-x-[2px] max-w-2xl mx-auto overflow-hidden">
        {buckets.map((b, idx) => {
          const heightPct = Math.max(8, Math.min(100, (b.total / maxTotal) * 100));
          const isErrorSpike = b.error > 2 && b.error / Math.max(1, b.total) > 0.3;

          // Height distribution
          const errorPct = b.total > 0 ? (b.error / b.total) * 100 : 0;
          const warnPct = b.total > 0 ? (b.warn / b.total) * 100 : 0;
          const infoPct = b.total > 0 ? (b.info / b.total) * 100 : 100;

          return (
            <div
              key={idx}
              className="flex-1 h-full flex flex-col justify-end items-center relative group cursor-pointer"
              onMouseEnter={(e) => {
                const rect = e.currentTarget.getBoundingClientRect();
                setHoveredBucket({
                  bucket: b,
                  x: rect.left + rect.width / 2,
                  y: rect.top,
                });
              }}
              onMouseLeave={() => setHoveredBucket(null)}
              onClick={() => onSelectInterval?.(b)}
            >
              {/* Anomaly Dot indicator */}
              {isErrorSpike && (
                <div className="w-1 h-1 rounded-full bg-rose-500 mb-0.5 animate-ping" />
              )}

              {/* Stacked bar */}
              <div
                style={{ height: `${heightPct}%` }}
                className={`w-full rounded-t-sm flex flex-col justify-end overflow-hidden transition-all duration-200 group-hover:brightness-125 ${
                  b.total === 0 ? 'bg-slate-800/40' : ''
                }`}
              >
                {b.error > 0 && (
                  <div
                    style={{ height: `${errorPct}%` }}
                    className="w-full bg-rose-500 shrink-0"
                  />
                )}
                {b.warn > 0 && (
                  <div
                    style={{ height: `${warnPct}%` }}
                    className="w-full bg-amber-400 shrink-0"
                  />
                )}
                {b.info > 0 && (
                  <div
                    style={{ height: `${infoPct}%` }}
                    className="w-full bg-sky-500/80 shrink-0"
                  />
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Legend & Stats */}
      <div className="hidden sm:flex items-center space-x-3 text-[10px] text-slate-400 shrink-0 ml-4">
        <div className="flex items-center space-x-1">
          <span className="w-2 h-2 rounded-full bg-sky-400 inline-block" />
          <span>Info</span>
        </div>
        <div className="flex items-center space-x-1">
          <span className="w-2 h-2 rounded-full bg-amber-400 inline-block" />
          <span>Warn</span>
        </div>
        <div className="flex items-center space-x-1">
          <span className="w-2 h-2 rounded-full bg-rose-500 inline-block" />
          <span>Error</span>
        </div>
      </div>

      {/* Floating Tooltip */}
      {hoveredBucket && (
        <div
          style={{
            left: `${hoveredBucket.x}px`,
            top: `${hoveredBucket.y - 8}px`,
            transform: 'translate(-50%, -100%)',
          }}
          className="fixed z-50 pointer-events-none px-2.5 py-1.5 rounded-lg bg-surface border border-border shadow-xl text-slate-200 text-[11px] font-mono whitespace-nowrap animate-in fade-in duration-100"
        >
          <div className="font-semibold text-slate-300 pb-1 border-b border-border/50 flex justify-between gap-3">
            <span>{new Date(hoveredBucket.bucket.timestamp * 1000).toLocaleTimeString()}</span>
            <span className="text-sky-400">{hoveredBucket.bucket.total} logs/sec</span>
          </div>
          <div className="pt-1 flex items-center space-x-3">
            <span className="text-sky-300">Info: {hoveredBucket.bucket.info}</span>
            <span className="text-amber-300">Warn: {hoveredBucket.bucket.warn}</span>
            <span className="text-rose-400 font-bold">Error: {hoveredBucket.bucket.error}</span>
          </div>
        </div>
      )}
    </div>
  );
};
