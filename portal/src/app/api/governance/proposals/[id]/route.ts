/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextRequest, NextResponse } from 'next/server';
import { env } from '@/config/env';
import type {
  GovernanceProposal,
  GovernanceProposalDetailResponse,
  GovernanceVote,
  TallyParams,
  TallyResult,
  VotingParams,
} from '@/types/governance';

export const runtime = 'nodejs';

const RELATED_LIMIT = 30;
const VOTES_LIMIT = 50;

function buildChainUrl(path: string, params?: Record<string, string | number | boolean>) {
  const base = env.chainRest.endsWith('/') ? env.chainRest : `${env.chainRest}/`;
  const url = new URL(path.replace(/^\//, ''), base);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.set(key, String(value));
    });
  }
  return url.toString();
}

async function fetchChainJson<T>(path: string, params?: Record<string, string | number | boolean>) {
  const response = await fetch(buildChainUrl(path, params), { cache: 'no-store' });
  if (!response.ok) {
    throw new Error(`Chain request failed: ${response.status}`);
  }
  return (await response.json()) as T;
}

export async function GET(request: NextRequest, { params }: { params: { id: string } }) {
  try {
    const { searchParams } = request.nextUrl;
    const voter = searchParams.get('voter') ?? undefined;

    const proposalResponse = await fetchChainJson<{ proposal: GovernanceProposal }>(
      `/cosmos/gov/v1/proposals/${params.id}`
    );

    const proposal = proposalResponse.proposal;

    const [tallyResponse, votesResponse, tallyParamsResponse, votingParamsResponse, poolResponse] =
      await Promise.allSettled([
        fetchChainJson<{ tally: TallyResult }>(`/cosmos/gov/v1/proposals/${params.id}/tally`),
        fetchChainJson<{ votes: GovernanceVote[] }>(`/cosmos/gov/v1/proposals/${params.id}/votes`, {
          'pagination.limit': VOTES_LIMIT,
        }),
        fetchChainJson<{ tally_params: TallyParams }>('/cosmos/gov/v1/params/tallying'),
        fetchChainJson<{ voting_params: VotingParams }>('/cosmos/gov/v1/params/voting'),
        fetchChainJson<{ pool: { bonded_tokens: string } }>('/cosmos/staking/v1beta1/pool'),
      ]);

    let voterVote: GovernanceVote | null = null;
    if (voter) {
      try {
        const voteResponse = await fetchChainJson<{ vote: GovernanceVote }>(
          `/cosmos/gov/v1/proposals/${params.id}/votes/${voter}`
        );
        voterVote = voteResponse.vote ?? null;
      } catch {
        voterVote = null;
      }
    }

    let relatedProposals: GovernanceProposal[] | undefined;
    if (proposal?.proposer) {
      try {
        const relatedResponse = await fetchChainJson<{ proposals: GovernanceProposal[] }>(
          '/cosmos/gov/v1/proposals',
          {
            'pagination.limit': RELATED_LIMIT,
            'pagination.reverse': true,
          }
        );
        relatedProposals = (relatedResponse.proposals ?? [])
          .filter((item) => item.proposer === proposal.proposer && item.id !== proposal.id)
          .slice(0, 6);
      } catch {
        relatedProposals = undefined;
      }
    }

    const response: GovernanceProposalDetailResponse = {
      proposal,
      tally: tallyResponse.status === 'fulfilled' ? tallyResponse.value.tally : null,
      votes: votesResponse.status === 'fulfilled' ? votesResponse.value.votes : undefined,
      tallyParams:
        tallyParamsResponse.status === 'fulfilled'
          ? tallyParamsResponse.value.tally_params
          : undefined,
      votingParams:
        votingParamsResponse.status === 'fulfilled'
          ? votingParamsResponse.value.voting_params
          : undefined,
      bondedTokens:
        poolResponse.status === 'fulfilled' ? poolResponse.value.pool?.bonded_tokens : undefined,
      relatedProposals,
      voterVote,
    };

    return NextResponse.json(response);
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to load proposal';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
