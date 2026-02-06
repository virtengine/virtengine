# Task 29G: Aggregated Multi-Provider Client

**ID:** 29G  
**Title:** feat(portal): Aggregated multi-provider client  
**Priority:** P1 (High)  
**Wave:** 3 (After 29D, 29E)  
**Estimated LOC:** ~2500  
**Dependencies:** 29D (ProviderAPIClient), 29E (Wallet-signed auth)  
**Blocking:** 29K (Metrics aggregation)  

---

## Problem Statement

Users may have deployments across **multiple providers**. The current architecture requires connecting to each provider individually. We need:

1. **Unified API** - Single interface across all providers
2. **Provider discovery** - Auto-discover provider endpoints from on-chain data
3. **Automatic failover** - Handle provider outages gracefully
4. **Aggregated views** - Combined metrics/status across providers
5. **Connection pooling** - Efficient management of multiple connections

---

## Acceptance Criteria

### AC-1: MultiProviderClient Class
- [ ] Single client managing multiple provider connections
- [ ] Automatic provider endpoint discovery from chain
- [ ] Connection pooling and reuse
- [ ] Health checking and status tracking
- [ ] Graceful handling of provider failures

### AC-2: Provider Discovery
- [ ] Query chain for provider endpoints
- [ ] Cache provider info with TTL
- [ ] Auto-refresh stale provider data
- [ ] Handle providers coming online/offline

### AC-3: Unified Deployment View
- [ ] `listAllDeployments()` - Aggregated across providers
- [ ] `getDeployment(id)` - Routes to correct provider
- [ ] Deployment-to-provider mapping
- [ ] Parallel queries to multiple providers

### AC-4: Aggregated Metrics
- [ ] `getAllMetrics()` - Combined metrics from all providers
- [ ] `getTotalUsage()` - Sum usage across deployments
- [ ] `getTotalCost()` - Sum costs across providers
- [ ] Time-synchronized data collection

### AC-5: Failover Handling
- [ ] Detect provider unavailability
- [ ] Mark providers as unhealthy
- [ ] Automatic reconnection attempts
- [ ] Graceful degradation of views

### AC-6: React Integration
- [ ] `MultiProviderProvider` context
- [ ] `useMultiProvider()` hook
- [ ] `useAggregatedDeployments()` hook
- [ ] `useAggregatedMetrics()` hook

### AC-7: Performance
- [ ] Parallel provider queries
- [ ] Request batching where possible
- [ ] Response caching with invalidation
- [ ] Minimal re-renders in React

---

## Technical Requirements

### MultiProviderClient Implementation

