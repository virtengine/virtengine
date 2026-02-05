# Hybrid Decentralized Portal - Implementation Plan

## Overview

This document details **how** to implement the hybrid architecture where:
- **Blockchain** handles trust-critical operations (orders, payments, identity)
- **Provider APIs** handle operational features (logs, metrics, dashboards)
- **Portal** seamlessly integrates both, appearing as a unified Web 3.0 application

## Architecture Summary

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CUSTOMER VIEW                                  │
│  "I'm using a decentralized cloud marketplace"                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         VE PORTAL (React)                            │   │
│  │  Single unified interface - customer doesn't see the complexity      │   │
│  └──────────────────────────────┬──────────────────────────────────────┘   │
│                                 │                                           │
│              ┌──────────────────┼──────────────────┐                       │
│              │                  │                  │                       │
│              ▼                  ▼                  ▼                       │
│  ┌───────────────────┐ ┌────────────────┐ ┌────────────────────────────┐  │
│  │   Chain Layer     │ │ Provider Layer │ │   Aggregation Layer        │  │
│  │                   │ │  (Per-Provider)│ │   (Portal-side)            │  │
│  │ • Browse offerings│ │ • Get logs     │ │ • Combine metrics from     │  │
│  │ • Place orders    │ │ • View metrics │ │   multiple providers       │  │
│  │ • Check balances  │ │ • Open shell   │ │ • Unified dashboard        │  │
│  │ • Submit payments │ │ • Download     │ │ • Cross-provider search    │  │
│  │ • Verify identity │ │   artifacts    │ │                            │  │
│  └───────────────────┘ └────────────────┘ └────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Part 1: Provider API Layer

### 1.1 Enhanced Provider Daemon API

Expand `pkg/provider_daemon/portal_api.go` to expose rich operational endpoints:

