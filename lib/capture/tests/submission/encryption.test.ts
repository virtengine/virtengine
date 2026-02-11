import { describe, expect, it } from 'vitest';
import nacl from 'tweetnacl';
import { computeKeyFingerprint, encryptPayloadForRecipients } from '../../src/submission/encryption';

function toRecipientKey(keyPair: nacl.BoxKeyPair, keyFingerprint: string) {
  return {
    validatorAddress: 'val1',
    publicKey: keyPair.publicKey,
    keyFingerprint,
    keyVersion: 1,
    algorithmId: 'X25519-XSALSA20-POLY1305',
  };
}

describe('submission encryption', () => {
  it('encrypts and decrypts for a single recipient', async () => {
    const recipient = nacl.box.keyPair();
    const fingerprint = await computeKeyFingerprint(recipient.publicKey);
    const plaintext = new Uint8Array([1, 2, 3, 4, 5, 6]);

    const envelope = await encryptPayloadForRecipients(plaintext, [toRecipientKey(recipient, fingerprint)]);

    const decrypted = nacl.box.open(
      envelope.ciphertext,
      envelope.nonce,
      envelope.senderPubKey,
      recipient.secretKey
    );

    expect(decrypted).not.toBeNull();
    expect(Array.from(decrypted ?? [])).toEqual(Array.from(plaintext));
    expect(envelope.recipientKeyIds[0]).toContain(fingerprint);
  });

  it('encrypts and decrypts for multiple recipients', async () => {
    const recipientA = nacl.box.keyPair();
    const recipientB = nacl.box.keyPair();
    const fingerprintA = await computeKeyFingerprint(recipientA.publicKey);
    const fingerprintB = await computeKeyFingerprint(recipientB.publicKey);
    const plaintext = new Uint8Array([9, 8, 7, 6, 5, 4, 3]);

    const envelope = await encryptPayloadForRecipients(plaintext, [
      toRecipientKey(recipientA, fingerprintA),
      toRecipientKey(recipientB, fingerprintB),
    ]);

    const wrapped = envelope.wrappedKeys?.[0]?.wrappedKey ?? envelope.encryptedKeys?.[0];
    expect(wrapped).toBeDefined();

    const keyNonce = wrapped!.slice(0, 24);
    const encryptedDek = wrapped!.slice(24);
    const dek = nacl.box.open(encryptedDek, keyNonce, envelope.senderPubKey, recipientA.secretKey);
    expect(dek).not.toBeNull();

    const decrypted = nacl.secretbox.open(envelope.ciphertext, envelope.nonce, dek!);
    expect(decrypted).not.toBeNull();
    expect(Array.from(decrypted ?? [])).toEqual(Array.from(plaintext));
  });
});
