'use client';

import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import type {
  ExtensionWalletType,
  WalletActions,
  WalletAdapter,
  WalletAdapterContext,
  WalletChainConfig,
  WalletConnectConfig,
  WalletContextValue,
  WalletError,
  WalletState,
  FeeEstimate,
  AminoSignDoc,
  DirectSignDoc,
  ArbitrarySignResponse,
  AminoSignResponse,
  DirectSignResponse,
} from './types';
import { createKeplrAdapter, createLeapAdapter, createCosmostationAdapter, createWalletConnectAdapter } from './adapters';
import { toBase64 } from './utils';

export interface WalletProviderConfig {
  chain: WalletChainConfig;
  wallets?: ExtensionWalletType[];
  autoConnect?: boolean;
  persistKey?: string;
  walletConnect?: WalletConnectConfig;
  onError?: (error: WalletError) => void;
}

export interface WalletProviderProps {
  children: ReactNode;
  config: WalletProviderConfig;
}

const WalletContext = createContext<WalletContextValue | null>(null);

const DEFAULT_STORAGE_KEY = 've_portal_wallet_session';

interface PersistedSession {
  walletType: ExtensionWalletType;
  address?: string;
}

function getPersistedSession(storageKey: string): PersistedSession | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(storageKey);
    if (!raw) return null;
    return JSON.parse(raw) as PersistedSession;
  } catch {
    return null;
  }
}

function persistSession(storageKey: string, session: PersistedSession | null) {
  if (typeof window === 'undefined') return;
  if (!session) {
    window.localStorage.removeItem(storageKey);
    return;
  }
  window.localStorage.setItem(storageKey, JSON.stringify(session));
}

function createInitialState(chain: WalletChainConfig): WalletState {
  return {
    isConnecting: false,
    isConnected: false,
    walletType: null,
    address: null,
    accounts: [],
    chainId: chain.chainId,
    networkName: chain.chainName,
    lastConnectedAt: null,
    error: null,
  };
}

function mapError(error: unknown, fallbackCode: WalletError['code']): WalletError {
  if (error instanceof Error) {
    return { code: fallbackCode, message: error.message };
  }
  return { code: fallbackCode, message: 'Wallet operation failed' };
}

