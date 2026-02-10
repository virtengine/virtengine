const decoder = new TextDecoder("utf-8");
const encoder = new TextEncoder();

export function decodeBinary(input: Uint8Array): string {
  return decoder.decode(input);
}

export function encodeBinary(input: string): Uint8Array {
  return encoder.encode(input);
}
