# State Pruning Package

The `pkg/pruning` package provides optimized blockchain state pruning for VirtEngine with storage efficiency while maintaining historical state query support and state sync compatibility.

## Features

- **Multiple Pruning Strategies**: Default, Nothing (archive), Everything (minimal), Custom, and Tiered
- **Tiered Retention**: Keep full resolution for recent blocks, sample older blocks
- **Snapshot Management**: Automatic snapshot lifecycle with configurable retention
- **Disk Monitoring**: Usage tracking with configurable alert thresholds
- **Growth Projections**: Predict storage requirements and days until full
- **Performance Benchmarks**: Compare storage savings across strategies
- **State Sync Compatible**: Validated compatibility with CometBFT state sync

## Pruning Strategies

### Default
Keeps the last 362,880 states (~3 weeks at 6s blocks), pruning every 10 blocks.
Best for: Standard validator nodes.

### Nothing
Archives all historical states forever.
Best for: Archive nodes, block explorers, analytics services.

### Everything
Keeps only 2 most recent states.
Best for: Light nodes, temporary testing.

### Custom
Manual specification of keep-recent and pruning interval.
Best for: Fine-tuned node configurations.

### Tiered
Intelligent sampling based on block age:
- **Tier 1**: Full resolution for recent blocks (default: ~1 week)
- **Tier 2**: Sample every 100 blocks (default: ~1 month)
- **Tier 3**: Sample every 1000 blocks (archive)

Best for: Nodes needing historical access with reduced storage.

## Usage

### Basic Configuration

```go
import "github.com/virtengine/virtengine/pkg/pruning"

// Default configuration
cfg := pruning.DefaultConfig()

// Archive node
cfg := pruning.NothingConfig()

// Minimal storage
cfg := pruning.EverythingConfig()

// Tiered retention
cfg := pruning.TieredRetentionConfig()
```

### State Manager

```go
manager, err := pruning.NewStateManager(cfg, logger)
if err != nil {
    return err
}

// Check if a height should be pruned
shouldPrune := manager.ShouldPruneHeight(currentHeight, targetHeight)

// Calculate prune range
from, to, shouldPrune := manager.CalculatePruneRange(currentHeight, lastPruned)
```

### Snapshot Manager

```go
snapshotMgr, err := pruning.NewSnapshotManager(cfg.Snapshot, dataDir, logger)
if err != nil {
    return err
}

// Check if snapshot should be created
if snapshotMgr.ShouldCreateSnapshot(height) {
    // Create snapshot...
}

// Register snapshot
snapshotMgr.RegisterSnapshot(pruning.SnapshotInfo{
    Height:    height,
    Format:    1,
    CreatedAt: time.Now(),
    Size:      size,
})
```

### Disk Monitoring

```go
diskMon := pruning.NewDiskMonitor(cfg.DiskMonitor, dataDir, logger)
diskMon.Start(ctx)
defer diskMon.Stop()

// Get status
status, err := diskMon.GetStatus()
if status.AlertLevel == pruning.AlertLevelCritical {
    // Handle critical disk usage
}

// Get growth projection
projection := diskMon.CalculateGrowthProjection()
fmt.Printf("Days until full: %d\n", projection.DaysUntilFull)
```

### Complete Manager

```go
manager, err := pruning.NewManager(cfg, dataDir, logger)
if err != nil {
    return err
}

// Start background monitoring
manager.Start(ctx)
defer manager.Stop()

// Get complete status
status, err := manager.GetStatus()
```

## Configuration Reference

```go
type Config struct {
    Strategy        Strategy  // Pruning strategy
    KeepRecent      uint64    // Heights to keep (default: 362880)
    Interval        uint64    // Pruning interval (default: 10)
    MinRetainBlocks uint64    // Min blocks for CometBFT
    Snapshot        SnapshotConfig
    Tiered          TieredConfig
    DiskMonitor     DiskMonitorConfig
    HistoricalQueries HistoricalQueryConfig
}

type SnapshotConfig struct {
    Enabled          bool      // Enable snapshots
    Interval         uint64    // Creation interval (default: 1000)
    KeepRecent       uint32    // Snapshots to keep (default: 2)
    Compression      bool      // Enable compression
    CompressionLevel int       // 1-9 (default: 6)
}

type DiskMonitorConfig struct {
    Enabled                  bool          // Enable monitoring
    CheckInterval            time.Duration // Check frequency
    WarningThresholdPercent  float64       // Warning level (default: 80%)
    CriticalThresholdPercent float64       // Critical level (default: 90%)
    AutoPruneOnCritical      bool          // Auto-prune on critical
}
```

## Benchmarks

Compare storage savings across strategies:

```go
results := pruning.CompareBenchmarks(1000000, 1024)

for strategy, result := range results {
    fmt.Printf("%s: %.1f%% savings\n", strategy, result.SavingsPercent)
}
```

## State Sync Compatibility

Validate configuration for state sync:

```go
if !cfg.IsStateSyncCompatible() {
    return ErrStateSyncIncompatible
}

err := manager.ValidateStateSyncCompatibility()
```

## Metrics

Access pruning metrics for monitoring:

```go
metrics := manager.Metrics().GetMetrics()

// Available metrics:
// - TotalPruningOperations
// - TotalHeightsPruned
// - TotalBytesPruned
// - LastPruningDuration
// - AveragePruningDuration
// - CurrentDiskUsage
// - DailyGrowthRate
// - DaysUntilFull
```

## Testing

```bash
go test -v ./pkg/pruning/...
```

## Related Documentation

- [Cosmos SDK Pruning](https://docs.cosmos.network/main/run-node/run-node#pruning)
- [CometBFT State Sync](https://docs.cometbft.com/v0.38/core/state-sync)