export function WalletProvider({ children, config }: WalletProviderProps): JSX.Element {
  const storageKey = config.persistKey ?? DEFAULT_STORAGE_KEY;
  const [state, setState] = useState<WalletState>(() => createInitialState(config.chain));
  const adapterRef = useRef<WalletAdapter | null>(null);

  const adapters = useMemo<Record<ExtensionWalletType, WalletAdapter>>(() => ({
    keplr: createKeplrAdapter(),
    leap: createLeapAdapter(),
    cosmostation: createCosmostationAdapter(),
    walletconnect: createWalletConnectAdapter(),
  }), []);

  const walletAdapterContext = useMemo<WalletAdapterContext>(() => ({
    chain: config.chain,
    walletConnect: config.walletConnect,
  }), [config.chain, config.walletConnect]);

  const setError = useCallback((error: WalletError | null) => {
    setState(prev => ({ ...prev, error }));
    if (error) {
      config.onError?.(error);
    }
  }, [config]);

  const connect = useCallback(async (walletType: ExtensionWalletType) => {
    if (config.wallets && !config.wallets.includes(walletType)) {
      setError({ code: 'unknown', message: 'Wallet type not enabled' });
      return;
    }
    const adapter = adapters[walletType];
    if (!adapter) {
      setError({ code: 'unknown', message: 'Unsupported wallet type' });
      return;
    }

    if (!adapter.isInstalled()) {
      setError({ code: 'wallet_not_installed', message: `${adapter.name} is not installed` });
      return;
    }

    setState(prev => ({ ...prev, isConnecting: true, error: null }));

    try {
      if (adapter.suggestChain) {
        await adapter.suggestChain(config.chain);
      }

      const connection = await adapter.connect(walletAdapterContext);
      const accounts = connection.accounts ?? [];
      const activeAccount = connection.activeAccount ?? accounts[0];

      if (!activeAccount?.address) {
        throw new Error('No account returned by wallet');
      }

      adapterRef.current = adapter;

      setState(prev => ({
        ...prev,
        isConnecting: false,
        isConnected: true,
        walletType,
        address: activeAccount.address,
        accounts,
        lastConnectedAt: Date.now(),
        error: null,
      }));

      persistSession(storageKey, { walletType, address: activeAccount.address });
    } catch (error) {
      adapterRef.current = null;
      setState(prev => ({ ...prev, isConnecting: false }));
      setError(mapError(error, 'connection_failed'));
    }
  }, [adapters, config.chain, setError, storageKey, walletAdapterContext]);

  const reconnect = useCallback(async () => {
    const session = getPersistedSession(storageKey);
    if (!session?.walletType) {
      return;
    }

    await connect(session.walletType);

    if (session.address) {
      setState(prev => ({ ...prev, address: session.address ?? null }));
    }
  }, [connect, storageKey]);

  const disconnect = useCallback(async () => {
    try {
      await adapterRef.current?.disconnect();
    } finally {
      adapterRef.current = null;
      setState(createInitialState(config.chain));
      persistSession(storageKey, null);
    }
  }, [config.chain, storageKey]);

  const refreshAccounts = useCallback(async () => {
    if (!state.walletType) return;
    const adapter = adapters[state.walletType];
    if (!adapter) return;

    try {
      const accounts = await adapter.getAccounts(config.chain.chainId);
      if (accounts.length === 0) {
        setError({ code: 'account_not_found', message: 'No accounts found' });
        return;
      }

      const active = accounts.find(account => account.address === state.address) ?? accounts[0];

      setState(prev => ({
        ...prev,
        accounts,
        address: active?.address ?? prev.address,
      }));
    } catch (error) {
      setError(mapError(error, 'connection_failed'));
    }
  }, [adapters, config.chain.chainId, setError, state.address, state.walletType]);

  const switchAccount = useCallback((address: string) => {
    setState(prev => ({
      ...prev,
      address,
    }));
    if (state.walletType) {
      persistSession(storageKey, { walletType: state.walletType, address });
    }
  }, [state.walletType, storageKey]);

  const signAmino = useCallback(async (
    signDoc: AminoSignDoc,
    signerAddress?: string
  ): Promise<AminoSignResponse> => {
    if (!state.walletType || !adapterRef.current) {
      throw new Error('No wallet connected');
    }
    const signer = signerAddress ?? state.address;
    if (!signer) {
      throw new Error('No signer address available');
    }
    return adapterRef.current.signAmino(config.chain.chainId, signer, signDoc);
  }, [config.chain.chainId, state.address, state.walletType]);

  const signDirect = useCallback(async (
    signDoc: DirectSignDoc,
    signerAddress?: string
  ): Promise<DirectSignResponse> => {
    if (!state.walletType || !adapterRef.current) {
      throw new Error('No wallet connected');
    }
    const signer = signerAddress ?? state.address;
    if (!signer) {
      throw new Error('No signer address available');
    }
    return adapterRef.current.signDirect(config.chain.chainId, signer, signDoc);
  }, [config.chain.chainId, state.address, state.walletType]);

  const signArbitrary = useCallback(async (
    data: string | Uint8Array,
    signerAddress?: string
  ): Promise<ArbitrarySignResponse> => {
    if (!state.walletType || !adapterRef.current?.signArbitrary) {
      throw new Error('Wallet does not support arbitrary signing');
    }
    const signer = signerAddress ?? state.address;
    if (!signer) {
      throw new Error('No signer address available');
    }
    return adapterRef.current.signArbitrary(config.chain.chainId, signer, data);
  }, [config.chain.chainId, state.address, state.walletType]);

  const estimateFee = useCallback(async (
    txBytes: Uint8Array,
    gasAdjustment = 1.2
  ): Promise<FeeEstimate> => {
    if (!config.chain.restEndpoint) {
      throw new Error('REST endpoint not configured');
    }

    const response = await fetch(`${config.chain.restEndpoint}/cosmos/tx/v1beta1/simulate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ tx_bytes: toBase64(txBytes) }),
    });

    if (!response.ok) {
      throw new Error('Failed to simulate transaction');
    }

    const data = await response.json();
    const gasUsed = parseInt(data.gas_info?.gas_used ?? '0', 10);
    const gasWanted = Math.ceil(gasUsed * gasAdjustment);

    const feeCurrency = config.chain.feeCurrencies[0];
    const gasPrice = feeCurrency?.gasPriceStep?.average ?? 0.025;
    const feeAmount = Math.ceil(gasWanted * gasPrice).toString();

    return {
      gasUsed,
      gasWanted,
      feeAmount,
      denom: feeCurrency?.coinMinimalDenom ?? config.chain.stakeCurrency.coinMinimalDenom,
    };
  }, [config.chain]);

  const clearError = useCallback(() => setError(null), [setError]);

  useEffect(() => {
    if (!config.autoConnect) return;
    void reconnect();
  }, [config.autoConnect, reconnect]);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const handleKeystoreChange = () => {
      void refreshAccounts();
    };

    window.addEventListener('keplr_keystorechange', handleKeystoreChange);
    window.addEventListener('leap_keystorechange', handleKeystoreChange);
    window.addEventListener('cosmostation_keystorechange', handleKeystoreChange);

    return () => {
      window.removeEventListener('keplr_keystorechange', handleKeystoreChange);
      window.removeEventListener('leap_keystorechange', handleKeystoreChange);
      window.removeEventListener('cosmostation_keystorechange', handleKeystoreChange);
    };
  }, [refreshAccounts]);

  const actions: WalletActions = {
    connect,
    reconnect,
    disconnect,
    refreshAccounts,
    switchAccount,
    signAmino,
    signDirect,
    signArbitrary,
    estimateFee,
    clearError,
  };

  return (
    <WalletContext.Provider value={{ state, actions }}>
      {children}
    </WalletContext.Provider>
  );
}

export function useWallet(): WalletContextValue {
  const context = useContext(WalletContext);
  if (!context) {
    throw new Error('useWallet must be used within a WalletProvider');
  }
  return context;
}
