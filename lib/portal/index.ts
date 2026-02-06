/**
 * VirtEngine Portal Library
 * VE-700 to VE-705: Portal foundation, VEID, MFA, Marketplace, Provider, and HPC UI
 *
 * This library provides the complete portal SDK for VirtEngine including:
 *
 * - Wallet-based authentication and session management (VE-700)
 * - VEID identity onboarding and score display (VE-701)
 * - MFA enrollment and policy configuration (VE-702)
 * - Marketplace discovery and checkout (VE-703)
 * - Provider console operations (VE-704)
 * - HPC/Supercomputer job submission (VE-705)
 *
 * @example
 * ```tsx
 * import { PortalProvider, useAuth, useIdentity, useMarketplace } from '@virtengine/portal';
 *
 * function App() {
 *   return (
 *     <PortalProvider config={{ chainEndpoint: 'wss://rpc.virtengine.com' }}>
 *       <Dashboard />
 *     </PortalProvider>
 *   );
 * }
 * ```
 *
 * @packageDocumentation
 */

// ============================================================================
// Core Exports
// ============================================================================

export { PortalProvider, usePortal } from "./components/PortalProvider";
export type { PortalProviderProps, PortalConfig } from "./types/config";

// ============================================================================
// Authentication (VE-700)
// ============================================================================

export { useAuth, AuthProvider } from "./hooks/useAuth";
export type {
  AuthState,
  AuthActions,
  WalletCredentials,
  SSOCredentials,
  SessionToken,
  AuthError,
} from "./types/auth";

export { SessionManager } from "./utils/session";
export type { SessionConfig, SessionInfo } from "./utils/session";

export { WalletAdapter, MnemonicWallet, KeypairWallet } from "./utils/wallet";
export type { WalletConfig, SigningResult } from "./types/wallet";

// ============================================================================
// Wallet Connections (Portal UI)
// ============================================================================

export { WalletProvider, useWallet } from "./src/wallet";
export type {
  WalletType as PortalWalletType,
  WalletConnectionStatus,
  WalletChainInfo,
  WalletAccount,
  WalletError as WalletErrorType,
  WalletState,
  WalletSignOptions,
  AminoSignDoc,
  AminoSignResponse,
  DirectSignDoc,
  DirectSignResponse,
  WalletContextValue,
  WalletProviderConfig,
} from "./src/wallet";

// ============================================================================
// Wallet Utilities (Enhanced - VE-700)
// ============================================================================

export {
  // Error handling
  WalletError as WalletErrorClass,
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
} from "./src/wallet";

export type {
  WalletSession,
  SessionConfig as WalletSessionConfig,
  WalletDetectionResult,
  GasTier,
  GasSettings,
  FeeEstimate,
  TransactionPreview,
  TransactionOptions,
  TransactionValidationResult,
} from "./src/wallet";

// ============================================================================
// Identity / VEID (VE-701)
// ============================================================================

export { useIdentity, IdentityProvider } from "./hooks/useIdentity";
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
} from "./types/identity";

export { getTierFromScore } from "./types/identity";
export { IdentityStatusCard } from "./components/identity/IdentityStatusCard";
export { IdentityScoreDisplay } from "./components/identity/IdentityScoreDisplay";
export { ScopeRequirements } from "./components/identity/ScopeRequirements";
export { UploadHistory } from "./components/identity/UploadHistory";
export { RemediationGuide } from "./components/identity/RemediationGuide";

// ============================================================================
// MFA (VE-702)
// ============================================================================

export { useMFA, MFAProvider } from "./hooks/useMFA";
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
} from "./types/mfa";

export { MFAEnrollmentWizard } from "./components/mfa/MFAEnrollmentWizard";
export { MFAPolicyConfig } from "./components/mfa/MFAPolicyConfig";
export { TrustedBrowserManager } from "./components/mfa/TrustedBrowserManager";
export { MFAPrompt } from "./components/mfa/MFAPrompt";
export { MFAAuditLog } from "./components/mfa/MFAAuditLog";

// ============================================================================
// Marketplace (VE-703)
// ============================================================================

export { useMarketplace, MarketplaceProvider } from "./hooks/useMarketplace";
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
} from "./types/marketplace";

export { OfferingList } from "./components/marketplace/OfferingList";
export { OfferingCard } from "./components/marketplace/OfferingCard";
export { OfferingDetail } from "./components/marketplace/OfferingDetail";
export { CheckoutFlow } from "./components/marketplace/CheckoutFlow";
export { OrderDetail } from "./components/marketplace/OrderDetail";
export { OrderTimeline } from "./components/marketplace/OrderTimeline";

