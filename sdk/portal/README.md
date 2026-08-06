# VirtEngine Portal (Vite + React)

React + TypeScript SDK example shell for the VirtEngine blockchain. This package speaks directly to the chain and provider-facing APIs and is intended to mirror the same runtime contract the production portal uses, without depending on the Next.js application.

## Requirements

- Node.js 20+
- pnpm (`corepack enable`)
- A reachable VirtEngine RPC/LCD pair
- Keplr / Leap / Cosmostation browser extension
- Optional WalletConnect project ID if mobile or QR-based wallet flows are required

## Quick start

```bash
pnpm --ignore-workspace -C sdk/portal install
pnpm -C sdk/portal dev
```

Open `http://localhost:5173`.

## Environment

The SDK accepts either Vite-style `VITE_*` vars or the same `NEXT_PUBLIC_*` vars used by the main portal. If both are set, `VITE_*` wins.

```bash
VITE_CHAIN_ID=virtengine-testnet-1
VITE_CHAIN_RPC=https://rpc.testnet.virtengine.com
VITE_CHAIN_REST=https://api.testnet.virtengine.com
VITE_CHAIN_WS=wss://ws.testnet.virtengine.com
VITE_APP_URL=https://staging.portal.virtengine.com
VITE_PROVIDER_DAEMON_URL=https://provider.staging.virtengine.com
VITE_WALLET_CONNECT_PROJECT_ID=your_walletconnect_project_id
```

For localnet development, override the chain endpoints explicitly rather than relying on an implicit localhost default.

## Docker (dev mode)

```bash
docker compose up ve-portal
```

This runs Vite in dev mode with port `5173` exposed.

## Scripts

```bash
pnpm -C sdk/portal type-check
pnpm -C sdk/portal test
pnpm -C sdk/portal build
pnpm -C sdk/portal preview
```
