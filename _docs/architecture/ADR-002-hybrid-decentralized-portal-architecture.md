# ADR-002: Hybrid Decentralized Portal Architecture

## Status

**PROPOSED** - February 2026

## Executive Summary

This ADR defines a **hybrid architecture** that delivers rich Web 2.0 functionality (dashboards, logs, metrics, organization management, multi-tenant, real-time updates) while maintaining Web 3.0 decentralization principles. The key insight is:

> **Blockchain for trust, provider APIs for operations.**

## The Problem

Building a purely blockchain-based portal faces fundamental constraints:

| Feature | Blockchain Suitability | Reason |
|---------|----------------------|--------|
| Orders, payments, settlements | ✅ Excellent | Consensus-critical, immutable |
| Identity verification | ✅ Excellent | Trust anchors, attestations |
| Escrow, billing | ✅ Excellent | Financial integrity |
| Real-time logs | ❌ Poor | Too frequent, ephemeral data |
| Metrics dashboards | ❌ Poor | Time-series data, high volume |
| Shell access | ❌ Impossible | WebSocket sessions |
| File uploads | ❌ Poor | Storage costs, latency |
| Organization management | ⚠️ Mixed | Auth is Web2, permissions could be on-chain |

Waldur provides all the "❌ Poor" features via its REST API. But we don't want a centralized Waldur - we want **decentralized Waldur instances** run by each provider.

## Proposed Architecture

### Core Principle: Provider-Operated APIs with On-Chain Discovery

```
┌─────────────────────────────────────────────────────────────────────────┐
│                          VE PORTAL (Customer)                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│   ┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐  │
│   │ Wallet Connect  │     │   Chain Client  │     │ Provider Client │  │
│   │ (Keplr/Leap)    │     │   (CosmJS SDK)  │     │ (Dynamic URLs)  │  │
│   └────────┬────────┘     └────────┬────────┘     └────────┬────────┘  │
│            │                       │                       │           │
└────────────┼───────────────────────┼───────────────────────┼───────────┘
             │                       │                       │
             ▼                       ▼                       ▼
┌────────────────────────┐  ┌────────────────────┐  ┌────────────────────┐
│    VirtEngine Chain    │  │   Provider A API   │  │   Provider B API   │
│  (Source of Truth)     │  │   (Waldur/Custom)  │  │   (Waldur/Custom)  │
│                        │  │                    │  │                    │
│  • x/market (orders)   │  │  • Logs streaming  │  │  • Logs streaming  │
│  • x/provider (registry)│ │  • Metrics/graphs  │  │  • Metrics/graphs  │
│  • x/veid (identity)   │  │  • Shell access    │  │  • Shell access    │
│  • x/escrow (payments) │  │  • File download   │  │  • File download   │
│  • x/hpc (jobs)        │  │  • Organization UI │  │  • Organization UI │
│  • x/support (tickets) │  │  • Support tickets │  │  • Support tickets │
└────────────────────────┘  └────────────────────┘  └────────────────────┘
             │                       │                       │
             │              ┌────────┴────────┐              │
             │              │ Provider Daemon │              │
             └──────────────┤ (Bridge Layer)  ├──────────────┘
                            │                 │
                            │ • Order routing │
                            │ • Status sync   │
                            │ • Usage report  │
                            └─────────────────┘
```

### Provider API Discovery (On-Chain)

Providers register their API endpoint on-chain via the existing `x/provider` module:

```protobuf
// Already exists in provider.proto
message Provider {
  string owner = 1;      // Provider address
  string host_uri = 2;   // Provider API endpoint (e.g., "https://api.provider-a.com")
  repeated Attribute attributes = 3;
  Info info = 4;
}
```

The portal queries the chain for provider endpoints:

```typescript
// Portal discovers provider API from chain
const provider = await queryClient.provider.get(providerAddress);
const providerAPI = provider.hostUri; // "https://api.provider-a.com"

// Now call provider's operational APIs
const logs = await fetch(`${providerAPI}/deployments/${deploymentId}/logs`);
const metrics = await fetch(`${providerAPI}/deployments/${deploymentId}/metrics`);
```

### Data Flow Matrix

