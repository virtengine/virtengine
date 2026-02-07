import type { DeviceAttestation } from "./captureModels";
import { createId } from "../utils/id";

export function createDeviceAttestation(appVersion: string): DeviceAttestation {
  return {
    deviceId: createId("device"),
    deviceModel: "unknown",
    osVersion: "unknown",
    appVersion,
    attestedAt: Date.now(),
    attestationSignature: "attestation_pending"
  };
}
