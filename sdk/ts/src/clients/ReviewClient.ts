import type { MsgDeleteReview, MsgSubmitReview } from "../generated/protos/virtengine/review/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface ReviewClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Review module
 */
export class ReviewClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: ReviewClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async submitReview(params: MsgSubmitReview, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.review.v1.submitReview(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "submitReview");
    }
  }

  async deleteReview(params: MsgDeleteReview, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.review.v1.deleteReview(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "deleteReview");
    }
  }
}
