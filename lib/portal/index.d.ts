/* eslint-disable @typescript-eslint/no-explicit-any */

import type * as React from "react";

export type UnsupportedDocumentReasonCode =
  | "CATEGORY_NOT_SUPPORTED"
  | "COUNTRY_NOT_SUPPORTED"
  | "DOCUMENT_EXPIRED"
  | "DOCUMENT_FORMAT_NOT_SUPPORTED"
  | "DOCUMENT_QUALITY_INSUFFICIENT"
  | "LANGUAGE_NOT_SUPPORTED"
  | "REQUIRED_SIDE_MISSING";
export type UnsupportedDocumentCategory =
  | "DRIVER_LICENSE"
  | "NATIONAL_ID"
  | "OTHER_GOVERNMENT_DOCUMENT"
  | "PASSPORT"
  | "RESIDENCE_PERMIT";
export type SupportLanguageCode =
  | "ar"
  | "de"
  | "en"
  | "es"
  | "fr"
  | "hi"
  | "id"
  | "ja"
  | "ko"
  | "pt"
  | "zh";
export type RedactedReasonToken =
  | "category-mismatch"
  | "country-restricted"
  | "expired-document"
  | "format-unsupported"
  | "language-assistance"
  | "quality-insufficient"
  | "review-required"
  | "side-missing";
export type SafeNextStepOption =
  | "CONTACT_SUPPORT"
  | "REQUEST_ACCESSIBLE_ASSISTANCE"
  | "RETRY_WITH_SUPPORTED_DOCUMENT"
  | "REVIEW_SUPPORTED_DOCUMENTS";
export interface UnsupportedDocumentTriageRequest {
  schemaVersion: "unsupported-document-triage.v1";
  reasonCodes: UnsupportedDocumentReasonCode[];
  supportLanguageCode: SupportLanguageCode;
  redactedNotes: RedactedReasonToken[];
  documentCategory?: UnsupportedDocumentCategory;
  countryCode?: string;
}
export interface UnsupportedDocumentTriageOutput {
  schemaVersion: "unsupported-document-triage.v1";
  unsupportedDocumentCategory: UnsupportedDocumentCategory;
  safeNextStepOptions: SafeNextStepOption[];
  redactedCaseSummary: string;
}
export interface UnsupportedDocumentTriageDraft extends UnsupportedDocumentTriageOutput {
  status: "draft";
  draftReference: string;
}
export interface HumanTriageConfirmation {
  confirmed: true;
  draftReference: string;
}
export interface AcceptedUnsupportedDocumentSupportNote extends UnsupportedDocumentTriageOutput {
  status: "accepted-support-note";
  draftReference: string;
  humanConfirmed: true;
}
export interface UnsupportedDocumentModelAuthority {
  proposeUnsupportedDocumentTriage(
    request: Readonly<UnsupportedDocumentTriageRequest>,
  ): Promise<unknown>;
}
export interface UnsupportedDocumentOutputValidator {
  validate(
    output: Readonly<UnsupportedDocumentTriageOutput>,
  ): void | Promise<void>;
}
export interface TriageDigestAuthority {
  sha256(canonicalValue: string): string | Promise<string>;
}
export interface UnsupportedDocumentTriagePolicy {
  allowDocumentCategory?: boolean;
  allowCountryCode?: boolean;
}
export interface UnsupportedDocumentTriageDependencies {
  modelAuthority?: UnsupportedDocumentModelAuthority;
  outputValidator?: UnsupportedDocumentOutputValidator;
  digestAuthority?: TriageDigestAuthority;
  policy?: UnsupportedDocumentTriagePolicy;
}

