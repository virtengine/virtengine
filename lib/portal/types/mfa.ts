/**
 * MFA Types
 * VE-702: MFA enrollment, policy configuration, trusted browser UX
 *
 * @packageDocumentation
 */

/**
 * MFA state for a user
 */
export interface MFAState {
  /**
   * Whether MFA data is loading
   */
  isLoading: boolean;

  /**
   * Whether MFA is enabled for the account
   */
  isEnabled: boolean;

  /**
   * Enrolled MFA factors
   */
  enrolledFactors: MFAFactor[];

  /**
   * Current MFA policy
   */
  policy: MFAPolicy | null;

  /**
   * Trusted browsers
   */
  trustedBrowsers: TrustedBrowser[];

  /**
   * Active challenge (if any)
   */
  activeChallenge: MFAChallenge | null;

  /**
   * Audit history of MFA-gated actions
   */
  auditHistory: MFAAuditEntry[];

  /**
   * Any MFA-related error
   */
  error: MFAError | null;
}

/**
 * MFA factor types
 */
export type MFAFactorType =
  | 'otp'       // Time-based OTP (TOTP) app
  | 'fido2'     // FIDO2/WebAuthn security key
  | 'sms'       // SMS verification (less secure)
  | 'biometric' // VEID biometric verification
  | 'email';    // Email verification (backup only)

/**
 * MFA factor enrollment
 */
export interface MFAFactor {
  /**
   * Factor ID
   */
  id: string;

  /**
   * Factor type
   */
  type: MFAFactorType;

  /**
   * Factor name (user-defined)
   */
  name: string;

  /**
   * When factor was enrolled
   */
  enrolledAt: number;

  /**
   * When factor was last used
   */
  lastUsedAt: number | null;

  /**
   * Whether this is the primary factor
   */
  isPrimary: boolean;

  /**
   * Factor status
   */
  status: MFAFactorStatus;

  /**
   * Factor-specific metadata (non-sensitive)
   */
  metadata: MFAFactorMetadata;
}

/**
 * MFA factor status
 */
export type MFAFactorStatus =
  | 'active'
  | 'suspended'
  | 'expired';

/**
 * MFA factor metadata (non-sensitive)
 */
export interface MFAFactorMetadata {
  /**
   * For OTP: issuer name
   */
  issuer?: string;

  /**
   * For FIDO2: credential type
   */
  credentialType?: 'platform' | 'cross-platform';

  /**
   * For FIDO2: authenticator AAGUID
   */
  aaguid?: string;

  /**
   * For SMS: masked phone number
   */
  maskedPhone?: string;

  /**
   * For biometric: biometric type
   */
  biometricType?: 'face' | 'fingerprint';
}

/**
 * MFA enrollment state (during enrollment flow)
 */
export interface MFAEnrollment {
  /**
   * Factor type being enrolled
   */
  type: MFAFactorType;

  /**
   * Enrollment step
   */
  step: MFAEnrollmentStep;

  /**
   * Challenge data (non-sensitive)
   */
  challengeData?: MFAEnrollmentChallengeData;

  /**
   * Error during enrollment
   */
  error?: MFAError;
}

/**
 * MFA enrollment steps
 */
export type MFAEnrollmentStep =
  | 'select_type'
  | 'configure'
  | 'verify'
  | 'confirm'
  | 'complete';

/**
 * MFA enrollment challenge data
 */
export interface MFAEnrollmentChallengeData {
  /**
   * For OTP: QR code data URL (contains secret URI)
   */
  qrCodeDataUrl?: string;

  /**
   * For OTP: manual entry key (backup)
   */
  manualEntryKey?: string;

  /**
   * For FIDO2: creation options
   */
  fido2Options?: PublicKeyCredentialCreationOptions;

  /**
   * For SMS: masked phone for confirmation
   */
  maskedPhone?: string;

  /**
   * Backup codes (shown once during enrollment)
   */
  backupCodes?: string[];
}

/**
 * MFA policy configuration
 */
export interface MFAPolicy {
  /**
   * Policy ID
   */
  id: string;

  /**
   * When policy was last updated
   */
  updatedAt: number;

  /**
   * Required factor types (at least one must be satisfied)
   */
  requiredFactorTypes: MFAFactorType[];

  /**
   * Number of factors required
   */
  requiredFactorCount: number;

  /**
   * Sensitive transaction types that require MFA
   */
  sensitiveTransactions: SensitiveTransactionType[];

  /**
   * Whether to allow trusted browsers
   */
  allowTrustedBrowsers: boolean;

  /**
   * Trusted browser duration in seconds
   */
  trustedBrowserDurationSeconds: number;

  /**
   * Whether biometric can be used alone for low-value transactions
   */
  biometricForLowValue: boolean;

