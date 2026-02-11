import { describe, expect, it } from 'vitest';
import {
  createClientSigningPayload,
  createUserSigningPayload,
  createUploadMetadata,
  createUploadNonce,
  createUploadSignatures,
} from '../../src/submission/envelope';
import { computeHash } from '../../utils/signature';

describe('submission envelope', () => {
  it('builds client and user signing payloads', async () => {
    const salt = new Uint8Array([1, 2, 3, 4, 5, 6, 7, 8]);
    const payloadHash = new Uint8Array(32).fill(9);
    const nonce = createUploadNonce(16);

    const clientPayload = await createClientSigningPayload(
      salt,
      'device-abc',
      'client-123',
      payloadHash,
      nonce
    );
    const clientSignature = await computeHash(clientPayload, 'SHA-256');
    const userPayload = await createUserSigningPayload(clientPayload, clientSignature);

    expect(clientPayload.length).toBe(32);
    expect(userPayload.length).toBe(32);
  });

  it('creates upload signatures and metadata with chained payloads', async () => {
    let clientPayloadSeen: Uint8Array | null = null;
    let userPayloadSeen: Uint8Array | null = null;

    const clientKeyProvider = {
      getClientId: () => Promise.resolve('client-123'),
      getClientVersion: () => Promise.resolve('1.0.0'),
      sign: async (data: Uint8Array) => {
        clientPayloadSeen = data;
        return computeHash(data, 'SHA-256');
      },
      getPublicKey: () => Promise.resolve(new Uint8Array(32)),
      getKeyType: () => Promise.resolve('ed25519' as const),
    };

    const userKeyProvider = {
      getAccountAddress: () => Promise.resolve('ve1user'),
      sign: async (data: Uint8Array) => {
        userPayloadSeen = data;
        return computeHash(data, 'SHA-256');
      },
      getPublicKey: () => Promise.resolve(new Uint8Array(32)),
      getKeyType: () => Promise.resolve('ed25519' as const),
    };

    const salt = new Uint8Array(32).map((_, index) => (index + 1) % 255);
    const payloadHash = new Uint8Array(32).fill(4);
    const uploadNonce = createUploadNonce(16);

    const { clientSignature, userSignature } = await createUploadSignatures({
      salt,
      deviceFingerprint: 'device-abc',
      clientId: 'client-123',
      payloadHash,
      uploadNonce,
      clientKeyProvider,
      userKeyProvider,
    });

    expect(clientSignature.length).toBe(32);
    expect(userSignature.length).toBe(32);
    expect(clientPayloadSeen).not.toBeNull();
    expect(userPayloadSeen).not.toBeNull();

    const metadata = await createUploadMetadata({
      salt,
      deviceFingerprint: 'device-abc',
      clientId: 'client-123',
      clientSignature,
      userSignature,
      payloadHash,
      uploadNonce,
      captureTimestamp: 1700000000,
      geoHint: 'US',
    });

    expect(metadata.clientId).toBe('client-123');
    expect(metadata.saltHash.length).toBe(32);
    expect(metadata.payloadHash).toEqual(payloadHash);
    expect(metadata.geoHint).toBe('US');
  });
});
