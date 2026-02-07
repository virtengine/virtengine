import { describe, expect, it } from 'vitest';

import {
  calculateQuorumProgress,
  calculateTallyPercentages,
  formatProposalStatus,
} from '@/lib/governance';

describe('governance utils', () => {
  it('calculates vote percentages from tally results', () => {
    const tally = {
      yes_count: '600',
      no_count: '200',
      abstain_count: '100',
      no_with_veto_count: '100',
    };

    const result = calculateTallyPercentages(tally);
    expect(result.yes).toBe(60);
    expect(result.no).toBe(20);
    expect(result.abstain).toBe(10);
    expect(result.veto).toBe(10);
  });

  it('calculates quorum progress using bonded tokens', () => {
    const tally = {
      yes_count: '300',
      no_count: '200',
      abstain_count: '0',
      no_with_veto_count: '0',
    };

    const progress = calculateQuorumProgress(tally, '1000');
    expect(progress).toBeCloseTo(0.5, 4);
  });

  it('formats proposal status labels', () => {
    expect(formatProposalStatus('PROPOSAL_STATUS_VOTING_PERIOD')).toBe('Voting');
    expect(formatProposalStatus('PROPOSAL_STATUS_REJECTED')).toBe('Rejected');
  });
});
