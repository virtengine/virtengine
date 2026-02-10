import type { GeneratedType } from "@cosmjs/proto-signing";

import type { MessageDesc, MessageInitShape, MessageShape, MethodDesc } from "../../client/types.ts";
import { TransportError } from "../TransportError.ts";
import type { Transport, TxCallOptions, UnaryResponse } from "../types.ts";
import type { TxClient } from "./TxClient.ts";
import { TxError } from "./TxError.ts";

export function createTxTransport(transportOptions: TransactionTransportOptions): Transport<TxCallOptions> {
  return {
    requiresTypePatching: true,
    async unary<I extends MessageDesc, O extends MessageDesc>(
      method: MethodDesc<"unary", I, O>,
      input: MessageInitShape<I>,
      options?: TxCallOptions,
    ): Promise<UnaryResponse<I, O>> {
      const messages = [{
        typeUrl: `/${method.input.$type}`,
        value: input,
      }];
      const txResponse = await transportOptions.client.signAndBroadcast(messages, {
        afterSign: options?.afterSign,
        fee: options?.fee,
        memo: options?.memo ?? `virtengine: ${method.name}`,
      });
      options?.afterBroadcast?.(txResponse);

      if (txResponse.code !== 0) {
        throw new TxError(`Transaction failed with code ${txResponse.code}`, txResponse);
      }

      const response = txResponse.msgResponses[0];
      let responseMessage: MessageShape<O>;
      if (response) {
        const MessageType = transportOptions.getMessageType(response.typeUrl);
        if (!MessageType) {
          throw new Error(`Cannot find message type ${response.typeUrl} in type registry. `
            + `If you use cosmos.authz.v1beta1.exec(), then provide custom message types to TxClient.`);
        }
        responseMessage = MessageType.decode(response.value);
      } else {
        responseMessage = {} as MessageShape<O>;
      }

      return {
        stream: false,
        header: new Headers(),
        trailer: new Headers(),
        message: responseMessage,
        method,
      };
    },
    async stream() {
      throw new TransportError(`Transaction transport doesn't support streaming`, TransportError.Code.Unimplemented);
    },
  };
}

export interface TransactionTransportOptions {
  client: TxClient;
  getMessageType: (typeUrl: string) => GeneratedType | undefined;
}
