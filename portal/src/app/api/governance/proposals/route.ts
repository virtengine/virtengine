/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextRequest, NextResponse } from 'next/server';
import { env } from '@/config/env';
import type { GovernanceProposalsResponse, TallyParams, TallyResult } from '@/types/governance';

export const runtime = 'nodejs';

const DEFAULT_LIMIT = 12;

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

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = request.nextUrl;
    const status = searchParams.get('status') ?? 'PROPOSAL_STATUS_VOTING_PERIOD';
    const page = Math.max(1, parseInt(searchParams.get('page') ?? '1', 10));
    const limit = Math.max(1, parseInt(searchParams.get('limit') ?? `${DEFAULT_LIMIT}`, 10));
    const includeTally = searchParams.get('includeTally') === 'true';

    const proposalParams: Record<string, string | number | boolean> = {
      'pagination.offset': (page - 1) * limit,
      'pagination.limit': limit,
      'pagination.reverse': true,
    };

    if (status && status !== 'all') {
      proposalParams.proposal_status = status;
    }

    const proposalsResponse = await fetchChainJson<GovernanceProposalsResponse>(
      '/cosmos/gov/v1/proposals',
      proposalParams
    );

    let proposals = proposalsResponse.proposals ?? [];

    if (includeTally && proposals.length > 0) {
      const tallies = await Promise.all(
        proposals.map(async (proposal) => {
          try {
            const tally = await fetchChainJson<{ tally: TallyResult }>(
              `/cosmos/gov/v1/proposals/${proposal.id}/tally`
            );
            return tally.tally;
          } catch {
            return null;
          }
        })
      );

      proposals = proposals.map((proposal, index) => ({
        ...proposal,
        tally: tallies[index] ?? undefined,
      }));
    }

    let tallyParams: TallyParams | undefined;
    let bondedTokens: string | undefined;
    try {
      const tallyResponse = await fetchChainJson<{ tally_params: TallyParams }>(
        '/cosmos/gov/v1/params/tallying'
      );
      tallyParams = tallyResponse.tally_params;
    } catch {
      tallyParams = undefined;
    }

    try {
      const poolResponse = await fetchChainJson<{ pool: { bonded_tokens: string } }>(
        '/cosmos/staking/v1beta1/pool'
      );
      bondedTokens = poolResponse.pool?.bonded_tokens;
    } catch {
      bondedTokens = undefined;
    }

    return NextResponse.json({
      proposals,
      pagination: proposalsResponse.pagination,
      tallyParams,
      bondedTokens,
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to load proposals';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
