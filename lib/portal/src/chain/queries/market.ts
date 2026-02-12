/**
 * Market module query helpers.
 */

import type { ChainQueryClient } from "../client";
import type {
  ChainQueryOptions,
  ChainRequestResult,
  PaginationRequest,
  PaginationResponse,
} from "../types";
import { buildPaginationParams, normalizePagination } from "../types";

export interface MarketOffering {
  id?: Record<string, unknown>;
  provider?: string;
  state?: string | number;
  category?: string | number;
  price?: Record<string, unknown> | string;
  [key: string]: unknown;
}

export interface MarketOfferingsResponse {
  offerings?: MarketOffering[];
  pagination?: PaginationResponse;
  next_key?: string | null;
  total?: string | number;
}

export interface MarketOrder {
  id?: Record<string, unknown>;
  state?: string | number;
  created_at?: string | number;
  [key: string]: unknown;
}

export interface MarketOrdersResponse {
  orders?: MarketOrder[];
  pagination?: PaginationResponse;
  next_key?: string | null;
  total?: string | number;
}

export interface MarketBid {
  bid?: Record<string, unknown>;
  escrow_account?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface MarketBidsResponse {
  bids?: MarketBid[];
  pagination?: PaginationResponse;
  next_key?: string | null;
  total?: string | number;
}

export interface MarketLease {
  lease?: Record<string, unknown>;
  escrow_payment?: Record<string, unknown>;
  [key: string]: unknown;
}

export interface MarketLeasesResponse {
  leases?: MarketLease[];
  pagination?: PaginationResponse;
  next_key?: string | null;
  total?: string | number;
}

export interface MarketOrderFilters extends Record<
  string,
  string | number | undefined
> {
  owner?: string;
  dseq?: number | string;
  gseq?: number | string;
  oseq?: number | string;
  state?: string;
}

export interface MarketBidFilters extends MarketOrderFilters {
  provider?: string;
  bseq?: number | string;
}

export type MarketLeaseFilters = MarketBidFilters;

export interface MarketOrderId {
  owner: string;
  dseq: number | string;
  gseq: number | string;
  oseq: number | string;
}

export interface MarketBidId extends MarketOrderId {
  provider: string;
  bseq: number | string;
}

export interface MarketLeaseId extends MarketBidId {}

const OFFERING_PATHS = [
  "/virtengine/market/v1/offerings",
  "/marketplace/offerings",
];
const OFFERING_DETAIL_PATHS = (offeringId: string) => [
  "/virtengine/market/v1/offerings/" + offeringId,
  "/marketplace/offerings/" + offeringId,
];
const ORDER_LIST_PATHS = ["/virtengine/market/v1beta5/orders/list"];
const ORDER_INFO_PATHS = ["/virtengine/market/v1beta5/orders/info"];
const BID_LIST_PATHS = ["/virtengine/market/v1beta5/bids/list"];
const BID_INFO_PATHS = ["/virtengine/market/v1beta5/bids/info"];
const LEASE_LIST_PATHS = ["/virtengine/market/v1beta5/leases/list"];
const LEASE_INFO_PATHS = ["/virtengine/market/v1beta5/leases/info"];

function buildFilterParams(
  filters?: Record<string, string | number | undefined>,
): Record<string, string> {
  if (!filters) return {};
  return Object.entries(filters).reduce<Record<string, string>>(
    (acc, [key, value]) => {
      if (value !== undefined && value !== null && value !== "") {
        acc["filters." + key] = String(value);
      }
      return acc;
    },
    {},
  );
}

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
 * Fetch marketplace offerings.
 */
export async function fetchMarketOfferings(
  client: ChainQueryClient,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketOfferingsResponse>> {
  const params = mergeParams({}, options.pagination);
  return client.getJsonWithPathFallback<MarketOfferingsResponse>(
    OFFERING_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch a single offering by its composite identifier.
 */
export async function fetchMarketOffering(
  client: ChainQueryClient,
  offeringId: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<{ offering?: MarketOffering } | MarketOffering>> {
  return client.getJsonWithPathFallback<
    { offering?: MarketOffering } | MarketOffering
  >(OFFERING_DETAIL_PATHS(offeringId), undefined, options.request);
}

/**
 * Fetch market orders with optional filters.
 */
export async function fetchMarketOrders(
  client: ChainQueryClient,
  filters?: MarketOrderFilters,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketOrdersResponse>> {
  const params = mergeParams(buildFilterParams(filters), options.pagination);
  return client.getJsonWithPathFallback<MarketOrdersResponse>(
    ORDER_LIST_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch a single market order by ID.
 */
export async function fetchMarketOrder(
  client: ChainQueryClient,
  id: MarketOrderId,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<{ order?: MarketOrder }>> {
  const params = {
    "id.owner": id.owner,
    "id.dseq": String(id.dseq),
    "id.gseq": String(id.gseq),
    "id.oseq": String(id.oseq),
  };
  return client.getJsonWithPathFallback<{ order?: MarketOrder }>(
    ORDER_INFO_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch bids with optional filters.
 */
export async function fetchMarketBids(
  client: ChainQueryClient,
  filters?: MarketBidFilters,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketBidsResponse>> {
  const params = mergeParams(buildFilterParams(filters), options.pagination);
  return client.getJsonWithPathFallback<MarketBidsResponse>(
    BID_LIST_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch a single bid by ID.
 */
export async function fetchMarketBid(
  client: ChainQueryClient,
  id: MarketBidId,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketBid>> {
  const params = {
    "id.owner": id.owner,
    "id.dseq": String(id.dseq),
    "id.gseq": String(id.gseq),
    "id.oseq": String(id.oseq),
    "id.provider": id.provider,
    "id.bseq": String(id.bseq),
  };
  return client.getJsonWithPathFallback<MarketBid>(
    BID_INFO_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch leases with optional filters.
 */
export async function fetchMarketLeases(
  client: ChainQueryClient,
  filters?: MarketLeaseFilters,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketLeasesResponse>> {
  const params = mergeParams(buildFilterParams(filters), options.pagination);
  return client.getJsonWithPathFallback<MarketLeasesResponse>(
    LEASE_LIST_PATHS,
    params,
    options.request,
  );
}

/**
 * Fetch a single lease by ID.
 */
export async function fetchMarketLease(
  client: ChainQueryClient,
  id: MarketLeaseId,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<MarketLease>> {
  const params = {
    "id.owner": id.owner,
    "id.dseq": String(id.dseq),
    "id.gseq": String(id.gseq),
    "id.oseq": String(id.oseq),
    "id.provider": id.provider,
    "id.bseq": String(id.bseq),
  };
  return client.getJsonWithPathFallback<MarketLease>(
    LEASE_INFO_PATHS,
    params,
    options.request,
  );
}

/**
 * Normalize the pagination data returned by market queries.
 */
export function normalizeMarketPagination(
  response:
    | MarketOfferingsResponse
    | MarketOrdersResponse
    | MarketBidsResponse
    | MarketLeasesResponse,
) {
  return normalizePagination(response.pagination ?? response);
}
