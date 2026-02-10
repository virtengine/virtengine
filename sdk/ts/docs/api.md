# VirtEngine TypeScript SDK API Overview

This document provides a quick reference for the TypeScript SDK.

## Entry points

- `@virtengine/chain-sdk`
- `@virtengine/chain-sdk/web`

## Core factories

- `createChainNodeSDK(options)`
- `createChainNodeWebSDK(options)`
- `createProviderSDK(options)`

## High-level module clients

Use `createVirtEngineClient` for a consolidated interface:

```ts
import { createVirtEngineClient } from "@virtengine/chain-sdk";

const client = await createVirtEngineClient({
  rpcEndpoint: "https://grpc.sandbox-2.aksh.pw:9090",
  restEndpoint: "https://api.sandbox-2.aksh.pw:443",
});

await client.veid.getIdentity("virt1...");
await client.hpc.listClusters();
```

## Transactions

- Provide a `TxClient` (or `createStargateClient`) to enable tx methods.
- You can pass `TxCallOptions` to override memo, fees, or capture broadcast results.

## Modules

- **VEID**: identity scoring, scopes, wallet management
- **MFA**: multi-factor enrollment and challenges
- **HPC**: cluster, offering, and job operations
- **Market**: orders, bids, leases
- **Escrow**: account and payment queries, deposits
- **Encryption**: key management and envelope validation
- **Roles**: role and account state queries

## Generated types

All protobuf message types are exported from the SDK bundle:

```ts
import { MsgSubmitJob, IdentityRecord, RoleAssignment } from "@virtengine/chain-sdk";
```

For full details, inspect the generated module exports under `dist/types/generated`.
