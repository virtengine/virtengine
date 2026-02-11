/**
 * Transaction helpers for VEID submission pipeline.
 */

import type { EncodeObject } from '@cosmjs/proto-signing';
import type { SigningStargateClient, StdFee } from '@cosmjs/stargate';
import { bytesToHex } from '../../utils/salt-generator';
import type {
  EncryptedPayloadEnvelope,
  ScopeTypeInput,
  TxBroadcastResult,
  TxBroadcaster,
  UploadMetadata,
  UploadScopeMessage,
} from './types';

export const MSG_UPLOAD_SCOPE_TYPE_URL = '/virtengine.veid.v1.MsgUploadScope';

const SCOPE_TYPE_MAP: Record<string, number> = {
  id_document: 1,
  selfie: 2,
  face_video: 3,
  biometric: 4,
  sso_metadata: 5,
  email_proof: 6,
  sms_proof: 7,
  domain_verify: 8,
  ad_sso: 9,
  biometric_hardware: 10,
  device_attestation: 11,
};

export function normalizeScopeType(scopeType: ScopeTypeInput): number {
  if (typeof scopeType === 'number') return scopeType;
  return SCOPE_TYPE_MAP[scopeType] ?? 0;
}

export function createScopeId(prefix = 'scope'): string {
  const crypto = globalThis.crypto;
  if (!crypto?.getRandomValues) {
    throw new Error('Web Crypto API not available');
  }
  const randomBytes = new Uint8Array(12);
  crypto.getRandomValues(randomBytes);
  const timestamp = Date.now().toString(36);
  const randomHex = bytesToHex(randomBytes);
  return `${prefix}_${timestamp}_${randomHex}`;
}

function normalizeEnvelope(envelope: EncryptedPayloadEnvelope): EncryptedPayloadEnvelope {
  return {
    ...envelope,
    recipientPublicKeys: envelope.recipientPublicKeys ?? [],
    encryptedKeys: envelope.encryptedKeys ?? [],
    metadata: envelope.metadata ?? {},
  };
}

export function buildUploadScopeMessage(options: {
  senderAddress: string;
  scopeId: string;
  scopeType: ScopeTypeInput;
  envelope: EncryptedPayloadEnvelope;
  metadata: UploadMetadata;
}): UploadScopeMessage {
  const normalizedEnvelope = normalizeEnvelope(options.envelope);
  return {
    typeUrl: MSG_UPLOAD_SCOPE_TYPE_URL,
    value: {
      sender: options.senderAddress,
      scopeId: options.scopeId,
      scopeType: normalizeScopeType(options.scopeType),
      encryptedPayload: {
        version: normalizedEnvelope.version,
        algorithmId: normalizedEnvelope.algorithmId,
        algorithmVersion: normalizedEnvelope.algorithmVersion,
        recipientKeyIds: normalizedEnvelope.recipientKeyIds,
        recipientPublicKeys: normalizedEnvelope.recipientPublicKeys ?? [],
        encryptedKeys: normalizedEnvelope.encryptedKeys ?? [],
        nonce: normalizedEnvelope.nonce,
        ciphertext: normalizedEnvelope.ciphertext,
        senderSignature: normalizedEnvelope.senderSignature,
        senderPubKey: normalizedEnvelope.senderPubKey,
        metadata: normalizedEnvelope.metadata ?? {},
      },
      salt: options.metadata.salt,
      deviceFingerprint: options.metadata.deviceFingerprint,
      clientId: options.metadata.clientId,
      clientSignature: options.metadata.clientSignature,
      userSignature: options.metadata.userSignature,
      payloadHash: options.metadata.payloadHash,
      captureTimestamp: options.metadata.captureTimestamp,
      geoHint: options.metadata.geoHint ?? '',
    },
  };
}

export function createCosmjsBroadcaster(options: {
  client: SigningStargateClient;
  senderAddress: string;
  fee?: StdFee | 'auto' | number;
  memo?: string;
}): TxBroadcaster {
  return {
    broadcast: async (msg: UploadScopeMessage, memo?: string): Promise<TxBroadcastResult> => {
      const result = await options.client.signAndBroadcast(
        options.senderAddress,
        [msg as EncodeObject],
        options.fee ?? 'auto',
        memo ?? options.memo ?? ''
      );
      return {
        txHash: result.transactionHash,
        code: result.code,
        rawLog: result.rawLog,
        gasUsed: result.gasUsed ? Number(result.gasUsed) : undefined,
        gasWanted: result.gasWanted ? Number(result.gasWanted) : undefined,
      };
    },
  };
}
