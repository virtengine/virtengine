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
import type {
  TOTPEnrollmentData,
  WebAuthnEnrollmentData,
  BackupCodesData,
  SMSEnrollmentData,
  EmailEnrollmentData,
} from './types';
import * as mfaApi from './api';

export type MFAPolicyStatus = 'idle' | 'loading' | 'ready' | 'error';
export type MFAGateOwnerToken = string;

const MFA_POLICY_FIELDS = new Set([
  'id',
  'updatedAt',
  'requiredFactorTypes',
  'requiredFactorCount',
  'sensitiveTransactions',
  'allowTrustedBrowsers',
  'trustedBrowserDurationSeconds',
  'biometricForLowValue',
  'lowValueThreshold',
]);

const MFA_FACTOR_TYPES = new Set(['otp', 'fido2', 'sms', 'biometric', 'email']);

const SENSITIVE_TRANSACTION_TYPES = new Set([
  'account_recovery',
  'key_rotation',
  'high_value_order',
  'provider_registration',
  'offering_creation',
  'hpc_job_submission',
  'withdrawal',
  'delegation_change',
]);

function isCanonicalId(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    value.length > 0 &&
    value === value.trim() &&
    /^[A-Za-z0-9][A-Za-z0-9._:-]*$/.test(value)
  );
}

function isCanonicalNonnegativeDecimal(value: unknown): value is string {
  return typeof value === 'string' && /^(?:0|[1-9]\d*)(?:\.\d+)?$/.test(value);
}

function hasUniqueKnownValues(value: unknown, knownValues: ReadonlySet<string>): value is string[] {
  return (
    Array.isArray(value) &&
    value.every((item) => typeof item === 'string' && knownValues.has(item)) &&
    new Set(value).size === value.length
  );
}

export function materializeMFAPolicy(policy: unknown): MFAPolicy | null {
  if (!policy || typeof policy !== 'object') return null;
  const source = policy as Record<string, unknown>;
  const keys = Object.keys(source);
  if (keys.length !== MFA_POLICY_FIELDS.size || !keys.every((key) => MFA_POLICY_FIELDS.has(key))) {
    return null;
  }
  if (
    !isCanonicalId(source.id) ||
    !Number.isSafeInteger(source.updatedAt) ||
    !hasUniqueKnownValues(source.requiredFactorTypes, MFA_FACTOR_TYPES) ||
    source.requiredFactorTypes.length === 0 ||
    !Number.isSafeInteger(source.requiredFactorCount) ||
    (source.requiredFactorCount as number) <= 0 ||
    (source.requiredFactorCount as number) > source.requiredFactorTypes.length ||
    !hasUniqueKnownValues(source.sensitiveTransactions, SENSITIVE_TRANSACTION_TYPES) ||
    typeof source.allowTrustedBrowsers !== 'boolean' ||
    !Number.isSafeInteger(source.trustedBrowserDurationSeconds) ||
    (source.trustedBrowserDurationSeconds as number) < 0 ||
    typeof source.biometricForLowValue !== 'boolean' ||
    !isCanonicalNonnegativeDecimal(source.lowValueThreshold)
  ) {
    return null;
  }

  return Object.freeze({
    id: source.id,
    updatedAt: source.updatedAt as number,
    requiredFactorTypes: Object.freeze([...source.requiredFactorTypes]),
    requiredFactorCount: source.requiredFactorCount as number,
    sensitiveTransactions: Object.freeze([...source.sensitiveTransactions]),
    allowTrustedBrowsers: source.allowTrustedBrowsers,
    trustedBrowserDurationSeconds: source.trustedBrowserDurationSeconds as number,
    biometricForLowValue: source.biometricForLowValue,
    lowValueThreshold: source.lowValueThreshold,
  }) as MFAPolicy;
}

export function hasValidSensitiveTransactions(policy: unknown): policy is MFAPolicy {
  return materializeMFAPolicy(policy) !== null;
}

