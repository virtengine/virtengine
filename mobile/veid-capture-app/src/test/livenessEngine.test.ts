import { describe, expect, it } from "vitest";
import { LivenessEngine, DEFAULT_LIVENESS_CONFIG } from "../core/liveness/engine";
import type { FaceSignal } from "../core/liveness/types";

const baseSignal: FaceSignal = {
  timestamp: Date.now(),
  faceConfidence: 0.9,
  yaw: 0,
  pitch: 0,
  roll: 0,
  leftEyeOpenProbability: 0.9,
  rightEyeOpenProbability: 0.9,
  smileProbability: 0.1
};

function advance(engine: LivenessEngine, signal: Partial<FaceSignal>): void {
  engine.update({ ...baseSignal, ...signal });
}

describe("LivenessEngine", () => {
  it("completes a blink challenge", () => {
    const engine = new LivenessEngine([
      { id: "blink", type: "blink", instruction: "blink", timeoutMs: 2000 }
    ], DEFAULT_LIVENESS_CONFIG);

    engine.start();
    advance(engine, { leftEyeOpenProbability: 0.05, rightEyeOpenProbability: 0.05 });
    advance(engine, { timestamp: baseSignal.timestamp + 200, leftEyeOpenProbability: 0.9, rightEyeOpenProbability: 0.9 });

    const result = engine.getResult();
    expect(result.passed).toBe(true);
    expect(result.challenges[0]?.completed).toBe(true);
  });

  it("handles timeout as failure", () => {
    const engine = new LivenessEngine([
      { id: "turn", type: "turn_left", instruction: "turn", timeoutMs: 10 }
    ], DEFAULT_LIVENESS_CONFIG);

    engine.start();
    advance(engine, { timestamp: baseSignal.timestamp + 1000, yaw: 0 });

    const result = engine.getResult();
    expect(result.passed).toBe(false);
  });
});