export const AccessibleAlert: any;
export const AccessibleButton: any;
export const AccessibleCheckbox: any;
export const AccessibleInput: any;
export const AccessibleProgress: any;
export const AccessibleSelect: any;
export const AccountSelector: any;
export const adjustGas: any;
export const AllocationList: any;
export const announce: any;
export const AuthProvider: any;
export const BidDashboard: any;
export const buildChatContext: any;
export const calculateFee: any;
export const CapacityMonitor: any;
export const ChainProvider: any;
export const ChatAgent: any;
export class UnsupportedDocumentTriageAssistant {
  constructor(dependencies?: UnsupportedDocumentTriageDependencies);
  isAvailable(): boolean;
  propose(value: unknown): Promise<UnsupportedDocumentTriageDraft>;
  confirm(
    originalRequest: unknown,
    draftValue: unknown,
    confirmationValue: unknown,
  ): Promise<AcceptedUnsupportedDocumentSupportNote>;
}
export class UnsupportedDocumentTriageConfirmationError extends Error {}
export class UnsupportedDocumentTriageUnavailableError extends Error {}
export class UnsupportedDocumentTriageValidationError extends Error {
  readonly path: string;
}
export const UNSUPPORTED_DOCUMENT_TRIAGE_SCHEMA_VERSION: "unsupported-document-triage.v1";
export const UNSUPPORTED_DOCUMENT_REASON_CODES: readonly UnsupportedDocumentReasonCode[];
export const UNSUPPORTED_DOCUMENT_CATEGORIES: readonly UnsupportedDocumentCategory[];
export const SUPPORT_LANGUAGE_CODES: readonly SupportLanguageCode[];
export const REDACTED_REASON_TOKENS: readonly RedactedReasonToken[];
export const REDACTED_SUMMARY_TOKENS: readonly string[];
export const SAFE_NEXT_STEP_OPTIONS: readonly SafeNextStepOption[];
export function validateUnsupportedDocumentTriageOutput(
  value: unknown,
): UnsupportedDocumentTriageOutput;
export function validateUnsupportedDocumentTriageRequest(
  value: unknown,
  policy?: UnsupportedDocumentTriagePolicy,
): UnsupportedDocumentTriageRequest;
export interface CheckoutFlowProps {
  offering: Offering;
  onComplete: (orderId: string) => void;
  onCancel?: () => void;
  className?: string;
  mutationAdapter?: CheckoutMutationAdapter;
  mutationContext?: CheckoutMutationContext;
  resultProjector?: CheckoutMutationProjector;
  mutationTimeoutMs?: number;
}
export const CheckoutFlow: React.ComponentType<CheckoutFlowProps>;
export interface CheckoutMutationContext {
  chainId: string;
  customerAddress: string;
}
export interface CheckoutMutationRequest {
  chainId: string;
  customerAddress: string;
  offeringId: string;
  providerAddress: string;
  durationSeconds: number;
  priceAmount: string;
  depositAmount: string;
  priceDenom: string;
}
export interface CheckoutMutationSubmission {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}
export interface CheckoutMutationAdapter {
  submitOrder(
    request: CheckoutMutationRequest,
    submission: CheckoutMutationSubmission,
  ): Promise<unknown>;
}
export type CheckoutMutationProjector = (result: unknown) => unknown;
export interface CheckoutCommittedResult {
  status: "committed";
  code: 0;
  orderId: string;
  txHash: string;
  blockHeight: number;
  requestDigest: string;
  idempotencyKey: string;
  request: CheckoutMutationRequest;
}
export const clearAnnouncements: any;
export const createChainConfig: any;
export const createChainQueryClient: any;
export const createChainSigningClient: any;
export const createChatAgent: any;
export const createChatProvider: any;
export const createChatSnapshot: any;
export const createChatSystemMessage: any;
export const createDefaultChatTools: any;
export const createDefaultGasSettings: any;
export const createFocusTrap: any;
export const CreateOrganizationDialog: any;
export const createSessionManager: any;
export const createTransactionPreview: any;
export const createWalletError: any;
export const createWalletSigner: any;
export const decryptPayload: any;
export const DEFAULT_GAS_ADJUSTMENT: any;
export const DEFAULT_GAS_LIMIT: any;
export const DomainVerificationPanel: any;
export const encryptPayload: any;
export const estimateGas: any;
export const fetchMarketOffering: any;
export const fetchMarketOfferings: any;
export const fetchProviderRegistration: any;
export const focusVisibleStyles: any;
export const formatDuration: any;
export const formatFeeAmount: any;
export const formatScore: any;
export const formatTimestamp: any;
export const formatTokenAmount: any;
export const generateA11yId: any;
export const getContrastRatio: any;
export const getErrorMessage: any;
export const getFocusableElements: any;
export const getSuggestedAction: any;
export const getTierFromScore: any;
export const hasPermission: any;
export interface HPCProviderProps {
  children: React.ReactNode;
  queryClient: any;
  chainId: string;
  accountAddress: string | null;
  getAuthHeader?: () => Promise<string>;
  mutationAdapter?: HPCSignerAdapter;
  outputAdapter?: HPCOutputAdapter;
  queryAdapter?: HPCQueryAdapter;
}
export const HPCProvider: React.ComponentType<HPCProviderProps>;
export type HPCClientCapability = "query" | "signer" | "provider";
export interface SubmitJobParams {
  offeringId: string;
  name: string;
  description?: string;
  templateId?: string;
  resources: {
    nodes: number;
    cpusPerNode: number;
    memoryGBPerNode: number;
    gpusPerNode?: number;
    gpuType?: string;
    maxRuntimeSeconds: number;
    storageGB: number;
  };
  command?: string;
  containerImage?: string;
  environment?: Record<string, string>;
  parameters?: Record<string, string | number | boolean>;
  encryptedInputs?: Record<string, unknown>;
  inputRefs?: string[];
  quote?: {
    estimatedTotal: string;
    depositRequired: string;
    pricePerHour: string;
    maxHours: number;
    denom: string;
  };
}
export function assertValidSubmitJobParams(
  params: SubmitJobParams,
  requireQuote?: boolean,
): void;
export interface CommittedJobMutation {
  committed: true;
  jobId: string;
  txHash: string;
  code: 0;
  blockHeight: number;
}
export interface HPCSignerAdapter {
  readonly state: "query-only" | "signing-ready";
  readonly chainId: string;
  readonly accountAddress: string;
  submitJob(params: SubmitJobParams): Promise<unknown>;
  cancelJob(jobId: string): Promise<unknown>;
}
export class HPCClientUnavailableError extends Error {
  readonly code: "hpc_client_unavailable";
  readonly capability: HPCClientCapability;
  constructor(capability: HPCClientCapability);
}
export class HPCMutationNotCommittedError extends Error {
  readonly code: "hpc_mutation_not_committed";
  constructor();
}
export interface HPCOutputAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  getOutputs(jobId: string): Promise<unknown>;
  resolveOutput(outputRef: JobOutputReference): Promise<unknown>;
}
export interface HPCQueryAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  getWorkloadTemplates(): Promise<unknown>;
  getQuote(request: HPCQuoteRequest): Promise<unknown>;
  getJobs(): Promise<unknown>;
  getJob(jobId: string): Promise<unknown>;
  subscribeToJob?(
    jobId: string,
    callback: (event: unknown) => void,
  ): () => void;
}
export interface HPCQuoteRequest {
  offeringId: string;
  resources: JobResources;
}
export interface HPCQueryEnvelope {
  chainId: string;
  accountAddress: string;
}
export class HPCQueryValidationError extends Error {
  readonly code: "hpc_query_invalid";
  constructor();
}
export function requireHPCQueryAdapter(
  adapter: HPCQueryAdapter | undefined,
  expected: HPCQueryEnvelope,
): HPCQueryAdapter;
export function validateHPCWorkloadTemplates(
  value: unknown,
  expected: HPCQueryEnvelope,
): readonly WorkloadTemplate[];
export function validateHPCJobPriceQuote(
  value: unknown,
  expected: HPCQueryEnvelope,
  expectedRequest: HPCQuoteRequest,
): JobPriceQuote;
export function validateHPCQuoteRequest(
  value: HPCQuoteRequest,
): HPCQuoteRequest;
export function validateHPCJobs(
  value: unknown,
  expected: HPCQueryEnvelope,
): readonly Job[];
export function validateHPCJob(
  value: unknown,
  expected: HPCQueryEnvelope & { jobId: string },
): Job;
export class HPCOutputValidationError extends Error {
  readonly code: "hpc_output_invalid";
  constructor();
}
export function requireHPCOutputAdapter(
  adapter: HPCOutputAdapter | undefined,
  expected: { chainId: string; accountAddress: string },
): HPCOutputAdapter;
export function validateHPCOutputReferences(
  value: unknown,
  expected: { chainId: string; accountAddress: string; jobId: string },
): readonly JobOutputReference[];
export function validateResolvedHPCOutput(
  value: unknown,
  expected: JobOutputReference,
  binding: { chainId: string; accountAddress: string; jobId: string },
  now?: number,
): JobOutput;
export function assertCommittedJobMutation(
  result: unknown,
  expectedJobId?: string,
): asserts result is CommittedJobMutation;
export function requireHPCSigner(
  adapter: HPCSignerAdapter | undefined,
  expected?: { chainId: string; accountAddress: string },
): HPCSignerAdapter;
export const IdentityProvider: any;
export const IdentityScoreDisplay: any;
export const IdentityStatusCard: any;
export type UniquenessEnrollmentStatusValue =
  | "processing"
  | "possible-match-review"
  | "unique"
  | "duplicate-confirmed"
  | "unavailable"
  | "appeal";