// Marketplace Pages (Customer Browse Experience)
export {
  MarketplacePage,
  useOfferings,
  OFFERING_CATEGORIES,
  REGIONS,
  SearchBar,
  FilterPanel,
  CategoryNav,
  OfferingGrid,
  OfferingDetailPage,
  ProviderInfo,
  ProviderBadge,
  MarketplaceOfferingCard,
} from "./src/pages/marketplace";
export type {
  MarketplacePageProps,
  UseOfferingsOptions,
  OfferingsState,
  OfferingsActions,
  OfferingCategory,
  Region,
  SearchBarProps,
  FilterPanelProps,
  CategoryNavProps,
  OfferingGridProps,
  OfferingDetailPageProps,
  ProviderInfoProps,
  ProviderBadgeProps,
  MarketplaceOfferingCardProps,
} from "./src/pages/marketplace";

// ============================================================================
// Landing Page (Portal Entry)
// ============================================================================

export { LandingPage } from "./src/pages";
export type { LandingPageProps } from "./src/pages";

export { HeroSection } from "./src/components/hero";
export type { HeroSectionProps, HeroCTA } from "./src/components/hero";

export { StatsSection } from "./src/components/stats";
export type { StatsSectionProps } from "./src/components/stats";

export { FeaturedOfferings } from "./src/components/offerings";
export type { FeaturedOfferingsProps } from "./src/components/offerings";

export { LandingFooter } from "./src/components/footer";
export type {
  LandingFooterProps,
  FooterLinkGroup,
} from "./src/components/footer";
// Order Tracking (VE-707)
// ============================================================================

export {
  OrderTrackingProvider,
  useOrderTracking,
} from "./src/hooks/useOrderTracking";
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
} from "./src/hooks/useOrderTracking";

export {
  OrderList,
  OrderStatus,
  ResourceAccess,
  UsageMonitor,
} from "./src/components/orders";
export type {
  OrderListProps,
  OrderListFilter,
  OrderStatusProps,
  ResourceAccessProps,
  UsageMonitorProps,
} from "./src/components/orders";

export { OrderTrackingPage } from "./src/pages/orders";
export type { OrderTrackingPageProps } from "./src/pages/orders";

// ============================================================================
// Provider Console (VE-704)
// ============================================================================

export { useProvider, ProviderProvider } from "./hooks/useProvider";
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
} from "./types/provider";

export { ProviderRegistrationFlow } from "./components/provider/ProviderRegistrationFlow";
export { OfferingEditor } from "./components/provider/OfferingEditor";
export { PricingEditor } from "./components/provider/PricingEditor";
export { CapacityMonitor } from "./components/provider/CapacityMonitor";
export { BidDashboard } from "./components/provider/BidDashboard";
export { AllocationList } from "./components/provider/AllocationList";
export { UsageReports } from "./components/provider/UsageReports";
export { SettlementView } from "./components/provider/SettlementView";
export { DomainVerificationPanel } from "./components/provider/DomainVerificationPanel";

// ============================================================================
// Provider API (VE-29D/29E)
// ============================================================================

export {
  ProviderAPIClient,
  ProviderAPIError,
  LogStream,
  ShellConnection,
} from "./src/provider-api";
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
} from "./src/provider-api";

export { signRequest } from "./src/auth/wallet-sign";
export type {
  SignedRequestHeaders,
  SignRequestOptions,
} from "./src/auth/wallet-sign";

// ============================================================================
// Multi-Provider Aggregation (VE-29G)
// ============================================================================

export {
  MultiProviderClient,
  MultiProviderProvider,
  useMultiProvider,
} from "./src/multi-provider";
export type {
  ProviderRecord,
  ProviderStatus,
  DeploymentWithProvider,
  AggregatedMetrics,
  MultiProviderWallet,
  MultiProviderClientOptions,
  MultiProviderProviderProps,
} from "./src/multi-provider";

export { useAggregatedDeployments } from "./src/hooks/useAggregatedDeployments";
export type {
  AggregatedDeploymentsState,
  AggregatedDeploymentsActions,
  UseAggregatedDeploymentsOptions,
} from "./src/hooks/useAggregatedDeployments";

export { useAggregatedMetrics } from "./src/hooks/useAggregatedMetrics";
export type {
  AggregatedMetricsState,
  AggregatedMetricsActions,
  UseAggregatedMetricsOptions,
} from "./src/hooks/useAggregatedMetrics";

export { useDeploymentWithProvider } from "./src/hooks/useDeploymentWithProvider";
export type { DeploymentWithProviderState } from "./src/hooks/useDeploymentWithProvider";

// ============================================================================
// HPC / Supercomputer (VE-705)
// ============================================================================

export { useHPC, HPCProvider } from "./hooks/useHPC";
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
} from "./types/hpc";

export { WorkloadLibrary } from "./components/hpc/WorkloadLibrary";
export { JobSubmissionForm } from "./components/hpc/JobSubmissionForm";
export { JobTracker } from "./components/hpc/JobTracker";
export { JobOutputViewer } from "./components/hpc/JobOutputViewer";
export { JobCancelDialog } from "./components/hpc/JobCancelDialog";

// ============================================================================
// Wallet UI Components
// ============================================================================

export {
  WalletButton,
  WalletAccountDisplay,
  WalletNetworkBadge,
  WalletModal,
  WalletIcon,
  WalletSkeleton,
  AccountSelector,
  TransactionModal,
  KEPLR_ICON_SVG,
  LEAP_ICON_SVG,
  COSMOSTATION_ICON_SVG,
  WALLETCONNECT_ICON_SVG,
} from "./src/components/wallet";
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
} from "./src/components/wallet";

