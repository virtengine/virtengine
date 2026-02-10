# VirtEngine Protobuf Migration Guide: Cosmos SDK v0.45.x to v0.53.4

This document outlines the protobuf definition changes required to migrate your VirtEngine protobuf files from version 0.45.x to 0.53.4. This guide focuses specifically on **protobuf schema changes** and complements the main API migration guide.

## Important Version Note

The VirtEngine protobuf definitions follow the same versioning as the main API:

- **v0.45.x**: Earlier VirtEngine protobuf schemas
- **v0.53.4**: Current VirtEngine protobuf schemas

## VirtEngine-Specific Protobuf Changes

### 1. Deployment Module Changes (v1beta3 → v1beta4)

#### MsgCreateDeployment Proto Definition Updates

**Key Changes:**
- `version` field renamed to `hash`
- `deposit` field type changed from `cosmos.base.v1beta1.Coin` to `virtengine.base.deposit.v1.Deposit`
- `depositor` field removed
- Import dependencies updated

**Before (v1beta3):**
```protobuf
// virtengine/deployment/v1beta3/deploymentmsg.proto
syntax = "proto3";
package virtengine.deployment.v1beta3;

import "gogoproto/gogo.proto";
import "virtengine/deployment/v1beta3/deployment.proto";
import "virtengine/deployment/v1beta3/groupspec.proto";
import "cosmos/base/v1beta1/coin.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta3";

message MsgCreateDeployment {
  option (gogoproto.equal) = false;

  DeploymentID id = 1 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID",
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];
  
  repeated GroupSpec groups = 2 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "groups",
    (gogoproto.moretags) = "yaml:\"groups\""
  ];
  
  bytes version = 3 [
    (gogoproto.jsontag)  = "version",
    (gogoproto.moretags) = "yaml:\"version\""
  ];
  
  cosmos.base.v1beta1.Coin deposit = 4 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "deposit",
    (gogoproto.moretags) = "yaml:\"deposit\""
  ];
  
  // Depositor pays for the deposit
  string depositor = 5 [
    (gogoproto.jsontag)  = "depositor",
    (gogoproto.moretags) = "yaml:\"depositor\""
  ];
}
```

**After (v1beta4):**
```protobuf
// virtengine/deployment/v1beta4/deploymentmsg.proto
syntax = "proto3";
package virtengine.deployment.v1beta4;

import "gogoproto/gogo.proto";
import "virtengine/deployment/v1/deployment.proto";
import "virtengine/deployment/v1beta4/groupspec.proto";
import "virtengine/base/deposit/v1/deposit.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4";

message MsgCreateDeployment {
  option (gogoproto.equal) = false;

  // ID is the unique identifier of the deployment.
  virtengine.deployment.v1.DeploymentID id = 1 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID",
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];

  // GroupSpec is a list of group specifications for the deployment.
  repeated GroupSpec groups = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "GroupSpecs",
    (gogoproto.jsontag)      = "groups",
    (gogoproto.moretags)     = "yaml:\"groups\""
  ];

  // Hash of the deployment (renamed from version).
  bytes hash = 3 [
    (gogoproto.jsontag)  = "hash",
    (gogoproto.moretags) = "yaml:\"hash\""
  ];

  // Deposit specifies the amount and source of deployment deposit.
  virtengine.base.deposit.v1.Deposit deposit = 4 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "deposit",
    (gogoproto.moretags) = "yaml:\"deposit\""
  ];
}
```

#### MsgUpdateDeployment Proto Definition Updates

**Before (v1beta3):**
```protobuf
// virtengine/deployment/v1beta3/deploymentmsg.proto
message MsgUpdateDeployment {
  option (gogoproto.equal) = false;

  DeploymentID id = 1 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID",
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];
  
  bytes version = 3 [
    (gogoproto.jsontag)  = "version",
    (gogoproto.moretags) = "yaml:\"version\""
  ];
}
```

**After (v1beta4):**
```protobuf
// virtengine/deployment/v1beta4/deploymentmsg.proto
message MsgUpdateDeployment {
  option (gogoproto.equal) = false;

  // ID is the unique identifier of the deployment.
  virtengine.deployment.v1.DeploymentID id = 1 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID",
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];

  // Hash of the deployment (renamed from version).
  bytes hash = 3 [
    (gogoproto.jsontag)  = "hash",
    (gogoproto.moretags) = "yaml:\"hash\""
  ];
}
```