export interface UniquenessStatusProjection {
  status: UniquenessEnrollmentStatusValue;
  receiptId: string;
  revision: number;
  supersedesReceiptId?: string;
  governedFinalAdjudication?: boolean;
}
export type UniquenessReceiptProjector = (
  receipt: unknown,
) => UniquenessStatusProjection;
export interface UniquenessEnrollmentState {
  status: UniquenessEnrollmentStatusValue;
  receiptId: string | null;
  revision: number | null;
}
export type UniquenessTransitionErrorCode =
  | "invalid-projection"
  | "invalid-transition"
  | "stale-receipt"
  | "superseded-receipt"
  | "final-adjudication-required";
export class UniquenessTransitionError extends Error {
  readonly code: UniquenessTransitionErrorCode;
}
export interface UniquenessEnrollmentAdapter {
  getState(): Readonly<UniquenessEnrollmentState>;
  beginProcessing(): Readonly<UniquenessEnrollmentState>;
  applyReceipt(receipt: unknown): Readonly<UniquenessEnrollmentState>;
  requestAppeal(): Readonly<UniquenessEnrollmentState>;
}
export interface UniquenessEnrollmentAdapterOptions {
  projectReceipt?: UniquenessReceiptProjector;
}
export function createUniquenessEnrollmentAdapter(
  options?: UniquenessEnrollmentAdapterOptions,
): UniquenessEnrollmentAdapter;
export interface UniquenessEnrollmentStatusProps {
  state: Pick<UniquenessEnrollmentState, "status">;
  onManualVerification: () => void;
  onAppeal: () => void;
  className?: string;
}
export const UniquenessEnrollmentStatus: (
  props: UniquenessEnrollmentStatusProps,
) => any;
export const initLiveRegions: any;
export const InviteMemberDialog: any;
export const isRetryableError: any;
export const isValidScore: any;
export const isWalletError: any;
export interface JobCancelDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  jobId: string;
  jobName?: string;
  onConfirm: () => Promise<void>;
}
export const JobCancelDialog: React.ComponentType<JobCancelDialogProps>;
export const JobOutputViewer: any;
export interface JobSubmissionFormProps {
  offeringId: string;
  template?: WorkloadTemplate;
  onSubmit?: (jobId: string) => void;
  onCancel?: () => void;
  className?: string;
}
export const JobSubmissionForm: React.ComponentType<JobSubmissionFormProps>;
export const JobTracker: any;
export const KeypairWallet: any;
export const LogStream: any;
export const MarketplaceProvider: any;
export const meetsContrastRequirement: any;
export const MemberList: any;
export const MFAAuditLog: any;
export const MFAEnrollmentWizard: any;
export const MFAPolicyConfig: any;
export const MFAPrompt: any;
export const MFAProvider: any;
export const MnemonicWallet: any;
export const MultiProviderClient: any;
export const MultiProviderProvider: any;
export const normalizeMarketPagination: any;
export const OfferingCard: any;
export const OfferingDetail: any;
export const OfferingEditor: any;
export const OfferingList: any;
export const OrderDetail: any;
export const OrderList: any;
export const OrderStatus: any;
export const OrderTimeline: any;
export const OrderTrackingPage: any;
export const OrderTrackingProvider: any;
export const OrganizationBilling: any;
export const OrganizationCard: any;
export const OrganizationDetail: any;
export const OrganizationList: any;
export const OrganizationProvider: React.ComponentType<OrganizationProviderProps>;
export const OrganizationSwitcher: any;
export const parseWalletError: any;
export interface PortalProviderProps {
  config: PortalConfig;
  chainConfig: ChainConfig;
  walletConfig?: WalletProviderConfig;
  marketplaceMutationAdapter?: CheckoutMutationAdapter;
  marketplaceResultProjector?: CheckoutMutationProjector;
  marketplaceMutationTimeoutMs?: number;
  hpcMutationAdapter?: HPCSignerAdapter;
  hpcOutputAdapter?: HPCOutputAdapter;
  hpcQueryAdapter?: HPCQueryAdapter;
  providerDomainVerifier?: ProviderDomainVerifier;
  providerOfferingMutationAdapter?: ProviderOfferingMutationAdapter;
  organizationMutationAdapter?: OrganizationMutationAdapter;
  children: React.ReactNode;
}
export const PortalProvider: React.ComponentType<PortalProviderProps>;
export const prefersHighContrast: any;
export const prefersReducedMotion: any;
export const PricingEditor: any;
export const ProviderAPIClient: any;
export const ProviderAPIError: any;
export class ProviderShellSessionError extends Error {
  readonly code: ProviderShellSessionErrorCode;
  readonly cause?: unknown;
  constructor(
    code: ProviderShellSessionErrorCode,
    message: string,
    cause?: unknown,
  );
}
export const buildProviderShellWebSocketUrl: any;
export const validateProviderShellSessionReceipt: any;
export const providerDeploymentActions: readonly [
  "start",
  "stop",
  "restart",
  "update",
  "terminate",
];
export class ProviderDeploymentActionError extends Error {
  readonly code: ProviderDeploymentActionErrorCode;
  readonly cause?: unknown;
  constructor(
    code: ProviderDeploymentActionErrorCode,
    message: string,
    cause?: unknown,
  );
}
export const validateProviderDeploymentActionReceipt: ProviderDeploymentActionReceiptValidator;
export interface ProviderDomainBinding {
  chainId: string;
  accountAddress: string;
}
export interface ProviderDomainChallenge
  extends DomainChallenge, ProviderDomainBinding {
  challengeId: string;
}
export interface ProviderDomainVerificationEvidence
  extends DomainVerification, ProviderDomainBinding {
  status: "verified";
  challengeId: string;
  evidenceId: string;
}
export interface ProviderDomainVerifier {
  readonly chainId: string;
  readonly accountAddress: string;
  issueChallenge(
    domain: string,
    method: DomainVerificationMethod,
  ): Promise<unknown>;
  verifyChallenge(challenge: ProviderDomainChallenge): Promise<unknown>;
}
export class ProviderDomainVerificationError extends Error {
  readonly code:
    | "feature_unavailable"
    | "invalid_domain"
    | "invalid_challenge"
    | "invalid_verification"
    | "authority_changed"
    | "challenge_in_progress"
    | "verification_in_progress";
  constructor(code: ProviderDomainVerificationError["code"]);
}
export interface ProviderProviderProps {
  children: React.ReactNode;
  queryClient: import("./types/chain").QueryClient;
  chainId: string;
  accountAddress: string | null;
  getAuthHeader?: () => Promise<string>;
  domainVerifier?: ProviderDomainVerifier;
  offeringMutationAdapter?: ProviderOfferingMutationAdapter;
}
export const ProviderProvider: React.ComponentType<ProviderProviderProps>;
export function normalizeProviderDomain(value: string): string;
export function requireProviderDomainVerifier(
  verifier: ProviderDomainVerifier | undefined,
  binding: ProviderDomainBinding,
): ProviderDomainVerifier;
export function validateProviderDomainChallenge(
  value: unknown,
  binding: ProviderDomainBinding,
  domain: string,
  method: DomainVerificationMethod,
  now?: number,
): ProviderDomainChallenge;
export function validateProviderDomainVerification(
  value: unknown,
  binding: ProviderDomainBinding,
  challenge: ProviderDomainChallenge,
  now?: number,
): ProviderDomainVerificationEvidence;
export type ProviderOfferingMutationAction =
  | "create"
  | "update"
  | "publish"
  | "pause";