```go
// pkg/provider_daemon/portal_api.go - ENHANCED

package provider_daemon

// PortalAPIServerConfig - expanded configuration
type PortalAPIServerConfig struct {
    // Existing fields...
    
    // NEW: Feature flags
    EnableMetrics       bool
    EnableFileDownload  bool
    EnableOrganizations bool
    EnableSupportTickets bool
    EnableBilling       bool
    
    // NEW: Backend integrations
    MetricsStore     MetricsStore     // InfluxDB, Prometheus, etc.
    FileStore        FileStore        // Local, S3, IPFS
    SupportBackend   SupportBackend   // Jira, Waldur, local
    BillingBackend   BillingBackend   // Chain-synced invoices
}

func (s *PortalAPIServer) RegisterRoutes(router *mux.Router) {
    // Health & Discovery
    router.HandleFunc("/health", s.handleHealth).Methods("GET")
    router.HandleFunc("/.well-known/provider-info", s.handleProviderInfo).Methods("GET")
    
    // Deployment Operations (existing + enhanced)
    router.HandleFunc("/deployments", s.handleListDeployments).Methods("GET")
    router.HandleFunc("/deployments/{id}", s.handleGetDeployment).Methods("GET")
    router.HandleFunc("/deployments/{id}/logs", s.handleLogs).Methods("GET")
    router.HandleFunc("/deployments/{id}/shell/session", s.handleShellSession).Methods("POST")
    router.HandleFunc("/deployments/{id}/shell", s.handleShell).Methods("GET") // WebSocket
    
    // NEW: Metrics & Monitoring
    router.HandleFunc("/deployments/{id}/metrics", s.handleMetrics).Methods("GET")
    router.HandleFunc("/deployments/{id}/metrics/cpu", s.handleCPUMetrics).Methods("GET")
    router.HandleFunc("/deployments/{id}/metrics/memory", s.handleMemoryMetrics).Methods("GET")
    router.HandleFunc("/deployments/{id}/metrics/network", s.handleNetworkMetrics).Methods("GET")
    router.HandleFunc("/deployments/{id}/metrics/storage", s.handleStorageMetrics).Methods("GET")
    router.HandleFunc("/deployments/{id}/events", s.handleEvents).Methods("GET")
    
    // NEW: File Operations
    router.HandleFunc("/deployments/{id}/files", s.handleListFiles).Methods("GET")
    router.HandleFunc("/deployments/{id}/files/{path:.*}", s.handleDownloadFile).Methods("GET")
    router.HandleFunc("/deployments/{id}/files/{path:.*}", s.handleUploadFile).Methods("PUT")
    
    // NEW: Organization Management (for multi-tenant)
    router.HandleFunc("/organizations", s.handleListOrganizations).Methods("GET")
    router.HandleFunc("/organizations/{id}", s.handleGetOrganization).Methods("GET")
    router.HandleFunc("/organizations/{id}/members", s.handleOrgMembers).Methods("GET", "POST", "DELETE")
    router.HandleFunc("/organizations/{id}/invitations", s.handleOrgInvitations).Methods("GET", "POST")
    
    // NEW: Support Tickets (hybrid with chain)
    router.HandleFunc("/support/tickets", s.handleListTickets).Methods("GET")
    router.HandleFunc("/support/tickets", s.handleCreateTicket).Methods("POST")
    router.HandleFunc("/support/tickets/{id}", s.handleGetTicket).Methods("GET")
    router.HandleFunc("/support/tickets/{id}/messages", s.handleTicketMessages).Methods("GET", "POST")
    router.HandleFunc("/support/tickets/{id}/attachments", s.handleTicketAttachments).Methods("POST")
    
    // NEW: Billing & Usage (synced from chain)
    router.HandleFunc("/billing/usage", s.handleUsageSummary).Methods("GET")
    router.HandleFunc("/billing/invoices", s.handleListInvoices).Methods("GET")
    router.HandleFunc("/billing/invoices/{id}", s.handleGetInvoice).Methods("GET")
    router.HandleFunc("/billing/invoices/{id}/pdf", s.handleInvoicePDF).Methods("GET")
    
    // NEW: Provider Dashboard (for provider admins)
    router.HandleFunc("/admin/dashboard", s.handleAdminDashboard).Methods("GET")
    router.HandleFunc("/admin/orders", s.handleAdminOrders).Methods("GET")
    router.HandleFunc("/admin/revenue", s.handleAdminRevenue).Methods("GET")
}
```

### 1.2 Metrics Collection & Storage

```go
// pkg/provider_daemon/metrics_collector.go

package provider_daemon

import (
    "context"
    "time"
)

// MetricsStore abstracts time-series storage
type MetricsStore interface {
    // Write metrics
    WritePoint(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error
    
    // Query metrics
    Query(ctx context.Context, query MetricsQuery) ([]MetricsSeries, error)
    
    // Get aggregations
    Aggregate(ctx context.Context, deploymentID string, metric string, period time.Duration, fn AggregateFunc) (float64, error)
}

// MetricsQuery defines a metrics query
type MetricsQuery struct {
    DeploymentID string
    Metrics      []string        // cpu, memory, network_rx, network_tx, disk_read, disk_write
    Start        time.Time
    End          time.Time
    Interval     time.Duration   // Bucketing interval
    Aggregation  AggregateFunc   // mean, max, min, sum
}

// MetricsSeries is a time series of metrics
type MetricsSeries struct {
    Metric string
    Points []MetricsPoint
}

type MetricsPoint struct {
    Timestamp time.Time
    Value     float64
}

// Implementations:
// - InfluxDBMetricsStore: Production, scalable
// - PrometheusMetricsStore: If already using Prometheus
// - InMemoryMetricsStore: Development/testing
// - WaldurMetricsStore: Proxy to Waldur metrics API
```

### 1.3 Provider Info Discovery Endpoint

