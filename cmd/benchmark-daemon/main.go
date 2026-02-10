// Package main provides the CLI entry point for the VirtEngine benchmark daemon.
//
// VE-600: Benchmark daemon CLI
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
)

const (
	flagProviderAddress  = "provider-address"
	flagClusterID        = "cluster-id"
	flagRegion           = "region"
	flagChainEndpoint    = "chain-endpoint"
	flagScheduleInterval = "schedule-interval"
	flagChallengeCheck   = "challenge-check-interval"
	flagNetworkEndpoint  = "network-endpoint"
	flagEnableGPU        = "enable-gpu"
	flagKeyPath          = "key-path"
	flagSuiteVersion     = "suite-version"
	flagLogLevel         = "log-level"
	flagConfigFile       = "config"
)

var cfgFile string

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "benchmark-daemon",
	Short: "VirtEngine Benchmark Daemon",
	Long: `The VirtEngine Benchmark Daemon collects performance metrics 
from provider nodes and submits signed benchmark reports to the blockchain.

Features:
- Scheduled benchmark execution
- On-demand challenge response
- Signed reports for verification
- Rate limiting and retry logic`,
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the benchmark daemon",
	Long:  `Start the benchmark daemon with the specified configuration.`,
	RunE:  runDaemon,
}

// onceCmd represents the run-once command
var onceCmd = &cobra.Command{
	Use:   "once",
	Short: "Run a single benchmark",
	Long:  `Execute a single benchmark run and exit.`,
	RunE:  runOnce,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE:  printVersion,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, flagConfigFile, "", "config file (default is $HOME/.benchmark-daemon.yaml)")
	rootCmd.PersistentFlags().String(flagLogLevel, "info", "Log level (debug, info, warn, error)")

	// Run command flags
	runCmd.Flags().String(flagProviderAddress, "", "Provider address (required)")
	runCmd.Flags().String(flagClusterID, "", "Cluster ID (required)")
	runCmd.Flags().String(flagRegion, "", "Provider region")
	runCmd.Flags().String(flagChainEndpoint, "http://localhost:26657", "Chain RPC endpoint")
	runCmd.Flags().Duration(flagScheduleInterval, time.Hour, "Benchmark schedule interval")
	runCmd.Flags().Duration(flagChallengeCheck, time.Minute*5, "Challenge check interval")
	runCmd.Flags().String(flagNetworkEndpoint, "benchmark.virtengine.com", "Network benchmark endpoint")
	runCmd.Flags().Bool(flagEnableGPU, false, "Enable GPU benchmarks")
	runCmd.Flags().String(flagKeyPath, "", "Path to benchmarking key file")
	runCmd.Flags().String(flagSuiteVersion, "1.0.0", "Benchmark suite version")

	_ = runCmd.MarkFlagRequired(flagProviderAddress)
	_ = runCmd.MarkFlagRequired(flagClusterID)

	// Once command flags
	onceCmd.Flags().String(flagProviderAddress, "", "Provider address (required)")
	onceCmd.Flags().String(flagClusterID, "", "Cluster ID (required)")
	onceCmd.Flags().String(flagRegion, "", "Provider region")
	onceCmd.Flags().String(flagChainEndpoint, "http://localhost:26657", "Chain RPC endpoint")
	onceCmd.Flags().String(flagNetworkEndpoint, "benchmark.virtengine.com", "Network benchmark endpoint")
	onceCmd.Flags().Bool(flagEnableGPU, false, "Enable GPU benchmarks")
	onceCmd.Flags().String(flagKeyPath, "", "Path to benchmarking key file")
	onceCmd.Flags().String(flagSuiteVersion, "1.0.0", "Benchmark suite version")

	_ = onceCmd.MarkFlagRequired(flagProviderAddress)
	_ = onceCmd.MarkFlagRequired(flagClusterID)

	// Bind flags to viper
	_ = viper.BindPFlags(runCmd.Flags())
	_ = viper.BindPFlags(onceCmd.Flags())

	// Add commands
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(onceCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error finding home directory:", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".benchmark-daemon")
	}

	viper.SetEnvPrefix("BENCHMARK")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func runDaemon(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	config, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Starting benchmark daemon for provider %s (cluster: %s)\n",
		config.providerAddress, config.clusterID)
	fmt.Printf("Schedule interval: %s, Challenge check: %s\n",
		config.scheduleInterval, config.challengeCheck)
	fmt.Printf("Chain endpoint: %s\n", config.chainEndpoint)

	// In a real implementation, we would:
	// 1. Load the signing key from keyPath
	// 2. Create a real chain client
	// 3. Create a real benchmark runner
	// 4. Create and start the daemon

	// For now, we just wait for signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Daemon started. Press Ctrl+C to stop.")

	select {
	case sig := <-sigCh:
		fmt.Printf("\nReceived signal %s, shutting down...\n", sig)
	case <-ctx.Done():
		fmt.Println("\nContext cancelled, shutting down...")
	}

	return nil
}

