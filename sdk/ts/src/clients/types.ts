import type { DeliverTxResponse } from "@cosmjs/stargate";

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
