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
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/virtengine/virtengine/pkg/observability"
	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/servicedesk"
	"github.com/virtengine/virtengine/pkg/waldur"
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

	// FlagTracingEnabled enables distributed tracing
	FlagTracingEnabled = "tracing-enabled"

	// FlagTracingEndpoint is the OTLP endpoint for tracing
	FlagTracingEndpoint = "tracing-endpoint"

	// FlagTracingSampleRate is the trace sampling rate
	FlagTracingSampleRate = "tracing-sample-rate"

	// FlagTracingEnvironment sets the deployment environment
	FlagTracingEnvironment = "tracing-environment"

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

	// FlagWaldurOrderRoutingEnabled enables chain->Waldur order routing
	FlagWaldurOrderRoutingEnabled = "waldur-order-routing-enabled"

	// FlagWaldurOrderStateFile is the order routing state file path
	FlagWaldurOrderStateFile = "waldur-order-state-file"

	// FlagWaldurOrderCheckpointFile is the order routing checkpoint file path
	FlagWaldurOrderCheckpointFile = "waldur-order-checkpoint-file"

	// FlagWaldurOrderCallbackListen is the listen address for order status callbacks
	FlagWaldurOrderCallbackListen = "waldur-order-callback-listen"

	// FlagWaldurOrderCallbackPath is the callback path for order status callbacks
	FlagWaldurOrderCallbackPath = "waldur-order-callback-path"

	// FlagWaldurLifecycleCallbackURL is callback URL for lifecycle operations
	FlagWaldurLifecycleCallbackURL = "waldur-lifecycle-callback-url"

	// FlagWaldurLifecycleCallbackListen is the listen address for lifecycle callbacks
	FlagWaldurLifecycleCallbackListen = "waldur-lifecycle-callback-listen"

	// FlagWaldurLifecycleCallbackPath is the callback path for lifecycle callbacks
	FlagWaldurLifecycleCallbackPath = "waldur-lifecycle-callback-path"

	// FlagWaldurLifecycleRequireConsent toggles consent enforcement for lifecycle actions
	FlagWaldurLifecycleRequireConsent = "waldur-lifecycle-require-consent"

	// FlagWaldurLifecycleConsentScope sets the consent scope for lifecycle actions
	FlagWaldurLifecycleConsentScope = "waldur-lifecycle-consent-scope"

	// FlagWaldurLifecycleAllowedRoles sets allowed roles for lifecycle actions (comma-separated)
	FlagWaldurLifecycleAllowedRoles = "waldur-lifecycle-allowed-roles"

	// FlagWaldurOrderRoutingMaxRetries is max retries for order routing
	FlagWaldurOrderRoutingMaxRetries = "waldur-order-routing-max-retries"

	// FlagWaldurOrderRoutingWorkers is worker count for order routing
	FlagWaldurOrderRoutingWorkers = "waldur-order-routing-workers"

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

	// FlagWaldurLifecycleQueueEnabled enables the lifecycle command queue
	FlagWaldurLifecycleQueueEnabled = "waldur-lifecycle-queue-enabled"

	// FlagWaldurLifecycleQueueBackend sets lifecycle queue storage backend
	FlagWaldurLifecycleQueueBackend = "waldur-lifecycle-queue-backend"

	// FlagWaldurLifecycleQueuePath sets lifecycle queue storage path
	FlagWaldurLifecycleQueuePath = "waldur-lifecycle-queue-path"

	// FlagWaldurLifecycleQueueWorkers sets lifecycle queue worker count
	FlagWaldurLifecycleQueueWorkers = "waldur-lifecycle-queue-workers"

	// FlagWaldurLifecycleQueueMaxRetries sets lifecycle queue max retries
	FlagWaldurLifecycleQueueMaxRetries = "waldur-lifecycle-queue-max-retries"

	// FlagWaldurLifecycleQueueRetryBackoff sets lifecycle queue retry backoff
	FlagWaldurLifecycleQueueRetryBackoff = "waldur-lifecycle-queue-retry-backoff"

	// FlagWaldurLifecycleQueueMaxBackoff sets lifecycle queue max backoff
	FlagWaldurLifecycleQueueMaxBackoff = "waldur-lifecycle-queue-max-backoff"

	// FlagWaldurLifecycleQueuePollInterval sets lifecycle queue poll interval
	FlagWaldurLifecycleQueuePollInterval = "waldur-lifecycle-queue-poll-interval"

	// FlagWaldurLifecycleQueueReconcileInterval sets lifecycle queue reconcile interval
	FlagWaldurLifecycleQueueReconcileInterval = "waldur-lifecycle-queue-reconcile-interval"

	// FlagWaldurLifecycleQueueReconcileOnStart toggles reconciliation on startup
	FlagWaldurLifecycleQueueReconcileOnStart = "waldur-lifecycle-queue-reconcile-on-start"

	// FlagWaldurLifecycleQueueStaleAfter sets stale executing command threshold
	FlagWaldurLifecycleQueueStaleAfter = "waldur-lifecycle-queue-stale-after"

	// Portal API flags
	FlagPortalAuthSecret      = "portal-auth-secret" // #nosec G101 -- flag name, not a credential
	FlagPortalAllowInsecure   = "portal-allow-insecure"
	FlagPortalRequireVEID     = "portal-require-veid"
	FlagPortalMinVEIDScore    = "portal-min-veid-score"
	FlagPortalShellSessionTTL = "portal-shell-session-ttl"
	FlagPortalTokenTTL        = "portal-token-ttl" // #nosec G101 -- flag name, not a credential
	FlagPortalAuditLogFile    = "portal-audit-log-file"

	// Vault flags
	FlagVaultEnabled          = "vault-enabled"
	FlagVaultBackend          = "vault-backend"
	FlagVaultAuditOwner       = "vault-audit-owner"
	FlagVaultRotateOverlap    = "vault-rotate-overlap"
	FlagVaultAnomalyWindow    = "vault-anomaly-window"
	FlagVaultAnomalyThreshold = "vault-anomaly-threshold"

	// Support service desk flags
	FlagSupportEnabled             = "support-enabled"
	FlagSupportWaldurBaseURL       = "support-waldur-base-url"
	FlagSupportWaldurToken         = "support-waldur-token" //nolint:gosec
	FlagSupportWaldurOrgUUID       = "support-waldur-org-uuid"
	FlagSupportWaldurProjectUUID   = "support-waldur-project-uuid"
	FlagSupportWebhookSecret       = "support-webhook-secret" //nolint:gosec
	FlagSupportWebhookListen       = "support-webhook-listen"
	FlagSupportWebhookRequireSig   = "support-webhook-require-signature"
	FlagSupportDecryptionKeyPath   = "support-decryption-key-path"
	FlagSupportDecryptionKeyBase64 = "support-decryption-key-base64"
	FlagSupportEncryptionKeyPath   = "support-encryption-key-path"
	FlagSupportEncryptionKeyBase64 = "support-encryption-key-base64"
	FlagSupportSyncInbound         = "support-sync-inbound"
	FlagSupportSyncOutbound        = "support-sync-outbound"
	FlagSupportSyncInterval        = "support-sync-interval"
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
	rootCmd.PersistentFlags().Bool(FlagTracingEnabled, false, "Enable distributed tracing")
	rootCmd.PersistentFlags().String(FlagTracingEndpoint, "localhost:4317", "OTLP gRPC endpoint for traces")
	rootCmd.PersistentFlags().Float64(FlagTracingSampleRate, 0.1, "Trace sampling rate (0.0-1.0)")
	rootCmd.PersistentFlags().String(FlagTracingEnvironment, "development", "Deployment environment for tracing")
	rootCmd.PersistentFlags().Bool(FlagWaldurEnabled, false, "Enable Waldur provider bridge")
	rootCmd.PersistentFlags().String(FlagWaldurBaseURL, "", "Waldur API base URL")
	rootCmd.PersistentFlags().String(FlagWaldurToken, "", "Waldur API token")
	rootCmd.PersistentFlags().String(FlagWaldurProjectUUID, "", "Waldur project UUID")
	rootCmd.PersistentFlags().String(FlagWaldurOfferingMap, "", "Path to Waldur offering map JSON (DEPRECATED: use --waldur-offering-sync-enabled)")
	rootCmd.PersistentFlags().String(FlagWaldurCallbackSinkDir, "data/callbacks", "Directory for Waldur callback files")
	rootCmd.PersistentFlags().String(FlagWaldurStateFile, "data/waldur_bridge_state.json", "Waldur bridge state file path")
	rootCmd.PersistentFlags().String(FlagWaldurCheckpointFile, "data/marketplace_checkpoint.json", "Marketplace checkpoint file path")
	rootCmd.PersistentFlags().String(FlagWaldurOrderCallbackURL, "", "Callback URL to include in Waldur order")
	rootCmd.PersistentFlags().Bool(FlagWaldurOrderRoutingEnabled, true, "Enable routing customer orders to Waldur")
	rootCmd.PersistentFlags().String(FlagWaldurOrderStateFile, "data/waldur_order_state.json", "Waldur order routing state file path")
	rootCmd.PersistentFlags().String(FlagWaldurOrderCheckpointFile, "data/waldur_order_checkpoint.json", "Order routing checkpoint file path")
	rootCmd.PersistentFlags().String(FlagWaldurOrderCallbackListen, ":8444", "Listen address for Waldur order status callbacks")
	rootCmd.PersistentFlags().String(FlagWaldurOrderCallbackPath, "/v1/callbacks/waldur/orders", "HTTP path for Waldur order status callbacks")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleCallbackURL, "", "Callback URL to include in Waldur lifecycle actions")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleCallbackListen, ":8445", "Listen address for Waldur lifecycle callbacks")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleCallbackPath, "/v1/callbacks/waldur", "Base HTTP path for Waldur callbacks (lifecycle is /lifecycle)")
	rootCmd.PersistentFlags().Bool(FlagWaldurLifecycleRequireConsent, true, "Require consent for lifecycle actions")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleConsentScope, "marketplace:lifecycle", "Consent scope ID for lifecycle actions")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleAllowedRoles, "customer,administrator,support_agent", "Comma-separated roles allowed to request lifecycle actions")
	rootCmd.PersistentFlags().Int(FlagWaldurOrderRoutingMaxRetries, 5, "Max retries for Waldur order routing")
	rootCmd.PersistentFlags().Int(FlagWaldurOrderRoutingWorkers, 4, "Number of Waldur order routing workers")
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
	rootCmd.PersistentFlags().Bool(FlagWaldurLifecycleQueueEnabled, true, "Enable durable lifecycle command queue")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleQueueBackend, "badger", "Lifecycle queue storage backend (badger)")
	rootCmd.PersistentFlags().String(FlagWaldurLifecycleQueuePath, "data/lifecycle_queue", "Lifecycle queue storage path")
	rootCmd.PersistentFlags().Int(FlagWaldurLifecycleQueueWorkers, 2, "Lifecycle queue worker count")
	rootCmd.PersistentFlags().Int(FlagWaldurLifecycleQueueMaxRetries, 5, "Lifecycle queue max retries")
	rootCmd.PersistentFlags().Duration(FlagWaldurLifecycleQueueRetryBackoff, 10*time.Second, "Lifecycle queue retry backoff")
	rootCmd.PersistentFlags().Duration(FlagWaldurLifecycleQueueMaxBackoff, 5*time.Minute, "Lifecycle queue max backoff")
	rootCmd.PersistentFlags().Duration(FlagWaldurLifecycleQueuePollInterval, 2*time.Second, "Lifecycle queue poll interval")
	rootCmd.PersistentFlags().Duration(FlagWaldurLifecycleQueueReconcileInterval, 5*time.Minute, "Lifecycle queue reconcile interval")
	rootCmd.PersistentFlags().Bool(FlagWaldurLifecycleQueueReconcileOnStart, true, "Run lifecycle reconciliation on startup")
	rootCmd.PersistentFlags().Duration(FlagWaldurLifecycleQueueStaleAfter, 20*time.Minute, "Lifecycle queue stale command threshold")

	// Portal API flags
	rootCmd.PersistentFlags().String(FlagPortalAuthSecret, "", "Shared secret for portal signed requests")
	rootCmd.PersistentFlags().Bool(FlagPortalAllowInsecure, true, "Allow portal requests without signature (dev only)")
	rootCmd.PersistentFlags().Bool(FlagPortalRequireVEID, true, "Require VEID verification for shell access")
	rootCmd.PersistentFlags().Int(FlagPortalMinVEIDScore, 80, "Minimum VEID score required for shell access")
	rootCmd.PersistentFlags().Duration(FlagPortalShellSessionTTL, 10*time.Minute, "Shell session TTL for portal access")
	rootCmd.PersistentFlags().Duration(FlagPortalTokenTTL, 5*time.Minute, "Portal session token TTL")
	rootCmd.PersistentFlags().String(FlagPortalAuditLogFile, "data/portal_audit.log", "Portal audit log file path")

	// Vault flags
	rootCmd.PersistentFlags().Bool(FlagVaultEnabled, true, "Enable data vault APIs")
	rootCmd.PersistentFlags().String(FlagVaultBackend, "memory", "Data vault backend (memory)")
	rootCmd.PersistentFlags().String(FlagVaultAuditOwner, "audit-system", "Vault audit owner account")
	rootCmd.PersistentFlags().Duration(FlagVaultRotateOverlap, 24*time.Hour, "Vault key rotation overlap window")
	rootCmd.PersistentFlags().Duration(FlagVaultAnomalyWindow, 10*time.Minute, "Vault access anomaly detection window")
	rootCmd.PersistentFlags().Int(FlagVaultAnomalyThreshold, 5, "Vault access anomaly threshold")

	// Support service desk flags
	rootCmd.PersistentFlags().Bool(FlagSupportEnabled, false, "Enable support service desk bridge")
	rootCmd.PersistentFlags().String(FlagSupportWaldurBaseURL, "", "Support Waldur API base URL")
	rootCmd.PersistentFlags().String(FlagSupportWaldurToken, "", "Support Waldur API token")
	rootCmd.PersistentFlags().String(FlagSupportWaldurOrgUUID, "", "Support Waldur organization UUID")
	rootCmd.PersistentFlags().String(FlagSupportWaldurProjectUUID, "", "Support Waldur project UUID")
	rootCmd.PersistentFlags().String(FlagSupportWebhookSecret, "", "Support webhook secret")
	rootCmd.PersistentFlags().String(FlagSupportWebhookListen, ":8480", "Support webhook listen address")
	rootCmd.PersistentFlags().Bool(FlagSupportWebhookRequireSig, true, "Require signatures for support webhooks")
	rootCmd.PersistentFlags().String(FlagSupportDecryptionKeyPath, "", "Support payload decryption key path")
	rootCmd.PersistentFlags().String(FlagSupportDecryptionKeyBase64, "", "Support payload decryption key (base64)")
	rootCmd.PersistentFlags().String(FlagSupportEncryptionKeyPath, "", "Support payload encryption key path")
	rootCmd.PersistentFlags().String(FlagSupportEncryptionKeyBase64, "", "Support payload encryption key (base64)")
	rootCmd.PersistentFlags().Bool(FlagSupportSyncInbound, true, "Enable inbound support sync from service desk")
	rootCmd.PersistentFlags().Bool(FlagSupportSyncOutbound, true, "Enable outbound support sync to service desk")
	rootCmd.PersistentFlags().Duration(FlagSupportSyncInterval, 30*time.Second, "Support sync interval")

	// Bind to viper
	_ = viper.BindPFlag(FlagChainID, rootCmd.PersistentFlags().Lookup(FlagChainID))
	_ = viper.BindPFlag(FlagNode, rootCmd.PersistentFlags().Lookup(FlagNode))
	_ = viper.BindPFlag(FlagProviderKey, rootCmd.PersistentFlags().Lookup(FlagProviderKey))
	_ = viper.BindPFlag(FlagProviderKeyDir, rootCmd.PersistentFlags().Lookup(FlagProviderKeyDir))
	_ = viper.BindPFlag(FlagListenAddr, rootCmd.PersistentFlags().Lookup(FlagListenAddr))
	_ = viper.BindPFlag(FlagMetricsAddr, rootCmd.PersistentFlags().Lookup(FlagMetricsAddr))
	_ = viper.BindPFlag(FlagTracingEnabled, rootCmd.PersistentFlags().Lookup(FlagTracingEnabled))
	_ = viper.BindPFlag(FlagTracingEndpoint, rootCmd.PersistentFlags().Lookup(FlagTracingEndpoint))
	_ = viper.BindPFlag(FlagTracingSampleRate, rootCmd.PersistentFlags().Lookup(FlagTracingSampleRate))
	_ = viper.BindPFlag(FlagTracingEnvironment, rootCmd.PersistentFlags().Lookup(FlagTracingEnvironment))
	_ = viper.BindPFlag(FlagWaldurEnabled, rootCmd.PersistentFlags().Lookup(FlagWaldurEnabled))
	_ = viper.BindPFlag(FlagWaldurBaseURL, rootCmd.PersistentFlags().Lookup(FlagWaldurBaseURL))
	_ = viper.BindPFlag(FlagWaldurToken, rootCmd.PersistentFlags().Lookup(FlagWaldurToken))
	_ = viper.BindPFlag(FlagWaldurProjectUUID, rootCmd.PersistentFlags().Lookup(FlagWaldurProjectUUID))
	_ = viper.BindPFlag(FlagWaldurOfferingMap, rootCmd.PersistentFlags().Lookup(FlagWaldurOfferingMap))
	_ = viper.BindPFlag(FlagWaldurCallbackSinkDir, rootCmd.PersistentFlags().Lookup(FlagWaldurCallbackSinkDir))
	_ = viper.BindPFlag(FlagWaldurStateFile, rootCmd.PersistentFlags().Lookup(FlagWaldurStateFile))
	_ = viper.BindPFlag(FlagWaldurCheckpointFile, rootCmd.PersistentFlags().Lookup(FlagWaldurCheckpointFile))
	_ = viper.BindPFlag(FlagWaldurOrderCallbackURL, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderCallbackURL))
	_ = viper.BindPFlag(FlagWaldurOrderRoutingEnabled, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderRoutingEnabled))
	_ = viper.BindPFlag(FlagWaldurOrderStateFile, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderStateFile))
	_ = viper.BindPFlag(FlagWaldurOrderCheckpointFile, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderCheckpointFile))
	_ = viper.BindPFlag(FlagWaldurOrderCallbackListen, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderCallbackListen))
	_ = viper.BindPFlag(FlagWaldurOrderCallbackPath, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderCallbackPath))
	_ = viper.BindPFlag(FlagWaldurLifecycleCallbackURL, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleCallbackURL))
	_ = viper.BindPFlag(FlagWaldurLifecycleCallbackListen, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleCallbackListen))
	_ = viper.BindPFlag(FlagWaldurLifecycleCallbackPath, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleCallbackPath))
	_ = viper.BindPFlag(FlagWaldurLifecycleRequireConsent, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleRequireConsent))
	_ = viper.BindPFlag(FlagWaldurLifecycleConsentScope, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleConsentScope))
	_ = viper.BindPFlag(FlagWaldurLifecycleAllowedRoles, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleAllowedRoles))
	_ = viper.BindPFlag(FlagWaldurOrderRoutingMaxRetries, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderRoutingMaxRetries))
	_ = viper.BindPFlag(FlagWaldurOrderRoutingWorkers, rootCmd.PersistentFlags().Lookup(FlagWaldurOrderRoutingWorkers))
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
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueEnabled, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueEnabled))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueBackend, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueBackend))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueuePath, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueuePath))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueWorkers, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueWorkers))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueMaxRetries, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueMaxRetries))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueRetryBackoff, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueRetryBackoff))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueMaxBackoff, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueMaxBackoff))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueuePollInterval, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueuePollInterval))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueReconcileInterval, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueReconcileInterval))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueReconcileOnStart, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueReconcileOnStart))
	_ = viper.BindPFlag(FlagWaldurLifecycleQueueStaleAfter, rootCmd.PersistentFlags().Lookup(FlagWaldurLifecycleQueueStaleAfter))

	// Portal API flags
	_ = viper.BindPFlag(FlagPortalAuthSecret, rootCmd.PersistentFlags().Lookup(FlagPortalAuthSecret))
	_ = viper.BindPFlag(FlagPortalAllowInsecure, rootCmd.PersistentFlags().Lookup(FlagPortalAllowInsecure))
	_ = viper.BindPFlag(FlagPortalRequireVEID, rootCmd.PersistentFlags().Lookup(FlagPortalRequireVEID))
	_ = viper.BindPFlag(FlagPortalMinVEIDScore, rootCmd.PersistentFlags().Lookup(FlagPortalMinVEIDScore))
	_ = viper.BindPFlag(FlagPortalShellSessionTTL, rootCmd.PersistentFlags().Lookup(FlagPortalShellSessionTTL))
	_ = viper.BindPFlag(FlagPortalTokenTTL, rootCmd.PersistentFlags().Lookup(FlagPortalTokenTTL))
	_ = viper.BindPFlag(FlagPortalAuditLogFile, rootCmd.PersistentFlags().Lookup(FlagPortalAuditLogFile))
	_ = viper.BindPFlag(FlagVaultEnabled, rootCmd.PersistentFlags().Lookup(FlagVaultEnabled))
	_ = viper.BindPFlag(FlagVaultBackend, rootCmd.PersistentFlags().Lookup(FlagVaultBackend))
	_ = viper.BindPFlag(FlagVaultAuditOwner, rootCmd.PersistentFlags().Lookup(FlagVaultAuditOwner))
	_ = viper.BindPFlag(FlagVaultRotateOverlap, rootCmd.PersistentFlags().Lookup(FlagVaultRotateOverlap))
	_ = viper.BindPFlag(FlagVaultAnomalyWindow, rootCmd.PersistentFlags().Lookup(FlagVaultAnomalyWindow))
	_ = viper.BindPFlag(FlagVaultAnomalyThreshold, rootCmd.PersistentFlags().Lookup(FlagVaultAnomalyThreshold))

	// Support service desk flags
	_ = viper.BindPFlag(FlagSupportEnabled, rootCmd.PersistentFlags().Lookup(FlagSupportEnabled))
	_ = viper.BindPFlag(FlagSupportWaldurBaseURL, rootCmd.PersistentFlags().Lookup(FlagSupportWaldurBaseURL))
	_ = viper.BindPFlag(FlagSupportWaldurToken, rootCmd.PersistentFlags().Lookup(FlagSupportWaldurToken))
	_ = viper.BindPFlag(FlagSupportWaldurOrgUUID, rootCmd.PersistentFlags().Lookup(FlagSupportWaldurOrgUUID))
	_ = viper.BindPFlag(FlagSupportWaldurProjectUUID, rootCmd.PersistentFlags().Lookup(FlagSupportWaldurProjectUUID))
	_ = viper.BindPFlag(FlagSupportWebhookSecret, rootCmd.PersistentFlags().Lookup(FlagSupportWebhookSecret))
	_ = viper.BindPFlag(FlagSupportWebhookListen, rootCmd.PersistentFlags().Lookup(FlagSupportWebhookListen))
	_ = viper.BindPFlag(FlagSupportWebhookRequireSig, rootCmd.PersistentFlags().Lookup(FlagSupportWebhookRequireSig))
	_ = viper.BindPFlag(FlagSupportDecryptionKeyPath, rootCmd.PersistentFlags().Lookup(FlagSupportDecryptionKeyPath))
	_ = viper.BindPFlag(FlagSupportDecryptionKeyBase64, rootCmd.PersistentFlags().Lookup(FlagSupportDecryptionKeyBase64))
	_ = viper.BindPFlag(FlagSupportEncryptionKeyPath, rootCmd.PersistentFlags().Lookup(FlagSupportEncryptionKeyPath))
	_ = viper.BindPFlag(FlagSupportEncryptionKeyBase64, rootCmd.PersistentFlags().Lookup(FlagSupportEncryptionKeyBase64))
	_ = viper.BindPFlag(FlagSupportSyncInbound, rootCmd.PersistentFlags().Lookup(FlagSupportSyncInbound))
	_ = viper.BindPFlag(FlagSupportSyncOutbound, rootCmd.PersistentFlags().Lookup(FlagSupportSyncOutbound))
	_ = viper.BindPFlag(FlagSupportSyncInterval, rootCmd.PersistentFlags().Lookup(FlagSupportSyncInterval))

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
	fmt.Printf("  Tracing Enabled: %t\n", viper.GetBool(FlagTracingEnabled))

	obsCfg := observability.DefaultConfig()
	obsCfg.ServiceName = "virtengine-provider-daemon"
	obsCfg.Environment = viper.GetString(FlagTracingEnvironment)
	obsCfg.TracingEnabled = viper.GetBool(FlagTracingEnabled)
	obsCfg.TracingEndpoint = viper.GetString(FlagTracingEndpoint)
	obsCfg.TracingSampleRate = viper.GetFloat64(FlagTracingSampleRate)
	observer, err := observability.New(obsCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize tracing: %w", err)
	}
	defer func() {
		_ = observer.Shutdown(context.Background())
	}()

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
	var supportService *provider_daemon.SupportService

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
	chainClient, err := provider_daemon.NewRPCChainClient(provider_daemon.RPCChainClientConfig{
		NodeURI:        viper.GetString(FlagNode),
		GRPCEndpoint:   viper.GetString(FlagWaldurChainGRPC),
		ChainID:        viper.GetString(FlagChainID),
		RequestTimeout: time.Second * 30,
	})
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}

	// Initialize HPC provider (VE-21C/VE-14B)
	var hpcProvider *provider_daemon.HPCProvider
	hpcProviderConfig := provider_daemon.DefaultHPCProviderConfig()
	if viper.IsSet("hpc_provider") {
		if err := viper.UnmarshalKey("hpc_provider", &hpcProviderConfig); err != nil {
			return fmt.Errorf("failed to load hpc_provider config: %w", err)
		}
	}
	if viper.IsSet("hpc") {
		if err := viper.UnmarshalKey("hpc", &hpcProviderConfig.HPC); err != nil {
			return fmt.Errorf("failed to load hpc config: %w", err)
		}
	}

	if hpcProviderConfig.HPC.ProviderAddress == "" {
		hpcProviderConfig.HPC.ProviderAddress = providerAddress
	}
	if hpcProviderConfig.HPC.NodeAggregator.ProviderAddress == "" {
		hpcProviderConfig.HPC.NodeAggregator.ProviderAddress = hpcProviderConfig.HPC.ProviderAddress
	}
	if hpcProviderConfig.HPC.NodeAggregator.ClusterID == "" {
		hpcProviderConfig.HPC.NodeAggregator.ClusterID = hpcProviderConfig.HPC.ClusterID
	}
	if hpcProviderConfig.HPC.SlurmK8s.ClusterName == "" {
		hpcProviderConfig.HPC.SlurmK8s.ClusterName = hpcProviderConfig.HPC.ClusterID
	}

	if hpcProviderConfig.HPC.Enabled {
		hpcChainClient, err := provider_daemon.NewHPCChainClient(provider_daemon.RPCChainClientConfig{
			NodeURI:        viper.GetString(FlagNode),
			GRPCEndpoint:   viper.GetString(FlagWaldurChainGRPC),
			ChainID:        viper.GetString(FlagChainID),
			RequestTimeout: time.Second * 30,
		})
		if err != nil {
			return fmt.Errorf("failed to create hpc chain client: %w", err)
		}

		hpcProvider, err = provider_daemon.NewHPCProvider(hpcProviderConfig, hpcChainClient, nil)
		if err != nil {
			return fmt.Errorf("failed to create hpc provider: %w", err)
		}

		if err := hpcProvider.Start(ctx); err != nil {
			return fmt.Errorf("failed to start hpc provider: %w", err)
		}
		fmt.Println("  HPC Provider: started")
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

	var lifecycleManager *provider_daemon.ResourceLifecycleManager
	var lifecycleController *provider_daemon.LifecycleController
	var lifecycleReconciler *provider_daemon.LifecycleDriftReconciler

	if viper.GetBool(FlagWaldurEnabled) {
		waldurCfg := waldur.DefaultConfig()
		waldurCfg.BaseURL = viper.GetString(FlagWaldurBaseURL)
		waldurCfg.Token = viper.GetString(FlagWaldurToken)
		waldurClient, err := waldur.NewClient(waldurCfg)
		if err != nil {
			return fmt.Errorf("failed to create Waldur client: %w", err)
		}
		marketplaceClient := waldur.NewMarketplaceClient(waldurClient)

		lifecycleControllerCfg := provider_daemon.DefaultLifecycleControllerConfig()
		lifecycleControllerCfg.ProviderAddress = providerAddress
		lifecycleControllerCfg.CallbackURL = viper.GetString(FlagWaldurLifecycleCallbackURL)
		lifecycleControllerCfg.StateFilePath = "data/lifecycle_state.json"

		controller, err := provider_daemon.NewLifecycleController(
			lifecycleControllerCfg,
			keyManager,
			callbackSink,
			marketplaceClient,
			portalAuditLogger,
		)
		if err != nil {
			return fmt.Errorf("failed to initialize lifecycle controller: %w", err)
		}
		lifecycleController = controller

		lifecycleClient := waldur.NewLifecycleClient(marketplaceClient)
		lifecycleManager = provider_daemon.NewResourceLifecycleManager(
			provider_daemon.DefaultResourceLifecycleConfig(),
			lifecycleController,
			lifecycleClient,
			portalAuditLogger,
		)

		if err := lifecycleController.Start(ctx); err != nil {
			return fmt.Errorf("failed to start lifecycle controller: %w", err)
		}

		reconciler := provider_daemon.NewLifecycleDriftReconciler(
			provider_daemon.DefaultLifecycleDriftReconcilerConfig(),
			lifecycleController,
			lifecycleManager,
			lifecycleClient,
			portalAuditLogger,
		)
		lifecycleReconciler = reconciler
		if err := lifecycleReconciler.Start(ctx); err != nil {
			return fmt.Errorf("failed to start lifecycle reconciler: %w", err)
		}

		callbackCfg := provider_daemon.DefaultWaldurCallbackConfig()
		callbackCfg.ListenAddr = viper.GetString(FlagWaldurLifecycleCallbackListen)
		callbackCfg.CallbackPath = viper.GetString(FlagWaldurLifecycleCallbackPath)
		callbackCfg.EnableAuditLogging = true
		callbackHandler := provider_daemon.NewWaldurCallbackHandler(
			callbackCfg,
			lifecycleController,
			callbackSink,
			portalAuditLogger,
			keyManager,
		)
		callbackHandler.SetLifecycleManager(lifecycleManager)
		go func() {
			if err := callbackHandler.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Printf("[WALDUR] lifecycle callback handler stopped: %v\\n", err)
			}
		}()
		fmt.Println("  Waldur Lifecycle Callback Handler: started")
	}

	var chainQuery provider_daemon.ChainQuery = provider_daemon.NoopChainQuery{}
	var roleConn *grpc.ClientConn

	vaultCfg := provider_daemon.DefaultVaultServiceConfig()
	vaultCfg.Enabled = viper.GetBool(FlagVaultEnabled)
	vaultCfg.Backend = viper.GetString(FlagVaultBackend)
	vaultCfg.AuditOwner = viper.GetString(FlagVaultAuditOwner)
	vaultCfg.RotateOverlap = viper.GetDuration(FlagVaultRotateOverlap)
	vaultCfg.AnomalyWindow = viper.GetDuration(FlagVaultAnomalyWindow)
	vaultCfg.AnomalyThreshold = viper.GetInt(FlagVaultAnomalyThreshold)
	vaultCfg.OrgResolver = provider_daemon.ChainOrgResolver{ChainQuery: chainQuery}
	vaultCfg.RoleResolver = provider_daemon.ChainRoleResolver{ChainQuery: chainQuery}
	if grpcEndpoint := viper.GetString(FlagWaldurChainGRPC); grpcEndpoint != "" {
		conn, err := grpc.NewClient(
			grpcEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithStatsHandler(observability.GRPCClientStatsHandler()),
		)
		if err != nil {
			return fmt.Errorf("failed to connect role query grpc: %w", err)
		}
		roleConn = conn
		defer func() {
			_ = roleConn.Close()
		}()

		rolesClient := rolesv1.NewQueryClient(roleConn)
		veidClient := veidv1.NewQueryClient(roleConn)
		if portalQuery := provider_daemon.NewGRPCPortalChainQuery(rolesClient, veidClient); portalQuery != nil {
			chainQuery = portalQuery
		}

		vaultCfg.RoleResolver = provider_daemon.NewGRPCRoleResolver(rolesClient)
	}

	vaultService, err := provider_daemon.NewVaultService(vaultCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize vault service: %w", err)
	}

	portalCfg := provider_daemon.DefaultPortalAPIServerConfig()
	portalCfg.ListenAddr = viper.GetString(FlagListenAddr)
	portalCfg.AuthSecret = viper.GetString(FlagPortalAuthSecret)
	portalCfg.AllowInsecure = viper.GetBool(FlagPortalAllowInsecure)
	portalCfg.RequireVEID = viper.GetBool(FlagPortalRequireVEID)
	portalCfg.MinVEIDScore = viper.GetInt(FlagPortalMinVEIDScore)
	portalCfg.ShellSessionTTL = viper.GetDuration(FlagPortalShellSessionTTL)
	portalCfg.TokenTTL = viper.GetDuration(FlagPortalTokenTTL)
	portalCfg.AuditLogger = portalAuditLogger
	portalCfg.WalletAuthChainID = viper.GetString(FlagChainID)
	portalCfg.ChainQuery = chainQuery
	portalCfg.VaultService = vaultService
	portalCfg.LifecycleExecutor = lifecycleManager
	portalCfg.LifecycleRequireConsent = viper.GetBool(FlagWaldurLifecycleRequireConsent)
	portalCfg.LifecycleConsentScope = viper.GetString(FlagWaldurLifecycleConsentScope)
	portalCfg.LifecycleAllowedRoles = parseCSVList(viper.GetString(FlagWaldurLifecycleAllowedRoles))

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

		orderRoutingCfg := provider_daemon.DefaultOrderRoutingConfig()
		orderRoutingCfg.Enabled = viper.GetBool(FlagWaldurOrderRoutingEnabled)
		orderRoutingCfg.ProviderAddress = providerAddress
		orderRoutingCfg.WaldurBaseURL = viper.GetString(FlagWaldurBaseURL)
		orderRoutingCfg.WaldurToken = viper.GetString(FlagWaldurToken)
		orderRoutingCfg.WaldurProjectID = viper.GetString(FlagWaldurProjectUUID)
		orderRoutingCfg.OrderCallbackURL = viper.GetString(FlagWaldurOrderCallbackURL)
		orderRoutingCfg.OfferingMap = offeringMap
		orderRoutingCfg.StateFile = viper.GetString(FlagWaldurOrderStateFile)
		orderRoutingCfg.MaxRetries = viper.GetInt(FlagWaldurOrderRoutingMaxRetries)
		orderRoutingCfg.WorkerCount = viper.GetInt(FlagWaldurOrderRoutingWorkers)

		var orderRouter *provider_daemon.OrderRouter
		if orderRoutingCfg.Enabled {
			router, err := provider_daemon.NewOrderRouter(orderRoutingCfg, nil)
			if err != nil {
				return fmt.Errorf("failed to create order router: %w", err)
			}
			orderRouter = router
			orderRouter.Start(ctx)
			fmt.Println("  Waldur Order Router: started")

			listenerCfg := provider_daemon.DefaultOrderListenerConfig()
			listenerCfg.Enabled = true
			listenerCfg.ProviderAddress = providerAddress
			listenerCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
			listenerCfg.CometWS = viper.GetString(FlagCometWS)
			listenerCfg.EventQuery = ""
			listenerCfg.CheckpointFile = viper.GetString(FlagWaldurOrderCheckpointFile)
			listenerCfg.SubscriberID = fmt.Sprintf("order-router-%s", providerID[:8])

			orderListener, err := provider_daemon.NewOrderListener(listenerCfg, orderRouter)
			if err != nil {
				return fmt.Errorf("failed to create order listener: %w", err)
			}
			go func() {
				if err := orderListener.Start(ctx); err != nil {
					fmt.Printf("[WALDUR] order listener stopped: %v\n", err)
				}
			}()
			fmt.Println("  Waldur Order Listener: started")

			statusHandler, err := provider_daemon.NewOrderStatusCallbackHandler(
				keyManager,
				callbackSink,
				orderRouter.Store(),
			)
			if err != nil {
				return fmt.Errorf("failed to create order status handler: %w", err)
			}
			webhookCfg := provider_daemon.DefaultOrderStatusWebhookConfig()
			webhookCfg.ListenAddr = viper.GetString(FlagWaldurOrderCallbackListen)
			webhookCfg.CallbackPath = viper.GetString(FlagWaldurOrderCallbackPath)
			webhookServer, err := provider_daemon.NewOrderStatusWebhookServer(webhookCfg, statusHandler)
			if err != nil {
				return fmt.Errorf("failed to create order status webhook: %w", err)
			}
			go func() {
				if err := webhookServer.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
					fmt.Printf("[WALDUR] order status webhook stopped: %v\n", err)
				}
			}()
			fmt.Println("  Waldur Order Status Webhook: started")
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
		bridgeCfg.LifecycleQueueEnabled = viper.GetBool(FlagWaldurLifecycleQueueEnabled)
		bridgeCfg.LifecycleQueueBackend = viper.GetString(FlagWaldurLifecycleQueueBackend)
		bridgeCfg.LifecycleQueuePath = viper.GetString(FlagWaldurLifecycleQueuePath)
		bridgeCfg.LifecycleQueueWorkerCount = viper.GetInt(FlagWaldurLifecycleQueueWorkers)
		bridgeCfg.LifecycleQueueMaxRetries = viper.GetInt(FlagWaldurLifecycleQueueMaxRetries)
		bridgeCfg.LifecycleQueueRetryBackoff = viper.GetDuration(FlagWaldurLifecycleQueueRetryBackoff)
		bridgeCfg.LifecycleQueueMaxBackoff = viper.GetDuration(FlagWaldurLifecycleQueueMaxBackoff)
		bridgeCfg.LifecycleQueuePollInterval = viper.GetDuration(FlagWaldurLifecycleQueuePollInterval)
		bridgeCfg.LifecycleQueueReconcileInterval = viper.GetDuration(FlagWaldurLifecycleQueueReconcileInterval)
		bridgeCfg.LifecycleQueueReconcileOnStart = viper.GetBool(FlagWaldurLifecycleQueueReconcileOnStart)
		bridgeCfg.LifecycleQueueStaleAfter = viper.GetDuration(FlagWaldurLifecycleQueueStaleAfter)

		waldurBridge, err := provider_daemon.NewWaldurBridge(bridgeCfg, keyManager, callbackSink, usageReporter)
		if err != nil {
			return fmt.Errorf("failed to create waldur bridge: %w", err)
		}
		if lifecycleManager != nil {
			waldurBridge.SetLifecycleManager(lifecycleManager)
		}

		go func() {
			if err := waldurBridge.Start(ctx); err != nil {
				fmt.Printf("[WALDUR] bridge stopped: %v\n", err)
			}
		}()
		fmt.Println("  Waldur Bridge: started")
	}

	// Initialize support service desk bridge (VE-25C)
	if viper.GetBool(FlagSupportEnabled) {
		supportCfg := provider_daemon.DefaultSupportServiceConfig()
		supportCfg.Enabled = true
		supportCfg.ProviderAddress = providerAddress
		supportCfg.ChainID = viper.GetString(FlagChainID)
		supportCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
		supportCfg.CometWS = viper.GetString(FlagCometWS)
		supportCfg.GRPCEndpoint = viper.GetString(FlagWaldurChainGRPC)
		supportCfg.Encryption.SenderPrivateKeyPath = viper.GetString(FlagSupportEncryptionKeyPath)
		supportCfg.Encryption.SenderPrivateKeyBase64 = viper.GetString(FlagSupportEncryptionKeyBase64)

		if supportCfg.ServiceDeskConfig != nil {
			supportCfg.ServiceDeskConfig.Enabled = true
			supportCfg.ServiceDeskConfig.SyncConfig.EnableInbound = viper.GetBool(FlagSupportSyncInbound)
			supportCfg.ServiceDeskConfig.SyncConfig.EnableOutbound = viper.GetBool(FlagSupportSyncOutbound)
			supportCfg.ServiceDeskConfig.SyncConfig.SyncInterval = viper.GetDuration(FlagSupportSyncInterval)
			supportCfg.ServiceDeskConfig.WebhookConfig.ListenAddr = viper.GetString(FlagSupportWebhookListen)
			supportCfg.ServiceDeskConfig.WebhookConfig.RequireSignature = viper.GetBool(FlagSupportWebhookRequireSig)
			if supportCfg.ServiceDeskConfig.Decryption == nil {
				supportCfg.ServiceDeskConfig.Decryption = &servicedesk.DecryptionConfig{}
			}
			supportCfg.ServiceDeskConfig.Decryption.PrivateKeyPath = viper.GetString(FlagSupportDecryptionKeyPath)
			supportCfg.ServiceDeskConfig.Decryption.PrivateKeyBase64 = viper.GetString(FlagSupportDecryptionKeyBase64)
			supportCfg.ServiceDeskConfig.WaldurConfig = &servicedesk.WaldurConfig{
				BaseURL:          viper.GetString(FlagSupportWaldurBaseURL),
				Token:            viper.GetString(FlagSupportWaldurToken),
				OrganizationUUID: viper.GetString(FlagSupportWaldurOrgUUID),
				ProjectUUID:      viper.GetString(FlagSupportWaldurProjectUUID),
				WebhookSecret:    viper.GetString(FlagSupportWebhookSecret),
				Timeout:          30 * time.Second,
			}
		}

		svc, err := provider_daemon.NewSupportService(supportCfg, keyManager, provider_daemon.NewSupportLogger())
		if err != nil {
			return fmt.Errorf("failed to create support service: %w", err)
		}
		supportService = svc
		if supportService != nil {
			if err := supportService.Start(ctx); err != nil {
				return fmt.Errorf("failed to start support service: %w", err)
			}
			fmt.Println("  Support Service: started")
		}
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

	if hpcProvider != nil {
		if err := hpcProvider.Stop(); err != nil {
			fmt.Printf("  HPC Provider: failed to stop cleanly: %v\n", err)
		} else {
			fmt.Println("  HPC Provider: stopped")
		}
	}

	usageMeter.Stop()
	fmt.Println("  Usage Meter: stopped")

	if supportService != nil {
		_ = supportService.Stop(ctx)
		fmt.Println("  Support Service: stopped")
	}

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

func parseCSVList(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			continue
		}
		result = append(result, value)
	}
	return result
}
