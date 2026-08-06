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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	resourcesv1 "github.com/virtengine/virtengine/sdk/go/node/resources/v1"
	rolesv1 "github.com/virtengine/virtengine/sdk/go/node/roles/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/virtengine/virtengine/pkg/dex"
	"github.com/virtengine/virtengine/pkg/observability"
	"github.com/virtengine/virtengine/pkg/payments/offramp"
	provider_daemon "github.com/virtengine/virtengine/pkg/provider_daemon"
	"github.com/virtengine/virtengine/pkg/security"
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

	// FlagProviderKeyPassphraseFile is a secret-mounted file containing the
	// encrypted file-keystore passphrase.
	FlagProviderKeyPassphraseFile = "key-passphrase-file" //nolint:gosec // flag name, not a credential

	// FlagProviderKeyFingerprint is the expected SHA-256 public-key fingerprint.
	FlagProviderKeyFingerprint = "key-fingerprint"

	// FlagSubmitterLeaseFile stores cross-process ownership and fencing tokens.
	FlagSubmitterLeaseFile = "submitter-lease-file"

	// FlagSubmitterLeaseOwner is a unique replica identity, normally the pod UID.
	FlagSubmitterLeaseOwner = "submitter-lease-owner"

	// FlagSubmitterLeaseBackend selects kubernetes or shared_file fencing.
	FlagSubmitterLeaseBackend = "submitter-lease-backend"

	// FlagSubmitterLeaseNamespace is the Kubernetes Lease namespace.
	FlagSubmitterLeaseNamespace = "submitter-lease-namespace"

	// FlagProviderProduction enforces persistent key and distributed lease profiles.
	FlagProviderProduction = "production"

	// FlagKeyBackupFile is the path to the encrypted provider key backup file
	FlagKeyBackupFile = "key-backup-file"

	// FlagKeyBackupRestore restores provider keys from the backup file before startup
	FlagKeyBackupRestore = "key-backup-restore"

	// FlagKeyBackupExport writes an encrypted provider key backup after key initialization
	FlagKeyBackupExport = "key-backup-export"

	// FlagKeyBackupPassphrase is the passphrase used for provider key backups
	FlagKeyBackupPassphrase = "key-backup-passphrase" //nolint:gosec // #nosec G101: CLI flag name, not a credential

	// FlagKubeconfig is the path to kubeconfig
	FlagKubeconfig = "kubeconfig"

	// FlagMeteringInterval is the metering interval
	FlagMeteringInterval = "metering-interval"

	// FlagChainUsageSubmit enables authenticated durable usage submission.
	FlagChainUsageSubmit = "chain-usage-submit"

	// FlagChainUsageQueueFile stores proof allocation and transaction state.
	FlagChainUsageQueueFile = "chain-usage-queue-file"

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
	FlagWaldurToken = "waldur-token" //nolint:gosec // #nosec G101: CLI flag name, not a credential

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

	// FlagProvisioningEnabled toggles provisioning worker
	FlagProvisioningEnabled = "provisioning-enabled"

	// FlagProvisioningStateFile is path to provisioning state file
	FlagProvisioningStateFile = "provisioning-state-file"

	// FlagProvisioningCheckpointFile is path to provisioning checkpoint file
	FlagProvisioningCheckpointFile = "provisioning-checkpoint-file"

	// FlagProvisioningMaxRetries is max retries for provisioning
	FlagProvisioningMaxRetries = "provisioning-max-retries"

	// FlagProvisioningRetryBackoff is base backoff duration for provisioning retries
	FlagProvisioningRetryBackoff = "provisioning-retry-backoff"

	// FlagProvisioningMaxBackoff is max backoff duration for provisioning retries
	FlagProvisioningMaxBackoff = "provisioning-max-backoff"

	// FlagProvisioningPollInterval is provisioning status poll interval
	FlagProvisioningPollInterval = "provisioning-poll-interval"

	// FlagProvisioningDryRun enables dry-run provisioning for container runtime
	FlagProvisioningDryRun = "provisioning-dry-run"

	// FlagWaldurChainSubmit enables on-chain Waldur callback submission
	FlagWaldurChainSubmit = "waldur-chain-submit"

	// FlagWaldurChainKey is the key name for on-chain Waldur callbacks
	FlagWaldurChainKey = "waldur-chain-key"

	// FlagWaldurChainKeyringBackend is the keyring backend for on-chain callbacks
	FlagWaldurChainKeyringBackend = "waldur-chain-keyring-backend"

	// FlagWaldurChainKeyringDir is the keyring dir for on-chain callbacks
	FlagWaldurChainKeyringDir = "waldur-chain-keyring-dir"

	// FlagWaldurChainKeyringPassphrase is the keyring passphrase for on-chain callbacks
	FlagWaldurChainKeyringPassphrase = "waldur-chain-keyring-passphrase" //nolint:gosec // #nosec G101: CLI flag name, not a credential

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
	FlagSupportWaldurToken         = "support-waldur-token" //nolint:gosec // #nosec G101: CLI flag name, not a credential
	FlagSupportWaldurOrgUUID       = "support-waldur-org-uuid"
	FlagSupportWaldurProjectUUID   = "support-waldur-project-uuid"
	FlagSupportWebhookSecret       = "support-webhook-secret" //nolint:gosec // #nosec G101: webhook secret flag name, not a credential
	FlagSupportWebhookListen       = "support-webhook-listen"
	FlagSupportWebhookRequireSig   = "support-webhook-require-signature"
	FlagSupportDecryptionKeyPath   = "support-decryption-key-path"
	FlagSupportDecryptionKeyBase64 = "support-decryption-key-base64"
	FlagSupportEncryptionKeyPath   = "support-encryption-key-path"
	FlagSupportEncryptionKeyBase64 = "support-encryption-key-base64"
	FlagSupportSyncInbound         = "support-sync-inbound"
	FlagSupportSyncOutbound        = "support-sync-outbound"
	FlagSupportSyncInterval        = "support-sync-interval"

	// FlagDomainVerificationEnabled enables automated provider domain verification checks.
	FlagDomainVerificationEnabled = "domain-verification-enabled"

	// Task 85B conversion flags. Runtime defaults disabled and composition never
	// fabricates external custody or payout-provider dependencies.
	FlagFiatConversionEnabled             = "fiat-conversion-enabled"
	FlagFiatConversionMode                = "fiat-conversion-mode"
	FlagFiatConversionProfileAuthorityKey = "fiat-conversion-profile-authority-public-key"
	FlagFiatConversionProfileAuthorityID  = "fiat-conversion-profile-authority-id"
	FlagFiatConversionDEXProfile          = "fiat-conversion-dex-profile"
	FlagFiatConversionPayoutProfile       = "fiat-conversion-payout-profile"
	FlagFiatConversionStateFile           = "fiat-conversion-state-file"
	FlagFiatConversionRepositoryFile      = "fiat-conversion-repository-file"
	FlagFiatConversionPollInterval        = "fiat-conversion-poll-interval"
	FlagFiatConversionRetryBackoff        = "fiat-conversion-retry-backoff"
	FlagFiatConversionMaxRetryBackoff     = "fiat-conversion-max-retry-backoff"
	FlagFiatConversionMaxAttempts         = "fiat-conversion-max-attempts"
	FlagFiatConversionWebhookListen       = "fiat-conversion-webhook-listen"
	FlagFiatConversionWebhookPath         = "fiat-conversion-webhook-path"
	FlagFiatConversionWebhookBodyLimit    = "fiat-conversion-webhook-body-limit"
	FlagFiatConversionWebhookTimeout      = "fiat-conversion-webhook-timeout"
	FlagFiatConversionCustodyBackend      = "fiat-conversion-custody-backend"
	FlagFiatConversionSecretResolver      = "fiat-conversion-secret-resolver"
	FlagFiatConversionDestinationResolver = "fiat-conversion-destination-resolver"
	FlagFiatConversionComplianceResolver  = "fiat-conversion-compliance-resolver"
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
	rootCmd.PersistentFlags().String(FlagProviderKeyPassphraseFile, "", "Secret file containing the provider file-keystore passphrase")
	rootCmd.PersistentFlags().String(FlagProviderKeyFingerprint, "", "Expected SHA-256 provider public-key fingerprint")
	rootCmd.PersistentFlags().String(FlagSubmitterLeaseFile, "", "Durable cross-process submitter lease state file")
	rootCmd.PersistentFlags().String(FlagSubmitterLeaseOwner, "", "Unique submitter lease owner identity")
	rootCmd.PersistentFlags().String(FlagSubmitterLeaseBackend, "kubernetes", "Submitter fencing backend (kubernetes or shared_file)")
	rootCmd.PersistentFlags().String(FlagSubmitterLeaseNamespace, "virtengine", "Kubernetes namespace for submitter Lease objects")
	rootCmd.PersistentFlags().Bool(FlagProviderProduction, false, "Enforce production key, store, and fencing requirements")
	rootCmd.PersistentFlags().Bool(FlagChainUsageSubmit, false, "Enable authenticated durable on-chain usage submission")
	rootCmd.PersistentFlags().String(FlagChainUsageQueueFile, "data/chain_usage_queue.json", "Authenticated usage durable queue state file")
	rootCmd.PersistentFlags().String(FlagKeyBackupFile, "", "Path to encrypted provider key backup")
	rootCmd.PersistentFlags().Bool(FlagKeyBackupRestore, false, "Restore provider keys from --key-backup-file before startup")
	rootCmd.PersistentFlags().Bool(FlagKeyBackupExport, false, "Write an encrypted provider key backup to --key-backup-file after key initialization")
	rootCmd.PersistentFlags().String(FlagKeyBackupPassphrase, "", "Passphrase for provider key backup encryption")
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
	rootCmd.PersistentFlags().Bool(FlagProvisioningEnabled, false, "Enable VM/container provisioning worker")
	rootCmd.PersistentFlags().String(FlagProvisioningStateFile, "data/provisioning_state.json", "Provisioning state file path")
	rootCmd.PersistentFlags().String(FlagProvisioningCheckpointFile, "data/provisioning_checkpoint.json", "Provisioning checkpoint file path")
	rootCmd.PersistentFlags().Int(FlagProvisioningMaxRetries, 5, "Max retries for provisioning")
	rootCmd.PersistentFlags().Duration(FlagProvisioningRetryBackoff, 10*time.Second, "Provisioning retry backoff")
	rootCmd.PersistentFlags().Duration(FlagProvisioningMaxBackoff, 5*time.Minute, "Provisioning max backoff")
	rootCmd.PersistentFlags().Duration(FlagProvisioningPollInterval, 15*time.Second, "Provisioning poll interval")
	rootCmd.PersistentFlags().Bool(FlagProvisioningDryRun, false, "Enable dry-run container provisioning (no Kubernetes API calls)")
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
	rootCmd.PersistentFlags().Bool(FlagDomainVerificationEnabled, false, "Enable automated provider domain verification checks")
	rootCmd.PersistentFlags().Bool(FlagFiatConversionEnabled, false, "Enable the Task 85B off-chain fiat conversion orchestrator (default disabled)")
	rootCmd.PersistentFlags().String(FlagFiatConversionMode, "production", "Fiat conversion mode (production or engineering_external_blocked)")
	rootCmd.PersistentFlags().String(FlagFiatConversionProfileAuthorityKey, "", "Base64 Ed25519 public key for independently signed certified profile files")
	rootCmd.PersistentFlags().String(FlagFiatConversionProfileAuthorityID, "", "Trusted fiat profile authority identifier")
	rootCmd.PersistentFlags().String(FlagFiatConversionDEXProfile, "", "Versioned DEX route profile JSON path")
	rootCmd.PersistentFlags().String(FlagFiatConversionPayoutProfile, "", "Versioned payout profile JSON path")
	rootCmd.PersistentFlags().String(FlagFiatConversionStateFile, "data/fiat_conversion_state.json", "Durable conversion orchestrator state file")
	rootCmd.PersistentFlags().String(FlagFiatConversionRepositoryFile, "data/fiat_conversion_repository.json", "Durable payout/limit/webhook repository file")
	rootCmd.PersistentFlags().Duration(FlagFiatConversionPollInterval, 10*time.Second, "Fiat conversion intent polling interval")
	rootCmd.PersistentFlags().Duration(FlagFiatConversionRetryBackoff, 2*time.Second, "Fiat conversion base retry backoff")
	rootCmd.PersistentFlags().Duration(FlagFiatConversionMaxRetryBackoff, 5*time.Minute, "Fiat conversion maximum retry backoff")
	rootCmd.PersistentFlags().Uint32(FlagFiatConversionMaxAttempts, 12, "Maximum external conversion attempts before safe failure/manual review")
	rootCmd.PersistentFlags().String(FlagFiatConversionWebhookListen, "127.0.0.1:8485", "Private payout webhook listen address")
	rootCmd.PersistentFlags().String(FlagFiatConversionWebhookPath, "/internal/v1/offramp/webhook", "Private authenticated payout webhook path")
	rootCmd.PersistentFlags().Int64(FlagFiatConversionWebhookBodyLimit, 1<<20, "Maximum payout webhook body bytes")
	rootCmd.PersistentFlags().Duration(FlagFiatConversionWebhookTimeout, 5*time.Second, "Payout webhook request timeout")
	rootCmd.PersistentFlags().String(FlagFiatConversionCustodyBackend, "", "External target-chain custody signer backend identifier")
	rootCmd.PersistentFlags().String(FlagFiatConversionSecretResolver, "", "Secret resolver backend identifier")
	rootCmd.PersistentFlags().String(FlagFiatConversionDestinationResolver, "", "Encrypted destination resolver backend identifier")
	rootCmd.PersistentFlags().String(FlagFiatConversionComplianceResolver, "", "Compliance decision resolver backend identifier")

	// Bind to viper
	_ = viper.BindPFlag(FlagChainID, rootCmd.PersistentFlags().Lookup(FlagChainID))
	_ = viper.BindPFlag(FlagNode, rootCmd.PersistentFlags().Lookup(FlagNode))
	_ = viper.BindPFlag(FlagProviderKey, rootCmd.PersistentFlags().Lookup(FlagProviderKey))
	_ = viper.BindPFlag(FlagProviderKeyDir, rootCmd.PersistentFlags().Lookup(FlagProviderKeyDir))
	_ = viper.BindPFlag(FlagProviderKeyPassphraseFile, rootCmd.PersistentFlags().Lookup(FlagProviderKeyPassphraseFile))
	_ = viper.BindPFlag(FlagProviderKeyFingerprint, rootCmd.PersistentFlags().Lookup(FlagProviderKeyFingerprint))
	_ = viper.BindPFlag(FlagSubmitterLeaseFile, rootCmd.PersistentFlags().Lookup(FlagSubmitterLeaseFile))
	_ = viper.BindPFlag(FlagSubmitterLeaseOwner, rootCmd.PersistentFlags().Lookup(FlagSubmitterLeaseOwner))
	_ = viper.BindPFlag(FlagSubmitterLeaseBackend, rootCmd.PersistentFlags().Lookup(FlagSubmitterLeaseBackend))
	_ = viper.BindPFlag(FlagSubmitterLeaseNamespace, rootCmd.PersistentFlags().Lookup(FlagSubmitterLeaseNamespace))
	_ = viper.BindPFlag(FlagProviderProduction, rootCmd.PersistentFlags().Lookup(FlagProviderProduction))
	_ = viper.BindPFlag(FlagChainUsageSubmit, rootCmd.PersistentFlags().Lookup(FlagChainUsageSubmit))
	_ = viper.BindPFlag(FlagChainUsageQueueFile, rootCmd.PersistentFlags().Lookup(FlagChainUsageQueueFile))
	_ = viper.BindPFlag(FlagKeyBackupFile, rootCmd.PersistentFlags().Lookup(FlagKeyBackupFile))
	_ = viper.BindPFlag(FlagKeyBackupRestore, rootCmd.PersistentFlags().Lookup(FlagKeyBackupRestore))
	_ = viper.BindPFlag(FlagKeyBackupExport, rootCmd.PersistentFlags().Lookup(FlagKeyBackupExport))
	_ = viper.BindPFlag(FlagKeyBackupPassphrase, rootCmd.PersistentFlags().Lookup(FlagKeyBackupPassphrase))
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
	_ = viper.BindPFlag(FlagProvisioningEnabled, rootCmd.PersistentFlags().Lookup(FlagProvisioningEnabled))
	_ = viper.BindPFlag(FlagProvisioningStateFile, rootCmd.PersistentFlags().Lookup(FlagProvisioningStateFile))
	_ = viper.BindPFlag(FlagProvisioningCheckpointFile, rootCmd.PersistentFlags().Lookup(FlagProvisioningCheckpointFile))
	_ = viper.BindPFlag(FlagProvisioningMaxRetries, rootCmd.PersistentFlags().Lookup(FlagProvisioningMaxRetries))
	_ = viper.BindPFlag(FlagProvisioningRetryBackoff, rootCmd.PersistentFlags().Lookup(FlagProvisioningRetryBackoff))
	_ = viper.BindPFlag(FlagProvisioningMaxBackoff, rootCmd.PersistentFlags().Lookup(FlagProvisioningMaxBackoff))
	_ = viper.BindPFlag(FlagProvisioningPollInterval, rootCmd.PersistentFlags().Lookup(FlagProvisioningPollInterval))
	_ = viper.BindPFlag(FlagProvisioningDryRun, rootCmd.PersistentFlags().Lookup(FlagProvisioningDryRun))
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
	_ = viper.BindPFlag(FlagDomainVerificationEnabled, rootCmd.PersistentFlags().Lookup(FlagDomainVerificationEnabled))
	for _, name := range []string{
		FlagFiatConversionEnabled, FlagFiatConversionMode, FlagFiatConversionDEXProfile,
		FlagFiatConversionProfileAuthorityKey, FlagFiatConversionProfileAuthorityID,
		FlagFiatConversionPayoutProfile, FlagFiatConversionStateFile, FlagFiatConversionRepositoryFile,
		FlagFiatConversionPollInterval, FlagFiatConversionRetryBackoff, FlagFiatConversionMaxRetryBackoff,
		FlagFiatConversionMaxAttempts, FlagFiatConversionWebhookListen, FlagFiatConversionWebhookPath,
		FlagFiatConversionWebhookBodyLimit, FlagFiatConversionWebhookTimeout, FlagFiatConversionCustodyBackend,
		FlagFiatConversionSecretResolver, FlagFiatConversionDestinationResolver, FlagFiatConversionComplianceResolver,
	} {
		_ = viper.BindPFlag(name, rootCmd.PersistentFlags().Lookup(name))
	}

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
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func buildDomainVerificationCheckerConfig(
	providerAddress string,
	defaultSignerKey string,
) (provider_daemon.DomainVerificationCheckerConfig, error) {
	cfg := provider_daemon.DefaultDomainVerificationCheckerConfig()

	if viper.IsSet("domain_verification") {
		if err := viper.UnmarshalKey("domain_verification", &cfg); err != nil {
			return cfg, fmt.Errorf("failed to load domain_verification config: %w", err)
		}
	}

	if viper.GetBool(FlagDomainVerificationEnabled) {
		cfg.Enabled = true
	}

	if !cfg.Enabled {
		return cfg, nil
	}

	if cfg.ProviderAddress == "" {
		cfg.ProviderAddress = providerAddress
	}
	if cfg.ChainID == "" {
		cfg.ChainID = viper.GetString(FlagChainID)
	}
	if cfg.CometRPC == "" {
		cfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
	}
	if cfg.GRPCEndpoint == "" {
		cfg.GRPCEndpoint = viper.GetString(FlagWaldurChainGRPC)
	}
	if cfg.SignerKeyName == "" {
		cfg.SignerKeyName = viper.GetString(FlagWaldurChainKey)
		if cfg.SignerKeyName == "" {
			cfg.SignerKeyName = defaultSignerKey
		}
	}
	if cfg.SignerKeyringBackend == "" {
		cfg.SignerKeyringBackend = viper.GetString(FlagWaldurChainKeyringBackend)
	}
	if cfg.SignerKeyringDir == "" {
		cfg.SignerKeyringDir = viper.GetString(FlagWaldurChainKeyringDir)
	}
	if cfg.SignerKeyringPassphrase == "" {
		cfg.SignerKeyringPassphrase = viper.GetString(FlagWaldurChainKeyringPassphrase)
	}
	if cfg.GasSetting.Gas == 0 && !cfg.GasSetting.Simulate {
		gasSetting, err := parseGasSetting(viper.GetString(FlagWaldurChainGas))
		if err != nil {
			return cfg, fmt.Errorf("invalid domain verification gas setting: %w", err)
		}
		cfg.GasSetting = gasSetting
	}
	if cfg.GasPrices == "" {
		cfg.GasPrices = viper.GetString(FlagWaldurChainGasPrices)
	}
	if cfg.Fees == "" {
		cfg.Fees = viper.GetString(FlagWaldurChainFees)
	}
	if cfg.GasAdjustment == 0 {
		cfg.GasAdjustment = viper.GetFloat64(FlagWaldurChainGasAdjustment)
	}
	if cfg.BroadcastTimeout == 0 {
		cfg.BroadcastTimeout = viper.GetDuration(FlagWaldurChainBroadcastTimeout)
	}

	return cfg, nil
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
	if viper.GetBool(FlagProviderProduction) {
		if err := validateProductionDurablePaths(keyDir); err != nil {
			return err
		}
	}
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

	keyPassphrase, err := providerKeyPassphrase(keyConfig.StorageType, viper.GetString(FlagProviderKeyPassphraseFile))
	if err != nil {
		return err
	}
	defer scrubStringBytes(keyPassphrase)
	if err := keyManager.Unlock(string(keyPassphrase)); err != nil {
		return fmt.Errorf("failed to unlock key manager: %w", err)
	}
	fmt.Println("  Key Manager: initialized")

	keyBackupFile := viper.GetString(FlagKeyBackupFile)
	keyBackupPassphrase := viper.GetString(FlagKeyBackupPassphrase)
	if viper.GetBool(FlagKeyBackupRestore) {
		restoreResult, err := restoreProviderKeysFromBackup(keyManager, keyBackupFile, keyBackupPassphrase)
		if err != nil {
			return fmt.Errorf("failed to restore provider key backup: %w", err)
		}
		fmt.Printf("  Key Backup: restored %d key(s), skipped %d\n", len(restoreResult.RestoredKeys), len(restoreResult.SkippedKeys))
	}

	// Initialize provider key material
	providerKeyName := viper.GetString(FlagProviderKey)
	var key *provider_daemon.ManagedKey
	generatedKey := false
	if viper.GetBool(FlagProviderProduction) {
		key, err = keyManager.GetActiveKey()
		if err != nil {
			return fmt.Errorf("production provider key must be provisioned or restored before startup: %w", err)
		}
	} else {
		key, generatedKey, err = ensureProviderKey(keyManager, providerKeyName)
	}
	if err != nil {
		return fmt.Errorf("failed to initialize provider key: %w", err)
	}
	providerAddress := key.ProviderAddress
	if providerAddress == "" {
		providerAddress = providerKeyName
	}
	if _, err := sdk.AccAddressFromBech32(providerAddress); err != nil {
		derivedAddress, deriveErr := provider_daemon.ManagedKeyAccountAddress(key)
		if deriveErr != nil {
			return fmt.Errorf("provider key does not identify a Cosmos account: %w", deriveErr)
		}
		providerAddress = derivedAddress
		key.ProviderAddress = derivedAddress
	}
	providerID := key.PublicKey
	fingerprint, err := keyManager.ActiveKeyFingerprint()
	if err != nil {
		return fmt.Errorf("failed to fingerprint provider key: %w", err)
	}
	if expected := strings.ToLower(strings.TrimSpace(viper.GetString(FlagProviderKeyFingerprint))); expected != "" && expected != fingerprint {
		return fmt.Errorf("provider key fingerprint mismatch")
	}
	if expected := strings.TrimSpace(viper.GetString(FlagProviderKeyFingerprint)); expected != "" {
		if err := keyManager.SetExpectedActiveKeyFingerprint(expected); err != nil {
			return fmt.Errorf("bind provider key fingerprint: %w", err)
		}
	}
	if viper.GetBool(FlagProviderProduction) && strings.TrimSpace(viper.GetString(FlagProviderKeyFingerprint)) == "" {
		return fmt.Errorf("production provider key fingerprint is required")
	}
	if generatedKey {
		fmt.Printf("  Provider Key: generated (%s)\n", key.KeyID)
	} else {
		fmt.Printf("  Provider Key: loaded (%s)\n", key.KeyID)
	}
	fmt.Printf("  Provider ID: %s...\n", providerID[:16])

	if viper.GetBool(FlagKeyBackupExport) {
		backup, err := writeProviderKeyBackup(keyManager, keyBackupFile, keyBackupPassphrase)
		if err != nil {
			return fmt.Errorf("failed to write provider key backup: %w", err)
		}
		keyCount := 0
		if backup.Metadata != nil {
			keyCount = backup.Metadata.KeyCount
		}
		fmt.Printf("  Key Backup: wrote %d key(s) to %s\n", keyCount, keyBackupFile)
	}

	var callbackSink provider_daemon.CallbackSink
	var usageReporter provider_daemon.UsageReporter
	var supportService *provider_daemon.SupportService
	var domainVerificationChecker *provider_daemon.DomainVerificationChecker
	var mutationSubmitter *provider_daemon.ProviderMutationSubmitter
	var waldurMarketplaceClient *waldur.MarketplaceClient
	var waldurReconciler *provider_daemon.WaldurReconciler

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
	productionChainClient, err := provider_daemon.NewProviderRPCChainClient(provider_daemon.RPCChainClientConfig{
		NodeURI:        viper.GetString(FlagNode),
		GRPCEndpoint:   viper.GetString(FlagWaldurChainGRPC),
		ChainID:        viper.GetString(FlagChainID),
		RequestTimeout: time.Second * 30,
	})
	if err != nil {
		return fmt.Errorf("failed to create chain client: %w", err)
	}
	defer func() { _ = productionChainClient.Close() }()
	mutationChain, err := provider_daemon.NewRPCProviderMutationChain(productionChainClient)
	if err != nil {
		return fmt.Errorf("failed to initialize provider mutation transport: %w", err)
	}
	mutationCfg := provider_daemon.DefaultProviderMutationSubmitterConfig()
	mutationCfg.ChainID = viper.GetString(FlagChainID)
	mutationCfg.ProviderAddress = providerAddress
	mutationCfg.QueueStatePath = viper.GetString(FlagChainUsageQueueFile) + ".mutations"
	mutationCfg.Chain = mutationChain
	mutationCfg.Production = viper.GetBool(FlagProviderProduction)
	leaseOwner := strings.TrimSpace(viper.GetString(FlagSubmitterLeaseOwner))
	leaseBackend := strings.ToLower(strings.TrimSpace(viper.GetString(FlagSubmitterLeaseBackend)))
	switch leaseBackend {
	case "kubernetes":
		if leaseOwner != "" {
			clusterConfig, configErr := rest.InClusterConfig()
			if configErr != nil {
				return fmt.Errorf("failed to load Kubernetes submitter lease credentials: %w", configErr)
			}
			clientset, clientErr := kubernetes.NewForConfig(clusterConfig)
			if clientErr != nil {
				return fmt.Errorf("failed to initialize Kubernetes submitter lease client: %w", clientErr)
			}
			lease, leaseErr := provider_daemon.NewKubernetesSubmitterLease(clientset.CoordinationV1(), viper.GetString(FlagSubmitterLeaseNamespace), leaseOwner)
			if leaseErr != nil {
				return fmt.Errorf("failed to initialize Kubernetes submitter lease: %w", leaseErr)
			}
			mutationCfg.Lease = lease
		}
	case "shared_file":
		lease, leaseErr := provider_daemon.NewFileSubmitterLease(viper.GetString(FlagSubmitterLeaseFile), leaseOwner)
		if leaseErr != nil {
			return fmt.Errorf("failed to initialize durable submitter lease: %w", leaseErr)
		}
		mutationCfg.Lease = lease
	default:
		return fmt.Errorf("unsupported submitter lease backend %q", leaseBackend)
	}
	mutationSubmitter, err = provider_daemon.NewProviderMutationSubmitter(mutationCfg, keyManager)
	if err != nil {
		return fmt.Errorf("failed to initialize provider mutation submitter: %w", err)
	}
	if err := mutationSubmitter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start provider mutation submitter: %w", err)
	}
	productionChainClient.SetMutationSubmitter(mutationSubmitter)
	productionChainClient.SetMutationGuard(func(guardCtx context.Context) error {
		readiness := mutationSubmitter.Readiness(guardCtx)
		if readiness.Ready {
			return nil
		}
		return fmt.Errorf("%w: %s", provider_daemon.ErrProviderMutationNotReady, readiness.Reason)
	})
	if readiness := mutationSubmitter.Readiness(ctx); !readiness.Ready {
		if !mutationCfg.Production || readiness.Reason != "submitter lease not held" {
			return fmt.Errorf("provider mutation submitter not ready: %s", readiness.Reason)
		}
		fmt.Println("  Provider Mutation Submitter: standby (waiting for durable fenced ownership)")
	}
	fiatQuery, err := provider_daemon.NewRPCFiatConversionQuery(productionChainClient)
	if err != nil && viper.GetBool(FlagFiatConversionEnabled) {
		return fmt.Errorf("fiat conversion startup blocked: %w", err)
	}
	if err := validateFiatConversionStartup(ctx, fiatQuery, mutationSubmitter); err != nil {
		return err
	}
	defer func(submitter *provider_daemon.ProviderMutationSubmitter) {
		drainCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := submitter.Stop(drainCtx); err != nil {
			fmt.Printf("  Provider Mutation Submitter: failed to drain cleanly: %v\n", err)
		}
	}(mutationSubmitter)
	fmt.Println("  Provider Mutation Submitter: started (durable signed pipeline)")

	if viper.GetBool(FlagWaldurEnabled) && viper.GetBool(FlagWaldurChainSubmit) {
		callbackSink, err = provider_daemon.NewDurableChainCallbackSink(providerAddress, mutationSubmitter)
		if err != nil {
			return fmt.Errorf("failed to initialize durable Waldur callback sink: %w", err)
		}
	}
	var chainClient provider_daemon.ChainClient = productionChainClient

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
		hpcProvider, err = provider_daemon.NewHPCProviderWithDeps(
			hpcProviderConfig,
			productionChainClient,
			nil,
			&provider_daemon.HPCProviderDeps{
				Signer: newHPCKeyManagerSigner(keyManager, hpcProviderConfig.HPC.ProviderAddress),
			},
		)
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

	// Initialize resource availability sync (RES-36C)
	if resourceChainClient, ok := chainClient.(provider_daemon.ResourceChainClient); ok {
		providerConfig, err := chainClient.GetProviderConfig(ctx, providerAddress)
		if err != nil {
			fmt.Printf("Warning: failed to load provider config for resource sync: %v\n", err)
		} else {
			resourceCfg := provider_daemon.DefaultResourceSyncConfig()
			resourceCfg.ProviderAddress = providerAddress
			resourceCfg.InventoryID = fmt.Sprintf("%s-compute", providerAddress)
			resourceCfg.ResourceClass = resourcesv1.ResourceClass_RESOURCE_CLASS_COMPUTE
			if len(providerConfig.Regions) > 0 {
				resourceCfg.Region = providerConfig.Regions[0]
			}

			snapshotCapacity := providerConfig.Capacity
			snapshotCapacity.ReservedCPUCores = 0
			snapshotCapacity.ReservedMemoryGB = 0
			snapshotCapacity.ReservedStorageGB = 0
			snapshotCapacity.ReservedGPUs = 0

			snapshot := provider_daemon.NewStaticResourceSnapshotProvider(
				snapshotCapacity,
				providerConfig.Attributes["gpu_type"],
				resourceCfg.TotalNetworkMbps,
				resourceCfg.ReservedNetwork,
			)
			resourceSync, err := provider_daemon.NewResourceAvailabilitySync(resourceCfg, resourceChainClient, snapshot)
			if err != nil {
				fmt.Printf("Warning: failed to initialize resource sync: %v\n", err)
			} else if err := resourceSync.Start(ctx); err != nil {
				fmt.Printf("Warning: failed to start resource sync: %v\n", err)
			} else {
				fmt.Println("  Resource Sync: started")
			}
		}
	}

	domainVerificationCfg, err := buildDomainVerificationCheckerConfig(providerAddress, providerKeyName)
	if err != nil {
		return err
	}
	if domainVerificationCfg.Enabled {
		if _, err := sdk.AccAddressFromBech32(domainVerificationCfg.ProviderAddress); err != nil {
			return fmt.Errorf(
				"domain verification provider address must be bech32; set domain_verification.provider_address to the on-chain owner address: %w",
				err,
			)
		}

		checker, err := provider_daemon.NewDomainVerificationChecker(domainVerificationCfg, keyManager, productionChainClient)
		if err != nil {
			return fmt.Errorf("failed to initialize domain verification checker: %w", err)
		}
		if err := checker.Start(ctx); err != nil {
			return fmt.Errorf("failed to start domain verification checker: %w", err)
		}
		domainVerificationChecker = checker
		fmt.Println("  Domain Verification Checker: started")
	}

	// Initialize Kubernetes adapter (VE-403)
	statusUpdateChan := make(chan provider_daemon.WorkloadStatusUpdate, 100)
	portalLogStore := provider_daemon.NewDeploymentLogStore()
	provisioningEnabled := viper.GetBool(FlagProvisioningEnabled)
	provisioningDryRun := viper.GetBool(FlagProvisioningDryRun)

	var workloadRuntime *kubernetesWorkloadRuntime
	if provisioningEnabled {
		workloadRuntime, err = newKubernetesWorkloadRuntime(kubernetesRuntimeConfig{
			ProviderID:        providerID,
			ResourcePrefix:    viper.GetString(FlagResourcePrefix),
			Kubeconfig:        viper.GetString(FlagKubeconfig),
			DryRun:            provisioningDryRun,
			ReconcileInterval: viper.GetDuration(FlagProvisioningPollInterval),
			StatusUpdateChan:  statusUpdateChan,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize kubernetes runtime: %w", err)
		}
		workloadRuntime.Start(ctx)
		if provisioningDryRun {
			fmt.Println("  Kubernetes Adapter: initialized (dry-run)")
		} else {
			fmt.Println("  Kubernetes Adapter: initialized")
		}
	} else {
		fmt.Println("  Kubernetes Adapter: disabled")
	}

	// Initialize usage meter (VE-404)
	recordChan := make(chan *provider_daemon.UsageRecord, 100)
	usageStore := provider_daemon.NewUsageSnapshotStore()
	usageReporter = usageStore
	var usageMeter *provider_daemon.UsageMeter
	var usageSubmitter *provider_daemon.ChainUsageSubmitterImpl
	var usagePipelineMu sync.Mutex
	stopUsagePipeline := func() {
		usagePipelineMu.Lock()
		defer usagePipelineMu.Unlock()
		if usageMeter != nil {
			usageMeter.Stop()
			usageMeter = nil
		}
		if usageSubmitter != nil {
			usageSubmitter.Stop()
			usageSubmitter = nil
		}
	}
	if viper.GetBool(FlagChainUsageSubmit) {
		if _, err := sdk.AccAddressFromBech32(providerAddress); err != nil {
			return fmt.Errorf("provider key name must be the on-chain bech32 provider address for authenticated metering: %w", err)
		}
		if workloadRuntime == nil {
			return fmt.Errorf("authenticated usage metering requires an initialized workload metrics collector")
		}
		startUsagePipeline := func() error {
			usagePipelineMu.Lock()
			defer usagePipelineMu.Unlock()
			if usageMeter != nil || !mutationSubmitter.Readiness(ctx).Ready {
				return nil
			}
			usageSubmitterCfg := provider_daemon.DefaultChainSubmitterConfig()
			usageSubmitterCfg.ProviderAddress = providerAddress
			usageSubmitterCfg.ChainID = viper.GetString(FlagChainID)
			usageSubmitterCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
			usageSubmitterCfg.QueueStatePath = viper.GetString(FlagChainUsageQueueFile)
			usageSubmitterCfg.ProviderSigningState = productionChainClient
			usageSubmitterCfg.UsageStreamState = productionChainClient
			usageSubmitterCfg.MutationSubmitter = mutationSubmitter
			candidateSubmitter, createErr := provider_daemon.NewChainUsageSubmitter(usageSubmitterCfg, keyManager, provider_daemon.NewUsageMetricsCollector())
			if createErr != nil {
				return fmt.Errorf("initialize authenticated usage submitter: %w", createErr)
			}
			if startErr := candidateSubmitter.Start(ctx); startErr != nil {
				candidateSubmitter.Stop()
				return fmt.Errorf("start authenticated usage submitter: %w", startErr)
			}
			usageRecorder, recorderErr := provider_daemon.NewUsageChainRecorder(candidateSubmitter)
			if recorderErr != nil {
				candidateSubmitter.Stop()
				return fmt.Errorf("initialize usage chain recorder: %w", recorderErr)
			}
			candidateMeter := provider_daemon.NewUsageMeter(provider_daemon.UsageMeterConfig{
				ProviderID: providerAddress, Interval: provider_daemon.MeteringInterval(viper.GetDuration(FlagMeteringInterval)),
				MetricsCollector: workloadRuntime, ChainRecorder: usageRecorder, KeyManager: keyManager, RecordChan: recordChan,
			})
			if startErr := candidateMeter.Start(ctx); startErr != nil {
				candidateSubmitter.Stop()
				return fmt.Errorf("start usage meter: %w", startErr)
			}
			usageSubmitter = candidateSubmitter
			usageMeter = candidateMeter
			fmt.Println("  Usage Meter: started by fenced submitter owner")
			return nil
		}
		if err := startUsagePipeline(); err != nil {
			return err
		}
		if viper.GetBool(FlagProviderProduction) {
			go func() {
				ticker := time.NewTicker(2 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
						if mutationSubmitter.Readiness(ctx).Ready {
							if startErr := startUsagePipeline(); startErr != nil {
								fmt.Printf("  Usage Meter: fenced startup deferred: %v\n", startErr)
							}
						} else {
							stopUsagePipeline()
						}
					}
				}
			}()
		}
	} else {
		fmt.Println("  Usage Meter: disabled (enable authenticated submission explicitly)")
	}

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
		waldurMarketplaceClient = marketplaceClient

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
			grpc.WithTransportCredentials(credentials.NewTLS(security.SecureTLSConfig())),
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
	portalCfg.Readiness = func(readinessCtx context.Context) provider_daemon.ProviderMutationReadiness {
		status := mutationSubmitter.Readiness(readinessCtx)
		if status.Ready && viper.GetBool(FlagChainUsageSubmit) {
			usagePipelineMu.Lock()
			usageReady := usageMeter != nil && usageSubmitter != nil
			usagePipelineMu.Unlock()
			if !usageReady {
				status.Ready = false
				status.Reason = "usage sequence and queue store unavailable"
			}
		}
		return status
	}
	portalCfg.LogStore = portalLogStore
	portalCfg.WalletAuthChainID = viper.GetString(FlagChainID)
	portalCfg.ChainQuery = chainQuery
	portalCfg.VaultService = vaultService
	portalCfg.LifecycleExecutor = lifecycleManager
	portalCfg.LifecycleRequireConsent = viper.GetBool(FlagWaldurLifecycleRequireConsent)
	portalCfg.LifecycleConsentScope = viper.GetString(FlagWaldurLifecycleConsentScope)
	portalCfg.LifecycleAllowedRoles = parseCSVList(viper.GetString(FlagWaldurLifecycleAllowedRoles))
	if workloadRuntime != nil && workloadRuntime.client != nil {
		portalCfg.WorkloadLogSource = workloadRuntime
		portalCfg.WorkloadShellExecutor = workloadRuntime
	}

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

		if waldurMarketplaceClient != nil {
			reconcilerCfg := provider_daemon.DefaultWaldurReconcilerConfig()
			waldurReconciler = provider_daemon.NewWaldurReconciler(
				reconcilerCfg,
				waldurMarketplaceClient,
				usageStore,
				nil,
				provider_daemon.NewWaldurBridgeStateStore(viper.GetString(FlagWaldurStateFile)),
			)
			if err := waldurReconciler.Start(ctx); err != nil {
				return fmt.Errorf("failed to start waldur reconciler: %w", err)
			}
			fmt.Println("  Waldur Reconciler: started")
		}
	}

	// Initialize provisioning worker (VE-36F)
	if viper.GetBool(FlagProvisioningEnabled) {
		if callbackSink == nil {
			return fmt.Errorf("provisioning requires durable chain callback sink; enable %s with %s", FlagWaldurChainSubmit, FlagWaldurEnabled)
		}

		provCfg := provider_daemon.DefaultProvisioningConfig()
		provCfg.Enabled = true
		provCfg.ProviderAddress = providerAddress
		provCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
		provCfg.CometWS = viper.GetString(FlagCometWS)
		provCfg.EventQuery = viper.GetString(FlagMarketplaceEventQuery)
		provCfg.StateFile = viper.GetString(FlagProvisioningStateFile)
		provCfg.CheckpointFile = viper.GetString(FlagProvisioningCheckpointFile)
		provCfg.MaxRetries = viper.GetInt(FlagProvisioningMaxRetries)
		provCfg.RetryBackoff = viper.GetDuration(FlagProvisioningRetryBackoff)
		provCfg.MaxBackoff = viper.GetDuration(FlagProvisioningMaxBackoff)
		provCfg.PollInterval = viper.GetDuration(FlagProvisioningPollInterval)
		provCfg.RetryOnFailure = true

		provisioners := make([]provider_daemon.Provisioner, 0, 2)

		// Waldur/OpenStack provisioning
		waldurBase := viper.GetString(FlagWaldurBaseURL)
		waldurToken := viper.GetString(FlagWaldurToken)
		if waldurBase != "" && waldurToken != "" {
			waldurCfg := waldur.DefaultConfig()
			waldurCfg.BaseURL = waldurBase
			waldurCfg.Token = waldurToken
			waldurClient, err := waldur.NewClient(waldurCfg)
			if err != nil {
				return fmt.Errorf("failed to create waldur client for provisioning: %w", err)
			}
			marketplaceClient := waldur.NewMarketplaceClient(waldurClient)
			lifecycleClient := waldur.NewLifecycleClient(marketplaceClient)
			stateStore := provider_daemon.NewWaldurBridgeStateStore(viper.GetString(FlagWaldurStateFile))
			resolver := provider_daemon.NewWaldurStateResolver(stateStore)
			provisioners = append(provisioners, provider_daemon.NewWaldurProvisioner(lifecycleClient, resolver))
		}

		// Container provisioning via the shared Kubernetes runtime.
		dryRun := viper.GetBool(FlagProvisioningDryRun)
		if workloadRuntime == nil || workloadRuntime.adapter == nil {
			return fmt.Errorf("kubernetes runtime not initialized")
		}
		provisioners = append(provisioners, provider_daemon.NewContainerProvisioner(workloadRuntime.adapter, 5*time.Minute, dryRun))

		provisioningWorker, err := provider_daemon.NewProvisioningWorker(provCfg, keyManager, callbackSink, provisioners...)
		if err != nil {
			return fmt.Errorf("failed to create provisioning worker: %w", err)
		}
		go func() {
			if err := provisioningWorker.Start(ctx); err != nil {
				fmt.Printf("[PROVISIONING] worker stopped: %v\n", err)
			}
		}()
		fmt.Println("  Provisioning Worker: started")
	}

	// Initialize support service desk bridge (VE-25C)
	if viper.GetBool(FlagSupportEnabled) {
		chainKeyName := viper.GetString(FlagWaldurChainKey)
		if chainKeyName == "" {
			chainKeyName = providerKeyName
		}
		gasSetting, err := parseGasSetting(viper.GetString(FlagWaldurChainGas))
		if err != nil {
			return fmt.Errorf("invalid support chain gas: %w", err)
		}

		supportCfg := provider_daemon.DefaultSupportServiceConfig()
		supportCfg.Enabled = true
		supportCfg.ProviderAddress = providerAddress
		supportCfg.ChainID = viper.GetString(FlagChainID)
		supportCfg.CometRPC = normalizeCometRPC(viper.GetString(FlagNode))
		supportCfg.CometWS = viper.GetString(FlagCometWS)
		supportCfg.GRPCEndpoint = viper.GetString(FlagWaldurChainGRPC)
		supportCfg.SignerKeyName = chainKeyName
		supportCfg.SignerKeyringBackend = viper.GetString(FlagWaldurChainKeyringBackend)
		supportCfg.SignerKeyringDir = viper.GetString(FlagWaldurChainKeyringDir)
		supportCfg.SignerKeyringPassphrase = viper.GetString(FlagWaldurChainKeyringPassphrase)
		supportCfg.GasSetting = gasSetting
		supportCfg.GasPrices = viper.GetString(FlagWaldurChainGasPrices)
		supportCfg.Fees = viper.GetString(FlagWaldurChainFees)
		supportCfg.GasAdjustment = viper.GetFloat64(FlagWaldurChainGasAdjustment)
		supportCfg.BroadcastTimeout = viper.GetDuration(FlagWaldurChainBroadcastTimeout)
		supportCfg.RequestTimeout = 30 * time.Second
		supportCfg.MutationSubmitter = mutationSubmitter
		supportCfg.StoreQuery = productionChainClient.ProviderStoreQueryClient()
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
	go handleStatusUpdates(ctx, statusUpdateChan, portalLogStore)
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

	stopUsagePipeline()
	fmt.Println("  Usage Meter: stopped")

	if supportService != nil {
		_ = supportService.Stop(ctx)
		fmt.Println("  Support Service: stopped")
	}

	if waldurReconciler != nil {
		waldurReconciler.Stop()
		fmt.Println("  Waldur Reconciler: stopped")
	}

	if domainVerificationChecker != nil {
		if err := domainVerificationChecker.Stop(); err != nil {
			fmt.Printf("  Domain Verification Checker: failed to stop cleanly: %v\n", err)
		} else {
			fmt.Println("  Domain Verification Checker: stopped")
		}
	}

	if mutationSubmitter != nil {
		drainDeadline := time.Now().Add(30 * time.Second)
		for {
			metrics := mutationSubmitter.Metrics(context.Background())
			if metrics.QueueDepth == 0 || time.Now().After(drainDeadline) {
				if metrics.QueueDepth > 0 {
					fmt.Printf("  Provider Mutation Submitter: drain deadline reached with %d explicit queued item(s)\n", metrics.QueueDepth)
				}
				break
			}
			processCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			_ = mutationSubmitter.ProcessDue(processCtx, 32)
			cancel()
		}
		stopCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := mutationSubmitter.Stop(stopCtx); err != nil {
			fmt.Printf("  Provider Mutation Submitter: failed to stop cleanly: %v\n", err)
		} else {
			fmt.Println("  Provider Mutation Submitter: drained and stopped")
		}
		cancel()
		mutationSubmitter = nil
	}

	keyManager.Lock()
	fmt.Println("  Key Manager: locked")

	fmt.Println("Provider daemon stopped.")
	return nil
}

