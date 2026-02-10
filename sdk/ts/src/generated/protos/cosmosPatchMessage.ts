import { patches } from "../patches/cosmosCustomTypePatches.ts";
import type { MessageDesc } from "../../sdk/client/types.ts";
export const patched = <T extends MessageDesc>(messageDesc: T): T => {
  const patchMessage = patches[messageDesc.$type as keyof typeof patches] as any;
  if (!patchMessage) return messageDesc;
  return {
    ...messageDesc,
    encode(message, writer) {
      return messageDesc.encode(patchMessage(message, 'encode'), writer);
    },
    decode(input, length) {
      return patchMessage(messageDesc.decode(input, length), 'decode');
    },
  };
};
