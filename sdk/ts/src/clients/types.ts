import type { DeliverTxResponse } from "@cosmjs/stargate";

import { bytesFromBase64, numberToLong } from "../encoding/typeEncodingHelpers.ts";
import type { PageRequest } from "../generated/protos/cosmos/base/query/v1beta1/pagination.ts";
import type { createChainNodeSDK } from "../sdk/chain/createChainNodeSDK.ts";
import type { TxCallOptions } from "../sdk/transport/types.ts";

/**
 * Common filter options for list operations
 */
export interface ListOptions {
  limit?: number;
  offset?: number;
  cursor?: string;
}

/**
 * Transaction result from a successful broadcast
 */
export interface ClientTxResult {
  transactionHash: string;
  height: number;
  code: number;
  gasWanted: bigint;
  gasUsed: bigint;
  rawLog?: string;
}

/**
 * Convert DeliverTxResponse to ClientTxResult
 */
export function toClientTxResult(response: DeliverTxResponse): ClientTxResult {
  return {
    transactionHash: response.transactionHash,
    height: response.height,
    code: response.code,
    gasWanted: response.gasWanted,
    gasUsed: response.gasUsed,
    rawLog: response.rawLog,
  };
}

/**
 * SDK type used by module clients (matches the Chain Node SDK shape).
 */
export type ChainNodeSDK = ReturnType<typeof createChainNodeSDK>;

/**
 * Convert list options into a Cosmos SDK pagination request.
 */
export function toPageRequest(options?: ListOptions): PageRequest | undefined {
  if (!options) return undefined;

  const hasCursor = Boolean(options.cursor);
  const hasOffset = typeof options.offset === "number";
  const hasLimit = typeof options.limit === "number";

  if (!hasCursor && !hasOffset && !hasLimit) return undefined;

  return {
    key: hasCursor ? bytesFromBase64(options.cursor!) : new Uint8Array(),
    offset: numberToLong(options.offset ?? 0),
    limit: numberToLong(options.limit ?? 0),
    countTotal: false,
    reverse: false,
  };
}

/**
 * Wrap a transaction call to capture the broadcast result alongside the response payload.
 */
export async function withTxResult<T>(
  executor: (options?: TxCallOptions) => Promise<T>,
  options?: TxCallOptions,
): Promise<{ response: T; txResult: ClientTxResult }> {
  let broadcast: DeliverTxResponse | undefined;
  const response = await executor({
    ...options,
    afterBroadcast: (tx) => {
      broadcast = tx;
      options?.afterBroadcast?.(tx);
    },
  });

  if (!broadcast) {
    throw new Error("Transaction broadcast result was not captured");
  }

  return { response, txResult: toClientTxResult(broadcast) };
}