```go
// pkg/provider_daemon/discovery.go

// ProviderInfo is served at /.well-known/provider-info
type ProviderInfo struct {
    // Identity
    ProviderAddress string `json:"provider_address"`
    Name            string `json:"name"`
    Description     string `json:"description"`
    Website         string `json:"website,omitempty"`
    
    // API Capabilities
    APIVersion      string   `json:"api_version"`
    Features        []string `json:"features"` // ["logs", "metrics", "shell", "files", "support"]
    
    // Connection Info
    Endpoints       ProviderEndpoints `json:"endpoints"`
    
    // Trust Info
    ChainID         string `json:"chain_id"`
    RegistrationTx  string `json:"registration_tx"`
    CertificateCID  string `json:"certificate_cid,omitempty"` // Optional: IPFS CID of provider cert
}

type ProviderEndpoints struct {
    REST      string `json:"rest"`       // Base REST API URL
    WebSocket string `json:"websocket"`  // WebSocket URL for streaming
    GRPC      string `json:"grpc,omitempty"` // Optional gRPC endpoint
}

func (s *PortalAPIServer) handleProviderInfo(w http.ResponseWriter, r *http.Request) {
    info := ProviderInfo{
        ProviderAddress: s.cfg.ProviderAddress,
        Name:            s.cfg.ProviderName,
        APIVersion:      "v1",
        Features:        s.getEnabledFeatures(),
        Endpoints: ProviderEndpoints{
            REST:      s.cfg.PublicURL,
            WebSocket: s.cfg.PublicWSURL,
        },
        ChainID: s.cfg.ChainID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(info)
}
```

## Part 2: Portal Client Libraries

### 2.1 Provider API Client

