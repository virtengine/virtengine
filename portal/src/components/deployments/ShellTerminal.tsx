'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

import { env } from '@/config/env';
import { cn } from '@/lib/utils';

export interface ShellTerminalProps {
  deploymentId: string;
  containerName: string;
  onDisconnect?: () => void;
}

const SHELL_STDOUT = 100;
const SHELL_STDERR = 101;
const SHELL_RESULT = 102;
const SHELL_FAILURE = 103;
const SHELL_STDIN = 104;
const SHELL_RESIZE = 105;

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  return (
    window.localStorage.getItem('ve_session_token') ??
    window.localStorage.getItem('ve_portal_token') ??
    null
  );
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

export function ShellTerminal({ deploymentId, containerName, onDisconnect }: ShellTerminalProps) {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const termInstanceRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const dataListenerRef = useRef<() => void>();
  const resizeObserverRef = useRef<ResizeObserver | null>(null);

  const [status, setStatus] = useState<'idle' | 'connecting' | 'connected' | 'closed' | 'error'>(
    'idle'
  );
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const sendResize = useCallback(() => {
    const term = termInstanceRef.current;
    const ws = wsRef.current;
    if (!term || !ws || ws.readyState !== WebSocket.OPEN) return;

    const payload = new ArrayBuffer(5);
    const view = new DataView(payload);
    view.setUint8(0, SHELL_RESIZE);
    view.setUint16(1, term.cols, false);
    view.setUint16(3, term.rows, false);
    ws.send(payload);
  }, []);

  const writeSystemLine = useCallback((text: string) => {
    termInstanceRef.current?.writeln(`\x1b[38;5;244m${text}\x1b[0m`);
  }, []);

  const closeSocket = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const connectSocket = useCallback(() => {
    if (!deploymentId || !containerName) return;
    const term = termInstanceRef.current;
    if (!term) return;

    closeSocket();
    setStatus('connecting');
    setErrorMessage(null);

    const wsBase = toWebSocketUrl(env.apiUrl);
    const wsUrl = new URL(`${wsBase}/deployments/${deploymentId}/shell`);
    wsUrl.searchParams.set('container', containerName);
    wsUrl.searchParams.set('tty', '1');
    wsUrl.searchParams.set('stdin', '1');
    const token = getAuthToken();
    if (token) {
      wsUrl.searchParams.set('token', token);
    }

    const ws = new WebSocket(wsUrl.toString());
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.onopen = () => {
      setStatus('connected');
      writeSystemLine('Shell session connected.');
      sendResize();
    };

    ws.onmessage = (event: MessageEvent<ArrayBuffer>) => {
      if (!(event.data instanceof ArrayBuffer)) {
        return;
      }
      const view = new DataView(event.data);
      const messageType = view.getUint8(0);
      const payload = new Uint8Array(event.data.slice(1));

      switch (messageType) {
        case SHELL_STDOUT:
        case SHELL_STDERR: {
          const text = new TextDecoder().decode(payload);
          term.write(text);
          break;
        }
        case SHELL_RESULT: {
          const text = new TextDecoder().decode(payload);
          writeSystemLine(`Session ended: ${text}`);
          break;
        }
        case SHELL_FAILURE: {
          writeSystemLine('Provider reported a shell failure.');
          break;
        }
        default:
          break;
      }
    };

    ws.onclose = (event) => {
      setStatus('closed');
      const reason = event.reason || 'Shell session closed.';
      writeSystemLine(reason);
      onDisconnect?.();
    };

    ws.onerror = () => {
      setStatus('error');
      setErrorMessage('Shell connection error.');
      writeSystemLine('Shell connection error.');
    };
  }, [closeSocket, containerName, deploymentId, onDisconnect, sendResize, writeSystemLine]);

  useEffect(() => {
    if (!terminalRef.current) return;

    const terminal = new Terminal({
      cursorBlink: true,
      fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
      fontSize: 13,
      theme: {
        background: '#0c0f12',
        foreground: '#d6e1ea',
        cursor: '#d6e1ea',
      },
    });
    const fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(terminalRef.current);
    fitAddon.fit();

    termInstanceRef.current = terminal;
    fitAddonRef.current = fitAddon;

    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
      sendResize();
    });
    resizeObserver.observe(terminalRef.current);
    resizeObserverRef.current = resizeObserver;

    const disposable = terminal.onData((data) => {
      const ws = wsRef.current;
      if (!ws || ws.readyState !== WebSocket.OPEN) return;
      const payload = new TextEncoder().encode(data);
      const buffer = new Uint8Array(payload.length + 1);
      buffer[0] = SHELL_STDIN;
      buffer.set(payload, 1);
      ws.send(buffer);
    });
    dataListenerRef.current = () => disposable.dispose();

    return () => {
      dataListenerRef.current?.();
      resizeObserver.disconnect();
      terminal.dispose();
      termInstanceRef.current = null;
      fitAddonRef.current = null;
    };
  }, [sendResize]);

  useEffect(() => {
    if (!termInstanceRef.current) return;
    connectSocket();
    return () => {
      closeSocket();
    };
  }, [closeSocket, connectSocket, containerName, deploymentId]);

  return (
    <div className="flex h-full flex-col rounded-lg border border-border bg-card">
      <div className="flex flex-wrap items-center justify-between gap-3 border-b border-border p-4 text-xs text-muted-foreground">
        <div>
          <div className="text-sm font-semibold text-foreground">Shell Session</div>
          <div>{containerName}</div>
        </div>
        <div className="flex items-center gap-2">
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
            {status === 'connected' ? 'Connected' : status}
          </span>
          {errorMessage && <span>{errorMessage}</span>}
          <button
            type="button"
            onClick={connectSocket}
            className="rounded-full border border-muted px-3 py-1 text-xs text-muted-foreground hover:border-primary hover:text-primary"
          >
            Reconnect
          </button>
        </div>
      </div>
      <div className="flex-1 bg-[#0c0f12]">
        <div ref={terminalRef} className="h-full w-full" />
      </div>
    </div>
  );
}
