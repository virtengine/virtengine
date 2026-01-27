import { fromByteArray } from "base64-js";

export function base64UrlEncode(value: string | Uint8Array): string {
  return toBase64Url(base64Encode(value));
}

/**
 * Converts a base64 encoded string to a base64url encoded string
 */
export function toBase64Url(base64Encoded: string): string {
  return base64Encoded.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
}

const textDecoder = new TextDecoder();
export function base64UrlDecode(value: string): string {
  let str = value;
  // Convert from base64url → base64
  str = str.replace(/-/g, "+").replace(/_/g, "/");
  str = str.padEnd(str.length + (4 - (str.length % 4)) % 4, "=");

  return textDecoder.decode(Uint8Array.from(atob(str), (c) => c.charCodeAt(0)));
}

/**
 * Decode a base64 string
 * @param base64String The base64 string to decode
 * @returns The decoded object
 */
export function base64Decode(base64String: string): Record<string, unknown> {
  const decoded = atob(base64String);
  return JSON.parse(decoded);
}

const textEncoder = new TextEncoder();
export function base64Encode(value: string | Uint8Array): string {
  const data = typeof value === "string" ? textEncoder.encode(value) : value;
  return fromByteArray(data);
}