```typescript
// lib/portal/src/services/provider-api.ts

import { SigningStargateClient } from "@cosmjs/stargate";

/**
 * ProviderAPIClient handles authenticated communication with provider APIs.
 * Uses wallet signatures for authentication - no centralized auth server.
 */
export class ProviderAPIClient {
  private baseURL: string;
  private wallet: WalletClient;
  private chainId: string;
  
  constructor(baseURL: string, wallet: WalletClient, chainId: string) {
    this.baseURL = baseURL.replace(/\/$/, '');
    this.wallet = wallet;
    this.chainId = chainId;
  }
  
  /**
   * Create authenticated request headers using wallet signature
   */
  private async createAuthHeaders(path: string): Promise<Headers> {
    const timestamp = Date.now().toString();
    const nonce = crypto.randomUUID();
    const message = `${path}:${timestamp}:${nonce}`;
    
    // Sign with wallet (Keplr/Leap/etc)
    const signature = await this.wallet.signArbitrary(
      this.chainId,
      this.wallet.address,
      message
    );
    
    return new Headers({
      'X-VE-Address': this.wallet.address,
      'X-VE-Timestamp': timestamp,
      'X-VE-Nonce': nonce,
      'X-VE-Signature': signature.signature,
      'X-VE-PubKey': signature.pub_key.value,
      'Content-Type': 'application/json',
    });
  }
  
  /**
   * Make authenticated GET request
   */
  async get<T>(path: string, params?: Record<string, string>): Promise<T> {
    const url = new URL(`${this.baseURL}${path}`);
    if (params) {
      Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
    }
    
    const headers = await this.createAuthHeaders(path);
    const response = await fetch(url.toString(), { headers });
    
    if (!response.ok) {
      throw new ProviderAPIError(response.status, await response.text());
    }
    
    return response.json();
  }
  
  /**
   * Make authenticated POST request
   */
  async post<T>(path: string, body: unknown): Promise<T> {
    const headers = await this.createAuthHeaders(path);
    const response = await fetch(`${this.baseURL}${path}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(body),
    });
    
    if (!response.ok) {
      throw new ProviderAPIError(response.status, await response.text());
    }
    
    return response.json();
  }
  
  /**
   * Open authenticated WebSocket connection
   */
  async openWebSocket(path: string): Promise<AuthenticatedWebSocket> {
    // First, get a short-lived session token
    const session = await this.post<{ token: string; expires: number }>(
      `${path}/session`,
      {}
    );
    
    // Connect WebSocket with token
    const wsURL = this.baseURL.replace('https://', 'wss://').replace('http://', 'ws://');
    const ws = new WebSocket(`${wsURL}${path}?token=${session.token}`);
    
    return new AuthenticatedWebSocket(ws, session);
  }
  
  // ============ Deployment Operations ============
  
  async listDeployments(): Promise<Deployment[]> {
    return this.get('/deployments');
  }
  
  async getDeployment(id: string): Promise<DeploymentDetail> {
    return this.get(`/deployments/${id}`);
  }
  
  async getLogs(deploymentId: string, opts?: LogOptions): Promise<LogEntry[]> {
    return this.get(`/deployments/${deploymentId}/logs`, {
      tail: opts?.tail?.toString() ?? '200',
      level: opts?.level ?? '',
      search: opts?.search ?? '',
    });
  }
  
  async streamLogs(deploymentId: string, opts?: LogStreamOptions): Promise<LogStream> {
    const ws = await this.openWebSocket(`/deployments/${deploymentId}/logs`);
    return new LogStream(ws, opts);
  }
  
  async openShell(deploymentId: string): Promise<ShellSession> {
    const ws = await this.openWebSocket(`/deployments/${deploymentId}/shell`);
    return new ShellSession(ws);
  }
  
  // ============ Metrics ============
  
  async getMetrics(deploymentId: string, opts: MetricsOptions): Promise<MetricsData> {
    return this.get(`/deployments/${deploymentId}/metrics`, {
      start: opts.start.toISOString(),
      end: opts.end.toISOString(),
      interval: opts.interval ?? '1m',
      metrics: opts.metrics?.join(',') ?? 'cpu,memory,network',
    });
  }
  
  async getDashboardSummary(deploymentId: string): Promise<DashboardSummary> {
    return this.get(`/deployments/${deploymentId}/metrics/summary`);
  }
  
  // ============ Support Tickets ============
  
  async listTickets(status?: TicketStatus): Promise<SupportTicket[]> {
    return this.get('/support/tickets', status ? { status } : undefined);
  }
  
  async createTicket(ticket: CreateTicketRequest): Promise<SupportTicket> {
    return this.post('/support/tickets', ticket);
  }
  
  async getTicketMessages(ticketId: string): Promise<TicketMessage[]> {
    return this.get(`/support/tickets/${ticketId}/messages`);
  }
  
  async addTicketMessage(ticketId: string, message: string): Promise<TicketMessage> {
    return this.post(`/support/tickets/${ticketId}/messages`, { message });
  }
  
  // ============ Organizations ============
  
  async listOrganizations(): Promise<Organization[]> {
    return this.get('/organizations');
  }
  
  async getOrganizationMembers(orgId: string): Promise<OrgMember[]> {
    return this.get(`/organizations/${orgId}/members`);
  }
}
```

### 2.2 Multi-Provider Aggregation

```typescript
// lib/portal/src/services/aggregated-client.ts

/**
 * AggregatedProviderClient combines data from multiple providers
 * to create a unified customer view.
 */
export class AggregatedProviderClient {
  private chainClient: ChainQueryClient;
  private wallet: WalletClient;
  private providerClients: Map<string, ProviderAPIClient> = new Map();
  
  constructor(chainClient: ChainQueryClient, wallet: WalletClient) {
    this.chainClient = chainClient;
    this.wallet = wallet;
  }
  
  /**
   * Get or create a provider API client
   */
  private async getProviderClient(providerAddress: string): Promise<ProviderAPIClient> {
    if (this.providerClients.has(providerAddress)) {
      return this.providerClients.get(providerAddress)!;
    }
    
    // Discover provider endpoint from chain
    const provider = await this.chainClient.provider.get(providerAddress);
    if (!provider?.hostUri) {
      throw new Error(`Provider ${providerAddress} has no hostUri registered`);
    }
    
    const client = new ProviderAPIClient(
      provider.hostUri,
      this.wallet,
      this.chainClient.chainId
    );
    
    this.providerClients.set(providerAddress, client);
    return client;
  }
  
