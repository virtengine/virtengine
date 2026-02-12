/**
 * Market transaction builders.
 */

import type { ChainTxMessage } from "../types";

export interface CreateOrderInput {
  cpu: number;
  memoryGb: number;
  storageGb: number;
  price?: string;
  customer?: string;
  metadata?: Record<string, unknown>;
}

export interface CloseOrderInput {
  orderId: string;
  reason?: string;
}

export interface CreateBidInput {
  orderId: string;
  provider?: string;
  price?: string;
  deposit?: string;
  resourcesOffer?: Record<string, unknown>[];
}

/**
 * Build MsgCreateOrder.
 */
export function buildMsgCreateOrder(input: CreateOrderInput): ChainTxMessage {
  return {
    typeUrl: "/virtengine.market.v1.MsgCreateOrder",
    value: {
      customer: input.customer,
      cpu: input.cpu,
      memory_gb: input.memoryGb,
      storage_gb: input.storageGb,
      price: input.price ?? "",
      metadata: input.metadata,
    },
  };
}

/**
 * Build MsgCloseOrder.
 */
export function buildMsgCloseOrder(input: CloseOrderInput): ChainTxMessage {
  return {
    typeUrl: "/virtengine.market.v1.MsgCloseOrder",
    value: {
      order_id: input.orderId,
      reason: input.reason,
    },
  };
}

/**
 * Build MsgCreateBid.
 */
export function buildMsgCreateBid(input: CreateBidInput): ChainTxMessage {
  return {
    typeUrl: "/virtengine.market.v1.MsgCreateBid",
    value: {
      order_id: input.orderId,
      provider: input.provider,
      price: input.price,
      deposit: input.deposit,
      resources_offer: input.resourcesOffer ?? [],
    },
  };
}
