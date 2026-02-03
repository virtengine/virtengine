import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";
import type { BidID } from "../generated/protos/virtengine/market/v1/bid.ts";
import type { LeaseID } from "../generated/protos/virtengine/market/v1/lease.ts";
import type { OrderID } from "../generated/protos/virtengine/market/v1/order.ts";
import type { Bid } from "../generated/protos/virtengine/market/v1beta5/bid.ts";
import type { Order } from "../generated/protos/virtengine/market/v1beta5/order.ts";
import type { Lease } from "../generated/protos/virtengine/market/v1/lease.ts";
import type {
  MsgCloseBid,
  MsgCreateBid,
} from "../generated/protos/virtengine/market/v1beta5/bidmsg.ts";
import type { MsgCloseLease } from "../generated/protos/virtengine/market/v1beta5/leasemsg.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

export interface MarketClientDeps {
  sdk: ChainNodeSDK;
}

export interface OrderFilters {
  owner?: string;
  state?: string;
  dseq?: string | number | bigint;
  gseq?: number;
  oseq?: number;
}

export interface BidFilters {
  owner?: string;
  provider?: string;
  state?: string;
  dseq?: string | number | bigint;
  gseq?: number;
  oseq?: number;
  bseq?: number;
}

export interface LeaseFilters {
  owner?: string;
  provider?: string;
  state?: string;
  dseq?: string | number | bigint;
  gseq?: number;
  oseq?: number;
}

/**
 * Client for Market module (orders, bids, leases)
 */
export class MarketClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: MarketClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  // Order queries

  /**
   * Get order by ID
   */
  async getOrder(orderId: OrderID): Promise<Order | null> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getOrder({ id: orderId });
      return result.order ?? null;
    } catch (error) {
      this.handleQueryError(error, "getOrder");
    }
  }

  /**
   * List orders
   */
  async listOrders(options?: ListOptions & OrderFilters): Promise<Order[]> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getOrders({
        filters: {
          owner: options?.owner ?? "",
          state: options?.state ?? "",
          dseq: options?.dseq ?? 0,
          gseq: options?.gseq ?? 0,
          oseq: options?.oseq ?? 0,
        },
        pagination: toPageRequest(options),
      });
      return result.orders;
    } catch (error) {
      this.handleQueryError(error, "listOrders");
    }
  }

  // Bid queries

  /**
   * Get bid by ID
   */
  async getBid(bidId: BidID): Promise<Bid | null> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getBid({ id: bidId });
      return result.bid ?? null;
    } catch (error) {
      this.handleQueryError(error, "getBid");
    }
  }

  /**
   * List bids
   */
  async listBids(options?: ListOptions & BidFilters): Promise<Bid[]> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getBids({
        filters: {
          owner: options?.owner ?? "",
          provider: options?.provider ?? "",
          state: options?.state ?? "",
          dseq: options?.dseq ?? 0,
          gseq: options?.gseq ?? 0,
          oseq: options?.oseq ?? 0,
          bseq: options?.bseq ?? 0,
        },
        pagination: toPageRequest(options),
      });
      return result.bids
        .map((entry: { bid?: Bid }) => entry.bid)
        .filter((bid): bid is Bid => Boolean(bid));
    } catch (error) {
      this.handleQueryError(error, "listBids");
    }
  }

  // Lease queries

  /**
   * Get lease by ID
   */
  async getLease(leaseId: LeaseID): Promise<Lease | null> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getLease({ id: leaseId });
      return result.lease ?? null;
    } catch (error) {
      this.handleQueryError(error, "getLease");
    }
  }

  /**
   * List leases
   */
  async listLeases(options?: ListOptions & LeaseFilters): Promise<Lease[]> {
    try {
      const result = await this.sdk.virtengine.market.v1beta5.getLeases({
        filters: {
          owner: options?.owner ?? "",
          provider: options?.provider ?? "",
          state: options?.state ?? "",
          dseq: options?.dseq ?? 0,
          gseq: options?.gseq ?? 0,
          oseq: options?.oseq ?? 0,
        },
        pagination: toPageRequest(options),
      });
      return result.leases
        .map((entry: { lease?: Lease }) => entry.lease)
        .filter((lease): lease is Lease => Boolean(lease));
    } catch (error) {
      this.handleQueryError(error, "listLeases");
    }
  }

  // Transaction methods

  /**
   * Create a bid on an order
   */
  async createBid(params: MsgCreateBid, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.market.v1beta5.createBid(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "createBid");
    }
  }

  /**
   * Close an open bid
   */
  async closeBid(params: MsgCloseBid, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.market.v1beta5.closeBid(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "closeBid");
    }
  }

  /**
   * Close an active lease
   */
  async closeLease(params: MsgCloseLease, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.market.v1beta5.closeLease(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "closeLease");
    }
  }
}
