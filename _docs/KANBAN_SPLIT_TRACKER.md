# Kanban Task Split Tracker

## Overview

As of **2026-02-02**, the VirtEngine backlog was split 60:40 across two kanban boards to enable parallel agent execution and reduce Copilot rate limit issues.

## Split Details

| Kanban                    | Task Count | Focus Areas                                    |
| ------------------------- | ---------- | ---------------------------------------------- |
| **Primary** (vibe-kanban) | 52 tasks   | Core chain modules, CLI, security, fixes       |
| **Secondary** (external)  | 34 tasks   | Portal, Waldur integration, HPC CLI, E2E tests |

## Secondary Kanban Tasks (DO NOT RECREATE)

The following 34 tasks exist in the secondary kanban and should NOT be recreated:

### Testing Tasks

- `test(provider): E2E tests for Waldur infrastructure adapters`
- `test(enclave): Verify all enclave CLI commands work`
- `test(ante): Comprehensive VEID ante handler test coverage`
- `test(hpc): Implement HPC E2E test suite`
- `test(e2e): Waldur integration tests - customer and provider flows - 25M`

### TEE & ML Tasks

- `feat(tee): Hardware integration and attestation testing`
- `feat(ml): Train and publish VEID ML model artifacts`
- `feat(bme,oracle): Integrate bank keeper for token operations`

### HPC CLI Tasks (Duplicate Series)

- `HPC-E2E-001: E2E Tests for HPC Job Lifecycle`
- `HPC-DOCS-001: User Documentation for HPC Job Submission`
- `HPC-CLI-003: Wire HPC Templates CLI into Main Binary`
- `HPC-CLI-002: Implement HPC Query CLI Commands`
- `HPC-CLI-001: Implement HPC Transaction CLI Commands`

### Waldur Integration (25-Series)

- `docs(waldur): comprehensive integration documentation and ADRs - 25N`
- `feat(portal): landing page - hero, stats, featured offerings - 25L`
- `feat(sdk): TypeScript SDK - query and transaction clients for all modules - 25K`
- `feat(portal): wallet authentication - connect Keplr, Leap, Cosmostation - 25J`
- `feat(portal): provider order dashboard - view and approve chain orders - 25I`
- `feat(waldur): offering publication - Waldur to chain sync and portal UI - 25H`
- `feat(portal): customer order tracking - real-time status and access - 25G`
- `feat(portal): customer order flow - configure and submit on-chain orders - 25F`
- `feat(portal): customer marketplace - browse offerings from chain - 25E`
- `feat(portal): provider admin separation - Waldur vs decentralized controls - 25D`
- `feat(waldur): support ticket routing - chain to provider service desk - 25C`
- `feat(waldur): customer order routing - chain to provider Waldur - 25B`
- `feat(waldur): auto-create marketplace categories on localnet startup - 25A`

### Portal Features (24-Series)

- `refactor(portal): integrate existing lib/portal and lib/capture components - 24K`
- `feat(portal): provider pricing and bid configuration - 24J`
- `feat(portal): governance participation (proposals, voting) - 24I`
- `feat(portal): wallet integration (Keplr, Leap, Cosmostation) - 24H`
- `docs(portal): feature parity checklist and tracking - 24G`
- `perf(portal): performance optimization and Core Web Vitals - 24F`
- `fix(security): portal security review and audit - 24E`
- `feat(portal): metrics and usage dashboards - 24D`

## Instructions for Future Planners

1. **Before creating new tasks**, check this list to avoid duplicates
2. **Primary kanban** handles: chain modules (x/\*), provider daemon, security, CI fixes
3. **Secondary kanban** handles: portal UI, Waldur integration, SDK, E2E testing
4. When the secondary kanban is merged back, delete this tracking file

## JSON Export for Secondary Kanban

The full JSON export for the secondary kanban is available in the git history or can be regenerated from this list.

---

_Last updated: 2026-02-02_
