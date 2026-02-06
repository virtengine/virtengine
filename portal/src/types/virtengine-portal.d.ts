/* eslint-disable @typescript-eslint/no-explicit-any */
declare module '@virtengine/portal' {
  export const PortalProvider: any;
  export const WalletProvider: any;
  export const useWallet: any;
  export const useMarketplace: any;
  export const OfferingList: any;
  export const OfferingDetail: any;
  export const OrderDetail: any;
  export const OrderTimeline: any;
  export const CheckoutFlow: any;
  export const useIdentity: any;
  export const IdentityStatusCard: any;
  export const IdentityScoreDisplay: any;
  export const ScopeRequirements: any;
  export const RemediationGuide: any;
  export const useHPC: any;
  export const JobTracker: any;
  export const JobOutputViewer: any;
  export const JobCancelDialog: any;
  export const JobSubmissionForm: any;
  export const WorkloadLibrary: any;

  export type WalletProviderConfig = any;
  export type WalletChainInfo = any;
  export type PortalConfig = any;
  export type ChainConfig = any;
  export type PortalWalletType = any;
  export type WalletError = any;
  export type Offering = any;
  export type MarketplaceAction = any;
  export type MFAFactorType =
    | 'otp'
    | 'fido2'
    | 'sms'
    | 'email'
    | 'biometric'
    | (string & Record<string, never>);
  export type MFAFactorStatus =
    | 'active'
    | 'inactive'
    | 'pending'
    | 'revoked'
    | (string & Record<string, never>);
  export type MFAFactor = {
    id: string;
    status: MFAFactorStatus;
    type: MFAFactorType;
    name?: string | null;
    isPrimary?: boolean;
    enrolledAt?: number | null;
    lastUsedAt?: number | null;
  };
  export type SensitiveTransactionType = string;
  export type MFAState = any;
  export type MFAPolicy = any;
  export type MFAEnrollment = any;
  export type MFAEnrollmentStep = any;
  export type MFAEnrollmentChallengeData = any;
  export type MFAChallenge = any;
  export type MFAChallengeResponse = any;
  export type MFAChallengeType = any;
  export type MFAAuditEntry = any;
  export type MFAError = any;
  export type MFAErrorCode = any;
  export type TrustedBrowser = any;
  export type IdentityTier = any;
  export type IdentityStatus = any;
  export type VerificationScopeType = any;
  export type IdentityScore = any;
}

declare module '@virtengine/portal/*' {
  const portalModule: Record<string, unknown>;
  export default portalModule;
}
