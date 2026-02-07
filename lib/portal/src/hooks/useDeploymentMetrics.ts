import { useCallback, useEffect, useRef, useState } from "react";
import type { ProviderAPIClientOptions } from "../provider-api/client";
import type { ResourceMetrics } from "../provider-api/types";
import { useProviderAPI } from "./useProviderAPI";

export interface UseDeploymentMetricsOptions extends ProviderAPIClientOptions {
  leaseId: string;
  /** Polling interval in milliseconds. Defaults to 5 000 (5 s). */
  pollingInterval?: number;
  /** Set to `false` to disable automatic polling. */
  enabled?: boolean;
}

export interface UseDeploymentMetricsResult {
  metrics: ResourceMetrics | null;
  isLoading: boolean;
  error: Error | null;
  refresh: () => Promise<void>;
}

export function useDeploymentMetrics(
  options: UseDeploymentMetricsOptions,
): UseDeploymentMetricsResult {
  const { leaseId, pollingInterval = 5_000, enabled = true } = options;
  const client = useProviderAPI(options);

  const [metrics, setMetrics] = useState<ResourceMetrics | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const mountedRef = useRef(true);

  const refresh = useCallback(async () => {
    if (!leaseId) return;
    setIsLoading(true);
    setError(null);
    try {
      const result = await client.getDeploymentMetrics(leaseId);
      if (mountedRef.current) {
        setMetrics(result);
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

  return { metrics, isLoading, error, refresh };
}
