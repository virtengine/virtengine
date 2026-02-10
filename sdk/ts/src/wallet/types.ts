// Network configuration
export interface ChainInfo {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bip44: { coinType: number };
  bech32Config: {
    bech32PrefixAccAddr: string;
    bech32PrefixAccPub: string;
    bech32PrefixValAddr: string;
    bech32PrefixValPub: string;
    bech32PrefixConsAddr: string;
    bech32PrefixConsPub: string;
  };
  currencies: CurrencyInfo[];
  feeCurrencies: CurrencyInfo[];
  stakeCurrency: CurrencyInfo;
}

export interface CurrencyInfo {
  coinDenom: string;
  coinMinimalDenom: string;
  coinDecimals: number;
  coinGeckoId?: string;
}

// Wallet state
export type WalletState = "disconnected" | "connecting" | "connected" | "error";

export interface WalletAccount {
  address: string;
  pubKey: Uint8Array;
  algo: string;
}

// Transaction types
export interface UnsignedTx {
  messages: EncodedMessage[];
  memo?: string;
  fee?: StdFee;
  timeoutHeight?: bigint;
}

export interface EncodedMessage {
  typeUrl: string;
  value: Uint8Array;
}

export interface StdFee {
  amount: Array<{ denom: string; amount: string }>;
  gas: string;
  granter?: string;
  payer?: string;
}

export interface SignedTx {
  bodyBytes: Uint8Array;
  authInfoBytes: Uint8Array;
  signatures: Uint8Array[];
}

export interface TxResult {
  transactionHash: string;
  code: number;
  rawLog?: string;
  height: number;
  gasWanted: number;
  gasUsed: number;
  events?: TxEvent[];
}

export interface TxEvent {
  type: string;
  attributes: Array<{ key: string; value: string }>;
}

// Wallet adapter interface
export interface WalletAdapter {
  readonly name: string;
  readonly icon?: string;
  readonly state: WalletState;
  readonly account: WalletAccount | null;

  isAvailable(): boolean;
  connect(chainInfo: ChainInfo): Promise<void>;
  disconnect(): Promise<void>;
  getAddress(): string;
  signTx(tx: UnsignedTx, chainId: string): Promise<SignedTx>;
  signArbitrary?(chainId: string, data: string | Uint8Array): Promise<{ signature: Uint8Array; pubKey: Uint8Array }>;

  // Event handlers
  onAccountChange?(callback: (account: WalletAccount | null) => void): () => void;
  onNetworkChange?(callback: (chainId: string) => void): () => void;
}

// Wallet detection
export interface WalletWindow extends Window {
  keplr?: KeplrLike;
  leap?: LeapLike;
  cosmostation?: CosmostationLike;
}

export interface KeplrLike {
  enable(chainId: string): Promise<void>;
  getKey(chainId: string): Promise<{ bech32Address: string; pubKey: Uint8Array; algo: string }>;
  signDirect(
    chainId: string,
    signer: string,
    signDoc: KeplrSignDoc
  ): Promise<{ signed: KeplrSignDoc; signature: { signature: string } }>;
  signArbitrary?(
    chainId: string,
    signer: string,
    data: string | Uint8Array
  ): Promise<{ signature: string; pub_key: { value: string } }>;
  experimentalSuggestChain?(chainInfo: KeplrChainInfo): Promise<void>;
}

export interface KeplrSignDoc {
  bodyBytes: string;
  authInfoBytes: string;
  chainId: string;
  accountNumber: number;
}

export interface KeplrChainInfo {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bip44: { coinType: number };
  bech32Config: {
    bech32PrefixAccAddr: string;
    bech32PrefixAccPub: string;
    bech32PrefixValAddr: string;
    bech32PrefixValPub: string;
    bech32PrefixConsAddr: string;
    bech32PrefixConsPub: string;
  };
  currencies: CurrencyInfo[];
  feeCurrencies: CurrencyInfo[];
  stakeCurrency: CurrencyInfo;
}

// eslint-disable-next-line @typescript-eslint/no-empty-object-type
export interface LeapLike extends KeplrLike {
  // Leap has same interface as Keplr
}

export interface CosmostationLike {
  cosmos: {
    request(params: { method: string; params: CosmostationRequestParams }): Promise<CosmostationResponse>;
  };
}

export interface CosmostationRequestParams {
  chainName?: string;
  chainId?: string;
  addressPrefix?: string;
  baseDenom?: string;
  displayDenom?: string;
  restURL?: string;
  coinType?: string;
  decimals?: number;
  doc?: CosmostationSignDoc;
}

export interface CosmostationSignDoc {
  body_bytes: string;
  auth_info_bytes: string;
  chain_id: string;
  account_number: string;
}

export interface CosmostationResponse {
  address?: string;
  publicKey?: string;
  signed_doc?: {
    body_bytes: string;
    auth_info_bytes: string;
  };
  signature?: string;
}
