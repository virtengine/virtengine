import { useCallback, useEffect, useRef, useState } from "react";
import type { ProviderAPIClientOptions } from "../provider-api/client";
import type { DeploymentStatus } from "../provider-api/types";
import { useProviderAPI } from "./useProviderAPI";

export interface UseDeploymentStatusOptions extends ProviderAPIClientOptions {
  leaseId: string;
  /** Polling interval in milliseconds. Defaults to 10 000 (10 s). */
  pollingInterval?: number;
  /** Set to `false` to disable automatic polling. */
  enabled?: boolean;
}

export interface UseDeploymentStatusResult {
  status: DeploymentStatus | null;
  isLoading: boolean;
  error: Error | null;
  refresh: () => Promise<void>;
}

export function useDeploymentStatus(
  options: UseDeploymentStatusOptions,
): UseDeploymentStatusResult {
  const { leaseId, pollingInterval = 10_000, enabled = true } = options;
  const client = useProviderAPI(options);

  const [status, setStatus] = useState<DeploymentStatus | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const mountedRef = useRef(true);

  const refresh = useCallback(async () => {
    if (!leaseId) return;
    setIsLoading(true);
    setError(null);
    try {
      const result = await client.getDeploymentStatus(leaseId);
      if (mountedRef.current) {
        setStatus(result);
      }
    } catch (err) {
      if (mountedRef.current) {
        setError(err instanceof Error ? err : new Error(String(err)));
      }
    } finally {
      if (mountedRef.current) {
        setIsLoading(false);
      }
    }
  }, [client, leaseId]);

  useEffect(() => {
    mountedRef.current = true;
    return () => {
      mountedRef.current = false;
    };
  }, []);

  useEffect(() => {
    if (!enabled || !leaseId) return;

    void refresh();
    const timer = setInterval(() => void refresh(), pollingInterval);
    return () => clearInterval(timer);
  }, [enabled, leaseId, pollingInterval, refresh]);

  return { status, isLoading, error, refresh };
}
