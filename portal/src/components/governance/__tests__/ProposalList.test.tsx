import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ProposalList } from '@/components/governance/ProposalList';
import type { GovernanceProposalWithTally, TallyParams } from '@/types/governance';

const proposals: GovernanceProposalWithTally[] = [
  {
    id: '42',
    title: 'Upgrade Network v2',
    summary: 'Enable the v2 upgrade',
    status: 'PROPOSAL_STATUS_VOTING_PERIOD',
    messages: [{ '@type': 'cosmos.upgrade.v1beta1.SoftwareUpgradeProposal' }],
    voting_end_time: '2026-02-12T18:00:00Z',
    tally: {
      yes_count: '600',
      no_count: '200',
      abstain_count: '100',
      no_with_veto_count: '100',
    },
  },
  {
    id: '43',
    title: 'Community Spend',
    summary: 'Allocate treasury for grants',
    status: 'PROPOSAL_STATUS_DEPOSIT_PERIOD',
    messages: [{ '@type': 'cosmos.distribution.v1beta1.MsgCommunityPoolSpend' }],
    voting_end_time: '2026-02-20T12:00:00Z',
    tally: {
      yes_count: '100',
      no_count: '50',
      abstain_count: '10',
      no_with_veto_count: '0',
    },
  },
];

const tallyParams: TallyParams = {
  quorum: '0.4',
  threshold: '0.5',
  veto_threshold: '0.334',
};

describe('ProposalList', () => {
  it('renders proposal cards with status and type badges', () => {
    render(
      <ProposalList
        proposals={proposals}
        tallyParams={tallyParams}
        bondedTokens={'10000'}
        isLoading={false}
        isLoadingMore={false}
        error={null}
        hasMore={false}
        onLoadMore={() => undefined}
        searchQuery=""
        typeFilter="all"
      />
    );

    expect(screen.getByText('Upgrade Network v2')).toBeInTheDocument();
    expect(screen.getByText('Community Spend')).toBeInTheDocument();
    expect(screen.getByText('Voting')).toBeInTheDocument();
    expect(screen.getByText('Software Upgrade')).toBeInTheDocument();
  });

  it('filters proposals by search query and type', () => {
    render(
      <ProposalList
        proposals={proposals}
        tallyParams={tallyParams}
        bondedTokens={'10000'}
        isLoading={false}
        isLoadingMore={false}
        error={null}
        hasMore={false}
        onLoadMore={() => undefined}
        searchQuery="upgrade"
        typeFilter="Software Upgrade"
      />
    );

    expect(screen.getByText('Upgrade Network v2')).toBeInTheDocument();
    expect(screen.queryByText('Community Spend')).not.toBeInTheDocument();
  });

  it('invokes load more when requested', () => {
    const onLoadMore = vi.fn();

    render(
      <ProposalList
        proposals={proposals}
        tallyParams={tallyParams}
        bondedTokens={'10000'}
        isLoading={false}
        isLoadingMore={false}
        error={null}
        hasMore
        onLoadMore={onLoadMore}
        searchQuery=""
        typeFilter="all"
      />
    );

    fireEvent.click(screen.getByRole('button', { name: /load more proposals/i }));
    expect(onLoadMore).toHaveBeenCalledTimes(1);
  });
});