  /**
   * Get all allocations for the current user across all providers
   */
  async getAllAllocations(): Promise<AggregatedAllocation[]> {
    // Query chain for all user's allocations
    const allocations = await this.chainClient.market.listAllocations({
      owner: this.wallet.address,
    });
    
    // Enrich with provider details
    return Promise.all(
      allocations.map(async (alloc) => {
        const provider = await this.chainClient.provider.get(alloc.providerId);
        return {
          ...alloc,
          providerName: provider.info?.name ?? provider.owner,
          providerEndpoint: provider.hostUri,
        };
      })
    );
  }
  
  /**
   * Get aggregated dashboard data across all providers
   */
  async getDashboard(): Promise<AggregatedDashboard> {
    const allocations = await this.getAllAllocations();
    
    // Fetch metrics from each provider in parallel
    const metricsPromises = allocations.map(async (alloc) => {
      try {
        const client = await this.getProviderClient(alloc.providerId);
        const metrics = await client.getDashboardSummary(alloc.deploymentId);
        return { allocation: alloc, metrics, error: null };
      } catch (error) {
        return { allocation: alloc, metrics: null, error: error.message };
      }
    });
    
    const results = await Promise.all(metricsPromises);
    
    // Aggregate
    return {
      totalAllocations: allocations.length,
      activeAllocations: allocations.filter(a => a.status === 'active').length,
      totalSpend: this.sumSpend(allocations),
      aggregatedMetrics: this.aggregateMetrics(results),
      byProvider: this.groupByProvider(results),
      errors: results.filter(r => r.error).map(r => ({
        provider: r.allocation.providerId,
        error: r.error!,
      })),
    };
  }
  
  /**
   * Search logs across all providers
   */
  async searchLogsAcrossProviders(
    query: string,
    opts?: { limit?: number; providers?: string[] }
  ): Promise<AggregatedLogResult[]> {
    const allocations = await this.getAllAllocations();
    
    // Filter by providers if specified
    const targetAllocations = opts?.providers
      ? allocations.filter(a => opts.providers!.includes(a.providerId))
      : allocations;
    
    // Search each provider
    const results = await Promise.all(
      targetAllocations.map(async (alloc) => {
        try {
          const client = await this.getProviderClient(alloc.providerId);
          const logs = await client.getLogs(alloc.deploymentId, {
            search: query,
            tail: opts?.limit ?? 100,
          });
          return {
            allocation: alloc,
            logs,
            error: null,
          };
        } catch (error) {
          return { allocation: alloc, logs: [], error: error.message };
        }
      })
    );
    
    // Merge and sort by timestamp
    return results
      .flatMap(r => r.logs.map(log => ({
        ...log,
        provider: r.allocation.providerId,
        deployment: r.allocation.deploymentId,
      })))
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      .slice(0, opts?.limit ?? 100);
  }
}
```

### 2.3 React Hooks

```typescript
// lib/portal/src/hooks/useProviderAPI.ts

import { useQuery, useMutation, useInfiniteQuery } from '@tanstack/react-query';
import { useWallet } from './useWallet';
import { useChainClient } from './useChainClient';

/**
 * Hook to get a provider API client for a specific provider
 */
export function useProviderAPI(providerAddress: string) {
  const { wallet, isConnected } = useWallet();
  const chainClient = useChainClient();
  
  const { data: provider, isLoading: isLoadingProvider } = useQuery(
    ['provider', providerAddress],
    () => chainClient.provider.get(providerAddress),
    { enabled: !!providerAddress }
  );
  
  const client = useMemo(() => {
    if (!provider?.hostUri || !wallet || !isConnected) return null;
    return new ProviderAPIClient(provider.hostUri, wallet, chainClient.chainId);
  }, [provider?.hostUri, wallet, isConnected, chainClient.chainId]);
  
  return {
    client,
    isReady: !!client,
    isLoading: isLoadingProvider,
    providerInfo: provider,
    endpoint: provider?.hostUri,
  };
}

