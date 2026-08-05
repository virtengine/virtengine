import { describe, expect, it } from "vitest";
import { toTerminalCameraError } from "../core/cameraFailure";
import { VerificationTerminalError } from "../core/verificationError";

describe("camera capture safety", () => {
  it("turns camera unavailability into a terminal verification error", () => {
    const error = toTerminalCameraError("camera_permission_denied");
    expect(error).toBeInstanceOf(VerificationTerminalError);
    expect(error.code).toBe("camera_permission_denied");
  });
});
