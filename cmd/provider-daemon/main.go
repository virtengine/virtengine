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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
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

	// FlagWaldurEnabled toggles Waldur bridge
	FlagWaldurEnabled = "waldur-enabled"

	// FlagWaldurBaseURL is Waldur API base URL
	FlagWaldurBaseURL = "waldur-base-url"

	// FlagWaldurToken is Waldur API token
	FlagWaldurToken = "waldur-token" //nolint:gosec

	// FlagWaldurProjectUUID is Waldur project UUID
	FlagWaldurProjectUUID = "waldur-project-uuid"

	// FlagWaldurOfferingMap is path to offering map JSON
	FlagWaldurOfferingMap = "waldur-offering-map"

	// FlagWaldurCallbackSinkDir is directory for callback files
	FlagWaldurCallbackSinkDir = "waldur-callback-sink-dir"

	// FlagWaldurStateFile is path to state file
	FlagWaldurStateFile = "waldur-state-file"

	// FlagWaldurCheckpointFile is path to checkpoint file
	FlagWaldurCheckpointFile = "waldur-checkpoint-file"

	// FlagWaldurOrderCallbackURL is optional callback URL for Waldur order
	FlagWaldurOrderCallbackURL = "waldur-order-callback-url"

	// FlagWaldurChainSubmit enables on-chain Waldur callback submission
	FlagWaldurChainSubmit = "waldur-chain-submit"

	// FlagWaldurChainKey is the key name for on-chain Waldur callbacks
	FlagWaldurChainKey = "waldur-chain-key"

	// FlagWaldurChainKeyringBackend is the keyring backend for on-chain callbacks
	FlagWaldurChainKeyringBackend = "waldur-chain-keyring-backend"

	// FlagWaldurChainKeyringDir is the keyring dir for on-chain callbacks
	FlagWaldurChainKeyringDir = "waldur-chain-keyring-dir"

	// FlagWaldurChainKeyringPassphrase is the keyring passphrase for on-chain callbacks
	FlagWaldurChainKeyringPassphrase = "waldur-chain-keyring-passphrase" //nolint:gosec

	// FlagWaldurChainGRPC is the gRPC endpoint for on-chain callbacks
	FlagWaldurChainGRPC = "waldur-chain-grpc"

	// FlagWaldurChainGas is the gas setting for on-chain callbacks ("auto" or number)
	FlagWaldurChainGas = "waldur-chain-gas"

	// FlagWaldurChainGasPrices is the gas prices string for on-chain callbacks
	FlagWaldurChainGasPrices = "waldur-chain-gas-prices"

	// FlagWaldurChainFees is the fees string for on-chain callbacks
	FlagWaldurChainFees = "waldur-chain-fees"

	// FlagWaldurChainGasAdjustment is the gas adjustment for on-chain callbacks
	FlagWaldurChainGasAdjustment = "waldur-chain-gas-adjustment"

	// FlagWaldurChainBroadcastTimeout is the broadcast timeout for on-chain callbacks
	FlagWaldurChainBroadcastTimeout = "waldur-chain-broadcast-timeout"

	// FlagMarketplaceEventQuery is the marketplace event query
	FlagMarketplaceEventQuery = "marketplace-event-query"

	// FlagCometWS is the CometBFT websocket endpoint
	FlagCometWS = "comet-ws"

	// VE-2D: Automatic offering sync flags
	// FlagWaldurOfferingSyncEnabled enables automatic offering sync
	FlagWaldurOfferingSyncEnabled = "waldur-offering-sync-enabled"

	// FlagWaldurOfferingSyncStateFile is the path for offering sync state
	FlagWaldurOfferingSyncStateFile = "waldur-offering-sync-state-file"

	// FlagWaldurCustomerUUID is the Waldur customer/org UUID for offerings
	FlagWaldurCustomerUUID = "waldur-customer-uuid"

	// FlagWaldurCategoryMap is path to category map JSON
	FlagWaldurCategoryMap = "waldur-category-map"

	// FlagWaldurOfferingSyncInterval is the reconciliation interval in seconds
	FlagWaldurOfferingSyncInterval = "waldur-offering-sync-interval"

	// FlagWaldurOfferingSyncMaxRetries is max retries before dead-letter
	FlagWaldurOfferingSyncMaxRetries = "waldur-offering-sync-max-retries"

	// Portal API flags
	FlagPortalAuthSecret      = "portal-auth-secret"
	FlagPortalAllowInsecure   = "portal-allow-insecure"
	FlagPortalRequireVEID     = "portal-require-veid"
	FlagPortalMinVEIDScore    = "portal-min-veid-score"
	FlagPortalShellSessionTTL = "portal-shell-session-ttl"
	FlagPortalTokenTTL        = "portal-token-ttl"
	FlagPortalAuditLogFile    = "portal-audit-log-file"
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
	rootCmd.PersistentFlags().Bool(FlagWaldurEnabled, false, "Enable Waldur provider bridge")
	rootCmd.PersistentFlags().String(FlagWaldurBaseURL, "", "Waldur API base URL")
	rootCmd.PersistentFlags().String(FlagWaldurToken, "", "Waldur API token")
	rootCmd.PersistentFlags().String(FlagWaldurProjectUUID, "", "Waldur project UUID")
	rootCmd.PersistentFlags().String(FlagWaldurOfferingMap, "", "Path to Waldur offering map JSON (DEPRECATED: use --waldur-offering-sync-enabled)")
	rootCmd.PersistentFlags().String(FlagWaldurCallbackSinkDir, "data/callbacks", "Directory for Waldur callback files")
	rootCmd.PersistentFlags().String(FlagWaldurStateFile, "data/waldur_bridge_state.json", "Waldur bridge state file path")
	rootCmd.PersistentFlags().String(FlagWaldurCheckpointFile, "data/marketplace_checkpoint.json", "Marketplace checkpoint file path")
	rootCmd.PersistentFlags().String(FlagWaldurOrderCallbackURL, "", "Callback URL to include in Waldur order")
	rootCmd.PersistentFlags().Bool(FlagWaldurChainSubmit, false, "Submit Waldur callbacks on-chain via MsgWaldurCallback")
	rootCmd.PersistentFlags().String(FlagWaldurChainKey, "", "Key name for on-chain Waldur callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainKeyringBackend, "test", "Keyring backend for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainKeyringDir, "", "Keyring directory for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainKeyringPassphrase, "", "Keyring passphrase for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainGRPC, "localhost:9090", "gRPC endpoint for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainGas, "auto", "Gas setting for on-chain callback submissions (auto or number)")
	rootCmd.PersistentFlags().String(FlagWaldurChainGasPrices, "", "Gas prices for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagWaldurChainFees, "", "Fees for on-chain callback submissions")
	rootCmd.PersistentFlags().Float64(FlagWaldurChainGasAdjustment, 1.2, "Gas adjustment for on-chain callback submissions")
	rootCmd.PersistentFlags().Duration(FlagWaldurChainBroadcastTimeout, 30*time.Second, "Broadcast timeout for on-chain callback submissions")
	rootCmd.PersistentFlags().String(FlagMarketplaceEventQuery, "", "Marketplace event query for CometBFT subscription")
	rootCmd.PersistentFlags().String(FlagCometWS, "/websocket", "CometBFT websocket endpoint path")

	// VE-2D: Automatic offering sync flags
	rootCmd.PersistentFlags().Bool(FlagWaldurOfferingSyncEnabled, false, "Enable automatic offering sync from chain to Waldur (replaces manual offering map)")
	rootCmd.PersistentFlags().String(FlagWaldurOfferingSyncStateFile, "data/offering_sync_state.json", "Path for offering sync state file")
	rootCmd.PersistentFlags().String(FlagWaldurCustomerUUID, "", "Waldur customer/organization UUID for creating offerings")
	rootCmd.PersistentFlags().String(FlagWaldurCategoryMap, "", "Path to JSON file mapping offering categories to Waldur category UUIDs")
	rootCmd.PersistentFlags().Int64(FlagWaldurOfferingSyncInterval, 300, "Offering sync reconciliation interval in seconds")
	rootCmd.PersistentFlags().Int(FlagWaldurOfferingSyncMaxRetries, 5, "Max sync retries before dead-lettering")

	// Portal API flags
	rootCmd.PersistentFlags().String(FlagPortalAuthSecret, "", "Shared secret for portal signed requests")
	rootCmd.PersistentFlags().Bool(FlagPortalAllowInsecure, true, "Allow portal requests without signature (dev only)")
	rootCmd.PersistentFlags().Bool(FlagPortalRequireVEID, true, "Require VEID verification for shell access")
	rootCmd.PersistentFlags().Int(FlagPortalMinVEIDScore, 80, "Minimum VEID score required for shell access")
	rootCmd.PersistentFlags().Duration(FlagPortalShellSessionTTL, 10*time.Minute, "Shell session TTL for portal access")
	rootCmd.PersistentFlags().Duration(FlagPortalTokenTTL, 5*time.Minute, "Portal session token TTL")
	rootCmd.PersistentFlags().String(FlagPortalAuditLogFile, "data/portal_audit.log", "Portal audit log file path")

	// Bind to viper
	_ = viper.BindPFlag(FlagChainID, rootCmd.PersistentFlags().Lookup(FlagChainID))
	_ = viper.BindPFlag(FlagNode, rootCmd.PersistentFlags().Lookup(FlagNode))
	_ = viper.BindPFlag(FlagProviderKey, rootCmd.PersistentFlags().Lookup(FlagProviderKey))
	_ = viper.BindPFlag(FlagProviderKeyDir, rootCmd.PersistentFlags().Lookup(FlagProviderKeyDir))
	_ = viper.BindPFlag(FlagListenAddr, rootCmd.PersistentFlags().Lookup(FlagListenAddr))
	_ = viper.BindPFlag(FlagMetricsAddr, rootCmd.PersistentFlags().Lookup(FlagMetricsAddr))
	_ = viper.BindPFlag(FlagWaldurEnabled, rootCmd.PersistentFlags().Lookup(FlagWaldurEnabled))
	_ = viper.BindPFlag(FlagWaldurBaseURL, rootCmd.PersistentFlags().Lookup(FlagWaldurBaseURL))
	_ = viper.BindPFlag(FlagWaldurToken, rootCmd.PersistentFlags().Lookup(FlagWaldurToken))
	_ = viper.BindPFlag(FlagWaldurProjectUUID, rootCmd.PersistentFlags().Lookup(FlagWaldurProjectUUID))
	_ = viper.BindPFlag(FlagWaldurOfferingMap, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingMap))
	_ = viper.BindPFlag(FlagWaldurCallbackSinkDir, rootCmd.PersistentFlags().Lookup(FlagWaldurCallbackSinkDir))
	_ = viper.BindPFlag(FlagWaldurStateFile, rootCmd.PersistentFlags().Lookup(FlagWaldurStateFile))
	_ = viper.BindPFlag(FlagWaldurCheckpointFile, rootCmd.PersistentFlags().Lookup(FlagWaldurCheckpointFile))
	_ = viper.BindPFlag(FlagWaldurOrderCallbackURL, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderCallbackURL))
	_ = viper.BindPFlag(FlagWaldurChainSubmit, rootCmd.PersistentFlags().Lookup(FlagWaldurChainSubmit))
	_ = viper.BindPFlag(FlagWaldurChainKey, rootCmd.PersistentFlags().Lookup(FlagWaldurChainKey))
	_ = viper.BindPFlag(FlagWaldurChainKeyringBackend, rootCmd.PersistentFlags().Lookup(FlagWaldurChainKeyringBackend))
	_ = viper.BindPFlag(FlagWaldurChainKeyringDir, rootCmd.PersistentFlags().Lookup(FlagWaldurChainKeyringDir))
	_ = viper.BindPFlag(FlagWaldurChainKeyringPassphrase, rootCmd.PersistentFlags().Lookup(FlagWaldurChainKeyringPassphrase))
	_ = viper.BindPFlag(FlagWaldurChainGRPC, rootCmd.PersistentFlags().Lookup(FlagWaldurChainGRPC))
	_ = viper.BindPFlag(FlagWaldurChainGas, rootCmd.PersistentFlags().Lookup(FlagWaldurChainGas))
	_ = viper.BindPFlag(FlagWaldurChainGasPrices, rootCmd.PersistentFlags().Lookup(FlagWaldurChainGasPrices))
	_ = viper.BindPFlag(FlagWaldurChainFees, rootCmd.PersistentFlags().Lookup(FlagWaldurChainFees))
	_ = viper.BindPFlag(FlagWaldurChainGasAdjustment, rootCmd.PersistentFlags().Lookup(FlagWaldurChainGasAdjustment))
	_ = viper.BindPFlag(FlagWaldurChainBroadcastTimeout, rootCmd.PersistentFlags().Lookup(FlagWaldurChainBroadcastTimeout))
	_ = viper.BindPFlag(FlagMarketplaceEventQuery, rootCmd.PersistentFlags().Lookup(FlagMarketplaceEventQuery))
	_ = viper.BindPFlag(FlagCometWS, rootCmd.PersistentFlags().Lookup(FlagCometWS))

	// VE-2D: Bind offering sync flags
	_ = viper.BindPFlag(FlagWaldurOfferingSyncEnabled, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingSyncEnabled))
	_ = viper.BindPFlag(FlagWaldurOfferingSyncStateFile, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingSyncStateFile))
	_ = viper.BindPFlag(FlagWaldurCustomerUUID, rootCmd.PersistentFlags().Lookup(FlagWaldurCustomerUUID))
	_ = viper.BindPFlag(FlagWaldurCategoryMap, rootCmd.PersistentFlags().Lookup(FlagWaldurCategoryMap))
	_ = viper.BindPFlag(FlagWaldurOfferingSyncInterval, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingSyncInterval))
	_ = viper.BindPFlag(FlagWaldurOfferingSyncMaxRetries, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingSyncMaxRetries))

	// Portal API flags
	_ = viper.BindPFlag(FlagPortalAuthSecret, rootCmd.PersistentFlags().Lookup(FlagPortalAuthSecret))
	_ = viper.BindPFlag(FlagPortalAllowInsecure, rootCmd.PersistentFlags().Lookup(FlagPortalAllowInsecure))
	_ = viper.BindPFlag(FlagPortalRequireVEID, rootCmd.PersistentFlags().Lookup(FlagPortalRequireVEID))
	_ = viper.BindPFlag(FlagPortalMinVEIDScore, rootCmd.PersistentFlags().Lookup(FlagPortalMinVEIDScore))
	_ = viper.BindPFlag(FlagPortalShellSessionTTL, rootCmd.PersistentFlags().Lookup(FlagPortalShellSessionTTL))
	_ = viper.BindPFlag(FlagPortalTokenTTL, rootCmd.PersistentFlags().Lookup(FlagPortalTokenTTL))
	_ = viper.BindPFlag(FlagPortalAuditLogFile, rootCmd.PersistentFlags().Lookup(FlagPortalAuditLogFile))

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

	_ = viper.BindPFlag(FlagKubeconfig, cmd.Flags().Lookup(FlagKubeconfig))
	_ = viper.BindPFlag(FlagMeteringInterval, cmd.Flags().Lookup(FlagMeteringInterval))
	_ = viper.BindPFlag(FlagBidRateLimitMinute, cmd.Flags().Lookup(FlagBidRateLimitMinute))
	_ = viper.BindPFlag(FlagBidRateLimitHour, cmd.Flags().Lookup(FlagBidRateLimitHour))
	_ = viper.BindPFlag(FlagResourcePrefix, cmd.Flags().Lookup(FlagResourcePrefix))

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
	keyDir := viper.GetString(FlagProviderKeyDir)
	keyConfig := provider_daemon.DefaultKeyManagerConfig()
	keyConfig.KeyDir = keyDir
	if keyDir == "" {
		keyConfig.StorageType = provider_daemon.KeyStorageTypeMemory
	} else {
		keyConfig.StorageType = provider_daemon.KeyStorageTypeFile
	}
	keyManager, err := provider_daemon.NewKeyManager(keyConfig)
	if err != nil {
		return fmt.Errorf("failed to create key manager: %w", err)
	}

	// Unlock key manager (use empty passphrase for memory storage)
	if err := keyManager.Unlock(""); err != nil {
		return fmt.Errorf("failed to unlock key manager: %w", err)
	}
	fmt.Println("  Key Manager: initialized")

	// Generate provider key
	providerKeyName := viper.GetString(FlagProviderKey)
	providerAddress := providerKeyName
	key, err := keyManager.GenerateKey(providerKeyName)
	if err != nil {
		return fmt.Errorf("failed to generate provider key: %w", err)
	}
	providerID := key.PublicKey
	fmt.Printf("  Provider ID: %s...\n", providerID[:16])

	var callbackSink provider_daemon.CallbackSink
	var usageReporter provider_daemon.UsageReporter

	if viper.GetBool(FlagWaldurEnabled) && viper.GetBool(FlagWaldurChainSubmit) {
		chainKeyName := viper.GetString(FlagWaldurChainKey)
		if chainKeyName == "" {
			chainKeyName = providerKeyName
		}

		gasSetting, err := parseGasSetting(viper.GetString(FlagWaldurChainGas))
		if err != nil {
			return fmt.Errorf("invalid waldur chain gas: %w", err)
		}

		chainCfg := provider_daemon.ChainCallbackSinkConfig{
			ChainID:           viper.GetString(FlagChainID),
			NodeURI:           viper.GetString(FlagNode),
			GRPCEndpoint:      viper.GetString(FlagWaldurChainGRPC),
			KeyName:           chainKeyName,
			KeyringBackend:    viper.GetString(FlagWaldurChainKeyringBackend),
			KeyringDir:        viper.GetString(FlagWaldurChainKeyringDir),
			KeyringPassphrase: viper.GetString(FlagWaldurChainKeyringPassphrase),
			GasSetting:        gasSetting,
			GasPrices:         viper.GetString(FlagWaldurChainGasPrices),
			Fees:              viper.GetString(FlagWaldurChainFees),
			GasAdjustment:     viper.GetFloat64(FlagWaldurChainGasAdjustment),
			BroadcastTimeout:  viper.GetDuration(FlagWaldurChainBroadcastTimeout),
		}

		sink, err := provider_daemon.NewChainCallbackSink(ctx, chainCfg)
		if err != nil {
			return fmt.Errorf("failed to create on-chain callback sink: %w", err)
		}
		callbackSink = sink
		providerAddress = sink.SenderAddress()
	}

	if viper.GetBool(FlagWaldurEnabled) {
		if _, err := sdk.AccAddressFromBech32(providerAddress); err != nil {
			return fmt.Errorf("provider address must be bech32 when Waldur is enabled: %w", err)
		}
	}

	// Initialize bid engine (VE-401)
	bidEngineConfig := provider_daemon.BidEngineConfig{
		ProviderAddress:    providerAddress,
		MaxBidsPerMinute:   viper.GetInt(FlagBidRateLimitMinute),
		MaxBidsPerHour:     viper.GetInt(FlagBidRateLimitHour),
		MaxConcurrentBids:  5,
		BidRetryDelay:      time.Second * 5,
		MaxBidRetries:      3,
		ConfigPollInterval: time.Second * 30,
		OrderPollInterval:  time.Second * 5,
	}

	// Create chain client for bid engine
	chainClient, err := provider_daemon.NewRPCChainClient(ctx, provider_daemon.RPCChainClientConfig{
		NodeURI:        viper.GetString(FlagNode),
		GRPCEndpoint:   viper.GetString(FlagWaldurChainGRPC),
		ChainID:        viper.GetString(FlagChainID),
		RequestTimeout: time.Second * 30,
	})
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}

	// Initialize Event Stream (PROVIDER-STREAM-001)
	var eventSubscriber provider_daemon.EventSubscriber

	// Create checkpoint store
	checkpointStore, err := provider_daemon.NewEventCheckpointStore(viper.GetString(FlagWaldurCheckpointFile))
	if err != nil {
		fmt.Printf("Warning: Failed to create checkpoint store: %v\n", err)
	}

	// Configure event stream
	streamCfg := provider_daemon.DefaultEventSubscriberConfig()
	streamCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
	streamCfg.CometWS = viper.GetString(FlagCometWS)
	streamCfg.CheckpointStore = checkpointStore
	streamCfg.SubscriberID = fmt.Sprintf("provider-%s", providerID[:8])

	// Create subscriber
	sub, err := provider_daemon.NewCometEventSubscriber(streamCfg)
	if err != nil {
		fmt.Printf("Warning: Failed to create event subscriber: %v. Falling back to polling.\n", err)
	} else {
		eventSubscriber = sub
		fmt.Println("  Event Stream: initialized")
	}

	var bidEngine *provider_daemon.BidEngine
	if eventSubscriber != nil {
		bidEngine = provider_daemon.NewBidEngineWithStreaming(bidEngineConfig, keyManager, chainClient, eventSubscriber)
	} else {
		bidEngine = provider_daemon.NewBidEngine(bidEngineConfig, keyManager, chainClient)
	}

	if err := bidEngine.Start(ctx); err != nil {
		return fmt.Errorf("failed to start bid engine: %w", err)
	}
	fmt.Println("  Bid Engine: started")

	// Initialize Kubernetes adapter (VE-403)
	statusUpdateChan := make(chan provider_daemon.WorkloadStatusUpdate, 100)

	// Note: In production, this would use a real Kubernetes client
	// For now, we'll use a placeholder that demonstrates the integration
	fmt.Println("  Kubernetes Adapter: initialized (placeholder)")

	// Initialize usage meter (VE-404)
	recordChan := make(chan *provider_daemon.UsageRecord, 100)
	usageStore := provider_daemon.NewUsageSnapshotStore()
	usageReporter = usageStore

	usageMeter := provider_daemon.NewUsageMeter(provider_daemon.UsageMeterConfig{
		ProviderID: providerID,
		Interval:   provider_daemon.MeteringInterval(viper.GetDuration(FlagMeteringInterval)),
		KeyManager: keyManager,
		RecordChan: recordChan,
	})

	if err := usageMeter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start usage meter: %w", err)
	}
	fmt.Println("  Usage Meter: started")

	portalAuditCfg := provider_daemon.DefaultAuditLogConfig()
	portalAuditCfg.LogFile = viper.GetString(FlagPortalAuditLogFile)
	portalAuditLogger, err := provider_daemon.NewAuditLogger(portalAuditCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize portal audit logger: %w", err)
	}
	defer portalAuditLogger.Close()

	portalCfg := provider_daemon.DefaultPortalAPIServerConfig()
	portalCfg.ListenAddr = viper.GetString(FlagListenAddr)
	portalCfg.AuthSecret = viper.GetString(FlagPortalAuthSecret)
	portalCfg.AllowInsecure = viper.GetBool(FlagPortalAllowInsecure)
	portalCfg.RequireVEID = viper.GetBool(FlagPortalRequireVEID)
	portalCfg.MinVEIDScore = viper.GetInt(FlagPortalMinVEIDScore)
	portalCfg.ShellSessionTTL = viper.GetDuration(FlagPortalShellSessionTTL)
	portalCfg.TokenTTL = viper.GetDuration(FlagPortalTokenTTL)
	portalCfg.AuditLogger = portalAuditLogger

	portalAPI, err := provider_daemon.NewPortalAPIServer(portalCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize portal API server: %w", err)
	}
	go func() {
		if err := portalAPI.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("[PORTAL] API server stopped: %v\n", err)
		}
	}()
	fmt.Println("  Portal API: started")

	// Initialize Waldur bridge (VE-2040+)
	if viper.GetBool(FlagWaldurEnabled) {
		offeringMap, err := loadOfferingMap(viper.GetString(FlagWaldurOfferingMap))
		if err != nil {
			return fmt.Errorf("failed to load waldur offering map: %w", err)
		}

		bridgeCfg := provider_daemon.DefaultWaldurBridgeConfig()
		bridgeCfg.Enabled = true
		bridgeCfg.ProviderAddress = providerAddress
		bridgeCfg.ProviderID = providerID
		bridgeCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
		bridgeCfg.CometWS = viper.GetString(FlagCometWS)
		bridgeCfg.EventQuery = viper.GetString(FlagMarketplaceEventQuery)
		bridgeCfg.CallbackSinkDir = viper.GetString(FlagWaldurCallbackSinkDir)
		bridgeCfg.StateFile = viper.GetString(FlagWaldurStateFile)
		bridgeCfg.CheckpointFile = viper.GetString(FlagWaldurCheckpointFile)
		bridgeCfg.WaldurBaseURL = viper.GetString(FlagWaldurBaseURL)
		bridgeCfg.WaldurToken = viper.GetString(FlagWaldurToken)
		bridgeCfg.WaldurProjectUUID = viper.GetString(FlagWaldurProjectUUID)
		bridgeCfg.WaldurOfferingMap = offeringMap
		bridgeCfg.OrderCallbackURL = viper.GetString(FlagWaldurOrderCallbackURL)

		waldurBridge, err := provider_daemon.NewWaldurBridge(bridgeCfg, keyManager, callbackSink, usageReporter)
		if err != nil {
			return fmt.Errorf("failed to create waldur bridge: %w", err)
		}

		go func() {
			if err := waldurBridge.Start(ctx); err != nil {
				fmt.Printf("[WALDUR] bridge stopped: %v\n", err)
			}
		}()
		fmt.Println("  Waldur Bridge: started")
	}

	// Start background workers
	bidResultChan := bidEngine.GetBidResults()
	go handleBidResults(ctx, bidResultChan)
	go handleStatusUpdates(ctx, statusUpdateChan)
	go handleUsageRecords(ctx, recordChan, usageStore)

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

