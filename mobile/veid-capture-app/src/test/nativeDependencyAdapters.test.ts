import { describe, expect, it } from "vitest";
import { VerificationTerminalError } from "../core/verificationError";
import { createOcrService } from "../services/ocr/ocrService";
import { createVisionFaceDetector } from "../services/faceDetection/visionFaceDetector";

describe("declared native dependency adapters", () => {
  it("makes an unavailable declared OCR bridge terminal", async () => {
    const extract = createOcrService({ loadRecognizer: async () => { throw new Error("bridge missing"); } });
    await expect(extract("file://document.jpg")).rejects.toMatchObject({
      name: "VerificationTerminalError", code: "ocr_module_unavailable"
    } satisfies Partial<VerificationTerminalError>);
  });

  it("distinguishes OCR recognition failures from module availability", async () => {
    const extract = createOcrService({ loadRecognizer: async () => ({ recognize: async () => { throw new Error("native failure"); } }) });
    await expect(extract("file://document.jpg")).rejects.toMatchObject({ code: "ocr_recognition_failed" });
  });

  it("does not turn a missing face detector into an empty face list", async () => {
    const detector = createVisionFaceDetector({ moduleLoader: async () => { throw new Error("bridge missing"); } });
    await expect(detector.detect({})).rejects.toMatchObject({ code: "face_detection_failed" });
  });

  it("rejects face results that omit a real confidence score", async () => {
    const detector = createVisionFaceDetector({ moduleLoader: async () => ({ scanFaces: () => [{ yawAngle: 0 }] }) });
    await expect(detector.detect({})).rejects.toMatchObject({ code: "face_detection_failed" });
  });

  it("preserves the native detector confidence instead of inventing one", async () => {
    const detector = createVisionFaceDetector({ moduleLoader: async () => ({ scanFaces: () => [{ confidence: 0.73, yawAngle: 1 }] }) });
    await expect(detector.detect({})).resolves.toMatchObject([{ faceConfidence: 0.73, yaw: 1 }]);
  });
});
