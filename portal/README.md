# VirtEngine Portal

A production-ready Next.js 14 portal application for VirtEngine - a decentralized cloud computing marketplace with ML-powered identity verification.

## Features

- 🛒 **Marketplace**: Browse and purchase compute resources from providers worldwide
- 🔐 **Identity (VEID)**: Complete identity verification with ML-powered scoring
- ⚡ **HPC Jobs**: Submit and manage high-performance computing workloads
- 🖥️ **Provider Console**: Manage offerings and infrastructure as a provider
- 🗳️ **Governance**: Participate in protocol governance and voting
- 🌙 **Dark Mode**: Full dark mode support with system preference detection
- ♿ **Accessibility**: WCAG 2.1 AA compliant with keyboard navigation and screen reader support

## Tech Stack

- **Framework**: [Next.js 14](https://nextjs.org/) with App Router
- **Styling**: [Tailwind CSS](https://tailwindcss.com/) + [shadcn/ui](https://ui.shadcn.com/)
- **State Management**: [Zustand](https://github.com/pmndrs/zustand)
- **Testing**: [Vitest](https://vitest.dev/) (unit) + [Playwright](https://playwright.dev/) (E2E)
- **Type Safety**: TypeScript with strict mode

## Getting Started

### Prerequisites

- Node.js 20+
- pnpm 8+ (recommended) or npm

### Installation

```bash
# From the repository root
pnpm install

# Or just for the portal
cd portal
pnpm install
```

### Development

```bash
# Start development server
pnpm dev

# Open http://localhost:3000
```

### Environment Variables

Copy `.env.example` to `.env.local` and configure:

```bash
cp .env.example .env.local
```

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NEXT_PUBLIC_CHAIN_ID` | Blockchain chain ID | `virtengine-1` |
| `NEXT_PUBLIC_CHAIN_RPC` | RPC endpoint | `https://rpc.virtengine.com` |
| `NEXT_PUBLIC_CHAIN_REST` | REST API endpoint | `https://api.virtengine.com` |
| `NEXT_PUBLIC_WALLET_CONNECT_PROJECT_ID` | WalletConnect project ID | - |

## Scripts

| Command | Description |
|---------|-------------|
| `pnpm dev` | Start development server |
| `pnpm build` | Build for production |
| `pnpm start` | Start production server |
| `pnpm lint` | Run ESLint |
| `pnpm type-check` | Run TypeScript type checking |
| `pnpm test` | Run unit tests |
| `pnpm test:e2e` | Run E2E tests |
| `pnpm format` | Format code with Prettier |

## Project Structure

```
portal/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── (auth)/             # Authentication routes
│   │   ├── (customer)/         # Customer routes (marketplace, orders, identity)
│   │   ├── (provider)/         # Provider routes (dashboard, offerings, pricing)
│   │   ├── hpc/                # HPC routes (jobs, templates)
│   │   └── governance/         # Governance routes (proposals)
│   ├── components/
│   │   ├── layout/             # Layout components (Header, Sidebar, Footer)
│   │   ├── wallet/             # Wallet components (WalletButton, WalletModal)
│   │   ├── shared/             # Shared UI components
│   │   └── ui/                 # shadcn/ui base components
│   ├── stores/                 # Zustand stores
│   ├── config/                 # Configuration (chains, wallets, env)
│   ├── hooks/                  # Custom React hooks
│   └── lib/                    # Utility functions
├── tests/
│   ├── e2e/                    # Playwright E2E tests
│   └── unit/                   # Vitest unit tests
└── public/                     # Static assets
```

## Integration with lib/portal and lib/capture

This portal integrates with the existing component libraries:

- **`lib/portal`**: Core portal components, hooks, and utilities
- **`lib/capture`**: Document and selfie capture for identity verification

Import them using the configured aliases:

```tsx
import { useAuth, useIdentity } from '@virtengine/portal';
import { DocumentCapture, SelfieCapture } from '@virtengine/capture';
```

## Testing

### Unit Tests

```bash
# Run all unit tests
pnpm test

# Run with coverage
pnpm test:coverage

# Run in watch mode
pnpm test:watch
```

### E2E Tests

```bash
# Run all E2E tests
pnpm test:e2e

# Run with UI
pnpm test:e2e:ui

# Run specific test file
pnpm test:e2e tests/e2e/marketplace.spec.ts
```

## Deployment

### Preview Deployments

PRs automatically get preview deployments via the CI/CD pipeline.

### Production Deployment

Production deployments are triggered by pushes to the `main` branch and require manual approval.

## Contributing

1. Create a feature branch from `main`
2. Make your changes
3. Run `pnpm lint && pnpm type-check && pnpm test`
4. Submit a PR with a conventional commit message

## License

See [LICENSE](../LICENSE) for details.
