import type { CameraAdapter } from "./cameraAdapter";
import { createMockImageAsset } from "./cameraAdapter";

let cachedModule: any;

async function loadVisionCamera() {
  if (cachedModule) {
    return cachedModule;
  }

  try {
    cachedModule = await import("react-native-vision-camera");
    return cachedModule;
  } catch (error) {
    return null;
  }
}

export const visionCameraAdapter: CameraAdapter = {
  isAvailable: () => true,
  requestPermission: async () => {
    const module = await loadVisionCamera();
    if (!module) {
      return false;
    }

    const status = await module.requestCameraPermission();
    return status === "authorized";
  },
  capturePhoto: async (label: string) => {
    const module = await loadVisionCamera();
    if (!module) {
      return createMockImageAsset(label);
    }

    // Real capture is implemented by the camera view component.
    // This fallback returns a placeholder until native binding is wired.
    return createMockImageAsset(label);
  }
};
