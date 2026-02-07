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

declare module '@virtengine/portal/types/billing' {
  export type InvoiceStatus = string;
  export type PaymentStatus = string;
  export type UsageGranularity = string;
  export type Invoice = any;
  export type CostProjection = any;
  export type UsageHistoryPoint = any;
  export type UsageSummary = any;
  export type ResourceUsage = any;
  export type DeploymentUsage = any;
  export type ProviderUsage = any;
}

declare module '@virtengine/portal/hooks/useBilling' {
  export const useInvoices: any;
  export const useInvoice: any;
  export const useCurrentUsage: any;
  export const useCostProjection: any;
  export const useUsageHistory: any;
}

declare module '@virtengine/portal/utils/billing' {
  export const calculateOutstanding: any;
  export const hasOverdueInvoices: any;
  export const formatBillingAmount: any;
  export const formatBillingPeriod: any;
  export const formatBillingDate: any;
  export const generateInvoiceText: any;
}

declare module '@virtengine/portal/utils/csv' {
  export const generateInvoicesCSV: any;
  export const generateUsageReportCSV: any;
  export const downloadFile: any;
}

declare module '@virtengine/portal/types/metrics' {
  export type Alert = any;
  export type AlertEvent = any;
  export type AlertMetric = string;
  export type AlertCondition = string;
  export type MetricPoint = any;
  export type MetricSeries = any;
  export type ResourceMetric = any;
  export type NetworkMetric = any;
  export type GpuMetric = any;
  export type DeploymentSnapshot = any;
  export type ServiceMetrics = any;
  export type DeploymentHistory = any;
  export type DeploymentMetrics = any;
  export type ProviderMetrics = any;
  export type MetricTrend = any;
  export type MetricsSummary = any;
  export type TimeRange = string;
  export type Granularity = string;
  export type WidgetType = string;
  export type DashboardWidget = any;
  export type DashboardConfig = any;
  export type WidgetConfig = any;
  export type WidgetPosition = any;

  export const ALERT_STATUS_VARIANT: Record<
    string,
    'destructive' | 'success' | 'secondary' | 'default' | 'warning' | 'info' | 'outline'
  >;
  export const TIME_RANGE_LABELS: Record<string, string>;
  export const granularityForRange: (range: string) => string;
  export const formatTimestamp: (timestamp: Date | number, granularity?: string) => string;
}

declare module '@virtengine/portal/*' {
  const portalModule: Record<string, unknown>;
  export default portalModule;
}
