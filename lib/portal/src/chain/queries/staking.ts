/**
 * Staking module query helpers.
 */

import type { ChainQueryClient } from "../client";
import type {
  ChainQueryOptions,
  ChainRequestResult,
  PaginationRequest,
  PaginationResponse,
} from "../types";
import { buildPaginationParams, normalizePagination } from "../types";

export interface StakingValidator {
  operator_address?: string;
  description?: Record<string, unknown>;
  status?: string;
  [key: string]: unknown;
}

export interface StakingValidatorsResponse {
  validators?: StakingValidator[];
  pagination?: PaginationResponse;
}

export interface StakingDelegation {
  delegator_address?: string;
  validator_address?: string;
  balance?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface StakingDelegationsResponse {
  delegation_responses?: StakingDelegation[];
  pagination?: PaginationResponse;
}

const VALIDATORS_PATH = "/cosmos/staking/v1beta1/validators";
const DELEGATIONS_PATH = (delegatorAddress: string) =>
  "/cosmos/staking/v1beta1/delegations/" + delegatorAddress;

function mergeParams(
  base: Record<string, string>,
  pagination?: PaginationRequest,
): Record<string, string> {
  return {
    ...base,
    ...buildPaginationParams(pagination),
  };
}

/**
 * Fetch validators with optional status filter.
 */
export async function fetchStakingValidators(
  client: ChainQueryClient,
  status?: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<StakingValidatorsResponse>> {
  const params: Record<string, string> = {};
  if (status) {
    params.status = status;
  }
  return client.getJson<StakingValidatorsResponse>(
    VALIDATORS_PATH,
    mergeParams(params, options.pagination),
    options.request,
  );
}

/**
 * Fetch delegations for a delegator address.
 */
export async function fetchStakingDelegations(
  client: ChainQueryClient,
  delegatorAddress: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<StakingDelegationsResponse>> {
  return client.getJson<StakingDelegationsResponse>(
    DELEGATIONS_PATH(delegatorAddress),
    buildPaginationParams(options.pagination),
    options.request,
  );
}

/**
 * Normalize pagination from staking queries.
 */
export function normalizeStakingPagination(
  response: StakingValidatorsResponse | StakingDelegationsResponse,
) {
  return normalizePagination(response.pagination ?? response);
}
