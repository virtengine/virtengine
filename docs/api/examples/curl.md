# VirtEngine API cURL Examples

This document provides cURL examples for common API operations.

## Prerequisites

```bash
# Set base URL
export VIRTENGINE_API="https://api.virtengine.com"

# Set API key (optional, for higher rate limits)
export VIRTENGINE_API_KEY="your_api_key"
```

---

## Node Information

### Get Node Info

```bash
curl "$VIRTENGINE_API/cosmos/base/tendermint/v1beta1/node_info"
```

### Get Syncing Status

```bash
curl "$VIRTENGINE_API/cosmos/base/tendermint/v1beta1/syncing"
```

### Get Latest Block

```bash
curl "$VIRTENGINE_API/cosmos/base/tendermint/v1beta1/blocks/latest"
```

---

## Account Queries

### Get Account Balance

```bash
curl "$VIRTENGINE_API/cosmos/bank/v1beta1/balances/{address}"
```

### Get Account Info

```bash
curl "$VIRTENGINE_API/cosmos/auth/v1beta1/accounts/{address}"
```

---

## VEID Module

### Query Identity

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/identity/{address}"
```

### Query Identity Score

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/score/{address}"
```

### Query Scopes

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/scopes/{address}"
```

### Query Specific Scope

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/scope/{address}/{scope_id}"
```

### Query Identity Wallet

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/wallet/{address}"
```

### Query Consent Settings

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/consent/{address}"
```

### Query Verification History

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/history/{address}"
```

### Query Module Parameters

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/params"
```

### Query Active Models

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/models"
```

### Query Compliance Status

```bash
curl "$VIRTENGINE_API/virtengine/veid/v1/compliance/{address}"
```

---

## MFA Module

### Query MFA Policy

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/policy/{address}"
```

### Query Factor Enrollments

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/enrollments/{address}"
```

### Query Pending Challenges

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/challenges/{address}"
```

### Check if MFA Required

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/required/{address}/{transaction_type}"
```

### Query Sensitive Transaction Configs

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/sensitive_tx"
```

### Query Trusted Devices

```bash
curl "$VIRTENGINE_API/virtengine/mfa/v1/devices/{address}"
```

---

## Market Module

### List Orders

```bash
# All orders
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list"

# Open orders only
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?filters.state=open"

# Orders by owner
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?filters.owner={owner_address}"

# With pagination
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.limit=10"
```

### Get Order Details

```bash
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/info?id.owner={owner}&id.dseq={dseq}&id.gseq={gseq}&id.oseq={oseq}"
```

### List Bids

```bash
# All bids for an order
curl "$VIRTENGINE_API/virtengine/market/v2beta1/bids/list?filters.owner={owner}&filters.dseq={dseq}"

# Bids by provider
curl "$VIRTENGINE_API/virtengine/market/v2beta1/bids/list?filters.provider={provider_address}"
```

### Get Bid Details

```bash
curl "$VIRTENGINE_API/virtengine/market/v2beta1/bids/info?id.owner={owner}&id.dseq={dseq}&id.gseq={gseq}&id.oseq={oseq}&id.provider={provider}"
```

### List Leases

```bash
# Active leases
curl "$VIRTENGINE_API/virtengine/market/v2beta1/leases/list?filters.state=active"

# Leases by owner
curl "$VIRTENGINE_API/virtengine/market/v2beta1/leases/list?filters.owner={owner_address}"

# Leases by provider
curl "$VIRTENGINE_API/virtengine/market/v2beta1/leases/list?filters.provider={provider_address}"
```

### Get Lease Details

```bash
curl "$VIRTENGINE_API/virtengine/market/v2beta1/leases/info?id.owner={owner}&id.dseq={dseq}&id.gseq={gseq}&id.oseq={oseq}&id.provider={provider}"
```

### Get Market Parameters

```bash
curl "$VIRTENGINE_API/virtengine/market/v2beta1/params"
```

---

## Provider Module

### List Providers

```bash
curl "$VIRTENGINE_API/virtengine/provider/v1beta4/providers/list"
```

### Get Provider Details

```bash
curl "$VIRTENGINE_API/virtengine/provider/v1beta4/providers/info?owner={provider_address}"
```

---

## Deployment Module

### List Deployments

```bash
# By owner
curl "$VIRTENGINE_API/virtengine/deployment/v1beta5/deployments/list?filters.owner={owner_address}"

# Active deployments
curl "$VIRTENGINE_API/virtengine/deployment/v1beta5/deployments/list?filters.state=active"
```

