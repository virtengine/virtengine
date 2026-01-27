import type { BinaryReader, BinaryWriter } from "@bufbuild/protobuf/wire";
import Long from "long";

import { Timestamp } from "../generated/protos/google/protobuf/timestamp.ts";

// eslint-disable-next-line @typescript-eslint/no-unsafe-function-type
type Builtin = Date | Function | Uint8Array | string | number | bigint | boolean | undefined | null;

export type DeepPartial<T> = T extends Builtin ? T
  : T extends Long ? string | number | bigint | Long : T extends globalThis.Array<infer U> ? globalThis.Array<DeepPartial<U>>
    : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepPartial<U>>
      // eslint-disable-next-line @typescript-eslint/no-empty-object-type
      : T extends {} ? { [K in keyof T]?: DeepPartial<T[K]> }
        : Partial<T>;

export type DeepSimplify<T> = T extends Builtin ? T
  : T extends Long ? string | number | bigint | Long : T extends globalThis.Array<infer U> ? globalThis.Array<DeepSimplify<U>>
    : T extends ReadonlyArray<infer U> ? ReadonlyArray<DeepSimplify<U>>
      : { [K in keyof T]: DeepSimplify<T[K]> };

export interface MessageFns<T, V extends string> {
  readonly $type: V;
  encode(message: T, writer?: BinaryWriter): BinaryWriter;
  decode(input: BinaryReader | Uint8Array, length?: number): T;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  fromJSON(object: any): T;
  toJSON(message: T): unknown;
  fromPartial(object: DeepPartial<T>): T;
}

export function isSet(value: unknown): boolean {
  return value !== null && value !== undefined;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const bytesFromBase64 = (globalThis as any).Buffer
  ? (b64: string): Uint8Array => Uint8Array.from(globalThis.Buffer.from(b64, "base64"))
  : (b64: string): Uint8Array => {
      if ("fromBase64" in Uint8Array && typeof Uint8Array.fromBase64 === "function") {
        return Uint8Array.fromBase64(b64) as Uint8Array;
      }

      const bin = globalThis.atob(b64);
      const arr = new Uint8Array(bin.length);
      for (let i = 0; i < bin.length; ++i) {
        arr[i] = bin.charCodeAt(i);
      }
      return arr;
    };

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const base64FromBytes = (globalThis as any).Buffer
  ? (arr: Uint8Array): string => globalThis.Buffer.from(arr).toString("base64")
  : (arr: Uint8Array): string => {
      if ("toBase64" in arr && typeof arr.toBase64 === "function") {
        return arr.toBase64() as string;
      }

      const bin: string[] = [];
      arr.forEach((byte) => {
        bin.push(globalThis.String.fromCharCode(byte));
      });
      return globalThis.btoa(bin.join(""));
    };

export function toTimestamp(date: Date): Timestamp {
  const seconds = numberToLong(Math.trunc(date.getTime() / 1_000));
  const nanos = (date.getTime() % 1_000) * 1_000_000;
  return { seconds, nanos };
}

export function fromTimestamp(t: Timestamp): Date {
  let millis = (t.seconds.toNumber() || 0) * 1_000;
  millis += (t.nanos || 0) / 1_000_000;
  return new globalThis.Date(millis);
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function fromJsonTimestamp(o: any): Date {
  if (o instanceof globalThis.Date) {
    return o;
  } else if (typeof o === "string") {
    return new globalThis.Date(o);
  } else {
    return fromTimestamp(Timestamp.fromJSON(o));
  }
}

export function numberToLong(number: number) {
  return Long.fromNumber(number);
}

export function isObject(value: unknown): boolean {
  return typeof value === "object" && value !== null;
}
