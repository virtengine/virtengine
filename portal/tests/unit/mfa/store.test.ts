/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Unit tests for the MFA Zustand store.
 */

import { describe, it, expect, beforeEach, vi, type Mock } from 'vitest';
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
    });
  });
});
