# HPC Proto Generation Setup

This directory contains HPC protobuf definition files ready to be moved to their correct location.

## Quick Start

### Windows (Command Prompt)
```cmd
setup_hpc_proto.bat
```

### Unix/Linux/macOS/WSL2
```bash
chmod +x setup_hpc_proto.sh
./setup_hpc_proto.sh
```

### With Auto-Generation (Unix only)
```bash
./setup_hpc_proto.sh --generate
```

## Files Created

| Temporary File | Target Location |
|---------------|-----------------|
| `hpc_types.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/types.proto` |
| `hpc_tx.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/tx.proto` |
| `hpc_query.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/query.proto` |
| `hpc_genesis.proto.txt` | `sdk/proto/node/virtengine/hpc/v1/genesis.proto` |

## Alternative Setup Methods

### Option 1: Python
```bash
python create_network_security_dirs.py
```

### Option 2: Node.js
```bash
node setup_hpc_dirs.js
```

### Option 4: Go
```bash
go run create_dirs.go
```

### Option 5: Manual
```bash
# Create directory
mkdir -p sdk/proto/node/virtengine/hpc/v1

# Copy files (rename .proto.txt to .proto)
cp hpc_types.proto.txt sdk/proto/node/virtengine/hpc/v1/types.proto
cp hpc_tx.proto.txt sdk/proto/node/virtengine/hpc/v1/tx.proto
cp hpc_query.proto.txt sdk/proto/node/virtengine/hpc/v1/query.proto
cp hpc_genesis.proto.txt sdk/proto/node/virtengine/hpc/v1/genesis.proto
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

After successful proto generation, remove the temporary files:

```bash
rm hpc_*.proto.txt setup_hpc_proto.bat create_dirs.go create_network_security_dirs.py setup_hpc_dirs.js HPC_PROTO_README.md
```
