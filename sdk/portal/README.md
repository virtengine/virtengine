# VirtEngine Portal (Vite + React)

React + TypeScript scaffold for the VirtEngine blockchain. This portal speaks directly to the chain (RPC/LCD) and is **not** tied to Waldur.

## Requirements

- Node.js 20+
- pnpm (`corepack enable`)
- Localnet running (`docker compose up virtengine-node`)
- Keplr / Leap / Cosmostation browser extension

## Quick start

```bash
pnpm -C sdk/portal install
pnpm -C sdk/portal dev
```

Open `http://localhost:5173`.

## Environment

The portal reads Vite env vars (`VITE_*`). Defaults target localnet.

```
VITE_CHAIN_ID=virtengine-localnet-1
VITE_CHAIN_RPC=http://localhost:26657
VITE_CHAIN_REST=http://localhost:1317
```

## Docker (dev mode)

```bash
docker compose up ve-portal
```

This runs Vite in dev mode with port `5173` exposed.

## Scripts

```bash
pnpm -C sdk/portal build
pnpm -C sdk/portal preview
```
