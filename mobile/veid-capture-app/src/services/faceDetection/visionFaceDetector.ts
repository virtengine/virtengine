import type { FaceDetector, FaceDetectionResult } from "./faceDetector";
import { VerificationTerminalError } from "../../core/verificationError";

let cachedModule: any;

async function loadFaceDetector() {
  if (cachedModule) {
    return cachedModule;
  }

  try {
    cachedModule = await import("vision-camera-face-detector");
    return cachedModule;
  } catch (error) {
    throw new VerificationTerminalError("face_detector_unavailable", "Native face detector is unavailable.", { cause: error });
  }
}

export const visionFaceDetector: FaceDetector = {
  detect: async (frame: unknown): Promise<FaceDetectionResult[]> => {
    try {
      const module = await loadFaceDetector();
      const faces = module.scanFaces(frame);
      if (!Array.isArray(faces)) {
        throw new VerificationTerminalError("face_detection_failed", "Face detector returned an invalid result.");
      }
      return faces.map((face: any) => ({
      faceConfidence: 0.8,
      yaw: face.yawAngle ?? 0,
      roll: face.rollAngle ?? 0,
      leftEyeOpenProbability: face.leftEyeOpenProbability,
      rightEyeOpenProbability: face.rightEyeOpenProbability,
      smileProbability: face.smilingProbability,
        bounds: face.bounds
      }));
    } catch (error) {
      if (error instanceof VerificationTerminalError) throw error;
      throw new VerificationTerminalError("face_detection_failed", "Face detection failed.", { cause: error });
    }
  }
};