| Data Type | Source | Why |
|-----------|--------|-----|
| **Offerings** | Chain | Consensus on what's available |
| **Orders/Bids** | Chain | Financial transactions |
| **Allocations** | Chain | Escrow-backed resource grants |
| **Provider List** | Chain | Decentralized registry |
| **VEID Status** | Chain | Identity attestations |
| **Usage Reports** | Chain (via provider) | Settlement-critical |
| **Logs** | Provider API | Ephemeral, high-volume |
| **Metrics/Graphs** | Provider API | Time-series data |
| **Shell Access** | Provider API | Real-time WebSocket |
| **File Downloads** | Provider API | Large data |
| **Support Tickets** | Chain (metadata) + Provider (content) | Hybrid for auditability |

## Detailed Component Design

### 1. Provider API Gateway (pkg/provider_daemon/portal_api.go)

Already implemented! The provider daemon exposes:

```go
// Existing endpoints
router.HandleFunc("/health", s.handleHealth)
router.HandleFunc("/deployments/{id}/logs", s.handleLogs)
router.HandleFunc("/deployments/{id}/shell/session", s.handleShellSession)
router.HandleFunc("/deployments/{id}/shell", s.handleShell)

// TO ADD: Rich operational endpoints
router.HandleFunc("/deployments/{id}/metrics", s.handleMetrics)
router.HandleFunc("/deployments/{id}/files", s.handleFiles)
router.HandleFunc("/organizations", s.handleOrganizations)
router.HandleFunc("/support/tickets", s.handleSupportTickets)
router.HandleFunc("/billing/usage", s.handleUsage)
router.HandleFunc("/billing/invoices", s.handleInvoices)
```

### 2. Authentication: Wallet-Signed Requests

Customers authenticate to provider APIs using **wallet signatures**, not centralized auth:

```typescript
// Portal signs request with customer's wallet
async function signedProviderRequest(providerAPI: string, path: string, wallet: Wallet) {
  const timestamp = Date.now();
  const nonce = crypto.randomUUID();
  const message = `${path}:${timestamp}:${nonce}`;
  
  const signature = await wallet.signArbitrary(
    CHAIN_ID,
    wallet.address,
    message
  );
  
  return fetch(`${providerAPI}${path}`, {
    headers: {
      'X-VE-Address': wallet.address,
      'X-VE-Timestamp': timestamp.toString(),
      'X-VE-Nonce': nonce,
      'X-VE-Signature': signature.signature,
      'X-VE-PubKey': signature.pub_key.value,
    }
  });
}
```

Provider daemon validates:

```go
func (s *PortalAPIServer) authenticateRequest(r *http.Request) (string, error) {
    address := r.Header.Get("X-VE-Address")
    timestamp := r.Header.Get("X-VE-Timestamp")
    nonce := r.Header.Get("X-VE-Nonce")
    signature := r.Header.Get("X-VE-Signature")
    pubKey := r.Header.Get("X-VE-PubKey")
    
    // Verify timestamp is recent (prevent replay)
    ts, _ := strconv.ParseInt(timestamp, 10, 64)
    if time.Now().Unix() - ts > 300 { // 5 minute window
        return "", errors.New("request expired")
    }
    
    // Verify signature matches claimed address
    message := fmt.Sprintf("%s:%s:%s", r.URL.Path, timestamp, nonce)
    if !verifySignature(address, pubKey, message, signature) {
        return "", errors.New("invalid signature")
    }
    
    // Verify caller has access to requested resource
    if !s.hasAccess(address, mux.Vars(r)["id"]) {
        return "", errors.New("access denied")
    }
    
    return address, nil
}
```

### 3. Multi-Tenant Organization Management

Organizations can be modeled as **multi-sig accounts on-chain** with off-chain membership management:

```
┌───────────────────────────────────────────────────────────────┐
│                    Organization Model                         │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  ON-CHAIN (x/authz + x/group)           PROVIDER API          │
│  ┌──────────────────────────┐          ┌────────────────────┐ │
│  │ Group Account            │          │ Organization UI    │ │
│  │  • Admin addresses       │◀────────▶│  • Member list     │ │
│  │  • Spending limits       │          │  • Roles           │ │
│  │  • Authorization grants  │          │  • Audit logs      │ │
│  └──────────────────────────┘          │  • Invitations     │ │
│                                        └────────────────────┘ │
│                                                               │
│  Example:                                                     │
│   - Cosmos x/group: Define org with decision policy          │
│   - Cosmos x/authz: Grant spend permissions to members       │
│   - Provider API: UI for managing members, viewing activity  │
└───────────────────────────────────────────────────────────────┘
```

