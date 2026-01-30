# VirtEngine API Documentation

Welcome to the VirtEngine API documentation. This comprehensive guide covers all aspects of interacting with the VirtEngine blockchain platform.

## Quick Links

| Guide | Description |
|-------|-------------|
| [Getting Started](./guides/getting-started.md) | Quick start guide for new developers |
| [Authentication](./guides/authentication.md) | API keys, signatures, and MFA |
| [API Versioning](./guides/versioning.md) | Version strategy and migration |
| [Rate Limits](../RATELIMIT_CLIENT_GUIDE.md) | Rate limiting policies |
| [Error Handling](./ERROR_HANDLING.md) | Error codes and handling |

## API Reference

### Blockchain Modules

| Module | Description | Reference |
|--------|-------------|-----------|
| **VEID** | Identity verification and management | [Reference](./reference/veid.md) |
| **MFA** | Multi-factor authentication | [Reference](./reference/mfa.md) |
| **Encryption** | Public-key encryption services | [Reference](./reference/encryption.md) |
| **Market** | Marketplace orders, bids, and leases | [Reference](./reference/market.md) |
| **Deployment** | Deployment management | [Reference](./reference/deployment.md) |
| **Provider** | Provider registration and management | [Reference](./reference/provider.md) |
| **Escrow** | Payment escrow services | [Reference](./reference/escrow.md) |
| **Cert** | Certificate management | [Reference](./reference/cert.md) |
| **Audit** | Audit logging and queries | [Reference](./reference/audit.md) |
| **Roles** | Role-based access control | [Reference](./reference/roles.md) |
| **Staking** | Staking and delegation | [Reference](./reference/staking.md) |
| **HPC** | High-performance computing | [Reference](./reference/hpc.md) |
| **Enclave** | TEE/Enclave management | [Reference](./reference/enclave.md) |

### Off-Chain Services

| Service | Description | Reference |
|---------|-------------|-----------|
| **Provider Daemon** | Provider-side services | [Reference](./reference/provider-daemon.md) |
| **Inventory** | Provider inventory management | [Reference](./reference/inventory.md) |
| **Lease Service** | Lease management for providers | [Reference](./reference/lease-service.md) |

## Code Examples

- [Go Examples](./examples/go/) - Go SDK and CLI examples
- [TypeScript Examples](./examples/typescript/) - TypeScript/JavaScript examples
- [cURL Examples](./examples/curl.md) - Command-line examples
- [Python Examples](./examples/python/) - Python SDK examples

## Interactive Documentation

- **Swagger UI**: Available at `/portal` when running the API gateway
- **Redoc**: See [redoc.html](./openapi/redoc.html) for rendered documentation

## Base URLs

| Environment | URL | Description |
|-------------|-----|-------------|
| Mainnet | `https://api.virtengine.com` | Production environment |
| Testnet | `https://api.testnet.virtengine.com` | Test environment |
| Local | `http://localhost:1317` | Local development |

## Transport Protocols

VirtEngine APIs support multiple transport protocols:

### REST/HTTP (gRPC-Gateway)

All gRPC services are exposed via REST endpoints through gRPC-Gateway.

```bash
curl https://api.virtengine.com/virtengine/market/v2beta1/orders/list
```

### gRPC

Native gRPC for high-performance clients:

```bash
grpcurl -plaintext localhost:9090 virtengine.market.v2beta1.Query/Orders
```

### WebSocket

Real-time subscriptions via CometBFT WebSocket:

```javascript
const ws = new WebSocket('wss://api.virtengine.com/websocket');
ws.send(JSON.stringify({
  jsonrpc: '2.0',
  method: 'subscribe',
  params: { query: "tm.event='NewBlock'" },
  id: 1
}));
```

## SDK Support

| Language | Package | Status |
|----------|---------|--------|
| Go | `github.com/virtengine/virtengine/sdk/go` | Stable |
| TypeScript | `@virtengine/sdk` | Stable |
| Python | `virtengine-sdk` | Beta |
| Rust | `virtengine-rs` | Alpha |

## Support

- **Documentation Issues**: [GitHub Issues](https://github.com/virtengine/virtengine/issues)
- **Discord**: [discord.gg/virtengine](https://discord.gg/virtengine)
- **Email**: [support@virtengine.com](mailto:support@virtengine.com)

---

*Last updated: January 2026*
