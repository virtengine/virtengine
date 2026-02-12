# Portal Chain Client

The portal chain module provides a reusable query + signing layer for VirtEngine's REST/RPC endpoints.
It bridges the existing wallet adapters with CosmJS clients so the portal app can request on-chain data
without scattering raw fetch calls.

## Getting Started

```ts
import {
  createChainConfig,
  createChainQueryClient,
  createWalletSigner,
  createChainSigningClient,
} from "virtengine-portal-lib";

const config = createChainConfig({
  chainId: "virtengine-1",
  rpcEndpoints: ["https://rpc.virtengine.com"],
  restEndpoints: ["https://api.virtengine.com"],
});

const queryClient = createChainQueryClient(config);
const status = await queryClient.getChainStatus();

const signer = createWalletSigner(config.endpoints.chainId, walletAdapter);
const signingClient = createChainSigningClient(config, signer);
```

## Queries

Market, escrow, veid, governance, staking, and provider helpers live in `queries/*`.
Each helper accepts a `ChainQueryClient` and optional pagination/request overrides.

```ts
import { fetchMarketOfferings } from "virtengine-portal-lib";

const { data } = await fetchMarketOfferings(queryClient, {
  pagination: { limit: 50 },
});
```

## Transactions

Transaction builders return `typeUrl` + `value` objects ready to be signed using a
`SigningStargateClient` instance (register module codecs in your registry when needed).

```ts
import { buildMsgCreateOrder } from "virtengine-portal-lib";

const msg = buildMsgCreateOrder({
  cpu: 4,
  memoryGb: 16,
  storageGb: 200,
  price: "1200000",
});
```
