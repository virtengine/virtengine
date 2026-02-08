// Package main implements the VirtEngine HPC Node Agent CLI.
//
// The node agent runs on HPC compute nodes and is responsible for:
// - VE-500: Node registration and identity management
// - VE-500: Periodic heartbeat with capacity, health, and latency metrics
// - VE-500: Signed payload submission to provider daemon
package main

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Configuration flags
	FlagNodeID            = "node-id"
	FlagClusterID         = "cluster-id"
	FlagProviderAddress   = "provider-address"
	FlagProviderDaemonURL = "provider-daemon-url"
	FlagHeartbeatInterval = "heartbeat-interval"
	FlagKeyFile           = "key-file"
	FlagHostname          = "hostname"
	FlagRegion            = "region"
	FlagDatacenter        = "datacenter"
	FlagZone              = "zone"
	FlagRack              = "rack"
	FlagRow               = "row"
	FlagPosition          = "position"
	FlagLatencyTargets    = "latency-targets"
	FlagLogLevel          = "log-level"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "hpc-node-agent",
		Short: "VirtEngine HPC Node Agent",
		Long: `The VirtEngine HPC Node Agent runs on HPC compute nodes to report
node health, capacity, and latency metrics to the provider daemon.

It handles:
- Node registration with Ed25519 key pair
- Periodic heartbeat with signed payloads
- Capacity metrics collection (CPU, memory, GPU, storage)
- Health metrics collection (load, utilization, temperature)
- Latency measurements to other cluster nodes`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/virtengine/hpc-node-agent.yaml)")
	rootCmd.PersistentFlags().String(FlagNodeID, "", "Unique node identifier")
	rootCmd.PersistentFlags().String(FlagClusterID, "", "Parent cluster identifier")
	rootCmd.PersistentFlags().String(FlagProviderAddress, "", "Provider's blockchain address (bech32)")
	rootCmd.PersistentFlags().String(FlagProviderDaemonURL, "http://localhost:8080", "Provider daemon API URL")
	rootCmd.PersistentFlags().Duration(FlagHeartbeatInterval, 30*time.Second, "Heartbeat interval")
	rootCmd.PersistentFlags().String(FlagKeyFile, "/etc/virtengine/virtengine-agent.key", "Path to Ed25519 key file")
	rootCmd.PersistentFlags().String(FlagHostname, "", "Node hostname (auto-detected if empty)")
	rootCmd.PersistentFlags().String(FlagRegion, "", "Geographic region")
	rootCmd.PersistentFlags().String(FlagDatacenter, "", "Datacenter identifier")
	rootCmd.PersistentFlags().String(FlagZone, "", "Availability zone")
	rootCmd.PersistentFlags().String(FlagRack, "", "Rack identifier")
	rootCmd.PersistentFlags().String(FlagRow, "", "Row identifier")
	rootCmd.PersistentFlags().String(FlagPosition, "", "Position within rack/row")
	rootCmd.PersistentFlags().StringSlice(FlagLatencyTargets, nil, "Node IDs to measure latency to")
	rootCmd.PersistentFlags().String(FlagLogLevel, "info", "Log level (debug, info, warn, error)")

	// Bind to viper
	_ = viper.BindPFlag(FlagNodeID, rootCmd.PersistentFlags().Lookup(FlagNodeID))
	_ = viper.BindPFlag(FlagClusterID, rootCmd.PersistentFlags().Lookup(FlagClusterID))
	_ = viper.BindPFlag(FlagProviderAddress, rootCmd.PersistentFlags().Lookup(FlagProviderAddress))
	_ = viper.BindPFlag(FlagProviderDaemonURL, rootCmd.PersistentFlags().Lookup(FlagProviderDaemonURL))
	_ = viper.BindPFlag(FlagHeartbeatInterval, rootCmd.PersistentFlags().Lookup(FlagHeartbeatInterval))
	_ = viper.BindPFlag(FlagKeyFile, rootCmd.PersistentFlags().Lookup(FlagKeyFile))
	_ = viper.BindPFlag(FlagHostname, rootCmd.PersistentFlags().Lookup(FlagHostname))
	_ = viper.BindPFlag(FlagRegion, rootCmd.PersistentFlags().Lookup(FlagRegion))
	_ = viper.BindPFlag(FlagDatacenter, rootCmd.PersistentFlags().Lookup(FlagDatacenter))
	_ = viper.BindPFlag(FlagZone, rootCmd.PersistentFlags().Lookup(FlagZone))
	_ = viper.BindPFlag(FlagRack, rootCmd.PersistentFlags().Lookup(FlagRack))
	_ = viper.BindPFlag(FlagRow, rootCmd.PersistentFlags().Lookup(FlagRow))
	_ = viper.BindPFlag(FlagPosition, rootCmd.PersistentFlags().Lookup(FlagPosition))
	_ = viper.BindPFlag(FlagLatencyTargets, rootCmd.PersistentFlags().Lookup(FlagLatencyTargets))
	_ = viper.BindPFlag(FlagLogLevel, rootCmd.PersistentFlags().Lookup(FlagLogLevel))

	// Add commands
	rootCmd.AddCommand(startCmd())
	rootCmd.AddCommand(initCmd())
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(statusCmd())
	rootCmd.AddCommand(versionCmd())
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/virtengine")
		viper.AddConfigPath("$HOME/.virtengine")
		viper.SetConfigType("yaml")
		viper.SetConfigName("hpc-node-agent")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("VE_NODE")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the node agent",
		Long:  `Starts the node agent, beginning heartbeat and metrics collection.`,
		RunE:  runStart,
	}
}

