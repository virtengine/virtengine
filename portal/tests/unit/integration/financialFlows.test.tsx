import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { EscrowBalance } from '@/components/escrow/EscrowBalance';
import { escrowAccount, fiatRates } from '@/components/escrow/data';
import { formatToken } from '@/components/escrow/utils';
import { VoteTally } from '@/components/governance/VoteTally';
import { InvoiceRow } from '@/components/billing/InvoiceRow';
import type { Invoice } from '@virtengine/portal/types/billing';

const noop = () => undefined;

describe('Financial integration flows', () => {
  it('updates escrow balance after a deposit flow', () => {
    const initialAccount = { ...escrowAccount, availableBalance: 1000 };
    const { rerender } = render(
      <EscrowBalance
        account={initialAccount}
        fiatRates={fiatRates}
        onDeposit={noop}
        onWithdraw={noop}
      />
    );

    expect(screen.getByText(formatToken(1000, initialAccount.currency))).toBeInTheDocument();

    const updatedAccount = { ...initialAccount, availableBalance: 1500 };
    rerender(
      <EscrowBalance
        account={updatedAccount}
        fiatRates={fiatRates}
        onDeposit={noop}
        onWithdraw={noop}
      />
    );

    expect(screen.getByText(formatToken(1500, updatedAccount.currency))).toBeInTheDocument();
  });

  it('reflects governance vote changes in the tally progress', () => {
    const { rerender } = render(
      <VoteTally
        tally={{ yes_count: '50', no_count: '50', abstain_count: '0', no_with_veto_count: '0' }}
        bondedTokens={'100'}
      />
    );

    expect(screen.getByText('Yes 50%')).toBeInTheDocument();

    rerender(
      <VoteTally
        tally={{ yes_count: '80', no_count: '20', abstain_count: '0', no_with_veto_count: '0' }}
        bondedTokens={'100'}
      />
    );

    expect(screen.getByText('Yes 80%')).toBeInTheDocument();
  });

  it('updates invoice status after payment confirmation', () => {
    const invoice: Invoice = {
      id: 'inv-1',
      number: 'INV-001',
      leaseId: 'lease-1',
      deploymentId: 'dep-1',
      provider: 'Provider Alpha',
      period: { start: new Date('2026-01-01'), end: new Date('2026-01-31') },
      status: 'pending',
      currency: 'VIRT',
      subtotal: '10',
      fees: { platformFee: '1', providerFee: '1', networkFee: '1' },
      total: '12',
      dueDate: new Date('2026-02-10'),
      createdAt: new Date('2026-02-01'),
      lineItems: [],
      payments: [],
    };

    const { rerender } = render(
      <table>
        <tbody>
          <InvoiceRow invoice={invoice} />
        </tbody>
      </table>
    );
    expect(screen.getByText('pending')).toBeInTheDocument();

    rerender(
      <table>
        <tbody>
          <InvoiceRow invoice={{ ...invoice, status: 'paid' }} />
        </tbody>
      </table>
    );

    expect(screen.getByText('paid')).toBeInTheDocument();
  });
});
