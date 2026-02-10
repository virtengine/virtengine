# VirtEngine Database Index Strategy

This document describes the indexing strategy for VirtEngine's Cosmos SDK-based store layer, including key prefixes, secondary indexes, and query optimization patterns.

## Overview

VirtEngine uses the Cosmos SDK KVStore for persistent state storage. All data is stored as key-value pairs with carefully designed key prefixes that enable efficient iteration and lookups.

## Key Design Principles

### 1. State-Based Partitioning
Objects are stored in separate key spaces based on their state (e.g., open, active, closed). This enables:
- Efficient queries filtered by state
- Reduced iteration overhead when querying specific states
- Natural data lifecycle management

### 2. Hierarchical Key Structure
Keys follow a hierarchical structure: `Module/State/Owner/Sequences...`
- Enables prefix-based iteration for related objects
- Supports filtering at multiple levels
- Maintains natural ordering for range queries

### 3. Secondary Indexes
For queries that don't align with primary key structure, secondary indexes are maintained:
- **Reverse indexes**: Enable queries in opposite direction (e.g., provider → leases)
- **State indexes**: Enable state-based filtering without full scan

## Module-Specific Index Structures

### Market Module

#### Orders
```
Primary:   0x11/0x00/{state}/{owner}/{dseq}/{gseq}/{oseq}
```

| State Prefix | Value | Description |
|--------------|-------|-------------|
| 0x01 | Open | Order awaiting bids |
| 0x02 | Active | Order matched with lease |
| 0x03 | Closed | Order completed or cancelled |

#### Bids
```
Primary:   0x12/0x00/{state}/{owner}/{dseq}/{gseq}/{oseq}/{provider}/{bseq}
Reverse:   0x12/0x01/{state}/{provider}/{bseq}/{dseq}/{gseq}/{oseq}/{owner}
```

The reverse index enables efficient queries like "all bids by provider X".

#### Leases
```
Primary:   0x13/0x00/{state}/{owner}/{dseq}/{gseq}/{oseq}/{provider}/{bseq}
Reverse:   0x13/0x01/{state}/{provider}/{bseq}/{dseq}/{gseq}/{oseq}/{owner}
```

The reverse index enables:
- O(1) `ProviderHasActiveLeases()` check
- Efficient "leases by provider" queries

### Escrow Module

#### Accounts
```
Primary:   account/{state}/{xid}
```

States: Open, Closed, Overdrawn

#### Payments
```
Primary:   payment/{state}/{account_id}/{payment_id}
```

### Deployment Module

#### Deployments
```
Primary:   deployment/{state}/{owner}/{dseq}
```

#### Groups
```
Primary:   group/{state}/{owner}/{dseq}/{gseq}
```

States: Open, Paused, InsufficientFunds, Closed

### VEID Module

#### Identity Records
```
Primary:   identity/{address}
```

#### Scopes
```
Primary:   scope/{address}/{scope_id}
```

#### Wallets
```
Primary:   wallet/{address}
```

#### Appeals
```
Primary:   appeal/{appeal_id}
By Account: appeal/account/{address}/{appeal_id}
```

## Query Optimization Patterns

### 1. Batch Queries
For operations that need multiple related objects, use batch methods:

```go
// Instead of:
for _, id := range deploymentIDs {
    groups := k.GetGroups(ctx, id)  // N store accesses
}

// Use:
groupsMap := k.GetGroupsBatch(ctx, deploymentIDs)  // Single batch operation
```

Available batch methods:
- `GetAccountsBatch()` - Fetch multiple escrow accounts
- `GetPaymentsBatch()` - Fetch multiple payments
- `GetGroupsBatch()` - Fetch groups for multiple deployments

### 2. Pagination
All list queries should use Cosmos SDK pagination:

```go
pageRes, err := sdkquery.FilteredPaginate(store, req.Pagination, func(...) (bool, error) {
    // Process each item
})
```

Key pagination patterns:
- Always set default and max limits
- Use cursor-based pagination (via NextKey) for efficiency
- Avoid offset-based pagination on large datasets

### 3. Secondary Index Usage
Use secondary indexes for reverse lookups:

```go
// Efficient: Uses provider reverse index
prefix := keys.LeasesByProviderPrefix(keys.LeaseStateActivePrefix, provider)
iter := storetypes.KVStorePrefixIterator(store, prefix)

// Inefficient: Scans all leases
k.WithLeases(ctx, func(lease Lease) bool {
    if lease.ID.Provider == provider { ... }
})
```

### 4. State-Based Filtering
Leverage state-based key partitioning:

```go
// Efficient: Only iterates open orders
iter := storetypes.KVStorePrefixIterator(store, keys.OrderStateOpenPrefix)

// Less efficient: Iterates all orders, filters in memory
k.WithOrders(ctx, func(order Order) bool {
    if order.State == OrderOpen { ... }
})
```

## Performance Considerations

### Hot Paths
The following queries are on critical paths and have been optimized:

1. **Order/Bid/Lease lookups by ID** - Direct key access, O(1)
2. **Provider active lease check** - Uses reverse index, O(1)
3. **Deployment list with groups** - Uses batch fetching
4. **Bids for order** - Uses order-prefixed iteration

### Benchmark Results
Run benchmarks with:
```bash
go test -bench=. ./x/market/keeper/...
```

Key benchmarks:
- `BenchmarkProviderHasActiveLeases` - Should be O(1) with index
- `BenchmarkGetOrder` - Direct key lookup
- `BenchmarkWithBidsForOrder` - Prefix-based iteration
- `BenchmarkBidCountForOrder` - Multiple prefix iterations

## Adding New Indexes

When adding new secondary indexes:

1. **Define key builder function** in the appropriate `keys.go` file
2. **Maintain index on mutations** - Update in Create/Update/Delete operations
3. **Add migration** if retrofitting existing data
4. **Document in this file**

Example:
```go
// In keys/key.go
func LeasesByProviderPrefix(statePrefix []byte, provider sdk.AccAddress) []byte {
    buf := bytes.NewBuffer(LeasePrefixReverse)
    buf.Write(statePrefix)
    buf.Write(address.MustLengthPrefix(provider))
    return buf.Bytes()
}
```

## Common Anti-Patterns to Avoid

### ❌ N+1 Queries in Loops
```go
// Bad: N+1 query pattern
for _, deployment := range deployments {
    account := k.ekeeper.GetAccount(ctx, deployment.ID)  // Called N times
    groups := k.GetGroups(ctx, deployment.ID)            // Called N times
}
```

### ❌ Full Table Scans for Existence Checks
```go
// Bad: Scans all items to check existence
k.WithItems(ctx, func(item Item) bool {
    if item.SomeField == value {
        found = true
        return true
    }
    return false
})
```

### ❌ In-Memory Filtering Without Limit
```go
// Bad: Loads all data then filters in memory
all := k.GetAll(ctx)
filtered := filterItems(all, criteria)  // Memory-intensive
```

## Future Improvements

1. **Composite indexes** for complex queries
2. **Materialized views** for frequently-accessed aggregations
3. **Index-only scans** where value isn't needed
4. **Bloom filters** for existence checks on large datasets
