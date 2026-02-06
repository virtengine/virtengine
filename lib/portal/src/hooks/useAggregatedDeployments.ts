import { useCallback, useEffect, useMemo, useState } from "react";
import type { DeploymentWithProvider } from "../multi-provider/types";
import { useMultiProvider } from "../multi-provider/context";

export interface AggregatedDeploymentsState {
  deployments: DeploymentWithProvider[];
  isLoading: boolean;
  error: string | null;
  lastUpdatedAt: number | null;
}

export interface AggregatedDeploymentsActions {
  refresh: () => Promise<void>;
}

export interface UseAggregatedDeploymentsOptions {
  pollIntervalMs?: number;
  autoRefresh?: boolean;
  status?: string;
}

export function useAggregatedDeployments(
  options: UseAggregatedDeploymentsOptions = {},
) {
  const { client, isInitialized } = useMultiProvider();
  const [state, setState] = useState<AggregatedDeploymentsState>({
    deployments: [],
    isLoading: false,
    error: null,
    lastUpdatedAt: null,
  });

  const refresh = useCallback(async () => {
    if (!client) return;
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const deployments = await client.listAllDeployments({
        refresh: true,
        status: options.status,
      });
      setState({
        deployments,
        isLoading: false,
        error: null,
        lastUpdatedAt: Date.now(),
      });
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isLoading: false,
        error:
          error instanceof Error ? error.message : "Failed to load deployments",
      }));
    }
  }, [client, options.status]);

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
