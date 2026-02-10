import type { MessageDesc, MethodDesc } from "../client/types.ts";

export function createSerialization(type: MessageDesc) {
  return {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    parse: (data: Uint8Array): any => type.decode(data),
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    serialize: (data: any) => type.encode(data).finish(),
  };
}

export function coerceTimeoutMs(timeoutMs: number | undefined, defaultTimeoutMs: number | undefined): number | undefined {
  const value = timeoutMs !== undefined && (timeoutMs <= 0 || Number.isNaN(timeoutMs)) ? undefined : timeoutMs;
  return value === undefined ? defaultTimeoutMs : value;
}

export function createMethodUrl(baseUrl: string | URL, method: MethodDesc): string {
  return baseUrl
    .toString()
    .replace(/\/?$/, `/${method.parent.typeName}/${method.name}`);
}
