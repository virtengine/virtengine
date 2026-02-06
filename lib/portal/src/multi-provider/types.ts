import type { ResourceMetrics, Deployment } from "../provider-api/types";
import type { WalletRequestSigner } from "../auth/wallet-sign";

export type ProviderStatus = "online" | "offline" | "unknown";

export interface ProviderRecord {
  address: string;
  endpoint: string;
  name?: string;
  status: ProviderStatus;
  lastHealthCheck?: Date;
  lastUpdatedAt?: Date;
  attributes?: Record<string, string>;
  metadata?: Record<string, unknown>;
  error?: string;
}

export interface DeploymentWithProvider extends Deployment {
  providerId: string;
  providerEndpoint: string;
}

export interface AggregatedMetrics {
  collectedAt: Date;
  totalCPU: { used: number; limit: number };
  totalMemory: { used: number; limit: number };
  totalStorage: { used: number; limit: number };
  totalCost: { amount: string; currency: string };
  byProvider: Map<string, ResourceMetrics>;
  byDeployment: Map<string, ResourceMetrics>;
}

export interface MultiProviderWallet {
  signer: WalletRequestSigner;
  address: string;
  chainId: string;
}

export interface MultiProviderClientOptions {
  wallet?: MultiProviderWallet;
  chainEndpoint: string;
  healthCheckIntervalMs?: number;
  providerCacheTtlMs?: number;
  deploymentCacheTtlMs?: number;
  requestTimeoutMs?: number;
  fetcher?: typeof fetch;
}
