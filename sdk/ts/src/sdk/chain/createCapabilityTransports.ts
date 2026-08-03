import type { GeneratedType } from "@cosmjs/proto-signing";

import type { MethodDesc } from "../client/types.ts";
import { createTxTransport } from "../transport/tx/createTxTransport.ts";
import type { CallOptions, Transport, TxCallOptions } from "../transport/types.ts";
import type { ChainCapabilityController } from "./ChainCapability.ts";

export function createCapabilityQueryTransport(
  capability: ChainCapabilityController,
  queryTransport: Transport<CallOptions>,
): Transport<CallOptions> {
  return {
    requiresTypePatching: queryTransport.requiresTypePatching,
    async unary(method, input, options) {
      capability.assertCanQuery(operationName(method));
      return queryTransport.unary(method, input, options);
    },
    async stream(method, input, options) {
      capability.assertCanQuery(operationName(method));
      return queryTransport.stream(method, input, options);
    },
  };
}

export function createCapabilityTxTransport(
  capability: ChainCapabilityController,
  getMessageType: (typeUrl: string) => GeneratedType | undefined,
): Transport<TxCallOptions> {
  return {
    requiresTypePatching: true,
    async unary(method, input, options) {
      const signer = capability.requireSigner(operationName(method));
      return createTxTransport({ client: signer, getMessageType }).unary(method, input, options);
    },
    async stream(method, input, options) {
      const signer = capability.requireSigner(operationName(method));
      return createTxTransport({ client: signer, getMessageType }).stream(method, input, options);
    },
  };
}

function operationName(method: Pick<MethodDesc, "name" | "parent">): string {
  return `${method.parent.typeName}.${method.name}`;
}