```typescript
// lib/portal/src/multi-provider/client.ts

import { ProviderAPIClient, ProviderAPIClientOptions } from '../provider-api/client';

export interface Provider {
  address: string;
  endpoint: string;
  name?: string;
  status: 'online' | 'offline' | 'unknown';
  lastHealthCheck?: Date;
  attributes?: Record<string, string>;
}

export interface Deployment {
  id: string;
  providerId: string;
  providerEndpoint: string;
  state: string;
  resources: ResourceMetrics;
  createdAt: Date;
}

export interface AggregatedMetrics {
  totalCPU: { used: number; limit: number };
  totalMemory: { used: number; limit: number };
  totalStorage: { used: number; limit: number };
  totalCost: { amount: string; currency: string };
  byProvider: Map<string, ResourceMetrics>;
}

export interface MultiProviderClientOptions {
  wallet: {
    signer: any;
    address: string;
    chainId: string;
  };
  chainEndpoint: string;
  healthCheckInterval?: number;
  cacheTimeout?: number;
}

export class MultiProviderClient {
  private providers: Map<string, Provider> = new Map();
  private clients: Map<string, ProviderAPIClient> = new Map();
  private deploymentProviderMap: Map<string, string> = new Map();
  private healthCheckTimer?: NodeJS.Timer;
  
  private readonly chainEndpoint: string;
  private readonly wallet: MultiProviderClientOptions['wallet'];
  private readonly healthCheckInterval: number;
  private readonly cacheTimeout: number;

  constructor(options: MultiProviderClientOptions) {
    this.chainEndpoint = options.chainEndpoint;
    this.wallet = options.wallet;
    this.healthCheckInterval = options.healthCheckInterval ?? 60_000;
    this.cacheTimeout = options.cacheTimeout ?? 300_000;
  }

  /**
   * Initialize client by discovering providers
   */
  async initialize(): Promise<void> {
    await this.discoverProviders();
    this.startHealthChecks();
  }

  /**
   * Discover providers from on-chain data
   */
  private async discoverProviders(): Promise<void> {
    // Query chain for all providers
    const response = await fetch(`${this.chainEndpoint}/virtengine/provider/v1/providers`);
    const data = await response.json();
    
    for (const p of data.providers) {
      const provider: Provider = {
        address: p.owner,
        endpoint: this.extractEndpoint(p.host_uri),
        name: p.attributes?.name,
        status: 'unknown',
        attributes: p.attributes,
      };
      
      this.providers.set(p.owner, provider);
      
      // Create client for this provider
      const client = new ProviderAPIClient({
        endpoint: provider.endpoint,
        wallet: this.wallet,
      });
      this.clients.set(p.owner, client);
    }
  }

  /**
   * Get or create client for a provider
   */
  getClient(providerAddress: string): ProviderAPIClient | null {
    return this.clients.get(providerAddress) ?? null;
  }

  /**
   * Get all online providers
   */
  getOnlineProviders(): Provider[] {
    return Array.from(this.providers.values())
      .filter(p => p.status === 'online');
  }

  /**
   * List all deployments across all providers
   */
  async listAllDeployments(): Promise<Deployment[]> {
    const onlineProviders = this.getOnlineProviders();
    
    // Query all providers in parallel
    const results = await Promise.allSettled(
      onlineProviders.map(async (provider) => {
        const client = this.clients.get(provider.address)!;
        const deployments = await client.listDeployments();
        
        return deployments.map(d => ({
          ...d,
          providerId: provider.address,
          providerEndpoint: provider.endpoint,
        }));
      })
    );
    
    // Aggregate successful results
    const allDeployments: Deployment[] = [];
    for (const result of results) {
      if (result.status === 'fulfilled') {
        allDeployments.push(...result.value);
      }
    }
    
    // Update deployment-provider mapping
    for (const d of allDeployments) {
      this.deploymentProviderMap.set(d.id, d.providerId);
    }
    
    return allDeployments;
  }

  /**
   * Get specific deployment (routes to correct provider)
   */
  async getDeployment(deploymentId: string): Promise<Deployment | null> {
    // Check cached mapping first
    let providerId = this.deploymentProviderMap.get(deploymentId);
    
    if (!providerId) {
      // Query chain for deployment to find provider
      const response = await fetch(
        `${this.chainEndpoint}/virtengine/deployment/v1/deployments/${deploymentId}`
      );
      const data = await response.json();
      providerId = data.deployment.provider;
      this.deploymentProviderMap.set(deploymentId, providerId);
    }
    
    const client = this.clients.get(providerId);
    if (!client) {
      return null;
    }
    
    const deployment = await client.getDeploymentStatus(deploymentId);
    return {
      ...deployment,
      id: deploymentId,
      providerId,
      providerEndpoint: this.providers.get(providerId)!.endpoint,
    };
  }

  /**
   * Get aggregated metrics across all deployments
   */
  async getAggregatedMetrics(): Promise<AggregatedMetrics> {
    const deployments = await this.listAllDeployments();
    
    // Query metrics in parallel
    const metricsPromises = deployments.map(async (d) => {
      const client = this.clients.get(d.providerId)!;
      try {
        return {
          deploymentId: d.id,
          providerId: d.providerId,
          metrics: await client.getDeploymentMetrics(d.id),
        };
      } catch {
        return null;
      }
    });
    
    const results = (await Promise.all(metricsPromises)).filter(Boolean);
    
    // Aggregate
    const aggregated: AggregatedMetrics = {
      totalCPU: { used: 0, limit: 0 },
      totalMemory: { used: 0, limit: 0 },
      totalStorage: { used: 0, limit: 0 },
      totalCost: { amount: '0', currency: 'uvirt' },
      byProvider: new Map(),
    };
    
    for (const result of results) {
      if (!result) continue;
      
      aggregated.totalCPU.used += result.metrics.cpu.usage;
      aggregated.totalCPU.limit += result.metrics.cpu.limit;
      aggregated.totalMemory.used += result.metrics.memory.usage;
      aggregated.totalMemory.limit += result.metrics.memory.limit;
      aggregated.totalStorage.used += result.metrics.storage.usage;
      aggregated.totalStorage.limit += result.metrics.storage.limit;
      
      // Aggregate by provider
      const existing = aggregated.byProvider.get(result.providerId);
      if (existing) {
        existing.cpu.usage += result.metrics.cpu.usage;
        existing.memory.usage += result.metrics.memory.usage;
        existing.storage.usage += result.metrics.storage.usage;
      } else {
        aggregated.byProvider.set(result.providerId, { ...result.metrics });
      }
    }
    
    return aggregated;
  }

  /**
   * Connect to deployment logs (routes to correct provider)
   */
  connectLogs(deploymentId: string): LogStream {
    const providerId = this.deploymentProviderMap.get(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }
    
    const client = this.clients.get(providerId)!;
    return client.connectLogStream(deploymentId);
  }

  /**
   * Connect to deployment shell (routes to correct provider)
   */
  connectShell(deploymentId: string): ShellConnection {
    const providerId = this.deploymentProviderMap.get(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }
    
    const client = this.clients.get(providerId)!;
    return client.connectShell(deploymentId);
  }

  /**
   * Perform action on deployment
   */
  async performAction(
    deploymentId: string,
    action: 'start' | 'stop' | 'restart',
  ): Promise<void> {
    const providerId = this.deploymentProviderMap.get(deploymentId);
    if (!providerId) {
      throw new Error(`Unknown deployment: ${deploymentId}`);
    }
    
    const client = this.clients.get(providerId)!;
    await client.performAction(deploymentId, action);
  }

  /**
   * Start periodic health checks
   */
  private startHealthChecks(): void {
    this.healthCheckTimer = setInterval(() => {
      this.checkProviderHealth();
    }, this.healthCheckInterval);
    
    // Initial check
    this.checkProviderHealth();
  }

  /**
   * Check health of all providers
   */
  private async checkProviderHealth(): Promise<void> {
    const checks = Array.from(this.providers.entries()).map(
      async ([address, provider]) => {
        const client = this.clients.get(address)!;
        try {
          await client.health();
          provider.status = 'online';
        } catch {
          provider.status = 'offline';
        }
        provider.lastHealthCheck = new Date();
      }
    );
    
    await Promise.allSettled(checks);
  }

  /**
   * Cleanup
   */
  destroy(): void {
    if (this.healthCheckTimer) {
      clearInterval(this.healthCheckTimer);
    }
  }

  private extractEndpoint(hostUri: string): string {
    // Parse provider host_uri to API endpoint
    // Format: "provider.example.com:8443" -> "https://provider.example.com:8443"
    if (!hostUri.startsWith('http')) {
      return `https://${hostUri}`;
    }
    return hostUri;
  }
}
```

### React Context and Hooks

```typescript
// lib/portal/src/multi-provider/context.tsx

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { MultiProviderClient, MultiProviderClientOptions, Provider, Deployment, AggregatedMetrics } from './client';

