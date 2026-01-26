/**
 * Identity Types
 * VE-701: VEID onboarding, identity score display, re-verification prompts
 *
 * @packageDocumentation
 */

/**
 * Identity state for a user
 */
export interface IdentityState {
  /**
   * Whether identity data is loading
   */
  isLoading: boolean;

  /**
   * Current identity status
   */
  status: IdentityStatus;

  /**
   * Identity score (0-100)
   */
  score: IdentityScore | null;

  /**
   * Verification scopes completed
   */
  completedScopes: VerificationScope[];

  /**
   * Upload history
   */
  uploadHistory: UploadRecord[];

  /**
   * Verification records
   */
  verificationRecords: VerificationRecord[];

  /**
   * Any identity-related error
   */
  error: IdentityError | null;
}

/**
 * Identity status enumeration
 */
export type IdentityStatus =
  | 'unknown'      // No identity data submitted
  | 'pending'      // Identity submitted, awaiting verification
  | 'processing'   // Identity being processed by validators
  | 'verified'     // Identity verified successfully
  | 'rejected'     // Identity verification failed
  | 'expired';     // Identity verification expired, needs re-verification

/**
 * Identity score with metadata
 */
export interface IdentityScore {
  /**
   * Score value (0-100)
   */
  value: number;

  /**
   * Score tier derived from value
   */
  tier: IdentityTier;

  /**
   * When the score was computed
   */
  computedAt: number;

  /**
   * Model version used for scoring
   */
  modelVersion: string;

  /**
   * Score breakdown by category
   */
  breakdown: ScoreBreakdown;

  /**
   * Block height where score was committed
   */
  blockHeight: number;
}

/**
 * Identity tier based on score
 */
export type IdentityTier =
  | 'unverified' // Score 0 or no score
  | 'basic'     // Score 1-40
  | 'standard'  // Score 41-70
  | 'premium'   // Score 71-90
  | 'elite';    // Score 91-100

/**
 * Score breakdown by verification category
 */
export interface ScoreBreakdown {
  /**
   * Document verification score (0-25)
   */
  document: number;

  /**
   * Facial verification score (0-25)
   */
  facial: number;

  /**
   * Metadata consistency score (0-25)
   */
  metadata: number;

  /**
   * Historical trust score (0-25)
   */
  trust: number;
}

/**
 * Verification scope types
 */
export type VerificationScopeType =
  | 'id_document'    // Government-issued ID
  | 'selfie'         // Live selfie verification
  | 'email'          // Email verification
  | 'sso'            // SSO provider linkage
  | 'domain'         // Domain ownership
  | 'biometric';     // Additional biometric factors

/**
 * Verification scope
 */
export interface VerificationScope {
  /**
   * Scope type
   */
  type: VerificationScopeType;

  /**
   * Whether scope is completed
   */
  completed: boolean;

  /**
   * When scope was verified
   */
  verifiedAt?: number;

  /**
   * Expiration time (if applicable)
   */
  expiresAt?: number;

  /**
   * Sub-type details (e.g., document type for id_document)
   */
  subType?: string;
}

/**
 * Upload record for identity documents
 */
export interface UploadRecord {
  /**
   * Upload ID
   */
  id: string;

  /**
   * Scope type uploaded
   */
  scopeType: VerificationScopeType;

  /**
   * Upload timestamp
   */
  uploadedAt: number;

  /**
   * Upload status
   */
  status: UploadStatus;

  /**
   * Hash of uploaded content (for reference only, content is encrypted)
   */
  contentHash: string;

  /**
   * Transaction hash where upload was committed
   */
  txHash: string;
}

/**
 * Upload status
 */
export type UploadStatus =
  | 'pending'     // Uploaded, awaiting processing
  | 'processing'  // Being processed by validators
  | 'accepted'    // Upload accepted and verified
  | 'rejected'    // Upload rejected
  | 'expired';    // Upload expired

/**
 * Verification record (summary, no sensitive data)
 */
export interface VerificationRecord {
  /**
   * Record ID
   */
  id: string;

