/**
 * Upload metadata + signature helpers for VEID submission.
 */

import type { ClientKeyProvider, UserKeyProvider } from '../../types/capture';
import { concatBytes } from '../../utils/salt-generator';
import { computeHash } from '../../utils/signature';
import type { UploadMetadata } from './types';

const DEFAULT_UPLOAD_NONCE_LENGTH = 16;

function getCrypto(): Crypto {
  if (typeof window !== 'undefined' && window.crypto) {
    return window.crypto;
  }
  if (typeof globalThis !== 'undefined' && globalThis.crypto) {
    return globalThis.crypto;
  }
  throw new Error('Web Crypto API not available');
}

export function createUploadNonce(length = DEFAULT_UPLOAD_NONCE_LENGTH): Uint8Array {
  const crypto = getCrypto();
  const nonce = new Uint8Array(length);
  crypto.getRandomValues(nonce);
  return nonce;
}

export async function computeSaltHash(salt: Uint8Array): Promise<Uint8Array> {
  return computeHash(salt, 'SHA-256');
}

export async function createClientSigningPayload(
  salt: Uint8Array,
  deviceFingerprint: string,
  clientId: string,
  payloadHash: Uint8Array,
  uploadNonce?: Uint8Array
): Promise<Uint8Array> {
  const encoder = new TextEncoder();
  const components: Uint8Array[] = [
    salt,
    encoder.encode(deviceFingerprint),
    encoder.encode(clientId),
    payloadHash,
  ];
  if (uploadNonce && uploadNonce.length > 0) {
    components.push(uploadNonce);
  }

  const message = concatBytes(...components);
  return computeHash(message, 'SHA-256');
}

export async function createUserSigningPayload(
  clientSigningPayload: Uint8Array,
  clientSignature: Uint8Array
): Promise<Uint8Array> {
  const message = concatBytes(clientSigningPayload, clientSignature);
  return computeHash(message, 'SHA-256');
}

export async function createUploadSignatures(options: {
  salt: Uint8Array;
  deviceFingerprint: string;
  clientId: string;
  payloadHash: Uint8Array;
  uploadNonce: Uint8Array;
  clientKeyProvider: ClientKeyProvider;
  userKeyProvider: UserKeyProvider;
}): Promise<{ clientSignature: Uint8Array; userSignature: Uint8Array }> {
  const clientPayload = await createClientSigningPayload(
    options.salt,
    options.deviceFingerprint,
    options.clientId,
    options.payloadHash,
    options.uploadNonce
  );
  const clientSignature = await options.clientKeyProvider.sign(clientPayload);

  const userPayload = await createUserSigningPayload(clientPayload, clientSignature);
  const userSignature = await options.userKeyProvider.sign(userPayload);

  return { clientSignature, userSignature };
}

export async function createUploadMetadata(options: {
  salt: Uint8Array;
  deviceFingerprint: string;
  clientId: string;
  clientSignature: Uint8Array;
  userSignature: Uint8Array;
  payloadHash: Uint8Array;
  uploadNonce: Uint8Array;
  captureTimestamp: number;
  geoHint?: string;
}): Promise<UploadMetadata> {
  const saltHash = await computeSaltHash(options.salt);
  return {
    salt: options.salt,
    saltHash,
    deviceFingerprint: options.deviceFingerprint,
    clientId: options.clientId,
    clientSignature: options.clientSignature,
    userSignature: options.userSignature,
    payloadHash: options.payloadHash,
    uploadNonce: options.uploadNonce,
    captureTimestamp: options.captureTimestamp,
    geoHint: options.geoHint ?? '',
  };
}
