/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Zustand store for portal MFA state management.
 */

import { create } from 'zustand';
import type {
  MFAFactor,
  MFAPolicy,
  MFAChallenge,
  MFAAuditEntry,
  TrustedBrowser,
} from '@/lib/portal-adapter';
import type { TOTPEnrollmentData, WebAuthnEnrollmentData, BackupCodesData } from './types';
import * as mfaApi from './api';

export interface MFAStoreState {
  /** Whether initial data is loading */
  isLoading: boolean;
  /** Whether MFA is enabled (at least one active factor) */
  isEnabled: boolean;
  /** Enrolled MFA factors */
  factors: MFAFactor[];
  /** Current MFA policy */
  policy: MFAPolicy | null;
  /** Trusted browsers */
  trustedBrowsers: TrustedBrowser[];
  /** Audit log entries */
  auditLog: MFAAuditEntry[];
  /** Active challenge (when verifying a sensitive action) */
  activeChallenge: MFAChallenge | null;
  /** In-flight TOTP enrollment data */
  totpEnrollment: TOTPEnrollmentData | null;
  /** In-flight WebAuthn enrollment data */
  webAuthnEnrollment: WebAuthnEnrollmentData | null;
  /** Generated backup codes (shown once) */
  backupCodes: BackupCodesData | null;
  /** General error message */
  error: string | null;
  /** Whether a mutation is in progress */
  isMutating: boolean;
}

export interface MFAStoreActions {
  /** Load all MFA data for the current account */
  loadMFAData: () => Promise<void>;
  /** Start TOTP enrollment */
  startTOTPEnrollment: () => Promise<void>;
  /** Verify TOTP enrollment */
  verifyTOTPEnrollment: (code: string, name?: string) => Promise<MFAFactor>;
  /** Start WebAuthn enrollment */
  startWebAuthnEnrollment: () => Promise<void>;
  /** Complete WebAuthn enrollment */
  completeWebAuthnEnrollment: (
    credential: PublicKeyCredential,
    name?: string
  ) => Promise<MFAFactor>;
  /** Generate backup codes */
  generateBackupCodes: () => Promise<void>;
  /** Remove a factor */
  removeFactor: (factorId: string) => Promise<void>;
  /** Toggle factor enabled/disabled */
  toggleFactor: (factorId: string, enabled: boolean) => Promise<void>;
  /** Set primary factor */
  setPrimaryFactor: (factorId: string) => Promise<void>;
  /** Create an MFA challenge */
  createChallenge: (
    transactionType: Parameters<typeof mfaApi.createChallenge>[0]
  ) => Promise<MFAChallenge>;
  /** Verify an MFA challenge with code */
  verifyChallenge: (factorId: string, code: string) => Promise<boolean>;
  /** Revoke a trusted browser */
  revokeTrustedBrowser: (browserId: string) => Promise<void>;
  /** Clear enrollment state */
  clearEnrollment: () => void;
  /** Clear backup codes from memory */
  clearBackupCodes: () => void;
  /** Clear error */
  clearError: () => void;
  /** Reset store */
  reset: () => void;
}

const initialState: MFAStoreState = {
  isLoading: false,
  isEnabled: false,
  factors: [],
  policy: null,
  trustedBrowsers: [],
  auditLog: [],
  activeChallenge: null,
  totpEnrollment: null,
  webAuthnEnrollment: null,
  backupCodes: null,
  error: null,
  isMutating: false,
};

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = '';
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

