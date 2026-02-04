'use client';

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { env } from '@/config/env';
import { cn, generateId } from '@/lib/utils';

export interface LogViewerProps {
  deploymentId: string;
  containerName?: string;
  tail?: number;
  follow?: boolean;
}

type LogLevel = 'error' | 'warn' | 'info' | 'debug';

interface LogLine {
  id: string;
  raw: string;
  message: string;
  level: LogLevel;
  timestamp?: string;
}

const MAX_LOG_LINES = 2000;
const LOG_LEVELS: LogLevel[] = ['error', 'warn', 'info', 'debug'];
const LEVEL_LABELS: Record<LogLevel, string> = {
  error: 'Error',
  warn: 'Warn',
  info: 'Info',
  debug: 'Debug',
};

const LEVEL_STYLES: Record<LogLevel, string> = {
  error: 'text-destructive',
  warn: 'text-warning',
  info: 'text-primary',
  debug: 'text-muted-foreground',
};

const WS_BACKOFF_MS = [800, 1500, 2500, 4000, 8000, 12000];

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  return (
    window.localStorage.getItem('ve_session_token') ??
    window.localStorage.getItem('ve_portal_token') ??
    null
  );
}

function buildAuthHeaders(): HeadersInit {
  const token = getAuthToken();
  if (!token) return {};
  return {
    Authorization: `Bearer ${token}`,
  };
}

function toWebSocketUrl(url: string): string {
  if (url.startsWith('https://')) {
    return url.replace('https://', 'wss://');
  }
  if (url.startsWith('http://')) {
    return url.replace('http://', 'ws://');
  }
  return url;
}

function parseLogLine(raw: string): LogLine {
  const trimmed = raw.replace(/\r?\n$/, '');
  const levelMatch = trimmed.match(/\b(ERROR|WARN|WARNING|INFO|DEBUG)\b/i);
  const levelRaw = levelMatch?.[1]?.toLowerCase() ?? 'info';
  const level: LogLevel =
    levelRaw === 'warning' ? 'warn' : (levelRaw as LogLevel) ?? 'info';
  const timestampMatch = trimmed.match(
    /^(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2}(?:\.\d+)?Z?)\s+/
  );
  const timestamp = timestampMatch?.[1];
  const message = timestamp ? trimmed.slice(timestamp.length).trim() : trimmed;

  return {
    id: generateId('log'),
    raw: trimmed,
    message,
    level,
    timestamp,
  };
}

