# HPC Proto Generation Setup

Helper scripts and temporary HPC protobuf definition files live under `scripts/hpc`.

## Quick Start

### Windows (Command Prompt)
```cmd
scripts\hpc\setup_hpc_proto.bat
```

### Unix/Linux/macOS/WSL2
```bash
chmod +x scripts/hpc/setup_hpc_proto.sh
./scripts/hpc/setup_hpc_proto.sh
```

### With Auto-Generation (Unix only)
```bash
./scripts/hpc/setup_hpc_proto.sh --generate
```

## Files Created

| Temporary File | Target Location |
|---------------|-----------------|
| `scripts/hpc/proto/hpc_types.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/types.proto` |
| `scripts/hpc/proto/hpc_tx.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/tx.proto` |
| `scripts/hpc/proto/hpc_query.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/query.proto` |
| `scripts/hpc/proto/hpc_genesis.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/genesis.proto` |

## Alternative Setup Methods

### Option 1: Python
```bash
python scripts/hpc/create_network_security_dirs.py
```

### Option 2: Node.js
```bash
node scripts/hpc/setup_hpc_dirs.js
```

### Option 4: Go
```bash
go run scripts/hpc/create_dirs.go
```

### Option 5: Manual
```bash
# Create directory
mkdir -p sdk/proto/node/virtengine/hpc/v1

# Copy files (rename .proto.txt to .proto)
cp scripts/hpc/proto/hpc_types.proto.txt sdk/proto/node/virtengine/hpc/v1/types.proto
cp scripts/hpc/proto/hpc_tx.proto.txt sdk/proto/node/virtengine/hpc/v1/tx.proto
cp scripts/hpc/proto/hpc_query.proto.txt sdk/proto/node/virtengine/hpc/v1/query.proto
cp scripts/hpc/proto/hpc_genesis.proto.txt sdk/proto/node/virtengine/hpc/v1/genesis.proto
```

## Generate Go Files

After the proto files are in place, generate the Go files:

```bash
cd sdk

# Using buf (recommended)
buf generate

# Or using the proto generation script
./script/protocgen.sh go github.com/virtengine/virtengine/sdk/go/node go
```

## Proto Contents

### types.proto
Defines all HPC module types including:
- Enums: `ClusterState`, `JobState`, `HPCRewardSource`, `DisputeStatus`
- Cluster types: `HPCCluster`, `Partition`, `ClusterMetadata`
- Offering types: `HPCOffering`, `QueueOption`, `HPCPricing`, `PreconfiguredWorkload`
- Job types: `HPCJob`, `JobWorkloadSpec`, `JobResources`, `DataReference`, `HPCUsageMetrics`
- Scheduling types: `NodeMetadata`, `LatencyMeasurement`, `NodeResources`, `SchedulingDecision`, `ClusterCandidate`
- Rewards types: `HPCRewardRecord`, `HPCRewardRecipient`, `RewardCalculationDetails`, `HPCDispute`
- Accounting types: `JobAccounting`, `NodeReward`
- `Params` type

### tx.proto
Defines the `Msg` service with transaction messages:
- `MsgRegisterCluster` / `MsgRegisterClusterResponse`
- `MsgUpdateCluster` / `MsgUpdateClusterResponse`
- `MsgDeregisterCluster` / `MsgDeregisterClusterResponse`
- `MsgCreateOffering` / `MsgCreateOfferingResponse`
- `MsgUpdateOffering` / `MsgUpdateOfferingResponse`
- `MsgSubmitJob` / `MsgSubmitJobResponse`
- `MsgCancelJob` / `MsgCancelJobResponse`
- `MsgReportJobStatus` / `MsgReportJobStatusResponse`
- `MsgUpdateNodeMetadata` / `MsgUpdateNodeMetadataResponse`
- `MsgFlagDispute` / `MsgFlagDisputeResponse`
- `MsgResolveDispute` / `MsgResolveDisputeResponse`
- `MsgUpdateParams` / `MsgUpdateParamsResponse`

### query.proto
Defines the `Query` service with query methods:
- Cluster queries: `Cluster`, `Clusters`, `ClustersByProvider`
- Offering queries: `Offering`, `Offerings`, `OfferingsByCluster`
- Job queries: `Job`, `Jobs`, `JobsByCustomer`, `JobsByProvider`, `JobAccounting`
- Node queries: `NodeMetadata`, `NodesByCluster`
- Scheduling queries: `SchedulingDecision`, `SchedulingDecisionByJob`
- Reward queries: `Reward`, `RewardsByJob`
- Dispute queries: `Dispute`, `Disputes`
- Module queries: `Params`

### genesis.proto
Defines `GenesisState` including:
- Module parameters
- Initial clusters, offerings, jobs
- Job accounting records
- Node metadata records
- Scheduling decisions
- HPC rewards and disputes
- Sequence counters

## Cleanup

After successful proto generation, remove the temporary files if you no longer need them:

```bash
rm scripts/hpc/proto/hpc_*.proto.txt scripts/hpc/setup_hpc_proto.* scripts/hpc/create_dirs.go scripts/hpc/create_network_security_dirs.py scripts/hpc/setup_hpc_dirs.js
```