  /**
   * Related upload IDs
   */
  uploadIds: string[];

  /**
   * Verification timestamp
   */
  verifiedAt: number;

  /**
   * Verification result
   */
  result: VerificationResult;

  /**
   * Score contribution
   */
  scoreContribution: number;

  /**
   * Validator that performed verification
   */
  validatorAddress: string;

  /**
   * Block height of verification
   */
  blockHeight: number;
}

/**
 * Verification result
 */
export type VerificationResult =
  | 'pass'
  | 'fail'
  | 'inconclusive'
  | 'needs_review';

/**
 * Identity error
 */
export interface IdentityError {
  /**
   * Error code
   */
  code: IdentityErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Remediation suggestions
   */
  remediations: string[];
}

/**
 * Identity error codes
 */
export type IdentityErrorCode =
  | 'upload_failed'
  | 'verification_failed'
  | 'score_insufficient'
  | 'scope_missing'
  | 'scope_expired'
  | 'document_invalid'
  | 'selfie_mismatch'
  | 'duplicate_identity'
  | 'rate_limited'
  | 'network_error'
  | 'unknown';

/**
 * Identity gating error (returned when action requires higher identity)
 */
export interface IdentityGatingError {
  /**
   * Action that was blocked
   */
  action: string;

  /**
   * Required score for action
   */
  requiredScore: number;

  /**
   * Current user score
   */
  currentScore: number;

  /**
   * Required scopes for action
   */
  requiredScopes: VerificationScopeType[];

  /**
   * Missing scopes
   */
  missingScopes: VerificationScopeType[];

  /**
   * Remediation path
   */
  remediation: RemediationPath;
}

/**
 * Remediation path for identity issues
 */
export interface RemediationPath {
  /**
   * Steps to resolve the issue
   */
  steps: RemediationStep[];

  /**
   * Estimated time to complete
   */
  estimatedTimeMinutes: number;

  /**
   * Link to approved capture client (if applicable)
   */
  captureClientUrl?: string;
}

/**
 * Remediation step
 */
export interface RemediationStep {
  /**
   * Step order
   */
  order: number;

  /**
   * Step title
   */
  title: string;

  /**
   * Step description
   */
  description: string;

  /**
   * Action to take
   */
  action: RemediationAction;

  /**
   * Whether step is completed
   */
  completed: boolean;
}

/**
 * Remediation action types
 */
export type RemediationAction =
  | { type: 'upload_scope'; scopeType: VerificationScopeType }
  | { type: 'reverify_scope'; scopeType: VerificationScopeType }
  | { type: 'complete_mfa' }
  | { type: 'contact_support' }
  | { type: 'wait_processing' }
  | { type: 'external_link'; url: string };

/**
 * Scope requirements for marketplace actions
 */
export interface ScopeRequirement {
  /**
   * Action name
   */
  action: MarketplaceAction;

  /**
   * Minimum required score
   */
  minScore: number;

  /**
   * Required scopes
   */
  requiredScopes: VerificationScopeType[];

  /**
   * Optional scopes that increase trust
   */
  optionalScopes: VerificationScopeType[];

  /**
   * Whether MFA is required
   */
  mfaRequired: boolean;
}

/**
 * Marketplace actions that may require identity verification
 */
export type MarketplaceAction =
  | 'browse_offerings'
  | 'view_offering_details'
  | 'place_order'
  | 'place_high_value_order'
  | 'register_provider'
  | 'create_offering'
  | 'submit_hpc_job'
  | 'access_outputs';

/**
 * Get tier from score
 */
export function getTierFromScore(score: number): IdentityTier {
  if (score <= 0) return 'unverified';
  if (score <= 40) return 'basic';
  if (score <= 70) return 'standard';
  if (score <= 90) return 'premium';
  return 'elite';
}

/**
 * Initial identity state
 */
export const initialIdentityState: IdentityState = {
  isLoading: false,
  status: 'unknown',
  score: null,
  completedScopes: [],
  uploadHistory: [],
  verificationRecords: [],
  error: null,
};
