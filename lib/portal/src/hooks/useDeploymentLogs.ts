import { useCallback, useEffect, useRef, useState } from "react";
import type { ProviderAPIClientOptions } from "../provider-api/client";
import type { LogStream } from "../provider-api/client";
import type { LogOptions } from "../provider-api/types";
import { useProviderAPI } from "./useProviderAPI";

export interface UseDeploymentLogsOptions extends ProviderAPIClientOptions {
  leaseId: string;
  /** Maximum number of log lines to keep in state. Oldest are dropped first. */
  maxLines?: number;
  /** Options forwarded to the log stream (tail, level, etc.). */
  logOptions?: LogOptions;
  /** Set to `false` to defer connection. */
  enabled?: boolean;
}

export interface UseDeploymentLogsResult {
  logs: string[];
  isConnected: boolean;
  error: Error | null;
  clear: () => void;
}

export function useDeploymentLogs(
  options: UseDeploymentLogsOptions,
): UseDeploymentLogsResult {
  const { leaseId, maxLines, logOptions, enabled = true } = options;
  const client = useProviderAPI(options);

  const [logs, setLogs] = useState<string[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const streamRef = useRef<LogStream | null>(null);

  useEffect(() => {
    if (!enabled || !leaseId) return;

    let cancelled = false;

    (async () => {
      try {
        const stream = await client.connectLogStream(leaseId, logOptions);
        if (cancelled) {
          stream.close();
          return;
        }
        streamRef.current = stream;

        stream.onOpen(() => {
          if (!cancelled) setIsConnected(true);
        });

        stream.onMessage((line: string) => {
          if (cancelled) return;
          setLogs((prev) => {
            const next = [...prev, line];
            return maxLines && next.length > maxLines
              ? next.slice(-maxLines)
              : next;
          });
        });

        stream.onClose(() => {
          if (!cancelled) setIsConnected(false);
        });

        stream.onError(() => {
          if (!cancelled)
            setError(new Error("Log stream connection error"));
        });
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err : new Error(String(err)));
        }
      }
    })();

    return () => {
      cancelled = true;
      streamRef.current?.close();
      streamRef.current = null;
      setIsConnected(false);
    };
  }, [client, leaseId, enabled, maxLines, logOptions]);

  const clear = useCallback(() => setLogs([]), []);

  return { logs, isConnected, error, clear };
}
