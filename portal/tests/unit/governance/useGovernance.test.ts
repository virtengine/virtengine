import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useGovernanceProposals, useGovernanceProposalDetail } from '@/hooks/useGovernance';

function mockFetch(payload: unknown, ok = true, status = 200) {
  return vi.fn().mockResolvedValue({
    ok,
    status,
    json: async () => payload,
  });
}

describe('useGovernance hooks', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('fetches proposals and tally params', async () => {
    const response = {
      proposals: [
        {
          id: '1',
          status: 'PROPOSAL_STATUS_VOTING_PERIOD',
          title: 'Test Proposal',
          messages: [{ '@type': 'cosmos.gov.v1beta1.TextProposal' }],
        },
      ],
      pagination: { total: '1' },
      tallyParams: { quorum: '0.4' },
      bondedTokens: '1000',
    };

    const fetchMock = mockFetch(response);
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() =>
      useGovernanceProposals({ status: 'PROPOSAL_STATUS_VOTING_PERIOD', limit: 1 })
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.proposals).toHaveLength(1);
    expect(result.current.tallyParams?.quorum).toBe('0.4');
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('includeTally=true'),
      expect.any(Object)
    );
  });

  it('captures errors from proposal fetch', async () => {
    const fetchMock = mockFetch({ message: 'error' }, false, 500);
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useGovernanceProposals());

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.error).toContain('Failed to load proposals');
  });

  it('fetches proposal detail payloads', async () => {
    const response = {
      proposal: {
        id: '9',
        status: 'PROPOSAL_STATUS_VOTING_PERIOD',
        title: 'Detail Proposal',
        messages: [{ '@type': 'cosmos.gov.v1beta1.TextProposal' }],
      },
      tally: { yes_count: '1', no_count: '0', abstain_count: '0', no_with_veto_count: '0' },
    };

    const fetchMock = mockFetch(response);
    vi.stubGlobal('fetch', fetchMock);

    const { result } = renderHook(() => useGovernanceProposalDetail('9'));

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.data?.proposal?.id).toBe('9');
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining('/api/governance/proposals/9'),
      expect.any(Object)
    );
  });
});
