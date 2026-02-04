/**
 * Wallet types for portal wallet connection (extensions + WalletConnect)
 * VE-700: Wallet-based authentication
 */

export type ExtensionWalletType = 'keplr' | 'leap' | 'cosmostation' | 'walletconnect';

export interface WalletChainCurrency {
  coinDenom: string;
  coinMinimalDenom: string;
  coinDecimals: number;
  coinGeckoId?: string;
  coinImageUrl?: string;
}

export interface WalletChainFeeCurrency extends WalletChainCurrency {
  gasPriceStep?: {
    low: number;
    average: number;
    high: number;
  };
}

export interface WalletChainConfig {
  chainId: string;
  chainName: string;
  rpcEndpoint: string;
  restEndpoint: string;
  wsEndpoint?: string;
  explorerUrl?: string;
  bech32Prefix: string;
  stakeCurrency: WalletChainCurrency;
  currencies: WalletChainCurrency[];
  feeCurrencies: WalletChainFeeCurrency[];
  features?: string[];
  slip44?: number;
}

export interface WalletAccount {
  address: string;
  algo?: string;
  pubkey?: Uint8Array;
  isNanoLedger?: boolean;
  name?: string;
}

export interface AminoSignDoc {
  chain_id: string;
  account_number: string;
  sequence: string;
  fee: {
    amount: Array<{ denom: string; amount: string }>;
    gas: string;
  };
  msgs: Array<Record<string, unknown>>;
  memo: string;
}

export interface AminoSignResponse {
  signed: AminoSignDoc;
  signature: {
    pub_key: {
      type: string;
      value: string;
    };
    signature: string;
  };
}

export interface DirectSignDoc {
  bodyBytes: Uint8Array;
  authInfoBytes: Uint8Array;
  chainId: string;
  accountNumber: LongLike;
}

export interface DirectSignResponse {
  signed: DirectSignDoc;
  signature: {
    pub_key: {
      type: string;
      value: string;
    };
    signature: string;
  };
}

export type LongLike = string | number | bigint;

export interface ArbitrarySignResponse {
  signature: string;
  pubKey?: string;
}

export interface WalletConnection {
  accounts: WalletAccount[];
  activeAccount?: WalletAccount;
}

export interface WalletAdapterContext {
  chain: WalletChainConfig;
  walletConnect?: WalletConnectConfig;
}

export interface WalletAdapter {
  id: ExtensionWalletType;
  name: string;
  supportsExtension: boolean;
  supportsMobile: boolean;
  isInstalled(): boolean;
  connect(context: WalletAdapterContext): Promise<WalletConnection>;
  disconnect(): Promise<void>;
  getAccounts(chainId: string): Promise<WalletAccount[]>;
  signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc
  ): Promise<AminoSignResponse>;
  signDirect(
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc
  ): Promise<DirectSignResponse>;
  signArbitrary?(
    chainId: string,
    signerAddress: string,
    data: string | Uint8Array
  ): Promise<ArbitrarySignResponse>;
  suggestChain?(chain: WalletChainConfig): Promise<void>;
}

export interface WalletConnectMetadata {
  name: string;
  description: string;
  url: string;
  icons: string[];
}

export interface WalletConnectConfig {
  projectId: string;
  relayUrl?: string;
  metadata: WalletConnectMetadata;
}

export interface WalletState {
  isConnecting: boolean;
  isConnected: boolean;
  walletType: ExtensionWalletType | null;
  address: string | null;
  accounts: WalletAccount[];
  chainId: string;
  networkName: string;
  lastConnectedAt: number | null;
  error: WalletError | null;
}

export interface WalletError {
  code: WalletErrorCode;
  message: string;
  details?: Record<string, unknown>;
}

export type WalletErrorCode =
  | 'wallet_not_installed'
  | 'wallet_locked'
  | 'wallet_rejected'
  | 'chain_not_supported'
  | 'chain_suggest_failed'
  | 'connection_failed'
  | 'sign_failed'
  | 'account_not_found'
  | 'walletconnect_unavailable'
  | 'walletconnect_failed'
  | 'unknown';

export interface WalletActions {
  connect: (walletType: ExtensionWalletType) => Promise<void>;
  reconnect: () => Promise<void>;
  disconnect: () => Promise<void>;
  refreshAccounts: () => Promise<void>;
  switchAccount: (address: string) => void;
  signAmino: (signDoc: AminoSignDoc, signerAddress?: string) => Promise<AminoSignResponse>;
  signDirect: (signDoc: DirectSignDoc, signerAddress?: string) => Promise<DirectSignResponse>;
  signArbitrary: (data: string | Uint8Array, signerAddress?: string) => Promise<ArbitrarySignResponse>;
  estimateFee: (txBytes: Uint8Array, gasAdjustment?: number) => Promise<FeeEstimate>;
  clearError: () => void;
}

export interface WalletContextValue {
  state: WalletState;
  actions: WalletActions;
}

export interface FeeEstimate {
  gasUsed: number;
  gasWanted: number;
  feeAmount: string;
  denom: string;
}
