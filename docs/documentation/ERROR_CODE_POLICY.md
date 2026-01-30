# VirtEngine Error Code Policy

## Overview

All VirtEngine modules use standardized error codes to avoid conflicts with upstream dependencies and ensure consistent error handling across the platform.

## Error Code Allocation Rules

### Reserved Ranges

The following error code ranges are **RESERVED** and must not be used:

- **1-50**: Cosmos SDK core modules
- **1-50** (per module): IBC-Go modules
- **1-50** (per module): CosmWasm modules

### VirtEngine Module Ranges

VirtEngine modules use error codes starting from **100** to avoid conflicts.

#### Blockchain Modules (x/)

Each blockchain module has an allocated range of **100 error codes**:

| Module | Code Range | Description |
|--------|------------|-------------|
| veid | 1000-1099 | Identity verification and ML scoring |
| mfa | 1200-1299 | Multi-factor authentication |
| encryption | 1300-1399 | Encryption and key management |
| market | 1400-1499 | Marketplace orders and bids |
| escrow | 1500-1599 | Payment escrow |
| roles | 1600-1699 | Role-based access control |
| hpc | 1700-1799 | High-performance computing |
| provider | 1800-1899 | Provider registration and management |
| deployment | 1900-1999 | Deployment management |
| cert | 2000-2099 | Certificate management |
| audit | 2100-2199 | Audit logging |
| settlement | 2200-2299 | Payment settlement |
| benchmark | 2300-2399 | Provider benchmarking |
| staking | 2400-2499 | Staking and rewards |
| delegation | 2500-2599 | Stake delegation |
| fraud | 2600-2699 | Fraud detection |
| review | 2700-2799 | Provider reviews |
| enclave | 2800-2899 | Trusted execution environments |
| config | 2900-2999 | On-chain configuration |
| take | 3000-3099 | Fee distribution |
| marketplace | 3100-3199 | Marketplace integration |

#### Off-Chain Services (pkg/)

Off-chain services also have 100-code ranges:

| Module | Code Range | Description |
|--------|------------|-------------|
| provider_daemon | 100-199 | Provider daemon service |
| inference | 200-299 | ML inference service |
| workflow | 300-399 | Workflow engine |
| benchmark_daemon | 400-499 | Benchmark daemon |
| enclave_runtime | 500-599 | Enclave runtime |
| waldur | 600-699 | Waldur integration |
| govdata | 700-799 | Government data verification |
| edugain | 800-899 | EduGAIN federation |
| nli | 900-999 | Natural language interface |
| artifact_store | 3200-3299 | Artifact storage |
| capture_protocol | 3300-3399 | Identity capture protocol |
| payment | 3400-3499 | Payment processing |
| dex | 3500-3599 | DEX integration |
| jira | 3600-3699 | JIRA integration |
| slurm_adapter | 3700-3799 | SLURM adapter |
| ood_adapter | 3800-3899 | Open OnDemand adapter |
| moab_adapter | 3900-3999 | MOAB adapter |
| sre | 4000-4099 | SRE tooling |
| observability | 4100-4199 | Observability |
| ratelimit | 4200-4299 | Rate limiting |

## Error Code Patterns

Within each module's 100-code range, follow these conventions:

- **00-09**: Invalid input/validation errors
- **10-19**: Not found errors
- **20-29**: Already exists/conflict errors
- **30-39**: Unauthorized/permission errors
- **40-49**: State/lifecycle errors
- **50-59**: External service errors
- **60-69**: Internal errors
- **70-79**: Verification/validation errors
- **80-89**: Rate limiting/quota errors
- **90-99**: Reserved for future use

### Examples

For the `veid` module (range 1000-1099):

- `1001`: Validation error (invalid scope)
- `1010`: Not found error (scope not found)
- `1020`: Conflict error (scope already exists)
- `1030`: Unauthorized error (unauthorized access)
- `1036`: Internal error (ML inference failed)

## Usage Guidelines

### For Blockchain Modules (x/)

Use `errorsmod.Register()` from Cosmos SDK:

```go
package types

import (
    errorsmod "cosmossdk.io/errors"
)

var (
    ErrInvalidScope = errorsmod.Register(ModuleName, 1001, "invalid scope")
    ErrScopeNotFound = errorsmod.Register(ModuleName, 1010, "scope not found")
)
```

### For Off-Chain Services (pkg/)

Use the standardized error types from `pkg/errors`:

```go
import "github.com/virtengine/virtengine/pkg/errors"

// Validation error
err := errors.NewValidationError("provider_daemon", 100, "manifest", "invalid manifest format")

// Not found error
err := errors.NewNotFoundError("provider_daemon", 110, "deployment", deploymentID)

// External service error
err := errors.NewExternalError("waldur", 650, "openstack", "create_vm", "API unavailable")
```

## Validation

Validate error codes before registration:

```go
import "github.com/virtengine/virtengine/pkg/errors"

if !errors.ValidateCode("veid", 1050) {
    panic("error code out of allocated range")
}
```

## Adding New Error Codes

When adding new error codes:

1. **Check allocation**: Verify the code is within the module's range
2. **Follow patterns**: Use the appropriate category offset (00-09, 10-19, etc.)
3. **Document**: Add to `docs/errors/ERROR_CATALOG.md`
4. **Test**: Add test cases for the new error
5. **Update**: Update module documentation

## Error Code Registry

The complete error code registry is maintained in:

- Code: `pkg/errors/codes.go`
- Documentation: `docs/errors/ERROR_CATALOG.md`

## References

- Error Handling Best Practices: `_docs/ERROR_HANDLING.md`
- Client Error Handling Guide: `docs/api/ERROR_HANDLING.md`
- Error Type System: `pkg/errors/types.go`
- Complete Error Catalog: `docs/errors/ERROR_CATALOG.md`

## Policy Enforcement

This policy is enforced through:

1. **Code review**: All PRs adding errors must follow this policy
2. **Validation**: Use `errors.ValidateCode()` in tests
3. **Documentation**: Keep ERROR_CATALOG.md up to date
4. **CI checks**: Automated validation of error code ranges (future)

## Questions?

For questions about error code allocation or usage, refer to:

- Developer documentation: `_docs/ERROR_HANDLING.md`
- Open an issue with the `error-handling` label
- Ask in the #dev-errors Slack channel