#### New Deposit Proto Definition

The new deposit structure requires a separate proto file:

**New File: virtengine/base/deposit/v1/deposit.proto**
```protobuf
syntax = "proto3";
package virtengine.base.deposit.v1;

import "gogoproto/gogo.proto";
import "cosmos/base/v1beta1/coin.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/types/deposit/v1";

// Source is an enum which lists source of funds for deployment deposit.
enum Source {
  option (gogoproto.goproto_enum_prefix) = false;

  // Prefix should start with 0 in enum. So declaring dummy state.
  invalid = 0 [(gogoproto.enumvalue_customname) = "SourceInvalid"];
  // DepositSourceBalance denotes account balance as source of funds
  balance = 1 [(gogoproto.enumvalue_customname)  = "SourceBalance"];
  // DepositSourceGrant denotes authz grants as source of funds
  grant = 2 [(gogoproto.enumvalue_customname)  = "SourceGrant"];
}

// Deposit is a data type used by MsgCreateDeployment, MsgDepositDeployment and MsgCreateBid
message Deposit {
  // Amount specifies the coins to include in the deposit.
  cosmos.base.v1beta1.Coin amount = 1 [
    (gogoproto.nullable) = false,
    (gogoproto.moretags) = "yaml:\"amount\""
  ];

  // Sources list of deposit sources, each entry must be unique
  repeated Source sources = 5 [
    (gogoproto.castrepeated) = "Sources",
    (gogoproto.jsontag)      = "deposit_sources",
    (gogoproto.moretags)     = "yaml:\"deposit_sources\""
  ];
}
```

### 2. Market Module: v1beta4 and v1beta5 Coexistence

**Important**: Both v1beta4 and v1beta5 exist in the current codebase. v1beta5 introduces enhanced features while maintaining v1beta4 compatibility.

#### Bid Proto Definition Updates

**Key Changes:**
- `BidID` field renamed to `ID`
- ID field type changed to `virtengine.market.v1.BidID`
- Enhanced resource specification with `ResourceOffer` array
- Improved resource management

**Before (v1beta4):**
```protobuf
// virtengine/market/v1beta4/bid.proto (partial)
message Bid {
  option (gogoproto.equal)            = false;
  option (gogoproto.goproto_stringer) = false;

  BidID bid_id = 1 [
    (gogoproto.customname) = "BidID",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = 'yaml:"id"'
  ];

  State state = 2 [
    (gogoproto.jsontag)  = "state",
    (gogoproto.moretags) = 'yaml:"state"'
  ];

  cosmos.base.v1beta1.DecCoin price = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "price",
    (gogoproto.moretags) = 'yaml:"price"'
  ];
  
  int64 created_at = 4;

  repeated ResourceOffer resources_offer = 5 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "ResourcesOffer",
    (gogoproto.customname)   = "ResourcesOffer",
    (gogoproto.jsontag)      = "resources_offer",
    (gogoproto.moretags)     = 'yaml:"resources_offer"'
  ];
}
```

**After (v1beta5):**
```protobuf
// virtengine/market/v1beta5/bid.proto
syntax = "proto3";
package virtengine.market.v1beta5;

import "gogoproto/gogo.proto";
import "cosmos/base/v1beta1/coin.proto";
import "virtengine/market/v1beta5/resourcesoffer.proto";
import "virtengine/market/v1/bid.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5";

message Bid {
  option (gogoproto.equal)            = true;
  option (gogoproto.goproto_stringer) = false;

  enum State {
    option (gogoproto.goproto_enum_prefix) = false;

    invalid = 0 [(gogoproto.enumvalue_customname) = "BidStateInvalid"];
    open = 1 [(gogoproto.enumvalue_customname) = "BidOpen"];
    active = 2 [(gogoproto.enumvalue_customname) = "BidActive"];
    lost = 3 [(gogoproto.enumvalue_customname) = "BidLost"];
    closed = 4 [(gogoproto.enumvalue_customname) = "BidClosed"];
  }

  // BidID stores owner and all other seq numbers (renamed from bid_id to id)
  virtengine.market.v1.BidID id = 1 [
    (gogoproto.customname) = "ID",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];

  State state = 2 [
    (gogoproto.jsontag)  = "state",
    (gogoproto.moretags) = "yaml:\"state\""
  ];

  cosmos.base.v1beta1.DecCoin price = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "price",
    (gogoproto.moretags) = "yaml:\"price\""
  ];

  int64 created_at = 4 [
    (gogoproto.jsontag)  = "createdAt",
    (gogoproto.moretags) = "yaml:\"created_at\""
  ];

  // ResourceOffer is a list of offers (enhanced resource specification)
  repeated ResourceOffer resources_offer = 5 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "ResourcesOffer",
    (gogoproto.customname)   = "ResourcesOffer",
    (gogoproto.jsontag)      = "resources_offer",
    (gogoproto.moretags)     = "yaml:\"resources_offer\""
  ];
}
```

