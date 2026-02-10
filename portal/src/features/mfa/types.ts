/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Portal-specific MFA types extending lib/portal MFA types.
 */

export type {
  MFAState,
  MFAFactor,
  MFAFactorType,
  MFAPolicy,
  MFAEnrollment,
  MFAEnrollmentStep,
  MFAEnrollmentChallengeData,
  MFAChallenge,
  MFAChallengeResponse,
  MFAChallengeType,
  MFAAuditEntry,
  MFAError,
  MFAErrorCode,
  TrustedBrowser,
  SensitiveTransactionType,
} from '@/lib/portal-adapter';

/** TOTP enrollment data returned from the API */
export interface TOTPEnrollmentData {
  qrCodeDataUrl: string;
  manualEntryKey: string;
  issuer: string;
  accountName: string;
}

/** WebAuthn enrollment data returned from the API */
export interface WebAuthnEnrollmentData {
  creationOptions: PublicKeyCredentialCreationOptions;
  challengeId: string;
}

/** SMS enrollment data returned from the API */
export interface SMSEnrollmentData {
  challengeId: string;
  maskedPhone: string;
}

/** Email enrollment data returned from the API */
export interface EmailEnrollmentData {
  challengeId: string;
  maskedEmail: string;
}

/** Backup codes data returned from the API */
export interface BackupCodesData {
  codes: string[];
  generatedAt: number;
}

/** Factor removal confirmation state */
export interface FactorRemovalState {
  factorId: string;
  factorName: string;
  factorType: string;
  isConfirming: boolean;
}

/** Recovery flow step */
export type RecoveryStep = 'identify' | 'verify' | 'reset' | 'complete';

/** Recovery flow state */
export interface RecoveryState {
  step: RecoveryStep;
  method: 'backup_code' | 'email' | null;
  isSubmitting: boolean;
  error: string | null;
}
