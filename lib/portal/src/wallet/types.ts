/**
 * Wallet Types for Portal
 * VE-700: Wallet-based authentication
 */

export type WalletType = 'keplr' | 'leap' | 'cosmostation' | 'walletconnect';

export type WalletConnectionStatus = 'idle' | 'connecting' | 'connected' | 'error';

export interface WalletChainInfo {
  chainId: string;
  chainName: string;
  rpcEndpoint: string;
  restEndpoint: string;
  bech32Config: {
    bech32PrefixAccAddr: string;
    bech32PrefixAccPub: string;
    bech32PrefixValAddr: string;
    bech32PrefixValPub: string;
    bech32PrefixConsAddr: string;
    bech32PrefixConsPub: string;
  };
  bip44?: {
    coinType: number;
  };
  stakeCurrency: {
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
  };
  currencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
  }>;
  feeCurrencies: Array<{
    coinDenom: string;
    coinMinimalDenom: string;
    coinDecimals: number;
    coinGeckoId?: string;
    gasPriceStep?: {
      low: number;
      average: number;
      high: number;
    };
  }>;
  features?: string[];
}

export interface WalletAccount {
  address: string;
  pubKey: Uint8Array;
  algo: string;
}

export interface WalletError {
  code: string;
  message: string;
  cause?: unknown;
}

export interface WalletState {
  status: WalletConnectionStatus;
  walletType: WalletType | null;
  chainId: string | null;
  accounts: WalletAccount[];
  activeAccountIndex: number;
  balance: string | null;
  error: WalletError | null;
  lastConnectedAt: number | null;
  autoConnect: boolean;
}

export interface WalletSignOptions {
  preferNoSetFee?: boolean;
  preferNoSetMemo?: boolean;
  disableBalanceCheck?: boolean;
}

export interface AminoSignDoc {
  chain_id: string;
  account_number: string;
  sequence: string;
  fee: {
    gas: string;
    amount: Array<{ denom: string; amount: string }>;
  };
  msgs: Array<{ type: string; value: unknown }>;
  memo: string;
}

export interface AminoSignResponse {
  signed: AminoSignDoc;
  signature: {
    pub_key: { type: string; value: string };
    signature: string;
  };
}

export interface DirectSignDoc {
  bodyBytes: Uint8Array;
  authInfoBytes: Uint8Array;
  chainId: string;
  accountNumber: number;
}

export interface DirectSignResponse {
  signed: DirectSignDoc;
  signature: {
    signature: string;
  };
}

export interface WalletAdapter {
  readonly type: WalletType;
  readonly name: string;
  readonly icon?: string;

  isAvailable(): boolean;
  connect(chainInfo: WalletChainInfo): Promise<WalletAccount[]>;
  disconnect(): Promise<void>;
  getAccounts(chainInfo: WalletChainInfo): Promise<WalletAccount[]>;
  signAmino(
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc,
    signOptions?: WalletSignOptions
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
  ): Promise<{ signature: string; pubKey: Uint8Array }>;
  onAccountChange?(handler: (accounts: WalletAccount[]) => void): () => void;
  onNetworkChange?(handler: (chainId: string) => void): () => void;
}

export interface WalletProviderConfig {
  chainInfo: WalletChainInfo;
  walletConnectProjectId?: string;
  autoConnect?: boolean;
  persistKey?: string;
  onError?: (error: WalletError) => void;
  metadata?: {
    name: string;
    description: string;
    url: string;
    icons: string[];
  };
}

export interface WalletContextValue extends WalletState {
  connect: (walletType: WalletType) => Promise<void>;
  disconnect: () => Promise<void>;
  refreshAccounts: () => Promise<void>;
  selectAccount: (index: number) => void;
  signAmino: (signDoc: AminoSignDoc, options?: WalletSignOptions) => Promise<AminoSignResponse>;
  signDirect: (signDoc: DirectSignDoc) => Promise<DirectSignResponse>;
  signArbitrary: (data: string | Uint8Array) => Promise<{ signature: string; pubKey: Uint8Array }>;
  estimateFee: (gasLimit: number, denom?: string) => { amount: Array<{ denom: string; amount: string }>; gas: string };
  refreshBalance: () => Promise<void>;
}
