/**
 * Portal Adapter
 * Re-exports from lib/portal for Next.js integration
 *
 * This adapter provides a clean import path for portal library components,
 * hooks, and utilities within the Next.js application.
 */

import {
  WalletProvider as BaseWalletProvider,
  useWallet as baseUseWallet,
} from '../../../lib/portal';
import type { WalletContextValue } from '../../../lib/portal';

// ============================================================================
// Core Exports
// ============================================================================

export { PortalProvider, usePortal } from '../../../lib/portal';
export type { PortalProviderProps, PortalConfig } from '../../../lib/portal';

// ============================================================================
// Authentication (VE-700)
// ============================================================================

export { useAuth, AuthProvider } from '../../../lib/portal';
export type {
  AuthState,
  AuthActions,
  WalletCredentials,
  SSOCredentials,
  SessionToken,
  AuthError,
} from '../../../lib/portal';

export { SessionManager } from '../../../lib/portal';
export type { SessionConfig, SessionInfo } from '../../../lib/portal';

export { WalletAdapter, MnemonicWallet, KeypairWallet } from '../../../lib/portal';
export type { WalletConfig, SigningResult } from '../../../lib/portal';

// ============================================================================
// Wallet Connections
// ============================================================================

export const WalletProvider = BaseWalletProvider;
export const useWallet = baseUseWallet as () => WalletContextValue;
export type {
  PortalWalletType as PortalWalletType,
  WalletConnectionStatus,
  WalletChainInfo,
  WalletAccount,
  WalletErrorType as WalletError,
  WalletState,
  WalletSignOptions,
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  WalletProviderConfig,
} from '../../../lib/portal';

export {
  WalletButton,
  WalletAccountDisplay,
  WalletNetworkBadge,
  WalletModal,
  WalletIcon,
  WalletSkeleton,
  AccountSelector,
  TransactionModal,
} from '../../../lib/portal';
export type {
  WalletButtonProps,
  WalletAccountDisplayProps,
  WalletNetworkBadgeProps,
  WalletModalProps,
  WalletOption,
  WalletIconProps,
  WalletSkeletonProps,
  AccountSelectorProps,
  TransactionModalProps,
} from '../../../lib/portal';

// ============================================================================
// Wallet Utilities (Enhanced)
// ============================================================================

export {
  // Error handling
  WalletErrorClass,
  WalletErrorCode,
  WALLET_ERROR_MESSAGES,
  createWalletError,
  isWalletError,
  getErrorMessage,
  getSuggestedAction,
  parseWalletError,
  isRetryableError,
  withWalletTimeout,
  wrapWithWalletError,
  // Session management
  WalletSessionManager,
  walletSessionManager,
  createSessionManager,
  // Wallet detection
  WalletDetector,
  walletDetector,
  WalletPriority,
  // Transaction utilities
  GAS_TIERS,
  DEFAULT_GAS_ADJUSTMENT,
  DEFAULT_GAS_LIMIT,
  estimateGas,
  calculateFee,
  adjustGas,
  formatFeeAmount,
  createTransactionPreview,
  validateTransaction,
  createDefaultGasSettings,
} from '../../../lib/portal';
export type {
  WalletSession,
  WalletSessionConfig,
  WalletDetectionResult,
  GasTier,
  GasSettings,
  FeeEstimate,
  TransactionPreview,
  TransactionOptions,
  TransactionValidationResult,
} from '../../../lib/portal';

// ============================================================================
// Identity / VEID (VE-701)
// ============================================================================

export { useIdentity, IdentityProvider } from '../../../lib/portal';
export type {
  IdentityState,
  IdentityStatus,
  IdentityScore,
  IdentityTier,
  VerificationScope,
  VerificationScopeType,
  UploadRecord,
  VerificationRecord,
  IdentityGatingError,
  MarketplaceAction,
  RemediationPath,
  ScopeRequirement,
} from '../../../lib/portal';

export { getTierFromScore } from '../../../lib/portal';

export {
  IdentityStatusCard,
  IdentityScoreDisplay,
  ScopeRequirements,
  UploadHistory,
  RemediationGuide,
} from '../../../lib/portal';

// ============================================================================
// MFA (VE-702)
// ============================================================================

