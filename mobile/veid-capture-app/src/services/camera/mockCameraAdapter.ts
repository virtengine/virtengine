import type { CameraAdapter } from "./cameraAdapter";
import { createMockImageAsset } from "./cameraAdapter";

export const mockCameraAdapter: CameraAdapter = {
  isAvailable: () => true,
  requestPermission: async () => true,
  capturePhoto: async (label: string) => createMockImageAsset(label)
};
