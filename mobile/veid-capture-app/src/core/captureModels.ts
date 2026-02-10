export type DocumentType = "id_card" | "passport" | "drivers_license";
export type DocumentSide = "front" | "back";

export interface ImageAsset {
  uri: string;
  width: number;
  height: number;
  format: "jpeg" | "png" | "heic" | "unknown";
  timestamp: number;
}

export interface DocumentCapture {
  type: DocumentType;
  side: DocumentSide;
  image: ImageAsset;
  qualityScore: number;
  warnings: string[];
}

export interface SelfieCapture {
  image: ImageAsset;
  faceConfidence: number;
  guidance: string[];
}

export interface LivenessChallengeResult {
  challengeId: string;
  type: "blink" | "turn_left" | "turn_right" | "smile" | "hold_still";
  completed: boolean;
  startedAt: number;
  completedAt?: number;
  attempts: number;
  notes?: string;
}

export interface LivenessResult {
  passed: boolean;
  score: number;
  startedAt: number;
  completedAt: number;
  challenges: LivenessChallengeResult[];
  failureReason?: string;
}

export type SocialMediaProvider = "google" | "facebook" | "microsoft";

export interface SocialMediaProfile {
  provider: SocialMediaProvider;
  profileNameHash: string;
  emailHash?: string;
  usernameHash?: string;
  orgHash?: string;
  accountCreatedAt?: number;
  accountAgeDays?: number;
  isVerified: boolean;
  friendCountRange?: string;
  connectedAt: number;
}

export type BiometricModality = "fingerprint" | "iris";
export type BiometricSensorType = "optical" | "capacitive" | "ultrasonic" | "iris" | "unknown";
export type BiometricSecurityLevel = "unknown" | "basic" | "strong" | "hardware_backed";

export interface BiometricLivenessResult {
  passed: boolean;
  score: number;
  method: "hardware" | "software" | "combined";
  detectedSignals: string[];
}

export interface BiometricAntiSpoofingResult {
  passed: boolean;
  score: number;
  signals: string[];
}

export interface BiometricDeviceInfo {
  manufacturer: string;
  model: string;
  sensorType: BiometricSensorType;
  securityLevel: BiometricSecurityLevel;
  firmwareVersion: string;
}

export interface BiometricCapture {
  modality: BiometricModality;
  templateFormat: "iso_19794_2" | "iso_19794_6" | "vendor" | "unknown";
  template: string;
  capturedAt: number;
  liveness: BiometricLivenessResult;
  antiSpoofing: BiometricAntiSpoofingResult;
  deviceInfo: BiometricDeviceInfo;
  supported: boolean;
  failureReason?: string;
}

export interface OcrField {
  key: string;
  value: string;
  confidence: number;
}

export interface OcrResult {
  rawText: string;
  fields: OcrField[];
}

export type DevicePlatform = "android" | "ios" | "unknown";
export type DeviceAttestationProvider =
  | "android_play_integrity"
  | "android_safetynet"
  | "ios_devicecheck"
  | "ios_app_attest"
  | "mock";
export type DeviceIntegrityLevel = "unknown" | "basic" | "strong" | "hardware_backed" | "unsupported";

export interface DeviceAttestation {
  deviceId: string;
  deviceModel: string;
  osVersion: string;
  appVersion: string;
  appId: string;
  platform: DevicePlatform;
  provider: DeviceAttestationProvider;
  integrityLevel: DeviceIntegrityLevel;
  integrityScore: number;
  supported: boolean;
  failureReason?: string;
  nonce: string;
  verdicts: Record<string, boolean>;
  attestationPayload: string;
  attestedAt: number;
  attestationSignature: string;
}

export interface CaptureSession {
  sessionId: string;
  createdAt: number;
  documentType: DocumentType;
  documentFront?: DocumentCapture;
  documentBack?: DocumentCapture;
  selfie?: SelfieCapture;
  liveness?: LivenessResult;
  biometric?: BiometricCapture;
  ocr?: OcrResult;
  deviceAttestation?: DeviceAttestation;
  socialMedia?: SocialMediaProfile[];
}

export interface CapturePayload {
  session: CaptureSession;
  encryptedPayload: string;
  payloadHash: string;
  transport: {
    uploadUrl: string;
    retryCount: number;
  };
}
