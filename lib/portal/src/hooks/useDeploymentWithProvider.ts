import { useCallback, useEffect, useState } from "react";
import type { DeploymentWithProvider } from "../multi-provider/types";
import { useMultiProvider } from "../multi-provider/context";

export interface DeploymentWithProviderState {
  deployment: DeploymentWithProvider | null;
  isLoading: boolean;
  error: string | null;
}

export function useDeploymentWithProvider(deploymentId: string | null) {
  const { client, isInitialized } = useMultiProvider();
  const [state, setState] = useState<DeploymentWithProviderState>({
    deployment: null,
    isLoading: false,
    error: null,
  });

  const refresh = useCallback(async () => {
    if (!client || !deploymentId) return;
    setState((prev) => ({ ...prev, isLoading: true, error: null }));

    try {
      const deployment = await client.getDeployment(deploymentId);
      setState({ deployment, isLoading: false, error: null });
    } catch (error) {
      setState({
        deployment: null,
        isLoading: false,
        error:
          error instanceof Error ? error.message : "Failed to load deployment",
      });
    }
  }, [client, deploymentId]);

  useEffect(() => {
    if (!client || !isInitialized || !deploymentId) return;
    void refresh();
  }, [client, isInitialized, deploymentId, refresh]);

  return { state, refresh };
}
