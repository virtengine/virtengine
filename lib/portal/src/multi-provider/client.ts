import { ProviderAPIClient } from "../provider-api/client";
import type { Deployment, ResourceMetrics } from "../provider-api/types";
import type {
  AggregatedMetrics,
  DeploymentWithProvider,
  MultiProviderClientOptions,
  ProviderRecord,
  ProviderStatus,
} from "./types";

interface ChainProviderAttribute {
  key: string;
  value: string;
}

interface ChainProviderInfo {
  name?: string;
  website?: string;
}

interface ChainProvider {
  owner: string;
  host_uri?: string;
  hostURI?: string;
  attributes?: ChainProviderAttribute[];
  info?: ChainProviderInfo;
}

interface ChainProviderResponse {
  providers?: ChainProvider[];
  pagination?: {
    next_key?: string | null;
    total?: string;
  };
}

interface DeploymentCache {
  timestamp: number;
  deployments: DeploymentWithProvider[];
}

const sumUsageTotals = (
  target: { used: number; limit: number },
  metric?: { usage: number; limit: number },
) => {
  if (!metric) return;
  target.used += metric.usage ?? 0;
  target.limit += metric.limit ?? 0;
};

const sumUsageMetrics = (
  target: { usage: number; limit: number },
  metric?: { usage: number; limit: number },
) => {
  if (!metric) return;
  target.usage += metric.usage ?? 0;
  target.limit += metric.limit ?? 0;
};

const normalizeAttributes = (
  attributes?: ChainProviderAttribute[],
): Record<string, string> => {
  if (!attributes || attributes.length === 0) return {};
  return attributes.reduce<Record<string, string>>((acc, attr) => {
    acc[attr.key] = attr.value;
    return acc;
  }, {});
};

const pickAttribute = (
  attrs: Record<string, string>,
  keys: string[],
): string | undefined => {
  for (const key of keys) {
    const value = attrs[key];
    if (value) return value;
  }
  return undefined;
};

const normalizeEndpoint = (raw?: string): string | null => {
  if (!raw) return null;
  const trimmed = raw.trim();
  if (!trimmed) return null;
  if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
    return trimmed.replace(/\/$/, "");
  }
  return `https://${trimmed.replace(/\/$/, "")}`;
};

export class MultiProviderClient {
  private readonly chainEndpoint: string;
  private readonly wallet?: MultiProviderClientOptions["wallet"];
  private readonly healthCheckIntervalMs: number;
  private readonly providerCacheTtlMs: number;
  private readonly deploymentCacheTtlMs: number;
  private readonly requestTimeoutMs?: number;
  private readonly fetcher: typeof fetch;
  private providers = new Map<string, ProviderRecord>();
  private clients = new Map<string, ProviderAPIClient>();
  private deploymentProviderMap = new Map<string, string>();
  private providerCacheTimestamp = 0;
  private deploymentCache: DeploymentCache | null = null;
  private healthCheckTimer?: ReturnType<typeof setInterval>;
  private subscribers = new Set<(providers: ProviderRecord[]) => void>();

  constructor(options: MultiProviderClientOptions) {
    this.chainEndpoint = options.chainEndpoint.replace(/\/$/, "");
    this.wallet = options.wallet;
    this.healthCheckIntervalMs = options.healthCheckIntervalMs ?? 60000;
    this.providerCacheTtlMs = options.providerCacheTtlMs ?? 300000;
    this.deploymentCacheTtlMs = options.deploymentCacheTtlMs ?? 60000;
    this.requestTimeoutMs = options.requestTimeoutMs;
    this.fetcher = options.fetcher ?? fetch;
  }

  subscribe(listener: (providers: ProviderRecord[]) => void): () => void {
    this.subscribers.add(listener);
    listener(this.getProviders());
    return () => this.subscribers.delete(listener);
  }

  async initialize(): Promise<void> {
    await this.refreshProviders(true);
    this.startHealthChecks();
  }