export interface ProviderOfferingMutationRequest {
  chainId: string;
  accountAddress: string;
  action: ProviderOfferingMutationAction;
  offeringId?: string;
  draft?: OfferingDraft | Partial<OfferingDraft>;
}
export interface ProviderOfferingMutationContext {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}
export interface ProviderOfferingMutationAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  mutateOffering(
    request: ProviderOfferingMutationRequest,
    context: ProviderOfferingMutationContext,
  ): Promise<unknown>;
}
export interface CommittedProviderOfferingMutation {
  status: "committed";
  code: 0;
  txHash: string;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: ProviderOfferingMutationRequest;
  offering: ProviderOffering;
}
export class ProviderOfferingMutationError extends Error {
  readonly code:
    | "feature_unavailable"
    | "invalid_request"
    | "invalid_committed_result"
    | "request_changed"
    | "submission_cancelled"
    | "submission_in_progress";
  constructor(code: ProviderOfferingMutationError["code"]);
}
export function buildProviderOfferingMutationRequest(
  action: ProviderOfferingMutationAction,
  binding: { chainId: string; accountAddress: string },
  offeringId?: string,
  draft?: OfferingDraft | Partial<OfferingDraft>,
): ProviderOfferingMutationRequest;
export function digestProviderOfferingRequest(
  request: ProviderOfferingMutationRequest,
): Promise<string>;
export function requireProviderOfferingMutationAdapter(
  adapter: ProviderOfferingMutationAdapter | undefined,
  binding: { chainId: string; accountAddress: string },
): ProviderOfferingMutationAdapter;
export function validateCommittedProviderOfferingMutation(
  value: unknown,
  request: ProviderOfferingMutationRequest,
  requestDigest: string,
): CommittedProviderOfferingMutation;
export const ProviderRegistrationFlow: any;
export const RemediationGuide: any;
export const ResourceAccess: any;
export const ROLE_DESCRIPTIONS: any;
export const ROLE_LABELS: any;
export const ROLE_PERMISSIONS: any;
export const sanitizeDigits: any;
export const sanitizeJsonInput: any;
export const sanitizeObject: any;
export const sanitizePlainText: any;
export const ScopeRequirements: any;
export const SessionManager: any;
export const SettlementView: any;
export const ShellConnection: any;
export const signRequest: any;
export const SkipLink: any;
export const SrOnly: any;
export const srOnlyStyles: any;
export const TransactionModal: any;
export const TrustedBrowserManager: any;
export const UploadHistory: any;
export const UsageMonitor: any;
export const UsageReports: any;
export const useAggregatedDeployments: any;
export const useAggregatedMetrics: any;
export const useAuth: any;
export const useChain: any;
export const useDeploymentWithProvider: any;
export interface HPCActions {
  refresh(): Promise<void>;
  getWorkloadTemplates(): Promise<void>;
  startJobSubmission(templateId?: string): void;
  updateJobManifest(manifest: Partial<JobManifest>): void;
  selectOffering(offeringId: string): void;
  getQuote(request?: JobResources | HPCQuoteRequest): Promise<JobPriceQuote>;
  validateJob(): JobValidationError[];
  submitJob(): Promise<CommittedJobMutation>;
  cancelJob(jobId: string): Promise<CommittedJobMutation>;
  cancelSubmission(): void;
  getJobs(): Promise<void>;
  getJob(jobId: string): Promise<Job>;
  getOutputs(jobId: string): Promise<readonly JobOutputReference[]>;
  decryptOutput(
    jobId: string,
    outputRef: JobOutputReference,
  ): Promise<JobOutput>;
  subscribeToJob(
    jobId: string,
    callback: (event: import("./types/chain").ChainEvent) => void,
  ): () => void;
  clearError(): void;
}
export interface HPCContextValue {
  state: HPCState;
  actions: HPCActions;
}
export function useHPC(): HPCContextValue;
export const useIdentity: any;
export const useMarketplace: any;
export const useMFA: any;
export const useMultiProvider: any;
export const useOrderTracking: any;
export function useOrganization(): OrganizationContextValue;
export interface OrganizationState {
  isLoading: boolean;
  organizations: Organization[];
  selectedOrgId: string | null;
  error: string | null;
}
export interface OrganizationDetailState {
  isLoading: boolean;
  organization: Organization | null;
  members: OrganizationMember[];
  billing: OrganizationBillingSummary | null;
  error: string | null;
}
export interface OrganizationActions {
  fetchOrganizations(): Promise<void>;
  selectOrganization(orgId: string | null): void;
  createOrganization(request: CreateOrganizationRequest): Promise<Organization>;
  fetchOrganizationDetail(orgId: string): Promise<void>;
  inviteMember(orgId: string, request: InviteMemberRequest): Promise<void>;
  removeMember(orgId: string, memberAddress: string): Promise<void>;
  updateMemberRole(
    orgId: string,
    memberAddress: string,
    role: OrganizationRole,
  ): Promise<void>;
  leaveOrganization(orgId: string): Promise<void>;
  fetchBilling(orgId: string): Promise<void>;
}
export interface OrganizationContextValue {
  state: OrganizationState;
  detail: OrganizationDetailState;
  actions: OrganizationActions;
  selectedOrganization: Organization | null;
  currentUserRole: OrganizationRole | null;
}
export interface OrganizationProviderProps {
  children: React.ReactNode;
  queryClient?: import("./types/chain").QueryClient;
  chainId: string;
  accountAddress?: string | null;
  mutationAdapter?: OrganizationMutationAdapter;
}
export const usePortal: any;
export interface ProviderActions {
  refresh(): Promise<void>;
  startRegistration(): void;
  updateRegistrationData(data: Partial<ProviderRegistration>): void;
  startDomainVerification(
    domain: string,
    method: DomainVerificationMethod,
  ): Promise<ProviderDomainChallenge>;
  checkDomainVerification(
    domain: string,
  ): Promise<ProviderDomainVerificationEvidence>;
  submitRegistration(): Promise<void>;
  createOffering(draft: OfferingDraft): Promise<ProviderOffering>;
  updateOffering(
    offeringId: string,
    updates: Partial<OfferingDraft>,
  ): Promise<ProviderOffering>;
  publishOffering(offeringId: string): Promise<void>;
  pauseOffering(offeringId: string): Promise<void>;
  getIncomingOrders(): Promise<void>;
  getActiveBids(): Promise<void>;
  getActiveAllocations(): Promise<void>;
  getUsageRecords(allocationId?: string): Promise<void>;
  getSettlementSummary(): Promise<void>;
  clearError(): void;
}
export interface ProviderContextValue {
  state: ProviderState;
  actions: ProviderActions;
}
export function useProvider(): ProviderContextValue;
export const validateAddress: any;
export const validateMnemonic: any;
export const validateTransaction: any;
export const WALLET_ERROR_MESSAGES: any;
export const WalletAccountDisplay: any;
export const WalletAdapter: any;
export const WalletButton: any;
export const walletDetector: any;
export const WalletErrorCode: any;
export const WalletIcon: any;
export const WalletModal: any;
export const WalletNetworkBadge: any;
export const WalletPriority: any;
export const walletSessionManager: any;
export const WalletSkeleton: any;
export const withWalletTimeout: any;
export const WorkloadLibrary: any;
export const wrapWithWalletError: any;
export type AccountSelectorProps = any;
export type AggregatedDeploymentsActions = any;
export type AggregatedDeploymentsState = any;
export type AggregatedMetrics = any;
export type AggregatedMetricsActions = any;
export type AggregatedMetricsState = any;
export type AllocationRecord = any;
export type AminoSignDoc = any;
export type AminoSignResponse = any;
export type AuthActions = any;
export type AuthError = any;
export type AuthState = any;
export type BidRecord = any;
export type CapacityConfig = any;
export type ChainClientConfig = any;
export type ChainConfig = any;
export type ChainQueryClient = any;
export type ChainState = any;
export type ChatAction = any;
export type ChatActionExecution = any;
export type ChatContextSnapshot = any;
export type ChatMessage = any;
export type ChatProviderConfig = any;
export type ChatProviderType = any;
export type ChatToolContext = any;
export type ChatToolDefinition = any;
export type ChatToolHandler = any;
export type CheckoutRequest = any;
export type CheckoutValidation = any;
export type CreateOrganizationDialogProps = any;
export type CreateOrganizationRequest =
  import("./types/organization").CreateOrganizationRequest;