/**
 * Hook for aggregated dashboard across all providers
 */
export function useAggregatedDashboard() {
  const { wallet, isConnected } = useWallet();
  const chainClient = useChainClient();
  
  const aggregatedClient = useMemo(() => {
    if (!wallet || !isConnected) return null;
    return new AggregatedProviderClient(chainClient, wallet);
  }, [wallet, isConnected, chainClient]);
  
  return useQuery(
    ['aggregated-dashboard', wallet?.address],
    () => aggregatedClient!.getDashboard(),
    {
      enabled: !!aggregatedClient,
      refetchInterval: 30000, // Refresh every 30s
    }
  );
}

/**
 * Hook for deployment logs with real-time streaming
 */
export function useDeploymentLogs(
  providerAddress: string,
  deploymentId: string,
  opts?: { follow?: boolean; level?: string }
) {
  const { client } = useProviderAPI(providerAddress);
  const [streamedLogs, setStreamedLogs] = useState<LogEntry[]>([]);
  
  // Initial load
  const { data: initialLogs, isLoading } = useQuery(
    ['logs', providerAddress, deploymentId],
    () => client!.getLogs(deploymentId, { tail: 500, level: opts?.level }),
    { enabled: !!client }
  );
  
  // Real-time streaming
  useEffect(() => {
    if (!client || !opts?.follow) return;
    
    let stream: LogStream | null = null;
    
    (async () => {
      stream = await client.streamLogs(deploymentId, { level: opts.level });
      stream.on('log', (entry) => {
        setStreamedLogs(prev => [...prev, entry].slice(-1000)); // Keep last 1000
      });
    })();
    
    return () => {
      stream?.close();
    };
  }, [client, deploymentId, opts?.follow, opts?.level]);
  
  return {
    logs: [...(initialLogs ?? []), ...streamedLogs],
    isLoading,
    isStreaming: opts?.follow && !!client,
  };
}

/**
 * Hook for deployment shell access
 */
export function useDeploymentShell(providerAddress: string, deploymentId: string) {
  const { client } = useProviderAPI(providerAddress);
  const [session, setSession] = useState<ShellSession | null>(null);
  const [status, setStatus] = useState<'disconnected' | 'connecting' | 'connected'>('disconnected');
  
  const connect = useCallback(async () => {
    if (!client) return;
    
    setStatus('connecting');
    try {
      const shell = await client.openShell(deploymentId);
      setSession(shell);
      setStatus('connected');
      
      shell.on('close', () => {
        setSession(null);
        setStatus('disconnected');
      });
    } catch (error) {
      setStatus('disconnected');
      throw error;
    }
  }, [client, deploymentId]);
  
  const disconnect = useCallback(() => {
    session?.close();
    setSession(null);
    setStatus('disconnected');
  }, [session]);
  
  return {
    session,
    status,
    connect,
    disconnect,
    isReady: !!client,
  };
}

/**
 * Hook for real-time metrics
 */
export function useDeploymentMetrics(
  providerAddress: string,
  deploymentId: string,
  opts?: { interval?: string; metrics?: string[] }
) {
  const { client } = useProviderAPI(providerAddress);
  
  return useQuery(
    ['metrics', providerAddress, deploymentId, opts],
    () => client!.getMetrics(deploymentId, {
      start: new Date(Date.now() - 3600000), // Last hour
      end: new Date(),
      interval: opts?.interval ?? '1m',
      metrics: opts?.metrics,
    }),
    {
      enabled: !!client,
      refetchInterval: 60000, // Refresh every minute
    }
  );
}
```

## Part 3: Organization & Multi-Tenant Support

### 3.1 On-Chain Organization (via x/group)

```go
// x/org/types/msg.go (new module or extend existing)

// Organizations are implemented using Cosmos SDK's x/group module
// with VirtEngine-specific policies

package types

