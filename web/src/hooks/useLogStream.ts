import { useState, useEffect, useRef, useCallback } from 'react';
import { LogRecord } from '../types/log';

interface UseLogStreamOptions {
  query: string;
  maxBuffer?: number;
}

export function useLogStream({ query, maxBuffer = 1000 }: UseLogStreamOptions) {
  const [logs, setLogs] = useState<LogRecord[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [isConnected, setIsConnected] = useState(false);
  const [linesPerSec, setLinesPerSec] = useState(0);

  const eventSourceRef = useRef<EventSource | null>(null);
  const isPausedRef = useRef(isPaused);
  isPausedRef.current = isPaused;

  const countSinceLastInterval = useRef(0);

  // Fetch initial logs snapshot on query change
  useEffect(() => {
    let ignore = false;
    const fetchSnapshot = async () => {
      try {
        const url = `/api/logs?q=${encodeURIComponent(query)}&limit=150`;
        const res = await fetch(url);
        if (res.ok) {
          const data = await res.json();
          if (!ignore && data.records) {
            setLogs(data.records);
          }
        }
      } catch (err) {
        console.error('Failed to load initial logs:', err);
      }
    };

    fetchSnapshot();
    return () => {
      ignore = true;
    };
  }, [query]);

  // Connect to SSE stream
  useEffect(() => {
    const streamUrl = `/api/stream?q=${encodeURIComponent(query)}`;
    const es = new EventSource(streamUrl);
    eventSourceRef.current = es;

    es.onopen = () => {
      setIsConnected(true);
    };

    es.onmessage = (event) => {
      countSinceLastInterval.current++;
      if (isPausedRef.current) {
        return;
      }

      try {
        const record: LogRecord = JSON.parse(event.data);
        setLogs((prev) => {
          const next = [...prev, record];
          if (next.length > maxBuffer) {
            return next.slice(next.length - maxBuffer);
          }
          return next;
        });
      } catch {
        // ignore parse errors
      }
    };

    es.onerror = () => {
      setIsConnected(false);
    };

    return () => {
      es.close();
      eventSourceRef.current = null;
    };
  }, [query, maxBuffer]);

  // Velocity calculation interval (lines/sec)
  useEffect(() => {
    const interval = setInterval(() => {
      setLinesPerSec(countSinceLastInterval.current);
      countSinceLastInterval.current = 0;
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const togglePause = useCallback(() => {
    setIsPaused((p) => !p);
  }, []);

  const clearLogs = useCallback(() => {
    setLogs([]);
  }, []);

  return {
    logs,
    isConnected,
    isPaused,
    linesPerSec,
    togglePause,
    clearLogs,
  };
}
