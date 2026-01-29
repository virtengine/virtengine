# Provider Daemon - Waldur Integration Design

## Purpose
Define the provider-daemon bridge between VirtEngine on-chain marketplace events and Waldur control-plane actions for provisioning, termination, and usage reporting. This doc is the authoritative integration design for decentralized operation.

## Design Principles
- On-chain state is authoritative; off-chain executes intent.
- Providers control provisioning; validators only verify signatures and state transitions.
- Every off-chain action is signed and replay-protected.
- All workflows are idempotent and resilient to retries.

## Components
- **VirtEngine chain**: source of truth for orders/allocations and settlement.
- **Provider daemon**: consumes chain events and calls Waldur APIs.
- **Waldur**: provider-operated control plane for cloud/HPC resources.
- **Artifact store (IPFS/Waldur)**: encrypted config payloads referenced by `encrypted_config_ref`.

## Event Flow Diagrams

### 1) Order -> Allocation -> Provision

```
[Customer] --(MsgCreateOrder)--> [Chain]
   |                                  |
   |                       Event: OrderCreated
   |                                  v
   |                        [Provider Daemon]
   |                                  |
   |                        (Bid Decision)
   |                                  v
   |                         MsgCreateBid
   |                                  v
[Chain] --(Bid Accepted)--> Event: AllocationCreated
   |                                  |
   |                                  v
   |                        [Provider Daemon]
   |                         create Waldur order
   |                                  |
   |                         approve Waldur order
   |                                  v
   |                send Callback: provision_requested
   v                                  v
[Chain] <--- MsgWaldurCallback ---- [Provider Daemon]
```

### 2) Provisioning -> Active -> Usage

```
[Provider Daemon] -- Waldur provision --> [Waldur]
          |                                  |
          |                         Resource provisioned
          |                                  v
          |                Callback: allocation_active
          v                                  v
        [Chain] <--- MsgWaldurCallback ---- [Provider Daemon]
          |
          |  periodic usage report
          v
       MsgUsageReport
```

### 3) Termination

```
[Customer/Admin] -- MsgTerminate --> [Chain]
   |                                   |
   |                      Event: TerminateRequested
   |                                   v
   |                           [Provider Daemon]
   |                             delete Waldur resource
   |                                   v
   |                Callback: allocation_terminated
   v                                   v
[Chain] <--- MsgWaldurCallback ---- [Provider Daemon]
```

## Canonical State Mapping

| VirtEngine Allocation State | Waldur Resource State | Notes |
|---|---|---|
| pending | provisioning | Allocation created, provisioning queued |
| provisioning | provisioning | In-progress |
| active | provisioned | Resource ready |
| terminating | terminating | Termination requested |
| terminated | terminated | Resource removed |
| failed | failed | Provisioning failed |
| rejected | rejected | Provider rejected |

Order mapping is derived from allocation state:
- matched -> provisioning
- provisioning -> active
- pending_termination -> terminated

## Callback Schema (Provider -> Chain)

### Payload (JSON)
```json
{
  "id": "wcb_allocation_abc123_8f4e2c1a",
  "action_type": "provision|terminate|status_update|usage_report",
  "waldur_id": "UUID",
  "chain_entity_type": "allocation",
  "chain_entity_id": "customer/1/1",
  "payload": {
    "state": "provisioning|active|failed|terminated",
    "reason": "string",
    "encrypted_config_ref": "ipfs://... or object://...",
    "usage_period_start": "unix",
    "usage_period_end": "unix",
    "usage": { "cpu_hours": "123", "gpu_hours": "4", "ram_gb_hours": "512" }
  },
  "signer_id": "provider_addr_or_key_id",
  "nonce": "unique_nonce",
  "timestamp": "unix",
  "expires_at": "unix",
  "signature": "base64(ed25519(sig(payload)))"
}
```

### Signature Rules
- Signature is over canonical bytes of: `id|action_type|waldur_id|chain_entity_type|chain_entity_id|nonce|timestamp|payload_hash`.
- Verify against provider public key in `x/provider`.
- Reject if nonce already used or expired.

## Provider Daemon Responsibilities
- Subscribe to chain events with checkpointing and idempotent replays.
- Resolve allocation to provider-owned offering.
- Invoke Waldur APIs using `pkg/waldur`:
  - Create/approve order
  - Provision or terminate resource
  - Pull usage data
- Emit signed callbacks to chain.

## Provider Attach Guide (Waldur)
This describes how a provider attaches their own Waldur instance to VirtEngine.

### 1) Deploy Waldur per provider
- Each provider runs its own Waldur control plane (VMs, OpenStack, SLURM, or cloud connectors).
- This keeps provider resources isolated and avoids shared global state.

### 2) Create the Waldur marketplace objects
- Create a **Customer** and a **Project** in Waldur (or use existing).
- Create a **Marketplace Offering** per VirtEngine offering you want to sell.
- Record the Waldur offering UUIDs and the project UUID.

### 3) Generate an API token
- Create a Waldur user with appropriate provider permissions.
- Generate a token; store it as `WALDUR_TOKEN`.

### 4) Map VirtEngine offerings to Waldur offerings
Populate a map file:
```
config/waldur-offering-map.json
{
  "<provider_bech32>/1": "<waldur-offering-uuid>",
  "<provider_bech32>/2": "<waldur-offering-uuid>"
}
```
The key is the on-chain offering ID (`provider/sequence`).

### 5) Start provider-daemon with Waldur enabled
Example CLI flags (use your endpoints):
```
provider-daemon start \
  --chain-id virtengine-1 \
  --node tcp://localhost:26657 \
  --provider-key provider \
  --waldur-enabled \
  --waldur-base-url https://waldur.example.com/api \
  --waldur-token $WALDUR_TOKEN \
  --waldur-project-uuid <project-uuid> \
  --waldur-offering-map config/waldur-offering-map.json \
  --waldur-chain-submit \
  --waldur-chain-key <chain-key-name> \
  --waldur-chain-grpc localhost:9090
```

### 6) Validate connectivity
- Ensure `provider-daemon` logs “Waldur Bridge: started”.
- Confirm Waldur health check passes at startup.
- Submit a test order and verify a Waldur order is created and approved.

### 7) Security and operational notes
- Store Waldur tokens in a secret manager; do not commit to repo.
- Run provider-daemon and Waldur in the same trust domain.
- Rotate API tokens and chain keys regularly.

## Waldur API Operations (Go client wrapper)
- Marketplace: list offerings, create order, approve order, terminate resource.
- OpenStack/AWS/Azure: create/terminate resources depending on offering type.
- SLURM: create allocation, manage associations, list jobs.

## HPC / Supercomputer Integration
- HPC offerings correspond to Waldur SLURM allocations.
- Provider daemon creates SLURM allocation with CPU/GPU limits.
- On-chain job submission maps to Waldur SLURM job create.
- Usage is captured from Waldur/SLURM and posted to settlement module.

## Idempotency and Reconciliation
- All operations keyed by `allocation_id` and `waldur_id`.
- Provider daemon must re-query Waldur state if callbacks fail.
- Chain must tolerate duplicate callbacks (idempotent transitions).

## Observability
- Metrics: callback latency, Waldur API errors, retries, provisioning duration.
- Logs include allocation ID, order ID, and Waldur UUID.

## Security
- Provider public key required for callbacks.
- Nonce replay protection.
- Strict state transition validation.
- Encrypted config payloads only referenced; never decrypted on-chain.

