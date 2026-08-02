import { useCallback, useEffect, useRef, useState } from "react";
import type { ProviderAPIClientOptions } from "../provider-api/client";
import type { ShellConnection } from "../provider-api/client";
import type { ShellEligibilityProjection } from "../provider-api/shell-session";
import { ProviderShellSessionError } from "../provider-api/shell-session";
import { useProviderAPI } from "./useProviderAPI";

export interface UseDeploymentShellOptions extends ProviderAPIClientOptions {
  eligibility?: ShellEligibilityProjection;
  /** Set to `false` to defer connection. */
  enabled?: boolean;
}

export interface UseDeploymentShellResult {
  isConnected: boolean;
  isConnecting: boolean;
  error: Error | null;
  expiresAt: Date | null;
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
  const { eligibility, enabled = true } = options;
  const client = useProviderAPI(options);

  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [expiresAt, setExpiresAt] = useState<Date | null>(null);
  const shellRef = useRef<ShellConnection | null>(null);
  const dataCallbackRef = useRef<((data: ArrayBuffer) => void) | null>(null);

  useEffect(() => {
    if (!enabled) return;

    let cancelled = false;
    let expiryTimer: ReturnType<typeof setTimeout> | undefined;
    setError(null);
    setIsConnecting(true);

    (async () => {
      try {
        if (!eligibility) {
          throw new ProviderShellSessionError(
            "eligibility_unavailable",
            "Authoritative shell eligibility is unavailable",
          );
        }
        const receipt = await client.createShellSession(eligibility);
        const expiry = new Date(receipt.expiresAt);
        const shell = client.connectShell(receipt);
        if (cancelled) {
          shell.close();
          return;
        }
        shellRef.current = shell;
        setExpiresAt(expiry);
        expiryTimer = setTimeout(
          () => {
            shell.close();
            shellRef.current = null;
            setIsConnected(false);
            setExpiresAt(null);
            setError(
              new ProviderShellSessionError(
                "receipt_expired",
                "Provider shell session expired",
              ),
            );
          },
          Math.max(0, expiry.getTime() - Date.now()),
        );

        shell.onOpen(() => {
          if (!cancelled) {
            setIsConnecting(false);
            setIsConnected(true);
          }
        });

        shell.onMessage((data: ArrayBuffer) => {
          dataCallbackRef.current?.(data);
        });

        shell.onClose(() => {
          if (!cancelled) {
            if (expiryTimer) clearTimeout(expiryTimer);
            shellRef.current = null;
            setIsConnecting(false);
            setIsConnected(false);
            setExpiresAt(null);
          }
        });

        shell.onError(() => {
          if (!cancelled) setError(new Error("Shell connection error"));
        });
      } catch (err) {
        if (!cancelled) {
          setIsConnecting(false);
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      }
    })();

    return () => {
      cancelled = true;
      if (expiryTimer) clearTimeout(expiryTimer);
      shellRef.current?.close();
      shellRef.current = null;
      setIsConnected(false);
      setIsConnecting(false);
      setExpiresAt(null);
    };
  }, [client, eligibility, enabled]);

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
    setExpiresAt(null);
  }, []);

  return {
    isConnected,
    isConnecting,
    error,
    expiresAt,
    send,
    onData,
    disconnect,
  };
}