### Get Deployment Details

```bash
curl "$VIRTENGINE_API/virtengine/deployment/v1beta5/deployments/info?id.owner={owner}&id.dseq={dseq}"
```

---

## Escrow Module

### List Accounts

```bash
curl "$VIRTENGINE_API/virtengine/escrow/v1/accounts/list?scope=deployment&owner={owner_address}"
```

### Get Account Balance

```bash
curl "$VIRTENGINE_API/virtengine/escrow/v1/accounts/info?id.scope=deployment&id.xid={account_xid}"
```

### List Payments

```bash
curl "$VIRTENGINE_API/virtengine/escrow/v1/payments/list?scope=deployment&owner={owner_address}"
```

---

## Certificate Module

### List Certificates

```bash
curl "$VIRTENGINE_API/virtengine/cert/v1/certificates/list?filter.owner={owner_address}"
```

### Get Certificate

```bash
curl "$VIRTENGINE_API/virtengine/cert/v1/certificates/info?id.owner={owner}&id.serial={serial}"
```

---

## Audit Module

### List Audited Providers

```bash
curl "$VIRTENGINE_API/virtengine/audit/v1/audit/attributes/list"
```

### Get Provider Attributes

```bash
curl "$VIRTENGINE_API/virtengine/audit/v1/audit/attributes/info?auditor={auditor}&owner={owner}"
```

---

## Staking Module

### Get Validators

```bash
curl "$VIRTENGINE_API/cosmos/staking/v1beta1/validators"
```

### Get Delegations

```bash
curl "$VIRTENGINE_API/cosmos/staking/v1beta1/delegations/{delegator_address}"
```

---

## With API Key Authentication

```bash
# Include API key in header
curl -H "x-api-key: $VIRTENGINE_API_KEY" \
  "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list"
```

---

## Pagination Examples

### Keyset Pagination

```bash
# First page
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.limit=10"

# Get next_key from response, then:
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.limit=10&pagination.key={next_key}"
```

### Offset Pagination

```bash
# First page
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.limit=10&pagination.offset=0"

# Second page
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.limit=10&pagination.offset=10"
```

### Get Total Count

```bash
curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list?pagination.count_total=true"
```

---

## Error Handling

### Check for errors in response

```bash
# Successful response
{
  "orders": [...],
  "pagination": {...}
}

# Error response
{
  "error": {
    "code": "market:1401",
    "message": "order not found",
    "category": "not_found"
  }
}
```

### Handle rate limiting

```bash
# Check rate limit headers
curl -I "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list"

# Response headers:
# X-RateLimit-Limit: 300
# X-RateLimit-Remaining: 299
# X-RateLimit-Reset: 1704067260
```

---

## WebSocket Subscriptions

### Subscribe to New Blocks

```bash
# Using websocat
websocat wss://api.virtengine.com/websocket

# Send subscription
{"jsonrpc":"2.0","method":"subscribe","params":{"query":"tm.event='NewBlock'"},"id":1}
```

### Subscribe to Transactions

```bash
# Subscribe to transactions for a specific address
{"jsonrpc":"2.0","method":"subscribe","params":{"query":"tm.event='Tx' AND message.sender='virtengine1abc...'"},"id":1}
```

### Subscribe to Market Events

```bash
# Subscribe to new orders
{"jsonrpc":"2.0","method":"subscribe","params":{"query":"tm.event='Tx' AND message.action='CreateOrder'"},"id":1}
```

---

## gRPC via grpcurl

### Install grpcurl

```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Query Orders

```bash
grpcurl -plaintext api.virtengine.com:9090 virtengine.market.v2beta1.Query/Orders
```

### Query Identity

```bash
grpcurl -plaintext -d '{"account_address":"virtengine1abc..."}' \
  api.virtengine.com:9090 virtengine.veid.v1.Query/Identity
```

### List Services

```bash
grpcurl -plaintext api.virtengine.com:9090 list
```

---

## Tips

1. **Use jq for JSON formatting:**
   ```bash
   curl "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list" | jq .
   ```

2. **Save common queries as aliases:**
   ```bash
   alias virt-orders='curl -s "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list" | jq .'
   ```

3. **Debug with verbose output:**
   ```bash
   curl -v "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list"
   ```

4. **Measure response time:**
   ```bash
   curl -w "%{time_total}s\n" -o /dev/null -s "$VIRTENGINE_API/virtengine/market/v2beta1/orders/list"
   ```
