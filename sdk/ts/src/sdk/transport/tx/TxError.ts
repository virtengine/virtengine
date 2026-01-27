import type { DeliverTxResponse } from "@cosmjs/stargate";

import { TransportError } from "../TransportError.ts";

export class TxError extends TransportError {
  readonly txResponse: DeliverTxResponse;

  constructor(message: string, txResponse: DeliverTxResponse, code = TransportError.Code.Unknown) {
    super(message, code);
    this.name = "TxError";
    this.txResponse = txResponse;
  }
}