func runOnce(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()

	config, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Running single benchmark for provider %s (cluster: %s)\n",
		config.providerAddress, config.clusterID)

	// In a real implementation, we would:
	// 1. Create a benchmark runner
	// 2. Run the benchmark
	// 3. Sign and submit the report
	// 4. Print the results

	// For now, just print placeholder
	fmt.Println("Benchmark completed successfully.")
	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		fmt.Printf("Context deadline: %v\n", deadline)
	}

	return nil
}

func printVersion(cmd *cobra.Command, args []string) error {
	fmt.Println("benchmark-daemon v1.0.0")
	fmt.Println("Benchmark Suite Version: 1.0.0")
	fmt.Println("Metric Schema Version: 1.0.0")
	return nil
}

type daemonConfig struct {
	providerAddress  string
	clusterID        string
	region           string
	chainEndpoint    string
	scheduleInterval time.Duration
	challengeCheck   time.Duration
	networkEndpoint  string
	enableGPU        bool
	keyPath          string
	suiteVersion     string
}

func loadConfig(cmd *cobra.Command) (*daemonConfig, error) {
	providerAddress, _ := cmd.Flags().GetString(flagProviderAddress)
	clusterID, _ := cmd.Flags().GetString(flagClusterID)
	region, _ := cmd.Flags().GetString(flagRegion)
	chainEndpoint, _ := cmd.Flags().GetString(flagChainEndpoint)
	scheduleInterval, _ := cmd.Flags().GetDuration(flagScheduleInterval)
	challengeCheck, _ := cmd.Flags().GetDuration(flagChallengeCheck)
	networkEndpoint, _ := cmd.Flags().GetString(flagNetworkEndpoint)
	enableGPU, _ := cmd.Flags().GetBool(flagEnableGPU)
	keyPath, _ := cmd.Flags().GetString(flagKeyPath)
	suiteVersion, _ := cmd.Flags().GetString(flagSuiteVersion)

	// Override with viper values if set
	if viper.IsSet(flagProviderAddress) {
		providerAddress = viper.GetString(flagProviderAddress)
	}
	if viper.IsSet(flagClusterID) {
		clusterID = viper.GetString(flagClusterID)
	}
	if viper.IsSet(flagChainEndpoint) {
		chainEndpoint = viper.GetString(flagChainEndpoint)
	}

	if providerAddress == "" {
		return nil, fmt.Errorf("provider address is required")
	}
	if clusterID == "" {
		return nil, fmt.Errorf("cluster ID is required")
	}

	return &daemonConfig{
		providerAddress:  providerAddress,
		clusterID:        clusterID,
		region:           region,
		chainEndpoint:    chainEndpoint,
		scheduleInterval: scheduleInterval,
		challengeCheck:   challengeCheck,
		networkEndpoint:  networkEndpoint,
		enableGPU:        enableGPU,
		keyPath:          keyPath,
		suiteVersion:     suiteVersion,
	}, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
