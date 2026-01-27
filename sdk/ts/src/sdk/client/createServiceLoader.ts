import type { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";

import type { MessageDesc, MessageInitShape, MessageShape, ServiceDesc } from "./types.ts";

type LoadGrpcService = () => unknown;

/**
 * Loads a service and registers its methods input and output message types.
 */
export function createServiceLoader<T extends ReadonlyArray<LoadGrpcService>>(fns: T): ServiceLoader<T> {
  const loadedTypes: Record<string, GRPCMessageType> = {};
  return {
    getLoadedType(typeUrl) {
      return loadedTypes[typeUrl];
    },
    async loadAt(index) {
      const service = await fns[index]() as ServiceDesc;
      Object.values(service.methods).forEach((method) => {
        loadedTypes[`/${method.input.$type}`] = createMessageType(method.input);
        loadedTypes[`/${method.output.$type}`] = createMessageType(method.output);
      });

      return service;
    },
  } as ServiceLoader<T>;
}

/**
 * Create a message type for a given protobuf schema.
 * @param schema - The protobuf schema to create a message type for.
 * @returns A message type for the given protobuf schema.
 */
export function createMessageType<T extends MessageDesc>({ $type, ...schema }: T): GRPCMessageType<T> {
  return {
    typeUrl: `/${$type}`,
    ...schema,
  } as GRPCMessageType<T>;
}
export interface GRPCMessageType<T extends MessageDesc = MessageDesc> {
  typeUrl: string;
  encode(message: MessageShape<T> | MessageInitShape<T>, writer?: BinaryWriter): BinaryWriter;
  decode(input: Uint8Array | BinaryReader, length?: number): MessageShape<T>;
  fromPartial(message: MessageInitShape<T>): MessageShape<T>;
}

export interface ServiceLoader<T extends ReadonlyArray<LoadGrpcService>> {
  loadAt<TIndex extends keyof T & number>(index: TIndex): ReturnType<T[TIndex]>;
  getLoadedType(typeUrl: string): GRPCMessageType | undefined;
}