#### MsgCreateBid Proto Updates

**Before (v1beta4):**
```protobuf
// virtengine/market/v1beta4/bid.proto
message MsgCreateBid {
  option (gogoproto.equal) = false;

  OrderID order = 1 [
    (gogoproto.customname) = "Order",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "order",
    (gogoproto.moretags)   = 'yaml:"order"'
  ];
  
  string provider = 2 [
    (gogoproto.jsontag)  = "provider",
    (gogoproto.moretags) = 'yaml:"provider"'
  ];
  
  cosmos.base.v1beta1.DecCoin price = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "price",
    (gogoproto.moretags) = 'yaml:"price"'
  ];
  
  cosmos.base.v1beta1.Coin deposit = 4 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "deposit",
    (gogoproto.moretags) = 'yaml:"deposit"'
  ];
  
  repeated ResourceOffer resources_offer = 5 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "ResourcesOffer",
    (gogoproto.customname)   = "ResourcesOffer",
    (gogoproto.jsontag)      = "resources_offer",
    (gogoproto.moretags)     = 'yaml:"resources_offer"'
  ];
}
```

**After (v1beta5):**
```protobuf
// virtengine/market/v1beta5/bidmsg.proto
syntax = "proto3";
package virtengine.market.v1beta5;

import "gogoproto/gogo.proto";
import "cosmos/base/v1beta1/coin.proto";
import "virtengine/market/v1beta5/resourcesoffer.proto";
import "virtengine/market/v1/bid.proto";
import "virtengine/base/deposit/v1/deposit.proto";
import "virtengine/market/v1/types.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5";

message MsgCreateBid {
  option (gogoproto.equal) = false;

  virtengine.market.v1.BidID id = 1 [
    (gogoproto.customname) = "ID",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];

  cosmos.base.v1beta1.DecCoin price = 2 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "price",
    (gogoproto.moretags) = "yaml:\"price\""
  ];

  // Uses new deposit structure
  virtengine.base.deposit.v1.Deposit deposit = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "deposit",
    (gogoproto.moretags) = "yaml:\"deposit\""
  ];

  // Enhanced resource specification
  repeated ResourceOffer resources_offer = 4 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "ResourcesOffer",
    (gogoproto.customname)   = "ResourcesOffer",
    (gogoproto.jsontag)      = "resources_offer",
    (gogoproto.moretags)     = "yaml:\"resources_offer\""
  ];
}
```

#### MsgWithdrawLease Proto Updates

**Before (v1beta4):**
```protobuf
// virtengine/market/v1beta4/lease.proto
message MsgWithdrawLease {
  option (gogoproto.equal) = false;

  LeaseID bid_id = 1 [
    (gogoproto.customname) = "LeaseID",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = 'yaml:"id"'
  ];
}
```

**After (v1beta5):**
```protobuf
// virtengine/market/v1beta5/leasemsg.proto
message MsgWithdrawLease {
  option (gogoproto.equal) = false;

  // BidId is the unique identifier of the Bid.
  virtengine.market.v1.LeaseID id = 1 [
    (gogoproto.customname) = "ID",
    (gogoproto.nullable)   = false,
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];
}
```

### 3. Escrow Module Changes (v1)

```protobuf
// virtengine/escrow/v1/msg.proto
message MsgAccountDeposit {
  option (cosmos.msg.v1.signer) = "signer";
  
  // Signer is the account bech32 address of the user who wants to deposit into
  // an escrow account. Does not necessarily need to be an owner of the deployment.
  string signer = 1 [
    (cosmos_proto.scalar) = "cosmos.AddressString",
    (gogoproto.jsontag)   = "signer",
    (gogoproto.moretags)  = "yaml:\"signer\""
  ];
  
  virtengine.escrow.id.v1.Account id = 2 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID",
    (gogoproto.jsontag)    = "id",
    (gogoproto.moretags)   = "yaml:\"id\""
  ];
  
  virtengine.base.deposit.v1.Deposit deposit = 3 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "deposit",
    (gogoproto.moretags) = "yaml:\"deposit\""
  ];
}
```