export type DecryptionResult = any;
export type Deployment = any;
export type DeploymentAction = any;
export type DeploymentListResponse = any;
export type DeploymentState = any;
export type DeploymentStatus = any;
export type DeploymentWithProvider = any;
export type DeploymentWithProviderState = any;
export type DirectSignDoc = any;
export type DirectSignResponse = any;
export type DomainVerification = any;
export type EncryptionResult = any;
export type EventSubscription = any;
export type FeeEstimate = any;
export type GasSettings = any;
export type GasTier = any;
export type HPCError = any;
export type HPCErrorCode = any;
export type HPCState = import("./types/hpc").HPCState;
export type IdentityGatingError = any;
export type IdentityScore = any;
export type IdentityState = any;
export type IdentityStatus = any;
export type IdentityTier = any;
export type InviteMemberDialogProps = any;
export type InviteMemberRequest =
  import("./types/organization").InviteMemberRequest;
export type InviteStatus = import("./types/organization").InviteStatus;
export type Job = import("./types/hpc").Job;
export type JobEvent = any;
export type JobEventType = any;
export type JobManifest = import("./types/hpc").JobManifest;
export type JobOutput = import("./types/hpc").JobOutput;
export type JobOutputReference = import("./types/hpc").JobOutputReference;
export type JobOutputType = import("./types/hpc").JobOutputType;
export type JobParameter = any;
export type JobPriceQuote = import("./types/hpc").JobPriceQuote;
export type JobResources = import("./types/hpc").JobResources;
export type JobStatus = any;
export type JobStatusChange = any;
export type JobSubmission = any;
export type JobSubmissionState = any;
export type JobSubmissionStep = any;
export type JobValidationError = import("./types/hpc").JobValidationError;
export type LogOptions = any;
export type MarketplaceAction = any;
export type MarketplaceState = any;
export type MemberListProps = any;
export type MemberMetadata = import("./types/organization").MemberMetadata;
export type MFAAuditEntry = any;
export type MFAChallenge = any;
export type MFAChallengeResponse = any;
export type MFAChallengeType = any;
export type MFAEnrollment = any;
export type MFAEnrollmentChallengeData = any;
export type MFAEnrollmentStep = any;
export type MFAError = any;
export type MFAErrorCode = any;
export type MFAFactor = any;
export type MFAFactorMetadata = any;
export type MFAFactorStatus = any;
export type MFAFactorType = any;
export type MFAPolicy = any;
export type MFAState = any;
export type MultiProviderClientOptions = any;
export type MultiProviderProviderProps = any;
export type MultiProviderWallet = any;
export type Offering = any;
export type OfferingDraft = import("./types/provider").OfferingDraft;
export type ProviderOffering = import("./types/provider").ProviderOffering;
export type OfferingFilter = any;
export type OfferingSortField = any;
export type Order = any;
export type OrderApiEndpoint = any;
export type OrderArtifact = any;
export type OrderConnectionStatus = any;
export type OrderCredential = any;
export type OrderEvent = any;
export type OrderListFilter = any;
export type OrderListProps = any;
export type OrderResourceAccess = any;
export type OrderResourceConnection = any;
export type OrderState = any;
export type OrderStatusProps = any;
export type OrderTrackingActions = any;
export type OrderTrackingContextValue = any;
export type OrderTrackingOrder = any;
export type OrderTrackingPageProps = any;
export type OrderTrackingProviderProps = any;
export type OrderTrackingState = any;
export type OrderTrackingStateValue = any;
export type OrderUsageAlert = any;
export type OrderUsageMetric = any;
export type OrderUsageSample = any;
export type OrderUsageSnapshot = any;
export type Organization = import("./types/organization").Organization;
export type OrganizationBillingPeriod =
  import("./types/organization").OrganizationBillingPeriod;
