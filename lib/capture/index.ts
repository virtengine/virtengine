/**
 * VirtEngine Capture Library
 * VE-210: Guided document + selfie capture with quality checks
 *
 * This library provides a complete capture experience for identity
 * document and selfie capture with:
 *
 * - Real-time quality validation (resolution, brightness, blur, skew, glare)
 * - Visual guidance overlay for optimal positioning
 * - Metadata stripping (EXIF/GPS removal) for privacy
 * - Signature packaging (client + user signatures with salt binding)
 * - Liveness detection for selfies (optional)
 *
 * @example
 * ```tsx
 * import { DocumentCapture, SelfieCapture } from '@virtengine/capture';
 *
 * function IdentityCapture() {
 *   return (
 *     <DocumentCapture
 *       documentType="id_card"
 *       onCapture={(result) => console.log('Captured:', result)}
 *       onError={(error) => console.error('Error:', error)}
 *       clientKeyProvider={clientKeyProvider}
 *       userKeyProvider={userKeyProvider}
 *     />
 *   );
 * }
 * ```
 *
 * @packageDocumentation
 */

// ============================================================================
// Components
// ============================================================================

export { DocumentCapture } from './components/DocumentCapture';
export type { DocumentCaptureProps } from './types/capture';

export { SelfieCapture } from './components/SelfieCapture';
export type { SelfieCaptureProps } from './types/capture';

export { CaptureGuidance } from './components/CaptureGuidance';
export type { CaptureGuidanceProps } from './components/CaptureGuidance';

export { QualityFeedback, QualityIndicator } from './components/QualityFeedback';
export type { QualityFeedbackProps, QualityIndicatorProps } from './components/QualityFeedback';

// ============================================================================
// Hooks
// ============================================================================

export { useCamera } from './hooks/useCamera';
export type { UseCameraOptions, UseCameraReturn } from './hooks/useCamera';

export { useQualityCheck, useStableQualityFeedback } from './hooks/useQualityCheck';
export type { UseQualityCheckOptions, UseQualityCheckReturn } from './hooks/useQualityCheck';

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
} from './utils/quality-checks';

// Metadata stripping
export { stripMetadata, hasMetadata, STRIPPED_METADATA_TYPES } from './utils/metadata-strip';
export type { MetadataStripResult } from './utils/metadata-strip';

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
} from './utils/salt-generator';
export type { SaltOptions, GeneratedSalt } from './utils/salt-generator';

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
} from './utils/signature';
export type { SignatureOptions } from './utils/signature';

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
} from './types/capture';

export { DEFAULT_QUALITY_THRESHOLDS } from './types/capture';

// ============================================================================
// Version
// ============================================================================

export const VERSION = '1.0.0';

// ============================================================================
// Submission (On-chain VEID)
// ============================================================================

export {
  submitCaptureScope,
  createUploadNonce,
  createUploadSignatures,
  createUploadMetadata,
  encryptPayloadForRecipients,
  fetchValidatorEncryptionKeys,
  buildUploadScopeMessage,
  createCosmjsBroadcaster,
  normalizeScopeType,
  createScopeId,
} from './src/submission';

export type {
  SubmissionRequest,
  SubmissionResult,
  SubmissionUpdate,
  SubmissionStatus,
  SubmissionErrorType,
  ScopeTypeInput,
  ValidatorRecipientKey,
  EncryptedPayloadEnvelope,
  UploadMetadata,
  TxBroadcastResult,
  TxBroadcaster,
  UploadScopeMessage,
  ScorePollOptions,
  ApprovedClientCheckOptions,
} from './src/submission';

export {
  SubmissionError,
  CaptureValidationError,
  EncryptionError,
  SigningError,
  BroadcastError,
  SubmissionTimeoutError,
} from './src/submission';
