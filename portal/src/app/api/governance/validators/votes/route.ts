/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextRequest, NextResponse } from 'next/server';
import { env } from '@/config/env';
import { getChainInfo } from '@/config/chains';
import { convertBech32Prefix } from '@/lib/bech32';
import type {
  GovernanceProposal,
  GovernanceVote,
  ValidatorVoteHistoryResponse,
} from '@/types/governance';

export const runtime = 'nodejs';

const HISTORY_LIMIT = 6;

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
    const validator = searchParams.get('validator');

    if (!validator) {
      return NextResponse.json({ error: 'Missing validator address' }, { status: 400 });
    }

    const chain = getChainInfo();
    const voterAddress = convertBech32Prefix(validator, chain.bech32Config.bech32PrefixAccAddr);

    const proposalsResponse = await fetchChainJson<{ proposals: GovernanceProposal[] }>(
      '/cosmos/gov/v1/proposals',
      {
        'pagination.limit': HISTORY_LIMIT,
        'pagination.reverse': true,
      }
    );

    const proposals = proposalsResponse.proposals ?? [];

    const history = await Promise.all(
      proposals.map(async (proposal) => {
        let vote: GovernanceVote | null = null;
        try {
          const voteResponse = await fetchChainJson<{ vote: GovernanceVote }>(
            `/cosmos/gov/v1/proposals/${proposal.id}/votes/${voterAddress}`
          );
          vote = voteResponse.vote ?? null;
        } catch {
          vote = null;
        }

        return {
          proposalId: proposal.id,
          vote,
          title: proposal.title,
          status: proposal.status,
        };
      })
    );

    const payload: ValidatorVoteHistoryResponse = {
      validator,
      voterAddress,
      proposals: history,
    };

    return NextResponse.json(payload);
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to load validator history';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
