import type { GeneratedType } from "@cosmjs/proto-signing";

import { serviceLoader as cosmosServiceLoader } from "../generated/createCosmosSDK.ts";
import { serviceLoader as nodeServiceLoader } from "../generated/createNodeSDK.ts";
import { TxRaw } from "../generated/protos/cosmos/tx/v1beta1/tx.ts";
import { createMessageType } from "./client/createServiceLoader.ts";

const TxRawType = createMessageType(TxRaw);
export function getMessageType(typeUrl: string): GeneratedType | undefined {
  const type = nodeServiceLoader.getLoadedType(typeUrl) || cosmosServiceLoader.getLoadedType(typeUrl);
  if (type) return type;
  if (typeUrl === `/${TxRaw.$type}`) return TxRawType;
  return undefined;
}
