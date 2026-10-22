import { LogRecord, SystemStats } from '../types/log';

export function exportHtmlSnapshot(
  logs: LogRecord[],
  stats: SystemStats | null,
  activeQuery: string = ''
) {
  const exportTime = new Date().toISOString();
  const safeLogs = logs.slice(0, 3000); // Snapshot limit for responsive standalone DOM
  const logsJson = JSON.stringify(safeLogs);

  const htmlContent = `<!DOCTYPE html>
<html lang="en" class="dark">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>TimberLens Snapshot — ${new Date().toLocaleString()}</title>
  <style>
    :root {
      --bg: #090D16;
      --surface: #0F172A;
      --surface-hover: #1E293B;
      --border: #1E293B;
      --text: #F8FAFC;
      --text-muted: #94A3B8;
      --sky: #38BDF8;
      --rose: #F43F5E;
      --amber: #FBBF24;
      --emerald: #34D399;
      --indigo: #818CF8;
    }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      background-color: var(--bg);
      color: var(--text);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
      font-size: 13px;
      line-height: 1.5;
    }
    header {
      background-color: var(--surface);
      border-bottom: 1px solid var(--border);
      padding: 12px 20px;
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      position: sticky;
      top: 0;
      z-index: 100;
    }
    .brand {
      display: flex;
      align-items: center;
      gap: 8px;
      font-weight: 700;
      font-size: 15px;
      color: var(--text);
    }
    .badge {
      font-size: 10px;
      text-transform: uppercase;
      font-family: monospace;
      padding: 2px 6px;
      border-radius: 4px;
      background: rgba(56, 189, 248, 0.1);
      color: var(--sky);
      border: 1px solid rgba(56, 189, 248, 0.2);
    }
    .meta {
      display: flex;
      gap: 16px;
      color: var(--text-muted);
      font-size: 12px;
      font-family: monospace;
    }
    .search-bar {
      padding: 10px 20px;
      background: var(--surface);
      border-bottom: 1px solid var(--border);
      display: flex;
      gap: 10px;
    }
    .search-input {
      flex: 1;
      background: #060911;
      border: 1px solid var(--border);
      color: var(--text);
      padding: 8px 12px;
      border-radius: 6px;
      font-family: monospace;
      font-size: 12px;
      outline: none;
    }
    .search-input:focus {
      border-color: var(--sky);
    }
    .log-list {
      display: flex;
      flex-direction: column;
    }
    .log-item {
      display: flex;
      flex-direction: column;
      border-bottom: 1px solid rgba(30, 41, 59, 0.6);
      font-family: monospace;
      font-size: 11.5px;
      transition: background-color 0.15s;
    }
    .log-item:hover {
      background-color: rgba(15, 23, 42, 0.8);
    }
    .log-summary {
      display: flex;
      align-items: baseline;
      padding: 8px 20px;
      gap: 12px;
      cursor: pointer;
      user-select: none;
    }
    .timestamp { color: var(--text-muted); width: 85px; flex-shrink: 0; }
    .level {
      width: 50px;
      flex-shrink: 0;
      text-transform: uppercase;
      font-weight: 700;
      font-size: 10px;
      text-align: center;
      padding: 1px 4px;
      border-radius: 3px;
    }
    .level-ERROR, .level-FATAL { background: rgba(244, 63, 94, 0.15); color: #fb7185; }
    .level-WARN { background: rgba(251, 191, 36, 0.15); color: #fde047; }
    .level-INFO { background: rgba(56, 189, 248, 0.15); color: #7dd3fc; }
    .level-DEBUG, .level-TRACE { background: rgba(148, 163, 184, 0.15); color: #cbd5e1; }
    .service {
      font-weight: 600;
      color: var(--sky);
      width: 120px;
      flex-shrink: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    .trace-pill {
      font-size: 10px;
      background: rgba(129, 140, 248, 0.15);
      color: var(--indigo);
      border: 1px solid rgba(129, 140, 248, 0.3);
      padding: 1px 6px;
      border-radius: 4px;
      flex-shrink: 0;
    }
    .message {
      flex: 1;
      color: #E2E8F0;
      word-break: break-word;
    }
    .log-detail {
      display: none;
      padding: 12px 20px 16px 20px;
      background: #060911;
      border-top: 1px dashed var(--border);
    }
    .log-detail.open { display: block; }
    pre {
      background: #030712;
      border: 1px solid var(--border);
      border-radius: 6px;
      padding: 10px;
      overflow-x: auto;
      color: #93C5FD;
      font-size: 11px;
    }
    .stacktrace-box {
      margin-top: 10px;
      background: rgba(244, 63, 94, 0.05);
      border: 1px solid rgba(244, 63, 94, 0.2);
      border-radius: 6px;
      padding: 10px;
      color: #FDA4AF;
      font-size: 11px;
    }
    .action-btn {
      background: var(--surface);
      color: var(--text);
      border: 1px solid var(--border);
      padding: 4px 10px;
      border-radius: 4px;
      cursor: pointer;
      font-size: 11px;
      margin-bottom: 8px;
    }
    .action-btn:hover { background: var(--surface-hover); }
  </style>
</head>
<body>
  <header>
    <div class="brand">
      <span>TimberLens</span>
      <span class="badge">Offline Snapshot</span>
    </div>
    <div class="meta">
      <span>Logs: <strong>${safeLogs.length}</strong></span>
      <span>Exported: <strong>${new Date().toLocaleTimeString()}</strong></span>
      ${stats ? `<span>Memory: <strong>${stats.memoryMb.toFixed(1)} MB</strong></span>` : ''}
      ${activeQuery ? `<span>Filter: <strong>${activeQuery}</strong></span>` : ''}
    </div>
  </header>

  <div class="search-bar">
    <input type="text" id="filterInput" class="search-input" placeholder="Quick search in snapshot (service, error message, trace ID, level)..." />
  </div>

  <div id="logList" class="log-list"></div>

  <script>
    const logs = ${logsJson};

    function renderLogs(filterText) {
      const container = document.getElementById('logList');
      container.innerHTML = '';
      const filter = (filterText || '').toLowerCase().trim();

      const filtered = logs.filter(l => {
        if (!filter) return true;
        const msg = (l.message || l.raw || '').toLowerCase();
        const svc = (l.service || '').toLowerCase();
        const lvl = (l.levelStr || '').toLowerCase();
        const trace = (l.traceId || '').toLowerCase();
        return msg.includes(filter) || svc.includes(filter) || lvl.includes(filter) || trace.includes(filter);
      });

      filtered.forEach((log, idx) => {
        const item = document.createElement('div');
        item.className = 'log-item';

        const summary = document.createElement('div');
        summary.className = 'log-summary';

        const time = new Date(log.timestamp).toLocaleTimeString();
        summary.innerHTML = \`
          <span class="timestamp">\${time}</span>
          <span class="level level-\${log.levelStr}">\${log.levelStr}</span>
          <span class="service" title="\${log.service}">\${log.service}</span>
          \${log.traceId ? \`<span class="trace-pill">\${log.traceId.slice(0, 8)}</span>\` : ''}
          <span class="message">\${escapeHtml(log.message || log.raw)}</span>
        \`;

        const detail = document.createElement('div');
        detail.className = 'log-detail';
        detail.id = 'detail-' + idx;

        let detailHtml = '<button class="action-btn" onclick="copyJson(' + idx + ')">Copy JSON</button>';
        if (log.fields) {
          detailHtml += '<pre>' + escapeHtml(JSON.stringify(log.fields, null, 2)) + '</pre>';
        } else {
          detailHtml += '<pre>' + escapeHtml(log.raw) + '</pre>';
        }

        if (log.stackTrace && log.stackTrace.raw) {
          detailHtml += '<div class="stacktrace-box"><strong>Stack Trace:</strong><br><pre>' + escapeHtml(log.stackTrace.raw) + '</pre></div>';
        }

        detail.innerHTML = detailHtml;

        summary.addEventListener('click', () => {
          detail.classList.toggle('open');
        });

        item.appendChild(summary);
        item.appendChild(detail);
        container.appendChild(item);
      });
    }

    function escapeHtml(str) {
      if (!str) return '';
      return str.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");
    }

    function copyJson(idx) {
      const log = logs[idx];
      navigator.clipboard.writeText(JSON.stringify(log.fields || log, null, 2));
      alert('Copied log JSON to clipboard!');
    }

    document.getElementById('filterInput').addEventListener('input', (e) => {
      renderLogs(e.target.value);
    });

    renderLogs('');
  </script>
</body>
</html>`;

  const blob = new Blob([htmlContent], { type: 'text/html;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  const timestampStr = exportTime.replace(/[:.]/g, '-').slice(0, 19);
  a.href = url;
  a.download = `timberlens-snapshot-${timestampStr}.html`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
