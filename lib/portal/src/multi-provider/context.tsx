import React, {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { MultiProviderClient } from "./client";
import type { MultiProviderClientOptions, ProviderRecord } from "./types";

interface MultiProviderContextValue {
  client: MultiProviderClient | null;
  providers: ProviderRecord[];
  isInitialized: boolean;
  error: Error | null;
  refreshProviders: () => Promise<void>;
}

const MultiProviderContext = createContext<MultiProviderContextValue | null>(
  null,
);

export interface MultiProviderProviderProps {
  options: MultiProviderClientOptions;
  children: React.ReactNode;
}

export function MultiProviderProvider({
  options,
  children,
}: MultiProviderProviderProps) {
  const [client, setClient] = useState<MultiProviderClient | null>(null);
  const [providers, setProviders] = useState<ProviderRecord[]>([]);
  const [isInitialized, setIsInitialized] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const memoOptions = useMemo(
    () => ({
      chainEndpoint: options.chainEndpoint,
      wallet: options.wallet,
      healthCheckIntervalMs: options.healthCheckIntervalMs,
      providerCacheTtlMs: options.providerCacheTtlMs,
      deploymentCacheTtlMs: options.deploymentCacheTtlMs,
      requestTimeoutMs: options.requestTimeoutMs,
      fetcher: options.fetcher,
    }),
    [
      options.chainEndpoint,
      options.wallet,
      options.healthCheckIntervalMs,
      options.providerCacheTtlMs,
      options.deploymentCacheTtlMs,
      options.requestTimeoutMs,
      options.fetcher,
    ],
  );

  useEffect(() => {
    let isActive = true;
    const multiClient = new MultiProviderClient(memoOptions);
    setClient(multiClient);
    setIsInitialized(false);
    setError(null);

    const unsubscribe = multiClient.subscribe((snapshot) => {
      if (!isActive) return;
      setProviders(snapshot);
    });

    multiClient
      .initialize()
      .then(() => {
        if (!isActive) return;
        setIsInitialized(true);
      })
      .catch((err) => {
        if (!isActive) return;
        setError(err as Error);
      });

    return () => {
      isActive = false;
      unsubscribe();
      multiClient.destroy();
    };
  }, [memoOptions]);

  const refreshProviders = React.useCallback(async () => {
    if (!client) return;
    await client.refreshProviders(true);
  }, [client]);

  return (
    <MultiProviderContext.Provider
      value={{
        client,
        providers,
        isInitialized,
        error,
        refreshProviders,
      }}
    >
      {children}
    </MultiProviderContext.Provider>
  );
}

export function useMultiProvider(): MultiProviderContextValue {
  const context = useContext(MultiProviderContext);
  if (!context) {
    throw new Error(
      "useMultiProvider must be used within MultiProviderProvider",
    );
  }
  return context;
}
