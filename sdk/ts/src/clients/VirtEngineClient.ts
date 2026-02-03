/**
 * VirtEngineClient - Main entry point for VirtEngine SDK
 *
 * This client provides a unified interface to interact with all VirtEngine modules.
 * It wraps the generated SDK and provides high-level, user-friendly APIs.
 *
 * @example
 * ```typescript
 * import { createVirtEngineClient } from '@virtengine/chain-sdk';
 *
 * const client = await createVirtEngineClient({
 *   rpcEndpoint: 'https://rpc.virtengine.network',
 * });
 *
 * // Query VEID identity
 * const identity = await client.veid.getIdentity('virt1...');
 *
 * // Submit HPC job (when txSigner is provided)
 * const result = await client.hpc.submitJob({ ... });
 * ```
 */

import { createChainNodeSDK } from "../sdk/chain/createChainNodeSDK.ts";
import { createChainNodeWebSDK } from "../sdk/chain/createChainNodeWebSDK.ts";
import type { TxClient } from "../sdk/transport/tx/TxClient.ts";
import { MemoryCache } from "../utils/cache.ts";
import type { ChainInfo, WalletAdapter } from "../wallet/types.ts";
import { WalletManager, type WalletType } from "../wallet/WalletManager.ts";
import type { ClientOptions } from "./BaseClient.ts";
import { EncryptionClient } from "./EncryptionClient.ts";
import { EscrowClient } from "./EscrowClient.ts";
import { HPCClient } from "./HPCClient.ts";
import { MarketClient } from "./MarketClient.ts";
import { MFAClient } from "./MFAClient.ts";
import { RolesClient } from "./RolesClient.ts";
import type { ChainNodeSDK } from "./types.ts";
import { VEIDClient } from "./VEIDClient.ts";

/**
 * Configuration options for the VirtEngine client
 */
export interface VirtEngineClientOptions {
  /**
   * gRPC endpoint URL for the VirtEngine chain
   */
  rpcEndpoint: string;

  /**
   * REST API endpoint URL (optional, defaults to rpcEndpoint)
   */
  restEndpoint?: string;

  /**
   * Optional gRPC endpoint override for node SDK queries
   */
  grpcEndpoint?: string;

  /**
   * Provide an existing SDK instance (skips internal SDK creation)
   */
  sdk?: ChainNodeSDK;

  /**
   * Transaction signer to enable tx methods
   */
  txSigner?: TxClient;

  /**
   * Use gRPC-Gateway (web SDK) instead of gRPC (node SDK)
   */
  useWeb?: boolean;

  /**
   * Chain configuration
   */
  chainInfo?: ChainInfo;

  /**
   * Wallet adapter for signing transactions
   */
  wallet?: WalletAdapter;

  /**
   * Enable query caching (default: true)
   */
  enableCaching?: boolean;

  /**
   * Cache TTL in milliseconds (default: 30000)
   */
  cacheTtlMs?: number;

  /**
   * Maximum cache size (default: 1000)
   */
  maxCacheSize?: number;
}

/**
 * Default chain configuration for VirtEngine mainnet
 */
