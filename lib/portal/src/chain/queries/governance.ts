/**
 * Governance module query helpers.
 */

import type { ChainQueryClient } from "../client";
import type {
  ChainQueryOptions,
  ChainRequestResult,
  PaginationRequest,
  PaginationResponse,
} from "../types";
import { buildPaginationParams, normalizePagination } from "../types";

export interface GovProposal {
  id?: string;
  title?: string;
  status?: string;
  [key: string]: unknown;
}

export interface GovProposalsResponse {
  proposals?: GovProposal[];
  pagination?: PaginationResponse;
}

export interface GovVote {
  proposal_id?: string;
  voter?: string;
  option?: string;
  [key: string]: unknown;
}

export interface GovVotesResponse {
  votes?: GovVote[];
  pagination?: PaginationResponse;
}

export interface GovParamsResponse {
  params?: Record<string, unknown>;
}

const PROPOSALS_PATH = "/cosmos/gov/v1/proposals";
const PROPOSAL_PATH = (proposalId: string) =>
  "/cosmos/gov/v1/proposals/" + proposalId;
const VOTES_PATH = (proposalId: string) =>
  "/cosmos/gov/v1/proposals/" + proposalId + "/votes";
const PARAMS_PATH = (paramsType: string) =>
  "/cosmos/gov/v1/params/" + paramsType;

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
 * Fetch governance proposals.
 */
export async function fetchGovProposals(
  client: ChainQueryClient,
  status?: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<GovProposalsResponse>> {
  const params: Record<string, string> = {};
  if (status) {
    params.proposal_status = status;
  }
  return client.getJson<GovProposalsResponse>(
    PROPOSALS_PATH,
    mergeParams(params, options.pagination),
    options.request,
  );
}

/**
 * Fetch a single governance proposal.
 */
export async function fetchGovProposal(
  client: ChainQueryClient,
  proposalId: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<{ proposal?: GovProposal }>> {
  return client.getJson<{ proposal?: GovProposal }>(
    PROPOSAL_PATH(proposalId),
    undefined,
    options.request,
  );
}

/**
 * Fetch votes for a proposal.
 */
export async function fetchGovVotes(
  client: ChainQueryClient,
  proposalId: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<GovVotesResponse>> {
  return client.getJson<GovVotesResponse>(
    VOTES_PATH(proposalId),
    buildPaginationParams(options.pagination),
    options.request,
  );
}

/**
 * Fetch governance module parameters.
 */
export async function fetchGovParams(
  client: ChainQueryClient,
  paramsType: "voting" | "deposit" | "tallying" = "voting",
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<GovParamsResponse>> {
  return client.getJson<GovParamsResponse>(
    PARAMS_PATH(paramsType),
    undefined,
    options.request,
  );
}

/**
 * Normalize pagination data from governance queries.
 */
export function normalizeGovPagination(
  response: GovProposalsResponse | GovVotesResponse,
) {
  return normalizePagination(response.pagination ?? response);
}
