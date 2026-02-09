# Chain SDK (sdk/ts) — AGENTS Guide

## Package Overview
- Purpose: TypeScript SDK for VirtEngine chain APIs, gRPC clients, wallet helpers, and provider SDK access (`sdk/ts/package.json:2`).
- Use this package for TypeScript/JS integrations; use Go modules under `x/` for on-chain logic.
- Key exports / public API surface:
  - Barrel exports from `sdk/ts/src/index.ts:1`.
  - Chain SDK factories: `createChainNodeSDK`, `createChainNodeWebSDK` (`sdk/ts/src/sdk/index.ts:2`).
  - Provider SDK factory: `createProviderSDK` (`sdk/ts/src/sdk/index.ts:4`).

## Architecture
- Entry points:
  - `sdk/ts/src/index.ts:1` — public API barrel.
  - `sdk/ts/src/sdk/index.ts:1` — SDK factory exports.
- Directory layout:
  - `sdk/ts/src/clients/` — generated gRPC/Connect clients.
  - `sdk/ts/src/generated/` — generated proto clients and patches.
  - `sdk/ts/src/sdk/` — SDK factories and transports.
  - `sdk/ts/src/network/` and `sdk/ts/src/wallet/` — network helpers and signing.
  - `sdk/ts/test/` — unit and functional tests.

## Core Concepts
- Query vs tx: `createChainNodeSDK` builds gRPC query transport and optional tx transport, falling back to a noop transport when a signer is missing (`sdk/ts/src/sdk/chain/createChainNodeSDK.ts:17`).
- Web vs node: `createChainNodeWebSDK` targets gRPC-gateway endpoints for browser usage (`sdk/ts/src/sdk/chain/createChainNodeWebSDK.ts:15`).
- Provider SDK: `createProviderSDK` constructs provider-service clients with optional mTLS and retry interceptors (`sdk/ts/src/sdk/provider/createProviderSDK.ts:12`).

## Usage Examples

### Node SDK (gRPC)
```ts
import { createChainNodeSDK } from "@virtengine/chain-sdk";

const sdk = createChainNodeSDK({
  query: { baseUrl: "https://grpc.testnet.virtengine.com" },
});

const identity = await sdk.virtengine.veid.v1.identity({ accountAddress: "..."} );
```

### Web SDK (gRPC-gateway)
```ts
import { createChainNodeWebSDK } from "@virtengine/chain-sdk";

const sdk = createChainNodeWebSDK({
  query: { baseUrl: "https://api.testnet.virtengine.com" },
});
```

### Provider SDK (mTLS)
```ts
import { createProviderSDK } from "@virtengine/chain-sdk";

const providerSDK = createProviderSDK({
  baseUrl: "https://provider.example.com:8443",
  authentication: { type: "mtls", cert: "<pem>", key: "<pem>" },
});
```

## Implementation Patterns
- Add new API surfaces by updating generators under `sdk/ts/src/generated/` and re-export via `sdk/ts/src/index.ts:1`.
- New SDK factories should live under `sdk/ts/src/sdk/` and be exported from `sdk/ts/src/sdk/index.ts:1`.
- Keep retry and transport options wired through to avoid breaking existing integrations (`sdk/ts/src/sdk/provider/createProviderSDK.ts:12`).
- Anti-patterns:
  - Do not hardcode endpoint URLs inside SDK factories; always accept them via options.
  - Do not bypass generated patches when composing SDKs (`sdk/ts/src/sdk/chain/createChainNodeSDK.ts:36`).

## API Reference
- `createChainNodeSDK(options: ChainNodeSDKOptions)` (`sdk/ts/src/sdk/chain/createChainNodeSDK.ts:17`).
- `createChainNodeWebSDK(options: ChainNodeWebSDKOptions)` (`sdk/ts/src/sdk/chain/createChainNodeWebSDK.ts:15`).
- `createProviderSDK(options: ProviderSDKOptions)` (`sdk/ts/src/sdk/provider/createProviderSDK.ts:12`).
- Public exports barrel: `sdk/ts/src/index.ts:1`.

## Dependencies & Environment
- Node engine: `>=22.14.0` (`sdk/ts/package.json:106`).
- Key deps: `@connectrpc/*`, `@cosmjs/*`, `jsrsasign`, `long` (`sdk/ts/package.json:60`).
- Build outputs live in `sdk/ts/dist` (package `files` list, `sdk/ts/package.json:34`).

## Configuration
- Configure endpoints via SDK factory options (baseUrl, auth settings).
- No required environment variables for the SDK itself.

## Testing
- Tests live in `sdk/ts/test/` (unit + functional).
- Commands:
  - `npm test` (runs Jest, `sdk/ts/package.json:41`).
  - `npm run test:unit` and `npm run test:functional` for focused suites.

## Troubleshooting
- gRPC transport errors:
  - Cause: Incorrect baseUrl or missing gateway endpoints.
  - Fix: Verify the query baseUrl matches the environment (testnet vs mainnet).
- Auth failures on provider SDK:
  - Cause: Missing or invalid mTLS certs.
  - Fix: Re-check `authentication` options and cert/key formatting.
