import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { waitFor } from '@testing-library/react';
import { DepositModal } from '@/components/escrow/DepositModal';
import { escrowAccount, fiatRates } from '@/components/escrow/data';
import type { EscrowMutationAdapter } from '@/components/escrow/mutations';

const mutationContext = {
  chainId: 'virtengine-1',
  accountAddress: 'virtengine1account',
  escrowAccountId: escrowAccount.accountId,
};

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
    expect(screen.getByRole('button', { name: /^deposit$/i })).toBeDisabled();
  });

  it('rejects non-canonical numeric syntax', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '5e2' } });

    expect(screen.getByText('Minimum deposit is 50 VIRT')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^deposit$/i })).toBeDisabled();
  });

  it('shows validation error when exceeding wallet balance', () => {
    renderModal();

    fireEvent.change(screen.getByLabelText('Amount'), {
      target: { value: (escrowAccount.walletBalance + 1).toString() },
    });

    expect(screen.getByText('Amount exceeds available wallet balance')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^deposit$/i })).toBeDisabled();
  });

  it('is unavailable without an adapter and hides unsupported sources', () => {
    renderModal();
    expect(screen.getByText('Wallet deposits unavailable')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /^deposit$/i })).toBeDisabled();
    expect(screen.queryByText(/wire transfer/i)).not.toBeInTheDocument();
    expect(screen.queryByText(/card top-up/i)).not.toBeInTheDocument();
  });

  it('shows success only after exact committed evidence resolves', async () => {
    let resolveMutation: (() => void) | undefined;
    const mutate = vi.fn(
      (request, submission) =>
        new Promise((resolve) => {
          resolveMutation = () =>
            resolve({
              status: 'committed',
              txHash: 'DEPOSIT123',
              code: 0,
              blockHeight: 42,
              operationId: 'deposit-operation-1',
              requestDigest: submission.requestDigest,
              idempotencyKey: submission.idempotencyKey,
              request,
            });
        })
    );
    const adapter: EscrowMutationAdapter = {
      mutate,
    };
    renderModal({ mutationAdapter: adapter, mutationContext, resultProjector: (value) => value });

    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    expect(screen.getByText('Submitting deposit')).toBeInTheDocument();
    expect(screen.queryByText('Deposit committed')).not.toBeInTheDocument();
    expect(screen.getByLabelText('Amount')).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: /submitting/i }));
    await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));

    resolveMutation?.();
    expect(await screen.findByText('Deposit committed')).toBeInTheDocument();
  });

  it('does not retain success after the amount changes', async () => {
    renderModal({
      mutationAdapter: {
        mutate: vi.fn((request, submission) =>
          Promise.resolve({
            status: 'committed',
            txHash: 'DEPOSIT123',
            code: 0,
            blockHeight: 42,
            operationId: 'deposit-operation-1',
            requestDigest: submission.requestDigest,
            idempotencyKey: submission.idempotencyKey,
            request,
          })
        ),
      },
      mutationContext,
      resultProjector: (value) => value,
    });

    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    expect(await screen.findByText('Deposit committed')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('Amount'), { target: { value: '600' } });
    expect(screen.queryByText('Deposit committed')).not.toBeInTheDocument();
  });

  it('ignores a pending result after cancellation', async () => {
    let resolveMutation: (() => void) | undefined;
    let submissionSignal: AbortSignal | undefined;
    const mutate = vi.fn(
      (request, submission) =>
        new Promise((resolve) => {
          submissionSignal = submission.signal;
          resolveMutation = () =>
            resolve({
              status: 'committed',
              txHash: 'STALE123',
              code: 0,
              blockHeight: 42,
              operationId: 'stale-operation',
              requestDigest: submission.requestDigest,
              idempotencyKey: submission.idempotencyKey,
              request,
            });
        })
    );
    const adapter: EscrowMutationAdapter = {
      mutate,
    };
    renderModal({ mutationAdapter: adapter, mutationContext, resultProjector: (value) => value });

    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    await screen.findByText('Submitting deposit');
    await waitFor(() => expect(mutate).toHaveBeenCalledTimes(1));
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(submissionSignal?.aborted).toBe(true);
    resolveMutation?.();

    await waitFor(() => expect(screen.queryByText('Deposit committed')).not.toBeInTheDocument());
  });

  it('does not start a mutation when cancelled during request hashing', async () => {
    let resolveDigest: ((value: ArrayBuffer) => void) | undefined;
    const digestSpy = vi
      .spyOn(globalThis.crypto.subtle, 'digest')
      .mockImplementationOnce(
        () => new Promise<ArrayBuffer>((resolve) => (resolveDigest = resolve))
      );
    const mutate = vi.fn();
    const adapter: EscrowMutationAdapter = { mutate };
    renderModal({ mutationAdapter: adapter, mutationContext, resultProjector: (value) => value });

    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    expect(screen.getByText('Submitting deposit')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    resolveDigest?.(new ArrayBuffer(32));

    await waitFor(() => expect(mutate).not.toHaveBeenCalled());
    digestSpy.mockRestore();
  });

  it('does not start a mutation after an external close during request hashing', async () => {
    let resolveDigest: ((value: ArrayBuffer) => void) | undefined;
    const digestSpy = vi
      .spyOn(globalThis.crypto.subtle, 'digest')
      .mockImplementationOnce(
        () => new Promise<ArrayBuffer>((resolve) => (resolveDigest = resolve))
      );
    const mutate = vi.fn();
    const adapter: EscrowMutationAdapter = { mutate };
    const props = {
      onOpenChange: vi.fn(),
      account: escrowAccount,
      fiatRates,
      mutationAdapter: adapter,
      mutationContext,
      resultProjector: (value: unknown) => value,
    };
    const { rerender } = render(<DepositModal {...props} open />);

    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    rerender(<DepositModal {...props} open={false} />);
    resolveDigest?.(new ArrayBuffer(32));

    await waitFor(() => expect(mutate).not.toHaveBeenCalled());
    digestSpy.mockRestore();
  });

  it('rejects malformed commit evidence', async () => {
    const adapter: EscrowMutationAdapter = {
      mutate: vi.fn().mockResolvedValue({ code: 0 }),
    };
    renderModal({ mutationAdapter: adapter, mutationContext, resultProjector: (value) => value });
    fireEvent.click(screen.getByRole('button', { name: /^deposit$/i }));
    expect(await screen.findByText('Deposit not committed')).toBeInTheDocument();
    expect(screen.queryByText('Deposit committed')).not.toBeInTheDocument();
  });

  it('calls onOpenChange when cancelling', () => {
    const { onOpenChange } = renderModal();

    fireEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
