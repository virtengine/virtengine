/**
 * VEID module query helpers.
 */

import type { ChainQueryClient } from "../client";
import type { ChainQueryOptions, ChainRequestResult } from "../types";

export interface VeidRecord {
  account_address?: string;
  status?: string;
  state?: string;
  score?: string | number;
  [key: string]: unknown;
}

export interface VeidScore {
  score?: string | number;
  tier?: string;
  [key: string]: unknown;
}

export interface VeidScope {
  scope_id?: string;
  scope_type?: string;
  state?: string;
  [key: string]: unknown;
}

export interface VeidScopesResponse {
  scopes?: VeidScope[];
}

const RECORD_PATHS = (accountAddress: string) => [
  "/virtengine/veid/v1/identity_record/" + accountAddress,
  "/virtengine/veid/v1/identity-record/" + accountAddress,
  "/virtengine/veid/v1/identity_records/" + accountAddress,
  "/virtengine/veid/v1/identity/" + accountAddress,
];

const SCORE_PATHS = (accountAddress: string) => [
  "/virtengine/veid/v1/score/" + accountAddress,
];

const SCOPES_PATH = (accountAddress: string) =>
  "/virtengine/veid/v1/scopes/" + accountAddress;
const SCOPES_BY_TYPE_PATH = (accountAddress: string, scopeType: string) =>
  "/virtengine/veid/v1/scopes/" + accountAddress + "/type/" + scopeType;

/**
 * Fetch an identity record for a given account address.
 */
export async function fetchVeidRecord(
  client: ChainQueryClient,
  accountAddress: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<{ record?: VeidRecord } | VeidRecord>> {
  return client.getJsonWithPathFallback<{ record?: VeidRecord } | VeidRecord>(
    RECORD_PATHS(accountAddress),
    undefined,
    options.request,
  );
}

/**
 * Fetch the VEID score for a given account address.
 */
export async function fetchVeidScore(
  client: ChainQueryClient,
  accountAddress: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<{ score?: VeidScore } | VeidScore>> {
  return client.getJsonWithPathFallback<{ score?: VeidScore } | VeidScore>(
    SCORE_PATHS(accountAddress),
    undefined,
    options.request,
  );
}

/**
 * Fetch VEID scopes for an account address.
 */
export async function fetchVeidScopes(
  client: ChainQueryClient,
  accountAddress: string,
  scopeType?: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<VeidScopesResponse>> {
  const path = scopeType
    ? SCOPES_BY_TYPE_PATH(accountAddress, scopeType)
    : SCOPES_PATH(accountAddress);
  return client.getJson<VeidScopesResponse>(path, undefined, options.request);
}
