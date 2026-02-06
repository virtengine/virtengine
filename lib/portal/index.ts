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
  VerificationScope,
  UploadRecord,
  VerificationRecord,
  IdentityGatingError,
  MarketplaceAction,
  RemediationPath,
  ScopeRequirement,
} from "./types/identity";

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
// HPC / Supercomputer (VE-705)
// ============================================================================

export { useHPC, HPCProvider } from "./hooks/useHPC";
export type {
  HPCState,
  WorkloadTemplate,
  JobManifest,
  JobSubmission,
  JobStatus,
  JobEvent,
  JobOutput,
  JobOutputReference,
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
