# VirtEngine Task Specifications

This directory contains detailed task specifications for the hybrid portal architecture implementation.

**Note:** This directory is gitignored - task files are for local agent planning only.

## Task Series 29: Hybrid Portal Architecture

| ID | Title | Priority | Wave | Est. LOC |
|----|-------|----------|------|----------|
| [29A](./29A-ml-training-savemodel.md) | ML Training + SavedModel Export | P0 | 1 | 500 |
| [29B](./29B-model-hash-governance.md) | Model Hash Computation + Governance | P0 | 2 | 1000 |
| [29C](./29C-lib-portal-shared-library.md) | lib/portal Shared Library | P0 | 1 | 3000-5000 |
| [29D](./29D-provider-api-typescript-client.md) | Provider API TypeScript Client | P0 | 2 | 2000-3000 |
| [29E](./29E-wallet-signed-auth.md) | Wallet-Signed Request Auth | P0 | 2 | 1500 |
| [29F](./29F-enhanced-portal-api.md) | Enhanced portal_api.go Endpoints | P1 | 2 | 2000 |
| [29G](./29G-multi-provider-client.md) | Aggregated Multi-Provider Client | P1 | 3 | 2500 |
| [29H](./29H-organization-management.md) | Organization Management (x/group) | P1 | 3 | 2000 |
| [29I](./29I-support-ticket-flow.md) | Support Ticket Hybrid Flow | P1 | 3 | 2000 |
| [29J](./29J-billing-dashboard.md) | Billing/Invoice Dashboard | P1 | 3 | 2000 |
| [29K](./29K-metrics-aggregation.md) | Metrics/Dashboard Aggregation | P2 | 4 | 2500 |
| [29L](./29L-openapi-spec.md) | OpenAPI Spec for Provider API | P2 | 4 | 500 |

## Execution Waves

```
Wave 1 (Parallel - 2 weeks):
├── 29A: ML training execution
└── 29C: lib/portal shared library

Wave 2 (Sequential - 3-4 weeks):
├── 29B: Model hash + governance (← 29A)
├── 29D: Provider API TypeScript client (← 29C)
├── 29E: Wallet-signed auth (← 29D)
└── 29F: Enhanced provider API (parallel)

Wave 3 (Portal Features - 3-4 weeks):
├── 29G: Multi-provider aggregation (← 29D, 29E)
├── 29H: Organization management (← 29D)
├── 29I: Support ticket flow (← 29D, 29F)
└── 29J: Billing dashboard (← 29D)

Wave 4 (Polish - 1-2 weeks):
├── 29K: Metrics aggregation (← 29G)
└── 29L: OpenAPI documentation (← 29F)
```

## Total Estimated Timeline: 12-16 weeks

---

## Task Series 30: Production Readiness & Patent Compliance

Tasks identified from gap analysis against patent specification AU2024203136A1.

| ID | Title | Priority | Status | Est. LOC |
|----|-------|----------|--------|----------|
| [30A](./30A-mobile-veid-capture-app.md) | React Native VEID Capture App | P1 | todo | 8000-12000 |
| [30B](./30B-hsm-key-management.md) | HSM Integration for Key Management | P0 | todo | 4400 |
| [30C](./30C-multi-region-deployment.md) | Multi-Region Deployment + DR | P1 | todo | 4050 |
| [30D](./30D-tokenomics-simulation.md) | Tokenomics Simulation & Validation | P1 | todo | 6300 |
| [30E](./30E-security-audit-coordination.md) | Third-Party Security Audit | **BLOCKER** | todo | 2200 |

## Series 30 Dependencies

```
Pre-Mainnet Blockers (P0):
├── 30E: Security audit coordination (START IMMEDIATELY)
│   └── Requires: All other development complete
└── 30B: HSM integration
    └── Required by: 30E (crypto audit scope)

Pre-Mainnet Critical (P1):
├── 30A: Mobile VEID app (patent requirement)
├── 30C: Multi-region + DR
│   └── Depends on: 30B (regional HSM)
└── 30D: Tokenomics simulation
    └── Independent analysis

Execution Order:
1. 30E started immediately (RFP, vendor selection)
2. 30B (HSM) - parallel with audit prep
3. 30D (Tokenomics) - parallel with development
4. 30A (Mobile) - feature development
5. 30C (Multi-region) - after 30B
6. 30E execution (after feature complete)
```

## Combined Timeline Estimate

| Phase | Duration | Tasks |
|-------|----------|-------|
| Series 29 Wave 1-2 | 5-6 weeks | 29A-F |
| Series 29 Wave 3-4 | 4-6 weeks | 29G-L |
| Series 30 Parallel | 4-6 weeks | 30A, 30B, 30D |
| Audit Execution | 6-8 weeks | 30E |
| Multi-Region | 2-3 weeks | 30C |

**Total to Production:** 14-20 weeks (with overlapping work)
