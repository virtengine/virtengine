import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import React from 'react';
import { ProposalDetail } from '@/components/governance/ProposalDetail';
import type {
  GovernanceProposal,
  TallyParams,
  TallyResult,
  GovernanceVote,
} from '@/types/governance';

vi.mock('next/link', () => ({
  default: ({
    href,
    children,
    ...props
  }: {
    href: string;
    children: React.ReactNode;
    [key: string]: unknown;
  }) => React.createElement('a', { href, ...props }, children),
}));

vi.mock('@/components/governance/VoteModal', () => ({
  VoteModal: () => React.createElement('div', { 'data-testid': 'vote-modal' }),
}));

const proposal: GovernanceProposal = {
  id: '101',
  metadata: JSON.stringify({
    title: 'Increase validator rewards',
    summary: 'Adjust staking rewards by 2%',
    description: 'Proposal to update validator rewards by 2%.',
    forum: 'https://forum.example/proposal/101',
  }),
  status: 'PROPOSAL_STATUS_VOTING_PERIOD',
  proposer: 'virtengine1proposal',
  submit_time: '2026-02-01T12:00:00Z',
  voting_end_time: '2026-02-12T18:00:00Z',
  messages: [{ '@type': 'cosmos.gov.v1beta1.TextProposal' }],
};

const tally: TallyResult = {
  yes_count: '600',
  no_count: '200',
  abstain_count: '100',
  no_with_veto_count: '100',
};

const tallyParams: TallyParams = {
  quorum: '0.4',
  threshold: '0.5',
  veto_threshold: '0.334',
};

const votes: GovernanceVote[] = [
  {
    voter: 'virtengine1voter',
    options: [{ option: 'VOTE_OPTION_YES', weight: '1.0' }],
  },
];

describe('ProposalDetail', () => {
  it('renders proposal content and current vote summary', () => {
    render(
      <ProposalDetail
        proposal={proposal}
        tally={tally}
        votes={votes}
        tallyParams={tallyParams}
        bondedTokens={'10000'}
        voterVote={votes[0]}
      />
    );

    expect(screen.getByText('Increase validator rewards')).toBeInTheDocument();
    expect(screen.getByText('Voting')).toBeInTheDocument();
    expect(screen.getByText('Vote on proposal')).toBeInTheDocument();
    expect(screen.getByText('Yes 60%')).toBeInTheDocument();
    expect(screen.getByText('Quorum target 40%')).toBeInTheDocument();
    expect(screen.getByText(/Your current vote:/)).toBeInTheDocument();
  });

  it('shows voting closed when proposal is not in voting period', () => {
    render(
      <ProposalDetail
        proposal={{ ...proposal, status: 'PROPOSAL_STATUS_PASSED' }}
        tally={tally}
        tallyParams={tallyParams}
      />
    );

    expect(screen.getByText('Voting is closed for this proposal.')).toBeInTheDocument();
  });
});