export function LogViewer({
  deploymentId,
  containerName,
  tail = 200,
  follow = true,
}: LogViewerProps) {
  const [logLines, setLogLines] = useState<LogLine[]>([]);
  const [selectedLevels, setSelectedLevels] = useState<Set<LogLevel>>(
    new Set(LOG_LEVELS)
  );
  const [searchTerm, setSearchTerm] = useState('');
  const [autoScroll, setAutoScroll] = useState(true);
  const [isLive, setIsLive] = useState(follow);
  const [status, setStatus] = useState<'idle' | 'connecting' | 'connected' | 'error'>(
    'idle'
  );
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const listRef = useRef<HTMLDivElement | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldReconnectRef = useRef(true);
  const reconnectAttemptRef = useRef(0);

  const appendLines = useCallback((lines: string[]) => {
    if (lines.length === 0) return;
    setLogLines((prev) => {
      const next = [...prev];
      for (const line of lines) {
        if (!line.trim()) continue;
        next.push(parseLogLine(line));
      }
      if (next.length > MAX_LOG_LINES) {
        return next.slice(next.length - MAX_LOG_LINES);
      }
      return next;
    });
  }, []);

  const handleSocketMessage = useCallback(
    async (event: MessageEvent) => {
      if (typeof event.data === 'string') {
        appendLines(event.data.split('\n'));
        return;
      }
      if (event.data instanceof Blob) {
        const text = await event.data.text();
        appendLines(text.split('\n'));
        return;
      }
      if (event.data instanceof ArrayBuffer) {
        const text = new TextDecoder().decode(event.data);
        appendLines(text.split('\n'));
      }
    },
    [appendLines]
  );

  const disconnectSocket = useCallback(() => {
    if (reconnectTimerRef.current) {
      clearTimeout(reconnectTimerRef.current);
      reconnectTimerRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const connectSocket = useCallback(() => {
    if (!deploymentId || !isLive) return;

    disconnectSocket();
    setStatus('connecting');
    setErrorMessage(null);

    const wsBase = toWebSocketUrl(env.apiUrl);
    const wsUrl = new URL(`${wsBase}/deployments/${deploymentId}/logs`);
    wsUrl.searchParams.set('tail', String(tail));
    wsUrl.searchParams.set('follow', '1');
    if (containerName) {
      wsUrl.searchParams.set('container', containerName);
    }
    const token = getAuthToken();
    if (token) {
      wsUrl.searchParams.set('token', token);
    }

    const ws = new WebSocket(wsUrl.toString());
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus('connected');
      reconnectAttemptRef.current = 0;
    };

    ws.onmessage = handleSocketMessage;

    ws.onerror = () => {
      setStatus('error');
      setErrorMessage('Log stream error');
    };

    ws.onclose = () => {
      if (!shouldReconnectRef.current || !isLive) return;
      setStatus('error');
      setErrorMessage('Disconnected. Reconnecting...');
      const delay =
        WS_BACKOFF_MS[Math.min(reconnectAttemptRef.current, WS_BACKOFF_MS.length - 1)];
      reconnectTimerRef.current = setTimeout(() => {
        reconnectAttemptRef.current += 1;
        connectSocket();
      }, delay);
    };
  }, [
    containerName,
    deploymentId,
    disconnectSocket,
    handleSocketMessage,
    isLive,
    tail,
  ]);

  const fetchLogs = useCallback(async () => {
    if (!deploymentId) return;
    setStatus('connecting');
    setErrorMessage(null);
    try {
      const url = new URL(`${env.apiUrl}/deployments/${deploymentId}/logs`);
      url.searchParams.set('tail', String(tail));
      if (containerName) {
        url.searchParams.set('container', containerName);
      }
      const response = await fetch(url.toString(), {
        headers: buildAuthHeaders(),
      });
      if (!response.ok) {
        throw new Error(`Failed to load logs (${response.status})`);
      }
      const text = await response.text();
      setLogLines(text.split('\n').filter(Boolean).map(parseLogLine));
      setStatus('connected');
    } catch (error) {
      setStatus('error');
      setErrorMessage(error instanceof Error ? error.message : 'Failed to load logs');
    }
  }, [containerName, deploymentId, tail]);

  useEffect(() => {
    shouldReconnectRef.current = true;
    if (isLive) {
      connectSocket();
    } else {
      disconnectSocket();
      void fetchLogs();
    }

    return () => {
      shouldReconnectRef.current = false;
      disconnectSocket();
    };
  }, [connectSocket, disconnectSocket, fetchLogs, isLive]);

  useEffect(() => {
    if (!autoScroll || !listRef.current) return;
    listRef.current.scrollTop = listRef.current.scrollHeight;
  }, [autoScroll, logLines]);

  const displayedLines = useMemo(() => {
    const term = searchTerm.trim().toLowerCase();
    return logLines.filter((line) => {
      if (!selectedLevels.has(line.level)) return false;
      if (!term) return true;
      return line.raw.toLowerCase().includes(term);
    });
  }, [logLines, searchTerm, selectedLevels]);

  const toggleLevel = (level: LogLevel) => {
    setSelectedLevels((prev) => {
      const next = new Set(prev);
      if (next.has(level)) {
        next.delete(level);
      } else {
        next.add(level);
      }
      return next;
    });
  };

  const clearLogs = () => {
    setLogLines([]);
  };

  const downloadLogs = () => {
    if (displayedLines.length === 0) return;
    const content = displayedLines.map((line) => line.raw).join('\n');
    const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
    const href = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = href;
    link.download = `deployment-${deploymentId}-logs.txt`;
    link.click();
    URL.revokeObjectURL(href);
  };

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card">
      <div className="border-b border-border p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 className="text-sm font-semibold">Live Logs</h3>
            <p className="text-xs text-muted-foreground">
              {containerName ? `Container: ${containerName}` : 'All containers'}
            </p>
          </div>
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <span
              className={cn(
                'inline-flex items-center gap-2 rounded-full border px-2 py-1',
                status === 'connected' ? 'border-success text-success' : 'border-muted'
              )}
            >
              <span
                className={cn(
                  'h-2 w-2 rounded-full',
                  status === 'connected' ? 'bg-success' : 'bg-muted-foreground'
                )}
              />
              {status === 'connected' ? 'Connected' : status === 'connecting' ? 'Connecting' : 'Offline'}
            </span>
            {errorMessage && <span>{errorMessage}</span>}
          </div>
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-2">
          {LOG_LEVELS.map((level) => {
            const active = selectedLevels.has(level);
            return (
              <button
                key={level}
                type="button"
                onClick={() => toggleLevel(level)}
                className={cn(
                  'rounded-full border px-3 py-1 text-xs font-medium',
                  active ? 'border-primary text-primary' : 'border-muted text-muted-foreground'
                )}
              >
                {LEVEL_LABELS[level]}
              </button>
            );
          })}

          <div className="ml-auto flex flex-wrap items-center gap-2">
            <label className="flex items-center gap-2 text-xs text-muted-foreground">
              <input
                type="checkbox"
                checked={isLive}
                onChange={(event) => setIsLive(event.target.checked)}
              />
              Live
            </label>
            <label className="flex items-center gap-2 text-xs text-muted-foreground">
              <input
                type="checkbox"
                checked={autoScroll}
                onChange={(event) => setAutoScroll(event.target.checked)}
              />
              Auto-scroll
            </label>
            <button
              type="button"
              onClick={clearLogs}
              className="rounded-full border border-muted px-3 py-1 text-xs text-muted-foreground hover:border-primary hover:text-primary"
            >
              Clear
            </button>
            <button
              type="button"
              onClick={downloadLogs}
              className="rounded-full border border-muted px-3 py-1 text-xs text-muted-foreground hover:border-primary hover:text-primary"
            >
              Download
            </button>
          </div>
        </div>

        <div className="mt-3">
          <input
            type="text"
            value={searchTerm}
            onChange={(event) => setSearchTerm(event.target.value)}
            placeholder="Search logs"
            className="w-full rounded-md border border-border bg-background px-3 py-2 text-sm focus:border-primary focus:outline-none"
          />
        </div>
      </div>

      <div ref={listRef} className="flex-1 overflow-y-auto p-4 font-mono text-xs leading-relaxed">
        {displayedLines.length === 0 ? (
          <div className="text-center text-sm text-muted-foreground">No logs yet.</div>
        ) : (
          displayedLines.map((line) => (
            <div key={line.id} className="flex flex-wrap gap-2">
              {line.timestamp && (
                <span className="text-muted-foreground">{line.timestamp}</span>
              )}
              <span className={cn('uppercase', LEVEL_STYLES[line.level])}>
                {line.level}
              </span>
              <span className="text-foreground">{line.message}</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