### 4. Provider Module Changes (v1beta3 → v1beta4)

#### Provider Info Structure Update

**Key Changes:**
- `ProviderInfo` message renamed to `Info`
- Simplified structure with same fields
- Updated import paths

**Before (v1beta3):**
```protobuf
// virtengine/provider/v1beta3/provider.proto
syntax = "proto3";
package virtengine.provider.v1beta3;

import "gogoproto/gogo.proto";
import "virtengine/base/v1beta3/attribute.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta3";

// ProviderInfo
message ProviderInfo {
  string email = 1 [
    (gogoproto.customname) = "EMail",
    (gogoproto.jsontag)    = "email",
    (gogoproto.moretags)   = "yaml:\"email\""
  ];
  string website = 2 [
    (gogoproto.jsontag)    = "website",
    (gogoproto.moretags)   = "yaml:\"website\""
  ];
}

message MsgCreateProvider {
  option (gogoproto.equal) = false;

  string owner = 1;
  string host_uri = 2;
  repeated virtengine.base.v1beta3.Attribute attributes = 3;
  
  ProviderInfo info = 4 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "info",
    (gogoproto.moretags) = "yaml:\"info\""
  ];
}
```

**After (v1beta4):**
```protobuf
// virtengine/provider/v1beta4/provider.proto
syntax = "proto3";
package virtengine.provider.v1beta4;

import "gogoproto/gogo.proto";
import "cosmos_proto/cosmos.proto";
import "virtengine/base/attributes/v1/attribute.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4";

// Info (renamed from ProviderInfo)
message Info {
  string email = 1 [
    (gogoproto.customname) = "EMail",
    (gogoproto.jsontag)    = "email",
    (gogoproto.moretags)   = "yaml:\"email\""
  ];
  string website = 2 [
    (gogoproto.jsontag)    = "website",
    (gogoproto.moretags)   = "yaml:\"website\""
  ];
}

// virtengine/provider/v1beta4/msg.proto - MsgCreateProvider
message MsgCreateProvider {
  option (gogoproto.equal) = false;

  string owner = 1 [
    (cosmos_proto.scalar) = "cosmos.AddressString",
    (gogoproto.jsontag)   = "owner",
    (gogoproto.moretags)  = "yaml:\"owner\""
  ];
  
  string host_uri = 2 [
    (gogoproto.customname) = "HostURI",
    (gogoproto.jsontag)    = "host_uri",
    (gogoproto.moretags)   = "yaml:\"host_uri\""
  ];
  
  repeated virtengine.base.attributes.v1.Attribute attributes = 3 [
    (gogoproto.castrepeated) = "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1.Attributes",
    (gogoproto.nullable)     = false,
    (gogoproto.jsontag)      = "attributes",
    (gogoproto.moretags)     = "yaml:\"attributes\""
  ];
  
  virtengine.provider.v1beta4.Info info = 4 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "info",
    (gogoproto.moretags) = "yaml:\"info\""
  ];
}
```

## Import Path Changes

### 1. Deployment Module Imports

**Before (v1beta3):**
```protobuf
import "virtengine/deployment/v1beta3/deployment.proto";
import "virtengine/deployment/v1beta3/groupspec.proto";
import "cosmos/base/v1beta1/coin.proto";
```

**After (v1beta4):**
```protobuf
import "virtengine/deployment/v1/deployment.proto";
import "virtengine/deployment/v1beta4/groupspec.proto";
import "virtengine/base/deposit/v1/deposit.proto";
```

### 2. Market Module Imports

**Before (v1beta4):**
```protobuf
// Basic market imports in v1beta4
import "gogoproto/gogo.proto";
import "cosmos/base/v1beta1/coin.proto";
```

**After (v1beta5):**
```protobuf
import "gogoproto/gogo.proto";
import "cosmos/base/v1beta1/coin.proto";
import "virtengine/market/v1beta5/resourcesoffer.proto";
import "virtengine/market/v1/bid.proto";
import "virtengine/base/deposit/v1/deposit.proto";
```

