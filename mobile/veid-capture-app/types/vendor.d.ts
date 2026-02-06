declare module "@react-native-ml-kit/text-recognition" {
  const TextRecognition: {
    recognize: (imageUri: string) => Promise<{ text?: string }>;
  };
  export default TextRecognition;
}

declare module "react-native-vision-camera" {
  import * as React from "react";
  import { ViewProps } from "react-native";

  export type CameraDevice = { id: string; position: "front" | "back" };
  export type CameraPermissionStatus = "authorized" | "denied" | "not-determined";

  export const Camera: React.ComponentType<ViewProps & {
    device: CameraDevice;
    isActive: boolean;
  }>;

  export const useCameraDevices: () => { front?: CameraDevice; back?: CameraDevice };
  export const getCameraPermissionStatus: () => Promise<CameraPermissionStatus>;
  export const requestCameraPermission: () => Promise<CameraPermissionStatus>;
}

declare module "vision-camera-face-detector" {
  export interface FaceDetectorResult {
    bounds?: { x: number; y: number; width: number; height: number };
    rollAngle?: number;
    yawAngle?: number;
    leftEyeOpenProbability?: number;
    rightEyeOpenProbability?: number;
    smilingProbability?: number;
  }

  export const scanFaces: (frame: unknown, options?: Record<string, unknown>) => FaceDetectorResult[];
}
