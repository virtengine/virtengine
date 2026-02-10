import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { WithdrawForm } from '@/components/escrow/WithdrawForm';
import { escrowAccount, fiatRates } from '@/components/escrow/data';

const account = { ...escrowAccount, availableBalance: 500 };

describe('WithdrawForm', () => {
  it('submits successfully with a valid amount', () => {
    render(
      <WithdrawForm account={account} fiatRates={fiatRates} fiatOffRampUrl={'https://offramp'} />
    );

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '250' } });
    fireEvent.click(screen.getByRole('button', { name: /request withdrawal/i }));

    expect(screen.getByText('Withdrawal requested')).toBeInTheDocument();
  });

  it('shows validation error when amount exceeds available balance', () => {
    render(<WithdrawForm account={account} fiatRates={fiatRates} />);

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '1000' } });

    expect(screen.getByText('Amount exceeds available balance')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /request withdrawal/i })).toBeDisabled();
  });

  it('shows off-ramp link when configured', () => {
    render(
      <WithdrawForm account={account} fiatRates={fiatRates} fiatOffRampUrl={'https://offramp'} />
    );

    const link = screen.getByRole('link', { name: /open off-ramp/i });
    expect(link).toHaveAttribute('href', 'https://offramp');
  });
});
