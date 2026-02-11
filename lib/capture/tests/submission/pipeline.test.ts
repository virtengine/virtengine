import { describe, expect, it } from 'vitest';
import nacl from 'tweetnacl';
import { submitCaptureScope } from '../../src/submission/pipeline';
import { computeKeyFingerprint } from '../../src/submission/encryption';
import { computeHash } from '../../utils/signature';
import type { SubmissionUpdate } from '../../src/submission/types';
import type { CaptureResult } from '../../types/capture';

const clientKeyProvider = {
  getClientId: () => Promise.resolve('client-portal'),
  getClientVersion: () => Promise.resolve('1.0.0'),
  sign: async (data: Uint8Array) => computeHash(data, 'SHA-256'),
  getPublicKey: () => Promise.resolve(new Uint8Array(32)),
  getKeyType: () => Promise.resolve('ed25519' as const),
};

const userKeyProvider = {
  getAccountAddress: () => Promise.resolve('ve1user'),
  sign: async (data: Uint8Array) => computeHash(data, 'SHA-256'),
  getPublicKey: () => Promise.resolve(new Uint8Array(32)),
  getKeyType: () => Promise.resolve('ed25519' as const),
};

describe('submission pipeline', () => {
  it('submits capture payload through encryption and broadcast', async () => {
    const recipient = nacl.box.keyPair();
    const fingerprint = await computeKeyFingerprint(recipient.publicKey);
    const capture: CaptureResult = {
      imageBlob: new Blob([new Uint8Array([0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43])], {
        type: 'image/jpeg',
      }),
      salt: new Uint8Array(32).map((_, index) => (index + 5) % 251),
      payloadHash: new Uint8Array(32).fill(1),
      clientSignature: new Uint8Array(64).fill(2),
      userSignature: new Uint8Array(64).fill(3),
      metadata: {
        deviceFingerprint: 'device-xyz',
        clientVersion: '1.0.0',
        capturedAt: new Date().toISOString(),
        documentType: 'id_card',
        qualityScore: 92,
        sessionId: 'session-abc',
      },
      dimensions: { width: 1024, height: 768 },
      mimeType: 'image/jpeg',
    };

    const updates: SubmissionUpdate[] = [];
    const result = await submitCaptureScope({
      capture,
      scopeType: 'id_document',
      senderAddress: 've1user',
      restEndpoint: 'https://example.test',
      clientKeyProvider,
      userKeyProvider,
      broadcaster: {
        broadcast: async (msg) => {
          expect(msg.typeUrl).toBe('/virtengine.veid.v1.MsgUploadScope');
          return {
            txHash: 'tx123',
            code: 0,
            rawLog: '',
            gasUsed: 1000,
            gasWanted: 1200,
          };
        },
      },
      validatorKeys: [
        {
          validatorAddress: 'val1',
          publicKey: recipient.publicKey,
          keyFingerprint: fingerprint,
          keyVersion: 1,
          algorithmId: 'X25519-XSALSA20-POLY1305',
        },
      ],
      approvedClientCheck: { enabled: false },
      onStatus: (update) => updates.push(update),
    });

    expect(result.txHash).toBe('tx123');
    expect(result.scopeId).toContain('scope_');
    expect(result.status).toBe('confirmed');
    expect(updates.some((update) => update.status === 'encrypting')).toBe(true);
    expect(updates.some((update) => update.status === 'signing')).toBe(true);
    expect(updates.some((update) => update.status === 'broadcasting')).toBe(true);
    expect(updates.some((update) => update.status === 'confirmed')).toBe(true);
  });
});