  /**
   * Low-value transaction threshold
   */
  lowValueThreshold: string; // Token amount as string
}

/**
 * Sensitive transaction types
 */
export type SensitiveTransactionType =
  | 'account_recovery'
  | 'key_rotation'
  | 'high_value_order'
  | 'provider_registration'
  | 'offering_creation'
  | 'hpc_job_submission'
  | 'withdrawal'
  | 'delegation_change';

/**
 * Trusted browser registration
 */
export interface TrustedBrowser {
  /**
   * Browser ID
   */
  id: string;

  /**
   * Browser name (derived from user agent)
   */
  browserName: string;

  /**
   * Device name (user-defined)
   */
  deviceName: string;

  /**
   * When browser was trusted
   */
  trustedAt: number;

  /**
   * When trust expires
   */
  expiresAt: number;

  /**
   * Last used timestamp
   */
  lastUsedAt: number;

  /**
   * Device fingerprint hash (non-identifying)
   */
  fingerprintHash: string;

  /**
   * IP address region (non-specific)
   */
  region?: string;
}

/**
 * MFA challenge during transaction
 */
export interface MFAChallenge {
  /**
   * Challenge ID
   */
  id: string;

  /**
   * Challenge type
   */
  type: MFAChallengeType;

  /**
   * Transaction type requiring MFA
   */
  transactionType: SensitiveTransactionType;

  /**
   * Available factors for this challenge
   */
  availableFactors: MFAFactor[];

  /**
   * When challenge was created
   */
  createdAt: number;

  /**
   * When challenge expires
   */
  expiresAt: number;

  /**
   * Challenge nonce (for replay protection)
   */
  nonce: string;

  /**
   * Transaction summary (non-sensitive)
   */
  transactionSummary: string;
}

/**
 * MFA challenge types
 */
export type MFAChallengeType =
  | 'otp_verify'
  | 'fido2_assert'
  | 'sms_verify'
  | 'biometric_verify'
  | 'email_verify';

/**
 * MFA challenge response
 */
export interface MFAChallengeResponse {
  /**
   * Challenge ID
   */
  challengeId: string;

  /**
   * Factor ID used
   */
  factorId: string;

  /**
   * Response type
   */
  type: MFAChallengeType;

  /**
   * Response data
   * NOTE: Actual credentials are handled internally, never exposed
   */
  verified: boolean;

  /**
   * Timestamp of response
   */
  respondedAt: number;
}

/**
 * MFA audit entry
 */
export interface MFAAuditEntry {
  /**
   * Entry ID
   */
  id: string;

  /**
   * Transaction type
   */
  transactionType: SensitiveTransactionType;

  /**
   * Factor used
   */
  factorType: MFAFactorType;

  /**
   * Whether MFA was successful
   */
  success: boolean;

  /**
   * Timestamp
   */
  timestamp: number;

  /**
   * Transaction hash (if completed)
   */
  txHash?: string;

  /**
   * IP region (non-specific)
   */
  region?: string;
}

/**
 * MFA error
 */
export interface MFAError {
  /**
   * Error code
   */
  code: MFAErrorCode;

  /**
   * Human-readable message
   */
  message: string;

  /**
   * Remaining attempts (if applicable)
   */
  remainingAttempts?: number;

  /**
   * Lockout until (if applicable)
   */
  lockoutUntil?: number;
}

/**
 * MFA error codes
 */
export type MFAErrorCode =
  | 'invalid_code'
  | 'expired_challenge'
  | 'factor_not_found'
  | 'factor_suspended'
  | 'too_many_attempts'
  | 'lockout'
  | 'enrollment_failed'
  | 'verification_failed'
  | 'fido2_error'
  | 'network_error'
  | 'unknown';

/**
 * Initial MFA state
 */
export const initialMFAState: MFAState = {
  isLoading: false,
  isEnabled: false,
  enrolledFactors: [],
  policy: null,
  trustedBrowsers: [],
  activeChallenge: null,
  auditHistory: [],
  error: null,
};

/**
 * Get factor display name
 */
export function getFactorDisplayName(type: MFAFactorType): string {
  switch (type) {
    case 'otp':
      return 'Authenticator App (TOTP)';
    case 'fido2':
      return 'Security Key (FIDO2)';
    case 'sms':
      return 'SMS Verification';
    case 'biometric':
      return 'Biometric (VEID)';
    case 'email':
      return 'Email Verification';
    default:
      return 'Unknown Factor';
  }
}

/**
 * Get factor security level
 */
export function getFactorSecurityLevel(type: MFAFactorType): 'high' | 'medium' | 'low' {
  switch (type) {
    case 'fido2':
      return 'high';
    case 'otp':
    case 'biometric':
      return 'medium';
    case 'sms':
    case 'email':
      return 'low';
    default:
      return 'low';
  }
}
