import { describe, expect, it } from "vitest";
import type { BiometricCapture, DocumentCapture, OcrResult } from "../core/captureModels";
import { captureReducer, createInitialCaptureState } from "../state/captureStore";

describe("capture store wipe", () => {
  it("zeros buffers and removes sensitive data from old state references", () => {
    const state = createInitialCaptureState();
    const buffer = new Uint8Array([1, 2, 3]);
    const document = {
      side: "front",
      image: { uri: "file:///document.jpg" }
    } as DocumentCapture;
    const biometric = {
      template: "biometric-template",
      liveness: { detectedSignals: ["pulse"] },
      antiSpoofing: { signals: ["depth"] }
    } as BiometricCapture;
    const ocr = { rawText: "raw identity", fields: [{ value: "Alice" }] } as OcrResult;

    state.session.documentFront = document;
    state.session.biometric = biometric;
    state.session.ocr = ocr;
    state.session.socialMedia = [{ profileNameHash: "hash" }] as never[];
    state.temporaryEmbeddingReferences.push("embedding://temporary");
    state.encryptedPayloadBuffers.push(buffer);
    state.offlineQueuePlaintext.push({ raw: "queued identity" });

    const reset = captureReducer(state, { type: "wipe" });

    expect([...buffer]).toEqual([0, 0, 0]);
    expect(document.image.uri).toBe("");
    expect(biometric.template).toBe("");
    expect(ocr.rawText).toBe("");
    expect(state.temporaryEmbeddingReferences).toEqual([]);
    expect(state.encryptedPayloadBuffers).toEqual([]);
    expect(state.offlineQueuePlaintext).toEqual([]);
    expect(reset.session.documentFront).toBeUndefined();
    expect(reset.currentStep).toBe("consent");
  });
});