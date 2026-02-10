import type { MessageDesc, MessageShape } from "./types.ts";

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type TypePatches = Record<string, (value: any, transform: "encode" | "decode") => any>;

export function applyPatches<T extends MessageDesc<unknown>>(transform: "encode" | "decode", schema: T, message: MessageShape<T>, patches: TypePatches): MessageShape<T> {
  if (Object.hasOwn(patches, schema.$type)) {
    return patches[schema.$type](message, transform);
  }
  return message;
}
