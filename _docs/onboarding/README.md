# VirtEngine Developer Onboarding Guide

Welcome to VirtEngine! This guide will help you get started as a contributor to the project.

## Quick Start

1. **[Environment Setup](./01-environment-setup.md)** - Set up your development environment
2. **[Architecture Overview](./02-architecture-overview.md)** - Understand the system design
3. **[Code Contribution](./03-code-contribution.md)** - Learn how to contribute
4. **[Testing Guide](./04-testing-guide.md)** - Write and run tests
5. **[Code Review Checklist](./05-code-review-checklist.md)** - Review standards
6. **[Module Development](./06-module-development.md)** - Build blockchain modules
7. **[Patterns & Anti-patterns](./07-patterns-antipatterns.md)** - Best practices
8. **[Debugging Guide](./08-debugging-guide.md)** - Troubleshooting techniques

## What is VirtEngine?

VirtEngine is a Cosmos SDK-based blockchain for decentralized cloud computing with:

- **VEID** - ML-powered identity verification
- **MFA Module** - Multi-factor authentication for sensitive transactions
- **Encryption Module** - Public-key encryption for on-chain data
- **Market Module** - Cloud marketplace with order/bid/lease lifecycle
- **Provider Daemon** - Off-chain bidding and infrastructure provisioning
- **HPC Module** - High-performance computing job scheduling

## Repository Structure

```
virtengine/
├── app/              # Cosmos SDK app wiring, ante handlers, genesis
├── cmd/              # CLI binaries (virtengine, provider-daemon)
├── x/                # Blockchain modules
│   ├── veid/         # Identity verification
│   ├── mfa/          # Multi-factor authentication
│   ├── encryption/   # Encryption primitives
│   ├── market/       # Marketplace orders/bids/leases
│   ├── escrow/       # Payment escrow
│   ├── roles/        # Role-based access control
│   └── hpc/          # HPC job scheduling
├── pkg/              # Off-chain services and utilities
│   ├── provider_daemon/  # Provider bidding engine
│   ├── inference/        # ML scoring service
│   └── workflow/         # Workflow orchestration
├── ml/               # Python ML pipelines
├── tests/            # Integration and E2E tests
├── _docs/            # Internal documentation
└── docs/             # Public-facing documentation
```

## Prerequisites

Before you begin, ensure you have:

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.21+ | Core language |
| GNU Make | 4.0+ | Build system |
| Docker | 20.10+ | Containerization |
| Node.js | 18+ | TypeScript SDK |
| Python | 3.10+ | ML pipelines |
| direnv | 2.32+ | Environment management |

## First Day Checklist

- [ ] Clone the repository
- [ ] Set up development environment ([guide](./01-environment-setup.md))
- [ ] Build the project (`make virtengine`)
- [ ] Run unit tests (`go test ./x/... ./pkg/...`)
- [ ] Start the localnet (`./scripts/localnet.sh start`)
- [ ] Read architecture overview ([guide](./02-architecture-overview.md))
- [ ] Review contribution guidelines ([guide](./03-code-contribution.md))

## First Week Goals

- [ ] Complete all onboarding documentation
- [ ] Successfully run integration tests
- [ ] Make a small contribution (typo fix, doc improvement)
- [ ] Attend a code review
- [ ] Familiarize yourself with 2-3 modules in `x/`

## Getting Help

| Resource | Link |
|----------|------|
| Discord | [discord.gg/virtengine](https://discord.gg/virtengine) |
| Documentation | [docs.virtengine.com](https://docs.virtengine.com) |
| GitHub Issues | [github.com/virtengine/virtengine/issues](https://github.com/virtengine/virtengine/issues) |
| Architecture Docs | [_docs/architecture.md](../architecture.md) |

## Next Steps

Start with the [Environment Setup](./01-environment-setup.md) guide to get your development environment ready.
