/**
 * VirtEngine Capture Library - Type Definitions
 * VE-210: Capture UX v1 - guided document + selfie capture
 *
 * @packageDocumentation
 */

// ============================================================================
// Document Types
// ============================================================================

/**
 * Supported document types for identity verification
 */
export type DocumentType = 'id_card' | 'passport' | 'drivers_license';

/**
 * Selfie capture mode
 */
export type SelfieCaptureMode = 'photo' | 'video';

/**
 * Document capture side
 */
export type DocumentSide = 'front' | 'back';

// ============================================================================
// Quality Check Types
// ============================================================================

/**
 * Types of quality issues that can be detected
 */
export type QualityIssueType =
  | 'blur'
  | 'dark'
  | 'bright'
  | 'skew'
  | 'resolution'
  | 'glare'
  | 'noise'
  | 'partial'
  | 'reflection';

/**
 * Severity level of a quality issue
 */
export type QualityIssueSeverity = 'warning' | 'error';

/**
 * Individual quality issue detected during capture
 */
export interface QualityIssue {
  /** Type of quality issue */
  type: QualityIssueType;
  /** Severity level */
  severity: QualityIssueSeverity;
  /** Human-readable message describing the issue */
  message: string;
  /** Actionable suggestion for the user */
  suggestion: string;
  /** Confidence score (0-1) */
  confidence: number;
}

/**
 * Result of quality validation checks
 */
export interface QualityCheckResult {
  /** Whether all quality checks passed */
  passed: boolean;
  /** Overall quality score (0-100) */
  score: number;
  /** List of issues found */
  issues: QualityIssue[];
  /** Individual check results */
  checks: {
    resolution: QualityCheckDetail;
    brightness: QualityCheckDetail;
    blur: QualityCheckDetail;
    skew: QualityCheckDetail;
    glare: QualityCheckDetail;
    noise: QualityCheckDetail;
  };
  /** Time taken for quality analysis (ms) */
  analysisTimeMs: number;
}

/**
 * Detail for individual quality check
 */
export interface QualityCheckDetail {
  /** Check passed */
  passed: boolean;
  /** Raw value from check */
  value: number;
  /** Threshold used */
  threshold: number;
  /** Description */
  description: string;
}

// ============================================================================
// Camera Types
// ============================================================================

/**
 * Camera device information
 */
export interface CameraDevice {
  deviceId: string;
  label: string;
  facing: 'user' | 'environment' | 'unknown';
}

/**
 * Camera stream constraints
 */
export interface CameraConstraints {
  /** Preferred camera facing mode */
  facingMode: 'user' | 'environment';
  /** Minimum resolution width */
  minWidth: number;
  /** Minimum resolution height */
  minHeight: number;
  /** Ideal resolution width */
  idealWidth: number;
  /** Ideal resolution height */
  idealHeight: number;
  /** Frame rate */
  frameRate: number;
}

/**
 * Camera state
 */
export interface CameraState {
  /** Whether camera is ready */
  isReady: boolean;
  /** Whether stream is active */
  isStreaming: boolean;
  /** Current error if any */
  error: CameraError | null;
  /** Active device */
  activeDevice: CameraDevice | null;
  /** Available devices */
  availableDevices: CameraDevice[];
  /** Current stream dimensions */
  dimensions: { width: number; height: number } | null;
}

/**
 * Camera error types
 */
export type CameraErrorType =
  | 'permission_denied'
  | 'not_found'
  | 'not_readable'
  | 'overconstrained'
  | 'security_error'
  | 'unknown';

/**
 * Camera error
 */
export interface CameraError {
  type: CameraErrorType;
  message: string;
  originalError?: Error;
}

// ============================================================================
// Capture Result Types
// ============================================================================

/**
 * Metadata about the capture
 */
export interface CaptureMetadata {
  /** Device fingerprint for binding */
  deviceFingerprint: string;
  /** Client version */
  clientVersion: string;
  /** ISO timestamp of capture */
  capturedAt: string;
  /** Document type captured */
  documentType: DocumentType | 'selfie';
  /** Quality score achieved */
  qualityScore: number;
  /** Document side (for documents) */
  documentSide?: DocumentSide;
  /** Camera device used */
  cameraDeviceId?: string;
  /** Capture session ID */
  sessionId: string;
}

/**
 * Capture result with all required signatures and bindings
 */
export interface CaptureResult {
  /** Stripped image blob (no EXIF/GPS) */
  imageBlob: Blob;
  /** Per-upload salt for binding */
  salt: Uint8Array;
  /** Hash of the encrypted payload */
  payloadHash: Uint8Array;
  /** Client signature (approved client) */
  clientSignature: Uint8Array;
  /** User signature */
  userSignature: Uint8Array;
  /** Capture metadata */
  metadata: CaptureMetadata;
  /** Image dimensions */
  dimensions: { width: number; height: number };
  /** MIME type of image */
  mimeType: string;
}

/**
 * Selfie capture result (extends CaptureResult for selfies)
 */
export interface SelfieResult extends CaptureResult {
  /** Liveness check result if performed */
  livenessCheck?: LivenessCheckResult;
}

/**
 * Liveness check result
 */
export interface LivenessCheckResult {
  /** Whether liveness was verified */
  passed: boolean;
  /** Liveness score (0-1) */
  score: number;
  /** Type of liveness challenge completed */
  challengeType: 'blink' | 'smile' | 'turn' | 'passive';
  /** Time taken for challenge */
  challengeDurationMs: number;
}

