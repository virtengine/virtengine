/**
 * Signature Utility
 * VE-210: Client and user signature creation for capture uploads
 *
 * Compatible with VE-201 approved client signature requirements
 * and VE-207 salt-binding anti-replay protection.
 */

import type {
  ClientKeyProvider,
  UserKeyProvider,
  SignaturePackage,
  CaptureMetadata,
} from '../types/capture';
import { concatBytes, bytesToHex } from './salt-generator';

/**
 * Signature creation options
 */
export interface SignatureOptions {
  /** Include metadata in signature */
  includeMetadata?: boolean;
  /** Hash algorithm to use */
  hashAlgorithm?: 'SHA-256' | 'SHA-384' | 'SHA-512';
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
 * Compute SHA-256 hash of data
 */
export async function computeHash(
  data: Uint8Array,
  algorithm: 'SHA-256' | 'SHA-384' | 'SHA-512' = 'SHA-256'
): Promise<Uint8Array> {
  const crypto = getCrypto();
  const hashBuffer = await crypto.subtle.digest(algorithm, data);
  return new Uint8Array(hashBuffer);
}

/**
 * Create the payload hash for signing
 * This includes the image data and metadata
 */
export async function createPayloadHash(
  imageBlob: Blob,
  metadata: CaptureMetadata,
  options: SignatureOptions = {}
): Promise<Uint8Array> {
  const { hashAlgorithm = 'SHA-256', includeMetadata = true } = options;

  // Get image bytes
  const imageBuffer = await imageBlob.arrayBuffer();
  const imageBytes = new Uint8Array(imageBuffer);

  let dataToHash: Uint8Array;

  if (includeMetadata) {
    // Create deterministic metadata representation
    const metadataString = JSON.stringify({
      deviceFingerprint: metadata.deviceFingerprint,
      clientVersion: metadata.clientVersion,
      capturedAt: metadata.capturedAt,
      documentType: metadata.documentType,
      qualityScore: metadata.qualityScore,
      sessionId: metadata.sessionId,
    });
    const metadataBytes = new TextEncoder().encode(metadataString);

    // Concatenate image and metadata
    dataToHash = concatBytes(imageBytes, metadataBytes);
  } else {
    dataToHash = imageBytes;
  }

  return computeHash(dataToHash, hashAlgorithm);
}

/**
 * Create client signature over (salt || payloadHash)
 * This binds the capture to the specific salt and payload
 */
export async function createClientSignature(
  clientKeyProvider: ClientKeyProvider,
  salt: Uint8Array,
  payloadHash: Uint8Array
): Promise<{
  signature: Uint8Array;
  clientId: string;
  clientVersion: string;
}> {
  // Create message to sign: salt || payloadHash
  const message = concatBytes(salt, payloadHash);

  // Sign with client key
  const signature = await clientKeyProvider.sign(message);
  const clientId = await clientKeyProvider.getClientId();
  const clientVersion = await clientKeyProvider.getClientVersion();

  return {
    signature,
    clientId,
    clientVersion,
  };
}

/**
 * Create user signature over (salt || payloadHash || clientSignature)
 * This binds the user's identity to the complete package
 */
export async function createUserSignature(
  userKeyProvider: UserKeyProvider,
  salt: Uint8Array,
  payloadHash: Uint8Array,
  clientSignature: Uint8Array
): Promise<{
  signature: Uint8Array;
  userAddress: string;
}> {
  // Create message to sign: salt || payloadHash || clientSignature
  const message = concatBytes(salt, payloadHash, clientSignature);

  // Sign with user key
  const signature = await userKeyProvider.sign(message);
  const userAddress = await userKeyProvider.getAccountAddress();

  return {
    signature,
    userAddress,
  };
}

/**
 * Create complete signature package for upload
 * Includes both client and user signatures with binding
 */
export async function createSignaturePackage(
  imageBlob: Blob,
  metadata: CaptureMetadata,
  salt: Uint8Array,
  clientKeyProvider: ClientKeyProvider,
  userKeyProvider: UserKeyProvider,
  options: SignatureOptions = {}
): Promise<SignaturePackage> {
  // Create payload hash
  const payloadHash = await createPayloadHash(imageBlob, metadata, options);

  // Create client signature
  const { signature: clientSignature, clientId, clientVersion } = await createClientSignature(
    clientKeyProvider,
    salt,
    payloadHash
  );

  // Create user signature
  const { signature: userSignature, userAddress } = await createUserSignature(
    userKeyProvider,
    salt,
    payloadHash,
    clientSignature
  );

  return {
    salt,
    payloadHash,
    clientSignature,
    userSignature,
    clientId,
    clientVersion,
    userAddress,
    signedAt: new Date().toISOString(),
  };
}

/**
 * Verify client signature
 * Used for verification on the receiving end
 */
export async function verifyClientSignature(
  signaturePackage: SignaturePackage,
  clientPublicKey: Uint8Array,
  keyType: 'ed25519' | 'secp256k1'
): Promise<boolean> {
  const crypto = getCrypto();
  const { salt, payloadHash, clientSignature } = signaturePackage;

  // Reconstruct message
  const message = concatBytes(salt, payloadHash);

  // Import the public key based on type
  if (keyType === 'ed25519') {
    // Note: Web Crypto doesn't natively support Ed25519 in all browsers
    // This is a simplified check - real implementation would use a library
    const key = await crypto.subtle.importKey(
      'raw',
      clientPublicKey,
      { name: 'Ed25519' },
      false,
      ['verify']
    );

    return crypto.subtle.verify('Ed25519', key, clientSignature, message);
  } else {
    // Secp256k1 - also requires library support in most browsers
    throw new Error('Secp256k1 verification requires external library');
  }
}

/**
 * Verify user signature
 */
export async function verifyUserSignature(
  signaturePackage: SignaturePackage,
  userPublicKey: Uint8Array,
  keyType: 'ed25519' | 'secp256k1'
): Promise<boolean> {
  const crypto = getCrypto();
  const { salt, payloadHash, clientSignature, userSignature } = signaturePackage;

  // Reconstruct message
  const message = concatBytes(salt, payloadHash, clientSignature);

  if (keyType === 'ed25519') {
    const key = await crypto.subtle.importKey(
      'raw',
      userPublicKey,
      { name: 'Ed25519' },
      false,
      ['verify']
    );

    return crypto.subtle.verify('Ed25519', key, userSignature, message);
  } else {
    throw new Error('Secp256k1 verification requires external library');
  }
}

/**
 * Serialize signature package for transmission
 */
export function serializeSignaturePackage(pkg: SignaturePackage): string {
  return JSON.stringify({
    salt: bytesToHex(pkg.salt),
    payloadHash: bytesToHex(pkg.payloadHash),
    clientSignature: bytesToHex(pkg.clientSignature),
    userSignature: bytesToHex(pkg.userSignature),
    clientId: pkg.clientId,
    clientVersion: pkg.clientVersion,
    userAddress: pkg.userAddress,
    signedAt: pkg.signedAt,
  });
}

/**
 * Generate a device fingerprint for binding
 * This creates a stable identifier for the capture device
 */
export async function generateDeviceFingerprint(): Promise<string> {
  const crypto = getCrypto();

  // Collect device characteristics
  const components: string[] = [];

  // Screen info
  if (typeof screen !== 'undefined') {
    components.push(`screen:${screen.width}x${screen.height}x${screen.colorDepth}`);
  }

  // User agent
  if (typeof navigator !== 'undefined') {
    components.push(`ua:${navigator.userAgent}`);
    components.push(`lang:${navigator.language}`);
    components.push(`platform:${navigator.platform || 'unknown'}`);
  }

  // Timezone
  components.push(`tz:${Intl.DateTimeFormat().resolvedOptions().timeZone}`);

  // Canvas fingerprint (limited for privacy)
  if (typeof document !== 'undefined') {
    try {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      if (ctx) {
        canvas.width = 200;
        canvas.height = 50;
        ctx.textBaseline = 'top';
        ctx.font = '14px Arial';
        ctx.fillText('VirtEngine Fingerprint', 2, 2);
        components.push(`canvas:${canvas.toDataURL().slice(-50)}`);
      }
    } catch {
      // Canvas fingerprinting may be blocked
    }
  }

  // WebGL info
  if (typeof document !== 'undefined') {
    try {
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl');
      if (gl) {
        const renderer = gl.getParameter(gl.RENDERER);
        const vendor = gl.getParameter(gl.VENDOR);
        components.push(`webgl:${vendor}|${renderer}`);
      }
    } catch {
      // WebGL may not be available
    }
  }

  // Hash all components
  const fingerprintData = components.join('|');
  const fingerprintBytes = new TextEncoder().encode(fingerprintData);
  const hashBuffer = await crypto.subtle.digest('SHA-256', fingerprintBytes);
  const hashArray = new Uint8Array(hashBuffer);

  return bytesToHex(hashArray);
}

/**
 * Create a capture session ID
 */
export function createSessionId(): string {
  const crypto = getCrypto();
  const randomBytes = new Uint8Array(16);
  crypto.getRandomValues(randomBytes);
  const timestamp = Date.now().toString(36);
  const random = bytesToHex(randomBytes).slice(0, 16);
  return `cap_${timestamp}_${random}`;
}