interface MultiProviderContextValue {
  client: MultiProviderClient | null;
  providers: Provider[];
  isInitialized: boolean;
  error: Error | null;
}

const MultiProviderContext = createContext<MultiProviderContextValue | null>(null);

interface MultiProviderProviderProps {
  options: MultiProviderClientOptions;
  children: ReactNode;
}

export function MultiProviderProvider({ options, children }: MultiProviderProviderProps) {
  const [client, setClient] = useState<MultiProviderClient | null>(null);
  const [providers, setProviders] = useState<Provider[]>([]);
  const [isInitialized, setIsInitialized] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const multiClient = new MultiProviderClient(options);
    
    multiClient.initialize()
      .then(() => {
        setClient(multiClient);
        setProviders(multiClient.getOnlineProviders());
        setIsInitialized(true);
      })
      .catch((err) => {
        setError(err);
      });

    return () => {
      multiClient.destroy();
    };
  }, [options.wallet.address, options.chainEndpoint]);

  return (
    <MultiProviderContext.Provider value={{ client, providers, isInitialized, error }}>
      {children}
    </MultiProviderContext.Provider>
  );
}

export function useMultiProvider() {
  const context = useContext(MultiProviderContext);
  if (!context) {
    throw new Error('useMultiProvider must be used within MultiProviderProvider');
  }
  return context;
}

// lib/portal/src/hooks/useAggregatedDeployments.ts
import { useQuery } from '@tanstack/react-query';
import { useMultiProvider } from '../multi-provider/context';

export function useAggregatedDeployments() {
  const { client, isInitialized } = useMultiProvider();

  return useQuery({
    queryKey: ['aggregated-deployments'],
    queryFn: () => client!.listAllDeployments(),
    enabled: isInitialized && !!client,
    staleTime: 30_000,
  });
}

