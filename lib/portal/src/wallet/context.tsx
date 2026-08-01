"use client";

import * as React from "react";
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
} from "./types";
import { KeplrAdapter } from "./adapters/keplr";
import { LeapAdapter } from "./adapters/leap";
import { CosmostationAdapter } from "./adapters/cosmostation";
import { WalletConnectAdapter } from "./adapters/walletconnect";
import type { WalletAdapter } from "./types";
import { WalletError as TypedWalletError, WalletErrorCode, parseWalletError } from "./errors";
import { WalletSessionManager, type WalletSession } from "./session";

const DEFAULT_PERSIST_KEY = "virtengine_wallet_session";

const initialState: WalletState = {
  status: "idle",
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

export function WalletProvider({
  children,
  config,
}: WalletProviderProps): JSX.Element {
  const [state, setState] = React.useState<WalletState>(() => ({
    ...initialState,
    autoConnect: config.autoConnect ?? true,
  }));

  const persistKey = config.persistKey ?? DEFAULT_PERSIST_KEY;
  const sessionManager = React.useMemo(
    () => new WalletSessionManager({ persistKey, autoReconnect: config.autoConnect ?? true }),
    [config.autoConnect, persistKey],
  );

  const adaptersRef = React.useRef<Map<WalletType, WalletAdapter> | null>(null);

  if (!adaptersRef.current) {
    const adapters = new Map<WalletType, WalletAdapter>();
    adapters.set("keplr", new KeplrAdapter());
    adapters.set("leap", new LeapAdapter());
    adapters.set("cosmostation", new CosmostationAdapter());

    if (config.walletConnectProjectId) {
      const metadata = config.metadata ?? {
        name: "VirtEngine Portal",
        description: "VirtEngine wallet connection",
        url: "https://portal.virtengine.com",
        icons: ["https://portal.virtengine.com/favicon.ico"],
      };

      adapters.set(
        "walletconnect",
        new WalletConnectAdapter(config.walletConnectProjectId, metadata),
      );
    }

    adaptersRef.current = adapters;
  }

  const chainInfo = config.chainInfo;
  const previousChainIdRef = React.useRef(chainInfo.chainId);

  React.useEffect(() => {
    if (previousChainIdRef.current !== chainInfo.chainId) {
      sessionManager.clearLiveAuthorization();
      sessionManager.clearSession();
      setState((prev) => ({
        ...initialState,
        autoConnect: prev.autoConnect,
      }));
    }
    previousChainIdRef.current = chainInfo.chainId;
    sessionManager.setExpectedChainId(chainInfo.chainId);
  }, [chainInfo.chainId, sessionManager]);

  React.useEffect(() => {
    const clearProviderState = () => {
      setState((prev) => ({
        ...initialState,
        autoConnect: prev.autoConnect,
      }));
    };
    const unsubscribe = sessionManager.onInvalidated(clearProviderState);
    return () => {
      unsubscribe();
      sessionManager.dispose();
    };
  }, [sessionManager]);

  const setError = React.useCallback(
    (error: WalletError | null) => {
      setState((prev) => ({
        ...prev,
        status: error ? "error" : prev.status,
        error,
      }));
      if (error && config.onError) {
        config.onError(error);
      }
    },
    [config],
  );

  const getAdapter = React.useCallback(
    (walletType: WalletType | null): WalletAdapter | null => {
      if (!walletType || !adaptersRef.current) return null;
      return adaptersRef.current.get(walletType) ?? null;
    },
    [],
  );

  const persistReconnectMetadata = React.useCallback(
    (walletType: WalletType, account: WalletAccount) => {
      sessionManager.setExpectedContext({
        walletType,
        account: account.address,
        chainId: chainInfo.chainId,
      });
      sessionManager.clearLiveAuthorization();
      sessionManager.saveSession(
        sessionManager.createSession({
          walletType,
          address: account.address,
          chainId: chainInfo.chainId,
          autoReconnect: config.autoConnect ?? true,
        }),
      );
    },
    [chainInfo.chainId, config.autoConnect, sessionManager],
  );

  const connectWallet = React.useCallback(
    async (walletType: WalletType, reconnect: WalletSession | null) => {
      const adapter = getAdapter(walletType);
      if (!adapter) {
        const error = new TypedWalletError(
          WalletErrorCode.UNKNOWN,
          "Unsupported wallet type",
        );
        setError(error);
        throw error;
      }

      if (!adapter.isAvailable()) {
        const error = new TypedWalletError(
          WalletErrorCode.WALLET_NOT_INSTALLED,
          `${adapter.name} wallet is not available`,
        );
        setError(error);
        throw error;
      }

      setState((prev) => ({
        ...prev,
        status: "connecting",
        walletType,
        error: null,
      }));

      try {
        if (reconnect?.chainId !== undefined && reconnect.chainId !== chainInfo.chainId) {
          throw new TypedWalletError(WalletErrorCode.INVALID_CHAIN_ID);
        }
        if (reconnect && (reconnect.walletType !== walletType || adapter.type !== walletType)) {
          throw new TypedWalletError(WalletErrorCode.SESSION_EXPIRED);
        }
        const accounts = await adapter.connect(chainInfo);
        const account = accounts[0];
        if (!account) {
          throw new TypedWalletError(WalletErrorCode.ACCOUNT_NOT_FOUND);
        }
        if (reconnect && account.address !== reconnect.address) {
          throw new TypedWalletError(WalletErrorCode.SESSION_EXPIRED);
        }
        const connectedAt = Date.now();
        if (!reconnect) persistReconnectMetadata(walletType, account);
        setState((prev) => {
          const nextState: WalletState = {
            ...prev,
            status: "connected",
            walletType,
            chainId: chainInfo.chainId,
            accounts,
            activeAccountIndex: 0,
            error: null,
            lastConnectedAt: connectedAt,
          };
          return nextState;
        });
      } catch (error) {
        sessionManager.clearSession();
        const walletError = parseWalletError(error);
        if (reconnect) {
          setState((prev) => ({
            ...initialState,
            status: "error",
            error: walletError,
            autoConnect: prev.autoConnect,
          }));
        }
        setError(walletError);
        throw walletError;
      }
    },
    [chainInfo, getAdapter, persistReconnectMetadata, sessionManager, setError],
  );

  const connect = React.useCallback(
    async (walletType: WalletType) => connectWallet(walletType, null),
    [connectWallet],
  );

  const disconnect = React.useCallback(async () => {
    const adapter = getAdapter(state.walletType);
    try {
      if (adapter) {
        await adapter.disconnect();
      }
    } finally {
      sessionManager.clearSession();
      setState((prev) => ({
        ...initialState,
        autoConnect: prev.autoConnect,
      }));
    }
  }, [getAdapter, sessionManager, state.walletType]);

  const refreshAccounts = React.useCallback(async () => {
    const adapter = getAdapter(state.walletType);
    if (!adapter) return;

    try {
      const accounts = await adapter.getAccounts(chainInfo);
      const activeAccountIndex = Math.min(
        state.activeAccountIndex,
        Math.max(accounts.length - 1, 0),
      );
      sessionManager.clearLiveAuthorization();
      const account = accounts[activeAccountIndex];
      if (state.walletType && account) persistReconnectMetadata(state.walletType, account);
      else sessionManager.clearSession();
      setState((prev) => ({ ...prev, accounts, activeAccountIndex }));
    } catch (error) {
      setError({
        code: "account_refresh_failed",
        message:
          error instanceof Error ? error.message : "Failed to refresh accounts",
        cause: error,
      });
    }
  }, [
    chainInfo,
    getAdapter,
    persistReconnectMetadata,
    sessionManager,
    setError,
    state.activeAccountIndex,
    state.walletType,
  ]);

  const selectAccount = React.useCallback(
    (index: number) => {
      sessionManager.clearLiveAuthorization();
      const account = state.accounts[index];
      if (state.walletType && account) persistReconnectMetadata(state.walletType, account);
      setState((prev) => ({ ...prev, activeAccountIndex: index }));
    },
    [persistReconnectMetadata, sessionManager, state.accounts, state.walletType],
  );

  const getActiveAccount = React.useCallback(
    (accounts: WalletAccount[], index: number): WalletAccount => {
      if (accounts.length === 0) {
        throw new Error("No wallet account available");
      }
      return accounts[index] ?? accounts[0];
    },
    [],
  );

  const signAmino = React.useCallback(
    async (
      signDoc: AminoSignDoc,
      options?: WalletSignOptions,
    ): Promise<AminoSignResponse> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter) {
        throw new Error("No wallet connected");
      }

      const account = getActiveAccount(
        state.accounts,
        state.activeAccountIndex,
      );
      return adapter.signAmino(
        chainInfo.chainId,
        account.address,
        signDoc,
        options,
      );
    },
    [
      chainInfo.chainId,
      getActiveAccount,
      getAdapter,
      state.accounts,
      state.activeAccountIndex,
      state.walletType,
    ],
  );

  const signDirect = React.useCallback(
    async (signDoc: DirectSignDoc): Promise<DirectSignResponse> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter) {
        throw new Error("No wallet connected");
      }

      const account = getActiveAccount(
        state.accounts,
        state.activeAccountIndex,
      );
      return adapter.signDirect(chainInfo.chainId, account.address, signDoc);
    },
    [
      chainInfo.chainId,
      getActiveAccount,
      getAdapter,
      state.accounts,
      state.activeAccountIndex,
      state.walletType,
    ],
  );

  const signArbitrary = React.useCallback(
    async (
      data: string | Uint8Array,
    ): Promise<{ signature: string; pubKey: Uint8Array }> => {
      const adapter = getAdapter(state.walletType);
      if (!adapter?.signArbitrary) {
        throw new Error("Wallet does not support arbitrary signing");
      }

      const account = getActiveAccount(
        state.accounts,
        state.activeAccountIndex,
      );
      return adapter.signArbitrary(chainInfo.chainId, account.address, data);
    },
    [
      chainInfo.chainId,
      getActiveAccount,
      getAdapter,
      state.accounts,
      state.activeAccountIndex,
      state.walletType,
    ],
  );

  const estimateFee = React.useCallback(
    (gasLimit: number, denom?: string) => {
      const feeCurrency = denom
        ? chainInfo.feeCurrencies.find(
            (currency) => currency.coinMinimalDenom === denom,
          )
        : chainInfo.feeCurrencies[0];

      if (!feeCurrency) {
        return { amount: [], gas: String(gasLimit) };
      }

      const gasPrice = feeCurrency.gasPriceStep?.average ?? 0.025;
      const feeAmount = Math.ceil(gasLimit * gasPrice);

      return {
        amount: [
          { denom: feeCurrency.coinMinimalDenom, amount: String(feeAmount) },
        ],
        gas: String(gasLimit),
      };
    },
    [chainInfo.feeCurrencies],
  );

  const refreshBalance = React.useCallback(async () => {
    try {
      if (state.accounts.length === 0) return;
      const account = getActiveAccount(
        state.accounts,
        state.activeAccountIndex,
      );
      const response = await fetch(
        `${chainInfo.restEndpoint}/cosmos/bank/v1beta1/balances/${account.address}`,
      );

      if (!response.ok) {
        throw new Error("Failed to fetch balance");
      }

      const data = await response.json();
      const denom = chainInfo.stakeCurrency.coinMinimalDenom;
      const balance = (
        data.balances as Array<{ denom: string; amount: string }>
      ).find((item) => item.denom === denom);

      const formatted = balance
        ? formatBalance(balance.amount, chainInfo.stakeCurrency.coinDecimals)
        : "0";

      setState((prev) => ({
        ...prev,
        balance: formatted,
      }));
    } catch (error) {
      setError({
        code: "balance_fetch_failed",
        message:
          error instanceof Error ? error.message : "Failed to refresh balance",
        cause: error,
      });
    }
  }, [
    chainInfo.restEndpoint,
    chainInfo.stakeCurrency.coinDecimals,
    chainInfo.stakeCurrency.coinMinimalDenom,
    getActiveAccount,
    setError,
    state.accounts,
    state.activeAccountIndex,
  ]);

  const autoConnectAttemptRef = React.useRef<string | null>(null);

  React.useEffect(() => {
    if (!state.autoConnect) return;
    const reconnect = sessionManager.loadSession();
    if (!reconnect || !sessionManager.shouldAutoReconnect()) return;
    const reconnectBinding = `${reconnect.walletType}:${reconnect.chainId}:${reconnect.address}`;
    if (autoConnectAttemptRef.current === reconnectBinding) return;
    autoConnectAttemptRef.current = reconnectBinding;
    void connectWallet(reconnect.walletType, reconnect).catch(() => undefined);
  }, [connectWallet, sessionManager, state.autoConnect]);

  React.useEffect(() => {
    if (state.status !== "connected") return;

    const handler = () => {
      refreshAccounts();
    };

    if (state.walletType === "keplr") {
      window.addEventListener("keplr_keystorechange", handler);
      return () => window.removeEventListener("keplr_keystorechange", handler);
    }

    if (state.walletType === "leap") {
      window.addEventListener("leap_keystorechange", handler);
      return () => window.removeEventListener("leap_keystorechange", handler);
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

  return (
    <WalletContext.Provider value={value}>{children}</WalletContext.Provider>
  );
}

export function useWallet(): WalletContextValue {
  const context = React.useContext(WalletContext);
  if (!context) {
    throw new Error("useWallet must be used within a WalletProvider");
  }
  return context;
}

function formatBalance(amount: string, decimals: number): string {
  if (!amount) return "0";
  const padded = amount.padStart(decimals + 1, "0");
  const integer = padded.slice(0, -decimals);
  const fraction = padded.slice(-decimals).replace(/0+$/, "");
  return fraction ? `${integer}.${fraction}` : integer;
}