import (
    grouptypes "github.com/cosmos/cosmos-sdk/x/group"
)

// OrgPolicy defines spending and authorization policies
type OrgPolicy struct {
    // Maximum spend per day (in uve)
    DailySpendLimit sdk.Coins `json:"daily_spend_limit"`
    
    // Who can create orders
    OrderCreators OrgRole `json:"order_creators"`
    
    // Who can terminate allocations
    TerminationApprovers OrgRole `json:"termination_approvers"`
    
    // Who can manage members
    MemberManagers OrgRole `json:"member_managers"`
}

type OrgRole int

const (
    OrgRoleAnyMember OrgRole = iota
    OrgRoleAdmin
    OrgRoleThreshold // Requires N of M approval
)

// CreateOrgProposal creates a Cosmos group with VE-specific metadata
func CreateOrgProposal(
    admin string,
    members []grouptypes.MemberRequest,
    policy OrgPolicy,
) *grouptypes.MsgCreateGroupWithPolicy {
    // ... implementation using x/group
}
```

### 3.2 Provider-Side Organization UI

```go
// pkg/provider_daemon/organizations.go

package provider_daemon

// OrganizationStore manages organization data synced from chain
type OrganizationStore struct {
    db *sql.DB // Local database for fast queries
}

// Organization mirrors on-chain group with enriched data
type Organization struct {
    // From chain
    GroupID     uint64   `json:"group_id"`
    Admin       string   `json:"admin"`
    Members     []Member `json:"members"`
    PolicyType  string   `json:"policy_type"`
    
    // Local enrichment
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Logo        string   `json:"logo,omitempty"`
    CreatedAt   time.Time `json:"created_at"`
    
    // Computed
    TotalSpend  sdk.Coins `json:"total_spend"`
    ActiveAllocs int      `json:"active_allocations"`
}

// SyncOrganizationsFromChain syncs group data from blockchain
func (s *OrganizationStore) SyncOrganizationsFromChain(ctx context.Context, chainClient *ChainClient) error {
    // Query x/group for all groups where any member matches known customers
    // Store locally for fast queries
}

func (s *PortalAPIServer) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
    address, err := s.authenticateRequest(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
    
    // Get organizations where this address is a member
    orgs, err := s.orgStore.GetOrganizationsForMember(r.Context(), address)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(orgs)
}
```

## Part 4: Support Ticket Hybrid Architecture

### 4.1 Ticket Flow

```
Customer                    Portal                  Chain                   Provider Daemon           Provider Backend
   │                          │                       │                          │                         │
   │  Create ticket           │                       │                          │                         │
   ├─────────────────────────►│                       │                          │                         │
   │                          │  Store encrypted      │                          │                         │
   │                          │  content at provider  │                          │                         │
   │                          ├──────────────────────────────────────────────────►│                         │
   │                          │                       │                          │  Store content           │
   │                          │                       │                          ├─────────────────────────►│
   │                          │                       │                          │                         │
   │                          │  Submit MsgCreate     │                          │                         │
   │                          │  SupportTicket        │                          │                         │
   │                          ├──────────────────────►│                          │                         │
   │                          │                       │  Emit TicketCreated     │                         │
   │                          │                       │  event                   │                         │
   │                          │                       ├─────────────────────────►│                         │
   │                          │                       │                          │  Create Jira/Waldur     │
   │                          │                       │                          │  ticket (optional)      │
   │                          │                       │                          ├─────────────────────────►│
   │                          │                       │                          │                         │
   │  View ticket             │                       │                          │                         │
   ├─────────────────────────►│                       │                          │                         │
   │                          │  Get ticket from chain│                          │                         │
   │                          ├──────────────────────►│                          │                         │
   │                          │  Get content from     │                          │                         │
   │                          │  provider             │                          │                         │
   │                          ├──────────────────────────────────────────────────►│                         │
   │                          │                       │                          │                         │
   │  ◄─────────────────────────────────────────────────────────────────────────────────────────────────────│
   │  Merged view (chain metadata + provider content)                                                        │