// ============================================================================
// Chain Integration
// ============================================================================

export { useChain, ChainProvider } from "./hooks/useChain";
export type {
  ChainState,
  ChainConfig,
  EventSubscription,
  QueryClient,
  TransactionResult,
} from "./types/chain";

// TODO: Implement chain utilities
// export { ChainEventListener } from './utils/chain-events';
// export { RPCClient } from './utils/rpc-client';
// export { WebSocketClient } from './utils/websocket-client';

// ============================================================================
// Utilities
// ============================================================================

export {
  formatScore,
  formatTokenAmount,
  formatDuration,
  formatTimestamp,
} from "./utils/format";
export {
  validateAddress,
  validateMnemonic,
  isValidScore,
} from "./utils/validation";
export {
  sanitizePlainText,
  sanitizeDigits,
  sanitizeJsonInput,
  sanitizeObject,
} from "./utils/security";
export { encryptPayload, decryptPayload } from "./utils/encryption";
export type { EncryptionResult, DecryptionResult } from "./utils/encryption";
export {
  createOAuthRequest,
  persistOAuthRequest,
  consumeOAuthRequest,
  buildAuthorizationUrl,
  createPKCE,
} from "./utils/oidc";
export type { OAuthRequest } from "./utils/oidc";

// ============================================================================
// Accessibility (VE-UI-002)
// ============================================================================

export {
  // Accessible components
  SrOnly,
  SkipLink,
  AccessibleButton,
  AccessibleInput,
  AccessibleSelect,
  AccessibleCheckbox,
  AccessibleAlert,
  AccessibleProgress,
} from "./components/accessible";

export type {
  SrOnlyProps,
  SkipLinkProps,
  AccessibleButtonProps,
  AccessibleInputProps,
  AccessibleSelectProps,
  AccessibleCheckboxProps,
  AccessibleAlertProps,
  AccessibleProgressProps,
} from "./components/accessible";

export {
  // Accessibility utilities
  generateA11yId,
  announce,
  clearAnnouncements,
  initLiveRegions,
  createFocusTrap,
  getFocusableElements,
  meetsContrastRequirement,
  getContrastRatio,
  hexToRgb,
  getLuminance,
  handleArrowNavigation,
  manageRovingTabindex,
  prefersReducedMotion,
  prefersHighContrast,
  srOnlyStyles,
  focusVisibleStyles,
  A11Y_COLORS,
} from "./utils/a11y";

export type { FocusTrap, ArrowNavOptions } from "./utils/a11y";

export {
  // Accessibility testing utilities
  runA11yTests,
  expectNoA11yViolations,
  checkContrastRatio,
  checkFocusIndicator,
  checkTouchTargetSize,
  analyzeKeyboardNav,
  validateAriaAttributes,
  checkScreenReaderContent,
  generateA11yReport,
  formatViolations,
  toHaveNoViolations,
  WCAG_21_AA_CONFIG,
} from "./utils/a11y-testing";

export type {
  A11yTestConfig,
  A11yReport,
  KeyboardNavTestResult,
} from "./utils/a11y-testing";

// ============================================================================
// Organization Management (VE-29H)
// ============================================================================

export { useOrganization, OrganizationProvider } from "./hooks/useOrganization";
export type {
  OrganizationState,
  OrganizationDetailState,
  OrganizationActions,
  OrganizationContextValue,
  OrganizationProviderProps,
} from "./hooks/useOrganization";

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
} from "./types/organization";
export {
  ROLE_PERMISSIONS,
  hasPermission,
  ROLE_LABELS,
  ROLE_DESCRIPTIONS,
} from "./types/organization";

export { OrganizationList } from "./components/organization/OrganizationList";
export type { OrganizationListProps } from "./components/organization/OrganizationList";
export { OrganizationCard } from "./components/organization/OrganizationCard";
export type { OrganizationCardProps } from "./components/organization/OrganizationCard";
export { OrganizationDetail } from "./components/organization/OrganizationDetail";
export type { OrganizationDetailProps } from "./components/organization/OrganizationDetail";
export { MemberList } from "./components/organization/MemberList";
export type { MemberListProps } from "./components/organization/MemberList";
export { InviteMemberDialog } from "./components/organization/InviteMemberDialog";
export type { InviteMemberDialogProps } from "./components/organization/InviteMemberDialog";
export { CreateOrganizationDialog } from "./components/organization/CreateOrganizationDialog";
export type { CreateOrganizationDialogProps } from "./components/organization/CreateOrganizationDialog";
export { OrganizationSwitcher } from "./components/organization/OrganizationSwitcher";
export type { OrganizationSwitcherProps } from "./components/organization/OrganizationSwitcher";
export { OrganizationBilling } from "./components/organization/OrganizationBilling";
export type { OrganizationBillingProps } from "./components/organization/OrganizationBilling";
