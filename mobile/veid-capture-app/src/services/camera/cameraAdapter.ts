import type { ImageAsset } from "../../core/captureModels";

export interface CameraAdapter {
  isAvailable: () => boolean;
  requestPermission: () => Promise<boolean>;
  capturePhoto: (label: string) => Promise<ImageAsset>;
}

export type CameraUnavailableCode =
  | "camera_module_unavailable"
  | "camera_device_unavailable"
  | "camera_permission_denied"
  | "camera_timeout"
  | "camera_capture_unavailable"
  | "camera_capture_failed";

export class CameraUnavailableError extends Error {
  readonly code: CameraUnavailableCode;

  constructor(code: CameraUnavailableCode, message: string) {
    super(message);
    this.name = "CameraUnavailableError";
    this.code = code;
  }
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