export type OrganizationBillingProps = any;
export type OrganizationBillingSummary =
  import("./types/organization").OrganizationBillingSummary;
export type OrganizationCardProps = any;
export type OrganizationDetailProps = any;
export type OrganizationInvite =
  import("./types/organization").OrganizationInvite;
export type OrganizationListProps = any;
export type OrganizationMember =
  import("./types/organization").OrganizationMember;
export type OrganizationMetadata =
  import("./types/organization").OrganizationMetadata;
export type OrganizationRole = import("./types/organization").OrganizationRole;
export type OrganizationSwitcherProps = any;
export type OrganizationMutationAction =
  | "create"
  | "invite"
  | "remove"
  | "update_role"
  | "leave";
export type OrganizationMutationRequest =
  | {
      chainId: string;
      accountAddress: string;
      action: "create";
      organization: CreateOrganizationRequest;
    }
  | {
      chainId: string;
      accountAddress: string;
      action: "invite";
      organizationId: string;
      invitation: InviteMemberRequest;
    }
  | {
      chainId: string;
      accountAddress: string;
      action: "remove";
      organizationId: string;
      memberAddress: string;
    }
  | {
      chainId: string;
      accountAddress: string;
      action: "update_role";
      organizationId: string;
      memberAddress: string;
      role: OrganizationRole;
    }
  | {
      chainId: string;
      accountAddress: string;
      action: "leave";
      organizationId: string;
    };