export { useMFA, MFAProvider } from '../../../lib/portal';
export type {
  MFAState,
  MFAFactor,
  MFAFactorType,
  MFAFactorStatus,
  MFAFactorMetadata,
  MFAPolicy,
  MFAEnrollment,
  MFAEnrollmentStep,
  MFAEnrollmentChallengeData,
  TrustedBrowser,
  MFAChallenge,
  MFAChallengeType,
  MFAChallengeResponse,
  MFAAuditEntry,
  MFAError,
  MFAErrorCode,
  SensitiveTransactionType,
} from '../../../lib/portal';

export {
  MFAEnrollmentWizard,
  MFAPolicyConfig,
  TrustedBrowserManager,
  MFAPrompt,
  MFAAuditLog,
} from '../../../lib/portal';

// ============================================================================
// Marketplace (VE-703)
// ============================================================================

export { useMarketplace, MarketplaceProvider } from '../../../lib/portal';
export type {
  MarketplaceState,
  Offering,
  OfferingFilter,
  OfferingSortField,
  Order,
  OrderState,
  OrderEvent,
  CheckoutRequest,
  CheckoutValidation,
} from '../../../lib/portal';

export {
  OfferingList,
  OfferingCard,
  OfferingDetail,
  CheckoutFlow,
  OrderDetail,
  OrderTimeline,
} from '../../../lib/portal';

// ============================================================================
// Order Tracking (VE-707)
// ============================================================================

export { OrderTrackingProvider, useOrderTracking } from '../../../lib/portal';
export type {
  OrderTrackingState,
  OrderConnectionStatus,
  OrderResourceConnection,
  OrderCredential,
  OrderApiEndpoint,
  OrderResourceAccess,
  OrderUsageMetric,
  OrderUsageSample,
  OrderUsageAlert,
  OrderUsageSnapshot,
  OrderArtifact,
  OrderTrackingOrder,
  OrderTrackingStateValue,
  OrderTrackingActions,
  OrderTrackingContextValue,
  OrderTrackingProviderProps,
} from '../../../lib/portal';

export {
  OrderList,
  OrderStatus,
  ResourceAccess,
  UsageMonitor,
  OrderTrackingPage,
} from '../../../lib/portal';
export type {
  OrderListProps,
  OrderListFilter,
  OrderStatusProps,
  ResourceAccessProps,
  UsageMonitorProps,
  OrderTrackingPageProps,
} from '../../../lib/portal';

// ============================================================================
// Provider Console (VE-704)
// ============================================================================

export { useProvider, ProviderProvider } from '../../../lib/portal';
export type {
  ProviderState,
  ProviderProfile,
  ProviderRegistration,
  DomainVerification,
  OfferingDraft,
  PricingConfig,
  CapacityConfig,
  BidRecord,
  AllocationRecord,
  UsageRecord,
  SettlementSummary,
} from '../../../lib/portal';

export {
  ProviderRegistrationFlow,
  OfferingEditor,
  PricingEditor,
  CapacityMonitor,
  BidDashboard,
  AllocationList,
  UsageReports,
  SettlementView,
  DomainVerificationPanel,
} from '../../../lib/portal';

// ============================================================================
// Provider API + Multi-Provider (VE-29D/29E/29G)
// ============================================================================

export {
  ProviderAPIClient,
  ProviderAPIError,
  LogStream,
  ShellConnection,
  signRequest,
  MultiProviderClient,
  MultiProviderProvider,
  useMultiProvider,
  useAggregatedDeployments,
  useAggregatedMetrics,
  useDeploymentWithProvider,
} from '../../../lib/portal';
export type {
  ProviderAPIClientOptions,
  ProviderHealthStatus,
  ProviderHealth,
  LogOptions,
  DeploymentState,
  UsageMetric,
  ResourceMetrics,
  Deployment,
  DeploymentStatus,
  ServiceStatus,
  DeploymentListResponse,
  DeploymentAction,
  ShellSessionResponse,
  ProviderAPIErrorDetails,
  SignedRequestHeaders,
  SignRequestOptions,
  ProviderRecord,
  ProviderStatus,
  DeploymentWithProvider,
  AggregatedMetrics,
  MultiProviderWallet,
  MultiProviderClientOptions,
  MultiProviderProviderProps,
  AggregatedDeploymentsState,
  AggregatedDeploymentsActions,
  UseAggregatedDeploymentsOptions,
  AggregatedMetricsState,
  AggregatedMetricsActions,
  UseAggregatedMetricsOptions,
  DeploymentWithProviderState,
} from '../../../lib/portal';

