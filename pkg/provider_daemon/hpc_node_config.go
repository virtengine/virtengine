package provider_daemon

import (
	"fmt"
	"time"
)

// Validate validates node aggregator configuration.
func (c *HPCNodeAggregatorConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.ProviderAddress == "" {
		return fmt.Errorf("node_aggregator.provider_address required")
	}
	if c.ListenAddr == "" {
		return fmt.Errorf("node_aggregator.listen_addr required")
	}
	if c.HeartbeatTimeout < 10*time.Second {
		return fmt.Errorf("node_aggregator.heartbeat_timeout must be >= 10s")
	}
	if c.MaxBatchSize < 1 {
		return fmt.Errorf("node_aggregator.max_batch_size must be >= 1")
	}
	if c.BatchSubmitInterval < time.Second {
		return fmt.Errorf("node_aggregator.batch_submit_interval must be >= 1s")
	}
	if c.CheckpointInterval < time.Second {
		return fmt.Errorf("node_aggregator.checkpoint_interval must be >= 1s")
	}
	return nil
}
