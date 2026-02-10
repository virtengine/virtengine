/**
 * Chat context builder
 */

import type { ChatContextOptions, ChatContextSnapshot } from "./types";
import type { QueryClient } from "../../types/chain";

const DEFAULT_DENOM = "uve";

async function safeQuery<T>(fn: () => Promise<T>): Promise<T | null> {
  try {
    return await fn();
  } catch {
    return null;
  }
}

async function queryActiveLeases(
  queryClient: QueryClient,
  walletAddress: string,
): Promise<ChatContextSnapshot["activeLeases"]> {
  const response = await safeQuery(() =>
    queryClient.query<{ leases?: Array<any> }>(
      "/virtengine/market/v1beta5/leases/list",
      {
        "filters.owner": walletAddress,
        "filters.state": "active",
        "pagination.limit": "50",
      },
    ),
  );

  if (!response?.leases) return [];

  return response.leases.map((lease) => ({
    id: lease.lease_id ?? lease.id ?? "unknown",
    provider: lease.provider ?? lease.provider_address,
    state: lease.state ?? lease.status,
    createdAt: lease.created_at
      ? Number.parseInt(lease.created_at, 10)
      : undefined,
  }));
}

async function queryActiveOrders(
  queryClient: QueryClient,
  walletAddress: string,
): Promise<ChatContextSnapshot["activeOrders"]> {
  const response = await safeQuery(() =>
    queryClient.query<{ orders?: Array<any> }>("/virtengine/market/v1/orders", {
      "filters.owner": walletAddress,
      "pagination.limit": "25",
    }),
  );

  if (!response?.orders) return [];

  return response.orders.map((order) => ({
    id: order.order_id ?? order.id ?? "unknown",
    state: order.state ?? order.status,
    offeringId: order.offering_id ?? order.offeringId,
    createdAt: order.created_at
      ? Number.parseInt(order.created_at, 10)
      : undefined,
  }));
}

export async function buildChatContext(
  options: ChatContextOptions,
): Promise<ChatContextSnapshot> {
  const walletAddress = options.walletAddress ?? undefined;
  const chainId = options.chainId ?? undefined;
  const queryClient = options.queryClient ?? undefined;
  const denom = options.tokenDenom ?? DEFAULT_DENOM;

  const context: ChatContextSnapshot = {
    walletAddress,
    chainId,
    roles: options.roles ?? [],
    permissions: options.permissions ?? [],
  };

  if (!walletAddress || !queryClient) {
    return context;
  }

  const balancePromise = safeQuery(() =>
    queryClient.queryBalance(walletAddress, denom),
  );
  const identityPromise = safeQuery(() =>
    queryClient.queryIdentity(walletAddress),
  );
  const leasesPromise = queryActiveLeases(queryClient, walletAddress);
  const ordersPromise = queryActiveOrders(queryClient, walletAddress);

  const [balance, identity, activeLeases, activeOrders] = await Promise.all([
    balancePromise,
    identityPromise,
    leasesPromise,
    ordersPromise,
  ]);

  if (balance) {
    context.balance = {
      denom: balance.denom,
      amount: balance.amount,
    };
  }

  if (identity) {
    context.veid = {
      status: identity.status,
      score: identity.score,
      updatedAt: identity.updatedAt,
    };
  }

  if (activeLeases) {
    context.activeLeases = activeLeases;
  }

  if (activeOrders) {
    context.activeOrders = activeOrders;
  }

  return context;
}

export function formatChatContext(context: ChatContextSnapshot): string {
  const payload: Record<string, unknown> = {
    walletAddress: context.walletAddress,
    chainId: context.chainId,
    balance: context.balance,
    veid: context.veid,
    activeLeases: context.activeLeases,
    activeOrders: context.activeOrders,
    roles: context.roles,
    permissions: context.permissions,
  };

  return JSON.stringify(payload, null, 2);
}
