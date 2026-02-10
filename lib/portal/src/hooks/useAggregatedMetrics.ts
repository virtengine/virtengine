import { useCallback, useEffect, useMemo, useState } from "react";
import type { AggregatedMetrics } from "../multi-provider/types";
import { useMultiProvider } from "../multi-provider/context";

export interface AggregatedMetricsState {
  metrics: AggregatedMetrics | null;
  isLoading: boolean;
  error: string | null;
  lastUpdatedAt: number | null;
}

export interface AggregatedMetricsActions {
  refresh: () => Promise<void>;
}

export interface UseAggregatedMetricsOptions {
  pollIntervalMs?: number;
  autoRefresh?: boolean;
}

export function useAggregatedMetrics(
  options: UseAggregatedMetricsOptions = {},
) {
  const { client, isInitialized } = useMultiProvider();
  const [state, setState] = useState<AggregatedMetricsState>({
    metrics: null,
    isLoading: false,
    error: null,
    lastUpdatedAt: null,
  });

  const refresh = useCallback(async () => {
    if (!client) return;
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const metrics = await client.getAggregatedMetrics();
      setState({
        metrics,
        isLoading: false,
        error: null,
        lastUpdatedAt: Date.now(),
      });
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error:
          error instanceof Error ? error.message : "Failed to load metrics",
      }));
    }
  }, [client]);

  useEffect(() => {
    if (!client || !isInitialized) return;
    void refresh();
  }, [client, isInitialized, refresh]);

  useEffect(() => {
    if (
      !client ||
      !isInitialized ||
      !options.autoRefresh ||
      !options.pollIntervalMs
    )
      return;
    const interval = setInterval(() => {
      void refresh();
    }, options.pollIntervalMs);
    return () => clearInterval(interval);
  }, [
    client,
    isInitialized,
    options.autoRefresh,
    options.pollIntervalMs,
    refresh,
  ]);

  const actions = useMemo(() => ({ refresh }), [refresh]);

  return { state, actions };
}
