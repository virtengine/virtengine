import { describe, expect, it } from 'vitest';
import { buildUploadScopeMessage, normalizeScopeType } from '../../src/submission/transaction';


describe('submission transaction', () => {
  it('builds an upload scope message with expected fields', () => {
    const message = buildUploadScopeMessage({
      senderAddress: 've1sender',
      scopeId: 'scope_test_123',
      scopeType: 'selfie',
      envelope: {
        version: 1,
        algorithmId: 'X25519-XSALSA20-POLY1305',
        algorithmVersion: 1,
        recipientKeyIds: ['fingerprint-1'],
        recipientPublicKeys: [new Uint8Array([1, 2, 3])],
        encryptedKeys: [new Uint8Array([4, 5, 6])],
        nonce: new Uint8Array([7, 8, 9]),
        ciphertext: new Uint8Array([10, 11, 12]),
        senderSignature: new Uint8Array([13, 14, 15]),
        senderPubKey: new Uint8Array([16, 17, 18]),
        metadata: { capture_type: '2' },
      },
      metadata: {
        salt: new Uint8Array([21, 22, 23]),
        saltHash: new Uint8Array([24, 25, 26]),
        deviceFingerprint: 'device-abc',
        clientId: 'client-xyz',
        clientSignature: new Uint8Array([27, 28, 29]),
        userSignature: new Uint8Array([30, 31, 32]),
        payloadHash: new Uint8Array([33, 34, 35]),
        uploadNonce: new Uint8Array([36, 37, 38]),
        captureTimestamp: 1700000000,
        geoHint: 'US',
      },
    });

    expect(message.typeUrl).toBe('/virtengine.veid.v1.MsgUploadScope');
    const value = message.value as Record<string, any>;
    const encryptedPayload = value.encryptedPayload as Record<string, any>;
    expect(value.sender).toBe('ve1sender');
    expect(value.scopeId).toBe('scope_test_123');
    expect(value.scopeType).toBe(2);
    expect(encryptedPayload.algorithmId).toBe('X25519-XSALSA20-POLY1305');
    expect(encryptedPayload.recipientKeyIds).toEqual(['fingerprint-1']);
    expect(value.deviceFingerprint).toBe('device-abc');
    expect(value.clientId).toBe('client-xyz');
    expect(value.captureTimestamp).toBe(1700000000);
    expect(value.geoHint).toBe('US');
  });

  it('normalizes scope types to numeric values', () => {
    expect(normalizeScopeType('id_document')).toBe(1);
    expect(normalizeScopeType('selfie')).toBe(2);
    expect(normalizeScopeType(7)).toBe(7);
  });
});