// ============================================================================
// Key Provider Types
// ============================================================================

/**
 * Client key provider for approved client signature
 * Compatible with VE-201 approved client requirements
 */
export interface ClientKeyProvider {
  /** Get the client ID */
  getClientId(): Promise<string>;
  /** Get the client version */
  getClientVersion(): Promise<string>;
  /** Sign data with client key */
  sign(data: Uint8Array): Promise<Uint8Array>;
  /** Get public key for verification */
  getPublicKey(): Promise<Uint8Array>;
  /** Key type (ed25519 or secp256k1) */
  getKeyType(): Promise<'ed25519' | 'secp256k1'>;
}

/**
 * User key provider for user signature
 */
export interface UserKeyProvider {
  /** Get the user's account address */
  getAccountAddress(): Promise<string>;
  /** Sign data with user key */
  sign(data: Uint8Array): Promise<Uint8Array>;
  /** Get public key for verification */
  getPublicKey(): Promise<Uint8Array>;
  /** Key type */
  getKeyType(): Promise<'ed25519' | 'secp256k1'>;
}

// ============================================================================
// Error Types
// ============================================================================

/**
 * Capture error types
 */
export type CaptureErrorType =
  | 'camera_error'
  | 'quality_check_failed'
  | 'metadata_strip_failed'
  | 'signing_failed'
  | 'validation_failed'
  | 'timeout'
  | 'cancelled'
  | 'unknown';

/**
 * Capture error
 */
export interface CaptureError {
  type: CaptureErrorType;
  message: string;
  details?: Record<string, unknown>;
  qualityIssues?: QualityIssue[];
  originalError?: Error;
}

// ============================================================================
// Component Props Types
// ============================================================================

/**
 * Guidance overlay state
 */
export interface GuidanceState {
  /** Whether document is detected in frame */
  documentDetected: boolean;
  /** Whether alignment is acceptable */
  alignmentOk: boolean;
  /** Current quality issues (real-time) */
  currentIssues: QualityIssue[];
  /** Ready to capture */
  readyToCapture: boolean;
  /** Current instruction message */
  instruction: string;
}

/**
 * Document capture component props
 */
export interface DocumentCaptureProps {
  /** Type of document to capture */
  documentType: DocumentType;
  /** Which side of document to capture */
  documentSide?: DocumentSide;
  /** Callback when capture is successful */
  onCapture: (result: CaptureResult) => void;
  /** Callback when an error occurs */
  onError: (error: CaptureError) => void;
  /** Client key provider for approved client signature */
  clientKeyProvider: ClientKeyProvider;
  /** User key provider for user signature */
  userKeyProvider: UserKeyProvider;
  /** Optional callback when guidance state changes */
  onGuidanceChange?: (state: GuidanceState) => void;
  /** Optional custom quality thresholds */
  qualityThresholds?: Partial<QualityThresholds>;
  /** Enable debug mode */
  debug?: boolean;
  /** Custom class name */
  className?: string;
  /** Session ID for tracking */
  sessionId?: string;
}

/**
 * Selfie capture component props
 */
export interface SelfieCaptureProps {
  /** Capture mode */
  mode: SelfieCaptureMode;
  /** Enable liveness check */
  livenessCheck?: boolean;
  /** Callback when capture is successful */
  onCapture: (result: SelfieResult) => void;
  /** Callback when an error occurs */
  onError: (error: CaptureError) => void;
  /** Client key provider for approved client signature */
  clientKeyProvider: ClientKeyProvider;
  /** User key provider for user signature */
  userKeyProvider: UserKeyProvider;
  /** Optional callback when guidance state changes */
  onGuidanceChange?: (state: GuidanceState) => void;
  /** Enable debug mode */
  debug?: boolean;
  /** Custom class name */
  className?: string;
  /** Session ID for tracking */
  sessionId?: string;
}

/**
 * Quality thresholds configuration
 */
export interface QualityThresholds {
  /** Minimum resolution (width * height) */
  minResolution: { width: number; height: number };
  /** Minimum brightness (0-255) */
  minBrightness: number;
  /** Maximum brightness (0-255) */
  maxBrightness: number;
  /** Maximum blur (Laplacian variance threshold) */
  maxBlur: number;
  /** Maximum skew angle (degrees) */
  maxSkew: number;
  /** Maximum glare percentage */
  maxGlare: number;
  /** Maximum noise level */
  maxNoise: number;
  /** Minimum overall score to pass */
  minScore: number;
}

/**
 * Default quality thresholds
 */
export const DEFAULT_QUALITY_THRESHOLDS: QualityThresholds = {
  minResolution: { width: 1024, height: 768 },
  minBrightness: 40,
  maxBrightness: 220,
  maxBlur: 100,
  maxSkew: 10,
  maxGlare: 0.15,
  maxNoise: 0.2,
  minScore: 70,
};

// ============================================================================
// Signature Packaging Types
// ============================================================================

/**
 * Signature package for upload
 */
export interface SignaturePackage {
  /** Salt used for this upload */
  salt: Uint8Array;
  /** Hash of the payload */
  payloadHash: Uint8Array;
  /** Client signature over (salt || payloadHash) */
  clientSignature: Uint8Array;
  /** User signature over (salt || payloadHash || clientSignature) */
  userSignature: Uint8Array;
  /** Client ID */
  clientId: string;
  /** Client version */
  clientVersion: string;
  /** User account address */
  userAddress: string;
  /** Timestamp of signing */
  signedAt: string;
}