// ============================================================================
// HPC / Supercomputer (VE-705)
// ============================================================================

export { useHPC, HPCProvider } from '../../../lib/portal';
export type {
  HPCState,
  WorkloadTemplate,
  WorkloadCategory,
  JobResources,
  JobParameter,
  JobManifest,
  JobSubmission,
  JobSubmissionState,
  JobSubmissionStep,
  JobPriceQuote,
  JobValidationError,
  Job,
  JobStatus,
  JobStatusChange,
  JobEvent,
  JobEventType,
  JobOutputReference,
  JobOutputType,
  JobOutput,
  HPCError,
  HPCErrorCode,
} from '../../../lib/portal';

export {
  WorkloadLibrary,
  JobSubmissionForm,
  JobTracker,
  JobOutputViewer,
  JobCancelDialog,
} from '../../../lib/portal';

// ============================================================================
// Chain Integration
// ============================================================================

export { useChain, ChainProvider } from '../../../lib/portal';
export type {
  ChainState,
  ChainConfig,
  EventSubscription,
  QueryClient,
  TransactionResult,
} from '../../../lib/portal';

// ============================================================================
// Utilities
// ============================================================================

export {
  formatScore,
  formatTokenAmount,
  formatDuration,
  formatTimestamp,
} from '../../../lib/portal';
export { validateAddress, validateMnemonic, isValidScore } from '../../../lib/portal';
export {
  sanitizePlainText,
  sanitizeDigits,
  sanitizeJsonInput,
  sanitizeObject,
} from '../../../lib/portal';
export { encryptPayload, decryptPayload } from '../../../lib/portal';
export type { EncryptionResult, DecryptionResult } from '../../../lib/portal';

// ============================================================================
// Accessibility (VE-UI-002)
// ============================================================================

export {
  SrOnly,
  SkipLink,
  AccessibleButton,
  AccessibleInput,
  AccessibleSelect,
  AccessibleCheckbox,
  AccessibleAlert,
  AccessibleProgress,
  generateA11yId,
  announce,
  clearAnnouncements,
  initLiveRegions,
  createFocusTrap,
  getFocusableElements,
  meetsContrastRequirement,
  getContrastRatio,
  prefersReducedMotion,
  prefersHighContrast,
  srOnlyStyles,
  focusVisibleStyles,
} from '../../../lib/portal';

// ============================================================================
// Organization Management (VE-29H)
// ============================================================================

export { useOrganization, OrganizationProvider } from '../../../lib/portal';
export type {
  OrganizationState,
  OrganizationDetailState,
  OrganizationActions,
  OrganizationContextValue,
  OrganizationProviderProps,
} from '../../../lib/portal';

export type {
  Organization,
  OrganizationMetadata,
  OrganizationRole,
  OrganizationMember,
  MemberMetadata,
  OrganizationInvite,
  InviteStatus,
  CreateOrganizationRequest,
  InviteMemberRequest,
  OrganizationBillingPeriod,
  OrganizationBillingSummary,
} from '../../../lib/portal';
export {
  ROLE_PERMISSIONS,
  hasPermission,
  ROLE_LABELS,
  ROLE_DESCRIPTIONS,
} from '../../../lib/portal';

export {
  OrganizationList,
  OrganizationCard,
  OrganizationDetail,
  MemberList,
  InviteMemberDialog,
  CreateOrganizationDialog,
  OrganizationSwitcher,
  OrganizationBilling,
} from '../../../lib/portal';
export type {
  OrganizationListProps,
  OrganizationCardProps,
  OrganizationDetailProps,
  MemberListProps,
  InviteMemberDialogProps,
  CreateOrganizationDialogProps,
  OrganizationSwitcherProps,
  OrganizationBillingProps,
} from '../../../lib/portal';