export interface MFAStoreState {
  /** Whether initial data is loading */
  isLoading: boolean;
  /** Authoritative policy initialization state */
  policyStatus: MFAPolicyStatus;
  /** Changes whenever authoritative policy state is invalidated or reloaded */
  policyGeneration: number;
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
  /** Hook instance currently authorized to own a challenge */
  gateOwnerToken: MFAGateOwnerToken | null;
  /** Owner bound to the published challenge */
  activeChallengeOwnerToken: MFAGateOwnerToken | null;
  /** Transaction bound to the published challenge */
  activeChallengeTransactionType: Parameters<typeof mfaApi.createChallenge>[0] | null;
  /** Changes whenever challenge creation is invalidated */
  challengeGeneration: number;
  /** In-flight TOTP enrollment data */
  totpEnrollment: TOTPEnrollmentData | null;
  /** In-flight WebAuthn enrollment data */
  webAuthnEnrollment: WebAuthnEnrollmentData | null;
  /** In-flight SMS enrollment data */
  smsEnrollment: SMSEnrollmentData | null;
  /** In-flight email enrollment data */
  emailEnrollment: EmailEnrollmentData | null;
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
  /** Start SMS enrollment */
  startSMSEnrollment: (phone: string) => Promise<void>;
  /** Verify SMS enrollment */
  verifySMSEnrollment: (code: string, name?: string) => Promise<MFAFactor>;
  /** Start email enrollment */
  startEmailEnrollment: (email: string) => Promise<void>;
  /** Verify email enrollment */
  verifyEmailEnrollment: (code: string, name?: string) => Promise<MFAFactor>;
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
    ownerToken: MFAGateOwnerToken,
    transactionType: Parameters<typeof mfaApi.createChallenge>[0]
  ) => Promise<MFAChallenge>;
  /** Acquire exclusive ownership of the global MFA gate */
  acquireGateOwner: () => MFAGateOwnerToken | null;
  /** Whether a token owns the gate and, optionally, its exact challenge */
  ownsGate: (ownerToken: MFAGateOwnerToken, challengeId?: string) => boolean;
  /** Release the gate only when called by its current owner */
  releaseGateOwner: (ownerToken: MFAGateOwnerToken) => void;
  /** Verify an MFA challenge with code */
  verifyChallenge: (factorId: string, code: string) => Promise<boolean>;
  /** Verify a WebAuthn MFA challenge */
  verifyWebAuthnChallenge: (factorId: string, credential: PublicKeyCredential) => Promise<boolean>;
  /** Revoke a trusted browser */
  revokeTrustedBrowser: (browserId: string) => Promise<void>;
  /** Clear enrollment state */
  clearEnrollment: () => void;
  /** Clear active challenge */
  clearChallenge: (ownerToken?: MFAGateOwnerToken) => void;
  /** Clear backup codes from memory */
  clearBackupCodes: () => void;
  /** Clear error */
  clearError: () => void;
  /** Reset store */
  reset: () => void;
}