  destroy(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer);
      this.healthCheckTimer = undefined;
    }
  }

  getProviders(): ProviderRecord[] {
    return Array.from(this.providers.values());
  }

  getOnlineProviders(): ProviderRecord[] {
    return this.getProviders().filter(
      (provider) => provider.status === "online",
    );
  }

  getProvider(address: string): ProviderRecord | undefined {
    return this.providers.get(address);
  }

  getClient(address: string): ProviderAPIClient | null {
    const provider = this.providers.get(address);
    if (!provider) return null;
    const existing = this.clients.get(address);
    if (existing) return existing;

    const client = new ProviderAPIClient({
      endpoint: provider.endpoint,
      wallet: this.wallet,
      timeoutMs: this.requestTimeoutMs,
    });
    this.clients.set(address, client);
    return client;
  }

  async refreshProviders(force = false): Promise<void> {
    const now = Date.now();
    if (!force && now - this.providerCacheTimestamp < this.providerCacheTtlMs) {
      return;
    }

    const discovered = await this.discoverProviders();
    const nextProviders = new Map<string, ProviderRecord>();

    discovered.forEach((provider) => {
      const existing = this.providers.get(provider.address);
      const status = existing?.status ?? "unknown";
      nextProviders.set(provider.address, {
        ...provider,
        status,
        lastHealthCheck: existing?.lastHealthCheck,
        error: existing?.error,
      });
    });

    this.providers = nextProviders;
    this.providerCacheTimestamp = now;
    this.notifyProviders();
  }

  async listAllDeployments(
    options: { refresh?: boolean; status?: string } = {},
  ): Promise<DeploymentWithProvider[]> {
    if (!options.refresh && this.deploymentCache) {
      const age = Date.now() - this.deploymentCache.timestamp;
      if (age < this.deploymentCacheTtlMs) {
        return this.deploymentCache.deployments;
      }
    }

    await this.refreshProviders();

    const providers = this.getProviders().filter(
      (provider) => provider.status !== "offline",
    );
    const results = await Promise.allSettled(
      providers.map(async (provider) => {
        const client = this.getClient(provider.address);
        if (!client) return [];
        const response = await client.listDeployments({
          status: options.status,
        });
        return response.deployments.map((deployment) => ({
          ...deployment,
          providerId: provider.address,
          providerEndpoint: provider.endpoint,
        }));
      }),
    );

    const deployments: DeploymentWithProvider[] = [];
    results.forEach((result) => {
      if (result.status === "fulfilled") {
        deployments.push(...result.value);
      }
    });

    deployments.forEach((deployment) => {
      this.deploymentProviderMap.set(deployment.id, deployment.providerId);
    });

    this.deploymentCache = {
      timestamp: Date.now(),
      deployments,
    };

    return deployments;
  }

  async getDeployment(
    deploymentId: string,
  ): Promise<DeploymentWithProvider | null> {
    const providerId = this.deploymentProviderMap.get(deploymentId);
    if (providerId) {
      const client = this.getClient(providerId);
      if (client) {
        const deployment = await client.getDeployment(deploymentId);
        return {
          ...deployment,
          providerId,
          providerEndpoint: this.providers.get(providerId)?.endpoint ?? "",
        };
      }
    }

    const found = await this.findDeploymentProvider(deploymentId);
    if (!found) return null;

    this.deploymentProviderMap.set(deploymentId, found.providerId);
    return found;
  }

  async getAggregatedMetrics(): Promise<AggregatedMetrics> {
    const deployments = await this.listAllDeployments();
    const collectedAt = new Date();

    const metricsResults = await Promise.allSettled(
      deployments.map(async (deployment) => {
        const client = this.getClient(deployment.providerId);
        if (!client) return null;
        const metrics = await client.getDeploymentMetrics(deployment.id);
        return { deployment, metrics };
      }),
    );

    const aggregated: AggregatedMetrics = {
      collectedAt,
      totalCPU: { used: 0, limit: 0 },
      totalMemory: { used: 0, limit: 0 },
      totalStorage: { used: 0, limit: 0 },
      totalCost: { amount: "0", currency: "uvirt" },
      byProvider: new Map(),
      byDeployment: new Map(),
    };

    metricsResults.forEach((result) => {
      if (result.status !== "fulfilled" || !result.value) return;
      const { deployment, metrics } = result.value;

      aggregated.byDeployment.set(deployment.id, metrics);

      sumUsageTotals(aggregated.totalCPU, metrics.cpu);
      sumUsageTotals(aggregated.totalMemory, metrics.memory);
      sumUsageTotals(aggregated.totalStorage, metrics.storage);

      if (
        metrics.cost &&
        metrics.cost.currency === aggregated.totalCost.currency
      ) {
        const current = Number(aggregated.totalCost.amount);
        const addition = Number(metrics.cost.amount);
        if (!Number.isNaN(addition)) {
          aggregated.totalCost.amount = (current + addition).toString();
        }
      }

      const existing = aggregated.byProvider.get(deployment.providerId);
      if (existing) {
        sumUsageMetrics(existing.cpu, metrics.cpu);
        sumUsageMetrics(existing.memory, metrics.memory);
        sumUsageMetrics(existing.storage, metrics.storage);
      } else {
        aggregated.byProvider.set(deployment.providerId, {
          ...metrics,
          cpu: { ...metrics.cpu },
          memory: { ...metrics.memory },
          storage: { ...metrics.storage },
        });
      }
    });

    return aggregated;
  }

  async getAllMetrics(): Promise<AggregatedMetrics> {
    return this.getAggregatedMetrics();
  }

  async getTotalUsage(): Promise<{
    cpu: { used: number; limit: number };
    memory: { used: number; limit: number };
    storage: { used: number; limit: number };
  }> {
    const aggregated = await this.getAggregatedMetrics();
    return {
      cpu: aggregated.totalCPU,
      memory: aggregated.totalMemory,
      storage: aggregated.totalStorage,
    };
  }

  async getTotalCost(): Promise<{ amount: string; currency: string }> {
    const aggregated = await this.getAggregatedMetrics();
    return aggregated.totalCost;
  }

  async connectLogs(deploymentId: string) {
    const providerId = await this.resolveDeploymentProvider(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }

    const client = this.getClient(providerId);
    if (!client) {
      throw new Error(`No client for provider ${providerId}`);
    }

    return client.connectLogStream(deploymentId);
  }

  async connectShell(
    deploymentId: string,
    sessionToken?: string,
    container?: string,
  ) {
    const providerId = await this.resolveDeploymentProvider(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }

    const client = this.getClient(providerId);
    if (!client) {
      throw new Error(`No client for provider ${providerId}`);
    }

    return client.connectShell(deploymentId, sessionToken, container);
  }

  async performAction(deploymentId: string, action: string): Promise<void> {
    const providerId = await this.resolveDeploymentProvider(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }

    const client = this.getClient(providerId);
    if (!client) {
      throw new Error(`No client for provider ${providerId}`);
    }

    await client.performAction(deploymentId, action);
  }

  private async resolveDeploymentProvider(
    deploymentId: string,
  ): Promise<string | null> {
    const existing = this.deploymentProviderMap.get(deploymentId);
    if (existing) return existing;

    const found = await this.findDeploymentProvider(deploymentId);
    if (!found) return null;

    this.deploymentProviderMap.set(deploymentId, found.providerId);
    return found.providerId;
  }

  private async findDeploymentProvider(
    deploymentId: string,
  ): Promise<DeploymentWithProvider | null> {
    await this.refreshProviders();

    const providers = this.getProviders().filter(
      (provider) => provider.status !== "offline",
    );
    const results = await Promise.allSettled(
      providers.map(async (provider) => {
        const client = this.getClient(provider.address);
        if (!client) return null;
        try {
          const deployment = await client.getDeployment(deploymentId);
          return {
            ...deployment,
            providerId: provider.address,
            providerEndpoint: provider.endpoint,
          };
        } catch {
          return null;
        }
      }),
    );

    for (const result of results) {
      if (result.status === "fulfilled" && result.value) {
        return result.value;
      }
    }

    return null;
  }

  private async discoverProviders(): Promise<ProviderRecord[]> {
    const response = await this.fetchChainProviders();
    return response
      .map((provider) => {
        const attrs = normalizeAttributes(provider.attributes);
        const hostUri =
          provider.host_uri ??
          provider.hostURI ??
          pickAttribute(attrs, [
            "host_uri",
            "hostURI",
            "endpoint",
            "provider_endpoint",
            "api_endpoint",
            "portal_endpoint",
            "url",
            "uri",
          ]);

        const endpoint = normalizeEndpoint(hostUri);
        if (!endpoint) return null;

        return {
          address: provider.owner,
          endpoint,
          name:
            provider.info?.name ??
            pickAttribute(attrs, ["name", "provider_name"]),
          status: "unknown" as ProviderStatus,
          attributes: attrs,
          lastUpdatedAt: new Date(),
        };
      })
      .filter(Boolean) as ProviderRecord[];
  }

  private async fetchChainProviders(): Promise<ChainProvider[]> {
    const paths = [
      "/virtengine/provider/v1beta4/providers",
      "/virtengine/provider/v1/providers",
    ];

    let lastError: Error | null = null;

    for (const path of paths) {
      try {
        const providers = await this.fetchAllPages(path);
        if (providers.length > 0) {
          return providers;
        }
      } catch (error) {
        lastError = error as Error;
      }
    }

    if (lastError) {
      throw lastError;
    }

    return [];
  }

  private async fetchAllPages(path: string): Promise<ChainProvider[]> {
    const providers: ChainProvider[] = [];
    let nextKey: string | null | undefined = null;

    do {
      const url = new URL(`${this.chainEndpoint}${path}`);
      if (nextKey) {
        url.searchParams.set("pagination.key", nextKey);
      } else {
        url.searchParams.set("pagination.limit", "200");
      }

      const response = await this.fetcher(url.toString());
      if (!response.ok) {
        throw new Error(`Provider discovery failed: ${response.statusText}`);
      }

      const payload = (await response.json()) as ChainProviderResponse;
      if (payload.providers) {
        providers.push(...payload.providers);
      }

      nextKey = payload.pagination?.next_key ?? null;
    } while (nextKey);

    return providers;
  }

  private startHealthChecks() {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer);
    }

    this.healthCheckTimer = setInterval(() => {
      void this.checkProviderHealth();
    }, this.healthCheckIntervalMs);

    void this.checkProviderHealth();
  }

  private async checkProviderHealth() {
    const checks = Array.from(this.providers.values()).map(async (provider) => {
      const client = this.getClient(provider.address);
      if (!client) return;

      try {
        await client.health();
        provider.status = "online";
        provider.error = undefined;
      } catch (error) {
        provider.status = "offline";
        provider.error =
          error instanceof Error ? error.message : "Health check failed";
      }
      provider.lastHealthCheck = new Date();
    });

    await Promise.allSettled(checks);
    this.notifyProviders();
  }

  private notifyProviders() {
    const snapshot = this.getProviders();
    this.subscribers.forEach((listener) => listener(snapshot));
  }
}
