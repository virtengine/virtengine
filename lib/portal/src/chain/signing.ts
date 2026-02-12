/**
 * Signing client helpers for wallet-connected transactions.
 */

import type { AminoSignResponse, StdSignDoc } from "@cosmjs/amino";
import type {
  AccountData,
  DirectSignResponse,
  OfflineSigner,
} from "@cosmjs/proto-signing";
import {
  SigningStargateClient,
  type SigningStargateClientOptions,
  GasPrice,
} from "@cosmjs/stargate";
import type {
  AminoSignDoc,
  DirectSignDoc,
  WalletAccount,
  WalletSignOptions,
} from "../wallet/types";
import {
  ChainClientConfig,
  ChainClientError,
  ChainTimeoutError,
} from "./types";

export interface ChainWalletSignerSource {
  getAccounts: () => Promise<WalletAccount[]>;
  signAmino: (
    chainId: string,
    signerAddress: string,
    signDoc: AminoSignDoc,
    signOptions?: WalletSignOptions,
  ) => Promise<AminoSignResponse>;
  signDirect: (
    chainId: string,
    signerAddress: string,
    signDoc: DirectSignDoc,
  ) => Promise<DirectSignResponse>;
}

export interface ChainSigningClientOptions {
  /** Override the gas price used for fee estimation. */
  gasPrice?: GasPrice | string;
  /** Optional registry/amino configuration. */
  signingOptions?: SigningStargateClientOptions;
}

const DEFAULT_CONNECT_TIMEOUT_MS = 5000;

function withTimeout<T>(
  promise: Promise<T>,
  timeoutMs: number,
  endpoint?: string,
): Promise<T> {
  if (!timeoutMs || timeoutMs <= 0) return promise;
  let timeoutHandle: ReturnType<typeof setTimeout> | null = null;
  const timeoutPromise = new Promise<T>((_, reject) => {
    timeoutHandle = setTimeout(() => {
      reject(
        new ChainTimeoutError(
          "Connection timed out after " + timeoutMs + "ms",
          endpoint,
        ),
      );
    }, timeoutMs);
  });

  return Promise.race([promise, timeoutPromise]).finally(() => {
    if (timeoutHandle) {
      clearTimeout(timeoutHandle);
    }
  }) as Promise<T>;
}

function normalizeGasPrice(gasPrice?: GasPrice | string): GasPrice | undefined {
  if (!gasPrice) return undefined;
  if (typeof gasPrice === "string") {
    return GasPrice.fromString(gasPrice);
  }
  return gasPrice;
}

function getRpcEndpoints(config: ChainClientConfig): string[] {
  const endpoints = [config.endpoints, ...(config.fallbackEndpoints ?? [])];
  return endpoints.flatMap((entry) => entry.rpcEndpoints ?? []).filter(Boolean);
}

function toAccountData(account: WalletAccount): AccountData {
  return {
    address: account.address,
    algo: account.algo as AccountData["algo"],
    pubkey: account.pubKey,
  };
}

class WalletSigner {
  private readonly chainId: string;
  private readonly source: ChainWalletSignerSource;

  constructor(chainId: string, source: ChainWalletSignerSource) {
    this.chainId = chainId;
    this.source = source;
  }

  async getAccounts(): Promise<readonly AccountData[]> {
    const accounts = await this.source.getAccounts();
    return accounts.map(toAccountData);
  }

  async signAmino(
    signerAddress: string,
    signDoc: StdSignDoc,
    signOptions?: WalletSignOptions,
  ): Promise<AminoSignResponse> {
    if (!this.source.signAmino) {
      throw new ChainClientError(
        "wallet_error",
        "Wallet does not support amino signing",
      );
    }
    return this.source.signAmino(
      this.chainId,
      signerAddress,
      signDoc as unknown as AminoSignDoc,
      signOptions,
    );
  }

  async signDirect(
    signerAddress: string,
    signDoc: DirectSignDoc,
  ): Promise<DirectSignResponse> {
    if (!this.source.signDirect) {
      throw new ChainClientError(
        "wallet_error",
        "Wallet does not support direct signing",
      );
    }
    return this.source.signDirect(this.chainId, signerAddress, signDoc);
  }
}

/**
 * Create an OfflineSigner backed by the portal wallet adapter.
 */
export function createWalletSigner(
  chainId: string,
  source: ChainWalletSignerSource,
): OfflineSigner {
  return new WalletSigner(chainId, source);
}

/**
 * Thin wrapper around SigningStargateClient that connects with fallback endpoints.
 */
export class ChainSigningClient {
  private readonly config: ChainClientConfig;
  private readonly signer: OfflineSigner;
  private readonly options?: ChainSigningClientOptions;
  private client: SigningStargateClient | null = null;

  constructor(
    config: ChainClientConfig,
    signer: OfflineSigner,
    options?: ChainSigningClientOptions,
  ) {
    this.config = config;
    this.signer = signer;
    this.options = options;
  }

  /** Connects to the first healthy RPC endpoint. */
  async getClient(): Promise<SigningStargateClient> {
    if (this.client) return this.client;

    const rpcEndpoints = getRpcEndpoints(this.config);
    if (rpcEndpoints.length === 0) {
      throw new ChainClientError(
        "connection_failed",
        "No RPC endpoints configured",
      );
    }

    let lastError: unknown = null;
    for (const endpoint of rpcEndpoints) {
      try {
        const gasPrice = normalizeGasPrice(this.options?.gasPrice);
        const client = await withTimeout(
          SigningStargateClient.connectWithSigner(endpoint, this.signer, {
            ...(this.options?.signingOptions ?? {}),
            ...(gasPrice ? { gasPrice } : {}),
          }),
          this.config.timeouts.connectMs ?? DEFAULT_CONNECT_TIMEOUT_MS,
          endpoint,
        );
        this.client = client;
        return client;
      } catch (error) {
        lastError = error;
      }
    }

    throw new ChainClientError(
      "connection_failed",
      "Unable to connect to any RPC endpoint",
      {
        cause: lastError,
      },
    );
  }

  /** Disconnects the cached signing client. */
  disconnect(): void {
    if (this.client) {
      this.client.disconnect();
    }
    this.client = null;
  }
}

/**
 * Convenience factory for creating a ChainSigningClient.
 */
export function createChainSigningClient(
  config: ChainClientConfig,
  signer: OfflineSigner,
  options?: ChainSigningClientOptions,
): ChainSigningClient {
  return new ChainSigningClient(config, signer, options);
}