// lib/portal/src/hooks/useAggregatedMetrics.ts
import { useQuery } from '@tanstack/react-query';
import { useMultiProvider } from '../multi-provider/context';

export function useAggregatedMetrics() {
  const { client, isInitialized } = useMultiProvider();

  return useQuery({
    queryKey: ['aggregated-metrics'],
    queryFn: () => client!.getAggregatedMetrics(),
    enabled: isInitialized && !!client,
    refetchInterval: 30_000,
  });
}

// lib/portal/src/hooks/useDeploymentWithProvider.ts
import { useQuery } from '@tanstack/react-query';
import { useMultiProvider } from '../multi-provider/context';

export function useDeploymentWithProvider(deploymentId: string) {
  const { client, isInitialized } = useMultiProvider();

  return useQuery({
    queryKey: ['deployment', deploymentId],
    queryFn: () => client!.getDeployment(deploymentId),
    enabled: isInitialized && !!client && !!deploymentId,
  });
}
```

### Aggregated Components

```typescript
// lib/portal/src/components/multi-provider/ProviderStatusBar.tsx

import { useMultiProvider } from '../../multi-provider/context';

export function ProviderStatusBar() {
  const { providers } = useMultiProvider();
  
  const online = providers.filter(p => p.status === 'online').length;
  const total = providers.length;

  return (
    <div className="flex items-center gap-2 text-sm">
      <span className={online === total ? 'text-green-500' : 'text-yellow-500'}>
        ●
      </span>
      <span>
        {online}/{total} providers online
      </span>
    </div>
  );
}

// lib/portal/src/components/multi-provider/AggregatedDashboard.tsx

import { useAggregatedDeployments, useAggregatedMetrics } from '../../hooks';
import { DeploymentList } from '../deployment/DeploymentList';
import { MetricsSummary } from '../metrics/MetricsSummary';
import { ProviderStatusBar } from './ProviderStatusBar';

export function AggregatedDashboard() {
  const { data: deployments, isLoading: deploymentsLoading } = useAggregatedDeployments();
  const { data: metrics, isLoading: metricsLoading } = useAggregatedMetrics();

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <ProviderStatusBar />
      </div>

      {metrics && (
        <MetricsSummary
          cpu={metrics.totalCPU}
          memory={metrics.totalMemory}
          storage={metrics.totalStorage}
          cost={metrics.totalCost}
        />
      )}

      <div>
        <h2 className="text-xl font-semibold mb-4">Deployments</h2>
        <DeploymentList
          deployments={deployments ?? []}
          isLoading={deploymentsLoading}
          showProvider
        />
      </div>
    </div>
  );
}
```

---

## Files to Create

| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/multi-provider/client.ts` | MultiProviderClient | 450 |
| `lib/portal/src/multi-provider/context.tsx` | React context | 100 |
| `lib/portal/src/multi-provider/types.ts` | Type definitions | 80 |
| `lib/portal/src/multi-provider/cache.ts` | Caching utilities | 100 |
| `lib/portal/src/multi-provider/index.ts` | Module exports | 15 |
| `lib/portal/src/hooks/useAggregatedDeployments.ts` | Deployments hook | 40 |
| `lib/portal/src/hooks/useAggregatedMetrics.ts` | Metrics hook | 40 |
| `lib/portal/src/hooks/useDeploymentWithProvider.ts` | Single deployment | 35 |
| `lib/portal/src/hooks/useProviderStatus.ts` | Provider status | 30 |
| `lib/portal/src/components/multi-provider/ProviderStatusBar.tsx` | Status bar | 50 |
| `lib/portal/src/components/multi-provider/AggregatedDashboard.tsx` | Dashboard | 150 |
| `lib/portal/src/components/multi-provider/ProviderSelector.tsx` | Provider dropdown | 80 |
| `lib/portal/src/multi-provider/__tests__/client.test.ts` | Client tests | 300 |

**Total: ~1470 lines**

---

## Validation Checklist

- [ ] Provider discovery works from chain
- [ ] Health checks run periodically
- [ ] Deployments aggregated correctly
- [ ] Metrics aggregated correctly
- [ ] WebSocket connections route to correct provider
- [ ] Failover handles offline providers
- [ ] React hooks work correctly
- [ ] Performance acceptable with many providers

---

## Vibe-Kanban Task ID

`9f6d4aa2-6a7a-4a90-a2b4-c9286b63fba6`
