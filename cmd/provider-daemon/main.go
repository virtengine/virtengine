// Package main implements the VirtEngine provider daemon CLI.
//
// The provider daemon is responsible for:
// - VE-400: Key management and transaction signing
// - VE-401: Bid engine and provider configuration watcher
// - VE-402: Manifest parsing and validation
// - VE-403: Kubernetes orchestration adapter
// - VE-404: Usage metering and on-chain recording
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	provider_daemon "pkg.akt.dev/node/pkg/provider_daemon"
)

const (
	// FlagChainID is the blockchain chain ID
	FlagChainID = "chain-id"

	// FlagNode is the blockchain node RPC endpoint
	FlagNode = "node"

	// FlagProviderKey is the provider's key name
	FlagProviderKey = "provider-key"

	// FlagProviderKeyDir is the directory containing provider keys
	FlagProviderKeyDir = "key-dir"

	// FlagKubeconfig is the path to kubeconfig
	FlagKubeconfig = "kubeconfig"

	// FlagMeteringInterval is the metering interval
	FlagMeteringInterval = "metering-interval"

	// FlagBidRateLimitMinute is the per-minute bid rate limit
	FlagBidRateLimitMinute = "bid-rate-limit-minute"

	// FlagBidRateLimitHour is the per-hour bid rate limit
	FlagBidRateLimitHour = "bid-rate-limit-hour"

	// FlagResourcePrefix is the prefix for Kubernetes resources
	FlagResourcePrefix = "resource-prefix"

	// FlagListenAddr is the API listen address
	FlagListenAddr = "listen"

	// FlagMetricsAddr is the metrics listen address
	FlagMetricsAddr = "metrics"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "provider-daemon",
		Short: "VirtEngine Provider Daemon",
		Long: `The VirtEngine Provider Daemon manages compute resources and workloads
on behalf of a provider in the VirtEngine decentralized cloud marketplace.

It handles:
- Key management and transaction signing
- Automatic bidding on orders matching provider capacity
- Manifest parsing and validation
- Kubernetes workload orchestration
- Usage metering and on-chain recording`,
	}
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.virtengine/provider-daemon.yaml)")
	rootCmd.PersistentFlags().String(FlagChainID, "virtengine-1", "Blockchain chain ID")
	rootCmd.PersistentFlags().String(FlagNode, "tcp://localhost:26657", "Blockchain node RPC endpoint")
	rootCmd.PersistentFlags().String(FlagProviderKey, "provider", "Provider key name")
	rootCmd.PersistentFlags().String(FlagProviderKeyDir, "", "Directory containing provider keys")
	rootCmd.PersistentFlags().String(FlagListenAddr, ":8080", "API listen address")
	rootCmd.PersistentFlags().String(FlagMetricsAddr, ":9090", "Metrics listen address")

	// Bind to viper
	viper.BindPFlag(FlagChainID, rootCmd.PersistentFlags().Lookup(FlagChainID))
	viper.BindPFlag(FlagNode, rootCmd.PersistentFlags().Lookup(FlagNode))
	viper.BindPFlag(FlagProviderKey, rootCmd.PersistentFlags().Lookup(FlagProviderKey))
	viper.BindPFlag(FlagProviderKeyDir, rootCmd.PersistentFlags().Lookup(FlagProviderKeyDir))
	viper.BindPFlag(FlagListenAddr, rootCmd.PersistentFlags().Lookup(FlagListenAddr))
	viper.BindPFlag(FlagMetricsAddr, rootCmd.PersistentFlags().Lookup(FlagMetricsAddr))

	// Add commands
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(initKeyCmd())
	rootCmd.AddCommand(rotateKeyCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(validateManifestCmd())
	rootCmd.AddCommand(versionCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home + "/.virtengine")
		viper.SetConfigType("yaml")
		viper.SetConfigName("provider-daemon")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("VIRTENGINE")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the provider daemon",
		Long:  `Starts the provider daemon, enabling bidding, workload management, and metering.`,
		RunE:  runStart,
	}

	cmd.Flags().String(FlagKubeconfig, "", "Path to kubeconfig file (defaults to in-cluster config)")
	cmd.Flags().Duration(FlagMeteringInterval, time.Hour, "Usage metering interval")
	cmd.Flags().Int(FlagBidRateLimitMinute, 10, "Maximum bids per minute")
	cmd.Flags().Int(FlagBidRateLimitHour, 100, "Maximum bids per hour")
	cmd.Flags().String(FlagResourcePrefix, "ve", "Prefix for Kubernetes resources")

	viper.BindPFlag(FlagKubeconfig, cmd.Flags().Lookup(FlagKubeconfig))
	viper.BindPFlag(FlagMeteringInterval, cmd.Flags().Lookup(FlagMeteringInterval))
	viper.BindPFlag(FlagBidRateLimitMinute, cmd.Flags().Lookup(FlagBidRateLimitMinute))
	viper.BindPFlag(FlagBidRateLimitHour, cmd.Flags().Lookup(FlagBidRateLimitHour))
	viper.BindPFlag(FlagResourcePrefix, cmd.Flags().Lookup(FlagResourcePrefix))

	return cmd
}