const initialState: MFAStoreState = {
  isLoading: false,
  policyStatus: 'idle',
  policyGeneration: 0,
  isEnabled: false,
  factors: [],
  policy: null,
  trustedBrowsers: [],
  auditLog: [],
  activeChallenge: null,
  gateOwnerToken: null,
  activeChallengeOwnerToken: null,
  activeChallengeTransactionType: null,
  challengeGeneration: 0,
  totpEnrollment: null,
  webAuthnEnrollment: null,
  smsEnrollment: null,
  emailEnrollment: null,
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

interface MFAVerificationOwnership {
  challengeId: string;
  challengeOwnerToken: MFAGateOwnerToken | null;
  transactionType: Parameters<typeof mfaApi.createChallenge>[0] | null;
  challengeGeneration: number;
  gateOwnerToken: MFAGateOwnerToken | null;
}

function hasVerificationOwnership(
  state: MFAStoreState,
  ownership: MFAVerificationOwnership
): boolean {
  return (
    state.activeChallenge?.id === ownership.challengeId &&
    state.activeChallengeOwnerToken === ownership.challengeOwnerToken &&
    state.activeChallengeTransactionType === ownership.transactionType &&
    state.challengeGeneration === ownership.challengeGeneration &&
    state.gateOwnerToken === ownership.gateOwnerToken
  );
}

let inFlightMFALoad: Promise<void> | null = null;
let nextGateOwnerToken = 0;

export const useMFAStore = create<MFAStoreState & MFAStoreActions>((set, get) => ({
  ...initialState,

  loadMFAData: () => {
    if (inFlightMFALoad) return inFlightMFALoad;

    const generation = get().policyGeneration + 1;
    set({
      isLoading: true,
      policyStatus: 'loading',
      policyGeneration: generation,
      isEnabled: false,
      factors: [],
      policy: null,
      trustedBrowsers: [],
      auditLog: [],
      activeChallenge: null,
      gateOwnerToken: null,
      activeChallengeOwnerToken: null,
      activeChallengeTransactionType: null,
      challengeGeneration: get().challengeGeneration + 1,
      error: null,
    });

    const load = (async () => {
      try {
        const [factors, policy, trustedBrowsers, auditLog] = await Promise.all([
          mfaApi.fetchFactors(),
          mfaApi.fetchPolicy(),
          mfaApi.fetchTrustedBrowsers(),
          mfaApi.fetchAuditLog(),
        ]);
        const materializedPolicy = materializeMFAPolicy(policy);
        if (!materializedPolicy) {
          throw new Error('Invalid MFA policy');
        }
        if (get().policyGeneration !== generation) return;
        set({
          isLoading: false,
          policyStatus: 'ready',
          isEnabled: factors.some((factor) => factor.status === 'active'),
          factors,
          policy: materializedPolicy,
          trustedBrowsers,
          auditLog,
        });
      } catch (err) {
        if (get().policyGeneration !== generation) return;
        set({
          isLoading: false,
          policyStatus: 'error',
          isEnabled: false,
          factors: [],
          policy: null,
          trustedBrowsers: [],
          auditLog: [],
          error: err instanceof Error ? err.message : 'Failed to load MFA data',
        });
      }
    })();

    inFlightMFALoad = load;
    void load.finally(() => {
      if (inFlightMFALoad === load) inFlightMFALoad = null;
    });
    return load;
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

  startSMSEnrollment: async (phone) => {
    set({ isMutating: true, error: null, smsEnrollment: null });
    try {
      const data = await mfaApi.startSMSEnrollment(phone);
      set({ isMutating: false, smsEnrollment: data });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to start SMS enrollment',
      });
      throw err;
    }
  },

  verifySMSEnrollment: async (code, name) => {
    const enrollment = get().smsEnrollment;
    if (!enrollment) throw new Error('No SMS enrollment in progress');

    set({ isMutating: true, error: null });
    try {
      const factor = await mfaApi.verifySMSEnrollment(enrollment.challengeId, code, name);
      set((s) => ({
        isMutating: false,
        smsEnrollment: null,
        factors: [...s.factors, factor],
        isEnabled: true,
      }));
      return factor;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'SMS verification failed',
      });
      throw err;
    }
  },

  startEmailEnrollment: async (email) => {
    set({ isMutating: true, error: null, emailEnrollment: null });
    try {
      const data = await mfaApi.startEmailEnrollment(email);
      set({ isMutating: false, emailEnrollment: data });
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Failed to start email enrollment',
      });
      throw err;
    }
  },

  verifyEmailEnrollment: async (code, name) => {
    const enrollment = get().emailEnrollment;
    if (!enrollment) throw new Error('No email enrollment in progress');

    set({ isMutating: true, error: null });
    try {
      const factor = await mfaApi.verifyEmailEnrollment(enrollment.challengeId, code, name);
      set((s) => ({
        isMutating: false,
        emailEnrollment: null,
        factors: [...s.factors, factor],
        isEnabled: true,
      }));
      return factor;
    } catch (err) {
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Email verification failed',
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

  acquireGateOwner: () => {
    if (get().gateOwnerToken !== null) return null;
    const ownerToken = `mfa-gate-${++nextGateOwnerToken}`;
    set({ gateOwnerToken: ownerToken });
    return ownerToken;
  },

  ownsGate: (ownerToken, challengeId) => {
    const state = get();
    return (
      state.gateOwnerToken === ownerToken &&
      (challengeId === undefined ||
        (state.activeChallengeOwnerToken === ownerToken &&
          state.activeChallenge?.id === challengeId))
    );
  },

  releaseGateOwner: (ownerToken) => {
    if (get().gateOwnerToken !== ownerToken) return;
    set((state) => ({
      gateOwnerToken: null,
      activeChallenge: null,
      activeChallengeOwnerToken: null,
      activeChallengeTransactionType: null,
      challengeGeneration: state.challengeGeneration + 1,
    }));
  },

  createChallenge: async (ownerToken, transactionType) => {
    if (get().gateOwnerToken !== ownerToken) throw new Error('authorization_pending');
    const generation = get().challengeGeneration + 1;
    set({
      isMutating: true,
      error: null,
      activeChallenge: null,
      activeChallengeOwnerToken: null,
      activeChallengeTransactionType: null,
      challengeGeneration: generation,
    });
    try {
      const challenge = await mfaApi.createChallenge(transactionType);
      const returnedType = (challenge as { transactionType?: unknown }).transactionType;
      if (
        !isCanonicalId(challenge?.id) ||
        (returnedType !== undefined && returnedType !== transactionType)
      ) {
        throw new Error('Invalid MFA challenge');
      }
      const current = get();
      if (current.gateOwnerToken !== ownerToken || current.challengeGeneration !== generation) {
        throw new Error('Stale MFA challenge');
      }
      set({
        isMutating: false,
        activeChallenge: challenge,
        activeChallengeOwnerToken: ownerToken,
        activeChallengeTransactionType: transactionType,
      });
      return challenge;
    } catch (err) {
      if (get().gateOwnerToken === ownerToken && get().challengeGeneration === generation) {
        set((state) => ({
          isMutating: false,
          error: err instanceof Error ? err.message : 'Failed to create challenge',
          activeChallenge: null,
          gateOwnerToken: null,
          activeChallengeOwnerToken: null,
          activeChallengeTransactionType: null,
          challengeGeneration: state.challengeGeneration + 1,
        }));
      }
      throw err;
    }
  },

  verifyChallenge: async (factorId, code) => {
    const initial = get();
    const challenge = initial.activeChallenge;
    if (!challenge) throw new Error('No active challenge');
    const ownership: MFAVerificationOwnership = {
      challengeId: challenge.id,
      challengeOwnerToken: initial.activeChallengeOwnerToken,
      transactionType: initial.activeChallengeTransactionType,
      challengeGeneration: initial.challengeGeneration,
      gateOwnerToken: initial.gateOwnerToken,
    };

    set({ isMutating: true, error: null });
    try {
      const result = await mfaApi.verifyChallenge(ownership.challengeId, factorId, code);
      if (!hasVerificationOwnership(get(), ownership)) {
        throw new Error('Stale MFA verification');
      }
      if (result.verified) {
        set((state) => ({
          isMutating: false,
          activeChallenge: state.gateOwnerToken ? state.activeChallenge : null,
        }));
      } else {
        set({ isMutating: false });
      }
      return result.verified;
    } catch (err) {
      if (!hasVerificationOwnership(get(), ownership)) {
        throw new Error('Stale MFA verification');
      }
      set({
        isMutating: false,
        error: err instanceof Error ? err.message : 'Verification failed',
      });
      throw err;
    }
  },

  verifyWebAuthnChallenge: async (factorId, credential) => {
    const initial = get();
    const challenge = initial.activeChallenge;
    if (!challenge) throw new Error('No active challenge');
    const ownership: MFAVerificationOwnership = {
      challengeId: challenge.id,
      challengeOwnerToken: initial.activeChallengeOwnerToken,
      transactionType: initial.activeChallengeTransactionType,
      challengeGeneration: initial.challengeGeneration,
      gateOwnerToken: initial.gateOwnerToken,
    };

    set({ isMutating: true, error: null });
    try {
      const assertionResponse = credential.response as AuthenticatorAssertionResponse;
      const result = await mfaApi.verifyWebAuthnChallenge(ownership.challengeId, factorId, {
        id: credential.id,
        rawId: arrayBufferToBase64(credential.rawId),
        type: credential.type,
        response: {
          authenticatorData: arrayBufferToBase64(assertionResponse.authenticatorData),
          clientDataJSON: arrayBufferToBase64(assertionResponse.clientDataJSON),
          signature: arrayBufferToBase64(assertionResponse.signature),
        },
      });
      if (!hasVerificationOwnership(get(), ownership)) {
        throw new Error('Stale MFA verification');
      }
      if (result.verified) {
        set((state) => ({
          isMutating: false,
          activeChallenge: state.gateOwnerToken ? state.activeChallenge : null,
        }));
      } else {
        set({ isMutating: false });
      }
      return result.verified;
    } catch (err) {
      if (!hasVerificationOwnership(get(), ownership)) {
        throw new Error('Stale MFA verification');
      }
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
    set({
      totpEnrollment: null,
      webAuthnEnrollment: null,
      smsEnrollment: null,
      emailEnrollment: null,
      error: null,
    });
  },

  clearChallenge: (ownerToken) => {
    if (
      (ownerToken && get().gateOwnerToken !== ownerToken) ||
      (!ownerToken && get().gateOwnerToken !== null)
    ) {
      return;
    }
    set((state) => ({
      activeChallenge: null,
      activeChallengeOwnerToken: null,
      activeChallengeTransactionType: null,
      challengeGeneration: state.challengeGeneration + 1,
    }));
  },

  clearBackupCodes: () => {
    set({ backupCodes: null });
  },

  clearError: () => {
    set({ error: null });
  },

  reset: () => {
    inFlightMFALoad = null;
    set((state) => ({
      ...initialState,
      policyGeneration: state.policyGeneration + 1,
      challengeGeneration: state.challengeGeneration + 1,
    }));
  },
}));
