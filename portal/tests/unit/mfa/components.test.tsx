/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Unit tests for MFA UI components rendering.
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';

// We need to mock the store before importing components
const mockStoreState: Record<string, unknown> = {
  factors: [] as unknown[],
  trustedBrowsers: [],
  auditLog: [],
  revokeTrustedBrowser: vi.fn(),
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

// Import after mock setup
import { MFAChallenge } from '../../../src/components/mfa/MFAChallenge';
import { MFASettings } from '../../../src/components/mfa/MFASettings';

describe('MFAChallenge', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    window.localStorage.clear();
  });

  it('renders trust this browser checkbox', () => {
    render(
      <MFAChallenge
        open={true}
        onOpenChange={vi.fn()}
        transactionType="high_value_order"
        actionDescription="Send tokens"
      />
    );
    const checkbox = screen.getByLabelText('Trust this browser');
    expect(checkbox).toBeDefined();
    expect(checkbox).toHaveProperty('type', 'checkbox');
    expect(checkbox).toHaveProperty('checked', false);
  });

  it('sets localStorage when trust browser checkbox is toggled', () => {
    const setItemSpy = vi.spyOn(window.localStorage, 'setItem');

    render(
      <MFAChallenge
        open={true}
        onOpenChange={vi.fn()}
        transactionType="high_value_order"
        actionDescription="Send tokens"
      />
    );

    const checkbox = screen.getByLabelText('Trust this browser') as HTMLInputElement;

    // Check the checkbox
    fireEvent.click(checkbox);
    expect(setItemSpy).toHaveBeenCalledWith('mfa_trust_browser', 'true');

    // Uncheck the checkbox
    fireEvent.click(checkbox);
    expect(setItemSpy).toHaveBeenCalledWith('mfa_trust_browser', 'false');

    setItemSpy.mockRestore();
  });
});

describe('MFASettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockStoreState.trustedBrowsers = [];
  });

  it('revokes all trusted browsers via individual revoke calls', async () => {
    const browser1 = {
      id: 'tb-1',
      deviceName: 'Chrome on Mac',
      browserName: 'Chrome',
      trustedAt: Date.now(),
      expiresAt: Date.now() + 86400000,
      region: 'US',
    };
    const browser2 = {
      id: 'tb-2',
      deviceName: 'Firefox on Windows',
      browserName: 'Firefox',
      trustedAt: Date.now(),
      expiresAt: Date.now() + 86400000,
      region: 'EU',
    };
    const browser3 = {
      id: 'tb-3',
      deviceName: 'Safari on iPhone',
      browserName: 'Safari',
      trustedAt: Date.now(),
      expiresAt: Date.now() + 86400000,
      region: 'US',
    };
    mockStoreState.trustedBrowsers = [browser1, browser2, browser3];

    render(<MFASettings />);

    const trustedTab = screen.getByRole('tab', { name: /trusted \(3\)/i });

    // Trigger full click sequence for Radix UI
    fireEvent.pointerDown(trustedTab);
    fireEvent.mouseDown(trustedTab);
    fireEvent.mouseUp(trustedTab);
    fireEvent.click(trustedTab);

    const revokeButtons = await waitFor(() => screen.getAllByRole('button', { name: /^revoke$/i }));

    // Click each Revoke button to revoke all browsers individually
    expect(revokeButtons).toHaveLength(3);
    fireEvent.click(revokeButtons[0]);
    fireEvent.click(revokeButtons[1]);
    fireEvent.click(revokeButtons[2]);

    // Assert revokeTrustedBrowser was called for each browser
    expect(mockStoreState.revokeTrustedBrowser).toHaveBeenCalledWith('tb-1');
    expect(mockStoreState.revokeTrustedBrowser).toHaveBeenCalledWith('tb-2');
    expect(mockStoreState.revokeTrustedBrowser).toHaveBeenCalledWith('tb-3');
    expect(mockStoreState.revokeTrustedBrowser).toHaveBeenCalledTimes(3);
  });
});
