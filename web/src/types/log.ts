export type SeverityLevel = 'TRACE' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL';

export interface StackFrame {
  file: string;
  line: number;
  function?: string;
  isVendor: boolean;
  raw: string;
}

export interface StackTraceInfo {
  raw: string;
  language: string;
  rootCause: string;
  totalFrames: number;
  hiddenFrames: number;
  frames: StackFrame[];
}

export interface LogRecord {
  id: number;
  timestamp: string;
  level: number;
  levelStr: SeverityLevel;
  service: string;
  traceId?: string;
  spanId?: string;
  message: string;
  raw: string;
  format: number;
  formatStr: string;
  fields?: Record<string, any>;
  numeric?: Record<string, number>;
  stackTrace?: StackTraceInfo;
}

export interface ErrorCluster {
  signature: string;
  pattern: string;
  rootCause?: string;
  count: number;
  firstSeen: string;
  lastSeen: string;
  service: string;
  sample?: LogRecord;
}

export interface SystemStats {
  totalIngested: number;
  capacity: number;
  bufferedCount: number;
  memoryBytes: number;
  memoryMb: number;
  services: Record<string, number>;
  levels: Record<string, number>;
}
