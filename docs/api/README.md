# VirtEngine API Documentation

Welcome to the VirtEngine API documentation. This comprehensive guide covers all aspects of interacting with the VirtEngine blockchain platform.

## 🚀 Quick Links

| Guide | Description |
|-------|-------------|
| [Getting Started](./guides/getting-started.md) | Quick start guide for new developers |
| [Authentication](./guides/authentication.md) | API keys, wallet signatures, and MFA |
| [API Versioning](./guides/versioning.md) | Version strategy and migration |
| [Rate Limits & Quotas](./guides/rate-limits.md) | Rate limiting policies and quotas |
| [Error Handling](./ERROR_HANDLING.md) | Error codes and handling strategies |

## 📚 API Reference

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
| **Take** | Network take rate configuration | [Reference](./reference/take.md) |
| **Oracle** | Price oracle for resources | [Reference](./reference/oracle.md) |
| **Epochs** | Epoch management | [Reference](./reference/epochs.md) |
| **Fraud** | Fraud detection and reporting | [Reference](./reference/fraud.md) |

### Off-Chain Services

| Service | Description | Reference |
|---------|-------------|-----------|
| **Provider Daemon** | Provider-side services | [Reference](./reference/provider-daemon.md) |
| **Inventory** | Provider inventory management | [Reference](./reference/inventory.md) |
| **Lease Service** | Lease management for providers | [Reference](./reference/lease-service.md) |

## 💻 Code Examples

- [Go Examples](./examples/go/) - Go SDK and CLI examples
- [TypeScript Examples](./examples/typescript/) - TypeScript/JavaScript examples
- [cURL Examples](./examples/curl.md) - Command-line examples
- [Python Examples](./examples/python/) - Python SDK examples

## 📖 Interactive Documentation

### OpenAPI/Swagger Documentation

- **Main API Spec**: [virtengine-api.yaml](../../api/openapi/virtengine-api.yaml)
- **Provider Portal Spec**: [portal_api.yaml](../../api/openapi/portal_api.yaml)
- **Swagger UI**: [swagger-ui.html](./openapi/swagger-ui.html) - Interactive API explorer
- **Redoc**: [redoc.html](./openapi/redoc.html) - Beautiful rendered documentation

### Accessing Interactive Docs

**Locally:**
```bash
# Serve static files
npx http-server docs/api/openapi -p 8080

# Open in browser
open http://localhost:8080/swagger-ui.html
open http://localhost:8080/redoc.html
```

**Via API Gateway:**
```bash
# Start Kong gateway with Swagger UI
docker-compose up -d kong swagger-ui

# Access at
http://localhost:8000/portal
```

## 🌐 Base URLs

| Environment | URL | Description |
|-------------|-----|-------------|
| Mainnet | `https://api.virtengine.com` | Production environment |
| Testnet | `https://api.testnet.virtengine.com` | Test environment |
| Local | `http://localhost:1317` | Local development |

## 🔌 Transport Protocols

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

## 🔐 Authentication

VirtEngine supports multiple authentication methods:

| Method | Use Case | Rate Limit Tier |
|--------|----------|-----------------|
| Anonymous (IP) | Public queries | Basic (10 req/s) |
| API Key | Applications, integrations | Standard (50 req/s) |
| Wallet Signature | User transactions | Standard (50 req/s) |
| MFA-Verified | Sensitive operations | Standard (50 req/s) |
| Provider Certificate | Provider operations | Provider (500 req/s) |

See [Authentication Guide](./guides/authentication.md) for detailed information.

## ⚡ Rate Limits

| Tier | Requests/sec | Daily Quota | Burst |
|------|--------------|-------------|-------|
| Anonymous | 10 | 10,000 | 20 |
| Standard (API Key) | 50 | 100,000 | 100 |
| Premium | 200 | 1,000,000 | 500 |
| Provider | 500 | Unlimited | 1,000 |

See [Rate Limits & Quotas Guide](./guides/rate-limits.md) for details.

## 📦 SDK Support

| Language | Package | Status | Documentation |
|----------|---------|--------|---------------|
| Go | `github.com/virtengine/virtengine/sdk/go` | Stable | [Docs](./examples/go/) |
| TypeScript | `@virtengine/sdk` | Stable | [Docs](./examples/typescript/) |
| Python | `virtengine-sdk` | Beta | [Docs](./examples/python/) |
| Rust | `virtengine-rs` | Alpha | Coming soon |

## 🛠️ Development Tools

### Generate Client from OpenAPI

```bash
# Install OpenAPI Generator
npm install -g @openapitools/openapi-generator-cli

# Generate TypeScript client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g typescript-fetch \
  -o generated/typescript

# Generate Go client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g go \
  -o generated/go

# Generate Python client
openapi-generator-cli generate \
  -i api/openapi/virtengine-api.yaml \
  -g python \
  -o generated/python
```

### Validate OpenAPI Spec

```bash
# Using Redocly
npx @redocly/openapi-cli lint api/openapi/virtengine-api.yaml

# Using Swagger
npx swagger-cli validate api/openapi/virtengine-api.yaml
```

## 🔍 Common Operations

### Query Market Orders

```bash
# REST
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.state=open"

# gRPC
grpcurl -d '{"filters":{"state":"open"}}' \
  api.virtengine.com:9090 \
  virtengine.market.v2beta1.Query/Orders
```

### Query Identity

```bash
curl "https://api.virtengine.com/virtengine/veid/v1/identity/virtengine1..."
```

### List Providers

```bash
curl "https://api.virtengine.com/virtengine/provider/v1beta4/providers"
```

## 📊 API Versioning Strategy

VirtEngine uses semantic versioning with stability indicators:

- **v1** - Stable, production-ready
- **v1beta1**, **v1beta2** - Beta, may change
- **v2beta1** - Beta for next major version

See [Versioning Guide](./guides/versioning.md) for migration information.

## 🚨 Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "veid:1001",
    "message": "identity not found",
    "category": "not_found",
    "context": {
      "account_address": "virtengine1..."
    }
  }
}
```

See [Error Handling Guide](./ERROR_HANDLING.md) for all error codes.

## 📈 Monitoring & Observability

### Health Check

```bash
curl https://api.virtengine.com/health
```

### Prometheus Metrics

```bash
curl https://api.virtengine.com/metrics
```

### Rate Limit Status

```bash
curl -H "x-api-key: YOUR_KEY" \
  https://api.virtengine.com/quota/status
```

## 🤝 Support

- **Documentation Issues**: [GitHub Issues](https://github.com/virtengine/virtengine/issues)
- **Discord**: [discord.gg/virtengine](https://discord.gg/virtengine)
- **Email**: [support@virtengine.com](mailto:support@virtengine.com)
- **Enterprise Support**: [enterprise@virtengine.com](mailto:enterprise@virtengine.com)

## 📝 Contributing

Found an error in the documentation? Please submit a PR!

1. Fork the repository
2. Make your changes
3. Submit a pull request

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## 🔗 Related Resources

- [Developer Guide](../../_docs/developer-guide.md)
- [Testing Guide](../../_docs/testing-guide.md)
- [Provider Guide](../../_docs/provider-guide.md)
- [Validator Onboarding](../../_docs/validator-onboarding.md)

---

**Last updated**: February 2026

**API Version**: v1.0.0
