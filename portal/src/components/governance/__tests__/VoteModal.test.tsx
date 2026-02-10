import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { VoteModal } from '@/components/governance/VoteModal';

const sendTransaction = vi.fn();
const estimateFee = vi.fn();
const toast = vi.fn();

vi.mock('@/hooks/use-toast', () => ({
  useToast: () => ({ toast }),
}));

vi.mock('@/lib/portal-adapter', () => ({
  useWallet: () => ({
    status: 'connected',
    accounts: [{ address: 'virtengine1voter' }],
    activeAccountIndex: 0,
    balance: '12345678',
  }),
}));

vi.mock('@/hooks/useWalletTransaction', () => ({
  useWalletTransaction: () => ({
    estimateFee,
    sendTransaction,
    isLoading: false,
  }),
}));

vi.mock('@/config/chains', () => ({
  getChainInfo: () => ({
    stakeCurrency: { coinDenom: 'VE', coinDecimals: 6 },
  }),
}));

describe('VoteModal', () => {
  beforeEach(() => {
    sendTransaction.mockResolvedValue({});
    estimateFee.mockReturnValue({ gas: '220000', amount: [{ amount: '2500', denom: 'uve' }] });
    toast.mockClear();
    sendTransaction.mockClear();
  });

  it.each([
    { label: 'Yes', option: 'VOTE_OPTION_YES' },
    { label: 'No', option: 'VOTE_OPTION_NO' },
    { label: 'Abstain', option: 'VOTE_OPTION_ABSTAIN' },
    { label: 'No with Veto', option: 'VOTE_OPTION_NO_WITH_VETO' },
  ])('submits a $label vote', async ({ label, option }) => {
    const onClose = vi.fn();
    render(<VoteModal proposalId={'101'} open onClose={onClose} />);

    fireEvent.click(screen.getByRole('button', { name: label }));
    fireEvent.click(screen.getByRole('button', { name: /review vote/i }));
    fireEvent.click(screen.getByRole('button', { name: /submit vote/i }));

    await waitFor(() => expect(sendTransaction).toHaveBeenCalled());

    const payload = sendTransaction.mock.calls[0]?.[0];
    expect(payload[0].typeUrl).toBe('/cosmos.gov.v1.MsgVote');
    expect(payload[0].value).toMatchObject({
      proposalId: '101',
      voter: 'virtengine1voter',
      option,
    });
    expect(toast).toHaveBeenCalled();
    expect(onClose).toHaveBeenCalled();
  });
});