**Why this works:**

- **Trust**: Authorization to spend escrow funds is on-chain (cannot be forged)
- **Rich UI**: Provider APIs handle the complex UI for member management
- **Decentralized**: Each provider can offer their own org management UI

### 4. Dashboard & Analytics

Provider daemon collects and exposes metrics:

```go
// pkg/provider_daemon/metrics_api.go

type MetricsEndpoint struct {
    prometheus *prometheus.Registry
    store      *TimeSeriesStore // Local InfluxDB/TimescaleDB
}

// Endpoints for portal dashboards
router.HandleFunc("/metrics/cpu", s.handleCPUMetrics)
router.HandleFunc("/metrics/memory", s.handleMemoryMetrics)
router.HandleFunc("/metrics/network", s.handleNetworkMetrics)
router.HandleFunc("/metrics/storage", s.handleStorageMetrics)
router.HandleFunc("/metrics/cost", s.handleCostMetrics)
router.HandleFunc("/dashboard/summary", s.handleDashboardSummary)
```

The portal queries each provider for their customers' metrics and aggregates:

```typescript
// Portal aggregates metrics from multiple providers
async function getCustomerDashboard(wallet: string, allocations: Allocation[]) {
  const metrics = await Promise.all(
    allocations.map(async (alloc) => {
      const provider = await getProvider(alloc.providerId);
      return signedProviderRequest(
        provider.hostUri, 
        `/deployments/${alloc.deploymentId}/metrics`,
        wallet
      );
    })
  );
  
  return aggregateMetrics(metrics);
}
```

### 5. Support Ticket Architecture (Hybrid)

Leverage the existing `x/support` module with provider-side enrichment:

```
┌───────────────────────────────────────────────────────────────┐
│                    Support Ticket Flow                        │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│  CUSTOMER PORTAL                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 1. Create ticket in portal                             │  │
│  │    - Select allocation/provider                        │  │
│  │    - Enter summary + description                       │  │
│  │    - Attach files (encrypted, stored at provider)      │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 2. On-chain: MsgCreateSupportTicket                    │  │
│  │    - ticket_id, customer, provider, allocation_ref     │  │
│  │    - encrypted_content_ref (IPFS/provider CID)         │  │
│  │    - status: OPEN                                      │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 3. Provider daemon receives event                      │  │
│  │    - Fetches encrypted content from CID                │  │
│  │    - Decrypts with provider key                        │  │
│  │    - Creates Jira/Waldur ticket (optional)             │  │
│  │    - Stores in local service desk for UI               │  │
│  └────────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ 4. Provider responds via provider API                  │  │
│  │    - Rich conversation UI at provider portal           │  │
│  │    - Status updates synced to chain                    │  │
│  └────────────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────────────┘
```

## Provider Backend Flexibility

### Waldur (Current)

Providers can run full Waldur for rich backend:

```yaml
# Provider A: Full Waldur deployment
provider_daemon:
  waldur:
    enabled: true
    base_url: "https://waldur.provider-a.internal"
    features:
      - marketplace
      - openstack
      - kubernetes
      - slurm
      - support_tickets
```

### Alternative: Lightweight Provider

Providers without Waldur can run minimal backend:

```yaml
# Provider B: Kubernetes-only, no Waldur
provider_daemon:
  waldur:
    enabled: false
  backends:
    - type: kubernetes
      kubeconfig: /etc/kubernetes/admin.conf
    - type: slurm
      controller: slurm-controller:6817
```

### Future: Pluggable Provider Backends

Define an interface for provider backends:

```go
// pkg/provider_daemon/backend.go

type ProviderBackend interface {
    // Resource lifecycle
    Provision(ctx context.Context, order Order) (Allocation, error)
    Terminate(ctx context.Context, allocationID string) error
    
    // Operational APIs (for portal)
    GetLogs(ctx context.Context, deploymentID string, opts LogOptions) ([]LogEntry, error)
    GetMetrics(ctx context.Context, deploymentID string, opts MetricsOptions) (Metrics, error)
    StreamLogs(ctx context.Context, deploymentID string) (<-chan LogEntry, error)
    OpenShell(ctx context.Context, deploymentID string) (ShellSession, error)
    
    // Usage reporting
    GetUsage(ctx context.Context, allocationID string, period Period) (Usage, error)
}

// Implementations:
// - WaldurBackend: Full Waldur integration
// - KubernetesBackend: Direct K8s, minimal features
// - SlurmBackend: Direct SLURM integration
// - CompositeBackend: Combines multiple backends
```

