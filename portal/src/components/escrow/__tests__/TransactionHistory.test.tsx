import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { TransactionHistory } from '@/components/escrow/TransactionHistory';
import { escrowTransactions } from '@/components/escrow/data';
import { formatToken } from '@/components/escrow/utils';

function totals() {
  const inflow = escrowTransactions
    .filter((txn) => txn.direction === 'credit')
    .reduce((sum, txn) => sum + txn.amount, 0);
  const outflow = escrowTransactions
    .filter((txn) => txn.direction === 'debit')
    .reduce((sum, txn) => sum + txn.amount, 0);
  return { inflow, outflow };
}

describe('TransactionHistory', () => {
  it('renders inflow/outflow totals and transaction rows', () => {
    const { inflow, outflow } = totals();
    const expectedTotals =
      formatToken(inflow, 'VIRT') + ' in · ' + formatToken(outflow, 'VIRT') + ' out';

    render(<TransactionHistory transactions={escrowTransactions} />);

    expect(screen.getByText('Transaction History')).toBeInTheDocument();
    expect(screen.getByText(expectedTotals)).toBeInTheDocument();
    expect(screen.getAllByRole('row')).toHaveLength(escrowTransactions.length + 1);
  });

  it('shows transaction statuses for escrow state transitions', () => {
    render(<TransactionHistory transactions={escrowTransactions} />);

    expect(screen.getAllByText('completed').length).toBeGreaterThan(0);
    expect(screen.getByText('processing')).toBeInTheDocument();
    expect(screen.getByText('pending')).toBeInTheDocument();
  });
});