func runStart(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Starting VirtEngine Provider Daemon...")
	fmt.Printf("  Chain ID: %s\n", viper.GetString(FlagChainID))
	fmt.Printf("  Node: %s\n", viper.GetString(FlagNode))
	fmt.Printf("  API Address: %s\n", viper.GetString(FlagListenAddr))
	fmt.Printf("  Metrics Address: %s\n", viper.GetString(FlagMetricsAddr))

	// Initialize key manager (VE-400)
	keyManager := provider_daemon.NewKeyManager()
	fmt.Println("  Key Manager: initialized")

	// Load or generate provider key
	keyName := viper.GetString(FlagProviderKey)
	keyDir := viper.GetString(FlagProviderKeyDir)
	if keyDir != "" {
		// Import key from file
		keyPath := keyDir + "/" + keyName + ".key"
		if _, err := os.Stat(keyPath); err == nil {
			data, err := os.ReadFile(keyPath)
			if err != nil {
				return fmt.Errorf("failed to read key file: %w", err)
			}
			if err := keyManager.ImportKey(ctx, keyName, data); err != nil {
				return fmt.Errorf("failed to import key: %w", err)
			}
			fmt.Printf("  Provider Key: loaded from %s\n", keyPath)
		} else {
			// Generate new key
			if err := keyManager.GenerateKey(ctx, keyName, provider_daemon.KeyStorageFile, map[string]string{
				"path": keyPath,
			}); err != nil {
				return fmt.Errorf("failed to generate key: %w", err)
			}
			fmt.Printf("  Provider Key: generated at %s\n", keyPath)
		}
	} else {
		// Use memory storage for development
		if err := keyManager.GenerateKey(ctx, keyName, provider_daemon.KeyStorageMemory, nil); err != nil {
			return fmt.Errorf("failed to generate key: %w", err)
		}
		fmt.Println("  Provider Key: generated (memory only - for development)")
	}

	// Get provider ID from key
	key, err := keyManager.GetKey(keyName)
	if err != nil {
		return fmt.Errorf("failed to get provider key: %w", err)
	}
	providerID := key.PublicKeyHex()
	fmt.Printf("  Provider ID: %s...\n", providerID[:16])

	// Initialize bid engine (VE-401)
	configChan := make(chan provider_daemon.ProviderConfig, 1)
	orderChan := make(chan provider_daemon.Order, 100)
	bidResultChan := make(chan provider_daemon.BidResult, 100)

	bidEngine := provider_daemon.NewBidEngine(provider_daemon.BidEngineConfig{
		ProviderID:      providerID,
		ConfigChan:      configChan,
		OrderChan:       orderChan,
		BidResultChan:   bidResultChan,
		PerMinuteLimit:  viper.GetInt(FlagBidRateLimitMinute),
		PerHourLimit:    viper.GetInt(FlagBidRateLimitHour),
	})

	// Load initial config
	initialConfig := provider_daemon.ProviderConfig{
		ProviderID: providerID,
		Capacity: provider_daemon.CapacityConfig{
			TotalCPU:    1000000, // 1000 cores
			TotalMemory: 1024 * 1024 * 1024 * 1024, // 1 TB
			TotalGPU:    100,
			UsedCPU:     0,
			UsedMemory:  0,
			UsedGPU:     0,
		},
		Pricing: provider_daemon.PricingConfig{
			CPURate:     "0.001",
			MemoryRate:  "0.0001",
			StorageRate: "0.00001",
			GPURate:     "0.01",
		},
		Regions:  []string{"us-east", "us-west", "eu-west"},
		Tags:     []string{"gpu", "ssd", "trusted"},
		Enabled:  true,
	}
	configChan <- initialConfig

	bidEngine.Start(ctx)
	fmt.Println("  Bid Engine: started")

	// Initialize Kubernetes adapter (VE-403)
	statusUpdateChan := make(chan provider_daemon.WorkloadStatusUpdate, 100)

	// Note: In production, this would use a real Kubernetes client
	// For now, we'll use a placeholder that demonstrates the integration
	fmt.Println("  Kubernetes Adapter: initialized (placeholder)")

	// Initialize usage meter (VE-404)
	recordChan := make(chan *provider_daemon.UsageRecord, 100)

	usageMeter := provider_daemon.NewUsageMeter(provider_daemon.UsageMeterConfig{
		ProviderID: providerID,
		Interval:   provider_daemon.MeteringInterval(viper.GetDuration(FlagMeteringInterval)),
		KeyManager: keyManager,
		RecordChan: recordChan,
	})

	usageMeter.Start(ctx)
	fmt.Println("  Usage Meter: started")

	// Start background workers
	go handleBidResults(ctx, bidResultChan)
	go handleStatusUpdates(ctx, statusUpdateChan)
	go handleUsageRecords(ctx, recordChan)

	fmt.Println("\nProvider daemon is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	select {
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
	case <-ctx.Done():
		fmt.Println("\nContext cancelled, shutting down...")
	}

	// Graceful shutdown
	fmt.Println("Stopping services...")

	bidEngine.Stop()
	fmt.Println("  Bid Engine: stopped")

	usageMeter.Stop()
	fmt.Println("  Usage Meter: stopped")

	keyManager.Lock()
	fmt.Println("  Key Manager: locked")

	fmt.Println("Provider daemon stopped.")
	return nil
}

