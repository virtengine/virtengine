export type LivenessChallengeType = "blink" | "turn_left" | "turn_right" | "smile" | "hold_still";

export interface FaceSignal {
  timestamp: number;
  faceConfidence: number;
  yaw: number;
  pitch: number;
  roll: number;
  leftEyeOpenProbability?: number;
  rightEyeOpenProbability?: number;
  smileProbability?: number;
}

export interface LivenessChallenge {
  id: string;
  type: LivenessChallengeType;
  instruction: string;
  timeoutMs: number;
}

export interface LivenessConfig {
  minFaceConfidence: number;
  blinkEyeClosedThreshold: number;
  blinkEyeOpenThreshold: number;
  yawThreshold: number;
  smileThreshold: number;
  holdStillDurationMs: number;
}

export interface LivenessUpdate {
  challengeId: string;
  challengeType: LivenessChallengeType;
  instruction: string;
  completed: boolean;
  progress: number;
  note?: string;
}