export const useMFAStore = create<MFAStoreState & MFAStoreActions>((set, get) => ({
  ...initialState,

  loadMFAData: async () => {
    set({ isLoading: true, error: null });
    try {
      const [factors, policy, trustedBrowsers, auditLog] = await Promise.all([
        mfaApi.fetchFactors(),
        mfaApi.fetchPolicy(),
        mfaApi.fetchTrustedBrowsers(),
        mfaApi.fetchAuditLog(),
      ]);
      set({
        isLoading: false,
        isEnabled: factors.some((f) => f.status === 'active'),
        factors,
        policy,
        trustedBrowsers,
        auditLog,
      });
    } catch (err) {
      set({
        isLoading: false,
        error: err instanceof Error ? err.message : 'Failed to load MFA data',
      });
    }
  },

  startTOTPEnrollment: async () => {
    set({ isMutating: true, error: null, totpEnrollment: null });
    try {
      const data = await mfaApi.startTOTPEnrollment();
      set({ isMutating: false, totpEnrollment: data });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to start TOTP enrollment',
      });
    }
  },

  verifyTOTPEnrollment: async (code, name) => {
    set({ isMutating: true, error: null });
    try {
      const factor = await mfaApi.verifyTOTPEnrollment(code, name);
      set((s) => ({
        isMutating: false,
        totpEnrollment: null,
        factors: [...s.factors, factor],
        isEnabled: true,
      }));
      return factor;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'TOTP verification failed',
      });
      throw err;
    }
  },

  startWebAuthnEnrollment: async () => {
    set({ isMutating: true, error: null, webAuthnEnrollment: null });
    try {
      const data = await mfaApi.startWebAuthnEnrollment();
      set({ isMutating: false, webAuthnEnrollment: data });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to start WebAuthn enrollment',
      });
    }
  },

  completeWebAuthnEnrollment: async (credential, name) => {
    const enrollment = get().webAuthnEnrollment;
    if (!enrollment) throw new Error('No WebAuthn enrollment in progress');

    set({ isMutating: true, error: null });
    try {
      const attestationResponse = credential.response as AuthenticatorAttestationResponse;
      const factor = await mfaApi.completeWebAuthnEnrollment(
        enrollment.challengeId,
        {
          id: credential.id,
          rawId: arrayBufferToBase64(credential.rawId),
          type: credential.type,
          response: {
            attestationObject: arrayBufferToBase64(attestationResponse.attestationObject),
            clientDataJSON: arrayBufferToBase64(attestationResponse.clientDataJSON),
          },
        },
        name
      );
      set((s) => ({
        isMutating: false,
        webAuthnEnrollment: null,
        factors: [...s.factors, factor],
        isEnabled: true,
      }));
      return factor;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'WebAuthn enrollment failed',
      });
      throw err;
    }
  },

  generateBackupCodes: async () => {
    set({ isMutating: true, error: null, backupCodes: null });
    try {
      const data = await mfaApi.generateBackupCodes();
      set({ isMutating: false, backupCodes: data });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to generate backup codes',
      });
    }
  },

  removeFactor: async (factorId) => {
    set({ isMutating: true, error: null });
    try {
      await mfaApi.removeFactor(factorId);
      set((s) => {
        const factors = s.factors.filter((f) => f.id !== factorId);
        return {
          isMutating: false,
          factors,
          isEnabled: factors.some((f) => f.status === 'active'),
        };
      });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to remove factor',
      });
      throw err;
    }
  },

  toggleFactor: async (factorId, enabled) => {
    set({ isMutating: true, error: null });
    try {
      const updated = await mfaApi.toggleFactor(factorId, enabled);
      set((s) => ({
        isMutating: false,
        factors: s.factors.map((f) => (f.id === factorId ? updated : f)),
        isEnabled: s.factors.some((f) =>
          f.id === factorId ? updated.status === 'active' : f.status === 'active'
        ),
      }));
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to update factor',
      });
      throw err;
    }
  },

  setPrimaryFactor: async (factorId) => {
    set({ isMutating: true, error: null });
    try {
      await mfaApi.setPrimaryFactor(factorId);
      set((s) => ({
        isMutating: false,
        factors: s.factors.map((f) => ({
          ...f,
          isPrimary: f.id === factorId,
        })),
      }));
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to set primary factor',
      });
      throw err;
    }
  },

  createChallenge: async (transactionType) => {
    set({ isMutating: true, error: null });
    try {
      const challenge = await mfaApi.createChallenge(transactionType);
      set({ isMutating: false, activeChallenge: challenge });
      return challenge;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to create challenge',
      });
      throw err;
    }
  },

  verifyChallenge: async (factorId, code) => {
    const challenge = get().activeChallenge;
    if (!challenge) throw new Error('No active challenge');

    set({ isMutating: true, error: null });
    try {
      const result = await mfaApi.verifyChallenge(challenge.id, factorId, code);
      if (result.verified) {
        set({ isMutating: false, activeChallenge: null });
      } else {
        set({ isMutating: false });
      }
      return result.verified;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Verification failed',
      });
      throw err;
    }
  },

  revokeTrustedBrowser: async (browserId) => {
    set({ isMutating: true, error: null });
    try {
      await mfaApi.revokeTrustedBrowser(browserId);
      set((s) => ({
        isMutating: false,
        trustedBrowsers: s.trustedBrowsers.filter((b) => b.id !== browserId),
      }));
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to revoke browser',
      });
      throw err;
    }
  },

  clearEnrollment: () => {
    set({ totpEnrollment: null, webAuthnEnrollment: null, error: null });
  },

  clearBackupCodes: () => {
    set({ backupCodes: null });
  },

  clearError: () => {
    set({ error: null });
  },

  reset: () => {
    set(initialState);
  },
}));