export const VIRTENGINE_MAINNET: ChainInfo = {
  chainId: "virtengine-1",
  chainName: "VirtEngine",
  rpc: "https://rpc.virtengine.network",
  rest: "https://api.virtengine.network",
  bip44: { coinType: 118 },
  bech32Config: {
    bech32PrefixAccAddr: "virt",
    bech32PrefixAccPub: "virtpub",
    bech32PrefixValAddr: "virtvaloper",
    bech32PrefixValPub: "virtvaloperpub",
    bech32PrefixConsAddr: "virtvalcons",
    bech32PrefixConsPub: "virtvalconspub",
  },
  currencies: [
    { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
    { coinDenom: "ACT", coinMinimalDenom: "uact", coinDecimals: 6 },
  ],
  feeCurrencies: [
    { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
  ],
  stakeCurrency: { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
};

/**
 * Default chain configuration for VirtEngine testnet
 */
export const VIRTENGINE_TESTNET: ChainInfo = {
  chainId: "virtengine-testnet-1",
  chainName: "VirtEngine Testnet",
  rpc: "https://rpc.testnet.virtengine.network",
  rest: "https://api.testnet.virtengine.network",
  bip44: { coinType: 118 },
  bech32Config: {
    bech32PrefixAccAddr: "virt",
    bech32PrefixAccPub: "virtpub",
    bech32PrefixValAddr: "virtvaloper",
    bech32PrefixValPub: "virtvaloperpub",
    bech32PrefixConsAddr: "virtvalcons",
    bech32PrefixConsPub: "virtvalconspub",
  },
  currencies: [
    { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
    { coinDenom: "ACT", coinMinimalDenom: "uact", coinDecimals: 6 },
  ],
  feeCurrencies: [
    { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
  ],
  stakeCurrency: { coinDenom: "VIRT", coinMinimalDenom: "uvirt", coinDecimals: 6 },
};

/**
 * Main VirtEngine client providing access to all modules
 */
export class VirtEngineClient {
  /** VEID identity verification module */
  public readonly veid: VEIDClient;

  /** MFA multi-factor authentication module */
  public readonly mfa: MFAClient;

  /** HPC high-performance computing module */
  public readonly hpc: HPCClient;

  /** Market marketplace module */
  public readonly market: MarketClient;

  /** Escrow payment module */
  public readonly escrow: EscrowClient;

  /** Encryption key management module */
  public readonly encryption: EncryptionClient;

  /** Roles access control module */
  public readonly roles: RolesClient;

  /** Wallet manager for handling wallet connections */
  public readonly wallets: WalletManager;

  /** Underlying chain SDK */
  public readonly sdk: ChainNodeSDK;

  private readonly options: VirtEngineClientOptions;
  private readonly cache: MemoryCache;

  constructor(options: VirtEngineClientOptions) {
    this.options = options;

    // Initialize cache
    this.cache = new MemoryCache({
      ttlMs: options.cacheTtlMs ?? 30000,
      maxSize: options.maxCacheSize ?? 1000,
    });

    const queryEndpoint = options.grpcEndpoint ?? options.rpcEndpoint;
    this.sdk = options.sdk ?? (options.useWeb
      ? createChainNodeWebSDK({
        query: { baseUrl: options.restEndpoint ?? options.rpcEndpoint },
        tx: options.txSigner ? { signer: options.txSigner } : undefined,
      })
      : createChainNodeSDK({
        query: { baseUrl: queryEndpoint },
        tx: options.txSigner ? { signer: options.txSigner } : undefined,
      }));

    // Client options with shared cache
    const clientOptions: ClientOptions = {
      cache: this.cache,
      enableCaching: options.enableCaching ?? true,
    };

    // SDK deps
    const deps = { sdk: this.sdk };

    // Initialize module clients
    this.veid = new VEIDClient(deps, clientOptions);
    this.mfa = new MFAClient(deps, clientOptions);
    this.hpc = new HPCClient(deps, clientOptions);
    this.market = new MarketClient(deps, clientOptions);
    this.escrow = new EscrowClient(deps, clientOptions);
    this.encryption = new EncryptionClient(deps, clientOptions);
    this.roles = new RolesClient(deps, clientOptions);

    // Initialize wallet manager
    this.wallets = new WalletManager();
  }

  /**
   * Connect a wallet to the client
   */
  async connectWallet(
    walletType: WalletType,
    chainInfo?: ChainInfo,
  ): Promise<string> {
    const chain = chainInfo ?? this.options.chainInfo ?? VIRTENGINE_MAINNET;
    const account = await this.wallets.connect(walletType, chain);
    return account.address;
  }

  /**
   * Disconnect the current wallet
   */
  async disconnectWallet(): Promise<void> {
    await this.wallets.disconnect();
  }

  /**
   * Check if a wallet is connected
   */
  isWalletConnected(): boolean {
    return this.wallets.isConnected();
  }

  /**
   * Get the connected wallet address
   */
  getAddress(): string {
    return this.wallets.getAddress();
  }

  /**
   * Clear the query cache
   */
  clearCache(): void {
    this.cache.clear();
  }

  /**
   * Get RPC endpoint
   */
  get rpcEndpoint(): string {
    return this.options.rpcEndpoint;
  }

  /**
   * Get REST endpoint
   */
  get restEndpoint(): string {
    return this.options.restEndpoint ?? this.options.rpcEndpoint;
  }

  /**
   * Get chain info
   */
  get chainInfo(): ChainInfo {
    return this.options.chainInfo ?? VIRTENGINE_MAINNET;
  }
}

/**
 * Create a new VirtEngine client instance
 *
 * @example
 * ```typescript
 * // Basic usage
 * const client = await createVirtEngineClient({
 *   rpcEndpoint: 'https://rpc.virtengine.network',
 * });
 *
 * // With wallet
 * const client = await createVirtEngineClient({
 *   rpcEndpoint: 'https://rpc.virtengine.network',
 *   chainInfo: VIRTENGINE_MAINNET,
 * });
 * await client.connectWallet('keplr');
 * ```
 */
export async function createVirtEngineClient(
  options: VirtEngineClientOptions,
): Promise<VirtEngineClient> {
  const client = new VirtEngineClient(options);

  // If a wallet adapter is provided, connect it
  if (options.wallet) {
    const chainInfo = options.chainInfo ?? VIRTENGINE_MAINNET;
    await options.wallet.connect(chainInfo);
  }

  return client;
}

/**
 * Create a VirtEngine client for mainnet
 */
export function createMainnetClient(
  options?: Partial<VirtEngineClientOptions>,
): Promise<VirtEngineClient> {
  return createVirtEngineClient({
    rpcEndpoint: VIRTENGINE_MAINNET.rpc,
    chainInfo: VIRTENGINE_MAINNET,
    ...options,
  });
}

/**
 * Create a VirtEngine client for testnet
 */
export function createTestnetClient(
  options?: Partial<VirtEngineClientOptions>,
): Promise<VirtEngineClient> {
  return createVirtEngineClient({
    rpcEndpoint: VIRTENGINE_TESTNET.rpc,
    chainInfo: VIRTENGINE_TESTNET,
    ...options,
  });
}
