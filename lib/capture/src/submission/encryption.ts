/**
 * Encryption helpers for VEID submission pipeline.
 */

import nacl from 'tweetnacl';
import { bytesToHex, concatBytes } from '../../utils/salt-generator';
import { computeHash } from '../../utils/signature';
import type { EncryptedPayloadEnvelope, ValidatorRecipientKey, WrappedKeyEntry } from './types';

export const ENVELOPE_VERSION = 1;
export const ALGORITHM_ID = 'X25519-XSALSA20-POLY1305';
export const ALGORITHM_VERSION = 1;

const NONCE_SIZE = 24;
const KEY_SIZE = 32;
const FINGERPRINT_SIZE = 20;

function toUint32Bytes(value: number): Uint8Array {
  const bytes = new Uint8Array(4);
  bytes[0] = (value >>> 24) & 0xff;
  bytes[1] = (value >>> 16) & 0xff;
  bytes[2] = (value >>> 8) & 0xff;
  bytes[3] = value & 0xff;
  return bytes;
}

export async function computeKeyFingerprint(publicKey: Uint8Array): Promise<string> {
  const hash = await computeHash(publicKey, 'SHA-256');
  return bytesToHex(hash.slice(0, FINGERPRINT_SIZE));
}

export function formatRecipientKeyId(fingerprint: string, version: number): string {
  if (!version) return fingerprint;
  return `${fingerprint}:v${version}`;
}

export function normalizeRecipientKeyId(keyId: string): string {
  const splitIndex = keyId.indexOf(':v');
  if (splitIndex === -1) return keyId;
  return keyId.slice(0, splitIndex);
}

export async function buildEnvelopeSigningPayload(envelope: EncryptedPayloadEnvelope): Promise<Uint8Array> {
  const parts: Uint8Array[] = [];
  parts.push(toUint32Bytes(envelope.version));
  parts.push(new TextEncoder().encode(envelope.algorithmId));
  parts.push(toUint32Bytes(envelope.algorithmVersion));
  parts.push(envelope.ciphertext);
  parts.push(envelope.nonce);

  for (const recipientId of envelope.recipientKeyIds) {
    parts.push(new TextEncoder().encode(recipientId));
  }

  return computeHash(concatBytes(...parts), 'SHA-256');
}

export async function signEnvelope(envelope: EncryptedPayloadEnvelope): Promise<Uint8Array> {
  const payload = await buildEnvelopeSigningPayload(envelope);
  const binding = concatBytes(payload, envelope.senderPubKey);
  return computeHash(binding, 'SHA-256');
}

function assertKeySize(key: Uint8Array, label: string) {
  if (key.length !== KEY_SIZE) {
    throw new Error(`${label} must be ${KEY_SIZE} bytes`);
  }
}

export async function encryptPayloadForRecipients(
  plaintext: Uint8Array,
  recipients: ValidatorRecipientKey[]
): Promise<EncryptedPayloadEnvelope> {
  if (!recipients.length) {
    throw new Error('At least one recipient key is required');
  }

  const senderKeyPair = nacl.box.keyPair();
  const recipientKeyIds: string[] = [];
  const recipientPublicKeys: Uint8Array[] = [];
  const encryptedKeys: Uint8Array[] = [];
  const wrappedKeys: WrappedKeyEntry[] = [];

  const metadata: Record<string, string> = {};

  if (recipients.length === 1) {
    const recipient = recipients[0];
    assertKeySize(recipient.publicKey, 'Recipient public key');

    const nonce = nacl.randomBytes(NONCE_SIZE);
    const ciphertext = nacl.box(plaintext, nonce, recipient.publicKey, senderKeyPair.secretKey);

    const recipientId = formatRecipientKeyId(recipient.keyFingerprint, recipient.keyVersion);
    recipientKeyIds.push(recipientId);
    recipientPublicKeys.push(recipient.publicKey);

    const envelope: EncryptedPayloadEnvelope = {
      version: ENVELOPE_VERSION,
      algorithmId: ALGORITHM_ID,
      algorithmVersion: ALGORITHM_VERSION,
      recipientKeyIds,
      recipientPublicKeys,
      nonce,
      ciphertext,
      senderSignature: new Uint8Array(),
      senderPubKey: senderKeyPair.publicKey,
      metadata,
    };

    envelope.senderSignature = await signEnvelope(envelope);
    return envelope;
  }

  const dek = nacl.randomBytes(KEY_SIZE);
  const dataNonce = nacl.randomBytes(NONCE_SIZE);
  const ciphertext = nacl.secretbox(plaintext, dataNonce, dek);

  recipients.forEach((recipient, index) => {
    assertKeySize(recipient.publicKey, 'Recipient public key');
    const keyNonce = nacl.randomBytes(NONCE_SIZE);
    const encryptedDek = nacl.box(dek, keyNonce, recipient.publicKey, senderKeyPair.secretKey);
    const wrappedKey = concatBytes(keyNonce, encryptedDek);

    const recipientId = formatRecipientKeyId(recipient.keyFingerprint, recipient.keyVersion);
    recipientKeyIds[index] = recipientId;
    recipientPublicKeys[index] = recipient.publicKey;
    encryptedKeys[index] = wrappedKey;
    wrappedKeys[index] = {
      recipientId,
      wrappedKey,
    };
  });

  metadata._mode = 'multi-recipient';

  const envelope: EncryptedPayloadEnvelope = {
    version: ENVELOPE_VERSION,
    algorithmId: ALGORITHM_ID,
    algorithmVersion: ALGORITHM_VERSION,
    recipientKeyIds,
    recipientPublicKeys,
    encryptedKeys,
    wrappedKeys,
    nonce: dataNonce,
    ciphertext,
    senderSignature: new Uint8Array(),
    senderPubKey: senderKeyPair.publicKey,
    metadata,
  };

  envelope.senderSignature = await signEnvelope(envelope);
  return envelope;
}