### 3. Escrow Module Imports

**New v1 Escrow Imports:**
```protobuf
import "virtengine/escrow/v1/msg.proto";
import "virtengine/escrow/id/v1/account.proto";
import "virtengine/base/deposit/v1/deposit.proto";
```

### 4. Provider Module Imports

**Before (v1beta3):**
```protobuf
import "virtengine/base/v1beta3/attribute.proto";
```

**After (v1beta4):**
```protobuf
import "cosmos_proto/cosmos.proto";
import "virtengine/base/attributes/v1/attribute.proto";
```

### 5. New v1 Proto Modules

The migration introduces several new v1 modules with enhanced functionality and improved structure:

#### 5.1. Core v1 Modules

**virtengine/deployment/v1/deployment.proto** - Stable deployment types:
```protobuf
// DeploymentID with enhanced documentation and cosmos_proto annotations
message DeploymentID {
  option (gogoproto.equal) = true;
  
  string owner = 1 [
    (cosmos_proto.scalar) = "cosmos.AddressString",
    (gogoproto.jsontag)   = "owner",
    (gogoproto.moretags)  = "yaml:\"owner\""
  ];
  
  uint64 dseq = 2 [
    (gogoproto.customname) = "DSeq",
    (gogoproto.jsontag)    = "dseq",
    (gogoproto.moretags)   = "yaml:\"dseq\""
  ];
}
```

**virtengine/base/attributes/v1/attribute.proto** - Enhanced attribute system:
```protobuf
// Attribute with improved structure
message Attribute {
  string key = 1 [(gogoproto.moretags) = "yaml:\"key\""];
  string value = 2 [(gogoproto.moretags) = "yaml:\"value\""];
}

// New PlacementRequirements for advanced provider selection
message PlacementRequirements {
  SignedBy signed_by = 1 [
    (gogoproto.nullable) = false,
    (gogoproto.jsontag)  = "signed_by",
    (gogoproto.moretags) = "yaml:\"signed_by\""
  ];
  
  repeated Attribute attributes = 2 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "Attributes",
    (gogoproto.jsontag)      = "attributes",
    (gogoproto.moretags)     = "yaml:\"attributes\""
  ];
}
```

#### 5.2. New Service Modules

