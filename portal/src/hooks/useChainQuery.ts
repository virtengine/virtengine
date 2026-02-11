'use client';

import { useCallback, useEffect, useState } from 'react';
import type { VirtEngineClient } from '@virtengine/chain-sdk';
import { getChainClient } from '@/lib/chain-sdk';

export interface ChainQueryState<T> {
  data: T | null;
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
}

export function useChainQuery<T>(
  query: (client: VirtEngineClient) => Promise<T>,
  deps: React.DependencyList = []
): ChainQueryState<T> {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  const execute = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const client = await getChainClient();
      const result = await query(client);
      setData(result);
    } catch (err) {
      setError(err as Error);
    } finally {
      setIsLoading(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [query, ...deps]);

  useEffect(() => {
    void execute();
  }, [execute]);

  return { data, isLoading, error, refetch: execute };
}
