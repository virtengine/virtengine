/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * MFA API service layer for portal-to-chain SDK communication.
 */

import { apiClient } from '@/lib/api-client';
import type {
  MFAFactor,
  MFAPolicy,
  MFAChallenge,
  MFAChallengeResponse,
  MFAAuditEntry,
  TrustedBrowser,
  SensitiveTransactionType,
} from '@/lib/portal-adapter';
import type {
  TOTPEnrollmentData,
  WebAuthnEnrollmentData,
  BackupCodesData,
  SMSEnrollmentData,
  EmailEnrollmentData,
} from './types';

const MFA_BASE = '/api/mfa';

/** Fetch all enrolled MFA factors for the current account */
export async function fetchFactors(): Promise<MFAFactor[]> {
  return apiClient.get<MFAFactor[]>(`${MFA_BASE}/factors`);
}

/** Fetch current MFA policy */
export async function fetchPolicy(): Promise<MFAPolicy> {
  return apiClient.get<MFAPolicy>(`${MFA_BASE}/policy`);
}

/** Fetch trusted browsers */
export async function fetchTrustedBrowsers(): Promise<TrustedBrowser[]> {
  return apiClient.get<TrustedBrowser[]>(`${MFA_BASE}/trusted-browsers`);
}

/** Fetch MFA audit log */
export async function fetchAuditLog(limit = 50): Promise<MFAAuditEntry[]> {
  return apiClient.get<MFAAuditEntry[]>(`${MFA_BASE}/audit`, {
    query: { limit },
  });
}

/** Start TOTP enrollment */
export async function startTOTPEnrollment(): Promise<TOTPEnrollmentData> {
  return apiClient.post<TOTPEnrollmentData>(`${MFA_BASE}/enroll/totp`);
}

/** Verify TOTP enrollment with a 6-digit code */
export async function verifyTOTPEnrollment(code: string, name?: string): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/enroll/totp/verify`, { code, name });
}

/** Start WebAuthn enrollment - returns credential creation options */
export async function startWebAuthnEnrollment(): Promise<WebAuthnEnrollmentData> {
  return apiClient.post<WebAuthnEnrollmentData>(`${MFA_BASE}/enroll/webauthn`);
}

/** Start SMS enrollment */
export async function startSMSEnrollment(phone: string): Promise<SMSEnrollmentData> {
  return apiClient.post<SMSEnrollmentData>(`${MFA_BASE}/enroll/sms`, {
    phone,
  });
}

/** Verify SMS enrollment with a code */
export async function verifySMSEnrollment(
  challengeId: string,
  code: string,
  name?: string
): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/enroll/sms/verify`, {
    challengeId,
    code,
    name,
  });
}

/** Start email enrollment */
export async function startEmailEnrollment(email: string): Promise<EmailEnrollmentData> {
  return apiClient.post<EmailEnrollmentData>(`${MFA_BASE}/enroll/email`, {
    email,
  });
}

/** Verify email enrollment with a code */
export async function verifyEmailEnrollment(
  challengeId: string,
  code: string,
  name?: string
): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/enroll/email/verify`, {
    challengeId,
    code,
    name,
  });
}

/** Complete WebAuthn enrollment with the attestation result */
export async function completeWebAuthnEnrollment(
  challengeId: string,
  credential: {
    id: string;
    rawId: string;
    type: string;
    response: {
      attestationObject: string;
      clientDataJSON: string;
    };
  },
  name?: string
): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/enroll/webauthn/verify`, {
    challengeId,
    credential,
    name,
  });
}

/** Generate backup codes */
export async function generateBackupCodes(): Promise<BackupCodesData> {
  return apiClient.post<BackupCodesData>(`${MFA_BASE}/enroll/backup-codes`);
}

/** Remove an enrolled factor (requires MFA verification) */
export async function removeFactor(factorId: string): Promise<void> {
  await apiClient.request(`${MFA_BASE}/factors/${factorId}`, { method: 'DELETE' });
}

/** Enable or disable (suspend) a factor */
export async function toggleFactor(factorId: string, enabled: boolean): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/factors/${factorId}/toggle`, {
    enabled,
  });
}

/** Set a factor as the primary factor */
export async function setPrimaryFactor(factorId: string): Promise<MFAFactor> {
  return apiClient.post<MFAFactor>(`${MFA_BASE}/factors/${factorId}/primary`);
}

/** Create an MFA challenge for a sensitive transaction */
export async function createChallenge(
  transactionType: SensitiveTransactionType
): Promise<MFAChallenge> {
  return apiClient.post<MFAChallenge>(`${MFA_BASE}/challenge`, {
    transactionType,
  });
}

/** Verify an MFA challenge with a code (OTP/SMS/Email/Backup) */
export async function verifyChallenge(
  challengeId: string,
  factorId: string,
  code: string
): Promise<MFAChallengeResponse> {
  return apiClient.post<MFAChallengeResponse>(`${MFA_BASE}/challenge/verify`, {
    challengeId,
    factorId,
    code,
  });
}

/** Verify a FIDO2/WebAuthn challenge */
export async function verifyWebAuthnChallenge(
  challengeId: string,
  factorId: string,
  assertion: {
    id: string;
    rawId: string;
    type: string;
    response: {
      authenticatorData: string;
      clientDataJSON: string;
      signature: string;
    };
  }
): Promise<MFAChallengeResponse> {
  return apiClient.post<MFAChallengeResponse>(`${MFA_BASE}/challenge/verify-fido2`, {
    challengeId,
    factorId,
    assertion,
  });
}

/** Revoke a trusted browser */
export async function revokeTrustedBrowser(browserId: string): Promise<void> {
  await apiClient.request(`${MFA_BASE}/trusted-browsers/${browserId}`, {
    method: 'DELETE',
  });
}

/** Trust current browser after MFA verification */
export async function trustCurrentBrowser(deviceName: string): Promise<TrustedBrowser> {
  return apiClient.post<TrustedBrowser>(`${MFA_BASE}/trusted-browsers`, {
    deviceName,
  });
}

/** Submit a recovery request using backup code */
export async function submitRecovery(backupCode: string): Promise<{ success: boolean }> {
  return apiClient.post<{ success: boolean }>(`${MFA_BASE}/recovery`, {
    backupCode,
  });
}
