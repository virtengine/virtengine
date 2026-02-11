import Long from "long";

import type { MsgAddPriceEntry } from "../generated/protos/virtengine/oracle/v1/msgs.ts";
import type { AggregatedPrice, PriceData, PricesFilter } from "../generated/protos/virtengine/oracle/v1/prices.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult, ListOptions } from "./types.ts";
import { toPageRequest, withTxResult } from "./types.ts";

export interface OracleClientDeps {
  sdk: ChainNodeSDK;
}

export interface OraclePriceFilters {
  assetDenom?: string;
  baseDenom?: string;
  height?: bigint | number;
}

/**
 * Client for Oracle module
 */
export class OracleClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: OracleClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  /**
   * Get aggregated price for a denom
   */
  async getPrice(denom: string): Promise<AggregatedPrice | null> {
    try {
      const result = await this.sdk.virtengine.oracle.v1.getAggregatedPrice({ denom });
      return result.aggregatedPrice ?? null;
    } catch (error) {
      this.handleQueryError(error, "getPrice");
    }
  }

  /**
   * List historical price data
   */
  async listPrices(options?: ListOptions & OraclePriceFilters): Promise<PriceData[]> {
    try {
      const filters: PricesFilter | undefined = options?.assetDenom || options?.baseDenom || options?.height !== undefined
        ? {
            assetDenom: options?.assetDenom ?? "",
            baseDenom: options?.baseDenom ?? "",
            height: Long.fromValue(options?.height ?? 0),
          }
        : undefined;

      const result = await this.sdk.virtengine.oracle.v1.getPrices({
        filters,
        pagination: toPageRequest(options),
      });
      return result.prices;
    } catch (error) {
      this.handleQueryError(error, "listPrices");
    }
  }

  /**
   * Get exchange rate (twap) for a denom
   */
  async getExchangeRate(denom: string): Promise<string | null> {
    try {
      const result = await this.sdk.virtengine.oracle.v1.getAggregatedPrice({ denom });
      return result.aggregatedPrice?.twap ?? null;
    } catch (error) {
      this.handleQueryError(error, "getExchangeRate");
    }
  }

  /**
   * List supported assets by scanning available price data
   */
  async listSupportedAssets(): Promise<string[]> {
    try {
      const result = await this.sdk.virtengine.oracle.v1.getPrices({
        filters: undefined,
        pagination: toPageRequest({ limit: 200 }),
      });
      const denoms = new Set<string>();
      result.prices.forEach((price) => {
        const denom = price.id?.denom;
        if (denom) denoms.add(denom);
      });
      return Array.from(denoms);
    } catch (error) {
      this.handleQueryError(error, "listSupportedAssets");
    }
  }

  /**
   * Submit a price entry (requires signer)
   */
  async addPriceEntry(params: MsgAddPriceEntry, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.oracle.v1.addPriceEntry(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "addPriceEntry");
    }
  }
}
