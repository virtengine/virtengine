/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Unit tests for the MFA Zustand store.
 */

import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest';
import type { MFAPolicy } from '@/lib/portal-adapter';
import { useMFAStore } from '../../../src/features/mfa/store';
import * as mfaApi from '../../../src/features/mfa/api';

// Mock the API module
vi.mock('../../../src/features/mfa/api', () => ({
  fetchFactors: vi.fn(),
  fetchPolicy: vi.fn(),
  fetchTrustedBrowsers: vi.fn(),
  fetchAuditLog: vi.fn(),
  startTOTPEnrollment: vi.fn(),
  verifyTOTPEnrollment: vi.fn(),
  startWebAuthnEnrollment: vi.fn(),
  completeWebAuthnEnrollment: vi.fn(),
  generateBackupCodes: vi.fn(),
  removeFactor: vi.fn(),
  toggleFactor: vi.fn(),
  setPrimaryFactor: vi.fn(),
  createChallenge: vi.fn(),
  verifyChallenge: vi.fn(),
  verifyWebAuthnChallenge: vi.fn(),
  revokeTrustedBrowser: vi.fn(),
  trustCurrentBrowser: vi.fn(),
  submitRecovery: vi.fn(),
}));

const mockFactor = {
  id: 'factor-1',
  type: 'otp' as const,
  name: 'My Phone',
  enrolledAt: Date.now(),
  lastUsedAt: null,
  isPrimary: true,
  status: 'active' as const,
  metadata: { issuer: 'VirtEngine' },
};

const mockPolicy = {
  id: 'policy-1',
  updatedAt: Date.now(),
  requiredFactorTypes: ['otp' as const],
  requiredFactorCount: 1,
  sensitiveTransactions: ['withdrawal' as const],
  allowTrustedBrowsers: true,
  trustedBrowserDurationSeconds: 2592000,
  biometricForLowValue: false,
  lowValueThreshold: '100',
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason?: unknown) => void;
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
}

