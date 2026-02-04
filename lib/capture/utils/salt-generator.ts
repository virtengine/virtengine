// @ts-nocheck
/**
 * Salt Generator Utility
 * VE-210: Per-upload salt generation for binding
 *
 * Generates cryptographically secure random salts for each upload,
 * compatible with VE-207 salt-binding requirements.
 */

/**
 * Default salt length in bytes
 */
export const DEFAULT_SALT_LENGTH = 32;

/**
 * Salt generation options
 */
export interface SaltOptions {
  /** Length of salt in bytes */
  length?: number;
  /** Optional prefix to include */
  prefix?: Uint8Array;
  /** Timestamp binding (will be included in salt derivation) */
  timestampBinding?: boolean;
}

/**
 * Generated salt result
 */
export interface GeneratedSalt {
  /** The salt bytes */
  salt: Uint8Array;
  /** Timestamp when salt was generated */
  generatedAt: Date;
  /** Salt as hex string */
  hex: string;
  /** Salt as base64 string */
  base64: string;
}

/**
 * Get the Web Crypto API
 */
function getCrypto(): Crypto {
  if (typeof window !== 'undefined' && window.crypto) {
    return window.crypto;
  }
  if (typeof globalThis !== 'undefined' && globalThis.crypto) {
    return globalThis.crypto;
  }
  throw new Error('Web Crypto API not available');
}

/**
 * Generate a cryptographically secure random salt
 *
 * @param options - Salt generation options
 * @returns Generated salt with metadata
 */
export function generateSalt(options: SaltOptions = {}): GeneratedSalt {
  const { length = DEFAULT_SALT_LENGTH, prefix, timestampBinding = true } = options;

  const crypto = getCrypto();
  const now = new Date();

  // Generate random bytes
  let randomBytes = new Uint8Array(length);
  crypto.getRandomValues(randomBytes);

  // If timestamp binding is enabled, XOR with timestamp bytes
  if (timestampBinding) {
    const timestamp = BigInt(now.getTime());
    const timestampBytes = new Uint8Array(8);
    for (let i = 0; i < 8; i++) {
      timestampBytes[i] = Number((timestamp >> BigInt(i * 8)) & BigInt(0xff));
    }

    // XOR timestamp into first 8 bytes
    for (let i = 0; i < Math.min(8, randomBytes.length); i++) {
      randomBytes[i] ^= timestampBytes[i];
    }
  }

  // If prefix is provided, prepend it
  let salt: Uint8Array;
  if (prefix && prefix.length > 0) {
    salt = new Uint8Array(prefix.length + randomBytes.length);
    salt.set(prefix, 0);
    salt.set(randomBytes, prefix.length);
  } else {
    salt = randomBytes;
  }

  return {
    salt,
    generatedAt: now,
    hex: bytesToHex(salt),
    base64: bytesToBase64(salt),
  };
}

/**
 * Generate a device-bound salt that incorporates device fingerprint
 *
 * @param deviceFingerprint - Device fingerprint string
 * @param options - Additional salt options
 * @returns Generated salt bound to device
 */
export async function generateDeviceBoundSalt(
  deviceFingerprint: string,
  options: SaltOptions = {}
): Promise<GeneratedSalt> {
  const crypto = getCrypto();

  // Hash the device fingerprint
  const fingerprintBytes = new TextEncoder().encode(deviceFingerprint);
  const fingerprintHash = await crypto.subtle.digest('SHA-256', fingerprintBytes);
  const fingerprintHashArray = new Uint8Array(fingerprintHash);

  // Generate random salt
  const baseSalt = generateSalt(options);

  // Combine: XOR fingerprint hash into the salt
  const combinedSalt = new Uint8Array(baseSalt.salt.length);
  for (let i = 0; i < baseSalt.salt.length; i++) {
    combinedSalt[i] = baseSalt.salt[i] ^ fingerprintHashArray[i % fingerprintHashArray.length];
  }

  return {
    salt: combinedSalt,
    generatedAt: baseSalt.generatedAt,
    hex: bytesToHex(combinedSalt),
    base64: bytesToBase64(combinedSalt),
  };
}

/**
 * Generate a session-bound salt that includes a session identifier
 *
 * @param sessionId - Session identifier
 * @param options - Additional salt options
 * @returns Generated salt bound to session
 */
export async function generateSessionBoundSalt(
  sessionId: string,
  options: SaltOptions = {}
): Promise<GeneratedSalt> {
  const crypto = getCrypto();

  // Hash the session ID
  const sessionBytes = new TextEncoder().encode(sessionId);
  const sessionHash = await crypto.subtle.digest('SHA-256', sessionBytes);
  const sessionPrefix = new Uint8Array(sessionHash.slice(0, 8));

  // Generate salt with session prefix
  return generateSalt({
    ...options,
    prefix: sessionPrefix,
  });
}

/**
 * Verify that a salt meets minimum security requirements
 *
 * @param salt - Salt to verify
 * @returns True if salt is valid
 */
export function verifySalt(salt: Uint8Array): boolean {
  // Minimum length check
  if (salt.length < 16) {
    return false;
  }

  // Check for all zeros (indicates potential RNG failure)
  let allZeros = true;
  for (const byte of salt) {
    if (byte !== 0) {
      allZeros = false;
      break;
    }
  }
  if (allZeros) {
    return false;
  }

  // Check for all same value (indicates potential RNG failure)
  const firstByte = salt[0];
  let allSame = true;
  for (const byte of salt) {
    if (byte !== firstByte) {
      allSame = false;
      break;
    }
  }
  if (allSame) {
    return false;
  }

  return true;
}

/**
 * Convert bytes to hex string
 */
export function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Convert hex string to bytes
 */
export function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(hex.substr(i * 2, 2), 16);
  }
  return bytes;
}

/**
 * Convert bytes to base64 string
 */
export function bytesToBase64(bytes: Uint8Array): string {
  if (typeof btoa !== 'undefined') {
    return btoa(String.fromCharCode(...bytes));
  }
  // Node.js fallback
  return Buffer.from(bytes).toString('base64');
}

/**
 * Convert base64 string to bytes
 */
export function base64ToBytes(base64: string): Uint8Array {
  if (typeof atob !== 'undefined') {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }
  // Node.js fallback
  return new Uint8Array(Buffer.from(base64, 'base64'));
}

/**
 * Concatenate multiple Uint8Arrays
 */
export function concatBytes(...arrays: Uint8Array[]): Uint8Array {
  const totalLength = arrays.reduce((sum, arr) => sum + arr.length, 0);
  const result = new Uint8Array(totalLength);
  let offset = 0;
  for (const arr of arrays) {
    result.set(arr, offset);
    offset += arr.length;
  }
  return result;
}