## Portal Architecture

### Unified Customer Portal

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        VE PORTAL ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                     Next.js / Vite App                          │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────────┐ │   │
│  │  │ lib/portal   │ │ lib/capture  │ │   New Portal Pages       │ │   │
│  │  │ (existing)   │ │ (existing)   │ │   (to build)             │ │   │
│  │  │ • Components │ │ • Selfie     │ │   • /marketplace         │ │   │
│  │  │ • Hooks      │ │ • Document   │ │   • /orders              │ │   │
│  │  │ • Types      │ │ • Liveness   │ │   • /deployments         │ │   │
│  │  └──────────────┘ └──────────────┘ │   • /dashboard           │ │   │
│  │                                    │   • /support             │ │   │
│  │                                    │   • /settings            │ │   │
│  │                                    └──────────────────────────┘ │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                     │                                   │
│                    ┌────────────────┼────────────────┐                 │
│                    │                │                │                 │
│                    ▼                ▼                ▼                 │
│  ┌─────────────────────┐ ┌──────────────────┐ ┌────────────────────┐  │
│  │    Chain Client     │ │  Provider Client │ │  Wallet Client     │  │
│  │   (lib/chain-sdk)   │ │ (dynamic routing)│ │   (@cosmos-kit)    │  │
│  │                     │ │                  │ │                    │  │
│  │ • Query offerings   │ │ • Get logs       │ │ • Connect wallet   │  │
│  │ • Query providers   │ │ • Get metrics    │ │ • Sign transactions│  │
│  │ • Submit orders     │ │ • Open shell     │ │ • Sign API requests│  │
│  │ • Check escrow      │ │ • Manage support │ │                    │  │
│  └─────────────────────┘ └──────────────────┘ └────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Portal Hooks

```typescript
// lib/portal/hooks/useProviderAPI.ts

export function useProviderAPI(providerAddress: string) {
  const { wallet } = useWallet();
  const chainClient = useChainClient();
  
  // Discover provider endpoint from chain
  const { data: provider } = useQuery(
    ['provider', providerAddress],
    () => chainClient.provider.get(providerAddress)
  );
  
  // Create authenticated API client
  const api = useMemo(() => {
    if (!provider?.hostUri || !wallet) return null;
    return new ProviderAPIClient(provider.hostUri, wallet);
  }, [provider?.hostUri, wallet]);
  
  return {
    api,
    isReady: !!api,
    endpoint: provider?.hostUri,
  };
}

// lib/portal/hooks/useDeploymentLogs.ts

export function useDeploymentLogs(providerId: string, deploymentId: string) {
  const { api } = useProviderAPI(providerId);
  
  return useInfiniteQuery(
    ['logs', providerId, deploymentId],
    ({ pageParam }) => api?.getLogs(deploymentId, { cursor: pageParam }),
    { enabled: !!api }
  );
}

// lib/portal/hooks/useDeploymentShell.ts

export function useDeploymentShell(providerId: string, deploymentId: string) {
  const { api } = useProviderAPI(providerId);
  const [session, setSession] = useState<ShellSession | null>(null);
  
  const connect = useCallback(async () => {
    if (!api) return;
    const ws = await api.openShell(deploymentId);
    setSession(ws);
  }, [api, deploymentId]);
  
  return { session, connect, disconnect: () => session?.close() };
}
```

## Comparison: Pure Web3 vs Hybrid

| Aspect | Pure Web3 | Hybrid (Proposed) | Winner |
|--------|-----------|-------------------|--------|
| **Logs/Metrics** | IPFS pinning (slow, expensive) | Provider API (fast, free) | Hybrid |
| **Shell Access** | Libp2p streams (complex) | WebSocket (proven) | Hybrid |
| **Dashboard UX** | Subgraph queries | Direct API | Hybrid |
| **Decentralization** | Maximum | High (providers are decentralized) | Tie |
| **Trust Model** | Blockchain consensus | Wallet-signed requests + chain anchors | Hybrid |
| **Development Cost** | Very high | Moderate | Hybrid |
| **User Experience** | Slow, limited | Web2-like speed | Hybrid |

## Security Model