func runStart(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	nodeID := viper.GetString(FlagNodeID)
	clusterID := viper.GetString(FlagClusterID)
	providerAddress := viper.GetString(FlagProviderAddress)
	providerDaemonURL := viper.GetString(FlagProviderDaemonURL)
	heartbeatInterval := viper.GetDuration(FlagHeartbeatInterval)
	keyFile := viper.GetString(FlagKeyFile)

	if nodeID == "" {
		return fmt.Errorf("--%s is required", FlagNodeID)
	}
	if clusterID == "" {
		return fmt.Errorf("--%s is required", FlagClusterID)
	}
	if providerAddress == "" {
		return fmt.Errorf("--%s is required", FlagProviderAddress)
	}

	fmt.Println("Starting VirtEngine HPC Node Agent...")
	fmt.Printf("  Node ID: %s\n", nodeID)
	fmt.Printf("  Cluster ID: %s\n", clusterID)
	fmt.Printf("  Provider: %s\n", providerAddress)
	fmt.Printf("  Provider Daemon: %s\n", providerDaemonURL)
	fmt.Printf("  Heartbeat Interval: %s\n", heartbeatInterval)

	// Load or generate key
	privateKey, publicKey, err := loadOrGenerateKey(keyFile)
	if err != nil {
		return fmt.Errorf("failed to load key: %w", err)
	}
	fmt.Printf("  Public Key: %s...\n", base64.StdEncoding.EncodeToString(publicKey)[:32])

	// Get hostname
	hostname := viper.GetString(FlagHostname)
	if hostname == "" {
		hostname, _ = os.Hostname()
	}

	// Create agent configuration
	config := AgentConfig{
		NodeID:            nodeID,
		ClusterID:         clusterID,
		ProviderAddress:   providerAddress,
		ProviderDaemonURL: providerDaemonURL,
		HeartbeatInterval: heartbeatInterval,
		PrivateKey:        privateKey,
		PublicKey:         publicKey,
		Hostname:          hostname,
		Region:            viper.GetString(FlagRegion),
		Datacenter:        viper.GetString(FlagDatacenter),
		Zone:              viper.GetString(FlagZone),
		Rack:              viper.GetString(FlagRack),
		Row:               viper.GetString(FlagRow),
		Position:          viper.GetString(FlagPosition),
		LatencyTargets:    viper.GetStringSlice(FlagLatencyTargets),
	}

	// Create and start agent
	agent := NewAgent(config)
	if err := agent.Start(ctx); err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	fmt.Println("\nNode agent is running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	select {
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
	case <-ctx.Done():
		fmt.Println("\nContext cancelled, shutting down...")
	}

	agent.Stop()
	fmt.Println("Node agent stopped.")
	return nil
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize node agent keys and configuration",
		Long:  `Generates Ed25519 key pair and creates initial configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			keyFile := viper.GetString(FlagKeyFile)

			// Check if key already exists
			if _, err := os.Stat(keyFile); err == nil {
				return fmt.Errorf("key file already exists: %s", keyFile)
			}

			// Generate new key
			publicKey, privateKey, err := ed25519.GenerateKey(nil)
			if err != nil {
				return fmt.Errorf("failed to generate key: %w", err)
			}

			// Save key
			keyData := struct {
				PrivateKey string `json:"private_key"`
				PublicKey  string `json:"public_key"`
			}{
				PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
				PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
			}

			data, err := json.MarshalIndent(keyData, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal key: %w", err)
			}

			if err := os.WriteFile(keyFile, data, 0600); err != nil {
				return fmt.Errorf("failed to write key file: %w", err)
			}

			fmt.Printf("Generated key pair: %s\n", keyFile)
			fmt.Printf("  Public Key: %s\n", base64.StdEncoding.EncodeToString(publicKey))
			fmt.Println("\nShare the public key with your provider for allowlist registration.")
			return nil
		},
	}
}

func registerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register",
		Short: "Register node with provider daemon",
		Long:  `Sends registration request to provider daemon with node identity.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID := viper.GetString(FlagNodeID)
			clusterID := viper.GetString(FlagClusterID)
			providerAddress := viper.GetString(FlagProviderAddress)
			providerDaemonURL := viper.GetString(FlagProviderDaemonURL)
			keyFile := viper.GetString(FlagKeyFile)

			if nodeID == "" || clusterID == "" || providerAddress == "" {
				return fmt.Errorf("--node-id, --cluster-id, and --provider-address are required")
			}
			if providerDaemonURL == "" {
				return fmt.Errorf("--provider-daemon-url is required")
			}

			privateKey, publicKey, err := loadOrGenerateKey(keyFile)
			if err != nil {
				return fmt.Errorf("failed to load key: %w", err)
			}

			hostname, _ := os.Hostname()
			if h := viper.GetString(FlagHostname); h != "" {
				hostname = h
			}

			config := AgentConfig{
				NodeID:            nodeID,
				ClusterID:         clusterID,
				ProviderAddress:   providerAddress,
				ProviderDaemonURL: providerDaemonURL,
				HeartbeatInterval: 30 * time.Second,
				PrivateKey:        privateKey,
				PublicKey:         publicKey,
				Hostname:          hostname,
				Region:            viper.GetString(FlagRegion),
				Datacenter:        viper.GetString(FlagDatacenter),
				Zone:              viper.GetString(FlagZone),
				Rack:              viper.GetString(FlagRack),
				Row:               viper.GetString(FlagRow),
				Position:          viper.GetString(FlagPosition),
			}

			agent := NewAgent(config)
			if err := agent.registerNode(cmd.Context()); err != nil {
				return err
			}

			fmt.Println("Node registration accepted.")
			fmt.Printf("  Node ID: %s\n", nodeID)
			fmt.Printf("  Cluster ID: %s\n", clusterID)
			fmt.Printf("  Provider: %s\n", providerAddress)
			fmt.Printf("  Public Key: %s\n", base64.StdEncoding.EncodeToString(publicKey))
			return nil
		},
	}
}

func statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show node agent status",
		Long:  `Displays current node metrics and agent status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			collector := NewMetricsCollector()
			capacity, err := collector.CollectCapacity()
			if err != nil {
				return fmt.Errorf("failed to collect capacity: %w", err)
			}

			health, err := collector.CollectHealth()
			if err != nil {
				return fmt.Errorf("failed to collect health: %w", err)
			}

			fmt.Println("Node Status")
			fmt.Println("===========")
			fmt.Println("\nCapacity:")
			fmt.Printf("  CPU Cores: %d total, %d available\n", capacity.CPUCoresTotal, capacity.CPUCoresAvailable)
			fmt.Printf("  Memory: %d GB total, %d GB available\n", capacity.MemoryGBTotal, capacity.MemoryGBAvailable)
			fmt.Printf("  GPUs: %d total, %d available\n", capacity.GPUsTotal, capacity.GPUsAvailable)
			fmt.Printf("  Storage: %d GB total, %d GB available\n", capacity.StorageGBTotal, capacity.StorageGBAvailable)

			fmt.Println("\nHealth:")
			fmt.Printf("  Status: %s\n", health.Status)
			fmt.Printf("  Uptime: %d seconds\n", health.UptimeSeconds)
			fmt.Printf("  Load Average: %s, %s, %s\n", health.LoadAverage1m, health.LoadAverage5m, health.LoadAverage15m)
			fmt.Printf("  CPU Utilization: %d%%\n", health.CPUUtilizationPercent)
			fmt.Printf("  Memory Utilization: %d%%\n", health.MemoryUtilizationPercent)

			return nil
		},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("VirtEngine HPC Node Agent")
			fmt.Println("  Version: 0.1.0")
			fmt.Println("  Features: VE-500 (Node Registration & Heartbeat)")
		},
	}
}

func loadOrGenerateKey(keyFile string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Try to load existing key
	if data, err := os.ReadFile(keyFile); err == nil {
		var keyData struct {
			PrivateKey string `json:"private_key"`
			PublicKey  string `json:"public_key"`
		}
		if err := json.Unmarshal(data, &keyData); err != nil {
			return nil, nil, fmt.Errorf("failed to parse key file: %w", err)
		}

		privateKey, err := base64.StdEncoding.DecodeString(keyData.PrivateKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode private key: %w", err)
		}

		publicKey, err := base64.StdEncoding.DecodeString(keyData.PublicKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode public key: %w", err)
		}

		return ed25519.PrivateKey(privateKey), ed25519.PublicKey(publicKey), nil
	}

	// Generate new key
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key: %w", err)
	}

	// Save key
	keyData := struct {
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	}{
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
	}

	data, err := json.MarshalIndent(keyData, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal key: %w", err)
	}

	if err := os.WriteFile(keyFile, data, 0600); err != nil {
		fmt.Printf("Warning: failed to save key file: %v\n", err)
	}

	return privateKey, publicKey, nil
}
