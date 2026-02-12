/**
 * VEID transaction builders.
 */

import type { ChainTxMessage } from "../types";

export interface UploadScopeInput {
  sender: string;
  scopeId: string;
  scopeType: string | number;
  encryptedPayload: Record<string, unknown>;
  salt: Uint8Array | string;
  deviceFingerprint: string;
  clientId: string;
  clientSignature: Uint8Array | string;
  userSignature: Uint8Array | string;
  payloadHash: Uint8Array | string;
}

export interface RequestVerificationInput {
  sender: string;
  scopeId: string;
}

/**
 * Build MsgUploadScope.
 */
export function buildMsgUploadScope(input: UploadScopeInput): ChainTxMessage {
  return {
    typeUrl: "/virtengine.veid.v1.MsgUploadScope",
    value: {
      sender: input.sender,
      scope_id: input.scopeId,
      scope_type: input.scopeType,
      encrypted_payload: input.encryptedPayload,
      salt: input.salt,
      device_fingerprint: input.deviceFingerprint,
      client_id: input.clientId,
      client_signature: input.clientSignature,
      user_signature: input.userSignature,
      payload_hash: input.payloadHash,
    },
  };
}

/**
 * Build MsgRequestVerification.
 */
export function buildMsgRequestVerification(
  input: RequestVerificationInput,
): ChainTxMessage {
  return {
    typeUrl: "/virtengine.veid.v1.MsgRequestVerification",
    value: {
      sender: input.sender,
      scope_id: input.scopeId,
    },
  };
}