### Trust Boundaries

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        TRUST MODEL                                      │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │ TRUSTLESS (Chain Consensus)                                        │ │
│  │  • Orders, payments, settlements                                  │ │
│  │  • Identity verification status                                   │ │
│  │  • Escrow balances                                                │ │
│  │  • Provider registration                                          │ │
│  │  • Usage reports (provider-signed, chain-verified)                │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │ TRUSTED PROVIDER (Customer accepts provider's assertions)          │ │
│  │  • Logs accuracy (provider generates)                              │ │
│  │  • Metrics accuracy (provider collects)                            │ │
│  │  • Shell access (provider grants)                                  │ │
│  │  • File downloads (provider serves)                                │ │
│  │                                                                     │ │
│  │  MITIGATION: Review scores, dispute resolution, provider staking   │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │ CUSTOMER-CONTROLLED (Wallet signatures)                            │ │
│  │  • API authentication (signed requests)                           │ │
│  │  • Transaction authorization                                       │ │
│  │  • Consent grants                                                  │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Dispute Resolution

If a provider's logs/metrics don't match reality:

1. **x/review Module**: Customers can leave reviews
2. **x/fraud Module**: Report fraudulent providers
3. **Provider Staking**: Providers stake tokens, slashed for fraud
4. **On-chain Evidence**: Usage reports are signed and verifiable

## Implementation Phases

### Phase 1: Foundation (4-6 weeks)

```
Tasks:
- [ ] 28D: Portal scaffold verification (Next.js)
- [ ] 28E: TypeScript chain SDK
- [ ] Provider API client library
- [ ] Wallet-signed request authentication
```

### Phase 2: Core Features (6-8 weeks)

```
Tasks:
- [ ] 28K: Marketplace browser (chain data)
- [ ] 28L: Order creation wizard (chain tx)
- [ ] Deployment logs viewer (provider API)
- [ ] Deployment shell access (provider API)
- [ ] Basic metrics dashboard (provider API)
```

### Phase 3: Rich Features (4-6 weeks)

```
Tasks:
- [ ] 28M: Customer dashboard (aggregated)
- [ ] 28N: Provider dashboard
- [ ] Support ticket UI (hybrid)
- [ ] Organization management
- [ ] Billing/invoice viewer
```

### Phase 4: Advanced (Ongoing)

```
Tasks:
- [ ] 28O: HPC job submission
- [ ] Multi-provider aggregation
- [ ] Advanced analytics
- [ ] Mobile app (React Native)
```

## Alternative Approaches Considered

### 1. The Graph Indexing

**Idea**: Use The Graph to index chain data for dashboards.

**Verdict**: Partial fit. Good for chain data aggregation, but doesn't solve logs/metrics/shell.

**Hybrid usage**: Could add The Graph for chain analytics while using provider APIs for operational data.

### 2. IPFS for Logs

**Idea**: Providers pin logs to IPFS, customers fetch.

**Verdict**: Poor UX. High latency, no real-time streaming, expensive at scale.

### 3. Libp2p Direct P2P

**Idea**: Portal connects directly to provider nodes via libp2p.

**Verdict**: Over-engineered. Provider daemon already exposes HTTP/WebSocket, libp2p adds complexity without benefit.

### 4. Centralized Aggregator

**Idea**: Run a central service that aggregates from all providers.

**Verdict**: Defeats decentralization. Single point of failure/censorship.

## Conclusion

The **hybrid architecture** delivers:

1. **Decentralization**: Providers run their own APIs, no central control
2. **Rich UX**: Web2-like dashboards, logs, shell access
3. **Trust**: Financial/identity operations anchored to blockchain
4. **Flexibility**: Providers can use Waldur or any compatible backend
5. **Practicality**: Builds on existing provider daemon code

This is **Decentralized Web 2.0** - the operational layer is distributed across providers while the trust layer remains on-chain.

## Appendix: API Specification

See [provider-api-spec.md](./provider-api-spec.md) for full OpenAPI specification of provider-side APIs.

## References

- [ADR-001: Portal Architecture Decision](./ADR-001-portal-architecture.md)
- [Provider Daemon Waldur Integration](./../provider-daemon-waldur-integration.md)
- [Akash Network Provider](https://github.com/akash-network/provider)
- [Cosmos AuthZ Module](https://docs.cosmos.network/main/modules/authz)
- [Cosmos Group Module](https://docs.cosmos.network/main/modules/group)
