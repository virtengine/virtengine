# VirtEngine API Compatibility

This document defines the API versioning strategy, compatibility matrix, deprecation policy, and version support lifecycle for VirtEngine.

## Table of Contents

- [Version Support Lifecycle](#version-support-lifecycle)
- [API Compatibility Matrix](#api-compatibility-matrix)
- [Protocol Version Negotiation](#protocol-version-negotiation)
- [Deprecation Policy](#deprecation-policy)
- [Breaking Change Guidelines](#breaking-change-guidelines)
- [Client Compatibility](#client-compatibility)

## Version Support Lifecycle

VirtEngine follows a **N-2 support policy**, meaning we support the current major/minor version plus two previous minor versions.

### Support Levels

| Level | Description | Duration |
|-------|-------------|----------|
| **Active** | Full support, new features, bug fixes | Current release |
| **Maintenance** | Security fixes and critical bugs only | N-1 and N-2 releases |
| **Deprecated** | No support, migration strongly encouraged | Beyond N-2 |
| **End of Life** | Not supported, may be removed | 6 months after deprecation |

### Current Version Support Matrix

| Version | Status | Release Date | EOL Date | Notes |
|---------|--------|--------------|----------|-------|
| v0.11.x | Active | TBD | TBD | Current development |
| v0.10.x | Maintenance | TBD | TBD | Latest mainnet |
| v0.9.x | Maintenance | TBD | TBD | Previous testnet |
| v0.8.x | Deprecated | TBD | TBD | Previous mainnet |
| < v0.8 | EOL | - | - | Not supported |

### Version Numbering Convention

- **Odd minor versions** (0.9.x, 0.11.x): Testnet releases
- **Even minor versions** (0.8.x, 0.10.x): Mainnet releases
- **Patch versions**: Bug fixes and security patches (backwards compatible)

## API Compatibility Matrix

### Module API Versions

| Module | Current Version | Supported Versions | Next Version |
|--------|-----------------|-------------------|--------------|
| `x/veid` | v1 | v1 | v2 (planned) |
| `x/mfa` | v1 | v1 | - |
| `x/encryption` | v1 | v1 | v2 (PQ crypto) |
| `x/market` | v1beta5 | v1beta4, v1beta5 | v1 |
| `x/deployment` | v1beta4 | v1beta3, v1beta4 | v1 |
| `x/provider` | v1beta4 | v1beta3, v1beta4 | v1 |
| `x/escrow` | v1 | v1 | - |
| `x/audit` | v1 | v1 | - |
| `x/cert` | v1 | v1 | - |
| `x/hpc` | v1 | v1 | - |

### gRPC API Compatibility

| Service | Version | Status | Successor |
|---------|---------|--------|-----------|
| `virtengine.veid.v1.Query` | v1 | Active | - |
| `virtengine.veid.v1.Msg` | v1 | Active | - |
| `virtengine.market.v1beta5.Query` | v1beta5 | Active | v1 |
| `virtengine.market.v1beta4.Query` | v1beta4 | Deprecated | v1beta5 |
| `virtengine.deployment.v1beta4.Query` | v1beta4 | Active | v1 |
| `virtengine.provider.v1beta4.Query` | v1beta4 | Active | v1 |

### REST API Compatibility

| Endpoint Prefix | Version | Status | Successor |
|-----------------|---------|--------|-----------|
| `/virtengine/veid/v1/` | v1 | Active | - |
| `/virtengine/market/v1beta5/` | v1beta5 | Active | - |
| `/virtengine/market/v1beta4/` | v1beta4 | Deprecated | v1beta5 |

## Protocol Version Negotiation

VirtEngine uses protocol version negotiation for client-server communication to ensure compatibility.

### Protocol Versions

| Protocol | Current | Minimum Supported | Negotiation Header |
|----------|---------|-------------------|-------------------|
| Capture Protocol | v1 | v1 | `X-Capture-Protocol-Version` |
| Provider Protocol | v2 | v1 | `X-Provider-Protocol-Version` |
| Manifest Protocol | v2.1 | v2.0 | `X-Manifest-Version` |

### Version Negotiation Flow

1. **Client Request**: Client sends supported version range in headers
2. **Server Selection**: Server selects highest mutually supported version
3. **Response**: Server responds with selected version in headers
4. **Fallback**: If no compatible version, server returns `426 Upgrade Required`

### Example Header Exchange

```http
# Client Request
GET /api/v1/resource HTTP/1.1
X-API-Version: 1.0-2.0
X-Protocol-Version: 1,2

# Server Response
HTTP/1.1 200 OK
X-API-Version: 2.0
X-Protocol-Version: 2
```

## Deprecation Policy

### Deprecation Timeline

| Phase | Duration | Actions |
|-------|----------|---------|
| **Announcement** | - | Deprecation notice in changelog, docs, and API responses |
| **Warning Period** | 2 minor versions | Deprecation warnings in logs and responses |
| **Migration Period** | 2 minor versions | Feature still functional, warnings escalate |
| **Removal** | After migration period | Feature removed in next major/minor version |

### Deprecation Markers

#### Code Annotations

```go
// Deprecated: Use NewFunction instead. Will be removed in v0.12.0.
// See migration guide: docs/migrations/v0.10-v0.12.md
func OldFunction() {}
```

#### Protobuf Deprecation

```protobuf
message OldRequest {
  option deprecated = true;
  // Use NewRequest instead. Removal planned for v0.12.0.
}
```

#### API Response Headers

```http
Deprecation: true
Sunset: Sat, 01 Jun 2025 00:00:00 GMT
Link: <https://docs.virtengine.io/migrations/v0.12>; rel="successor-version"
```

### Deprecation Enforcement

1. **Compile-time**: Go `deprecated` comments trigger IDE/linter warnings
2. **Runtime**: Deprecated API calls emit warning logs
3. **CI/CD**: Deprecation tests fail if removed features are still referenced
4. **Metrics**: Track deprecated API usage via telemetry

## Breaking Change Guidelines

### Definition of Breaking Changes

A breaking change is any modification that could cause existing clients to fail:

- Removing or renaming API endpoints, methods, or fields
- Changing field types or semantics
- Removing enum values
- Changing default values with behavioral impact
- Modifying error codes or response formats
- Changing authentication/authorization requirements

### Non-Breaking Changes (Safe to make)

- Adding new optional fields (with defaults)
- Adding new API endpoints
- Adding new enum values
- Extending error information (while preserving error codes)
- Performance improvements
- Documentation updates

### Breaking Change Process

1. **Proposal**: Submit breaking change proposal via GitHub issue
2. **Review**: Architecture review and impact assessment
3. **Approval**: Requires maintainer approval
4. **Migration Guide**: Create migration documentation
5. **Implementation**: Implement with proper deprecation markers
6. **Testing**: Add compatibility tests for migration path
7. **Communication**: Announce in changelog and release notes

### Migration Guide Requirements

Every breaking change must include a migration guide with:

- Clear description of what changed
- Impact assessment
- Step-by-step migration instructions
- Code examples (before/after)
- Timeline and support commitments

See [MIGRATION_GUIDE_TEMPLATE.md](./MIGRATION_GUIDE_TEMPLATE.md) for the template.

## Client Compatibility

### Supported Client Versions

| Client | Min Version | Recommended | Notes |
|--------|-------------|-------------|-------|
| `virtengine` CLI | v0.8.0 | Latest | Primary CLI tool |
| `provider-daemon` | v0.8.0 | Latest | Provider software |
| Go SDK | v0.8.0 | Latest | `github.com/virtengine/go` |
| TypeScript SDK | v0.8.0 | Latest | `@virtengine/sdk` |

### Client Upgrade Guidance

1. **Check Compatibility**: Verify client version against API version matrix
2. **Review Changelog**: Check for breaking changes in target version
3. **Test in Staging**: Deploy to testnet first
4. **Gradual Rollout**: Use canary deployment for production
5. **Monitor**: Watch for deprecation warnings and errors

### Version Detection

Clients can detect server version via:

```bash
# CLI version check
virtengine version

# Server version check
virtengine query version

# API version check
curl -I https://rpc.virtengine.io/version
```

## Testing Requirements

### Compatibility Test Coverage

All API changes must include:

1. **Forward Compatibility Tests**: New server with old client requests
2. **Backward Compatibility Tests**: Old server with new client requests
3. **Migration Tests**: Data migration between versions
4. **Deprecation Tests**: Verify deprecation warnings work correctly

### Running Compatibility Tests

```bash
# Run all compatibility tests
make test-compatibility

# Run specific version compatibility
go test -tags="e2e.compatibility" ./tests/compatibility/... -run TestAPIv1

# Run deprecation enforcement
go test ./tests/compatibility/... -run TestDeprecation
```

## Monitoring and Telemetry

### Deprecation Metrics

| Metric | Description |
|--------|-------------|
| `api_deprecated_calls_total` | Total deprecated API calls |
| `api_version_requests` | Requests by API version |
| `protocol_version_negotiation_failures` | Version negotiation failures |
| `client_version_distribution` | Distribution of client versions |

### Alerts

- Alert when deprecated API usage exceeds threshold
- Alert when version negotiation failure rate spikes
- Alert on unsupported client version connections

## Related Documentation

- [MIGRATION_GUIDE_TEMPLATE.md](./MIGRATION_GUIDE_TEMPLATE.md) - Template for migration guides
- [RELEASE.md](../RELEASE.md) - Release management process
- [version-control.md](../_docs/version-control.md) - Branching and versioning
- [CHANGELOG.md](../CHANGELOG.md) - Version changelog
