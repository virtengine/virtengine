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

export interface OcrField {
  key: string;
  value: string;
  confidence: number;
}

export interface OcrResult {
  rawText: string;
  fields: OcrField[];
}

export interface DeviceAttestation {
  deviceId: string;
  deviceModel: string;
  osVersion: string;
  appVersion: string;
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
  ocr?: OcrResult;
  deviceAttestation?: DeviceAttestation;
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
