import { describe, expect, it } from "vitest";
import { CameraUnavailableError } from "../services/camera/cameraAdapter";
import { createVisionCameraAdapter } from "../services/camera/visionCameraAdapter";

describe("createVisionCameraAdapter", () => {
  it("fails closed when the native module is unavailable", async () => {
    const adapter = createVisionCameraAdapter({ moduleLoader: async () => null });

    await expect(adapter.requestPermission()).rejects.toMatchObject({
      name: "CameraUnavailableError",
      code: "camera_module_unavailable"
    });
    expect(adapter.isAvailable()).toBe(false);
  });

  it("fails closed when no camera device exists", async () => {
    const adapter = createVisionCameraAdapter({
      moduleLoader: async () => ({
        getAvailableCameraDevices: async () => [],
        requestCameraPermission: async () => "authorized"
      })
    });

    await expect(adapter.requestPermission()).rejects.toMatchObject({
      code: "camera_device_unavailable"
    });
  });

  it("does not manufacture an image without a native capture binding", async () => {
    const adapter = createVisionCameraAdapter({
      moduleLoader: async () => ({
        getAvailableCameraDevices: async () => [{ id: "back" }],
        requestCameraPermission: async () => "authorized"
      })
    });

    await expect(adapter.requestPermission()).resolves.toBe(true);
    await expect(adapter.capturePhoto("document_front")).rejects.toEqual(
      expect.objectContaining<Partial<CameraUnavailableError>>({
        code: "camera_capture_unavailable"
      })
    );
  });

  it("fails closed with a typed timeout", async () => {
    const adapter = createVisionCameraAdapter({
      moduleLoader: () => new Promise(() => undefined),
      timeoutMs: 1
    });

    await expect(adapter.requestPermission()).rejects.toMatchObject({
      code: "camera_timeout"
    });
  });

  it("maps native bridge failures to typed unavailable errors", async () => {
    const adapter = createVisionCameraAdapter({
      moduleLoader: async () => ({
        getAvailableCameraDevices: () => {
          throw new Error("native bridge failure");
        },
        requestCameraPermission: async () => "authorized"
      })
    });

    await expect(adapter.requestPermission()).rejects.toMatchObject({
      name: "CameraUnavailableError",
      code: "camera_device_unavailable"
    });
  });
});