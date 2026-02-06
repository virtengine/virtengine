# Task 31A: IBC Cross-Chain Settlement Integration

**vibe-kanban ID:** `2d2e0f1e-5e96-4529-8f7d-320d257101c8`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31A |
| **Title** | feat(ibc): Cross-chain settlement integration |
| **Priority** | P2 |
| **Wave** | 4 (Post-Launch) |
| **Estimated LOC** | 3000 |
| **Duration** | 3-4 weeks |
| **Dependencies** | IBC-Go v10 (in go.mod) |
| **Blocking** | None |

---

## Problem Statement

The patent specification implies multi-chain settlement capabilities for broader ecosystem integration. While IBC-Go v10 is included in go.mod, no active IBC channels are configured. Without cross-chain integration:

1. VirtEngine tokens are isolated to the native chain
2. Users cannot leverage existing assets from other Cosmos chains
3. VEID attestations are not portable across chains
4. Settlement options are limited to native tokens only

### Current State Analysis

```
go.mod:
  github.com/cosmos/ibc-go/v10  ✅ Dependency exists:

app/ibc.go:                     ❌ No channel configuration
x/veid/ibc/                     ❌ No IBC packet handlers
deploy/relayer/                 ❌ No relayer configuration
```

---

## Acceptance Criteria

### AC-1: IBC Channel Configuration
- [ ] Establish IBC channel with Cosmos Hub
- [ ] Establish IBC channel with Osmosis (for DEX)
- [ ] Configure transfer channel parameters
- [ ] Implement channel parameter governance proposals
- [ ] Document channel maintenance procedures

### AC-2: Cross-chain VEID Recognition
- [ ] Define VEID attestation IBC packet format
- [ ] Implement VEID packet handler on receiving chain
- [ ] Create remote chain verification protocol
- [ ] Implement score portability with degradation rules
- [ ] Map trust levels across different chains

### AC-3: Settlement Token Bridge
- [ ] Implement IBC token transfer wrappers
- [ ] Create escrow bridge module for cross-chain deposits
- [ ] Support multi-hop routing for complex paths
- [ ] Handle timeout and refund scenarios
- [ ] Implement atomic cross-chain settlements

### AC-4: Documentation & Tooling
- [ ] Create relayer setup guide (Hermes/Go Relayer)
- [ ] Document channel maintenance runbook
- [ ] Create emergency procedures for stuck packets
- [ ] Implement cross-chain E2E test suite

---

## Technical Requirements

### IBC Channel Setup

```go
// app/ibc_channels.go

type ChannelConfig struct {
    CounterpartyChainID string
    ConnectionID        string
    PortID              string
    Version             string
    Ordering            channeltypes.Order
}

var DefaultChannels = []ChannelConfig{
    {
        CounterpartyChainID: "cosmoshub-4",
        PortID:              "transfer",
        Version:             "ics20-1",
        Ordering:            channeltypes.UNORDERED,
    },
    {
        CounterpartyChainID: "osmosis-1",
        PortID:              "transfer",
        Version:             "ics20-1",
        Ordering:            channeltypes.UNORDERED,
    },
    {
        CounterpartyChainID: "cosmoshub-4",
        PortID:              "veid",
        Version:             "veid-1",
        Ordering:            channeltypes.ORDERED,
    },
}
```

### VEID Attestation IBC Packet

```go
// x/veid/ibc/types.go

type VEIDAttestationPacket struct {
    // Source chain VEID record
    SourceChainID   string
    AccountAddress  string
    VEIDHash        []byte
    TrustScore      uint32
    TierLevel       uint32
    
    // Attestation metadata
    AttestationTime int64
    ExpirationTime  int64
    ValidatorSet    []string  // Validators who attested
    
    // Proof of authenticity
    MerkleProof     []byte
    StateRootHash   []byte
}

// Score degradation rules for cross-chain recognition
type CrossChainScorePolicy struct {
    // Apply degradation based on chain trust
    TrustedChains      map[string]uint32  // ChainID -> max score allowed
    DegradationFactor  sdk.Dec            // Score multiplier (e.g., 0.9)
    MinRecognizedTier  uint32             // Minimum tier to accept
    RequiredValidators uint32             // Min validators attesting
}
```

### Escrow Bridge Module

```go
// x/escrow/ibc/bridge.go

type CrossChainDeposit struct {
    SourceChain     string
    SourceChannel   string
    OriginalDenom   string
    IBCDenom        string
    Amount          sdk.Int
    Sender          string
    DepositorOnDest string
    TimeoutHeight   clienttypes.Height
}

// Bridge locks tokens on source, mints vouchers on dest
func (k Keeper) InitiateCrossChainDeposit(
    ctx sdk.Context,
    deposit CrossChainDeposit,
) error {
    // 1. Lock tokens in escrow on source chain
    // 2. Send IBC packet to destination
    // 3. On ACK, confirm deposit
    // 4. On Timeout, refund to sender
}
```

### Relayer Configuration

```toml
# deploy/relayer/config.toml

[chains.virtengine]
id = "virtengine-1"
rpc_addr = "http://localhost:26657"
grpc_addr = "http://localhost:9090"
websocket_addr = "ws://localhost:26657/websocket"
account_prefix = "ve"
gas_price = { price = 0.025, denom = "uve" }

[chains.cosmoshub]
id = "cosmoshub-4"
rpc_addr = "https://cosmos-rpc.polkachu.com:443"
grpc_addr = "https://cosmos-grpc.polkachu.com:14290"
account_prefix = "cosmos"

[channels.transfer]
a_chain = "virtengine"
a_port = "transfer"
b_chain = "cosmoshub"
b_port = "transfer"

[channels.veid]
a_chain = "virtengine"
a_port = "veid"
b_chain = "cosmoshub"
b_port = "veid"
```

---

## Directory Structure

```
x/veid/ibc/
├── types.go              # IBC packet types
├── packet.go             # Packet encoding/decoding
├── handler.go            # OnRecvPacket, OnAcknowledgement
├── keeper.go             # IBC keeper methods
└── module.go             # IBC module registration

x/escrow/ibc/
├── bridge.go             # Cross-chain escrow bridge
├── transfer.go           # Token transfer helpers
└── timeout.go            # Timeout handling

deploy/relayer/
├── config.toml           # Relayer configuration
├── keys/                 # Relayer key storage
├── scripts/
│   ├── init-channels.sh  # Channel initialization
│   └── monitor.sh        # Health monitoring
└── docker-compose.yaml   # Relayer deployment
```

---

## Testing Requirements

### Unit Tests
- IBC packet serialization/deserialization
- VEID attestation validation
- Score degradation calculations
- Timeout handling

### Integration Tests
- Cross-chain token transfer E2E
- VEID attestation relay E2E
- Channel recovery after timeout
- Multi-hop path verification

### Testnet Deployment
- Deploy to testnet with Cosmos Hub testnet
- Run week-long reliability test
- Measure packet latency and success rate

---

## Security Considerations

1. **Channel Trust**: Only accept channels from approved counterparty chains
2. **Score Degradation**: Never accept cross-chain scores at full value
3. **Timeout Handling**: Ensure refunds always occur on timeout
4. **Validator Set Verification**: Verify attestation comes from valid validator set
5. **Replay Prevention**: Include nonces and timestamps in packets
