import { useCallback, useEffect, useRef, useState } from "react";
import type { ProviderAPIClientOptions } from "../provider-api/client";
import type { ShellConnection } from "../provider-api/client";
import { useProviderAPI } from "./useProviderAPI";

export interface UseDeploymentShellOptions extends ProviderAPIClientOptions {
  leaseId: string;
  /** Pre-created session token for the shell connection. */
  sessionToken?: string;
  /** Container / service name within the deployment. */
  container?: string;
  /** Set to `false` to defer connection. */
  enabled?: boolean;
}

export interface UseDeploymentShellResult {
  isConnected: boolean;
  error: Error | null;
  /** Send raw data (keystrokes) to the shell. */
  send: (data: ArrayBufferLike | ArrayBufferView) => void;
  /** Register a handler for incoming shell data. */
  onData: (callback: (data: ArrayBuffer) => void) => void;
  /** Disconnect the shell session. */
  disconnect: () => void;
}

export function useDeploymentShell(
  options: UseDeploymentShellOptions,
): UseDeploymentShellResult {
  const { leaseId, sessionToken, container, enabled = true } = options;
  const client = useProviderAPI(options);

  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const shellRef = useRef<ShellConnection | null>(null);
  const dataCallbackRef = useRef<((data: ArrayBuffer) => void) | null>(null);

  useEffect(() => {
    if (!enabled || !leaseId) return;

    let cancelled = false;

    (async () => {
      try {
        const shell = await client.connectShell(
          leaseId,
          sessionToken,
          container,
        );
        if (cancelled) {
          shell.close();
          return;
        }
        shellRef.current = shell;

        shell.onOpen(() => {
          if (!cancelled) setIsConnected(true);
        });

        shell.onMessage((data: ArrayBuffer) => {
          dataCallbackRef.current?.(data);
        });

        shell.onClose(() => {
          if (!cancelled) setIsConnected(false);
        });

        shell.onError(() => {
          if (!cancelled) setError(new Error("Shell connection error"));
        });
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      }
    })();

    return () => {
      cancelled = true;
      shellRef.current?.close();
      shellRef.current = null;
      setIsConnected(false);
    };
  }, [client, leaseId, sessionToken, container, enabled]);

  const send = useCallback((data: ArrayBufferLike | ArrayBufferView) => {
    shellRef.current?.send(data);
  }, []);

  const onData = useCallback((callback: (data: ArrayBuffer) => void) => {
    dataCallbackRef.current = callback;
  }, []);

  const disconnect = useCallback(() => {
    shellRef.current?.close();
    shellRef.current = null;
    setIsConnected(false);
  }, []);

  return { isConnected, error, send, onData, disconnect };
}
