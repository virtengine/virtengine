/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

import { NextRequest, NextResponse } from 'next/server';
import { env } from '@/config/env';
import type {
  GovernanceDelegationsResponse,
  GovernanceDelegation,
  GovernanceValidator,
} from '@/types/governance';

export const runtime = 'nodejs';

const VALIDATOR_LIMIT = 200;

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
    const address = searchParams.get('address');

    if (!address) {
      return NextResponse.json({ error: 'Missing address' }, { status: 400 });
    }

    const [delegationsResponse, validatorsResponse] = await Promise.all([
      fetchChainJson<{ delegation_responses: GovernanceDelegation[] }>(
        `/cosmos/staking/v1beta1/delegations/${address}`
      ),
      fetchChainJson<{ validators: GovernanceValidator[] }>('/cosmos/staking/v1beta1/validators', {
        'pagination.limit': VALIDATOR_LIMIT,
      }),
    ]);

    const payload: GovernanceDelegationsResponse = {
      delegations: delegationsResponse.delegation_responses ?? [],
      validators: validatorsResponse.validators ?? [],
    };

    return NextResponse.json(payload);
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Failed to load delegations';
    return NextResponse.json({ error: message }, { status: 500 });
  }
}
