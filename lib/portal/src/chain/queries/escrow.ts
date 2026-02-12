/**
 * Escrow and settlement query helpers.
 */

import type { ChainQueryClient } from "../client";
import type {
  ChainQueryOptions,
  ChainRequestResult,
  PaginationRequest,
  PaginationResponse,
} from "../types";
import { buildPaginationParams, normalizePagination } from "../types";

export interface EscrowAccount {
  id?: string;
  owner?: string;
  state?: string;
  [key: string]: unknown;
}

export interface EscrowAccountsResponse {
  accounts?: EscrowAccount[];
  pagination?: PaginationResponse;
}

export interface EscrowPayment {
  id?: string;
  state?: string;
  amount?: Record<string, unknown> | string;
  [key: string]: unknown;
}

export interface EscrowPaymentsResponse {
  payments?: EscrowPayment[];
  pagination?: PaginationResponse;
}

export interface SettlementRecord {
  settlement_id?: string;
  order_id?: string;
  provider?: string;
  customer?: string;
  state?: string;
  [key: string]: unknown;
}

export interface SettlementRecordsResponse {
  settlements?: SettlementRecord[];
}

export interface EscrowAccountsFilters extends Record<
  string,
  string | undefined
> {
  state?: string;
  xid?: string;
}

export interface EscrowPaymentsFilters extends Record<
  string,
  string | undefined
> {
  state?: string;
  xid?: string;
}

const ACCOUNTS_PATH = "/virtengine/escrow/v1/types/accounts";
const PAYMENTS_PATH = "/virtengine/escrow/v1/types/payments";
const SETTLEMENTS_BY_ORDER_PATH = (orderId: string) =>
  "/virtengine/settlement/v1/settlements/by-order/" + orderId;

function mergeParams(
  base: Record<string, string>,
  pagination?: PaginationRequest,
): Record<string, string> {
  return {
    ...base,
    ...buildPaginationParams(pagination),
  };
}

function buildFilterParams(
  filters?: Record<string, string | undefined>,
): Record<string, string> {
  if (!filters) return {};
  return Object.entries(filters).reduce<Record<string, string>>(
    (acc, [key, value]) => {
      if (value) acc[key] = value;
      return acc;
    },
    {},
  );
}

/**
 * Fetch escrow accounts with optional filters.
 */
export async function fetchEscrowAccounts(
  client: ChainQueryClient,
  filters?: EscrowAccountsFilters,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<EscrowAccountsResponse>> {
  const params = mergeParams(buildFilterParams(filters), options.pagination);
  return client.getJson<EscrowAccountsResponse>(
    ACCOUNTS_PATH,
    params,
    options.request,
  );
}

/**
 * Fetch escrow payments with optional filters.
 */
export async function fetchEscrowPayments(
  client: ChainQueryClient,
  filters?: EscrowPaymentsFilters,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<EscrowPaymentsResponse>> {
  const params = mergeParams(buildFilterParams(filters), options.pagination);
  return client.getJson<EscrowPaymentsResponse>(
    PAYMENTS_PATH,
    params,
    options.request,
  );
}

/**
 * Fetch settlement records by order id.
 */
export async function fetchEscrowSettlements(
  client: ChainQueryClient,
  orderId: string,
  options: ChainQueryOptions = {},
): Promise<ChainRequestResult<SettlementRecordsResponse>> {
  return client.getJson<SettlementRecordsResponse>(
    SETTLEMENTS_BY_ORDER_PATH(orderId),
    undefined,
    options.request,
  );
}

/**
 * Normalize pagination response from escrow queries.
 */
export function normalizeEscrowPagination(
  response: EscrowAccountsResponse | EscrowPaymentsResponse,
) {
  return normalizePagination(response.pagination ?? response);
}
