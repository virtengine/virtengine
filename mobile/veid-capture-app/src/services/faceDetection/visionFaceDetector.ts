import type { FaceDetector, FaceDetectionResult } from "./faceDetector";

let cachedModule: any;

async function loadFaceDetector() {
  if (cachedModule) {
    return cachedModule;
  }

  try {
    cachedModule = await import("vision-camera-face-detector");
    return cachedModule;
  } catch (error) {
    return null;
  }
}

export const visionFaceDetector: FaceDetector = {
  detect: async (frame: unknown): Promise<FaceDetectionResult[]> => {
    const module = await loadFaceDetector();
    if (!module) {
      return [];
    }

    const faces = module.scanFaces(frame) ?? [];
    return faces.map((face: any) => ({
      faceConfidence: 0.8,
      yaw: face.yawAngle ?? 0,
      roll: face.rollAngle ?? 0,
      leftEyeOpenProbability: face.leftEyeOpenProbability,
      rightEyeOpenProbability: face.rightEyeOpenProbability,
      smileProbability: face.smilingProbability,
      bounds: face.bounds
    }));
  }
};