describe('useMFAStore', () => {
  beforeEach(() => {
    useMFAStore.getState().reset();
    vi.clearAllMocks();
  });

  describe('loadMFAData', () => {
    it('loads factors, policy, browsers, and audit log', async () => {
      (mfaApi.fetchFactors as Mock).mockResolvedValue([mockFactor]);
      (mfaApi.fetchPolicy as Mock).mockResolvedValue(mockPolicy);
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      await useMFAStore.getState().loadMFAData();

      const state = useMFAStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.isEnabled).toBe(true);
      expect(state.factors).toHaveLength(1);
      expect(state.factors[0]!.id).toBe('factor-1');
      expect(state.policy).toEqual(mockPolicy);
      expect(state.policyStatus).toBe('ready');
    });

    it('sets error on failure', async () => {
      (mfaApi.fetchFactors as Mock).mockRejectedValue(new Error('Network error'));
      (mfaApi.fetchPolicy as Mock).mockRejectedValue(new Error('Network error'));
      (mfaApi.fetchTrustedBrowsers as Mock).mockRejectedValue(new Error('Network error'));
      (mfaApi.fetchAuditLog as Mock).mockRejectedValue(new Error('Network error'));

      await useMFAStore.getState().loadMFAData();

      const state = useMFAStore.getState();
      expect(state.isLoading).toBe(false);
      expect(state.error).toBe('Network error');
      expect(state.policyStatus).toBe('error');
      expect(state.policy).toBeNull();
      expect(state.factors).toEqual([]);
    });

    it('fails closed when the policy is malformed', async () => {
      (mfaApi.fetchFactors as Mock).mockResolvedValue([mockFactor]);
      (mfaApi.fetchPolicy as Mock).mockResolvedValue({
        ...mockPolicy,
        sensitiveTransactions: ['unknown_transaction'],
      });
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      await useMFAStore.getState().loadMFAData();

      const state = useMFAStore.getState();
      expect(state.policyStatus).toBe('error');
      expect(state.error).toBe('Invalid MFA policy');
      expect(state.policy).toBeNull();
      expect(state.factors).toEqual([]);
      expect(state.isEnabled).toBe(false);
    });

    it.each([
      ['missing field', ({ lowValueThreshold: _removed, ...rest }) => rest],
      ['extra field', (value) => ({ ...value, unexpected: true })],
      ['non-canonical id', (value) => ({ ...value, id: ' policy-1' })],
      ['unsafe updatedAt', (value) => ({ ...value, updatedAt: Number.MAX_SAFE_INTEGER + 1 })],
      ['unknown factor', (value) => ({ ...value, requiredFactorTypes: ['unknown'] })],
      ['duplicate factor', (value) => ({ ...value, requiredFactorTypes: ['otp', 'otp'] })],
      ['zero factor count', (value) => ({ ...value, requiredFactorCount: 0 })],
      ['excess factor count', (value) => ({ ...value, requiredFactorCount: 2 })],
      ['unknown transaction', (value) => ({ ...value, sensitiveTransactions: ['unknown'] })],
      [
        'duplicate transaction',
        (value) => ({ ...value, sensitiveTransactions: ['withdrawal', 'withdrawal'] }),
      ],
      ['non-boolean trusted flag', (value) => ({ ...value, allowTrustedBrowsers: 1 })],
      ['negative trusted duration', (value) => ({ ...value, trustedBrowserDurationSeconds: -1 })],
      [
        'unsafe trusted duration',
        (value) => ({
          ...value,
          trustedBrowserDurationSeconds: Number.MAX_SAFE_INTEGER + 1,
        }),
      ],
      ['non-boolean biometric flag', (value) => ({ ...value, biometricForLowValue: 'false' })],
      ['numeric threshold', (value) => ({ ...value, lowValueThreshold: 100 })],
      ['non-canonical threshold', (value) => ({ ...value, lowValueThreshold: '0100' })],
    ])('rejects a policy with %s', async (_name, mutate) => {
      (mfaApi.fetchFactors as Mock).mockResolvedValue([mockFactor]);
      (mfaApi.fetchPolicy as Mock).mockResolvedValue(mutate(mockPolicy));
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      await useMFAStore.getState().loadMFAData();

      expect(useMFAStore.getState()).toMatchObject({
        policyStatus: 'error',
        policy: null,
        error: 'Invalid MFA policy',
      });
    });

    it('materializes an immutable policy snapshot', async () => {
      const source: MFAPolicy = {
        ...mockPolicy,
        requiredFactorTypes: [...mockPolicy.requiredFactorTypes],
        sensitiveTransactions: [...mockPolicy.sensitiveTransactions],
      };
      (mfaApi.fetchFactors as Mock).mockResolvedValue([]);
      (mfaApi.fetchPolicy as Mock).mockResolvedValue(source);
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      await useMFAStore.getState().loadMFAData();
      const stored = useMFAStore.getState().policy!;
      source.id = 'changed';
      source.requiredFactorTypes.push('sms');
      source.sensitiveTransactions.length = 0;

      expect(stored).toEqual(mockPolicy);
      expect(Object.isFrozen(stored)).toBe(true);
      expect(Object.isFrozen(stored.requiredFactorTypes)).toBe(true);
      expect(Object.isFrozen(stored.sensitiveTransactions)).toBe(true);
    });

    it('shares concurrent loads and publishes only after all data succeeds', async () => {
      const policyRequest = deferred<typeof mockPolicy>();
      (mfaApi.fetchFactors as Mock).mockResolvedValue([mockFactor]);
      (mfaApi.fetchPolicy as Mock).mockReturnValue(policyRequest.promise);
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      const firstLoad = useMFAStore.getState().loadMFAData();
      const secondLoad = useMFAStore.getState().loadMFAData();

      expect(firstLoad).toBe(secondLoad);
      expect(useMFAStore.getState().policyStatus).toBe('loading');
      expect(useMFAStore.getState().policy).toBeNull();
      expect(useMFAStore.getState().factors).toEqual([]);
      expect(mfaApi.fetchPolicy).toHaveBeenCalledTimes(1);

      policyRequest.resolve(mockPolicy);
      await firstLoad;

      expect(useMFAStore.getState().policyStatus).toBe('ready');
      expect(useMFAStore.getState().policy).toEqual(mockPolicy);
    });

    it('sets isEnabled to false when no active factors', async () => {
      const suspendedFactor = { ...mockFactor, status: 'suspended' as const };
      (mfaApi.fetchFactors as Mock).mockResolvedValue([suspendedFactor]);
      (mfaApi.fetchPolicy as Mock).mockResolvedValue(mockPolicy);
      (mfaApi.fetchTrustedBrowsers as Mock).mockResolvedValue([]);
      (mfaApi.fetchAuditLog as Mock).mockResolvedValue([]);

      await useMFAStore.getState().loadMFAData();

      expect(useMFAStore.getState().isEnabled).toBe(false);
    });
  });

  describe('startTOTPEnrollment', () => {
    it('stores TOTP enrollment data', async () => {
      const totpData = {
        qrCodeDataUrl: 'data:image/png;base64,abc',
        manualEntryKey: 'JBSWY3DPEHPK3PXP',
        issuer: 'VirtEngine',
        accountName: 'user@example.com',
      };
      (mfaApi.startTOTPEnrollment as Mock).mockResolvedValue(totpData);

      await useMFAStore.getState().startTOTPEnrollment();

      const state = useMFAStore.getState();
      expect(state.totpEnrollment).toEqual(totpData);
      expect(state.isMutating).toBe(false);
    });
  });

  describe('verifyTOTPEnrollment', () => {
    it('adds new factor to store on success', async () => {
      const newFactor = { ...mockFactor, id: 'factor-2' };
      (mfaApi.verifyTOTPEnrollment as Mock).mockResolvedValue(newFactor);

      useMFAStore.setState({
        totpEnrollment: { qrCodeDataUrl: '', manualEntryKey: '', issuer: '', accountName: '' },
      });

      const result = await useMFAStore.getState().verifyTOTPEnrollment('123456', 'Test');

      expect(result.id).toBe('factor-2');
      const state = useMFAStore.getState();
      expect(state.factors).toHaveLength(1);
      expect(state.totpEnrollment).toBeNull();
      expect(state.isEnabled).toBe(true);
    });

    it('sets error on failure and re-throws', async () => {
      (mfaApi.verifyTOTPEnrollment as Mock).mockRejectedValue(new Error('Invalid code'));

      await expect(useMFAStore.getState().verifyTOTPEnrollment('999999')).rejects.toThrow(
        'Invalid code'
      );

      expect(useMFAStore.getState().error).toBe('Invalid code');
    });
  });

  describe('removeFactor', () => {
    it('removes factor from store', async () => {
      (mfaApi.removeFactor as Mock).mockResolvedValue(undefined);

      useMFAStore.setState({ factors: [mockFactor], isEnabled: true });

      await useMFAStore.getState().removeFactor('factor-1');

      const state = useMFAStore.getState();
      expect(state.factors).toHaveLength(0);
      expect(state.isEnabled).toBe(false);
    });
  });

  describe('toggleFactor', () => {
    it('updates factor status in store', async () => {
      const updatedFactor = { ...mockFactor, status: 'suspended' as const };
      (mfaApi.toggleFactor as Mock).mockResolvedValue(updatedFactor);

      useMFAStore.setState({ factors: [mockFactor] });

      await useMFAStore.getState().toggleFactor('factor-1', false);

      const state = useMFAStore.getState();
      expect(state.factors[0]!.status).toBe('suspended');
    });
  });

  describe('setPrimaryFactor', () => {
    it('updates isPrimary flags', async () => {
      (mfaApi.setPrimaryFactor as Mock).mockResolvedValue(undefined);

      const factor2 = { ...mockFactor, id: 'factor-2', isPrimary: false };
      useMFAStore.setState({ factors: [mockFactor, factor2] });

      await useMFAStore.getState().setPrimaryFactor('factor-2');

      const state = useMFAStore.getState();
      expect(state.factors.find((f) => f.id === 'factor-1')!.isPrimary).toBe(false);
      expect(state.factors.find((f) => f.id === 'factor-2')!.isPrimary).toBe(true);
    });
  });

  describe('generateBackupCodes', () => {
    it('stores generated backup codes', async () => {
      const codesData = {
        codes: ['AAA-BBB', 'CCC-DDD', 'EEE-FFF'],
        generatedAt: Date.now(),
      };
      (mfaApi.generateBackupCodes as Mock).mockResolvedValue(codesData);

      await useMFAStore.getState().generateBackupCodes();

      expect(useMFAStore.getState().backupCodes).toEqual(codesData);
    });
  });

  describe('challenge ownership', () => {
    it('binds challenge creation to its owner and transaction', async () => {
      (mfaApi.createChallenge as Mock).mockResolvedValue({
        id: 'challenge-1',
        transactionType: 'withdrawal',
      });
      const owner = useMFAStore.getState().acquireGateOwner()!;

      await useMFAStore.getState().createChallenge(owner, 'withdrawal');

      const state = useMFAStore.getState();
      expect(state.ownsGate(owner, 'challenge-1')).toBe(true);
      expect(state.activeChallengeOwnerToken).toBe(owner);
      expect(state.activeChallengeTransactionType).toBe('withdrawal');
    });

    it('rejects cross-owner and mismatched challenge responses', async () => {
      const owner = useMFAStore.getState().acquireGateOwner()!;
      await expect(
        useMFAStore.getState().createChallenge('other-owner', 'withdrawal')
      ).rejects.toThrow('authorization_pending');

      (mfaApi.createChallenge as Mock).mockResolvedValue({
        id: 'challenge-1',
        transactionType: 'delegation_change',
      });
      await expect(useMFAStore.getState().createChallenge(owner, 'withdrawal')).rejects.toThrow(
        'Invalid MFA challenge'
      );
      expect(useMFAStore.getState().gateOwnerToken).toBeNull();
      expect(useMFAStore.getState().activeChallenge).toBeNull();
    });

    it('does not publish a stale response over a new owner', async () => {
      const request = deferred<{ id: string; transactionType: 'withdrawal' }>();
      (mfaApi.createChallenge as Mock).mockReturnValue(request.promise);
      const ownerA = useMFAStore.getState().acquireGateOwner()!;
      const creationA = useMFAStore.getState().createChallenge(ownerA, 'withdrawal');
      useMFAStore.getState().reset();
      const ownerB = useMFAStore.getState().acquireGateOwner()!;

      request.resolve({ id: 'challenge-a', transactionType: 'withdrawal' });
      await expect(creationA).rejects.toThrow('Stale MFA challenge');

      expect(useMFAStore.getState().gateOwnerToken).toBe(ownerB);
      expect(useMFAStore.getState().activeChallenge).toBeNull();
    });

    it('clears the old challenge and owner when replacement creation fails', async () => {
      const owner = useMFAStore.getState().acquireGateOwner()!;
      useMFAStore.setState({
        activeChallenge: { id: 'old-challenge' } as never,
        activeChallengeOwnerToken: owner,
        activeChallengeTransactionType: 'withdrawal',
      });
      (mfaApi.createChallenge as Mock).mockRejectedValue(new Error('challenge failed'));

      await expect(useMFAStore.getState().createChallenge(owner, 'withdrawal')).rejects.toThrow(
        'challenge failed'
      );

      expect(useMFAStore.getState()).toMatchObject({
        activeChallenge: null,
        activeChallengeOwnerToken: null,
        activeChallengeTransactionType: null,
        gateOwnerToken: null,
      });
    });

    it('retains a verified owner-bound challenge until the owner releases it', async () => {
      const owner = useMFAStore.getState().acquireGateOwner()!;
      useMFAStore.setState({
        activeChallenge: { id: 'challenge-1' } as never,
        activeChallengeOwnerToken: owner,
        activeChallengeTransactionType: 'withdrawal',
      });
      (mfaApi.verifyChallenge as Mock).mockResolvedValue({ verified: true });

      await useMFAStore.getState().verifyChallenge('factor-1', '123456');

      expect(useMFAStore.getState().ownsGate(owner, 'challenge-1')).toBe(true);
      useMFAStore.getState().releaseGateOwner(owner);
      expect(useMFAStore.getState().activeChallenge).toBeNull();
    });

    it('rejects a stale OTP success without mutating the replacement challenge', async () => {
      const request = deferred<{ verified: boolean }>();
      (mfaApi.verifyChallenge as Mock).mockReturnValue(request.promise);
      const ownerA = useMFAStore.getState().acquireGateOwner()!;
      useMFAStore.setState({
        activeChallenge: { id: 'challenge-a' } as never,
        activeChallengeOwnerToken: ownerA,
        activeChallengeTransactionType: 'withdrawal',
      });
      const verificationA = useMFAStore.getState().verifyChallenge('factor-1', '123456');

      useMFAStore.getState().releaseGateOwner(ownerA);
      const ownerB = useMFAStore.getState().acquireGateOwner()!;
      const challengeB = { id: 'challenge-b' } as never;
      useMFAStore.setState({
        activeChallenge: challengeB,
        activeChallengeOwnerToken: ownerB,
        activeChallengeTransactionType: 'delegation_change',
        isMutating: true,
        error: 'challenge-b-error',
      });
      const expectedB = useMFAStore.getState();

      request.resolve({ verified: true });
      await expect(verificationA).rejects.toThrow('Stale MFA verification');
      const state = useMFAStore.getState();
      expect(state.activeChallenge).toBe(challengeB);
      expect(state).toMatchObject({
        activeChallengeOwnerToken: expectedB.activeChallengeOwnerToken,
        activeChallengeTransactionType: expectedB.activeChallengeTransactionType,
        challengeGeneration: expectedB.challengeGeneration,
        gateOwnerToken: expectedB.gateOwnerToken,
        isMutating: expectedB.isMutating,
        error: expectedB.error,
      });
    });

    it('rejects a stale OTP failure without overwriting replacement error or loading', async () => {
      const request = deferred<{ verified: boolean }>();
      (mfaApi.verifyChallenge as Mock).mockReturnValue(request.promise);
      const ownerA = useMFAStore.getState().acquireGateOwner()!;
      useMFAStore.setState({
        activeChallenge: { id: 'challenge-a' } as never,
        activeChallengeOwnerToken: ownerA,
        activeChallengeTransactionType: 'withdrawal',
      });
      const verificationA = useMFAStore.getState().verifyChallenge('factor-1', '123456');

      useMFAStore.getState().reset();
      const ownerB = useMFAStore.getState().acquireGateOwner()!;
      const challengeB = { id: 'challenge-b' } as never;
      useMFAStore.setState({
        activeChallenge: challengeB,
        activeChallengeOwnerToken: ownerB,
        activeChallengeTransactionType: 'delegation_change',
        isMutating: true,
        error: 'challenge-b-error',
      });
      const generationB = useMFAStore.getState().challengeGeneration;

      request.reject(new Error('challenge-a-error'));
      await expect(verificationA).rejects.toThrow('Stale MFA verification');
      const state = useMFAStore.getState();
      expect(state.activeChallenge).toBe(challengeB);
      expect(state).toMatchObject({
        activeChallengeOwnerToken: ownerB,
        activeChallengeTransactionType: 'delegation_change',
        challengeGeneration: generationB,
        gateOwnerToken: ownerB,
        isMutating: true,
        error: 'challenge-b-error',
      });
    });

    it('rejects a stale WebAuthn success without mutating the replacement challenge', async () => {
      const request = deferred<{ verified: boolean }>();
      (mfaApi.verifyWebAuthnChallenge as Mock).mockReturnValue(request.promise);
      const ownerA = useMFAStore.getState().acquireGateOwner()!;
      useMFAStore.setState({
        activeChallenge: { id: 'challenge-a' } as never,
        activeChallengeOwnerToken: ownerA,
        activeChallengeTransactionType: 'withdrawal',
      });
      const responseBuffer = new Uint8Array([1]).buffer;
      const credential = {
        id: 'credential-a',
        rawId: responseBuffer,
        type: 'public-key',
        response: {
          authenticatorData: responseBuffer,
          clientDataJSON: responseBuffer,
          signature: responseBuffer,
        },
      } as unknown as PublicKeyCredential;
      const verificationA = useMFAStore.getState().verifyWebAuthnChallenge('factor-1', credential);

      useMFAStore.getState().releaseGateOwner(ownerA);
      const ownerB = useMFAStore.getState().acquireGateOwner()!;
      const challengeB = { id: 'challenge-b' } as never;
      useMFAStore.setState({
        activeChallenge: challengeB,
        activeChallengeOwnerToken: ownerB,
        activeChallengeTransactionType: 'delegation_change',
        isMutating: true,
        error: 'challenge-b-error',
      });
      const generationB = useMFAStore.getState().challengeGeneration;

      request.resolve({ verified: true });
      await expect(verificationA).rejects.toThrow('Stale MFA verification');
      const state = useMFAStore.getState();
      expect(state.activeChallenge).toBe(challengeB);
      expect(state).toMatchObject({
        activeChallengeOwnerToken: ownerB,
        activeChallengeTransactionType: 'delegation_change',
        challengeGeneration: generationB,
        gateOwnerToken: ownerB,
        isMutating: true,
        error: 'challenge-b-error',
      });
    });
  });

  describe('clearEnrollment', () => {
    it('clears enrollment state', () => {
      useMFAStore.setState({
        totpEnrollment: { qrCodeDataUrl: '', manualEntryKey: '', issuer: '', accountName: '' },
        error: 'Some error',
      });

      useMFAStore.getState().clearEnrollment();

      const state = useMFAStore.getState();
      expect(state.totpEnrollment).toBeNull();
      expect(state.webAuthnEnrollment).toBeNull();
      expect(state.error).toBeNull();
    });
  });

  describe('clearBackupCodes', () => {
    it('clears backup codes from memory', () => {
      useMFAStore.setState({
        backupCodes: { codes: ['ABC'], generatedAt: Date.now() },
      });

      useMFAStore.getState().clearBackupCodes();

      expect(useMFAStore.getState().backupCodes).toBeNull();
    });
  });

  describe('revokeTrustedBrowser', () => {
    it('removes browser from store', async () => {
      (mfaApi.revokeTrustedBrowser as Mock).mockResolvedValue(undefined);

      const browser = {
        id: 'browser-1',
        browserName: 'Chrome',
        deviceName: 'Work laptop',
        trustedAt: Date.now(),
        expiresAt: Date.now() + 86400000,
        lastUsedAt: Date.now(),
        fingerprintHash: 'abc123',
      };
      useMFAStore.setState({ trustedBrowsers: [browser] });

      await useMFAStore.getState().revokeTrustedBrowser('browser-1');

      expect(useMFAStore.getState().trustedBrowsers).toHaveLength(0);
    });
  });

  describe('reset', () => {
    it('resets store to initial state', () => {
      useMFAStore.setState({
        isEnabled: true,
        factors: [mockFactor],
        error: 'some error',
      });

      useMFAStore.getState().reset();

      const state = useMFAStore.getState();
      expect(state.isEnabled).toBe(false);
      expect(state.factors).toHaveLength(0);
      expect(state.error).toBeNull();
      expect(state.policyStatus).toBe('idle');
    });
  });
});