// validateFiatConversionStartup enforces the production composition boundary.
// External custody, vendor, secret, destination and compliance implementations
// are intentionally not embedded in this binary, so identifiers alone never
// promote local engineering code to an executable production corridor.
func validateFiatConversionStartup(ctx context.Context, query *provider_daemon.RPCFiatConversionQuery, submitter *provider_daemon.ProviderMutationSubmitter) error {
	if !viper.GetBool(FlagFiatConversionEnabled) {
		fmt.Println("  Fiat Conversion Orchestrator: disabled (external route/corridor certification not claimed)")
		return nil
	}
	mode := strings.TrimSpace(viper.GetString(FlagFiatConversionMode))
	var authority provider_daemon.FiatProfileAuthority
	encodedAuthorityKey := strings.TrimSpace(viper.GetString(FlagFiatConversionProfileAuthorityKey))
	authorityID := strings.TrimSpace(viper.GetString(FlagFiatConversionProfileAuthorityID))
	if encodedAuthorityKey != "" || authorityID != "" {
		publicKey, decodeErr := base64.StdEncoding.Strict().DecodeString(encodedAuthorityKey)
		if decodeErr != nil {
			return fmt.Errorf("fiat conversion profile authority key: invalid base64")
		}
		configuredAuthority, authorityErr := provider_daemon.NewEd25519FiatProfileAuthority(authorityID, publicKey)
		if authorityErr != nil {
			return fmt.Errorf("fiat conversion profile authority: %w", authorityErr)
		}
		authority = configuredAuthority
	}
	profiles, err := provider_daemon.LoadTrustedFiatProfilesWithAuthority(
		viper.GetString(FlagFiatConversionDEXProfile),
		viper.GetString(FlagFiatConversionPayoutProfile),
		authority,
	)
	if err != nil {
		return fmt.Errorf("fiat conversion profiles: %w", err)
	}
	if err := validateFiatConversionProfileMode(mode, profiles); err != nil {
		return err
	}
	if mode == "engineering_external_blocked" {
		fmt.Println("  Fiat Conversion Orchestrator: engineering profile verified; external execution remains blocked until reviewed dependencies are injected")
		return nil
	}
	if query == nil || submitter == nil || !submitter.Readiness(ctx).Ready {
		return fmt.Errorf("fiat conversion startup blocked: query or mutation transport unavailable")
	}
	params, err := query.Params(ctx)
	if err != nil {
		return fmt.Errorf("fiat conversion startup blocked: %w", err)
	}
	if params.FiatConversionDexProfileId != profiles.DEX.ID || params.FiatConversionPayoutProfileId != profiles.Payout.ID ||
		!bytes.Equal(params.FiatConversionDexProfileDigest, profiles.DEXDigest[:]) ||
		!bytes.Equal(params.FiatConversionPayoutProfileDigest, profiles.PayoutDigest[:]) {
		return fmt.Errorf("fiat conversion startup blocked: chain profile commitments do not match local files")
	}
	missing := make([]string, 0, 4)
	for flag, value := range map[string]string{
		FlagFiatConversionCustodyBackend:      viper.GetString(FlagFiatConversionCustodyBackend),
		FlagFiatConversionSecretResolver:      viper.GetString(FlagFiatConversionSecretResolver),
		FlagFiatConversionDestinationResolver: viper.GetString(FlagFiatConversionDestinationResolver),
		FlagFiatConversionComplianceResolver:  viper.GetString(FlagFiatConversionComplianceResolver),
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, flag)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("fiat conversion startup blocked: external dependency composition unavailable (%s)", strings.Join(missing, ", "))
	}
	return fmt.Errorf("fiat conversion startup blocked: configured external backend identifiers have no in-binary production factories; inject reviewed DEX custody, payout partner, secret, destination, compliance, and webhook implementations")
}

