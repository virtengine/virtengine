import type { ImageAsset } from "../../core/captureModels";

export interface CameraAdapter {
  isAvailable: () => boolean;
  requestPermission: () => Promise<boolean>;
  capturePhoto: (label: string) => Promise<ImageAsset>;
}

export function createMockImageAsset(label: string): ImageAsset {
  return {
    uri: `mock://${label}`,
    width: 1080,
    height: 720,
    format: "jpeg",
    timestamp: Date.now()
  };
}
