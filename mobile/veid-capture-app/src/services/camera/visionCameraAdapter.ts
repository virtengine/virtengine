import type { ImageAsset } from "../../core/captureModels";
import type { CameraAdapter } from "./cameraAdapter";
import { CameraUnavailableError } from "./cameraAdapter";

interface VisionCameraModule {
  Camera?: {
    getAvailableCameraDevices?: () => Promise<unknown[]> | unknown[];
    requestCameraPermission?: () => Promise<string>;
  };
  getAvailableCameraDevices?: () => Promise<unknown[]> | unknown[];
  requestCameraPermission?: () => Promise<string>;
}

export interface VisionCameraAdapterOptions {
  moduleLoader?: () => Promise<VisionCameraModule | null>;
  capturePhoto?: (label: string) => Promise<ImageAsset>;
  timeoutMs?: number;
}

async function loadVisionCamera(): Promise<VisionCameraModule | null> {
  try {
    return (await import("react-native-vision-camera")) as unknown as VisionCameraModule;
  } catch {
    return null;
  }
}

export function createVisionCameraAdapter(options: VisionCameraAdapterOptions = {}): CameraAdapter {
  const moduleLoader = options.moduleLoader ?? loadVisionCamera;
  const timeoutMs = options.timeoutMs ?? 10_000;
  let available = false;

  return {
    isAvailable: () => available,
    requestPermission: async () => {
      const module = await mapCameraFailure(
        withTimeout(moduleLoader(), timeoutMs),
        "camera_module_unavailable",
        "Native camera module failed to load."
      );
      if (!module) {
        throw new CameraUnavailableError(
          "camera_module_unavailable",
          "Native camera module is not installed."
        );
      }

      const getDevices = module.Camera?.getAvailableCameraDevices ?? module.getAvailableCameraDevices;
      if (!getDevices) {
        throw new CameraUnavailableError(
          "camera_device_unavailable",
          "Native camera module cannot enumerate camera devices."
        );
      }

      const devices = await mapCameraFailure(
        withTimeout(Promise.resolve().then(() => getDevices()), timeoutMs),
        "camera_device_unavailable",
        "Camera devices could not be enumerated."
      );
      if (devices.length === 0) {
        throw new CameraUnavailableError("camera_device_unavailable", "No camera device is available.");
      }

      const requestPermission =
        module.Camera?.requestCameraPermission ?? module.requestCameraPermission;
      if (!requestPermission) {
        throw new CameraUnavailableError(
          "camera_module_unavailable",
          "Native camera module cannot request camera permission."
        );
      }

      const status = await mapCameraFailure(
        withTimeout(Promise.resolve().then(() => requestPermission()), timeoutMs),
        "camera_permission_denied",
        "Camera permission request failed."
      );
      if (status !== "authorized" && status !== "granted") {
        throw new CameraUnavailableError("camera_permission_denied", "Camera permission was not granted.");
      }

      available = true;
      return true;
    },
    capturePhoto: async (label: string) => {
      if (!available) {
        throw new CameraUnavailableError(
          "camera_device_unavailable",
          "Camera capture requires an available, permitted device."
        );
      }

      if (!options.capturePhoto) {
        throw new CameraUnavailableError(
          "camera_capture_unavailable",
          "Native camera capture is not bound to the camera view."
        );
      }

      try {
        return await withTimeout(options.capturePhoto(label), timeoutMs);
      } catch (error) {
        if (error instanceof CameraUnavailableError) {
          throw error;
        }
        throw new CameraUnavailableError("camera_capture_failed", "Native camera capture failed.");
      }
    }
  };
}

async function withTimeout<T>(operation: Promise<T>, timeoutMs: number): Promise<T> {
  let timeoutHandle: ReturnType<typeof setTimeout> | undefined;
  const timeout = new Promise<never>((_, reject) => {
    timeoutHandle = setTimeout(() => {
      reject(new CameraUnavailableError("camera_timeout", "Camera operation timed out."));
    }, timeoutMs);
  });

  try {
    return await Promise.race([operation, timeout]);
  } finally {
    if (timeoutHandle) {
      clearTimeout(timeoutHandle);
    }
  }
}

async function mapCameraFailure<T>(
  operation: Promise<T>,
  code: ConstructorParameters<typeof CameraUnavailableError>[0],
  message: string
): Promise<T> {
  try {
    return await operation;
  } catch (error) {
    if (error instanceof CameraUnavailableError) {
      throw error;
    }
    throw new CameraUnavailableError(code, message);
  }
}

export const visionCameraAdapter = createVisionCameraAdapter();