func validateFiatConversionProfileMode(mode string, profiles *provider_daemon.TrustedFiatProfiles) error {
	if profiles == nil {
		return fmt.Errorf("fiat conversion startup blocked: profiles unavailable")
	}
	if mode == "engineering_external_blocked" {
		if profiles.DEX.State != dex.RouteEngineeringCompleteExternalBlocked || profiles.Payout.State != offramp.ProfileEngineeringCompleteExternalBlocked {
			return fmt.Errorf("fiat conversion engineering mode requires explicit external-blocked profile rows")
		}
		return nil
	}
	if mode != "production" {
		return fmt.Errorf("invalid fiat conversion mode %q", mode)
	}
	if !profiles.DEXTrusted || !profiles.PayoutTrusted {
		return fmt.Errorf("fiat conversion startup blocked: certified profiles are not independently authorized")
	}
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

func handleStatusUpdates(ctx context.Context, ch <-chan provider_daemon.WorkloadStatusUpdate, logStore *provider_daemon.DeploymentLogStore) {
	for {
		select {
		case <-ctx.Done():
			return
		case update, ok := <-ch:
			if !ok {
				return
			}
			if logStore != nil {
				deploymentKey := update.DeploymentID
				if deploymentKey == "" {
					deploymentKey = update.WorkloadID
				}
				logStore.Append(deploymentKey, provider_daemon.LogEntry{
					Timestamp: update.Timestamp,
					Level:     "info",
					Message:   fmt.Sprintf("state=%s %s", update.State, update.Message),
				})
			}
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

			passphrase, err := providerKeyPassphrase(provider_daemon.KeyStorageTypeFile, viper.GetString(FlagProviderKeyPassphraseFile))
			if err != nil {
				return err
			}
			defer scrubStringBytes(passphrase)
			if err := keyManager.Unlock(string(passphrase)); err != nil {
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

func providerKeyPassphrase(storageType provider_daemon.KeyStorageType, path string) ([]byte, error) {
	if storageType == provider_daemon.KeyStorageTypeMemory {
		return nil, nil
	}
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("file key storage requires --%s from a mounted secret", FlagProviderKeyPassphraseFile)
	}
	value, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read provider key passphrase file: %w", err)
	}
	value = bytes.TrimSpace(value)
	if len(value) == 0 {
		return nil, fmt.Errorf("provider key passphrase must not be empty")
	}
	return value, nil
}

func scrubStringBytes(value []byte) {
	for i := range value {
		value[i] = 0
	}
}

func validateProductionDurablePaths(keyDir string) error {
	if strings.TrimSpace(keyDir) == "" || !filepath.IsAbs(keyDir) {
		return fmt.Errorf("production provider key directory must be an absolute durable path")
	}
	paths := map[string]string{
		FlagChainUsageQueueFile:          viper.GetString(FlagChainUsageQueueFile),
		FlagWaldurStateFile:              viper.GetString(FlagWaldurStateFile),
		FlagWaldurCheckpointFile:         viper.GetString(FlagWaldurCheckpointFile),
		FlagWaldurOrderStateFile:         viper.GetString(FlagWaldurOrderStateFile),
		FlagWaldurOrderCheckpointFile:    viper.GetString(FlagWaldurOrderCheckpointFile),
		FlagProvisioningStateFile:        viper.GetString(FlagProvisioningStateFile),
		FlagProvisioningCheckpointFile:   viper.GetString(FlagProvisioningCheckpointFile),
		FlagFiatConversionStateFile:      viper.GetString(FlagFiatConversionStateFile),
		FlagFiatConversionRepositoryFile: viper.GetString(FlagFiatConversionRepositoryFile),
		FlagWaldurOfferingSyncStateFile:  viper.GetString(FlagWaldurOfferingSyncStateFile),
		FlagWaldurLifecycleQueuePath:     viper.GetString(FlagWaldurLifecycleQueuePath),
		FlagWaldurCallbackSinkDir:        viper.GetString(FlagWaldurCallbackSinkDir),
		FlagPortalAuditLogFile:           viper.GetString(FlagPortalAuditLogFile),
	}
	if strings.EqualFold(strings.TrimSpace(viper.GetString(FlagSubmitterLeaseBackend)), "shared_file") {
		paths[FlagSubmitterLeaseFile] = viper.GetString(FlagSubmitterLeaseFile)
	}
	root := filepath.Clean(filepath.Dir(keyDir))
	for name, path := range paths {
		if strings.TrimSpace(path) == "" || !filepath.IsAbs(path) {
			return fmt.Errorf("production --%s must be an absolute durable path", name)
		}
		relative, err := filepath.Rel(root, filepath.Clean(path))
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("production --%s must be rooted under durable provider volume %s", name, root)
		}
	}
	return nil
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
