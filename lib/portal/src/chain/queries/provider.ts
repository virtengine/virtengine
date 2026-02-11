/**
 * Provider module query helpers.
 */

import type { ChainQueryClient } from "../client";
import type {
  ChainQueryOptions,
  ChainRequestResult,
  PaginationRequest,
  PaginationResponse,
} from "../types";
import { buildPaginationParams, normalizePagination } from "../types";

export interface ProviderRecord {
  owner?: string;
  name?: string;
  status?: string;
  [key: string]: unknown;
}

export interface ProvidersResponse {
  providers?: ProviderRecord[];
  pagination?: PaginationResponse;
}

export interface ProviderResponse {
  provider?: ProviderRecord;
}

export interface ProviderStatusResponse {
  status?: string;
  [key: string]: unknown;
}

const PROVIDERS_PATHS = [
  "/virtengine/provider/v1beta4/providers",
  "/virtengine/provider/v1/providers",
];
const PROVIDER_PATHS = (owner: string) => [
  "/virtengine/provider/v1beta4/providers/" + owner,
  "/virtengine/provider/v1/providers/" + owner,
];
const PROVIDER_STATUS_PATHS = (owner: string) => [
  "/virtengine/provider/v1beta4/providers/" + owner + "/status",
  "/virtengine/provider/v1/providers/" + owner + "/status",
  "/virtengine/provider/v1beta4/providers/" + owner,
  "/virtengine/provider/v1/providers/" + owner,
];

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
 * Fetch registered providers.
 */
export async function fetchProviders(
  client: ChainQueryClient,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<ProvidersResponse>> {
  return client.getJsonWithPathFallback<ProvidersResponse>(
    PROVIDERS_PATHS,
    mergeParams({}, options.pagination),
    options.request,
  );
}

/**
 * Fetch a provider registration by owner address.
 */
export async function fetchProviderRegistration(
  client: ChainQueryClient,
  owner: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<ProviderResponse>> {
  return client.getJsonWithPathFallback<ProviderResponse>(
    PROVIDER_PATHS(owner),
    undefined,
    options.request,
  );
}

/**
 * Fetch provider status metadata if available.
 */
export async function fetchProviderStatus(
  client: ChainQueryClient,
  owner: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<ProviderStatusResponse>> {
  return client.getJsonWithPathFallback<ProviderStatusResponse>(
    PROVIDER_STATUS_PATHS(owner),
    undefined,
    options.request,
  );
}

/**
 * Normalize pagination from provider queries.
 */
export function normalizeProviderPagination(response: ProvidersResponse) {
  return normalizePagination(response.pagination ?? response);
}
