export interface EncryptionOptions {
  allowInsecure?: boolean;
}

export interface EncryptionResult {
  ciphertext: string;
  keyId: string;
  algorithm: string;
}

export function encryptPayload(payload: string, options: EncryptionOptions = {}): EncryptionResult {
  if (!options.allowInsecure) {
    throw new Error("Encryption adapter not configured. Provide a native X25519-XSalsa20-Poly1305 implementation.");
  }

  const base64 = typeof Buffer !== "undefined"
    ? Buffer.from(payload, "utf8").toString("base64")
    : payload;

  return {
    ciphertext: base64,
    keyId: "insecure-dev-key",
    algorithm: "base64-dev"
  };
}
