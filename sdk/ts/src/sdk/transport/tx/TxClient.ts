import type { EncodeObject, GeneratedType } from "@cosmjs/proto-signing";
import type { DeliverTxResponse, StdFee } from "@cosmjs/stargate";

import type { TxRaw as GenTxRaw } from "../../../generated/protos/cosmos/tx/v1beta1/tx.ts";

export type TxRaw = Omit<GenTxRaw, "$typeName" | "$unknown">;
export { DeliverTxResponse, StdFee, EncodeObject, GeneratedType };

export interface TxClient {
  signAndBroadcast(messages: EncodeObject[], options?: TxSignAndBroadcastOptions): Promise<DeliverTxResponse>;
}

export interface TxSignAndBroadcastOptions {
  fee?: Partial<StdFee>;
  memo?: string;
  timeoutHeight?: bigint;
  afterSign?: (tx: TxRaw) => void;
};
