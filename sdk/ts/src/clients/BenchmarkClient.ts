import type {
  MsgFlagProvider,
  MsgRequestChallenge,
  MsgResolveAnomalyFlag,
  MsgRespondChallenge,
  MsgSubmitBenchmarks,
  MsgUnflagProvider,
} from "../generated/protos/virtengine/benchmark/v1/tx.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";
import { BaseClient, type ClientOptions } from "./BaseClient.ts";
import type { ChainNodeSDK, ClientTxResult } from "./types.ts";
import { withTxResult } from "./types.ts";

export interface BenchmarkClientDeps {
  sdk: ChainNodeSDK;
}

/**
 * Client for Benchmark module
 */
export class BenchmarkClient extends BaseClient {
  private sdk: ChainNodeSDK;

  constructor(deps: BenchmarkClientDeps, options?: ClientOptions) {
    super(options);
    this.sdk = deps.sdk;
  }

  async submitBenchmarks(params: MsgSubmitBenchmarks, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.submitBenchmarks(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "submitBenchmarks");
    }
  }

  async requestChallenge(params: MsgRequestChallenge, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.requestChallenge(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "requestChallenge");
    }
  }

  async respondChallenge(params: MsgRespondChallenge, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.respondChallenge(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "respondChallenge");
    }
  }

  async flagProvider(params: MsgFlagProvider, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.flagProvider(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "flagProvider");
    }
  }

  async unflagProvider(params: MsgUnflagProvider, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.unflagProvider(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "unflagProvider");
    }
  }

  async resolveAnomalyFlag(params: MsgResolveAnomalyFlag, options?: TxCallOptions): Promise<ClientTxResult> {
    try {
      const { txResult } = await withTxResult((txOptions) =>
        this.sdk.virtengine.benchmark.v1.resolveAnomalyFlag(params, txOptions), options);
      return txResult;
    } catch (error) {
      this.handleQueryError(error, "resolveAnomalyFlag");
    }
  }
}
