/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * Billing component unit tests
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { UsageSummaryCard } from '@/components/billing/UsageSummaryCard';

// Mock fetch globally for billing API hooks
beforeEach(() => {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ invoices: [], usages: [] }),
    })
  );
});

describe('UsageSummaryCard', () => {
  it('renders title and value', () => {
    render(<UsageSummaryCard title="Current Period" value="150.00" currency="VIRT" />);
    expect(screen.getByText('Current Period')).toBeInTheDocument();
    expect(screen.getByText('150.00')).toBeInTheDocument();
    expect(screen.getByText('VIRT')).toBeInTheDocument();
  });

  it('renders subtitle when provided', () => {
    render(<UsageSummaryCard title="Projected" value="200" subtitle="5 days remaining" />);
    expect(screen.getByText('5 days remaining')).toBeInTheDocument();
  });

  it('shows skeleton when loading', () => {
    const { container } = render(<UsageSummaryCard title="Loading" value="0" loading />);
    expect(container.querySelector('[aria-hidden="true"]')).toBeInTheDocument();
  });

  it('applies warning border when status is warning', () => {
    const { container } = render(
      <UsageSummaryCard title="Outstanding" value="50" status="warning" />
    );
    const card = container.firstElementChild;
    expect(card?.className).toContain('border-warning');
  });

  it('does not apply warning border by default', () => {
    const { container } = render(<UsageSummaryCard title="Normal" value="100" />);
    const card = container.firstElementChild;
    expect(card?.className).not.toContain('border-warning');
  });
});
