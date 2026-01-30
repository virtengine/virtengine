# VirtEngine Infrastructure Tests

This directory contains Terratest-based infrastructure tests for validating the Terraform modules.

## Prerequisites

- Go 1.22+
- AWS credentials configured
- Terraform 1.5+

## Running Tests

### Quick tests (networking only)
```bash
go test -v -timeout 30m -run TestNetworkingModule
```

### Full module tests
```bash
go test -v -timeout 2h
```

### Short mode (skip long-running tests)
```bash
go test -v -short
```

### Specific environment test
```bash
go test -v -timeout 2h -run TestFullStackIntegration
```

## Test Stages

The integration tests use test_structure to support staged execution:

```bash
# Run only setup stage
SKIP_deploy=true SKIP_validate=true SKIP_cleanup=true go test -v -run TestFullStackIntegration

# Run only validation (assumes infrastructure exists)
SKIP_setup=true SKIP_deploy=true SKIP_cleanup=true go test -v -run TestFullStackIntegration

# Skip cleanup for debugging
SKIP_cleanup=true go test -v -run TestFullStackIntegration
```

## Test Coverage

| Module | Test |
|--------|------|
| networking | VPC, subnets, NAT gateway, security groups |
| eks | Cluster creation, node groups, OIDC provider |
| rds | Instance, encryption, credentials |
| vault | (Requires K8s cluster - manual testing) |
| monitoring | (Requires K8s cluster - manual testing) |

## Cost Warning

These tests create real AWS resources. Ensure tests complete or resources are manually cleaned up to avoid unexpected charges.
