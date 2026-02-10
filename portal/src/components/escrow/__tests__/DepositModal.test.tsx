import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DepositModal } from '@/components/escrow/DepositModal';
import { escrowAccount, fiatRates } from '@/components/escrow/data';

function renderModal(props?: Partial<React.ComponentProps<typeof DepositModal>>) {
  const onOpenChange = vi.fn();
  render(
    <DepositModal
      open
      onOpenChange={onOpenChange}
      account={escrowAccount}
      fiatRates={fiatRates}
      {...props}
    />
  );
  return { onOpenChange };
}

describe('DepositModal', () => {
  it('shows validation error for amounts below the minimum', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '10' } });

    expect(screen.getByText('Minimum deposit is 50 VIRT')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue/i })).toBeDisabled();
  });

  it('shows validation error when exceeding wallet balance', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('Amount'), {
      target: { value: (escrowAccount.walletBalance + 1).toString() },
    });

    expect(screen.getByText('Amount exceeds available wallet balance')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /continue/i })).toBeDisabled();
  });

  it('submits successfully with a valid amount', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '500' } });
    fireEvent.click(screen.getByRole('button', { name: /continue/i }));

    expect(screen.getByText('Deposit queued')).toBeInTheDocument();
  });

  it('calls onOpenChange when cancelling', () => {
    const { onOpenChange } = renderModal();

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