func handleUsageRecords(ctx context.Context, ch <-chan *provider_daemon.UsageRecord, usageStore *provider_daemon.UsageSnapshotStore) {
	for {
		select {
		case <-ctx.Done():
			return
		case record := <-ch:
			if usageStore != nil {
				usageStore.Track(record)
			}
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

			keyConfig := provider_daemon.KeyManagerConfig{
				StorageType: provider_daemon.KeyStorageTypeFile,
				KeyDir:      keyDir,
			}
			keyManager, err := provider_daemon.NewKeyManager(keyConfig)
			if err != nil {
				return fmt.Errorf("failed to create key manager: %w", err)
			}

			if err := keyManager.Unlock(""); err != nil {
				return fmt.Errorf("failed to unlock key manager: %w", err)
			}

			key, err := keyManager.GenerateKey(keyName)
			if err != nil {
				return fmt.Errorf("failed to generate key: %w", err)
			}

			fmt.Printf("Generated key '%s'\n", keyName)
			fmt.Printf("  Key ID: %s\n", key.KeyID)
			fmt.Printf("  Public Key: %s\n", key.PublicKey)

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

			//nolint:gosec // G304: filePath is a user-provided CLI argument for manifest validation
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

func loadOfferingMap(path string) (map[string]string, error) {
	if path == "" {
		return map[string]string{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]string{}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func parseGasSetting(value string) (provider_daemon.GasSetting, error) {
	if strings.EqualFold(value, "auto") || value == "" {
		return provider_daemon.GasSetting{Simulate: true}, nil
	}
	gas, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return provider_daemon.GasSetting{}, err
	}
	return provider_daemon.GasSetting{Gas: gas}, nil
}

func normalizeCometRPC(node string) string {
	if strings.HasPrefix(node, "tcp://") {
		return "http://" + strings.TrimPrefix(node, "tcp://")
	}
	if strings.HasPrefix(node, "http://") || strings.HasPrefix(node, "https://") {
		return node
	}
	return "http://" + node
}
