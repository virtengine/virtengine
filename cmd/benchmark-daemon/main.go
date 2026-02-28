// Package main provides the CLI entry point for the VirtEngine benchmark daemon.
//
// VE-600: Benchmark daemon CLI
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	benchmarkdaemon "github.com/virtengine/virtengine/pkg/benchmark_daemon"
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

	config, err := loadConfig(cmd)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	return executeDaemon(ctx, config, os.Stdout)
}

func executeDaemon(ctx context.Context, config *daemonConfig, out io.Writer) error {
	daemonCfg := buildBenchmarkDaemonConfig(config)
	runner := benchmarkdaemon.NewDefaultBenchmarkRunner()
	daemon, err := benchmarkdaemon.NewBenchmarkDaemon(daemonCfg, nil, runner, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize benchmark daemon: %w", err)
	}
	if err := daemon.Start(ctx); err != nil {
		return fmt.Errorf("failed to start benchmark daemon: %w", err)
	}
	defer daemon.Stop()

	fmt.Fprintf(out, "Starting benchmark daemon for provider %s (cluster: %s)\n",
		config.providerAddress, config.clusterID)
	fmt.Fprintf(out, "Schedule interval: %s, Challenge check: %s\n",
		config.scheduleInterval, config.challengeCheck)
	fmt.Fprintf(out, "Chain endpoint: %s\n", config.chainEndpoint)
	fmt.Fprintln(out, "Chain submission disabled in standalone CLI mode; local benchmark reports will not be broadcast.")

	reporterCtx, reporterCancel := context.WithCancel(ctx)
	defer reporterCancel()
	go streamDaemonResults(reporterCtx, daemon, out)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	fmt.Fprintln(out, "Daemon started. Press Ctrl+C to stop.")

	select {
	case sig := <-sigCh:
		fmt.Fprintf(out, "\nReceived signal %s, shutting down...\n", sig)
	case <-ctx.Done():
		fmt.Fprintln(out, "\nContext cancelled, shutting down...")
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

	return executeSingleBenchmark(ctx, config, os.Stdout)
}

func executeSingleBenchmark(ctx context.Context, config *daemonConfig, out io.Writer) error {
	fmt.Fprintf(out, "Running single benchmark for provider %s (cluster: %s)\n",
		config.providerAddress, config.clusterID)
	fmt.Fprintln(out, "Chain submission disabled in standalone CLI mode; printing the local benchmark report.")

	result, err := benchmarkdaemon.RunLocalBenchmark(ctx, buildBenchmarkDaemonConfig(config), benchmarkdaemon.NewDefaultBenchmarkRunner())
	if err != nil {
		if result != nil {
			_ = printBenchmarkResult(out, result)
		}
		return fmt.Errorf("benchmark execution failed: %w", err)
	}

	return printBenchmarkResult(out, result)
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

func buildBenchmarkDaemonConfig(config *daemonConfig) benchmarkdaemon.BenchmarkDaemonConfig {
	daemonCfg := benchmarkdaemon.DefaultBenchmarkDaemonConfig()
	daemonCfg.ProviderAddress = config.providerAddress
	daemonCfg.ClusterID = config.clusterID
	daemonCfg.Region = config.region
	daemonCfg.ChainEndpoint = config.chainEndpoint
	daemonCfg.ScheduleInterval = config.scheduleInterval
	daemonCfg.ChallengeCheckInterval = config.challengeCheck
	daemonCfg.EnableGPU = config.enableGPU
	daemonCfg.NetworkReferenceEndpoint = config.networkEndpoint
	daemonCfg.SuiteVersion = config.suiteVersion

	return daemonCfg
}

func printBenchmarkResult(out io.Writer, result *benchmarkdaemon.BenchmarkResult) error {
	payload, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode benchmark result: %w", err)
	}

	_, err = fmt.Fprintln(out, string(payload))
	return err
}

func streamDaemonResults(ctx context.Context, daemon *benchmarkdaemon.BenchmarkDaemon, out io.Writer) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var lastReportID string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, ok := daemon.GetLatestResult()
			if !ok || result.ReportID == "" || result.ReportID == lastReportID {
				continue
			}

			if err := printBenchmarkResult(out, result); err == nil {
				lastReportID = result.ReportID
			}
		}
	}
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
