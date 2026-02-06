export interface FaceDetectionResult {
  faceConfidence: number;
  yaw: number;
  roll: number;
  leftEyeOpenProbability?: number;
  rightEyeOpenProbability?: number;
  smileProbability?: number;
  bounds?: { x: number; y: number; width: number; height: number };
}

export interface FaceDetector {
  detect: (frame: unknown) => Promise<FaceDetectionResult[]>;
}
