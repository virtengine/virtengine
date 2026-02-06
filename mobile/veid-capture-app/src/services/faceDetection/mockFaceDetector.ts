import type { FaceDetector } from "./faceDetector";

export const mockFaceDetector: FaceDetector = {
  detect: async () => [
    {
      faceConfidence: 0.95,
      yaw: 0,
      roll: 0,
      leftEyeOpenProbability: 0.9,
      rightEyeOpenProbability: 0.9,
      smileProbability: 0.4
    }
  ]
};
