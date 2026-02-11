/**
 * Submission pipeline types for VEID on-chain scope uploads.
 */

import type { CaptureResult, ClientKeyProvider, UserKeyProvider } from '../../types/capture';

export type SubmissionStatus =
  | 'pending'
  | 'encrypting'
  | 'signing'
  | 'broadcasting'
  | 'confirmed'
  | 'scored'
  | 'failed';

export interface SubmissionUpdate {
  status: SubmissionStatus;
  message?: string;
  scopeId?: string;
  txHash?: string;
  score?: number;
  error?: SubmissionError;
}

export interface SubmissionRequest {
  capture: CaptureResult;
  scopeType: ScopeTypeInput;
  senderAddress: string;
  restEndpoint: string;
  clientKeyProvider: ClientKeyProvider;
  userKeyProvider: UserKeyProvider;
  broadcaster: TxBroadcaster;
  memo?: string;
  scopeId?: string;
  geoHint?: string;
  uploadNonce?: Uint8Array;
  validatorKeys?: ValidatorRecipientKey[];
  fetchJson?: ChainJsonFetcher;
  onStatus?: (update: SubmissionUpdate) => void;
  scorePoll?: ScorePollOptions;
  approvedClientCheck?: ApprovedClientCheckOptions;
}

export interface SubmissionResult {
  scopeId: string;
  txHash: string;
  payloadHash: Uint8Array;
  envelope: EncryptedPayloadEnvelope;
  score?: number;
  status: SubmissionStatus;
}

export interface ScorePollOptions {
  enabled?: boolean;
  timeoutMs?: number;
  intervalMs?: number;
}

export interface ApprovedClientCheckOptions {
  enabled?: boolean;
  allowedClientIds?: string[];
}

export type ScopeTypeInput =
  | 'id_document'
  | 'selfie'
  | 'face_video'
  | 'biometric'
  | 'sso_metadata'
  | 'email_proof'
  | 'sms_proof'
  | 'domain_verify'
  | 'ad_sso'
  | 'biometric_hardware'
  | 'device_attestation'
  | number;

export interface UploadScopeMessage {
  typeUrl: string;
  value: Record<string, unknown>;
}

export interface TxBroadcastResult {
  txHash: string;
  code: number;
  rawLog?: string;
  gasUsed?: number;
  gasWanted?: number;
}

export interface TxBroadcaster {
  broadcast: (msg: UploadScopeMessage, memo?: string) => Promise<TxBroadcastResult>;
}

export interface ChainJsonFetcher {
  (path: string, params?: Record<string, string | number | boolean | undefined>): Promise<unknown>;
}

export interface ValidatorRecipientKey {
  validatorAddress: string;
  publicKey: Uint8Array;
  keyFingerprint: string;
  keyVersion: number;
  algorithmId: string;
}

export interface EncryptedPayloadEnvelope {
  version: number;
  algorithmId: string;
  algorithmVersion: number;
  recipientKeyIds: string[];
  recipientPublicKeys?: Uint8Array[];
  encryptedKeys?: Uint8Array[];
  wrappedKeys?: WrappedKeyEntry[];
  nonce: Uint8Array;
  ciphertext: Uint8Array;
  senderSignature: Uint8Array;
  senderPubKey: Uint8Array;
  metadata?: Record<string, string>;
}

export interface WrappedKeyEntry {
  recipientId: string;
  wrappedKey: Uint8Array;
  algorithm?: string;
  ephemeralPubKey?: Uint8Array;
}

export interface UploadMetadata {
  salt: Uint8Array;
  saltHash: Uint8Array;
  deviceFingerprint: string;
  clientId: string;
  clientSignature: Uint8Array;
  userSignature: Uint8Array;
  payloadHash: Uint8Array;
  uploadNonce: Uint8Array;
  captureTimestamp: number;
  geoHint: string;
}

export type SubmissionErrorType =
  | 'capture_validation_failed'
  | 'encryption_failed'
  | 'signing_rejected'
  | 'broadcast_failed'
  | 'timeout';

export class SubmissionError extends Error {
  code: SubmissionErrorType;
  details?: Record<string, unknown>;

  constructor(code: SubmissionErrorType, message: string, details?: Record<string, unknown>) {
    super(message);
    this.name = 'SubmissionError';
    this.code = code;
    this.details = details;
  }
}

export class CaptureValidationError extends SubmissionError {
  constructor(message: string, details?: Record<string, unknown>) {
    super('capture_validation_failed', message, details);
    this.name = 'CaptureValidationError';
  }
}

export class EncryptionError extends SubmissionError {
  constructor(message: string, details?: Record<string, unknown>) {
    super('encryption_failed', message, details);
    this.name = 'EncryptionError';
  }
}

export class SigningError extends SubmissionError {
  constructor(message: string, details?: Record<string, unknown>) {
    super('signing_rejected', message, details);
    this.name = 'SigningError';
  }
}

export class BroadcastError extends SubmissionError {
  constructor(message: string, details?: Record<string, unknown>) {
    super('broadcast_failed', message, details);
    this.name = 'BroadcastError';
  }
}

export class SubmissionTimeoutError extends SubmissionError {
  constructor(message: string, details?: Record<string, unknown>) {
    super('timeout', message, details);
    this.name = 'SubmissionTimeoutError';
  }
}
