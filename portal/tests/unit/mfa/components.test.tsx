/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Unit tests for MFA UI components rendering.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';

// We need to mock the store before importing components
const mockStoreState: Record<string, unknown> = {
  isLoading: false,
  isEnabled: false,
  factors: [] as unknown[],
  policy: null,
  trustedBrowsers: [],
  auditLog: [],
  activeChallenge: null,
  totpEnrollment: null,
  webAuthnEnrollment: null,
  backupCodes: null,
  error: null,
  isMutating: false,
  loadMFAData: vi.fn(),
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
  revokeTrustedBrowser: vi.fn(),
  clearEnrollment: vi.fn(),
  clearBackupCodes: vi.fn(),
  clearError: vi.fn(),
  reset: vi.fn(),
};

vi.mock('../../../src/features/mfa/store', () => ({
  useMFAStore: Object.assign(
    (selector?: (state: typeof mockStoreState) => unknown) => {
      if (selector) return selector(mockStoreState);
      return mockStoreState;
    },
    {
      getState: () => mockStoreState,
      setState: vi.fn(),
      subscribe: vi.fn(),
    }
  ),
}));

vi.mock('../../../src/features/mfa/api', () => ({
  submitRecovery: vi.fn(),
}));

// Import after mock setup
import { FactorList } from '../../../src/components/mfa/FactorList';
import { MFASetup } from '../../../src/components/mfa/MFASetup';

describe('FactorList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockStoreState.factors = [];
    mockStoreState.error = null;
  });

  it('renders empty state when no factors', () => {
    render(<FactorList />);
    expect(screen.getByText('No MFA Factors Enrolled')).toBeDefined();
  });

  it('renders add button when onAddFactor provided and no factors', () => {
    const onAdd = vi.fn();
    render(<FactorList onAddFactor={onAdd} />);
    expect(screen.getByText('Add First Factor')).toBeDefined();
  });

  it('renders enrolled factors', () => {
    mockStoreState.factors = [
      {
        id: 'f-1',
        type: 'otp' as const,
        name: 'Work Phone',
        enrolledAt: Date.now(),
        lastUsedAt: null,
        isPrimary: true,
        status: 'active' as const,
        metadata: {},
      },
    ];

    render(<FactorList />);
    expect(screen.getByText('Work Phone')).toBeDefined();
    expect(screen.getByText('active')).toBeDefined();
    expect(screen.getByText('Primary')).toBeDefined();
  });

  it('shows disable button for active factors', () => {
    mockStoreState.factors = [
      {
        id: 'f-1',
        type: 'otp' as const,
        name: 'Phone',
        enrolledAt: Date.now(),
        lastUsedAt: null,
        isPrimary: false,
        status: 'active' as const,
        metadata: {},
      },
    ];

    render(<FactorList />);
    expect(screen.getByText('Disable')).toBeDefined();
  });

  it('shows enable button for suspended factors', () => {
    mockStoreState.factors = [
      {
        id: 'f-1',
        type: 'otp' as const,
        name: 'Phone',
        enrolledAt: Date.now(),
        lastUsedAt: null,
        isPrimary: false,
        status: 'suspended' as const,
        metadata: {},
      },
    ];

    render(<FactorList />);
    expect(screen.getByText('Enable')).toBeDefined();
  });
});

describe('MFASetup', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders factor selection options', () => {
    render(<MFASetup />);
    expect(screen.getByText('Authenticator App')).toBeDefined();
    expect(screen.getByText('Security Key')).toBeDefined();
    expect(screen.getByText('Backup Codes')).toBeDefined();
  });

  it('renders title and description', () => {
    render(<MFASetup />);
    expect(screen.getByText('Set Up Two-Factor Authentication')).toBeDefined();
    expect(screen.getByText('Add an extra layer of security to your account')).toBeDefined();
  });
});
