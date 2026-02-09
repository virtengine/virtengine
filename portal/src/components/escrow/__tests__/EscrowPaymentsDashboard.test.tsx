import { describe, it, expect, beforeAll, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EscrowPaymentsDashboard } from '@/components/escrow/EscrowPaymentsDashboard';
import { escrowAccount } from '@/components/escrow/data';

const scrollIntoViewMock = vi.fn();

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
    configurable: true,
    value: scrollIntoViewMock,
  });
});

describe('EscrowPaymentsDashboard', () => {
  it('renders escrow balances and account metadata', () => {
    render(<EscrowPaymentsDashboard />);

    expect(screen.getByText('Escrow & Payments')).toBeInTheDocument();
    expect(
      screen.getByText((content) => content.includes(escrowAccount.accountId))
    ).toBeInTheDocument();
    expect(screen.getByText('Locked in escrow')).toBeInTheDocument();
    expect(screen.getByText('Available balance')).toBeInTheDocument();
    expect(screen.getByText('Pending settlement')).toBeInTheDocument();
  });

  it('opens the deposit modal from the dashboard', () => {
    render(<EscrowPaymentsDashboard />);

    fireEvent.click(screen.getByRole('button', { name: /deposit/i }));
    expect(screen.getByText('Deposit to Escrow')).toBeInTheDocument();
  });

  it('scrolls to the withdraw form when clicking withdraw', () => {
    render(<EscrowPaymentsDashboard />);

    fireEvent.click(screen.getByRole('button', { name: /^withdraw$/i }));
    expect(scrollIntoViewMock).toHaveBeenCalled();
  });

  it('switches between transaction, settlement, and payout views', () => {
    render(<EscrowPaymentsDashboard />);

    expect(screen.getByRole('tab', { name: 'Transactions' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Settlements' })).toBeInTheDocument();
    expect(screen.getByRole('tab', { name: 'Payouts' })).toBeInTheDocument();
  });
});