export interface OrganizationMutationContext {
  requestDigest: string;
  idempotencyKey: string;
  signal: AbortSignal;
}
export interface OrganizationMutationAdapter {
  readonly chainId: string;
  readonly accountAddress: string;
  mutateOrganization(
    request: OrganizationMutationRequest,
    context: OrganizationMutationContext,
  ): Promise<unknown>;
}
interface CommittedOrganizationMutationBase {
  status: "committed";
  code: 0;
  txHash: string;
  blockHeight: number;
  operationId: string;
  requestDigest: string;
  idempotencyKey: string;
  request: OrganizationMutationRequest;
}
export type CommittedOrganizationMutation =
  | (CommittedOrganizationMutationBase & {
      action: "create";
      organization: Organization;
    })
  | (CommittedOrganizationMutationBase & {
      action: "invite" | "remove" | "update_role";
      members: readonly OrganizationMember[];
    })
  | (CommittedOrganizationMutationBase & {
      action: "leave";
      organizationId: string;
    });
export class OrganizationMutationError extends Error {
  readonly code:
    | "feature_unavailable"
    | "invalid_request"
    | "invalid_committed_result"
    | "request_changed"
    | "submission_cancelled"
    | "submission_in_progress";
  constructor(code: OrganizationMutationError["code"]);
}
export function buildOrganizationMutationRequest(
  action: OrganizationMutationAction,
  binding: { chainId: string; accountAddress: string },
  input: CreateOrganizationRequest | Record<string, unknown>,
): OrganizationMutationRequest;
export function digestOrganizationMutationRequest(
  request: OrganizationMutationRequest,
): Promise<string>;
export function requireOrganizationMutationAdapter(
  adapter: OrganizationMutationAdapter | undefined,
  binding: { chainId: string; accountAddress: string },
): OrganizationMutationAdapter;
export function validateCommittedOrganizationMutation(
  value: unknown,
  request: OrganizationMutationRequest,
  requestDigest: string,
): CommittedOrganizationMutation;
export type PortalConfig = any;
export type PortalWalletType = any;
export type PricingConfig = any;
export type ProviderAPIClientOptions = any;
export type ProviderAPIErrorDetails = any;
export type ProviderDeploymentAction =
  | "start"
  | "stop"
  | "restart"
  | "update"
  | "terminate";
