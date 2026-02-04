'use client';

import * as React from 'react';
import type {
  WalletContextValue,
  WalletProviderConfig,
  WalletState,
  WalletType,
  WalletError,
  WalletChainInfo,
  WalletAccount,
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  WalletSignOptions,
} from './types';
import { KeplrAdapter } from './adapters/keplr';
import { LeapAdapter } from './adapters/leap';
import { CosmostationAdapter } from './adapters/cosmostation';
import { WalletConnectAdapter } from './adapters/walletconnect';
import type { WalletAdapter } from './types';

const DEFAULT_PERSIST_KEY = 've_wallet_session';

const initialState: WalletState = {
  status: 'idle',
  walletType: null,
  chainId: null,
  accounts: [],
  activeAccountIndex: 0,
  balance: null,
  error: null,
  lastConnectedAt: null,
  autoConnect: true,
};

const WalletContext = React.createContext<WalletContextValue | null>(null);

export interface WalletProviderProps {
  children: React.ReactNode;
  config: WalletProviderConfig;
}

export function WalletProvider({ children, config }: WalletProviderProps): JSX.Element {
  const [state, setState] = React.useState<WalletState>(() => ({
    ...initialState,
    autoConnect: config.autoConnect ?? true,
  }));

  const adaptersRef = React.useRef<Map<WalletType, WalletAdapter> | null>(null);

  if (!adaptersRef.current) {
    const adapters = new Map<WalletType, WalletAdapter>();
    adapters.set('keplr', new KeplrAdapter());
    adapters.set('leap', new LeapAdapter());
    adapters.set('cosmostation', new CosmostationAdapter());

    if (config.walletConnectProjectId) {
      const metadata = config.metadata ?? {
        name: 'VirtEngine Portal',
        description: 'VirtEngine wallet connection',
        url: 'https://portal.virtengine.io',
        icons: ['https://portal.virtengine.io/favicon.ico'],
      };

      adapters.set(
        'walletconnect',
        new WalletConnectAdapter(
          config.walletConnectProjectId,
          metadata
        )
      );
    }

    adaptersRef.current = adapters;
  }

  const chainInfo = config.chainInfo;
  const persistKey = config.persistKey ?? DEFAULT_PERSIST_KEY;

  const setError = React.useCallback(
    (error: WalletError | null) => {
      setState((prev) => ({
        ...prev,
        status: error ? 'error' : prev.status,
        error,
      }));
      if (error && config.onError) {
        config.onError(error);
      }
    },
    [config]
  );

  const getAdapter = React.useCallback(
    (walletType: WalletType | null): WalletAdapter | null => {
      if (!walletType || !adaptersRef.current) return null;
      return adaptersRef.current.get(walletType) ?? null;
    },
    []
  );

  const persistSession = React.useCallback(
    (nextState: WalletState) => {
      if (typeof window === 'undefined') return;
      const payload = {
        walletType: nextState.walletType,
        activeAccountIndex: nextState.activeAccountIndex,
        chainId: nextState.chainId,
        autoConnect: nextState.autoConnect,
        lastConnectedAt: nextState.lastConnectedAt,
      };
      window.localStorage.setItem(persistKey, JSON.stringify(payload));
    },
    [persistKey]
  );

  const clearSession = React.useCallback(() => {
    if (typeof window === 'undefined') return;
    window.localStorage.removeItem(persistKey);
  }, [persistKey]);

  const connect = React.useCallback(
    async (walletType: WalletType) => {
      const adapter = getAdapter(walletType);
      if (!adapter) {
        setError({ code: 'wallet_unavailable', message: 'Unsupported wallet type' });
        return;
      }

      if (!adapter.isAvailable()) {
        setError({ code: 'wallet_not_installed', message: `${adapter.name} wallet is not available` });
        return;
      }

      setState((prev) => ({
        ...prev,
        status: 'connecting',
        walletType,
        error: null,
      }));

      try {
        const accounts = await adapter.connect(chainInfo);
        setState((prev) => {
          const nextState: WalletState = {
            ...prev,
            status: 'connected',
            walletType,
            chainId: chainInfo.chainId,
            accounts,
            activeAccountIndex: 0,
            error: null,
            lastConnectedAt: Date.now(),
          };
          persistSession(nextState);
          return nextState;
        });
      } catch (error) {
        setError({
          code: 'connect_failed',
          message: error instanceof Error ? error.message : 'Failed to connect wallet',
          cause: error,
        });
        setState((prev) => ({
          ...prev,
          status: 'error',
        }));
      }
    },
    [chainInfo, getAdapter, persistSession, setError]
  );

  const disconnect = React.useCallback(async () => {
    const adapter = getAdapter(state.walletType);
    if (adapter) {
      await adapter.disconnect();
    }

    setState((prev) => ({
      ...initialState,
      autoConnect: prev.autoConnect,
    }));
    clearSession();
  }, [clearSession, getAdapter, state.walletType]);

  const refreshAccounts = React.useCallback(async () => {
    const adapter = getAdapter(state.walletType);
    if (!adapter) return;

    try {
      const accounts = await adapter.getAccounts(chainInfo);
      setState((prev) => {
        const nextState = {
          ...prev,
          accounts,
          activeAccountIndex: Math.min(prev.activeAccountIndex, Math.max(accounts.length - 1, 0)),
        };
        persistSession(nextState);
        return nextState;
      });
    } catch (error) {
      setError({
        code: 'account_refresh_failed',
        message: error instanceof Error ? error.message : 'Failed to refresh accounts',
        cause: error,
      });
    }
  }, [chainInfo, getAdapter, setError, state.walletType]);

  const selectAccount = React.useCallback((index: number) => {
    setState((prev) => {
      const nextState = { ...prev, activeAccountIndex: index };
      persistSession(nextState);
      return nextState;
    });
  }, [persistSession]);

  const getActiveAccount = React.useCallback(
    (accounts: WalletAccount[], index: number): WalletAccount => {
      if (accounts.length === 0) {
        throw new Error('No wallet account available');
      }
      return accounts[index] ?? accounts[0];
    },
    []
  );

  const signAmino = React.useCallback(
    async (signDoc: AminoSignDoc, options?: WalletSignOptions): Promise<AminoSignResponse> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter) {
        throw new Error('No wallet connected');
      }

      const account = getActiveAccount(state.accounts, state.activeAccountIndex);
      return adapter.signAmino(chainInfo.chainId, account.address, signDoc, options);
    },
    [chainInfo.chainId, getActiveAccount, getAdapter, state.accounts, state.activeAccountIndex, state.walletType]
  );

  const signDirect = React.useCallback(
    async (signDoc: DirectSignDoc): Promise<DirectSignResponse> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter) {
        throw new Error('No wallet connected');
      }

      const account = getActiveAccount(state.accounts, state.activeAccountIndex);
      return adapter.signDirect(chainInfo.chainId, account.address, signDoc);
    },
    [chainInfo.chainId, getActiveAccount, getAdapter, state.accounts, state.activeAccountIndex, state.walletType]
  );

  const signArbitrary = React.useCallback(
    async (data: string | Uint8Array): Promise<{ signature: string; pubKey: Uint8Array }> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter?.signArbitrary) {
        throw new Error('Wallet does not support arbitrary signing');
      }

      const account = getActiveAccount(state.accounts, state.activeAccountIndex);
      return adapter.signArbitrary(chainInfo.chainId, account.address, data);
    },
    [chainInfo.chainId, getActiveAccount, getAdapter, state.accounts, state.activeAccountIndex, state.walletType]
  );

  const estimateFee = React.useCallback(
    (gasLimit: number, denom?: string) => {
      const feeCurrency = denom
        ? chainInfo.feeCurrencies.find((currency) => currency.coinMinimalDenom === denom)
        : chainInfo.feeCurrencies[0];

      if (!feeCurrency) {
        return { amount: [], gas: String(gasLimit) };
      }

      const gasPrice = feeCurrency.gasPriceStep?.average ?? 0.025;
      const feeAmount = Math.ceil(gasLimit * gasPrice);

      return {
        amount: [{ denom: feeCurrency.coinMinimalDenom, amount: String(feeAmount) }],
        gas: String(gasLimit),
      };
    },
    [chainInfo.feeCurrencies]
  );

  const refreshBalance = React.useCallback(async () => {
    try {
      if (state.accounts.length === 0) return;
      const account = getActiveAccount(state.accounts, state.activeAccountIndex);
      const response = await fetch(
        `${chainInfo.restEndpoint}/cosmos/bank/v1beta1/balances/${account.address}`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch balance');
      }

      const data = await response.json();
      const denom = chainInfo.stakeCurrency.coinMinimalDenom;
      const balance = (data.balances as Array<{ denom: string; amount: string }>).find(
        (item) => item.denom === denom
      );

      const formatted = balance
        ? formatBalance(balance.amount, chainInfo.stakeCurrency.coinDecimals)
        : '0';

      setState((prev) => ({
        ...prev,
        balance: formatted,
      }));
    } catch (error) {
      setError({
        code: 'balance_fetch_failed',
        message: error instanceof Error ? error.message : 'Failed to refresh balance',
        cause: error,
      });
    }
  }, [chainInfo.restEndpoint, chainInfo.stakeCurrency.coinDecimals, chainInfo.stakeCurrency.coinMinimalDenom, getActiveAccount, setError, state.accounts, state.activeAccountIndex]);

  React.useEffect(() => {
    if (typeof window === 'undefined') return;
    if (!state.autoConnect) return;

    const stored = window.localStorage.getItem(persistKey);
    if (!stored) return;

    try {
      const parsed = JSON.parse(stored) as Partial<WalletState>;
      if (parsed.walletType) {
        connect(parsed.walletType as WalletType);
      }
    } catch (error) {
      clearSession();
    }
  }, [clearSession, connect, persistKey, state.autoConnect]);

  React.useEffect(() => {
    if (state.status !== 'connected') return;

    const handler = () => {
      refreshAccounts();
    };

    if (state.walletType === 'keplr') {
      window.addEventListener('keplr_keystorechange', handler);
      return () => window.removeEventListener('keplr_keystorechange', handler);
    }

    if (state.walletType === 'leap') {
      window.addEventListener('leap_keystorechange', handler);
      return () => window.removeEventListener('leap_keystorechange', handler);
    }

    return undefined;
  }, [refreshAccounts, state.status, state.walletType]);

  const value: WalletContextValue = {
    ...state,
    connect,
    disconnect,
    refreshAccounts,
    selectAccount,
    signAmino,
    signDirect,
    signArbitrary,
    estimateFee,
    refreshBalance,
  };

  return <WalletContext.Provider value={value}>{children}</WalletContext.Provider>;
}

export function useWallet(): WalletContextValue {
  const context = React.useContext(WalletContext);
  if (!context) {
    throw new Error('useWallet must be used within a WalletProvider');
  }
  return context;
}

function formatBalance(amount: string, decimals: number): string {
  if (!amount) return '0';
  const padded = amount.padStart(decimals + 1, '0');
  const integer = padded.slice(0, -decimals);
  const fraction = padded.slice(-decimals).replace(/0+$/, '');
  return fraction ? `${integer}.${fraction}` : integer;
}
