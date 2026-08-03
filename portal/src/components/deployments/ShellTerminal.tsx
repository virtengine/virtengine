'use client';

import { useCallback, useEffect, useRef } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';

import {
  useDeploymentShell,
  type ProviderShellSessionCapability,
  type ShellEligibilityProjection,
} from '@/lib/portal-adapter';
import { cn } from '@/lib/utils';

export interface ShellTerminalProps {
  deploymentId: string;
  containerName: string;
  providerEndpoint?: string;
  capability?: ProviderShellSessionCapability;
  eligibility?: ShellEligibilityProjection;
  onDisconnect?: () => void;
}

const SHELL_STDOUT = 100;
const SHELL_STDERR = 101;
const SHELL_RESULT = 102;
const SHELL_FAILURE = 103;
const SHELL_STDIN = 104;
const SHELL_RESIZE = 105;

export function ShellTerminal({
  deploymentId,
  containerName,
  providerEndpoint,
  capability,
  eligibility,
  onDisconnect,
}: ShellTerminalProps) {
  const terminalRef = useRef<HTMLDivElement | null>(null);
  const termInstanceRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const dataListenerRef = useRef<() => void>();
  const resizeObserverRef = useRef<ResizeObserver | null>(null);
  const wasConnectedRef = useRef(false);
  const authorityMatchesProps =
    eligibility?.deploymentId === deploymentId && eligibility?.container === containerName;
  const shell = useDeploymentShell({
    endpoint: providerEndpoint ?? 'https://shell-authority-unavailable.invalid',
    shellSessionCapability: capability,
    eligibility: authorityMatchesProps ? eligibility : undefined,
  });
  const { isConnected, isConnecting, error, send, onData } = shell;

  const sendResize = useCallback(() => {
    const term = termInstanceRef.current;
    if (!term) return;

    const payload = new ArrayBuffer(5);
    const view = new DataView(payload);
    view.setUint8(0, SHELL_RESIZE);
    view.setUint16(1, term.cols, false);
    view.setUint16(3, term.rows, false);
    send(payload);
  }, [send]);

  const writeSystemLine = useCallback((text: string) => {
    termInstanceRef.current?.writeln(`\x1b[38;5;244m${text}\x1b[0m`);
  }, []);

  const handleShellData = useCallback(
    (data: ArrayBuffer) => {
      const term = termInstanceRef.current;
      if (!term) return;
      if (data.byteLength === 0) return;
      const view = new DataView(data);
      const messageType = view.getUint8(0);
      const payload = new Uint8Array(data.slice(1));

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
    },
    [writeSystemLine]
  );

  useEffect(() => {
    if (!terminalRef.current) return;

    const terminal = new Terminal({
      cursorBlink: true,
      fontFamily:
        'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace',
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
      const payload = new TextEncoder().encode(data);
      const buffer = new Uint8Array(payload.length + 1);
      buffer[0] = SHELL_STDIN;
      buffer.set(payload, 1);
      send(buffer);
    });
    dataListenerRef.current = () => disposable.dispose();

    return () => {
      dataListenerRef.current?.();
      resizeObserver.disconnect();
      terminal.dispose();
      termInstanceRef.current = null;
      fitAddonRef.current = null;
    };
  }, [send, sendResize]);

  useEffect(() => {
    onData(handleShellData);
  }, [handleShellData, onData]);

  useEffect(() => {
    if (isConnected) {
      writeSystemLine('Shell session connected.');
      sendResize();
    } else if (wasConnectedRef.current) {
      writeSystemLine('Shell session closed.');
      onDisconnect?.();
    }
    wasConnectedRef.current = isConnected;
  }, [isConnected, onDisconnect, sendResize, writeSystemLine]);

  const status = isConnected
    ? 'connected'
    : isConnecting
      ? 'connecting'
      : error
        ? 'unavailable'
        : 'closed';

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
          {error && <span>{error.message}</span>}
        </div>
      </div>
      <div className="flex-1 bg-[#0c0f12]">
        <div ref={terminalRef} className="h-full w-full" />
      </div>
    </div>
  );
}
