import { getMessageType } from "../../../getMessageType.ts";
import type { BaseGenericStargateClientOptions, StargateTxClient, WithSigner } from "./createGenericStargateClient.ts";
import { createGenericStargateClient } from "./createGenericStargateClient.ts";

export type StargateClientOptions = WithSigner<Omit<BaseGenericStargateClientOptions, "createClient" | "getMessageType">>;
export function createStargateClient(options: StargateClientOptions): StargateTxClient {
  return createGenericStargateClient({
    ...options,
    getMessageType,
  });
}