**virtengine/escrow/v1/** - Dedicated escrow management:
```protobuf
// MsgAccountDeposit for escrow account deposits
message MsgAccountDeposit {
  option (cosmos.msg.v1.signer) = "signer";
  
  string signer = 1 [
    (cosmos_proto.scalar) = "cosmos.AddressString"
  ];
  
  virtengine.escrow.id.v1.Account id = 2 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ID"
  ];
  
  virtengine.base.deposit.v1.Deposit deposit = 3 [
    (gogoproto.nullable) = false
  ];
}
```

**virtengine/take/v1/** - New fee management system:
```protobuf
// DenomTakeRate for denomination-specific take rates
message DenomTakeRate {
  string denom = 1 [(gogoproto.customname) = "Denom"];
  uint32 rate = 2 [(gogoproto.customname) = "Rate"];
}

// Params for take module configuration
message Params {
  repeated DenomTakeRate denom_take_rates = 1 [
    (gogoproto.nullable)     = false,
    (gogoproto.castrepeated) = "DenomTakeRates"
  ];
  
  uint32 default_take_rate = 2 [
    (gogoproto.customname) = "DefaultTakeRate"
  ];
}
```

**virtengine/discovery/v1/** - Client discovery and information:
```protobuf
// VirtEngine-specific RPC parameters
message VirtEngine {
  ClientInfo client_info = 1 [
    (gogoproto.nullable)   = false,
    (gogoproto.customname) = "ClientInfo"
  ];
}
```

#### 5.3. Enhanced Certificate Management

**virtengine/cert/v1/cert.proto** - Improved certificate handling:
```protobuf
// Enhanced certificate ID with cosmos_proto annotations
message ID {
  string owner = 1 [
    (cosmos_proto.scalar) = "cosmos.AddressString",
    (gogoproto.jsontag)   = "owner",
    (gogoproto.moretags)  = "yaml:\"owner\""
  ];
  
  string serial = 2 [
    (gogoproto.jsontag)  = "serial",
    (gogoproto.moretags) = "yaml:\"serial\""
  ];
}

// Certificate with enhanced state management
enum State {
  invalid = 0 [(gogoproto.enumvalue_customname) = "CertificateStateInvalid"];
  valid = 1 [(gogoproto.enumvalue_customname) = "CertificateValid"];
  revoked = 2 [(gogoproto.enumvalue_customname) = "CertificateRevoked"];
}
```

#### 5.4. Enhanced Audit System

**virtengine/audit/v1/audit.proto** - Improved auditing:
```protobuf
// AuditedProvider with enhanced attribute support
message AuditedProvider {
  string owner = 1;
  string auditor = 2;
  
  repeated virtengine.base.attributes.v1.Attribute attributes = 4 [
    (gogoproto.castrepeated) = "github.com/virtengine/virtengine/sdk/go/node/types/attributes/v1.Attributes",
    (gogoproto.nullable)     = false
  ];
}

// AttributesFilters for advanced filtering
message AttributesFilters {
  repeated string auditors = 1;
  repeated string owners = 2;
}
```

## Service Definition Updates

### 1. Deployment Service (v1beta4)

**New Service Definition:**
```protobuf
// virtengine/deployment/v1beta4/service.proto
syntax = "proto3";
package virtengine.deployment.v1beta4;

import "virtengine/deployment/v1beta4/deploymentmsg.proto";
import "virtengine/deployment/v1beta4/groupmsg.proto";
import "virtengine/deployment/v1beta4/paramsmsg.proto";
import "cosmos/msg/v1/msg.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/deployment/v1beta4";

service Msg {
  option (cosmos.msg.v1.service) = true;

  rpc CreateDeployment(MsgCreateDeployment) returns (MsgCreateDeploymentResponse);
  rpc UpdateDeployment(MsgUpdateDeployment) returns (MsgUpdateDeploymentResponse);
  rpc CloseDeployment(MsgCloseDeployment) returns (MsgCloseDeploymentResponse);
  rpc CloseGroup(MsgCloseGroup) returns (MsgCloseGroupResponse);
  rpc PauseGroup(MsgPauseGroup) returns (MsgPauseGroupResponse);
  rpc StartGroup(MsgStartGroup) returns (MsgStartGroupResponse);
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
}
```

### 2. Market Service (v1beta5)

**Enhanced Service Definition:**
```protobuf
// virtengine/market/v1beta5/service.proto
syntax = "proto3";
package virtengine.market.v1beta5;

import "virtengine/market/v1beta5/bidmsg.proto";
import "virtengine/market/v1beta5/leasemsg.proto";
import "virtengine/market/v1beta5/paramsmsg.proto";
import "cosmos/msg/v1/msg.proto";

option go_package = "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5";

service Msg {
  option (cosmos.msg.v1.service) = true;
  
  rpc CreateBid(MsgCreateBid) returns (MsgCreateBidResponse);
  rpc CloseBid(MsgCloseBid) returns (MsgCloseBidResponse);
  rpc WithdrawLease(MsgWithdrawLease) returns (MsgWithdrawLeaseResponse);
  rpc CreateLease(MsgCreateLease) returns (MsgCreateLeaseResponse);
  rpc CloseLease(MsgCloseLease) returns (MsgCloseLeaseResponse);
  rpc UpdateParams(MsgUpdateParams) returns (MsgUpdateParamsResponse);
}
```

## Cosmos SDK Specific Changes

### 1. Authz Module Changes

#### Context Migration in Authorization Interface

**Key Changes:**
- Authorization `Accept` method now uses `context.Context` instead of `sdk.Context`
- Enhanced authorization scoping with new enum types
- Updated protobuf annotations for amino compatibility

**Before (v0.45.x):**
```protobuf
// cosmos/authz/v1beta1/authz.proto
message Grant {
  google.protobuf.Any authorization = 1;
  google.protobuf.Timestamp expiration = 2;
}
```

**After (v0.53.4):**
```protobuf
// cosmos/authz/v1beta1/authz.proto  
message Grant {
  google.protobuf.Any authorization = 1;
  google.protobuf.Timestamp expiration = 2;
}

// Enhanced with amino annotations
message DepositAuthorization {
  option (cosmos_proto.message_added_in)     = "chain-sdk v0.1.0";
  option (cosmos_proto.implements_interface) = "cosmos.authz.v1beta1.Authorization";
  option (amino.name)                        = "virtengine/DepositAuthorization";
  
  cosmos.base.v1beta1.Coin spend_limit = 1;
  repeated Scope scopes = 2;
}
```

#### New Authorization Scoping System

**Enhanced VirtEngine Authorization:**
```protobuf
// virtengine/escrow/v1/authz.proto
message DepositAuthorization {
  enum Scope {
    option (gogoproto.goproto_enum_prefix) = false;

    invalid = 0 [(gogoproto.enumvalue_customname) = "DepositScopeInvalid"];
    deployment = 1 [(gogoproto.enumvalue_customname) = "DepositScopeDeployment"];
    bid = 2 [(gogoproto.enumvalue_customname) = "DepositScopeBid"];
  }

  cosmos.base.v1beta1.Coin spend_limit = 1;
  repeated Scope scopes = 2;
}
```

### 2. Query Service Updates

#### New Query Methods (v0.45.x → v0.53.4)

**Added Query Endpoints:**
```protobuf
service Query {
  // Returns list of Authorization, granted to the grantee by the granter.
  rpc Grants(QueryGrantsRequest) returns (QueryGrantsResponse);
  
  // getGranterGrants returns list of GrantAuthorization, granted by granter.
  // Added in v0.53.4
  rpc GranterGrants(QueryGranterGrantsRequest) returns (QueryGranterGrantsResponse);
  
  // getGranteeGrants returns a list of GrantAuthorization by grantee.
  // Added in v0.53.4  
  rpc GranteeGrants(QueryGranteeGrantsRequest) returns (QueryGranteeGrantsResponse);
}
```

### 3. Auth Module Updates

#### UpdateParams Message (v0.45.x → v0.53.4)

**New Governance Operation:**
```protobuf
// cosmos/auth/v1beta1/tx.proto
message MsgUpdateParams {
  option (cosmos.msg.v1.signer) = "authority";

  // authority is the address that controls the module (defaults to x/gov unless overwritten).
  string authority = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];

  // params defines the x/auth parameters to update.
  Params params = 2 [(gogoproto.nullable) = false, (amino.dont_omitempty) = true];
}
```

### 4. Amino Annotations

#### Enhanced Protobuf Annotations

**New Required Imports:**
```protobuf
import "amino/amino.proto";
import "cosmos_proto/cosmos.proto";
```

**Updated Message Annotations:**
```protobuf
message YourAuthorization {
  option (cosmos_proto.implements_interface) = "cosmos.authz.v1beta1.Authorization";
  option (amino.name) = "your-module/YourAuthorization";
  
  // fields...
}
```

### 5. Staking Authorization Updates

#### Enhanced Staking Authorizations

**Updated Authorization Types:**
```protobuf
// cosmos/staking/v1beta1/authz.proto
enum AuthorizationType {
  // UNSPECIFIED defines an invalid authorization type
  AUTHORIZATION_TYPE_UNSPECIFIED = 0;
  // DELEGATE defines a delegation authorization type
  AUTHORIZATION_TYPE_DELEGATE = 1;
  // UNDELEGATE defines an undelegation authorization type  
  AUTHORIZATION_TYPE_UNDELEGATE = 2;
  // REDELEGATE defines a redelegation authorization type
  AUTHORIZATION_TYPE_REDELEGATE = 3;
  // CANCEL_UNBONDING_DELEGATION defines cancel unbonding delegation authorization type
  AUTHORIZATION_TYPE_CANCEL_UNBONDING_DELEGATION = 4;
}
```

### 6. Bank Authorization Updates

#### Send Authorization Changes

**Enhanced Send Authorization:**
```protobuf
// cosmos/bank/v1beta1/authz.proto
message SendAuthorization {
  option (cosmos_proto.implements_interface) = "cosmos.authz.v1beta1.Authorization";
  option (amino.name) = "cosmos-sdk/SendAuthorization";

  repeated cosmos.base.v1beta1.Coin spend_limit = 1;
  repeated string allow_list = 2;
}
```

### 7. Module Account Permissions

#### Updated Module Account Structure

**Enhanced Permissions System:**
```protobuf
// cosmos/auth/v1beta1/auth.proto
message ModuleAccount {
  option (cosmos_proto.implements_interface) = "cosmos.auth.v1beta1.AccountI";
  option (amino.name) = "cosmos-sdk/ModuleAccount";

  BaseAccount base_account = 1 [(gogoproto.embed) = true];
  string name = 2;
  repeated string permissions = 3;
}
```
