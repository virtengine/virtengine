/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access */
/**
 * Capture Adapter
 * Re-exports from @virtengine/capture for Next.js integration.
 */

import * as capture from '@virtengine/capture';

export const DocumentCapture = (capture as any).DocumentCapture;
export const CaptureGuidance = (capture as any).CaptureGuidance;
export const SelfieCapture = (capture as any).SelfieCapture;
export const submitCaptureScope = (capture as any).submitCaptureScope;
export const createUploadNonce = (capture as any).createUploadNonce;

export type DocumentType = 'id_card' | 'passport' | 'drivers_license';
export type DocumentSide = 'front' | 'back';
export interface CaptureResult {
  imageBlob: Blob;
  salt: Uint8Array;
  payloadHash: Uint8Array;
  clientSignature: Uint8Array;
  userSignature: Uint8Array;
  metadata: { sessionId: string; [key: string]: unknown };
  dimensions: { width: number; height: number };
  mimeType: string;
}
export interface CaptureError {
  type:
    | 'camera_error'
    | 'quality_check_failed'
    | 'metadata_strip_failed'
    | 'signing_failed'
    | 'validation_failed'
    | 'timeout'
    | 'cancelled'
    | 'unknown';
  message: string;
  details?: Record<string, unknown>;
  originalError?: Error;
}
export type GuidanceState = any;
export interface ClientKeyProvider {
  getClientId(): Promise<string>;
  getClientVersion(): Promise<string>;
  sign(data: Uint8Array): Promise<Uint8Array>;
  getPublicKey(): Promise<Uint8Array>;
  getKeyType(): Promise<'ed25519' | 'secp256k1'>;
}
export interface UserKeyProvider {
  getAccountAddress(): Promise<string>;
  sign(data: Uint8Array): Promise<Uint8Array>;
  getPublicKey(): Promise<Uint8Array>;
  getKeyType(): Promise<'ed25519' | 'secp256k1'>;
}
export type SelfieCaptureMode = any;
export interface LivenessCheckResult {
  passed: boolean;
  providerId: string;
  providerVersion: string;
  challengeId: string;
  sessionId: string;
  score: number;
  challengeType: 'blink' | 'smile' | 'turn' | 'passive';
  challengeDurationMs: number;
  evidenceDigest: string;
}
export interface LivenessProvider {
  verify(request: {
    sessionId: string;
    challengeId: string;
    timeoutMs: number;
  }): Promise<LivenessCheckResult>;
}
export interface SelfieResult extends CaptureResult {
  livenessCheck?: LivenessCheckResult;
}
export type SubmissionRequest = any;
export type SubmissionResult = any;
export type SubmissionUpdate = any;
export type SubmissionStatus = any;
export type TxBroadcaster = any;
export type UploadScopeMessage = any;
export type ScopeTypeInput = any;