func handleBidResults(ctx context.Context, ch <-chan provider_daemon.BidResult) {
	for {
		select {
		case <-ctx.Done():
			return
		case result := <-ch:
			if result.Success {
				fmt.Printf("[BID] Submitted bid for order %s: %s\n", result.OrderID, result.BidID)
			} else {
				fmt.Printf("[BID] Failed to bid on order %s: %s\n", result.OrderID, result.Error)
			}
		}
	}
}

func handleStatusUpdates(ctx context.Context, ch <-chan provider_daemon.WorkloadStatusUpdate) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-ch:
			fmt.Printf("[WORKLOAD] %s: %s - %s\n", update.WorkloadID, update.State, update.Message)
		}
	}
}

func handleUsageRecords(ctx context.Context, ch <-chan *provider_daemon.UsageRecord) {
	for {
		select {
		case <-ctx.Done():
			return
		case record := <-ch:
			fmt.Printf("[USAGE] Workload %s: CPU=%dms, Mem=%d bytes\n",
				record.WorkloadID,
				record.Metrics.CPUMilliSeconds,
				record.Metrics.MemoryByteSeconds,
			)
		}
	}
}

func initKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init-key [name]",
		Short: "Initialize a new provider key",
		Long:  `Creates a new provider key for transaction signing.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			keyDir := viper.GetString(FlagProviderKeyDir)
			if keyDir == "" {
				home, _ := os.UserHomeDir()
				keyDir = home + "/.virtengine/keys"
			}

			if err := os.MkdirAll(keyDir, 0700); err != nil {
				return fmt.Errorf("failed to create key directory: %w", err)
			}

			ctx := context.Background()
			keyManager := provider_daemon.NewKeyManager()

			keyPath := keyDir + "/" + keyName + ".key"
			if err := keyManager.GenerateKey(ctx, keyName, provider_daemon.KeyStorageFile, map[string]string{
				"path": keyPath,
			}); err != nil {
				return fmt.Errorf("failed to generate key: %w", err)
			}

			key, _ := keyManager.GetKey(keyName)
			fmt.Printf("Generated key '%s'\n", keyName)
			fmt.Printf("  Path: %s\n", keyPath)
			fmt.Printf("  Public Key: %s\n", key.PublicKeyHex())

			return nil
		},
	}

	return cmd
}

func rotateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate-key [name]",
		Short: "Rotate a provider key",
		Long:  `Creates a new key and marks the old one for rotation.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			fmt.Printf("Key rotation for '%s' - this would rotate the key in production\n", keyName)
			return nil
		},
	}

	return cmd
}

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show provider daemon status",
		Long:  `Displays the current status of the provider daemon.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Provider Daemon Status")
			fmt.Println("======================")
			fmt.Printf("Chain ID: %s\n", viper.GetString(FlagChainID))
			fmt.Printf("Node: %s\n", viper.GetString(FlagNode))
			fmt.Println("\nNote: Connect to running daemon for live status")
			return nil
		},
	}

	return cmd
}

func validateManifestCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-manifest [file]",
		Short: "Validate a deployment manifest",
		Long:  `Parses and validates a deployment manifest file.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read manifest: %w", err)
			}

			parser := provider_daemon.NewManifestParser()
			manifest, err := parser.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse manifest: %w", err)
			}

			result := parser.Validate(manifest)

			fmt.Printf("Manifest: %s\n", filePath)
			fmt.Printf("  Name: %s\n", manifest.Name)
			fmt.Printf("  Version: %s\n", manifest.Version)
			fmt.Printf("  Services: %d\n", manifest.ServiceCount())

			resources := manifest.TotalResources()
			fmt.Printf("  Total CPU: %d millicores\n", resources.CPU)
			fmt.Printf("  Total Memory: %d bytes\n", resources.Memory)
			fmt.Printf("  Total GPU: %d\n", resources.GPU)

			if result.Valid {
				fmt.Println("\n✓ Manifest is valid")
			} else {
				fmt.Println("\n✗ Manifest validation failed:")
				for _, err := range result.Errors {
					fmt.Printf("  - [%s] %s: %s\n", err.Code, err.Field, err.Message)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Println("\nWarnings:")
				for _, warn := range result.Warnings {
					fmt.Printf("  - %s\n", warn)
				}
			}

			if !result.Valid {
				os.Exit(1)
			}

			return nil
		},
	}

	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("VirtEngine Provider Daemon")
			fmt.Println("  Version: 0.1.0")
			fmt.Println("  Features: VE-400, VE-401, VE-402, VE-403, VE-404")
		},
	}
}