export type ProviderDeploymentActionStatus = "accepted" | "committed";
export interface ProviderDeploymentActionTxEvidence {
  hash: string;
  chainId: string;
  height: number;
}
export interface ProviderDeploymentActionReceipt {
  operationId: string;
  action: ProviderDeploymentAction;
  deploymentId: string;
  providerId: string;
  status: ProviderDeploymentActionStatus;
  issuedAt: Date;
  completedAt: Date;
  state: string;
  version: string;
  revision: string;
  txEvidence?: ProviderDeploymentActionTxEvidence;
}
export interface ProviderDeploymentActionCapability {
  receiptVersion: "v1";
  requiresChainSigning: boolean;
}
export type ProviderDeploymentActionErrorCode =
  | "feature_unavailable"
  | "action_rejected"
  | "malformed_receipt"
  | "receipt_mismatch"
  | "refresh_failed"
  | "deployment_drift"
  | "duplicate_action"
  | "chain_signing_required";
export type ProviderDeploymentTxEvidenceValidator = (
  evidence: ProviderDeploymentActionTxEvidence,
  receipt: Omit<ProviderDeploymentActionReceipt, "txEvidence">,
) => boolean | Promise<boolean>;
export interface ProviderDeploymentActionValidationContext {
  action: ProviderDeploymentAction;
  deploymentId: string;
  providerId: string;
  validateTxEvidence?: ProviderDeploymentTxEvidenceValidator;
}
export type ProviderDeploymentActionReceiptValidator = (
  value: unknown,
  context: ProviderDeploymentActionValidationContext,
) => Promise<ProviderDeploymentActionReceipt>;
export type ProviderHealth = any;
export type ProviderHealthStatus = any;
export type ProviderProfile = any;
export type ProviderRecord = any;
export type ProviderRegistration = any;
export type ProviderState = any;
export type ProviderStatus = any;
export type QueryClient = any;
export type RemediationPath = any;
export type ResourceAccessProps = any;
export type ResourceMetrics = any;
export type ScopeRequirement = any;
export type SensitiveTransactionType = any;
export type ServiceStatus = any;
export type SessionConfig = any;
export type SessionInfo = any;
export type SessionToken = any;
export type SettlementSummary = any;
export type ProviderShellSessionErrorCode = any;
export type ProviderShellSessionCapability = any;
export type ProviderShellSessionReceipt = any;
export type ShellEligibilityProjection = any;
export type ShellSessionValidationContext = any;
export type SignedRequestHeaders = any;
export type SigningResult = any;
export type SignRequestOptions = any;
export type SSOCredentials = any;
export type TransactionModalProps = any;
export type TransactionOptions = any;
export type TransactionPreview = any;
export type TransactionResult = any;
export type TransactionValidationResult = any;
export type TrustedBrowser = any;
export type UploadRecord = any;
export type UsageMetric = any;
export type UsageMonitorProps = any;
export type UsageRecord = any;
export type UseAggregatedDeploymentsOptions = any;
export type UseAggregatedMetricsOptions = any;
export const useDeploymentShell: any;
export type UseDeploymentShellOptions = any;
export type UseDeploymentShellResult = any;
export type VerificationRecord = any;
export type VerificationScope = any;
export type VerificationScopeType = any;
export type WalletAccount = any;
export type WalletAccountDisplayProps = any;
export type WalletButtonProps = any;
export type WalletChainInfo = any;
export type WalletConfig = any;
export type WalletConnectionStatus = any;
export type WalletContextValue = any;
export type WalletCredentials = any;
export type WalletDetectionResult = any;
export type WalletError = any;
export type WalletIconProps = any;
export type WalletModalProps = any;
export type WalletNetworkBadgeProps = any;
export type WalletOption = any;
export type WalletProviderConfig = any;
export type WalletSession = any;
export type WalletSessionConfig = any;
export type WalletSignOptions = any;
export type WalletSkeletonProps = any;
export type WalletState = any;
export type WorkloadCategory = any;
export type WorkloadTemplate = import("./types/hpc").WorkloadTemplate;
export class ChainClientError extends Error {
  status?: number;
  code?: string;
  endpoint?: string;
}
export class ChainHttpError extends Error {
  status?: number;
  code?: string;
  endpoint?: string;
}
export const WalletProvider: any;
export const useWallet: any;
export const WalletErrorClass: any;
export const WalletSessionManager: any;
export const WalletDetector: any;
export const GAS_TIERS: any;
export type WalletErrorType = any;
export type MultiProviderClient = any;
