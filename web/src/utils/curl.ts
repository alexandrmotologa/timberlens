import { LogRecord } from '../types/log';

export interface ExtractedHttpRequest {
  method: string;
  url: string;
  headers?: Record<string, string>;
  body?: string;
}

/**
 * Checks if a log record contains HTTP request attributes.
 */
export function extractHttpRequest(log: LogRecord): ExtractedHttpRequest | null {
  const fields = log.fields || {};

  // Check nested or flat fields
  const method =
    fields.method ||
    fields.http_method ||
    fields.req?.method ||
    fields.request?.method ||
    fields['http.method'];

  const pathOrUrl =
    fields.url ||
    fields.path ||
    fields.uri ||
    fields.req?.url ||
    fields.request?.url ||
    fields['http.url'] ||
    fields.route;

  if (!method || !pathOrUrl) {
    // Check if raw line contains HTTP pattern e.g. "POST /api/v1/checkout HTTP/1.1"
    const httpRegex = /\b(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+([/\w\-._~:?#[\]@!$&'()*+,;=]+)\b/i;
    const match = (log.message || log.raw).match(httpRegex);
    if (match) {
      return {
        method: match[1].toUpperCase(),
        url: match[2],
      };
    }
    return null;
  }

  let finalUrl = String(pathOrUrl);
  const host =
    fields.host ||
    fields.hostname ||
    fields.req?.headers?.host ||
    fields.request?.headers?.host ||
    'localhost';

  if (!finalUrl.startsWith('http://') && !finalUrl.startsWith('https://')) {
    if (!finalUrl.startsWith('/')) {
      finalUrl = '/' + finalUrl;
    }
    finalUrl = `http://${host}${finalUrl}`;
  }

  // Extract headers
  const headers: Record<string, string> = {};
  const rawHeaders = fields.headers || fields.req?.headers || fields.request?.headers;
  if (rawHeaders && typeof rawHeaders === 'object') {
    for (const [k, v] of Object.entries(rawHeaders)) {
      if (typeof v === 'string' || typeof v === 'number') {
        headers[k] = String(v);
      }
    }
  }

  // Extract body
  let bodyStr: string | undefined;
  const rawBody = fields.body || fields.payload || fields.req?.body || fields.request?.body;
  if (rawBody) {
    if (typeof rawBody === 'string') {
      bodyStr = rawBody;
    } else {
      try {
        bodyStr = JSON.stringify(rawBody);
      } catch {
        bodyStr = String(rawBody);
      }
    }
  }

  return {
    method: String(method).toUpperCase(),
    url: finalUrl,
    headers: Object.keys(headers).length > 0 ? headers : undefined,
    body: bodyStr,
  };
}

/**
 * Formats an HTTP request as a ready-to-run curl command.
 */
export function generateCurlCommand(log: LogRecord): string | null {
  const req = extractHttpRequest(log);
  if (!req) return null;

  const parts: string[] = ['curl'];

  if (req.method !== 'GET') {
    parts.push(`-X ${req.method}`);
  }

  parts.push(`"${req.url}"`);

  if (req.headers) {
    for (const [key, value] of Object.entries(req.headers)) {
      if (key.toLowerCase() === 'host') continue; // curl handles host
      parts.push(`-H "${key}: ${value}"`);
    }
  }

  if (req.body) {
    // If JSON body, add header if not already present
    const hasContentType = req.headers && Object.keys(req.headers).some(
      (h) => h.toLowerCase() === 'content-type'
    );
    if (!hasContentType && (req.body.startsWith('{') || req.body.startsWith('['))) {
      parts.push(`-H "Content-Type: application/json"`);
    }

    const escapedBody = req.body.replace(/'/g, "'\\''");
    parts.push(`-d '${escapedBody}'`);
  }

  return parts.join(' ');
}
