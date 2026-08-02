import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { waitFor } from '@testing-library/react';
import { WithdrawForm } from '@/components/escrow/WithdrawForm';
import { escrowAccount, fiatRates } from '@/components/escrow/data';
import type { EscrowMutationAdapter } from '@/components/escrow/mutations';

const account = { ...escrowAccount, availableBalance: 500 };
const mutationContext = {
  chainId: 'virtengine-1',
  accountAddress: 'virtengine1account',
  escrowAccountId: account.accountId,
};

function committedAdapter(): EscrowMutationAdapter {
  return {
    mutate: vi.fn((request, submission) =>
      Promise.resolve({
        status: 'committed',
        txHash: 'WITHDRAW123',
        code: 0,
        blockHeight: 44,
        operationId: 'withdraw-operation-1',
        requestDigest: submission.requestDigest,
        idempotencyKey: submission.idempotencyKey,
        request,
      })
    ),
  };
}

describe('WithdrawForm', () => {
  it('is unavailable without an adapter and hides unsupported destinations', () => {
    render(<WithdrawForm account={account} fiatRates={fiatRates} />);
    expect(screen.getByText('Wallet withdrawals unavailable')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /withdraw to wallet/i })).toBeDisabled();
    expect(screen.queryByText(/bank account/i)).not.toBeInTheDocument();
    expect(screen.queryByRole('option')).not.toBeInTheDocument();
    expect(screen.getByText(/unavailable until checkpoint 89C/i)).toBeInTheDocument();
  });

  it('renders success only for a validated committed withdrawal', async () => {
    render(
      <WithdrawForm
        account={account}
        fiatRates={fiatRates}
        mutationAdapter={committedAdapter()}
        mutationContext={mutationContext}
        resultProjector={(value) => value}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /withdraw to wallet/i }));
    expect(await screen.findByText('Withdrawal committed')).toBeInTheDocument();
    expect(screen.getByText(/WITHDRAW123/)).toBeInTheDocument();
  });

  it('clears committed evidence when the amount changes', async () => {
    render(
      <WithdrawForm
        account={account}
        fiatRates={fiatRates}
        mutationAdapter={committedAdapter()}
        mutationContext={mutationContext}
        resultProjector={(value) => value}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /withdraw to wallet/i }));
    expect(await screen.findByText('Withdrawal committed')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '100' } });
    expect(screen.queryByText('Withdrawal committed')).not.toBeInTheDocument();
  });

  it('follows authority availability changes', () => {
    const { rerender } = render(<WithdrawForm account={account} fiatRates={fiatRates} />);
    expect(screen.getByRole('button', { name: /withdraw to wallet/i })).toBeDisabled();

    rerender(
      <WithdrawForm
        account={account}
        fiatRates={fiatRates}
        mutationAdapter={committedAdapter()}
        mutationContext={mutationContext}
        resultProjector={(value) => value}
      />
    );
    expect(screen.getByRole('button', { name: /withdraw to wallet/i })).toBeEnabled();
  });

  it('does not show success before commit evidence resolves', async () => {
    let resolveMutation: (() => void) | undefined;
    const mutate = vi.fn(
      (request, submission) =>
        new Promise((resolve) => {
          resolveMutation = () =>
            resolve({
              status: 'committed',
              txHash: 'WITHDRAW123',
              code: 0,
              blockHeight: 44,
              operationId: 'withdraw-operation-1',
              requestDigest: submission.requestDigest,
              idempotencyKey: submission.idempotencyKey,
              request,
            });
        })
    );
    const adapter: EscrowMutationAdapter = {
      mutate,
    };
    render(
      <WithdrawForm
        account={account}
        fiatRates={fiatRates}
        mutationAdapter={adapter}
        mutationContext={mutationContext}
        resultProjector={(value) => value}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /withdraw to wallet/i }));
    expect(screen.getByText('Submitting withdrawal')).toBeInTheDocument();
    expect(screen.queryByText('Withdrawal committed')).not.toBeInTheDocument();
    await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
    resolveMutation?.();
    expect(await screen.findByText('Withdrawal committed')).toBeInTheDocument();
  });

  it('does not start a mutation after unmounting during request hashing', async () => {
    let resolveDigest: ((value: ArrayBuffer) => void) | undefined;
    const digestSpy = vi
      .spyOn(globalThis.crypto.subtle, 'digest')
      .mockImplementationOnce(
        () => new Promise<ArrayBuffer>((resolve) => (resolveDigest = resolve))
      );
    const mutate = vi.fn();
    const adapter: EscrowMutationAdapter = { mutate };
    const { unmount } = render(
      <WithdrawForm
        account={account}
        fiatRates={fiatRates}
        mutationAdapter={adapter}
        mutationContext={mutationContext}
        resultProjector={(value) => value}
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /withdraw to wallet/i }));
    unmount();
    resolveDigest?.(new ArrayBuffer(32));

    await waitFor(() => expect(mutate).not.toHaveBeenCalled());
    digestSpy.mockRestore();
  });

  it('shows validation error when amount exceeds available balance', () => {
    render(<WithdrawForm account={account} fiatRates={fiatRates} />);

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '1000' } });

    expect(screen.getByText('Amount exceeds available balance')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /withdraw to wallet/i })).toBeDisabled();
  });

  it('rejects non-canonical numeric syntax', () => {
    render(<WithdrawForm account={account} fiatRates={fiatRates} />);

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '1e2' } });

    expect(screen.getByText('Enter a valid amount')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /withdraw to wallet/i })).toBeDisabled();
  });

  it('keeps the off-ramp as navigation without claiming a request', () => {
    render(
      <WithdrawForm account={account} fiatRates={fiatRates} fiatOffRampUrl={'https://offramp'} />
    );

    const link = screen.getByRole('link', { name: /open off-ramp/i });
    expect(link).toHaveAttribute('href', 'https://offramp');
    expect(screen.queryByText(/^Withdrawal requested$/i)).not.toBeInTheDocument();
  });
});
