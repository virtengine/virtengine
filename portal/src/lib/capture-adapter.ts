/**
 * Capture Adapter
 * Re-exports from lib/capture for Next.js integration
 *
 * This adapter provides a clean import path for capture library components,
 * hooks, and utilities within the Next.js application.
 */

// ============================================================================
// Components
// ============================================================================

export { DocumentCapture } from '../../../lib/capture';
export type { DocumentCaptureProps } from '../../../lib/capture';

export { SelfieCapture } from '../../../lib/capture';
export type { SelfieCaptureProps } from '../../../lib/capture';

export { CaptureGuidance } from '../../../lib/capture';
export type { CaptureGuidanceProps } from '../../../lib/capture';

export { QualityFeedback, QualityIndicator } from '../../../lib/capture';
export type { QualityFeedbackProps, QualityIndicatorProps } from '../../../lib/capture';

// ============================================================================
// Hooks
// ============================================================================

export { useCamera } from '../../../lib/capture';
export type { UseCameraOptions, UseCameraReturn } from '../../../lib/capture';

export { useQualityCheck, useStableQualityFeedback } from '../../../lib/capture';
export type { UseQualityCheckOptions, UseQualityCheckReturn } from '../../../lib/capture';

// ============================================================================
// Utilities
// ============================================================================

// Quality checks
export {
  performQualityChecks,
  quickQualityAssessment,
  checkResolution,
  checkBrightness,
  checkBlur,
  checkSkew,
  checkGlare,
  checkNoise,
  getImageDataFromVideo,
  getImageDataFromBlob,
} from '../../../lib/capture';

// Metadata stripping
export { stripMetadata, hasMetadata, STRIPPED_METADATA_TYPES } from '../../../lib/capture';
export type { MetadataStripResult } from '../../../lib/capture';

// Salt generation
export {
  generateSalt,
  generateDeviceBoundSalt,
  generateSessionBoundSalt,
  verifySalt,
  bytesToHex,
  hexToBytes,
  bytesToBase64,
  base64ToBytes,
  concatBytes,
  DEFAULT_SALT_LENGTH,
} from '../../../lib/capture';
export type { SaltOptions, GeneratedSalt } from '../../../lib/capture';

// Signatures
export {
  computeHash,
  createPayloadHash,
  createClientSignature,
  createUserSignature,
  createSignaturePackage,
  serializeSignaturePackage,
  generateDeviceFingerprint,
  createSessionId,
} from '../../../lib/capture';
export type { SignatureOptions } from '../../../lib/capture';

// ============================================================================
// Types
// ============================================================================

export type {
  // Document types
  DocumentType,
  SelfieCaptureMode,
  DocumentSide,
  // Quality types
  QualityIssueType,
  QualityIssueSeverity,
  QualityIssue,
  QualityCheckResult,
  QualityCheckDetail,
  QualityThresholds,
  // Camera types
  CameraDevice,
  CameraConstraints,
  CameraState,
  CameraErrorType,
  CameraError,
  // Capture result types
  CaptureResult,
  SelfieResult,
  CaptureMetadata,
  LivenessCheckResult,
  // Key provider types
  ClientKeyProvider,
  UserKeyProvider,
  // Error types
  CaptureErrorType,
  CaptureError,
  // Guidance types
  GuidanceState,
  // Signature types
  SignaturePackage,
} from '../../../lib/capture';

export { DEFAULT_QUALITY_THRESHOLDS, VERSION } from '../../../lib/capture';
