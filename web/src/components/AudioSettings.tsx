import React, { useState, useRef, useEffect } from 'react';
import { Volume2, VolumeX, PlayCircle } from 'lucide-react';
import { alertManager } from '../utils/alerts';

export const AudioSettings: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [soundOn, setSoundOn] = useState(alertManager.isSoundEnabled());
  const [desktopOn, setDesktopOn] = useState(alertManager.isDesktopEnabled());
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

  const toggleSound = () => {
    const next = !soundOn;
    setSoundOn(next);
    alertManager.setSoundEnabled(next);
    if (next) {
      alertManager.playErrorChime();
    }
  };

  const toggleDesktop = async () => {
    const next = !desktopOn;
    const ok = await alertManager.setDesktopEnabled(next);
    setDesktopOn(ok);
  };

  return (
    <div className="relative" ref={menuRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg border text-xs font-medium transition-colors ${
          soundOn || desktopOn
            ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
            : 'bg-surfaceHover text-slate-400 hover:text-slate-200 border-border'
        }`}
        title="Audio & Desktop Alerts Configuration"
      >
        {soundOn ? <Volume2 className="w-3.5 h-3.5 text-amber-400" /> : <VolumeX className="w-3.5 h-3.5" />}
        <span className="hidden xl:inline">Alerts</span>
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-64 rounded-xl bg-surface border border-border shadow-2xl p-3 z-50 animate-in fade-in duration-100 font-sans">
          <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider border-b border-border/50 pb-2 mb-2">
            Realtime Error Alerts
          </div>

          <div className="space-y-3 text-xs">
            {/* Audio Toggle */}
            <div className="flex items-center justify-between">
              <div>
                <div className="font-medium text-slate-200">Audio Chimes</div>
                <div className="text-[10px] text-slate-400">
                  Synthesized chord on error logs
                </div>
              </div>
              <button
                onClick={toggleSound}
                className={`w-9 h-5 rounded-full transition-colors relative p-0.5 ${
                  soundOn ? 'bg-amber-500' : 'bg-slate-700'
                }`}
              >
                <div
                  className={`w-4 h-4 rounded-full bg-white transition-transform ${
                    soundOn ? 'translate-x-4' : 'translate-x-0'
                  }`}
                />
              </button>
            </div>

            {/* Test sound button */}
            {soundOn && (
              <button
                onClick={() => alertManager.playErrorChime()}
                className="w-full flex items-center justify-center space-x-1 px-2 py-1 rounded bg-surfaceHover hover:bg-slate-700/60 text-slate-300 text-[11px] border border-border transition-colors"
              >
                <PlayCircle className="w-3 h-3 text-amber-400" />
                <span>Play Test Chime</span>
              </button>
            )}

            {/* Desktop Notification Toggle */}
            <div className="flex items-center justify-between pt-2 border-t border-border/50">
              <div>
                <div className="font-medium text-slate-200">Desktop Notifications</div>
                <div className="text-[10px] text-slate-400">
                  Alert when tab is in background
                </div>
              </div>
              <button
                onClick={toggleDesktop}
                className={`w-9 h-5 rounded-full transition-colors relative p-0.5 ${
                  desktopOn ? 'bg-indigo-500' : 'bg-slate-700'
                }`}
              >
                <div
                  className={`w-4 h-4 rounded-full bg-white transition-transform ${
                    desktopOn ? 'translate-x-4' : 'translate-x-0'
                  }`}
                />
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
