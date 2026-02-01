# VirtEngine Deployment and Operations Guide

This guide replaces the old "_run" references and provides a clear path for:
- Local deployment (single-node development)
- Multi-node deployment (VMs or bare metal)
- Kubernetes deployment (multi-node)
- Validator operations (VEID + consensus)
- Provider operations (Waldur + listings)
- Joining an existing cluster (state sync, peers)

If you only need a local dev environment, start with the localnet section below.

## Quick Links

- Local dev environment: `_docs/development-environment.md`
- Validator onboarding: `_docs/validator-onboarding.md`
- Provider onboarding: `_docs/provider-guide.md`
- Provider/Waldur integration: `_docs/provider-daemon-waldur-integration.md`
- Provider ops runbook: `docs/operations/runbooks/PROVIDER_OPERATIONS.md`
- State sync bootstrap: `scripts/state-sync-bootstrap.sh`
- Kubernetes manifests: `deploy/kubernetes/`

## 1) Deploy the System Locally (Single Node)

VirtEngine ships a localnet script that runs the chain, provider-daemon, portal,
gateway, and mock services using Docker Compose.

Prerequisites:
- Docker + Docker Compose
- Go 1.21+ (tests) / Go 1.22+ (localnet script notes)
- Bash shell (WSL2 on Windows)

Start localnet:
```bash
chmod +x scripts/localnet.sh scripts/init-chain.sh
./scripts/localnet.sh start
```

Useful commands:
```bash
./scripts/localnet.sh status
./scripts/localnet.sh logs
./scripts/localnet.sh logs virtengine-node
./scripts/localnet.sh test
./scripts/localnet.sh stop
./scripts/localnet.sh reset
```

Notes:
- This starts a single validator chain plus supporting services.
- Windows users should run localnet from WSL2 as documented in
  `_docs/development-environment.md`.

## 2) Deploy Across Multiple Nodes (VMs or Bare Metal)

Use this when you want a real multi-node devnet without Kubernetes.
The flow is standard Cosmos SDK: create a shared genesis, collect gentxs,
distribute `genesis.json`, then start each node with proper peer config.

High-level steps:
1) Build the binary:
   ```bash
   make virtengine
   ```
2) Initialize a "genesis coordinator" node:
   ```bash
   virtengine init "devnet-validator-0" --chain-id devnet-1
   ```
3) Create the first validator key + gentx on the coordinator:
   ```bash
   virtengine keys add validator-0 --keyring-backend file
   virtengine genesis add-account $(virtengine keys show validator-0 -a --keyring-backend file) 100000000000uve
   virtengine genesis gentx validator-0 10000000000uve --chain-id devnet-1 --keyring-backend file
   ```
4) For each additional validator:
   - Run `virtengine init` on that host with the same chain ID.
   - Create a key and gentx on that host.
   - Send the gentx file to the coordinator.
5) On the coordinator, collect gentxs:
   ```bash
   virtengine genesis collect
   ```
6) Distribute the final `~/.virtengine/config/genesis.json` to every node.
7) Configure peers in `config.toml` (seeds/persistent_peers) for all nodes.
8) Start nodes:
   ```bash
   virtengine start
   ```

Tip: for rapid bootstrap, use state sync via `scripts/state-sync-bootstrap.sh`.

## 3) Deploy Across Multiple Nodes (Kubernetes)

The repo includes Kustomize overlays for dev/staging/prod plus ArgoCD apps.
This is the recommended path for a production-like multi-node deployment.

Options:
- Kustomize: apply manifests directly.
- ArgoCD: deploy via GitOps using `deploy/argocd/`.

Kustomize (example):
```bash
kubectl apply -k deploy/kubernetes/overlays/dev
```

ArgoCD (example):
```bash
kubectl apply -k deploy/argocd/base
kubectl apply -k deploy/argocd/apps
```

What gets deployed:
- `virtengine-node` (validator/full node)
- `provider-daemon`
- monitoring stack (optional)

Review and customize:
- `deploy/kubernetes/base/*` for core services
- `deploy/kubernetes/overlays/*` for env-specific overrides

## 4) Become a Validator Operator

If you plan to validate blocks and run VEID scoring, start here:
- `_docs/validator-onboarding.md` (end-to-end setup and operations)

Minimum workflow:
1) Install `virtengine` and initialize the node.
2) Configure P2P/RPC/REST/gRPC ports in `config.toml` and `app.toml`.
3) Sync the chain (use state sync when joining an existing network).
4) Create a validator transaction:
   ```bash
   virtengine tx staking create-validator ...
   ```
5) Set up VEID scoring dependencies (model + runtime) per the onboarding guide.

Operational essentials:
- Keep validator keys secure (HSM strongly recommended).
- Do not run two validators with the same key (double-sign risk).
- Monitor uptime and missed blocks.

## 5) Become a Provider Operator

Providers run the provider-daemon and connect a control plane (Waldur or
Kubernetes/SLURM adapters) to fulfill on-chain workloads.

Start here:
- `_docs/provider-guide.md`
- `_docs/provider-daemon-waldur-integration.md`
- `docs/operations/runbooks/PROVIDER_OPERATIONS.md`

Minimum workflow:
1) Install `provider-daemon` and create a provider config.
2) Register your provider on-chain:
   ```bash
   virtengine tx provider create ...
   ```
3) Register encryption keys for encrypted order payloads:
   ```bash
   virtengine tx encryption register-recipient-key ...
   ```
4) Connect to your orchestration backend:
   - Kubernetes adapter for container workloads
   - SLURM adapter for HPC workloads
   - Waldur bridge for cloud/HPC control planes
5) Publish listings (offerings) and start bidding.

Waldur note:
- Waldur is provider-operated; validators do not run Waldur.
- The provider-daemon signs callbacks back to chain.
- See `_docs/provider-daemon-waldur-integration.md` for the attach guide and
  offering mapping (Waldur offering UUIDs to on-chain offering IDs).

## 6) Joining an Existing Cluster (Devnet/Testnet/Mainnet)

If a network already exists, you need the official chain ID, genesis file,
and seed/persistent peers from the network operator.

Steps:
1) Initialize the node:
   ```bash
   virtengine init "my-node" --chain-id <chain-id>
   ```
2) Place the official `genesis.json` in `~/.virtengine/config/`.
3) Configure `seeds` and `persistent_peers` in `config.toml`.
4) Use state sync for fast bootstrapping:
   ```bash
   scripts/state-sync-bootstrap.sh --rpc-servers <rpc1>,<rpc2>
   ```
5) Start the node:
   ```bash
   virtengine start
   ```

Important:
- As of 2026-02-01, public mainnet/testnet deployments are not published in
  this repo. When they are, the official genesis and seed list will be
  published alongside the network release notes.

## 7) Operating the System

For day-2 operations and incident handling, use the existing runbooks:
- Provider ops: `docs/operations/runbooks/PROVIDER_OPERATIONS.md`
- Provider deployment troubleshooting: `docs/operations/runbooks/provider-deployment.md`
- Disaster recovery: `_docs/disaster-recovery.md`
- Horizontal scaling: `_docs/horizontal-scaling-guide.md`

Keep validators and providers on compatible versions:
- `docs/COMPATIBILITY.md`

