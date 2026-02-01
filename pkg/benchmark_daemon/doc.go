// Package benchmark_daemon implements the benchmarking daemon for VirtEngine providers.
//
// VE-600: Benchmark daemon documentation
package benchmark_daemon

// BenchmarkDaemon Package Documentation
//
// The benchmark daemon is a provider-managed agent that collects performance
// metrics and submits signed benchmark reports to the VirtEngine blockchain.
//
// # Features
//
// - Scheduled benchmark execution at configurable intervals
// - On-demand challenge response for anti-gaming verification
// - Signed reports using provider's benchmarking key
// - Rate limiting and retry logic for chain submissions
// - Support for CPU, memory, disk, network, and GPU metrics
//
// # Metrics Collected
//
// CPU:
// - Single-core performance score
// - Multi-core performance score
// - Core and thread counts
// - Base and boost frequencies
//
// Memory:
// - Total memory
// - Bandwidth (MB/s)
// - Latency (nanoseconds)
//
// Disk:
// - Read/Write IOPS
// - Read/Write throughput (MB/s)
// - Total storage capacity
//
// Network:
// - Throughput (Mbps)
// - Latency/RTT to reference endpoints
// - Packet loss rate
//
// GPU (optional):
// - Device count and type
// - Total memory
// - Compute score
// - Memory bandwidth
//
// # Security
//
// - Never logs or transmits any secrets
// - Configuration secrets are kept provider-side and encrypted at rest
// - All reports are signed with the provider's benchmarking key
// - Signature verification happens on-chain
//
// # Usage
//
//	config := benchmark_daemon.DefaultBenchmarkDaemonConfig()
//	config.ProviderAddress = "cosmos1..."
//	config.ClusterID = "cluster-1"
//	config.ChainEndpoint = "http://localhost:26657"
//
//	daemon, err := benchmark_daemon.NewBenchmarkDaemon(config, client, runner, signer)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	if err := daemon.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//
//	// Run an on-demand benchmark
//	result, err := daemon.RunBenchmark(ctx)
//	if err != nil {
//		log.Printf("Benchmark failed: %v", err)
//	}
//
//	// Stop the daemon
//	daemon.Stop()

