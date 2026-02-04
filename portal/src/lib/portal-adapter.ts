/**
 * Portal Adapter
 * Re-exports from lib/portal for Next.js integration
 *
 * This adapter provides a clean import path for portal library components,
 * hooks, and utilities within the Next.js application.
 */

// Re-export from the workspace package
// Note: These imports rely on the pnpm workspace setup with virtengine-portal-lib

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
// Identity / VEID (VE-701)
// ============================================================================

export { useIdentity, IdentityProvider } from '../../../lib/portal';
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
} from '../../../lib/portal';

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
  MFAPolicy,
  MFAEnrollment,
  TrustedBrowser,
  MFAChallenge,
  MFAChallengeResponse,
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
// HPC / Supercomputer (VE-705)
// ============================================================================

export { useHPC, HPCProvider } from '../../../lib/portal';
export type {
  HPCState,
  WorkloadTemplate,
  JobManifest,
  JobSubmission,
  JobStatus,
  JobEvent,
  JobOutput,
  JobOutputReference,
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

export { formatScore, formatTokenAmount, formatDuration, formatTimestamp } from '../../../lib/portal';
export { validateAddress, validateMnemonic, isValidScore } from '../../../lib/portal';
export { sanitizePlainText, sanitizeDigits, sanitizeJsonInput, sanitizeObject } from '../../../lib/portal';
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
