/**
 * useChain Hook
 * VE-700: Chain RPC/WebSocket integration
 */

import { useState, useCallback, useEffect, useContext, createContext, useRef } from 'react';
import type { ReactNode } from 'react';
import type {
  ChainState,
  ChainConfig,
  EventSubscription,
  QueryClient,
  ChainEvent,
  AccountInfo,
  Balance,
  IdentityInfo,
  OfferingInfo,
  OrderInfo,
  JobInfo,
  ProviderInfo,
  TransactionResult,
} from '../types/chain';
import { initialChainState, defaultChainConfig } from '../types/chain';

interface ChainContextValue {
  state: ChainState;
  queryClient: QueryClient;
  actions: ChainActions;
}

interface ChainActions {
  connect: () => Promise<void>;
  disconnect: () => void;
  subscribe: (query: string, callback: (event: ChainEvent) => void) => string;
  unsubscribe: (subscriptionId: string) => void;
  broadcastTx: (txBytes: Uint8Array) => Promise<TransactionResult>;
}

const ChainContext = createContext<ChainContextValue | null>(null);

export interface ChainProviderProps {
  children: ReactNode;
  config: Partial<ChainConfig>;
}

export function ChainProvider({ children, config: userConfig }: ChainProviderProps) {
  const config: ChainConfig = { ...defaultChainConfig, ...userConfig };
  const [state, setState] = useState<ChainState>(initialChainState);
  const wsRef = useRef<WebSocket | null>(null);
  const subscriptionsRef = useRef<Map<string, (event: ChainEvent) => void>>(new Map());
  const reconnectAttemptsRef = useRef(0);

  const connect = useCallback(async () => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    setState(prev => ({ ...prev, isConnecting: true, error: null }));

    try {
      const ws = new WebSocket(config.wsEndpoint);

      ws.onopen = () => {
        reconnectAttemptsRef.current = 0;
        setState(prev => ({
          ...prev,
          isConnected: true,
          isConnecting: false,
          chainId: config.chainId,
          networkName: 'VirtEngine',
        }));

        // Resubscribe to all subscriptions
        for (const [id, callback] of subscriptionsRef.current) {
          const subscription = state.subscriptions.find(s => s.id === id);
          if (subscription) {
            ws.send(JSON.stringify({
              jsonrpc: '2.0',
              method: 'subscribe',
              params: { query: subscription.query },
              id: subscription.id,
            }));
          }
        }
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          if (data.result?.query) {
            const subscriptionId = data.id;
            const callback = subscriptionsRef.current.get(subscriptionId);
            if (callback) {
              const chainEvent: ChainEvent = {
                query: data.result.query,
                type: data.result.data?.type || 'unknown',
                attributes: data.result.data?.value?.TxResult?.result?.events?.[0]?.attributes?.reduce(
                  (acc: Record<string, string>, attr: { key: string; value: string }) => {
                    acc[attr.key] = attr.value;
                    return acc;
                  },
                  {}
                ) || {},
                blockHeight: data.result.data?.value?.block?.header?.height || 0,
                txHash: data.result.data?.value?.TxResult?.txHash,
                timestamp: Date.now(),
              };
              callback(chainEvent);
            }
          }

          if (data.result?.block?.header?.height) {
            setState(prev => ({
              ...prev,
              blockHeight: parseInt(data.result.block.header.height, 10),
            }));
          }
        } catch (error) {
          // Ignore parse errors
        }
      };

      ws.onerror = () => {
        setState(prev => ({
          ...prev,
          error: { code: 'connection_failed', message: 'WebSocket error' },
        }));
      };

      ws.onclose = () => {
        setState(prev => ({ ...prev, isConnected: false }));
        wsRef.current = null;

        if (config.autoReconnect && reconnectAttemptsRef.current < config.maxReconnectAttempts) {
          reconnectAttemptsRef.current++;
          setTimeout(connect, config.reconnectDelayMs * reconnectAttemptsRef.current);
        }
      };

      wsRef.current = ws;
    } catch (error) {
      setState(prev => ({
        ...prev,
        isConnecting: false,
        error: {
          code: 'connection_failed',
          message: error instanceof Error ? error.message : 'Connection failed',
        },
      }));
    }
  }, [config, state.subscriptions]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setState(initialChainState);
  }, []);

  const subscribe = useCallback((query: string, callback: (event: ChainEvent) => void): string => {
    const subscriptionId = `sub-${Date.now()}-${Math.random().toString(36).slice(2)}`;

    subscriptionsRef.current.set(subscriptionId, callback);

    const subscription: EventSubscription = {
      id: subscriptionId,
      query,
      status: 'pending',
      createdAt: Date.now(),
      eventsReceived: 0,
    };

    setState(prev => ({
      ...prev,
      subscriptions: [...prev.subscriptions, subscription],
    }));

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        jsonrpc: '2.0',
        method: 'subscribe',
        params: { query },
        id: subscriptionId,
      }));

      setState(prev => ({
        ...prev,
        subscriptions: prev.subscriptions.map(s =>
          s.id === subscriptionId ? { ...s, status: 'active' as const } : s
        ),
      }));
    }

    return subscriptionId;
  }, []);

  const unsubscribe = useCallback((subscriptionId: string) => {
    subscriptionsRef.current.delete(subscriptionId);

    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        jsonrpc: '2.0',
        method: 'unsubscribe',
        params: { query: state.subscriptions.find(s => s.id === subscriptionId)?.query },
        id: subscriptionId,
      }));
    }

    setState(prev => ({
      ...prev,
      subscriptions: prev.subscriptions.filter(s => s.id !== subscriptionId),
    }));
  }, [state.subscriptions]);

  const broadcastTx = useCallback(async (txBytes: Uint8Array): Promise<TransactionResult> => {
    const response = await fetch(`${config.restEndpoint}/cosmos/tx/v1beta1/txs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        tx_bytes: btoa(String.fromCharCode(...txBytes)),
        mode: 'BROADCAST_MODE_SYNC',
      }),
    });

    const result = await response.json();

    if (result.tx_response?.code !== 0) {
      throw new Error(result.tx_response?.raw_log || 'Transaction failed');
    }

    return {
      txHash: result.tx_response.txhash,
      blockHeight: result.tx_response.height ? parseInt(result.tx_response.height, 10) : null,
      code: result.tx_response.code,
      rawLog: result.tx_response.raw_log,
      events: result.tx_response.events || [],
      gasUsed: parseInt(result.tx_response.gas_used || '0', 10),
      gasWanted: parseInt(result.tx_response.gas_wanted || '0', 10),
    };
  }, [config.restEndpoint]);

  const queryClient: QueryClient = {
    queryAccount: async (address: string): Promise<AccountInfo> => {
      const response = await fetch(`${config.restEndpoint}/cosmos/auth/v1beta1/accounts/${address}`);
      const data = await response.json();
      return {
        address: data.account?.address || address,
        publicKey: data.account?.pub_key?.key || null,
        accountNumber: parseInt(data.account?.account_number || '0', 10),
        sequence: parseInt(data.account?.sequence || '0', 10),
      };
    },

    queryBalance: async (address: string, denom: string): Promise<Balance> => {
      const response = await fetch(`${config.restEndpoint}/cosmos/bank/v1beta1/balances/${address}/by_denom?denom=${denom}`);
      const data = await response.json();
      return {
        denom: data.balance?.denom || denom,
        amount: data.balance?.amount || '0',
      };
    },

    queryIdentity: async (address: string): Promise<IdentityInfo> => {
      const response = await fetch(`${config.restEndpoint}/virtengine/veid/v1/identity/${address}`);
      const data = await response.json();
      return {
        address,
        status: data.identity?.status || 'unknown',
        score: parseInt(data.identity?.score || '0', 10),
        modelVersion: data.identity?.model_version || '',
        updatedAt: parseInt(data.identity?.updated_at || '0', 10),
        blockHeight: parseInt(data.identity?.block_height || '0', 10),
      };
    },

    queryOffering: async (id: string): Promise<OfferingInfo> => {
      const response = await fetch(`${config.restEndpoint}/virtengine/market/v1/offerings/${id}`);
      const data = await response.json();
      return {
        id,
        providerAddress: data.offering?.provider_address || '',
        status: data.offering?.status || 'unknown',
        metadata: data.offering?.metadata || {},
        createdAt: parseInt(data.offering?.created_at || '0', 10),
      };
    },

    queryOrder: async (id: string): Promise<OrderInfo> => {
      const response = await fetch(`${config.restEndpoint}/virtengine/market/v1/orders/${id}`);
      const data = await response.json();
      return {
        id,
        offeringId: data.order?.offering_id || '',
        customerAddress: data.order?.customer_address || '',
        providerAddress: data.order?.provider_address || '',
        state: data.order?.state || 'unknown',
        createdAt: parseInt(data.order?.created_at || '0', 10),
      };
    },

    queryJob: async (id: string): Promise<JobInfo> => {
      const response = await fetch(`${config.restEndpoint}/virtengine/hpc/v1/jobs/${id}`);
      const data = await response.json();
      return {
        id,
        customerAddress: data.job?.customer_address || '',
        providerAddress: data.job?.provider_address || '',
        status: data.job?.status || 'unknown',
        createdAt: parseInt(data.job?.created_at || '0', 10),
      };
    },

    queryProvider: async (address: string): Promise<ProviderInfo> => {
      const response = await fetch(`${config.restEndpoint}/virtengine/provider/v1/providers/${address}`);
      const data = await response.json();
      return {
        address,
        status: data.provider?.status || 'unknown',
        reliabilityScore: parseInt(data.provider?.reliability_score || '0', 10),
        registeredAt: parseInt(data.provider?.registered_at || '0', 10),
      };
    },

    query: async <T,>(path: string, params?: Record<string, string>): Promise<T> => {
      const url = new URL(`${config.restEndpoint}${path}`);
      if (params) {
        for (const [key, value] of Object.entries(params)) {
          url.searchParams.set(key, value);
        }
      }
      const response = await fetch(url.toString());
      return response.json();
    },
  };

  useEffect(() => {
    connect();
    return () => disconnect();
  }, []);

  const actions: ChainActions = {
    connect,
    disconnect,
    subscribe,
    unsubscribe,
    broadcastTx,
  };

  return (
    <ChainContext.Provider value={{ state, queryClient, actions }}>
      {children}
    </ChainContext.Provider>
  );
}

export function useChain(): ChainContextValue {
  const context = useContext(ChainContext);
  if (!context) {
    throw new Error('useChain must be used within a ChainProvider');
  }
  return context;
}