```

### 4.2 Ticket Types

```go
// x/support/types/ticket.go

// On-chain ticket (minimal, references encrypted content)
type SupportTicket struct {
    TicketID        string         `json:"ticket_id"`
    Customer        string         `json:"customer"`
    Provider        string         `json:"provider"`
    AllocationID    string         `json:"allocation_id,omitempty"`
    
    // Reference to encrypted content (NOT stored on-chain)
    ContentRef      string         `json:"content_ref"` // e.g., "provider://tickets/{id}"
    
    // Status tracking (on-chain for auditability)
    Status          TicketStatus   `json:"status"`
    Priority        TicketPriority `json:"priority"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    ResolvedAt      *time.Time     `json:"resolved_at,omitempty"`
    
    // For SLA tracking
    ResponseDeadline time.Time     `json:"response_deadline"`
}

// Off-chain ticket content (at provider)
type TicketContent struct {
    Subject     string            `json:"subject"`
    Description string            `json:"description"`
    Category    string            `json:"category"`
    Messages    []TicketMessage   `json:"messages"`
    Attachments []TicketAttachment `json:"attachments"`
    
    // External integrations
    JiraKey     string            `json:"jira_key,omitempty"`
    WaldurUUID  string            `json:"waldur_uuid,omitempty"`
}
```

## Part 5: Implementation Roadmap

### Phase 1: Foundation (Weeks 1-4)

```
Week 1-2: Provider API Enhancement
├── [ ] Expand portal_api.go with metrics, files, org endpoints
├── [ ] Implement wallet-signature authentication
├── [ ] Add /.well-known/provider-info endpoint
└── [ ] Write provider API OpenAPI spec

Week 3-4: Portal Client Libraries
├── [ ] Create ProviderAPIClient TypeScript class
├── [ ] Create AggregatedProviderClient for multi-provider
├── [ ] Implement React hooks (useProviderAPI, useDeploymentLogs, etc.)
└── [ ] Add to existing lib/portal package
```

### Phase 2: Core Features (Weeks 5-10)

```
Week 5-6: Logs & Shell
├── [ ] Implement log streaming WebSocket
├── [ ] Create LogViewer component with filtering
├── [ ] Implement shell WebSocket
├── [ ] Create Terminal component (xterm.js integration)

Week 7-8: Metrics & Dashboards
├── [ ] Implement metrics collection in provider daemon
├── [ ] Create MetricsChart components
├── [ ] Build aggregated dashboard view
├── [ ] Add real-time updates

Week 9-10: Organizations
├── [ ] Implement x/group integration for orgs
├── [ ] Add org management to provider API
├── [ ] Create org UI components
└── [ ] Test multi-member workflows
```

### Phase 3: Advanced Features (Weeks 11-16)

```
Week 11-12: Support Tickets
├── [ ] Implement hybrid ticket architecture
├── [ ] Add Jira/Waldur sync to provider daemon
├── [ ] Create ticket UI components
└── [ ] Test end-to-end ticket flow

Week 13-14: Billing & Invoices
├── [ ] Sync invoices from chain to provider API
├── [ ] Create invoice viewer components
├── [ ] Add PDF generation
└── [ ] Test billing workflows

Week 15-16: Polish & Testing
├── [ ] Cross-browser testing
├── [ ] Mobile responsiveness
├── [ ] Performance optimization
├── [ ] Documentation
└── [ ] E2E test suite
```

## Conclusion

This hybrid architecture achieves:

1. **Rich Waldur-like UX** via provider-operated APIs
2. **True decentralization** - no central server controls the network
3. **Trust where needed** - financial/identity operations on-chain
4. **Flexibility** - providers can use Waldur or any compatible backend
5. **Practical timeline** - builds on existing code, not from scratch

The key insight is that **decentralization doesn't mean everything on blockchain** - it means **no single point of control**. By having providers run their own APIs, we achieve decentralization while delivering Web 2.0-quality user experience.
